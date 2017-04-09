# SpotiCLI Remote

SpotiCLI Remote is a WIP command line interface, written in Go, that controls media playback over spotify connected devices using the [Spotify Web API](https://api.spotify.com).

## Todo

- ~~Control media playback (pause/play, next/previous song)~~
- ~~Login to Spotify through WebAPI~~
- ~~Store OAuth 2 Token in user's home directory to avoid having to login every run~~
- ~~Control volume of media playback~~
- ~~Get information about currently playing song~~
- ~~Control device selection~~
- ~~Skip track (next/previous song)~~
- ~~Add an option to clear the console~~
- Change options (shuffle/repeat)
- Seek options (fast forward, rewind)
- Allow selection of devices using a search
  - i.e. `play --device Amazon Echo`
- Search for a playlist to listen to
- Search for an ablum to listen to
- Search for a song to listen to

## Stretch Goals

- Daemon to monitor and report currently playing media in plain text for scripting purposes (polybar, conky, etc.)

## Currently Implemented Commands

Syntax: spotify> [command] [--flags] [values...]

```
spotify> devices                        List devices available for playback.

spotify> play                           Start/Resume playback on current device.
spotify> play --device <device_number>  Start/Resume playback on specified device.
spotify> pause                          Pause playback on current device.

spotify> vol                            Show the current volume.
spotify> vol --set <percent>            Set the volume to an amount between 0 and 100.
spotify> vol --up                       Increase the volume by 10%.
spotify> vol --down                     Decrease the volume by 10%.

spotify> next                           Skip to the next song in playlist.
spotify> prev                           Return to the previous song in playlist.
spotify> current                        Show information about the currently playing song.

spotify> clear                          Clear the command window.
spotify> help                           Show help.
spotify> quit                           Quit application.
```

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
