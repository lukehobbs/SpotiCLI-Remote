# SpotiCLI Remote

SpotiCLI Remote is a WIP command line interface, written in Go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Todo

- ~~Control media playback (pause/play, next/previous song)~~
- ~~Login to Spotify through WebAPI~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~
- ~~Control volume of media playback~~
- Get information about currently playing song
- Allow selection of devices using a search
  - i.e. `play --device Amazon Echo`
- Change options (shuffle/repeat)
- Seek options (fast forward, rewind)
- ~~Control device selection~~
- Add an option to clear the console
- ~~Skip track (next/previous song)~~

## Stretch Goals

- Daemon to monitor currently playing media in plain text for scripting purposes (polybar, conky, etc.)
- Select songs using a fuzzy search (i.e. spotify> play --song Anthem Part Two)

## Currently Implemented Commands

Syntax: spotify> [command] [--flags] [values...]

`devices`, `d` - list available devices and their IDs

`play [--device]`, `p [-d]` - start/resume playback on specified device or active device in none specified

`pause`, `pp` - pause playback on active device

`vol [--set, --up, --down]`, `v [-s, -u, -d]` - control volume on active device

`next`, `n` - skip to next track

`prev`, `pr` - skip to previous track

`current`, `c` - display information about the currently playing track


## Example Usage

```
$ spotifycli
You are logged in as: lukehobbs
spotify> devices
[1]=Samsung (TV)
[2]=Desktop (Computer) ACTIVE
[3]=Amazon Echo (Speaker)

spotify> play --device 3

spotify> devices
[1]=Samsung (TV)
[2]=Desktop (Computer)
[3]=Amazon Echo (Speaker) ACTIVE

spotify> vol --set 80
spotify> vol
Volume: 80

spotify> vol --up
spotify> vol
Volume: 90

spotify> current
Track:  Shut Up
Artist:	blink-182
Album:	Take Off Your Pants And Jacket
```


## Acknowlegments

Thanks to Zac Bergquist for the [Go wrapper for the Spotify Web API](https://github.com/zmb3/spotify) that made this project possible!
