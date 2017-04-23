package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/urfave/cli"
	"github.com/zmb3/spotify"
)

// LastSearch is the results of the last search query
var LastSearch *spotify.SearchResult

const (
	track  = "track"
	album  = "album"
	artist = "artist"
	plist  = "playlist"
)

func searchAction(c *cli.Context) {
	var t int
	client := auth.NewClient(tok)
	q := strings.Join(c.Args(), " ")
	if q == "" {
		displayLastSearch()
		return
	}
	if c.Bool(album) {
		t++
	}
	if c.Bool(artist) {
		t += 2
	}
	if c.Bool(plist) {
		t += 4
	}
	if c.Bool(track) {
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

func displaySearchResults(r *spotify.SearchResult) {
	if len(r.Tracks.Tracks) > 0 {
		displayTracks(r.Tracks.Tracks)
	}
	if len(r.Artists.Artists) > 0 {
		displayArtists(r.Artists.Artists)
	}
	if len(r.Albums.Albums) > 0 {
		displayAlbums(r.Albums.Albums)
	}
	if len(r.Playlists.Playlists) > 0 {
		displayPlaylists(r.Playlists.Playlists)
	}
}

func displayTracks(r []spotify.FullTrack) {
	fmt.Println("Tracks: ")
	t := template.New("shortTrackTemplate")
	t, err := t.Parse(shortTrackTemplate)
	checkErr(err)
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		fmt.Printf("  [%d]:\t", i+1)
		err = t.Execute(os.Stdout, v)
		checkErr(err)
	}
}

func displayArtists(r []spotify.FullArtist) {
	fmt.Println("Artists: ")
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		fmt.Printf("  [%d]:\t%v\n", i+1, v.Name)
	}
}

func displayAlbums(r []spotify.SimpleAlbum) {
	client := auth.NewClient(tok)
	fmt.Println("Albums: ")
	t := template.New("shortAlbumTemplate")
	t, err := t.Parse(shortAlbumTemplate)
	checkErr(err)
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		al, err := client.GetAlbum(v.ID)
		checkErr(err)
		fmt.Printf("  [%d]:\t", i+1)
		err = t.Execute(os.Stdout, al)
		checkErr(err)
	}
}

func displayPlaylists(r []spotify.SimplePlaylist) {
	fmt.Println("Playlists: ")
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		fmt.Printf("  [%d]:\t\"%v\" - %s\n", i+1, v.Name, v.Owner.ID)
	}
}

func repeatAction(c *cli.Context) {
	if c.NArg() < 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	switch c.Args().First() {
	case track:
		setRepeat("track")
	case plist:
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

func skipAction(c *cli.Context, b bool) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	if b {
		err := client.Next()
		checkErr(err)
		return
	}
	err := client.Previous()
	checkErr(err)
	return
}

func volAdjust(b bool) {
	v := getVolume()
	switch b {
	case true:
		if v >= 90 {
			setVolume(100)
			return
		}
		setVolume(v + 10)
		return
	case false:
		if v <= 10 {
			setVolume(0)
			return
		}
		setVolume(v - 10)
		return
	}
}

func volAction(c *cli.Context, b bool) {
	if c.NArg() > 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	volAdjust(b)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Volume: %v%%\n", getVolume())
}

func volSetAction(c *cli.Context) {
	if c.NArg() != 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	i, err := strconv.Atoi(c.Args().First())
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
		panic("Error: no devices are active, please begin playback.")
	}
	return a
}

