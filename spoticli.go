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
const tokendir = "/.spoticli"
const tokenfile = tokendir + "/token.gob"

var (
	auth = spotify.NewAuthenticator(
		redirectURL,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
	)
	state = "SpotiCLI-Remote"
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
	// Destination variables for command flags
	var (
		devid      int
		volpercent int
	)

	app := cli.NewApp()
	app.Name = "SpotiCLI-Remote"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Luke Hobbs",
			Email: "lukeehobbs@gmail.com",
		},
	}
	app.Usage = "Control Spotify Connect enabled devices via terminal."
	app.UsageText = "spotify> command [arguments...]"

	app.Commands = []cli.Command{
		{
			Name:    "devices",
			Aliases: []string{"d"},
			Usage:   "List available devices and their IDs",
			Action: func(c *cli.Context) error {
				listDevices()
				return nil
			},
		},
		{
			Name:    "play",
			Aliases: []string{"p"},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:        "device, d",
					Usage:       "Transfer playback to `DEVICE_NUMBER`",
					Destination: &devid,
				},
			},
			Usage: "Start/Resume playback on device, or currently playing device if none specified",
			Action: func(c *cli.Context) error {
				if c.IsSet("device") {
					if devid > 25 || devid < 0 { // Assuming user will not have more than 25 devices
						fmt.Println("Incorrect Usage: argument is not a valid device ID: ", devid)
						cli.ShowCommandHelp(c, "play")
						return nil
					}
					play(devid)
				}
				play(0)
				return nil
			},
		},
		{
			Name:    "pause",
			Aliases: []string{"pp"},
			Usage:   "Pause playback on currently playing device",
			Action: func(c *cli.Context) error {
				pause()
				return nil
			},
		},
		{
			Name:    "vol",
			Aliases: []string{"v"},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:        "set, s",
					Usage:       "Set volume to `PERCENT`",
					Value:       -1, //  Percent is not specified
					Destination: &volpercent,
				},
				cli.BoolFlag{
					Name:  "up, u",
					Usage: "Increase volume by 10%",
				},
				cli.BoolFlag{
					Name:  "down, d",
					Usage: "Decrease volume by 10%",
				},
			},
			Usage: "Change volume on currently playing device",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "vol")
          return nil
				}
        defer func() {  // Recover if no devices are active
          if r := recover(); r != nil {
            fmt.Println(r)
          }
        }()
				if c.NumFlags() == 2 { // Flag is one of: [--up, --down, --set]
					if c.IsSet("up") {
						volumePlus(10)
						return nil
					}
					if c.IsSet("down") {
						volumePlus(-10)
						return nil
					}
					if c.IsSet("set") {
						setVolume(volpercent)
						return nil
					}
				}
				if c.NumFlags() == 0 {
          current := getVolume()
					fmt.Println("Volume: ", current)
					return nil
				}
				// ERROR
				return nil
			},
		},
		{
			Name:    "current",
			Aliases: []string{"c"},
			Usage:   "Display information about the currently playing song",
			Action: func(c *cli.Context) error {
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
		app.Run(strings.Fields("spoticli " + line))
	}

	app.Run(os.Args)
}

func setVolume(p int) {
	client := auth.NewClient(tok)
	if err := client.Volume(p); err != nil {
		panic(err)
	}
}

func getVolume() int {
  current := -1
  client := auth.NewClient(tok)
  devices, err := client.PlayerDevices()
	if err != nil {
		panic(err)
	}
	for _, v := range devices {
		if v.Active == true {
      current = v.Volume
		}
	}
  if current == -1 {
    panic("Error: no devices are active, please begin playback first")
  }
  return current
}

func volumePlus(v int) {
  current := getVolume()
  newvol := current + v
  if newvol > 100 { newvol = 100}
  if newvol < 0 {newvol = 0}
  setVolume(newvol)
}

func play(i int) {
	client := auth.NewClient(tok)
	if i == 0 {
		err := client.Play()
		if err != nil {
			panic(err)
		}
		return
	}
	devices, err := client.PlayerDevices()
	if err != nil {
		panic(err)
	}
	ID := []spotify.ID{devices[i-1].ID}
	err = client.TransferPlayback(ID, true)
	if err != nil {
		panic(err)
	}
}

func pause() {
	client := auth.NewClient(tok)
	err := client.Pause()
	if err != nil {
		panic(err)
	}
}

func listDevices() {
	client := auth.NewClient(tok)
	devices, err := client.PlayerDevices()
	if err != nil {
		panic(err)
	}
	for i, v := range devices {
		fmt.Printf("[%d]=%v (%v)", i+1, v.Name, v.Type)
		if v.Active == true {
			fmt.Printf(" ACTIVE\n")
		} else {
			fmt.Printf("\n")
		}
	}
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
	_ = os.Mkdir(usr.HomeDir+"/.spoticli", 0755)
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
