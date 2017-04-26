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

// checkSaved looks for the specified string in the user's saved library
// t is the type of s and can be any of (artist, album, playlist, track)
// Returns the URI of s if found or "" if not found
func checkSaved(s string, t string) spotify.URI {
	s = strings.ToLower(s)
	switch t {
	case track:
		for _, v := range getSavedTracks() {
			if s == strings.ToLower(v.Name) {
				return v.URI
			}
		}
	case album:
		for _, v := range getSavedAlbums() {
			if s == strings.ToLower(v.Name) {
				return v.URI
			}
		}
	case artist:
		for _, v := range getSavedArtists() {
			if s == strings.ToLower(v.Name) {
				return v.URI
			}
		}
	case plist:
		for _, v := range getSavedPlaylists() {
			if s == strings.ToLower(v.Name) {
				return v.URI
			}
		}
	}
	return ""
}

// devicesAction is called with spotcon> devices
// Lists the user's Spotify Connected devices
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

// luckySearch searches Spotify for specified string
// t is the type of s and can be any of (artist, album, playlist, track)
// Returns the first result matching the string specified
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

// optAction is called with spotcon> opt
// Used to set options: (repeat, shuffle) to (on, off)
func optAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	if c.String("repeat") != "" {
		switch c.String("repeat") {
		case "on":
			setRepeat("context")
		case "off":
			setRepeat("off")
		default:
			err := cli.ShowCommandHelp(c, c.Command.Name)
			checkErr(err)
			return
		}
	}
	if c.String("shuffle") != "" {
		switch c.String("shuffle") {
		case "on":
			setShuffle(true)
		case "off":
			setShuffle(false)
		}
	}
	time.Sleep(200 * time.Millisecond)
	displayOpts()
}

// nowAction is called with spotcon> now
// Displays information about Now Playing
func nowAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	tr := getCurrentTrack()
	t := template.New("longTrackTemplate")
	t, err := t.Parse(longTrackTemplate)
	checkErr(err)

	fmt.Println("Device:", getActiveDeviceName())
	err = t.Execute(os.Stdout, tr)
	checkErr(err)
	displayVolume()
	displayProgress()
}

// pauseAction is called with spotcon> pause
// Pauses the current playback
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

