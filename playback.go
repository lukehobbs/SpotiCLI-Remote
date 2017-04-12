package main

import (
	"fmt"
	"github.com/lukehobbs/spotify"
	"github.com/urfave/cli"
	"os"
	"text/template"
)

func setRepeat(s string) {
	client := auth.NewClient(tok)
	switch s {
	case "off":
		err := client.Repeat("off")
		checkErr(err)
	case "track":
		err := client.Repeat("track")
		checkErr(err)
	case "playlist":
		err := client.Repeat("context")
		checkErr(err)
	}
}

func setShuffle(b bool) {
	client := auth.NewClient(tok)
	err := client.Shuffle(b)
	checkErr(err)
}

func optionsAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	checkErr(err)
	t := template.New("optionsTemplate")
	t, err = t.Parse(optionsTemplate)
	checkErr(err)
	err = t.Execute(os.Stdout, state)
	checkErr(err)
}

func currentAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	trk := getCurrentTrack()
	t := template.New("longTrackTemplate")
	t, err := t.Parse(longTrackTemplate)
	checkErr(err)
	err = t.Execute(os.Stdout, trk)
	checkErr(err)
}

func getCurrentTrack() *spotify.FullTrack {
	client := auth.NewClient(tok)
	current, err := client.PlayerCurrentlyPlaying()
	checkErr(err)
	return current.Item
}

func nextAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Next()
	checkErr(err)
}

func prevAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Previous()
	checkErr(err)
}

func setVolume(p int) {
	client := auth.NewClient(tok)
	err := client.Volume(p)
	checkErr(err)
}

func getVolume() int {
	a := -1
	client := auth.NewClient(tok)
	d, err := client.PlayerDevices()
	checkErr(err)
	for _, v := range d {
		if v.Active {
			a = v.Volume
		}
	}
	if a == -1 {
		panic("Error: no devices are active, please begin playback first")
	}
	return a
}

func volumePlus(a int) {
	vol := getVolume()
	s := vol + a
	if s > 100 {
		s = 100
	}
	if s < 0 {
		s = 0
	}
	setVolume(s)
}

func playAction(c *cli.Context) {
	i := c.Int("device")
	client := auth.NewClient(tok)

	if i == 0 {
		err := client.Play()
		checkErr(err)
		return
	}

	d, err := client.PlayerDevices()
	checkErr(err)

	if i > len(d) {
		fmt.Println("ERROR: Incorrect device ID, ", i)
		err = cli.ShowCommandHelp(c, c.App.Name)
		checkErr(err)
		return
	}

	ID := d[i-1].ID
	err = client.TransferPlayback(ID, true)
	checkErr(err)
}

func pauseAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Pause()
	checkErr(err)
}

func devicesAction(c *cli.Context) {
	if c.Args().Present() {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	d, err := client.PlayerDevices()
	checkErr(err)
	for i, v := range d {
		fmt.Printf("[%d]=%v (%v)", i+1, v.Name, v.Type)
		if v.Active {
			fmt.Println(" ACTIVE")
		} else {
			fmt.Println()
		}
	}
}
