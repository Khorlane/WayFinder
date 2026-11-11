package weg

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"unicode"

	"WayFinder/wne"
)

var gmcpRoomID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

const GMCPWaitingStatus = "Waiting for GMCP offer. Using room text."

// SetGMCPEnabled selects the authoritative room event source for the connection.
// The receive loop calls it between writes of transcript text.
func (g *LiveGateway) SetGMCPEnabled(enabled bool) {
	g.gmcpEnabled = enabled
	g.gotGMCPRoom = false
	g.gotGMCPExits = false
	g.rooms = roomStream{}

	g.commandMu.Lock()
	g.gameStarted = false
	g.pendingCommands = 0
	g.commandDirection = ""
	g.commandMu.Unlock()

	status := "GMCP disabled by server. Using room text."
	if enabled {
		status = "GMCP enabled. Waiting for HolyQuest.Room."
	}
	g.display(LiveText{GMCPStatus: status})
}

// ReceiveGMCP decodes a Telnet subnegotiation and returns the validated room event.
func (g *LiveGateway) ReceiveGMCP(message []byte) RoomEntry {
	if !g.gmcpEnabled {
		return RoomEntry{}
	}
	name, data, ok := bytes.Cut(message, []byte(" "))
	if !ok || !bytes.EqualFold(name, []byte("HolyQuest.Room")) {
		return RoomEntry{}
	}
	var room struct {
		ID    wne.RoomID            `json:"id"`
		Name  string                `json:"name"`
		Exits map[string]wne.RoomID `json:"exits"`
	}
	if err := json.Unmarshal(data, &room); err != nil || !gmcpRoomID.MatchString(string(room.ID)) {
		return RoomEntry{}
	}
	entry := RoomEntry{ID: room.ID}
	if room.Exits != nil {
		entry.Destinations = make(map[string]wne.RoomID, len(room.Exits))
		for key, destination := range room.Exits {
			dir := normalizeDirName(key)
			if dir == "" {
				dir = strings.ToUpper(strings.TrimSpace(key))
				switch dir {
				case "UP":
					dir = "U"
				case "DOWN":
					dir = "D"
				}
			}
			if !gmcpRoomID.MatchString(dir) || !gmcpRoomID.MatchString(string(destination)) {
				return RoomEntry{}
			}
			if previous, exists := entry.Destinations[dir]; exists && previous != destination {
				return RoomEntry{}
			}
			entry.Destinations[dir] = destination
		}
	}
	title := strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, room.Name))
	if title == "" {
		title = string(room.ID)
	}
	entry.Name = title
	if entry.Destinations != nil && !g.gotGMCPExits {
		g.gotGMCPExits = true
		g.display(LiveText{GMCPStatus: "GMCP active. Receiving rooms and exits."})
	} else if !g.gotGMCPRoom {
		g.display(LiveText{GMCPStatus: "GMCP active. Receiving room IDs."})
	}
	g.gotGMCPRoom = true
	if g.entered != nil {
		g.entered(entry)
	}
	return entry
}
