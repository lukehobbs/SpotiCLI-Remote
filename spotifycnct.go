package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/bobappleyard/readline"
	"github.com/urfave/cli"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURL = "http://localhost:8080/callback"
const tokendir = "/.spotifycnct"
const tokenfile = tokendir + "/token.gob"

var (
	auth  = spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate, spotify.ScopeUserReadPlaybackState)
	state = "spotifycnct"
	ch    = make(chan *spotify.Client)
	tok   *oauth2.Token
)

func init() {
	err := loadToken()
	if err != nil {
		startAuth()
	} else {
		// Create new client from the loaded token
		client := auth.NewClient(tok)
		// Save new token
		err = saveToken(tok)
		if err != nil {
			panic(err)
		}
		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			panic(err)
		}
		fmt.Println("You are logged in as:", user.ID)
	}
}

func startAuth() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
}

func main() {
	app := cli.NewApp()
	app.Name = "spotifycnct"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Luke Hobbs",
			Email: "lukeehobbs@gmail.com",
		},
	}
	app.Usage = "Control Spotify Connect enabled devices via terminal."
	app.UsageText = "command [command options] [arguments...]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, config",
			Usage: "Load configuration from `FILE`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "devices",
			Aliases: []string{"d"},
			Usage:   "List available devices",
			Action: func(c *cli.Context) error {
				listDevices()
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"q"},
			Usage:   "Quit application",
			Action: func(c *cli.Context) error {
				os.Exit(0)
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	for {
		line, err := readline.String("spotify> ")
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error: ", err)
			break
		}
		readline.AddHistory(line)
		app.Run(strings.Fields("spotifycnct " + line))
	}

	app.Run(os.Args)
}

func listDevices() {
  client := auth.NewClient(tok)
  devices, err := client.PlayerDevices()
  if err != nil {
    panic(err)
  }
  fmt.Println("Available devices: ", devices)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
	err = saveToken(tok)
	if err != nil {
		panic(err)
	}
}

func saveToken(t *oauth2.Token) error {
	tok := &t
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Mkdir(usr.HomeDir+"/.spotifycnct", 0755)
	tokenpath := usr.HomeDir + tokenfile
	file, err := os.OpenFile(tokenpath, os.O_CREATE|os.O_RDWR, 0755)
	if err == nil {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(tok)
	}
	file.Close()
	return err
}

func loadToken() error {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	tokenpath := usr.HomeDir + tokenfile
	file, err := os.Open(tokenpath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&tok)
	}
	file.Close()
	return err
}
