//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"sort"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	lineSpacing        = 32   //px
	lineThickness      = 3    //px
	width              = 1600 //px
	height             = 900  //px
	halfWidth          = width / 2
	halfHeight         = height / 2
	fpsCap             = 300
	sharpKeySignatures = "FCGDAEB"
	flatKeySignatures  = "BEADGCF"
)

const (
	//music fonts (only bravura? needs verify) draw notes two lines too low,
	//so we adjust the Y drawing accordingly
	middleCY int32 = halfHeight - ((iota + 4) * lineSpacing / 2)
	trebleDY
	trebleEY
	trebleFY
	trebleGY
	trebleAY
	trebleBY
)

const (
	bassBY int32 = 360 + ((iota - 3) * lineSpacing / 2)
	bassAY
	bassGY
	bassFY
	bassEY
	bassDY
)

const (
	trebleMiddleLineY = halfHeight - 3*lineSpacing
	bassMiddleLineY   = halfHeight + 3*lineSpacing
	octaveHeight      = 7 * (lineSpacing / 2)
)

var (
	//set preferred font for U+E000-U+F8FF to some music font to render symbols properly
	fontSize       = float32(lineSpacing * 5)
	fontCodePoints = []rune("")
	musicFont      rl.Font
	noteX          = float32(500)
	noteWidth      float32

	sustainStarTime   float32 = 1000000
	sostenutoStarTime float32 = 1000000

	keySigString = map[bool]string{
		false: sharpKeySignatures,
		true:  flatKeySignatures,
	}[useFlats][:keySignature]
	keySignatureSettingOpen = false

	//change these for the background, foreground and music colors
	//respectively
	BGCOL = rl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	FGCOL = rl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	MUSIC = rl.Color{R: 0xFF, G: 0xCB, B: 0x00, A: 0xFF}
)

func raylibWindow() {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.SetTraceLog(rl.LogNone)
	rl.InitWindow(width, height, "Live Score")
	rl.SetTargetFPS(fpsCap)

	musicFont = rl.LoadFontEx(
		"musicFont.otf",
		int32(fontSize),
		&fontCodePoints[0],
		int32(len(fontCodePoints)),
	)

	noteWidth = rl.MeasureTextEx(musicFont, "", fontSize, 1).X

	for !rl.WindowShouldClose() {
		//do this while not drawing -> better perf
		//sort the active notes so we can draw note beams easier (in the
		//future
		activeNotes = unique(activeNotes)
		sort.Slice(activeNotes, func(i, j int) bool {
			return activeNotes[i] > activeNotes[j]
		})

		rl.BeginDrawing()
		rl.ClearBackground(BGCOL)
		draw()
		rl.EndDrawing()

		for _, note := range notesToClear {
			activeNotes = remove(activeNotes, note)
		}
		notesToClear = []byte{}
	}

	rl.CloseWindow()
}

func draw() {
	drawStaff()
	drawKeySignature()
	drawNotes()
	drawPetalStatus()
	drawSettings()
}

func drawStaff() {
	for i := -2; i <= 2; i++ {
		lineY := float32(trebleMiddleLineY + i*lineSpacing)
		rl.DrawLineEx(
			rl.Vector2{X: 0, Y: lineY},
			rl.Vector2{X: width, Y: lineY},
			lineThickness,
			FGCOL,
		)
	}
	rl.DrawTextEx(
		musicFont,
		"",
		rl.Vector2{
			X: 50,
			Y: float32(trebleMiddleLineY) - 2*lineSpacing,
		},
		fontSize,
		1,
		FGCOL,
	)

	for i := -2; i <= 2; i++ {
		lineY := float32(bassMiddleLineY + i*lineSpacing)
		rl.DrawLineEx(
			rl.Vector2{X: 0, Y: lineY},
			rl.Vector2{X: width, Y: lineY},
			lineThickness,
			FGCOL,
		)
	}
	rl.DrawTextEx(
		musicFont,
		"",
		rl.Vector2{
			X: 50,
			Y: float32(bassMiddleLineY) - 2*lineSpacing,
		},
		fontSize,
		1,
		FGCOL,
	)
}

type keyoffsetMap = map[rune][2]int32

