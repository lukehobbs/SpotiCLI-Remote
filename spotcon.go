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
Repeat:  {{.RepeatState}}
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
	app.HelpName = "spotcon"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Luke Hobbs",
			Email: "lukeehobbs@gmail.com",
		},
	}
	app.Usage = "Control Spotify Connect enabled devices via terminal."
	app.UsageText = "spotify> command [subcommand] [--flags] [arguments...]"

	cli.AppHelpTemplate = `NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}
USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}
DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}
AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

	cli.CommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}
USAGE:
   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{end}}{{if .Category}}
CATEGORY:
   {{.Category}}{{end}}{{if .Description}}
DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	app.Commands = []cli.Command{
		{
			Name:      "devices",
			Aliases:   []string{"d"},
			Usage:     "List available devices",
			ArgsUsage: "",
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
					Usage: "Search for artist, `NAME`, on Spotify",
				},
				cli.BoolFlag{
					Name:  "album, al",
					Usage: "Search albums, `NAME`, on Spotify",
				},
				cli.BoolFlag{
					Name:  "track, tr",
					Usage: "Search tracks, `NAME`, on Spotify",
				},
				cli.BoolFlag{
					Name:  "playlist, pl",
					Usage: "Search playlists, `NAME`, on Spotify",
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
					Usage: "Start/resume playback on specified  `NAME` or number from device list",
				},
				cli.BoolFlag{
					Name:  "track, tr",
					Usage: "Play track with specified `NAME` or number from search results",
				},
				cli.BoolFlag{
					Name:  "album, al",
					Usage: "Play album with specified `NAME` or number from search results",
				},
				cli.BoolFlag{
					Name:  "artist, ar",
					Usage: "Play artist with specified `NAME` or number from search results",
				},
				cli.BoolFlag{
					Name:  "plist, pl",
					Usage: "Play playlist with specified `NAME` or number from search results",
				},
			},
			Usage: "Start/Resume playback",
			Action: func(c *cli.Context) error {
				playAction(c)
				return nil
			},
		},
		{
			Name:    "pause",
			Aliases: []string{"pp"},
			Usage:   "Pause playback",
			Action: func(c *cli.Context) error {
				pauseAction(c)
				return nil
			},
		},
		{
			Name:      "vol",
			Aliases:   []string{"v"},
			Usage:     "Options for changing volume of playback",
			ArgsUsage: "[arguments...]",
			After: func(c *cli.Context) error {
				time.Sleep(150 * time.Millisecond)
				displayVolume()
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "up",
					Usage: "Increase volume by PERCENT or 10% if not specified",
					Action: func(c *cli.Context) error {
						volAdjustAction(c, true)
						return nil
					},
				},
				{
					Name:  "down",
					Usage: "Decrease volume by PERCENT or 10% if not specified",
					Action: func(c *cli.Context) error {
						volAdjustAction(c, false)
						return nil
					},
				},
				{
					Name:  "set",
					Usage: "Set volume to PERCENT",
					Action: func(c *cli.Context) error {
						volSetAction(c)
						return nil
					},
				},
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
				skipAction(c, true)
				return nil
			},
		},
		{
			Name:    "prev",
			Aliases: []string{"pr"},
			Usage:   "Skip to the previous track in queue",
			Action: func(c *cli.Context) error {
				skipAction(c, false)
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
			Name:    "opt",
			Aliases: []string{"o"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "repeat, r",
					Usage: "Set playback option repeat [on, off]",
				},
				cli.StringFlag{
					Name:  "shuffle, s",
					Usage: "Set playback option shuffle [on, off]",
				},
			},
			Usage: "Options for changing current playback parameters",
			After: func(c *cli.Context) error {
				time.Sleep(200 * time.Millisecond)
				displayOpts()
				return nil
			},
			Action: func(c *cli.Context) error {
				optAction(c)
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
