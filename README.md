# SpotifyCNCT - Spotify Connect CLI Remote

SpotifyCNCT will be a command line interface, written in go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Todo

- ~~Control media playback (pause/play, next/previous song)~~
- ~~Login to Spotify through WebAPI~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~
- Control volume of media playback
- Change options (shuffle/repeat)
- Seek options (fast forward, rewind)
- ~~Control device selection~~

## Stretch Goals

- Daemon to monitor currently playing media in plain text for scripting purposes (polybar, conky, etc.)

# Currently Implemented Commands

Syntax: spotify> [command] [arguments..]

`devices`, `d` - list available devices and their IDs

`play DEVICE_ID`, `p DEVICE_ID` - start/resume playback on specified device or currently playing device in none specified

`pause` - pause playback on currently playing device