func drawKeySignature() {
	symbol := ""
	if useFlats {
		symbol = ""
	}

	symbolWidth := rl.MeasureTextEx(musicFont, symbol, fontSize, 1).X

	sharpOffsets := keyoffsetMap{
		'C': {middleCY - octaveHeight, middleCY + octaveHeight},
		'D': {trebleDY - octaveHeight, bassDY},
		'E': {trebleEY - octaveHeight, bassEY},
		'F': {trebleFY - octaveHeight, bassFY},
		'G': {trebleGY - octaveHeight, bassGY},
		'A': {trebleAY, bassAY + octaveHeight},
		'B': {trebleBY, bassBY + octaveHeight},
	}

	flatOffsets := keyoffsetMap{
		'C': {middleCY - octaveHeight, middleCY + octaveHeight},
		'D': {trebleDY - octaveHeight, bassDY},
		'E': {trebleEY - octaveHeight, bassEY},
		'F': {trebleFY, bassFY + octaveHeight},
		'G': {trebleGY, bassGY + octaveHeight},
		'A': {trebleAY, bassAY + octaveHeight},
		'B': {trebleBY, bassBY + octaveHeight},
	}

	offsets := map[bool]keyoffsetMap{
		false: sharpOffsets,
		true:  flatOffsets,
	}

	for i, changedNote := range keySigString {
		for staff := range []int{0, 1} {
			rl.DrawTextEx(
				musicFont,
				symbol,
				rl.Vector2{
					X: 150 + float32(i)*(symbolWidth+10),
					Y: float32(offsets[useFlats][changedNote][staff]),
				},
				fontSize,
				1,
				FGCOL,
			)
		}
	}
}

func yOffsetFor(note byte) int32 {
	//what note is it
	noteStr := strings.Split(noteName(note, useFlats), "-")
	name := noteStr[0][0]

	//what octave is it
	octave, _ := strconv.Atoi(noteStr[1])
	octaveOffset := int32((4 - octave) * octaveHeight)

	//which actual note
	baseY := map[byte]int32{
		'C': middleCY,
		'D': trebleDY,
		'E': trebleEY,
		'F': trebleFY,
		'G': trebleGY,
		'A': trebleAY,
		'B': trebleBY,
	}[name]
	return baseY + octaveOffset
}

//big yike... refactor this.
func drawNotes() {
	shiftXFactor := 0
	stemDown := false

	for noteIdx, note := range activeNotes {
		yOff := yOffsetFor(note)

		//......................in bravura
		apparentNoteY := yOff + 2*lineSpacing

		//......................half octave
		if noteIdx == 0 && yOff <= trebleBY {
			stemDown = true
		}

		drawNoteHead(note, noteIdx, yOff, &shiftXFactor, &stemDown)
		drawLedgerLines(yOff, apparentNoteY, shiftXFactor)
		drawStem(stemDown, apparentNoteY, shiftXFactor)
		drawAccidental(note, float32(yOff))
	}
}

func drawNoteHead(
	note byte,
	noteIdx int,
	yOff int32,
	shiftXFactor *int,
	stemDown *bool,
) {
	hasPrev := noteIdx > 0
	hasNext := noteIdx < len(activeNotes)-1
	shouldToggleShift := false

	couldCollide := false
	closeEnough := false

	if hasNext {
		next := activeNotes[noteIdx+1]
		visualDistance := yOffsetFor(next) - yOff

		couldCollide = isOnLine(note) != isOnLine(next)
		closeEnough = visualDistance <= lineSpacing/2
	}

	if hasPrev {
		prev := activeNotes[noteIdx-1]
		visualDistance := yOff - yOffsetFor(prev)
		if visualDistance >= 4*lineSpacing {
			if yOff > middleCY {
				*stemDown = false
			}

			shouldToggleShift = false
			*shiftXFactor = 0
		} else {
			//only update these if we're less than
			//an octave away
			couldCollide = isOnLine(note) != isOnLine(prev)
			closeEnough = visualDistance <= lineSpacing/2
		}
	}

	if couldCollide && closeEnough {
		shouldToggleShift = true
	}

	if shouldToggleShift {
		//toggle between 0 and 1
		*shiftXFactor = 1 - *shiftXFactor
	}

	if noteIdx == 0 && yOff <= trebleBY {
		*shiftXFactor = 1
	}

	rl.DrawTextEx(
		musicFont,
		"",
		rl.Vector2{
			X: noteX + (float32(*shiftXFactor) * (noteWidth - lineThickness)),
			Y: float32(yOff),
		},
		fontSize,
		1,
		MUSIC,
	)
}

func drawLedgerLineAt(y float32, shiftXFactor int) {
	rl.DrawRectangle(
		int32(noteX-lineSpacing/2+float32(shiftXFactor)*noteWidth),
		int32(y-lineThickness/2),
		int32(noteWidth+lineSpacing),
		lineThickness,
		MUSIC,
	)
}

