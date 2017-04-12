package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bobappleyard/readline"
	"github.com/urfave/cli"
	"github.com/lukehobbs/spotify"
	//"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURL = "http://localhost:8080/callback"
const tokenDir = "/.spoticli"
const tokenFile = tokenDir + "/token.gob"

const longTrackTemplate = `Track:  {{.Name}}
Artist:	{{range $index, $artist := .Artists}}{{if $index}}, {{end}}{{.Name}}{{end}}
Album:	{{.Album.Name}}
`
const optionsTemplate = `Shuffle: {{if .ShuffleState}}on{{end}}{{if not .ShuffleState}}off{{end}}
Repeat: {{.RepeatState}}
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
		checkErr(err)
		// use the client to make calls that require authorization
		usr, err := client.CurrentUser()
		checkErr(err)
		fmt.Println("You are logged in as:", usr.ID)
	}
}

func main() {
	// Destination variables for command flags
	var (
	//	devid      int
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
					err := cli.ShowCommandHelp(c, "devices")
					checkErr(err)
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
					Value:	     0,
				},
			},
			Usage: "Start/Resume playback on device, or currently playing device if none specified",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					err := cli.ShowCommandHelp(c, "play")
					checkErr(err)
					return nil
				}
				playAction(c)
				return nil
			},
		},
		{
			Name:    "pause",
			Aliases: []string{"pp"},
			Usage:   "Pause playback on currently playing device",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					err := cli.ShowCommandHelp(c, "pause")
					checkErr(err)
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
					err := cli.ShowCommandHelp(c, "vol")
					checkErr(err)
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
					err := cli.ShowCommandHelp(c, "current")
					checkErr(err)
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
					err := cli.ShowCommandHelp(c, "next")
					checkErr(err)
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
					err := cli.ShowCommandHelp(c, "prev")
					checkErr(err)
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
					err := cli.ShowCommandHelp(c, "clear")
					checkErr(err)
					return nil
				}
				clear(c)
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
					err := cli.ShowCommandHelp(c, "shuffle")
					checkErr(err)
					return nil
				}
				if c.IsSet("on") {
					setShuffle(true)
				}
				if c.IsSet("off") {
					setShuffle(false)
				}
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
					err := cli.ShowCommandHelp(c, "repeat")
					checkErr(err)
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
				return nil
			},
		},
		{
			Name:    "options",
			Aliases: []string{"o"},
			Usage:   "Display current playback options",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					err := cli.ShowCommandHelp(c, "options")
					checkErr(err)
					return nil
				}
				displayOptions()
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"q"},
			Usage:   "Quit application",
			Action: func(c *cli.Context) error {
				quitAction(c)
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
		err = app.Run(strings.Fields("spoticli " + line))
		checkErr(err)
	}

	err := app.Run(os.Args)
	checkErr(err)
}

func quitAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	os.Exit(0)
}

func clear(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	_, err := os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
	checkErr(err)
}


func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}
