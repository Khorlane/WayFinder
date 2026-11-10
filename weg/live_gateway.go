package weg

import (
	"strings"
	"sync"
	"unicode/utf8"
)

// LiveText carries transcript text and the current input masking state.
type LiveText struct {
	Spans      []LiveSpan
	MaskInput  bool
	GMCPStatus string
}

// LiveGateway decodes the transcript and recognizes login prompts and room entries.
type LiveGateway struct {
	display          func(LiveText)
	entered          func(RoomEntry)
	isRoomName       func(string) bool
	tail             string
	partial          string
	ansi             ansiRendition
	rooms            roomStream
	gmcpEnabled      bool
	gotGMCPRoom      bool
	commandMu        sync.Mutex
	gameStarted      bool
	pendingCommands  int
	commandDirection string
}

// The append-only transcript has no cursor backtracking.
var liveTextControls = strings.NewReplacer("\x00", "", "\b", "", "\r\n", "\r\n", "\n", "\r\n")

func NewLiveGateway(display func(LiveText)) *LiveGateway {
	return &LiveGateway{display: display}
}

func (g *LiveGateway) OnRoomEntry(entered func(RoomEntry), isRoomName func(string) bool) {
	g.entered = entered
	g.isRoomName = isRoomName
}

func (g *LiveGateway) CommandSent(command string) {
	g.commandMu.Lock()
	defer g.commandMu.Unlock()
	if !g.gameStarted || g.entered == nil {
		return
	}
	g.pendingCommands++
	g.commandDirection = ""
	if g.pendingCommands == 1 {
		g.commandDirection = normalizeDirName(command)
	}
}

func (g *LiveGateway) noteGamePrompt() {
	g.commandMu.Lock()
	defer g.commandMu.Unlock()
	g.gameStarted = true
	if g.pendingCommands > 0 {
		g.pendingCommands--
	}
}

func (g *LiveGateway) pendingDirection() string {
	g.commandMu.Lock()
	defer g.commandMu.Unlock()
	if g.pendingCommands == 1 {
		return g.commandDirection
	}
	return ""
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
		if g.entered != nil && !g.gmcpEnabled {
			for _, span := range spans {
				g.rooms.ingest(span.Text, g)
			}
		}
	}
	return len(p), nil
}
