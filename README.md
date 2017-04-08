# SpotifyCNCT - Spotify Connect CLI Remote

SpotifyCNCT is a WIP command line interface, written in Go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Todo

- ~~Control media playback (pause/play, next/previous song)~~
- ~~Login to Spotify through WebAPI~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~
- ~~Control volume of media playback~~
- Get information about currently playing song
- Allow selection of devices using a search
  - i.e. `play -d Amazon Echo`
- Change options (shuffle/repeat)
- Seek options (fast forward, rewind)
- ~~Control device selection~~
- Add an option to clear the console

## Stretch Goals

- Daemon to monitor currently playing media in plain text for scripting purposes (polybar, conky, etc.)

## Currently Implemented Commands

Syntax: spotify> [command] [--flags] [values...]

`devices`, `d` - list available devices and their IDs

`play [--device]`, `p [-d]` - start/resume playback on specified device or active device in none specified

`pause`, `pp` - pause playback on active device

`vol [--set, --up, --down]`, `v [-s, -u, -d]` - control volume on active device


## Example Usage

```
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
```
