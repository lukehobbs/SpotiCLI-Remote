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
	"text/template"
	"time"

	"github.com/bobappleyard/readline"
	"github.com/lukehobbs/spotify"
	"github.com/urfave/cli"
	//"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURL = "http://localhost:8080/callback"
const tokendir = "/.spoticli"
const tokenfile = tokendir + "/token.gob"

const longTrackTemplate = `Track:  {{.Name}}
Artist:	{{range $index, $artist := .Artists}}{{if $index}}, {{end}}{{.Name}}{{end}}
Album:	{{.Album.Name}}
`
const shuffleTemplate = `Shuffle: {{if .ShuffleState}}on{{end}}{{if not .ShuffleState}}off{{end}}
`
const repeatTemplate = `Repeat: {{.RepeatState}}
`

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
		usr, err := client.CurrentUser()
		if err != nil {
			panic(err)
		}
		fmt.Println("You are logged in as:", usr.ID)
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
	usr, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", usr.ID)
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
		{
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
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "devices")
					return nil
				}
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
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "play")
					return nil
				}
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
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "pause")
					return nil
				}
				pause()
				return nil
			},
		},
		{
			Name:    "vol",
			Aliases: []string{"v"},
			Usage:   "Change volume on currently playing device",
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
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "vol")
					return nil
				}
				defer func() { // Recover if no devices are active
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
			Usage:   "Display information about the currently playing track",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "current")
					return nil
				}
				displayCurrentTrack()
				return nil
			},
		},
		{
			Name:    "next",
			Aliases: []string{"n"},
			Usage:   "Skip to the next track in queue",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "next")
					return nil
				}
				next()
				return nil
			},
		},
		{
			Name:    "prev",
			Aliases: []string{"pr"},
			Usage:   "Skip to the previous track in queue",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "prev")
					return nil
				}
				prev()
				return nil
			},
		},
		{
			Name:    "clear",
			Aliases: []string{"clc"},
			Usage:   "Clear the command window",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "clear")
					return nil
				}
				clear()
				return nil
			},
		},
		{
			Name:    "shuffle",
			Aliases: []string{"s"},
			Usage:   "turn on playback option shuffle",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "on",
					Usage: "turn shuffle on",
				},
				cli.BoolFlag{
					Name:  "off",
					Usage: "turn shuffle off",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "shuffle")
					return nil
				}
				if c.IsSet("on") {
					setShuffle(true)
				}
				if c.IsSet("off") {
					setShuffle(false)
				}
				time.Sleep(200 * time.Millisecond) // Wait for shuffleState to update
				displayOptions("shuffle")
				return nil
			},
		},
		{
			Name:    "repeat",
			Aliases: []string{"r"},
			Usage:   "turn on playlist option repeat",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "off",
					Usage: "turn repeat off",
				},
				cli.BoolFlag{
					Name:  "track, t",
					Usage: "turn on repeat track",
				},
				cli.BoolFlag{
					Name:  "playlist, p",
					Usage: "turn on repeat playlist",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "repeat")
					return nil
				}
				if c.IsSet("off") {
					setRepeat("off")
				}
				if c.IsSet("track") {
					setRepeat("track")
				}
				if c.IsSet("playlist") {
					setRepeat("playlist")
				}
				time.Sleep(200 * time.Millisecond) // Wait for shuffleState to update
				displayOptions("repeat")
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"q"},
			Usage:   "Quit application",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					cli.ShowCommandHelp(c, "quit")
					return nil
				}
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

func clear() {
	os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
}

func setRepeat(s string) {
	client := auth.NewClient(tok)
	switch s {
	case "off":
		err := client.Repeat("off")
		if err != nil {
			panic(err)
		}
	case "track":
		err := client.Repeat("track")
		if err != nil {
			panic(err)
		}
	case "playlist":
		err := client.Repeat("context")
		if err != nil {
			panic(err)
		}
	}
}

func setShuffle(b bool) {
	client := auth.NewClient(tok)
	err := client.Shuffle(b)
	if err != nil {
		panic(err)
	}
}

func displayOptions(s string) {
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	if err != nil {
		panic(err)
	}
	switch s {
	case "shuffle":
		t := template.New("shuffleTemplate")
		t, err = t.Parse(shuffleTemplate)
		if err != nil {
			panic(err)
		}
		err = t.Execute(os.Stdout, state)
		if err != nil {
			panic(err)
		}
	case "repeat":
		t := template.New("repeatTemplate")
		t, err = t.Parse(repeatTemplate)
		if err != nil {
			panic(err)
		}
		err = t.Execute(os.Stdout, state)
		if err != nil {
			panic(err)
		}
	}
}

func displayCurrentTrack() {
	track := getCurrentTrack()
	t := template.New("longTrackTemplate")
	t, err := t.Parse(longTrackTemplate)
	if err != nil {
		panic(err)
	}
	err = t.Execute(os.Stdout, track)
	if err != nil {
		panic(err)
	}
}

func getCurrentTrack() *spotify.FullTrack {
	client := auth.NewClient(tok)
	current, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		panic(err)
	}
	return current.Item
}

func next() {
	client := auth.NewClient(tok)
	err := client.Next()
	if err != nil {
		panic(err)
	}
}

func prev() {
	client := auth.NewClient(tok)
	err := client.Previous()
	if err != nil {
		panic(err)
	}
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
	if newvol > 100 {
		newvol = 100
	}
	if newvol < 0 {
		newvol = 0
	}
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
	ID := devices[i-1].ID
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
			fmt.Println(" ACTIVE")
		} else {
			fmt.Println()
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
	fmt.Fprintln(w, "Login Completed!")
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