// play begins playback
//      - t determines the type which is one of (artist, album, playlist, track)
//      - if s is a number, playNum() is called to handle playback
//      - if s is a string, the user's saved tracks are searched for matches and
//        no matches are found, the first result from a search is played
func play(s string, t string) {
	var u spotify.URI
	client := auth.NewClient(tok)
	if i, err := strconv.Atoi(s); err == nil {
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
	if a := checkSaved(s, t); a != "" {
		u = a
	} else {
		u = luckySearch(s, t)
	}
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

// playAction is called with spotcon> play
// Start/Resumes playback and handles flags
func playAction(c *cli.Context) {
	if c.NArg() > 0 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	if c.NumFlags() > 2 {
		if !c.IsSet("device") {
			fmt.Println("ERROR: Too many flags set.")
			err := cli.ShowCommandHelp(c, c.Command.Name)
			checkErr(err)
			return
		}
		if c.NumFlags() > 4 {
			fmt.Println("ERROR: Too many flags set.")
			err := cli.ShowCommandHelp(c, c.Command.Name)
			checkErr(err)
			return
		}
	}
	client := auth.NewClient(tok)
	if c.IsSet("device") {
		e := setDevice(c.String("device"))
		if !e {
			err := cli.ShowCommandHelp(c, c.Command.Name)
			checkErr(err)
			return
		}
		if c.NumFlags() == 2 { // Device is the only flag set.
			err := client.Play()
			checkErr(err)
			return
		}
	}
	if c.IsSet(track) {
		play(c.String("track"), track)
		return
	}
	if c.IsSet(album) {
		play(c.String("album"), album)
		return
	}
	if c.IsSet(artist) {
		play(c.String("artist"), artist)
		return
	}
	if c.IsSet(plist) {
		play(c.String("plist"), plist)
	}
	err := client.Play()
	checkErr(err)
}

// playNum plays an item from LastSearch by referencing its number
// found with searchAction()
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

// searchAction is called with spotcon> search
// Preforms a Spotify search with the specified flags
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

// seekAction is called with spotcon> seek
// Seeks forwards if b is true and backwards if b is false
func seekAction(c *cli.Context, b bool) {
	var err error
	t := 15 * 1000
	_ = t
	if c.NumFlags() > 2 {
		fmt.Println("ERROR: Cannot seek forward and backwards")
		err = cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	client := auth.NewClient(tok)
	p, err := client.PlayerCurrentlyPlaying()
	checkErr(err)
	pr := p.Progress
	d := p.Item.Duration
	if c.Args().First() != "" {
		t, err = strconv.Atoi(c.Args().First())
		checkErr(err)
		t = t * 1000
	}
	if b {
		if pr+t > d {
			t = d - pr
		}
		err = client.Seek(pr + t)
		checkErr(err)
	} else {
		if pr-t < 0 {
			t = pr
		}
		err = client.Seek(pr - t)
		checkErr(err)
	}
	time.Sleep(150 * time.Millisecond)
	displayProgress()
}

// skipAction is called with either spotcon> next or spotcon> prev
// Playback skips forward if b is true or backwards if b is false
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
	} else {
		err := client.Previous()
		checkErr(err)
	}
	err := client.Play()
	checkErr(err)
	time.Sleep(200 * time.Millisecond)
	nowAction(c)
	return
}

// volAdjustAction is called by spotcon> vol (up/down)
// Increments volume by 10% if percent is not specified
func volAdjustAction(c *cli.Context, b bool) {
	p := 10
	if c.Args().First() != "" {
		var err error
		p, err = strconv.Atoi(c.Args().First())
		checkErr(err)
	}
	v := getVolume()
	switch b {
	case true:
		if v+p >= 100 {
			setVolume(100)
			break
		}
		setVolume(v + p)
	case false:
		if v-p <= 0 {
			setVolume(0)
			break
		}
		setVolume(v - p)
	}
	time.Sleep(150 * time.Millisecond)
	displayVolume()
}

// volSetAction is called with spotcon> vol set
// Sets volume to a specified percent
func volSetAction(c *cli.Context) {
	if c.NArg() != 1 {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		checkErr(err)
		return
	}
	i, err := strconv.Atoi(c.Args().First())
	checkErr(err)
	if i > 100 {
		i = 100
	}
	setVolume(i)
	time.Sleep(150 * time.Millisecond)
	displayVolume()
}

// displayFullTracks prints a shortTrackTemplate of each of the tracks in
// a []spotify.FullTrack
func displayFullTracks(r []spotify.FullTrack) {
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

// func displayFullArtists
func displayFullArtists(r []spotify.FullArtist) {
	fmt.Println("Artists: ")
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		fmt.Printf("  [%d]:\t%v\n", i+1, v.Name)
	}
}

// func displayLastSearch prints the results of the last search query
func displayLastSearch() {
	if LastSearch == nil {
		fmt.Println("No previous search results found.")
		return
	}
	displaySearchResults(LastSearch)
}

// displayOpts prints the current values of shuffle and repeat
// using the optionsTemplate
func displayOpts() {
	client := auth.NewClient(tok)
	state, err := client.PlayerState()
	checkErr(err)
	t := template.New("optionsTemplate")
	t, err = t.Parse(optionsTemplate)
	checkErr(err)
	err = t.Execute(os.Stdout, state)
	checkErr(err)
}

// displayProgress prints the current playback progress
func displayProgress() {
	client := auth.NewClient(tok)
	p, err := client.PlayerCurrentlyPlaying()
	checkErr(err)
	pr := p.Progress / 1000
	t := p.Item.Duration / 1000
	fmt.Printf("[%d:%02d/%d:%02d]\n", pr/60, pr%60, t/60, t%60)
}

// displaySearchResults is a helper function that calls the correct display
// functions to print out all the search results
func displaySearchResults(r *spotify.SearchResult) {
	if len(r.Tracks.Tracks) > 0 {
		displayFullTracks(r.Tracks.Tracks)
	}
	if len(r.Artists.Artists) > 0 {
		displayFullArtists(r.Artists.Artists)
	}
	if len(r.Albums.Albums) > 0 {
		displaySimpleAlbums(r.Albums.Albums)
	}
	if len(r.Playlists.Playlists) > 0 {
		displaySimplePlaylists(r.Playlists.Playlists)
	}
}

// displaySimpleAlbums prints a shortAlbumTemplate of each of the albums
// in a []spotify.SimpleAlbum
func displaySimpleAlbums(r []spotify.SimpleAlbum) {
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

// displaySimplePlaylists prints the names and owner IDs of all the playlists
// in a []spotify.SimplePlaylist
func displaySimplePlaylists(r []spotify.SimplePlaylist) {
	fmt.Println("Playlists: ")
	for i := 0; i < 5 && i < len(r); i++ {
		v := r[i]
		fmt.Printf("  [%d]:\t\"%v\" - %s\n", i+1, v.Name, v.Owner.ID)
	}
}

// displayVolume prints the current volume level as a percent
func displayVolume() {
	v := getVolume()
	if v == -1 {
		return
	}
	fmt.Printf("Volume: %v%%\n", v)
}

// getActiveDeviceName returns the name of the actively playing device
// Or "No devices active" if none are active
func getActiveDeviceName() string {
	client := auth.NewClient(tok)
	d, err := client.PlayerDevices()
	checkErr(err)
	for _, v := range d {
		if v.Active {
			return v.Name
		}
	}
	return "No devices active"
}

// getCurrentTrack returns a pointer to the currently playing track
func getCurrentTrack() *spotify.FullTrack {
	client := auth.NewClient(tok)
	current, err := client.PlayerCurrentlyPlaying()
	checkErr(err)
	return current.Item
}

// getInterfaceSlice takes one of ([]spotify.FullTrack, []spotify.FullArtist,
//                                 []spotify.SimpleAlbum, []spotify.SimplePlaylist)
// and returns an interface slice version
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

// getSavedAlbums returns the first 50 of the user's saved artists
func getSavedAlbums() []spotify.SavedAlbum {
	i := 50
	o := spotify.Options{Limit: &i}
	client := auth.NewClient(tok)
	s, err := client.CurrentUsersAlbumsOpt(&o)
	checkErr(err)
	sa := s.Albums
	return sa
}

// getSavedArtists returns the first 50 of the user's saved artists
func getSavedArtists() []spotify.FullArtist {
	client := auth.NewClient(tok)
	s, err := client.CurrentUsersFollowedArtistsOpt(50, "")
	checkErr(err)
	sa := s.Artists
	return sa
}

// getSavedPlaylists returns the first 50 of the user's saved playlists
func getSavedPlaylists() []spotify.SimplePlaylist {
	i := 50
	o := spotify.Options{Limit: &i}
	client := auth.NewClient(tok)
	s, err := client.CurrentUsersPlaylistsOpt(&o)
	checkErr(err)
	sa := s.Playlists
	return sa
}

// getSavedTracks returns the first 50 of the user's saved tracks
func getSavedTracks() []spotify.SavedTrack {
	i := 50
	o := spotify.Options{Limit: &i}
	client := auth.NewClient(tok)
	s, err := client.CurrentUsersTracksOpt(&o)
	checkErr(err)
	sa := s.Tracks
	return sa
}

// getURI accesses the URI property of the interface
// Input must be one of [spotify.FullTrack, spotify.SimplePlaylist,
//                       spotify.SimpleAlbum, spotify.FullArtist]
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

// getVolume retrieves the current volume level
// Returns an integer between 0 and 100
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
		//fmt.Println("Error: no devices are active, please begin playback.")
		return -1
	}
	return a
}

// setDevice transfers playback to a new device
// Either takes the name of a device as input or the number
// displayed from devicesAction()
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
	return false
}

// setRepeat sets repeat option to one of [on, off]
func setRepeat(s string) {
	client := auth.NewClient(tok)
	err := client.Repeat(s)
	checkErr(err)
}

// setShuffle sets shuffle option to one of [on, off]
func setShuffle(b bool) {
	client := auth.NewClient(tok)
	err := client.Shuffle(b)
	checkErr(err)
}

// setVolume sets volume to a percent
// 0 < i < 100
func setVolume(i int) {
	client := auth.NewClient(tok)
	err := client.Volume(i)
	checkErr(err)
}
