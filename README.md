# Spotcon

Spotcon is a WIP command line interface, written in Go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Usage

Syntax: spotcon> command [subcommand] [--flags] [arguments...]

```
NAME:
   Spotcon - Control Spotify Connect enabled devices via terminal.
USAGE:
   spotify> command [subcommand] [--flags] [arguments...]
AUTHOR:
   Luke Hobbs <lukeehobbs@gmail.com>
COMMANDS:
     clear, clc  Clear the command window
     current, c  Display information about the current playback
     devices, d  List available devices
     next, n     Skip to the next track in queue
     opt, o      Options for changing current playback parameters
     pause, pp   Pause playback
     play, p     Start/Resume playback
     prev, pr    Skip to the previous track in queue
     quit, q     Quit application
     search, s   Search Spotify for artists, albums, tracks, or playlists
     vol, v      Options for changing volume of playback
     help, h     Shows a list of commands or help for one command
GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Commands

```

```

## Todo

- Move repeat and shuffle commands to be flags of options command
- Configure command to remove the need for environment variables
- Search should be prioritized to search user's saved library first
- Seek options (fast forward, rewind)
- ~~Add an option to clear the console~~
- ~~Allow selection of devices using a search~~
- ~~Change options (shuffle/repeat)~~
- ~~Control device selection~~
- ~~Control media playback (pause/play, next/previous song)~~
- ~~Control volume of media playback~~
- ~~Get information about currently playing song~~
- ~~Login to Spotify through WebAPI~~
- ~~Search for a playlist to listen to~~
- ~~Search for a song to listen to~~
- ~~Search for an album to listen to~~
- ~~Skip track (next/previous song)~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~

## Stretch Goals

- Daemon to monitor and report currently playing media in plain text for scripting purposes (polybar, conky, etc.)

## Example

```
$ spotcon
You are logged in as: lukehobbs
spotcon> devices
Devices:
  [1]: Samsung (TV)
  [2]: Desktop (Computer) ACTIVE
  [3]: Amazon Echo (Speaker)

spotcon> play --device echo --track shut up blink

spotcon> devices
Devices:
  [1]: Samsung (TV)
  [2]: Desktop (Computer)
  [3]: Amazon Echo (Speaker) ACTIVE

spotcon> vol 80
Volume: 80%

spotcon> vol up
Volume: 90%

spotcon> current
Track:  Shut Up
Artist:	blink-182
Album:	Take Off Your Pants And Jacket
Volume: 90%

spotcon> options
Shuffle: On
Repeat:  Off

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
  [1]:	"Bridge Over Troubled Water" by Simon & Garfunkel
  [2]:	"Water Under the Bridge" by Adele
  [3]:	"One Day Remains" by Alter Bridge
  [4]:	"The Last Hero" by Alter Bridge
  [5]:	"Blackbird" by Alter Bridge
Playlists: 
  [1]:	"Alter Bridge Complete Collection" - officialalterbridge
  [2]:	"THE BRIDGE" - 1221493509
  [3]:	"bridge" - 11101296551
  [4]:	"Alter Bridge" - chemistry11
  [5]:	"Bridge Anytime" - 1259523134

spotcon> play --track 2

spotcon> current
Track:  Under The Bridge
Artist:	Red Hot Chili Peppers
Album:	Blood Sugar Sex Magik (Deluxe Version)
Volume: 90%
```


## Acknowledgements

Thanks to Zac Bergquist for the [Go wrapper for the Spotify Web API](https://github.com/zmb3/spotify) that made this project possible!
