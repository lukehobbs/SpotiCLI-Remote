package main

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"text/template"
	"github.com/lukehobbs/spotify"
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

func displayOptions() {
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	checkErr(err)
	t := template.New("optionsTemplate")
	t, err = t.Parse(optionsTemplate)
	checkErr(err)
	err = t.Execute(os.Stdout, state)
	checkErr(err)
}

func displayCurrentTrack() {
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

func next() {
	client := auth.NewClient(tok)
	err := client.Next()
	checkErr(err)
}

func prev() {
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
	new := vol + a
	if new > 100 {
		new = 100
	}
	if new < 0 {
		new = 0
	}
	setVolume(new)
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

func pause() {
	client := auth.NewClient(tok)
	err := client.Pause()
	checkErr(err)
}

func listDevices() {
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