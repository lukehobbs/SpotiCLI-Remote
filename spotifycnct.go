package main

import (
    "os"
    "sort"
    "time"

    "github.com/urfave/cli"
)

func main() {
    app := cli.NewApp()
    app.Name = "spotifycnct"
    app.Version = "0.0.1"
    app.Compiled = time.Now()
    app.Authors = []cli.Author {
        cli.Author {
            Name: "Luke Hobbs",
            Email: "lukeehobbs@gmail.com",
        },
    }
    app.Usage = "Control Spotify Connect enabled devices via terminal."
    app.UsageText = "command [command options] [arguments...]"

    app.Flags = []cli.Flag {
        cli.StringFlag {
            Name: "c, config",
            Usage: "Load configuration from `FILE`",
        },
    }

    app.Commands = []cli.Command {
        {
            Name:       "d, devices",
            Usage:      "list available devices",
            Action:     func(c *cli.Context) error {
                return nil
            },
        },
        {
            Name:       "n, now",
            Usage:      "get information about current playback",
            Action:      func(c *cli.Context) error {
                return nil
            },
            Subcommands: []cli.Command {
                {
                    Name:       "p, play",
                    Usage:      "toggle playback",
                    Action:      func(c *cli.Context) error {
                        return nil
                    },
                },
                {
                    Name:       "v, vol",
                    Usage:      "Increase or decrease volume of playback by `VALUE`",
                    Action:      func(c *cli.Context) error {
                        return nil
                    },
                },
            },
        },
    }

    sort.Sort(cli.FlagsByName(app.Flags))
    sort.Sort(cli.CommandsByName(app.Commands))

    app.Run(os.Args)
}
