package weg

import (
	"regexp"
	"strings"

	"WayFinder/wne"
)

type RoomEntry struct {
	ID        wne.RoomID
	Name      string
	Exits     map[string]struct{}
	Direction string
}

var (
	roomHeader = regexp.MustCompile(`^(.+?)[ \t]+\[([A-Za-z0-9_-]+)([ \t]+[^\[\]]*)?\]$`)
	gamePrompt = regexp.MustCompile(`^[0-9]+H[ \t]+[0-9]+M[ \t]*>[ \t]*`)
)

type roomStream struct {
	line        string
	discardLine bool
	sawPrompt   bool
	pending     RoomEntry
}

func (s *roomStream) ingest(text string, g *LiveGateway) {
	for text != "" {
		part, rest, complete := strings.Cut(text, "\n")
		if len(s.line)+len(part) > 4096 {
			s.discardLine = true
		}
		if !s.discardLine {
			s.line += part
			if !s.sawPrompt && gamePrompt.MatchString(strings.TrimSpace(s.line)) {
				s.sawPrompt = true
				g.noteGamePrompt()
			}
		}
		if !complete {
			return
		}
		if !s.discardLine {
			s.finishLine(strings.TrimSpace(s.line), g)
		}
		s.line, s.discardLine, s.sawPrompt = "", false, false
		text = rest
	}
}

func (s *roomStream) finishLine(line string, g *LiveGateway) {
	if prompt := gamePrompt.FindStringIndex(line); prompt != nil {
		s.pending = RoomEntry{}
		line = line[prompt[1]:]
	}
	if fields := roomHeader.FindStringSubmatch(line); fields != nil {
		s.pending = RoomEntry{ID: wne.RoomID(fields[2]), Name: fields[1]}
		return
	}
	if exits, ok := strings.CutPrefix(line, "Exits:"); ok && s.pending.Name != "" {
		room := s.pending
		s.pending = RoomEntry{}
		room.Exits = make(map[string]struct{})
		for _, exit := range strings.Fields(exits) {
			exit = strings.Trim(exit, ",()[]")
			if strings.EqualFold(exit, "none") || exit == "" {
				continue
			}
			if dir := normalizeDirName(exit); dir != "" {
				exit = dir
			}
			room.Exits[exit] = struct{}{}
		}
		room.Direction = g.pendingDirection()
		g.entered(room)
		return
	}
	if g.isRoomName != nil && g.isRoomName(line) {
		s.pending = RoomEntry{Name: line}
	}
}
