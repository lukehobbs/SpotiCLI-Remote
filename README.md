# Spotcon

Spotcon is a WIP command line interface, written in Go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Todo

- ~~Control media playback (pause/play, next/previous song)~~
- ~~Login to Spotify through WebAPI~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~
- ~~Control volume of media playback~~
- ~~Get information about currently playing song~~
- ~~Control device selection~~
- ~~Skip track (next/previous song)~~
- ~~Add an option to clear the console~~
- ~~Change options (shuffle/repeat)~~
- Seek options (fast forward, rewind)
- Allow selection of devices using a search
  - i.e. `play Amazon Echo`
- ~~Search for a playlist to listen to~~
- ~~Search for an album to listen to~~
- ~~Search for a song to listen to~~
  - Play track/album/artist/playlist from most recent search
  - i.e. `play -t 2`

## Stretch Goals

- Daemon to monitor and report currently playing media in plain text for scripting purposes (polybar, conky, etc.)

## Usage

Syntax: spotcon> command [subcommand] [arguments...]

```
spotcon> devices                        List devices available for playback.

spotcon> play                           Start/Resume playback on current device.
spotcon> play <device_number>           Start/Resume playback on specified device.
spotcon> pause                          Pause playback on current device.

spotcon> search <query>                 Search Spotify for <query>
spotcon> search --track <query>         Search Spotify tracks for <query>
spotcon> search --album <query>         Search Spotify albums for <query>
spotcon> search --artist <query>        Search Spotify artists for <query>
spotcon> search --playlist <query>      Search Spotify playlists for <query>

spotcon> vol                            Show the current volume.
spotcon> vol set <percent>              Set the volume to an amount between 0 and 100.
spotcon> vol up                         Increase the volume by 10%.
spotcon> vol down                       Decrease the volume by 10%.

spotcon> next                           Skip to the next song in playlist.
spotcon> prev                           Return to the previous song in playlist.
spotcon> current                        Show information about the currently playing song.

spotcon> options                        Display current state of playback options (Shuffle, Repeat, Volume)
spotcon> shuffle                        Toggle shuffle playback state
spotcon> repeat track                   Turn on track repeat playback
spotcon> repeat playlist                Turn on playlist repeat playback
spotcon> repeat off                     Turn repeat playback off

spotcon> clear                          Clear the command window.
spotcon> help                           Show help.
spotcon> quit                           Quit application.
```

## Example

```
$ spotcon
You are logged in as: lukehobbs
spotcon> devices
[1]=Samsung (TV)
[2]=Desktop (Computer) ACTIVE
[3]=Amazon Echo (Speaker)

spotcon> play 3

spotcon> devices
[1]=Samsung (TV)
[2]=Desktop (Computer)
[3]=Amazon Echo (Speaker) ACTIVE

spotcon> vol set 80
spotcon> vol
Volume: 80%

spotcon> vol up
spotcon> vol
Volume: 90%

spotcon> current
Track:  Shut Up
Artist:	blink-182
Album:	Take Off Your Pants And Jacket

spotcon> options
Shuffle: On
Repeat:  Off
Volume:  90%

spotcon> repeat track
Repeat:  track

spotcon> shuffle
Shuffle: Off

spotcon> search bridge
Tracks: 
  [1]:	"Water Under the Bridge" by Adele
  [2]:	"Under The Bridge" by Red Hot Chili Peppers
  [3]:	"Water Under the Bridge" by Adele
  [4]:	"London Bridge" by Fergie
  [5]:	"Under The Bridge" by Red Hot Chili Peppers
Artists: 
  [1]:	Alter Bridge
  [2]:	Bridge to Grace
  [3]:	Bridge
  [4]:	Marcus Bridge
  [5]:	The Bridge
Albums: 
  [1]:	Bridge Over Troubled Water
  [2]:	Water Under the Bridge
  [3]:	One Day Remains
  [4]:	The Last Hero
  [5]:	Blackbird
Playlists: 
  [1]:	Alter Bridge Complete Collection
  [2]:	THE BRIDGE
  [3]:	bridge
  [4]:	Alter Bridge
  [5]:	Bridge Anytime

spotcon> play -t 2

spotcon> current
Track:  Under The Bridge
Artist:	Red Hot Chili Peppers
Album:	Blood Sugar Sex Magik (Deluxe Version)
Volume: 100%
```


## Acknowledgements

Thanks to Zac Bergquist for the [Go wrapper for the Spotify Web API](https://github.com/zmb3/spotify) that made this project possible!