func drawLedgerLines(yOff, apparentNoteY int32, shiftXFactor int) {
	ledgersNeededAbove := apparentNoteY <= trebleMiddleLineY-3*lineSpacing
	if ledgersNeededAbove {
		for y := float32(trebleMiddleLineY - 3*lineSpacing); y >= float32(apparentNoteY); y -= lineSpacing {
			drawLedgerLineAt(y, shiftXFactor)
		}
	}

	if yOff == middleCY {
		drawLedgerLineAt(float32(halfHeight), shiftXFactor)
	}

	ledgersNeededBelow := apparentNoteY >= bassMiddleLineY+3*lineSpacing
	if ledgersNeededBelow {
		for y := float32(bassMiddleLineY + 3*lineSpacing); y <= float32(apparentNoteY); y += lineSpacing {
			drawLedgerLineAt(y, shiftXFactor)
		}
	}
}

func drawStem(stemDown bool, apparentNoteY int32, shiftXFactor int) {
	stemYOffset := int32(3.5 * lineSpacing)
	stemLength := int32(3.5*lineSpacing + shiftXFactor*(0.25*lineSpacing))

	if stemDown {
		stemYOffset = 0
	}

	//change stem length
	//TODO

	rl.DrawRectangle(
		int32(noteX+noteWidth-lineThickness),
		apparentNoteY-stemYOffset,
		lineThickness,
		stemLength,
		MUSIC,
	)
}

func drawAccidental(note byte, yOff float32) {
	noteStr := strings.Split(noteName(note, useFlats), "-")
	name := rune(noteStr[0][0])

	keyAffectsCurrentNote := strings.ContainsRune(
		keySigString,
		name,
	)

	if strings.ContainsRune(noteStr[0], '♮') {
		if keyAffectsCurrentNote {
			//draw natural
			rl.DrawTextEx(
				musicFont,
				"",
				rl.Vector2{
					X: noteX - lineSpacing*1.5,
					Y: yOff,
				},
				fontSize,
				1,
				MUSIC,
			)
		}
	} else {
		if !keyAffectsCurrentNote {
			if !useFlats {
				//draw sharp
				rl.DrawTextEx(
					musicFont,
					"",
					rl.Vector2{
						X: noteX - lineSpacing*1.5,
						Y: yOff,
					},
					fontSize,
					1,
					MUSIC,
				)
			} else {
				//draw flat
				rl.DrawTextEx(
					musicFont,
					"",
					rl.Vector2{
						X: noteX - lineSpacing*1.5,
						Y: yOff,
					},
					fontSize,
					1,
					MUSIC,
				)
			}
		}
	}
}

func drawPetalStatus() {
	//draw sustain
	if sustainPercent < 0.2 {
		if sustainStarTime < 0.25 /*seconds*/ {
			sustainStarTime += rl.GetFrameTime()

			rl.DrawTextEx(
				musicFont,
				"",
				rl.Vector2{X: lineSpacing / 2, Y: height - fontSize},
				fontSize,
				1,
				MUSIC,
			)
		}
	} else {
		//pedal pressed
		sustainStarTime = 0
		rl.DrawTextEx(
			musicFont,
			"",
			rl.Vector2{X: lineSpacing / 2, Y: height - fontSize},
			fontSize,
			1,
			rl.Fade(MUSIC, sustainPercent),
		)
	}

	//draw sostenuto
	if sostenutoPercent < 0.2 {
		if sostenutoStarTime < 0.25 /*seconds*/ {
			//pedal released -> blink pedal "star"
			sostenutoStarTime += rl.GetFrameTime()

			rl.DrawTextEx(
				musicFont,
				"",
				rl.Vector2{
					X: lineSpacing + rl.MeasureTextEx(
						musicFont,
						"",
						fontSize,
						1).X,
					Y: height - fontSize,
				},
				fontSize,
				1,
				MUSIC,
			)
		}
	} else {
		//pedal pressed
		sostenutoStarTime = 0
		rl.DrawTextEx(
			musicFont,
			"",
			rl.Vector2{
				X: lineSpacing + rl.MeasureTextEx(
					musicFont,
					"",
					fontSize,
					1).X,
				Y: height - fontSize,
			},
			fontSize,
			1,
			rl.Fade(MUSIC, sostenutoPercent),
		)
	}
}

