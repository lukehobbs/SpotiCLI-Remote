# SpotifyCNCT - Spotify Connect CLI Remote

SpotifyCNCT will be a command line interface, written in go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com). 

## Goals

- Control media playback (pause/play, next/previous song)
- Control volume of media playback
- Change options (shuffle/repeat)
- Seek options (fast forward, rewind)
- Control device selection

# Commands

`devices` - list available devices

`devices connect $DEVICE` - connect to a device for playback


`play` - start/resume playback
 
`pause` - pause playback

`next` - skip playback to next track

`prev` - skip playback to previous track

`ff $VALUE` - seek forward to value in track or default if none provided

`rw $VALUE` - seek backward to value in track or default if none provided

`vol (up/down)` - increase/decrease volume of playback

`now` - get information about currently playing track

`set repeat (on/off)` - toggle repeat mode

`set shuffle (on/off)` - toggle shuffle mode


## Stretch Goals

- Daemon to monitor currently playing media in plain text for scripting purposes (polybar, conky, etc.)

