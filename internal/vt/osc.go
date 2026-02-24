// Package vt provides a virtual terminal implementation.
// SKIP: Fix typecheck errors - function signature mismatches and undefined types
package vt

import (
	"bytes"
	"image/color"
	"io"

	"github.com/charmbracelet/x/ansi"
)

// handleOsc handles an OSC escape sequence.
func (e *Emulator) handleOsc(cmd int, data []byte) {
	e.flushGrapheme() // Flush any pending grapheme before handling OSC sequences.
	if !e.handlers.handleOsc(cmd, data) {
		e.logf("unhandled sequence: OSC %q", data)
	}
}

func (e *Emulator) handleTitle(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}
	switch cmd {
	case 0: // Set window title and icon name
		name := string(parts[1])
		e.iconName, e.title = name, name
		if e.cb.Title != nil {
			e.cb.Title(name)
		}
		if e.cb.IconName != nil {
			e.cb.IconName(name)
		}
	case 1: // Set icon name
		name := string(parts[1])
		e.iconName = name
		if e.cb.IconName != nil {
			e.cb.IconName(name)
		}
	case 2: // Set window title
		name := string(parts[1])
		e.title = name
		if e.cb.Title != nil {
			e.cb.Title(name)
		}
	}
}

func (e *Emulator) handleDefaultColor(cmd int, data []byte) {
	if cmd != 10 && cmd != 11 && cmd != 12 &&
		cmd != 110 && cmd != 111 && cmd != 112 {
		// Invalid, ignore
		return
	}

	parts := bytes.Split(data, []byte{';'})
	if len(parts) == 0 {
		// Invalid, ignore
		return
	}

	cb := func(c color.Color) {
		switch cmd {
		case 10, 110: // Foreground color
			e.SetForegroundColor(c)
		case 11, 111: // Background color
			e.SetBackgroundColor(c)
		case 12, 112: // Cursor color
			e.SetCursorColor(c)
		}
	}

	switch len(parts) {
	case 1: // Reset color
		cb(nil)
	case 2: // Set/Query color
		arg := string(parts[1])
		if arg == "?" {
			var xrgb ansi.XRGBColor
			switch cmd {
			case 10: // Query foreground color
				xrgb.Color = e.ForegroundColor()
				if xrgb.Color != nil {
					_, _ = io.WriteString(e.pw, ansi.SetForegroundColor(xrgb.String()))
				}
			case 11: // Query background color
				xrgb.Color = e.BackgroundColor()
				if xrgb.Color != nil {
					_, _ = io.WriteString(e.pw, ansi.SetBackgroundColor(xrgb.String()))
				}
			case 12: // Query cursor color
				xrgb.Color = e.CursorColor()
				if xrgb.Color != nil {
					_, _ = io.WriteString(e.pw, ansi.SetCursorColor(xrgb.String()))
				}
			}
		} else if c := ansi.XParseColor(arg); c != nil {
			cb(c)
		}
	}
}

func (e *Emulator) handleWorkingDirectory(cmd int, data []byte) {
	if cmd != 7 {
		// Invalid, ignore
		return
	}

	// The data is the working directory path.
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}

	path := string(parts[1])
	e.cwd = path

	if e.cb.WorkingDirectory != nil {
		e.cb.WorkingDirectory(path)
	}
}

func (e *Emulator) handleSemanticZone(data []byte) {
	// OSC 133 format: "133;<subcommand>[;params]"
	// data includes the "133;" prefix from the parser
	parts := bytes.Split(data, []byte{';'})
	if len(parts) < 2 || len(parts[1]) == 0 {
		return
	}

	subCmd := parts[1][0] // 'A', 'B', 'C', or 'D'
	switch subCmd {
	case 'A', 'B', 'C', 'D':
		// valid
	default:
		return
	}

	curX, curY := e.scr.CursorPosition()
	absLine := e.ScrollbackLen() + curY

	exitCode := -1
	if subCmd == 'D' && len(parts) >= 3 {
		// Parse exit code from params (e.g., "D;0" or "D;1")
		code := 0
		for _, b := range parts[2] {
			if b >= '0' && b <= '9' {
				code = code*10 + int(b-'0')
			}
		}
		if len(parts[2]) > 0 {
			exitCode = code
		}
	}

	if e.semanticMarkers != nil {
		marker := SemanticMarker{
			Type:     SemanticMarkerType(subCmd),
			AbsLine:  absLine,
			Col:      curX,
			ExitCode: exitCode,
		}

		// On C marker (command executed), capture the command text from the
		// terminal buffer before the program's output overwrites it.
		// This is the only reliable time to read the command text.
		if subCmd == 'C' {
			if bMarker := e.semanticMarkers.Last(MarkerCommandStart); bMarker != nil {
				marker.CapturedText = e.extractCommandText(bMarker.AbsLine, bMarker.Col, absLine, curX)
			}
		}

		e.semanticMarkers.Add(marker)
	}
}

func (e *Emulator) handleHyperlink(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 3 || cmd != 8 {
		// Invalid, ignore
		return
	}

	e.scr.cur.Link.URL = string(parts[1])
	e.scr.cur.Link.Params = string(parts[2])
}
