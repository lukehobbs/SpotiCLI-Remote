package main

import (
	"fmt"
	"os"
	"strconv"
	"text/template"
	"time"
	"strings"

	"github.com/zmb3/spotify"
	"github.com/urfave/cli"
)

var LastSearch *spotify.SearchResult

func searchAction(c *cli.Context) {
	var t int
	client := auth.NewClient(tok)
	q := strings.Join(c.Args(), " ")
	if q == "" {
		displayLastSearch()
		return
	}
	if c.Bool("album"){
		t += 1
	}
	if c.Bool("artist"){
		t += 2
	}
	if c.Bool("playlist"){
		t += 4
	}
	if c.Bool("track"){
		t += 8
	}
	if t == 0 {
		t = 15
	}
	st := spotify.SearchType(t)
	var err error
	LastSearch, err = client.Search(q, st)
	checkErr(err)
	displaySearchResults(LastSearch)
}

func displayLastSearch() {
	if LastSearch == nil {
		fmt.Println("No previous search results found.")
		return
	}
	displaySearchResults(LastSearch)
}

func displaySearchResults(r *spotify.SearchResult){
	if r.Tracks != nil {
		fmt.Println("Tracks: ")
		t := template.New("shortTrackTemplate")
		t, err := t.Parse(shortTrackTemplate)
		checkErr(err)
		for i, v := range r.Tracks.Tracks{
			if i > 4 {
				break
			}
			fmt.Printf("  [%d]:\t", i+1)
			err = t.Execute(os.Stdout, v)
			checkErr(err)
		}
	}
	if r.Artists != nil {
		fmt.Println("Artists: ")
		for i, v := range r.Artists.Artists {
			if i > 4 {
				break
			}
			fmt.Printf("  [%d]:\t%v\n", i+1, v.Name)
		}
	}
	if r.Albums != nil {
		fmt.Println("Albums: ")
		for i, v := range r.Albums.Albums {
			if i > 4 {
				break
			}
			fmt.Printf("  [%d]:\t%v\n", i+1, v.Name)
		}
	}
	if r.Playlists != nil {
		fmt.Println("Playlists: ")
		for i, v := range r.Playlists.Playlists {
			if i > 4 {
				break
			}
			fmt.Printf("  [%d]:\t%v\n", i+1, v.Name)
		}
	}

}
func repeatAction(c *cli.Context) {
	if c.NArg() < 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	switch c.Args().First() {
	case "track":
		setRepeat("track")
	case "playlist":
		setRepeat("context")
	case "off":
		setRepeat("off")
	default:
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	time.Sleep(250 * time.Millisecond)
	client := auth.NewClient(tok)
	s, err := client.PlayerState()
	checkErr(err)
	if s.RepeatState == "context" {
		fmt.Println("Repeat: ", "playlist")
	} else {
		fmt.Println("Repeat: ", s.RepeatState)
	}

}

func setRepeat(s string) {
	client := auth.NewClient(tok)
	err := client.Repeat(s)
	checkErr(err)
}

func shuffleAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	checkErr(err)
	if state.ShuffleState {
		err = client.Shuffle(false)
		checkErr(err)
	} else {
		err = client.Shuffle(true)
		checkErr(err)
	}
	time.Sleep(250 * time.Millisecond)
	fmt.Println("Shuffle: ", getShuffleState())
}

func getShuffleState() bool {
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	checkErr(err)
	return state.ShuffleState
}

func optionsAction(c *cli.Context) {
	if c.NArg() > 0 {
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
	if c.NArg() > 0 {
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
	fmt.Printf("Volume: %v%%\n", getVolume())
}

func getCurrentTrack() *spotify.FullTrack {
	client := auth.NewClient(tok)
	current, err := client.PlayerCurrentlyPlaying()
	checkErr(err)
	return current.Item
}

func nextAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Next()
	checkErr(err)
}

func prevAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Previous()
	checkErr(err)
}

func volUpAction(c *cli.Context) {
	if c.NArg() > 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	v := getVolume()
	i := v + 10
	if i > 100 {
		i = 100
	}
	setVolume(i)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Volume: %v%%\n", getVolume())
}

func volDownAction(c *cli.Context) {
	if c.NArg() > 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	v := getVolume()
	i := v - 10
	if i < 0 {
		i = 0
	}
	setVolume(i)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Volume: %v%%\n", getVolume())
}

func volSetAction(c *cli.Context) {
	if c.NArg() != 2 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	i, err := strconv.Atoi(c.Args().Get(1))
	checkErr(err)
	if i > 100 || i < 0 {
		fmt.Println("ERROR: Invalid argument, ", i)
		return
	}
	setVolume(i)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Volume: %v%%\n", getVolume())
}

func setVolume(i int) {
	client := auth.NewClient(tok)
	err := client.Volume(i)
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
		panic("Error: no devices are active, please begin playback on a Spotify Conneceted device first")
	}
	return a
}

func playAction(c *cli.Context) {
	client := auth.NewClient(tok)
	if c.Args().First() == "" {
		err := client.Play()
		checkErr(err)
		return
	}
	i, err := strconv.Atoi(c.Args().First())
	checkErr(err)

	d, err := client.PlayerDevices()
	checkErr(err)
	if i > len(d) {
		fmt.Println("ERROR: Incorrect device ID, ", i)
		err = cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}

	ID := d[i-1].ID
	err = client.TransferPlayback(ID, true)
	checkErr(err)
}

func pauseAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	err := client.Pause()
	checkErr(err)
}

func devicesAction(c *cli.Context) {
	if c.NArg() > 0 {
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
