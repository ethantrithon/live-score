# Live Score

## About

Live Score is a display program for MIDI devices (such as pianos) on Linux. It
shows whatever notes or pedals you are playing at any time. You can set it to
use sharps or flats, as well as the key signature (0 to 7 accidentals in
standard order). You can also edit `gui.go` to change the colors of the
background, staves and the notes. Made with
[these](https://github.com/gen2brain/raylib-go) raylib bindings.
GUI can also be disabled to get human-readable MIDI data on the commandline.
Uses "Bravura" as the default music font but any SMuFL font should work (I have
not tried any others); to change the font, place it in the directory and call
the font file `musicFont.otf`.

## Why

Why not :D

## Building

* Clone this repo
* [optional: edit code to your pleasure]
* `go build`
* `./live-score`

## Usage

```
Usage of ./live-score:
  -echo
        Echo (note) input back to midi source (default true)
  -echovel int
        Velocity to use for the echo (default 2)
  -flat
        alias for -flats
  -flats
        Use flats (♭) instead of sharps (♯)
  -key int
        How many accidentals your key signature has (e.g. A Major would have *3* sharps)
  -nogui
        disable gui
```

## Screenshots

Coming soon

## License

MPL 2.0
