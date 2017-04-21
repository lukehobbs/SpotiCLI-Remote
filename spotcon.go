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
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURL = "http://localhost:8080/callback"
const tokenDir = "/.spoticli"
const tokenFile = tokenDir + "/token.gob"
const longTrackTemplate = `Track:  {{.Name}}
Artist:	{{range $index, $artist := .Artists}}{{if $index}}, {{end}}{{.Name}}{{end}}
Album:	{{.Album.Name}}
`
const shortTrackTemplate = `"{{.Name}}" by {{range $index, $artist := .Artists}}{{if $index}}, {{end}}{{.Name}}{{end}}
`
const shortAlbumTemplate = `"{{.Name}}" by {{range $index, $artist := .Artists}}{{if $index}}, {{end}}{{.Name}}{{end}}
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
	state = "Spotcon"
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
	app := cli.NewApp()
	app.Name = "Spotcon"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Luke Hobbs",
			Email: "lukeehobbs@gmail.com",
		},
	}
	app.Usage = "Control Spotify Connect enabled devices via terminal."
	app.UsageText = "spotify> command [subcommand] [arguments...]"

	app.Commands = []cli.Command{
		{
			Name:    "devices",
			Aliases: []string{"d"},
			Usage:   "List available devices and their IDs",
			Action: func(c *cli.Context) error {
				devicesAction(c)
				return nil
			},
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "artist, ar",
					Usage: "Search artists on Spotify",
				},
				cli.BoolFlag{
					Name:  "album, al",
					Usage: "Search albums on Spotify",
				},
				cli.BoolFlag{
					Name:  "track, tr",
					Usage: "Search tracks on Spotify",
				},
				cli.BoolFlag{
					Name:  "playlist, pl",
					Usage: "Search playlists on Spotify",
				},
			},
			Usage: "Search Spotify for artists, albums, tracks, or playlists",
			Action: func(c *cli.Context) error {
				searchAction(c)
				return nil
			},
		},
		{
			Name:    "play",
			Aliases: []string{"p"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "device, d",
					Usage: "Start/resume playback on specified device",
				},
				cli.BoolFlag{
					Name: "track, t",
					Usage: "Play track with specified name.",
				},
				cli.BoolFlag{
					Name: "album, al",
					Usage: "Play album with specified name.",
				},
				cli.BoolFlag{
					Name: "artist, ar",
					Usage: "Play artist with specified name.",
				},
				cli.BoolFlag{
					Name: "plist, pl",
					Usage: "Play playlist with specified name",
				},
			},
			Usage:     "Start/Resume playback on device, or currently playing device if none specified",
			Action: func(c *cli.Context) error {
				playAction(c)
				return nil
			},
		},
		{
			Name:    "pause",
			Aliases: []string{"pp"},
			Usage:   "Pause playback on currently playing device",
			Action: func(c *cli.Context) error {
				pauseAction(c)
				return nil
			},
		},
		{
			Name:      "vol",
			Aliases:   []string{"v"},
			Usage:     "Options for changing volume on currently playing device",
			ArgsUsage: "[up | down | <percent>]",
			Action: func(c *cli.Context) error {
				if c.Args().First() == "up" {
					volUpAction(c)
					return nil
				}
				if c.Args().First() == "down" {
					volDownAction(c)
					return nil
				}
				if c.NArg() > 0 {
					volSetAction(c)
					return nil
				}
				fmt.Printf("Volume: %v%%\n", getVolume())
				return nil
			},
		},
		{
			Name:    "current",
			Aliases: []string{"c"},
			Usage:   "Display information about the current playback",
			Action: func(c *cli.Context) error {
				currentAction(c)
				return nil
			},
		},
		{
			Name:    "next",
			Aliases: []string{"n"},
			Usage:   "Skip to the next track in queue",
			Action: func(c *cli.Context) error {
				nextAction(c)
				return nil
			},
		},
		{
			Name:    "prev",
			Aliases: []string{"pr"},
			Usage:   "Skip to the previous track in queue",
			Action: func(c *cli.Context) error {
				prevAction(c)
				return nil
			},
		},
		{
			Name:    "clear",
			Aliases: []string{"clc"},
			Usage:   "Clear the command window",
			Action: func(c *cli.Context) error {
				clearAction(c)
				return nil
			},
		},
		{
			Name:    "shuffle",
			Aliases: []string{"s"},
			Usage:   "Toggle playback option shuffle",
			Action: func(c *cli.Context) error {
				shuffleAction(c)
				return nil
			},
		},
		{
			Name:      "repeat",
			Aliases:   []string{"r"},
			Usage:     "turn on playlist option repeat",
			ArgsUsage: "[off | track | playlist]",
			Action: func(c *cli.Context) error {
				repeatAction(c)
				return nil
			},
		},
		{
			Name:    "options",
			Aliases: []string{"o"},
			Usage:   "Display current playback options",
			Action: func(c *cli.Context) error {
				optionsAction(c)
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
		line, err := readline.String("\nspotcon> ")
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error: ", err)
			break
		}
		readline.AddHistory(line)
		err = app.Run(strings.Fields("spotcon " + line))
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

func clearAction(c *cli.Context) {
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
