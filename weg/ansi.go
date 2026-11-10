package weg

import (
	"strconv"
	"strings"
)

// LiveColor is zero for the default foreground, then the 16 ANSI foreground colors.
type LiveColor uint8

type LiveSpan struct {
	Text  string
	Color LiveColor
}

type ansiRendition struct {
	foreground LiveColor
	bright     bool
	partial    string
}

func (a *ansiRendition) spans(text string) []LiveSpan {
	text = a.partial + text
	a.partial = ""
	var spans []LiveSpan
	for len(text) > 0 {
		escape := strings.IndexByte(text, '\x1b')
		if escape < 0 {
			spans = append(spans, LiveSpan{Text: text, Color: a.color()})
			break
		}
		if escape > 0 {
			spans = append(spans, LiveSpan{Text: text[:escape], Color: a.color()})
		}
		text = text[escape:]
		if len(text) == 1 {
			a.partial = text
			break
		}
		if text[1] != '[' {
			text = text[1:]
			continue
		}
		end := 2
		for end < len(text) && text[end] >= 0x20 && text[end] <= 0x3f {
			end++
		}
		if end == len(text) {
			if end <= 64 {
				a.partial = text
			}
			break
		}
		if text[end] >= 0x40 && text[end] <= 0x7e {
			if text[end] == 'm' && end <= 64 {
				a.apply(text[2:end])
			}
			end++
		}
		text = text[end:]
	}
	return spans
}

func (a *ansiRendition) color() LiveColor {
	if a.bright && a.foreground >= 1 && a.foreground <= 8 {
		return a.foreground + 8
	}
	return a.foreground
}

func (a *ansiRendition) apply(parameters string) {
	fields := strings.Split(parameters, ";")
	for i := 0; i < len(fields); i++ {
		code, err := strconv.Atoi(fields[i])
		if err != nil && fields[i] != "" {
			continue
		}
		switch {
		case code == 0:
			a.foreground, a.bright = 0, false
		case code == 1:
			a.bright = true
		case code == 22:
			a.bright = false
		case code >= 30 && code <= 37:
			a.foreground = LiveColor(code - 29)
		case code == 39:
			a.foreground = 0
		case code >= 90 && code <= 97:
			a.foreground = LiveColor(code - 81)
		case code == 38 || code == 48 || code == 58:
			// Skip extended color parameters as a unit so they cannot reset the foreground.
			if i+1 < len(fields) {
				switch fields[i+1] {
				case "5":
					i += 2
				case "2":
					i += 4
				}
			}
		}
	}
}
