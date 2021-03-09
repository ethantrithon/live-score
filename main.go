//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.BoolVar(&shouldEchoBack, "echo", true, "Echo (note) input back to midi source")
	flag.IntVar(&echoVelocity, "echovel", 2, "Velocity to use for the echo")
	flag.IntVar(&keySignature, "key", 0, "How many accidentals your key signature has (e.g. A Major would have *3* sharps)")
	fs := flag.Bool("flats", false, "Use flats (♭) instead of sharps (♯)")
	f := flag.Bool("flat", false, "alias for -flats")
	nogui := flag.Bool("nogui", false, "disable gui")

	flag.Parse()

	if keySignature > 7 {
		keySignature = 7
	}
	if keySignature < 0 {
		keySignature = 0
	}

	useFlats = *f || *fs

	if *nogui {
		useGUI = false
	}

	devs, err := os.Open("/dev")
	assertOK(err)
	allDevices, err := devs.Readdirnames(0)
	assertOK(err)

	name := scanForDevices(allDevices)

	for name == "" {
		devs.Close()
		fmt.Println("Unable to find any midi devices in /dev.")
		fmt.Println("Check connection to device, then press enter to try again, or ^C to cancel.")
		fmt.Scanln()
		devs, _ := os.Open("/dev")
		allDevices, _ = devs.Readdirnames(0)
		name = scanForDevices(allDevices)
	}

	dev, err := os.OpenFile("/dev/"+name, os.O_RDWR, os.ModeDevice)
	assertOK(err)
	devs.Close()

	fmt.Println("# Found MIDI device /dev/"+name, "#")

	midi := bufio.NewReader(dev)

	//Allow manual writes to the midi device
	keyboardInput := listenForLines()
	//Or disallow it (recommended)
	// keyboardInput := make(chan []byte)

	go func() {
		for {
			select {
			case userInput := <-keyboardInput:
				dev.Write(userInput)

			default:
				midiReadAndUpdateValues(midi, dev)
			}
		}
	}()

	if useGUI {
		raylibWindow()
		os.Exit(0)
	} else {
		fmt.Println("^D to quit")

		for bufio.NewScanner(os.Stdin).Scan() {
		}
	}
}
