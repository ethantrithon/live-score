//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//constants taken from midi spec
const (
	NOTE_OFF             = 0x80
	NOTE_ON              = 0x90
	CONTROL              = 0xB0
	SUSTAIN              = 0x40
	SOSTENUTO            = 0x42
	SOFT_PEDAL           = 0x43
	SYSTEM               = 0xF0
	SYSTEM_EXCLUSIVE     = 0xF0
	SYSTEM_END_EXCLUSIVE = 0xF7
)

var (
	//--- flags ---
	shouldEchoBack bool
	echoVelocity   int
	useFlats       bool
	useGUI         bool = true
	keySignature   int

	//---for raylib---
	activeNotes  = []byte{}
	notesToClear = []byte{}
	//hasNewNote               = false
	lastVelocity     byte    = 0
	sustainPercent   float32 = 0
	sostenutoPercent float32 = 0
)

func assertOK(err error) {
	if err != nil {
		panic("assertion failed: " + err.Error())
	}
}

func scanForDevices(allDevices []string) string {
	for _, file := range allDevices {
		if strings.HasPrefix(file, "midi") {
			return file
		}
	}

	return ""
}

func midiReadAndUpdateValues(midi *bufio.Reader, dev *os.File) {
	msg, err := midi.ReadByte()
	assertOK(err)

	if msg == 0xFE {
		//safe to ignore
		return
	}

	//get just the status number, low nibble is the midi channel
	switch msg & 0xF0 {
	case NOTE_ON:
		note(msg, true, midi, dev)
	case NOTE_OFF:
		note(msg, false, midi, dev)

	case CONTROL:
		control(msg, midi)

	case SYSTEM:
		switch msg {
		case SYSTEM_EXCLUSIVE:
			fmt.Print("System exclusive message: [")
			for {
				b, _ := midi.ReadByte()
				if b == SYSTEM_END_EXCLUSIVE {
					break
				}
				fmt.Printf("%02X ", b)
			}
			fmt.Println("\b]")
		}
	default:
		fmt.Printf("Byte %02X (%03d, %08b)\n", msg, msg, msg)
	}
}

func note(msg byte, on bool, b *bufio.Reader, device *os.File) {
	note, _ := b.ReadByte()
	velocity, _ := b.ReadByte()

	fmt.Printf("Input   Channel %02d: Note %s %03d (%s) @ velocity %03d\n",
		msg&0x0F,
		map[bool]string{true: "on ", false: "off"}[on],
		note,
		noteName(note, useFlats),
		velocity,
	)

	if on {
		activeNotes = append(activeNotes, note)
		lastVelocity = velocity
		//hasNewNote = true
	} else {
		notesToClear = append(notesToClear, note)
	}

	if shouldEchoBack {
		device.Write([]byte{msg, note, byte(echoVelocity)})
	}
}

func noteName(midiValue byte, flats bool) string {
	noteLetter := map[bool][]string{
		false: {"C♮", "C♯", "D♮", "D♯", "E♮", "F♮", "F♯", "G♮", "G♯", "A♮", "A♯", "B♮"},
		true:  {"C♮", "D♭", "D♮", "E♭", "E♮", "F♮", "G♭", "G♮", "A♭", "A♮", "B♭", "B♮"},
	}[flats][midiValue%12]

	//integer divide; - 1 because midi says so (octaves start at -1)
	octave := midiValue/12 - 1

	return fmt.Sprintf("%s-%d", noteLetter, octave)
}

func isOnLine(note byte) bool {
	//look up table - naturals
	switch note {
	case 23, 26, 29, 33, 36, 40, 43, 47, 50, 53, 57, 60,
		64, 67, 71, 74, 77, 81, 84, 88, 91, 95, 98, 101, 105, 108:
		return true
	}

	if !useFlats {
		//sharps
		switch note {
		case 27, 30, 34, 37, 44, 51, 54, 58, 61,
			68, 75, 78, 82, 85, 92, 99, 102, 106:
			return true
		}
	} else {
		//flats
		switch note {
		case 22, 25, 32, 39, 42, 46, 49, 56, 63,
			66, 70, 73, 80, 87, 90, 94, 97, 104:
			return true
		}
	}

	return false
}

func control(msg byte, b *bufio.Reader) {
	ctrl, _ := b.ReadByte()
	value, _ := b.ReadByte()

	fmt.Printf("Control Channel %02d: ", msg&0x0F)

	switch ctrl {
	case SUSTAIN:
		sustainPercent = float32(value) / 127
		fmt.Printf("Sustain @ %06.2f%% (%02X)\n", 100*(float32(value)/127), value)

	case SOSTENUTO:
		sostenutoPercent = float32(value) / 127
		fmt.Printf("Sostenuto %s\n",
			map[bool]string{
				true:  "on",
				false: "off",
			}[value > 64])

	case SOFT_PEDAL:
		fmt.Printf("Soft Pedal %s\n",
			map[bool]string{
				true:  "on",
				false: "off",
			}[value > 64],
		)

	default:
		fmt.Printf("Byte %02X (%03d, %08b)\n", msg, msg, msg)
		fmt.Printf("Byte %02X (%03d, %08b)\n", ctrl, ctrl, ctrl)
		fmt.Printf("Byte %02X (%03d, %08b)\n", value, value, value)
	}
}

func listenForLines() <-chan []byte {
	ch := make(chan []byte)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			numbers := strings.Split(s.Text(), " ")
			bytes := []byte{}
			for _, n := range numbers {
				b, _ := strconv.ParseInt(n, 16, 0)
				bytes = append(bytes, byte(b))
			}
			ch <- bytes
		}
	}()
	return ch
}
