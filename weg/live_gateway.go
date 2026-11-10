package weg

import (
	"strings"
	"unicode/utf8"
)

// LiveText carries transcript text and the current input masking state.
type LiveText struct {
	Spans     []LiveSpan
	MaskInput bool
}

// LiveGateway recognizes login prompts in a stream without creating navigation events.
type LiveGateway struct {
	display func(LiveText)
	tail    string
	partial string
	ansi    ansiRendition
}

// The append-only transcript has no cursor backtracking.
var liveTextControls = strings.NewReplacer("\x00", "", "\b", "", "\r\n", "\r\n", "\n", "\r\n")

func NewLiveGateway(display func(LiveText)) *LiveGateway {
	return &LiveGateway{display: display}
}

func (g *LiveGateway) Write(p []byte) (int, error) {
	text := g.partial + string(p)
	g.partial = ""
	// A TCP read can split either a UTF-8 character or a CRLF pair.
	for i := max(0, len(text)-utf8.UTFMax+1); i < len(text); i++ {
		if !utf8.FullRuneInString(text[i:]) || text[i:] == "\r" {
			g.partial = text[i:]
			text = text[:i]
			break
		}
	}
	spans := g.ansi.spans(liveTextControls.Replace(text))
	if len(spans) > 0 {
		for _, span := range spans {
			g.tail += span.Text
			g.tail = g.tail[max(0, len(g.tail)-32):]
		}
		prompt := strings.ToLower(strings.TrimSpace(g.tail))
		g.display(LiveText{
			Spans:     spans,
			MaskInput: strings.HasSuffix(prompt, "password?") || strings.HasSuffix(prompt, "password:"),
		})
	}
	return len(p), nil
}
