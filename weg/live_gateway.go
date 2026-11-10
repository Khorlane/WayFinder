package weg

import (
	"strings"
	"unicode/utf8"
)

// LiveText carries transcript text and the current input masking state.
type LiveText struct {
	Text      string
	MaskInput bool
}

// LiveGateway recognizes login prompts in a stream without creating navigation events.
type LiveGateway struct {
	display func(LiveText)
	tail    string
	partial string
}

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
	if text != "" {
		text = strings.ReplaceAll(text, "\x00", "")
		text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\n", "\r\n")
		g.tail = strings.ToLower(g.tail + text)
		g.tail = g.tail[max(0, len(g.tail)-32):]
		prompt := strings.TrimSpace(g.tail)
		g.display(LiveText{
			Text:      text,
			MaskInput: strings.HasSuffix(prompt, "password?") || strings.HasSuffix(prompt, "password:"),
		})
	}
	return len(p), nil
}