func drawSettings() {
	const buttonSize = 2 * lineSpacing

	//draw key signature option
	rl.DrawRectangle(
		int32(width-buttonSize*1),
		0,
		buttonSize,
		buttonSize,
		rl.Gray)
	rl.DrawRectangleLines(
		int32(width-buttonSize*1),
		0,
		buttonSize,
		buttonSize,
		BGCOL)

	mouseInsideButton := rl.CheckCollisionPointRec(
		rl.GetMousePosition(),
		rl.Rectangle{
			X:      width - buttonSize*1,
			Y:      0,
			Width:  buttonSize,
			Height: buttonSize,
		})
	if mouseInsideButton && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		keySignatureSettingOpen = !keySignatureSettingOpen
	}

	sign := ""
	if useFlats {
		sign = ""
	}
	rl.DrawTextEx(
		musicFont,
		sign,
		rl.Vector2{
			X: width - buttonSize*1 + 14,
			Y: 4,
		},
		42,
		1,
		MUSIC,
	)
	rl.DrawTextEx(
		musicFont,
		sign,
		rl.Vector2{
			X: width - buttonSize*1 + 30,
			Y: 24,
		},
		42,
		1,
		MUSIC,
	)
	rl.DrawTextEx(
		musicFont,
		sign,
		rl.Vector2{
			X: width - buttonSize*1 + 46,
			Y: 8,
		},
		42,
		1,
		MUSIC,
	)

	if keySignatureSettingOpen {
		//minus button
		rl.DrawRectangle(
			int32(width-buttonSize*1),
			buttonSize,
			buttonSize,
			buttonSize,
			rl.Gray)
		rl.DrawRectangleLines(
			int32(width-buttonSize*1),
			buttonSize,
			buttonSize,
			buttonSize,
			BGCOL)

		//minus sign
		rl.DrawRectangle(
			int32(width-buttonSize*1+buttonSize/4),
			1.5*buttonSize-2,
			buttonSize/2,
			4,
			BGCOL)

		mouseInsideButton = rl.CheckCollisionPointRec(
			rl.GetMousePosition(),
			rl.Rectangle{
				X:      width - buttonSize*1,
				Y:      buttonSize,
				Width:  buttonSize,
				Height: buttonSize,
			},
		)

		if mouseInsideButton && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			keySignature--
			if keySignature < 0 {
				keySignature = 0
			}
			keySigString = map[bool]string{
				false: sharpKeySignatures,
				true:  flatKeySignatures,
			}[useFlats][:keySignature]
		}

		//plus button
		rl.DrawRectangle(
			int32(width-buttonSize*2),
			buttonSize,
			buttonSize,
			buttonSize,
			rl.Gray)
		rl.DrawRectangleLines(
			int32(width-buttonSize*2),
			buttonSize,
			buttonSize,
			buttonSize,
			BGCOL)
		//plus sign
		rl.DrawRectangle(
			int32(width-buttonSize*2+buttonSize/4),
			1.5*buttonSize-2,
			buttonSize/2,
			4,
			BGCOL)
		rl.DrawRectangle(
			int32(width-buttonSize*1.5-2),
			1.25*buttonSize,
			4,
			buttonSize/2,
			BGCOL)

		mouseInsideButton = rl.CheckCollisionPointRec(
			rl.GetMousePosition(),
			rl.Rectangle{
				X:      width - buttonSize*2,
				Y:      buttonSize,
				Width:  buttonSize,
				Height: buttonSize,
			},
		)
		if mouseInsideButton && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			keySignature++
			if keySignature > 7 {
				keySignature = 7
			}
			keySigString = map[bool]string{
				false: sharpKeySignatures,
				true:  flatKeySignatures,
			}[useFlats][:keySignature]
		}
	}
	//end key signature option

	//draw flat/sharp button
	rl.DrawRectangle(
		int32(width-buttonSize*2),
		0,
		buttonSize,
		buttonSize,
		rl.Gray)
	rl.DrawRectangleLines(
		int32(width-buttonSize*2),
		0,
		buttonSize,
		buttonSize,
		BGCOL)
	rl.DrawLine(
		int32(width-buttonSize*2+buttonSize-4),
		4,
		int32(width-buttonSize*2+4),
		buttonSize-4,
		BGCOL,
	)

	mouseInsideButton = rl.CheckCollisionPointRec(
		rl.GetMousePosition(),
		rl.Rectangle{
			X:      width - buttonSize*2,
			Y:      0,
			Width:  buttonSize,
			Height: buttonSize,
		})
	if mouseInsideButton && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		keySignatureSettingOpen = false
		useFlats = !useFlats
		keySigString = map[bool]string{
			false: sharpKeySignatures,
			true:  flatKeySignatures,
		}[useFlats][:keySignature]
	}

	colors := map[bool]rl.Color{
		false: BGCOL,
		true:  MUSIC,
	}
	sharpColor, flatColor := colors[!useFlats], colors[useFlats]
	rl.DrawTextEx(
		musicFont,
		"",
		rl.Vector2{
			X: width - buttonSize*2 + 8,
			Y: -4, //font size doesn't matter here
		},
		64,
		1,
		sharpColor,
	)
	rl.DrawTextEx(
		musicFont,
		"",
		rl.Vector2{
			X: width - buttonSize*2 + buttonSize - 24,
			Y: 24,
		},
		64,
		1,
		flatColor,
	)
	//end flat/sharp button
}