func playAction(c *cli.Context) {
	if c.NumFlags() > 4 {
		fmt.Println("ERROR: Too many flags set.")
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	if c.NumFlags() == 0 {
		err := client.Play()
		checkErr(err)
		return
	}
	if c.IsSet("device") {
		e := setDevice(c.String("device"))
		if !e {
			err := cli.ShowCommandHelp(c, c.Command.Name)
			checkErr(err)
			return
		}
		if c.NumFlags() == 1 { // Device is the only flag set.
			fmt.Println("Device is only flag. Playing...")
			err := client.Play()
			checkErr(err)
			return
		}
	}
	if c.IsSet(track) {
		play(c.Args(), track)
		return
	}
	if c.IsSet(album) {
		play(c.Args(), album)
	}
	if c.IsSet(artist) {
		play(c.Args(), artist)
	}
	if c.IsSet(plist) {
		play(c.Args(), plist)
	}
}

func setDevice(s string) bool {
	client := auth.NewClient(tok)
	d, err := client.PlayerDevices()
	checkErr(err)
	var xi int
	if xi, err = strconv.Atoi(s); err == nil {
		if xi > 0 && xi <= len(d) {
			err = client.Pause() // Pause playback before transfer.
			checkErr(err)
			err = client.TransferPlayback(d[xi-1].ID, false)
			checkErr(err)
			return true
		}
		fmt.Println("ERROR: Incorrect device ID, ", s)
		return false
	}

	for _, v := range d {
		if strings.Contains(strings.ToLower(v.Name), strings.ToLower(s)) {
			err = client.Pause() // Pause playback before transfer.
			checkErr(err)
			err = client.TransferPlayback(v.ID, false)
			checkErr(err)
			return true
		}
	}

	fmt.Println("ERROR: Could not connect to device, ", s)
	checkErr(err)
	return false
}

func play(s []string, t string) {
	client := auth.NewClient(tok)
	a := strings.Join(s, " ")
	if i, err := strconv.Atoi(a); err == nil {
		switch t {
		case track:
			t := getInterfaceSlice(LastSearch.Tracks.Tracks)
			playNum(i, t)
		case artist:
			t := getInterfaceSlice(LastSearch.Artists.Artists)
			playNum(i, t)
		case album:
			t := getInterfaceSlice(LastSearch.Albums.Albums)
			playNum(i, t)
		case plist:
			t := getInterfaceSlice(LastSearch.Playlists.Playlists)
			playNum(i, t)
		}
		return
	}
	u := luckySearch(a, t)
	if t == track && u != "" {
		o := spotify.PlayOptions{URIs: []spotify.URI{u}}
		err := client.PlayOpt(&o)
		checkErr(err)
		return
	}
	if u != "" {
		o := spotify.PlayOptions{PlaybackContext: &u}
		err := client.PlayOpt(&o)
		checkErr(err)
		return
	}
}

func playNum(i int, t []interface{}) {
	client := auth.NewClient(tok)
	if len(t) == 0 {
		fmt.Println("No search results found.")
		return
	}
	if i > 0 && i < len(t) {
		u := getURI(t[i-1])
		if _, ok := t[0].(spotify.FullTrack); ok {
			o := spotify.PlayOptions{URIs: []spotify.URI{u}}
			err := client.PlayOpt(&o)
			checkErr(err)
			return
		}
		o := spotify.PlayOptions{PlaybackContext: &u}
		err := client.PlayOpt(&o)
		checkErr(err)
		return
	}
}

func getURI(r interface{}) spotify.URI {
	switch r := r.(type) {
	case spotify.FullTrack:
		return r.URI
	case spotify.SimplePlaylist:
		return r.URI
	case spotify.SimpleAlbum:
		return r.URI
	case spotify.FullArtist:
		return r.URI
	}
	return ""
}

func getInterfaceSlice(r interface{}) []interface{} {
	switch r := r.(type) {
	case []spotify.FullTrack:
		var interfaceSlice = make([]interface{}, len(r))
		for i, d := range r {
			interfaceSlice[i] = d
		}
		return interfaceSlice
	case []spotify.FullArtist:
		var interfaceSlice = make([]interface{}, len(r))
		for i, d := range r {
			interfaceSlice[i] = d
		}
		return interfaceSlice
	case []spotify.SimpleAlbum:
		var interfaceSlice = make([]interface{}, len(r))
		for i, d := range r {
			interfaceSlice[i] = d
		}
		return interfaceSlice
	case []spotify.SimplePlaylist:
		var interfaceSlice = make([]interface{}, len(r))
		for i, d := range r {
			interfaceSlice[i] = d
		}
		return interfaceSlice
	}
	return nil
}

func luckySearch(s string, t string) spotify.URI {
	client := auth.NewClient(tok)
	switch t {
	case track:
		r, err := client.Search(s, spotify.SearchType(8))
		checkErr(err)
		tr := r.Tracks.Tracks
		if len(tr) == 0 {
			fmt.Printf("No %ss found matching: %s", t, s)
			return ""
		}
		return tr[0].URI
	case artist:
		r, err := client.Search(s, spotify.SearchType(2))
		checkErr(err)
		ar := r.Artists.Artists
		if len(ar) == 0 {
			fmt.Printf("No %ss found matching: %s", t, s)
			return ""
		}
		return ar[0].URI
	case album:
		r, err := client.Search(s, spotify.SearchType(1))
		checkErr(err)
		al := r.Albums.Albums
		if len(al) == 0 {
			fmt.Printf("No %ss found matching: %s", t, s)
			return ""
		}
		return al[0].URI
	case plist:
		r, err := client.Search(s, spotify.SearchType(4))
		checkErr(err)
		pl := r.Playlists.Playlists
		if len(pl) == 0 {
			fmt.Printf("No %ss found matching: %s", t, s)
			return ""
		}
		return pl[0].URI
	default:
		return ""
	}
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
	fmt.Println("Devices:")
	for i, v := range d {
		fmt.Printf("  [%d]: %v (%v)", i+1, v.Name, v.Type)
		if v.Active {
			fmt.Println(" ACTIVE")
		} else {
			fmt.Println()
		}
	}
}
