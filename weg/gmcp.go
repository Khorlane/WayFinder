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

// ReceiveGMCP decodes a Telnet subnegotiation and returns the validated room ID.
func (g *LiveGateway) ReceiveGMCP(message []byte) wne.RoomID {
	if !g.gmcpEnabled {
		return ""
	}
	name, data, ok := bytes.Cut(message, []byte(" "))
	if !ok || !bytes.EqualFold(name, []byte("HolyQuest.Room")) {
		return ""
	}
	var room struct {
		ID   wne.RoomID `json:"id"`
		Name string     `json:"name"`
	}
	if err := json.Unmarshal(data, &room); err != nil || !gmcpRoomID.MatchString(string(room.ID)) {
		return ""
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
	if !g.gotGMCPRoom {
		g.gotGMCPRoom = true
		g.display(LiveText{GMCPStatus: "GMCP active. Receiving room IDs."})
	}
	if g.entered != nil {
		g.entered(RoomEntry{ID: room.ID, Name: title})
	}
	return room.ID
}
