package wmr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"WayFinder/wne"
)

const (
	simulatedAreaName = "LocalDevArea"
	simulatedPromptHP = 100
	simulatedPromptMP = 100
	descWrapWidth     = 72
)

// Simulated MUD output is a development harness feed.
// It emulates the inbound text source until WTL live mode becomes the primary
// runtime text source.
type localRoomPresentation struct {
	Title       string
	AreaName    string
	Description []string
	Contents    []string
}

func emitLocalPrompt() {
	uiPrintf("%dH %dM > ", simulatedPromptHP, simulatedPromptMP)
}

func emitSimulatedRoomOutput(worldPath string, roomID RoomID, exits map[string]RoomID) {
	room := loadLocalRoomPresentation(worldPath, roomID)

	uiPrintln()
	uiPrintf("%s [%s %s]\n", room.Title, roomID, room.AreaName)
	for _, line := range room.Description {
		uiPrintln(line)
	}

	formattedExits := formatSimulatedExits(exits)
	uiPrintf("Exits: %s\n", strings.Join(formattedExits, " "))
	for _, line := range room.Contents {
		uiPrintln(line)
	}
	uiPrintln()
}

func emitSimulatedMoveFailure(dir string) {
	uiPrintln()
	if dir == "" {
		uiPrintln("You cannot go that way.")
		uiPrintln()
		return
	}
	uiPrintf("You cannot go %s.\n", strings.ToLower(formatExitDisplayName(dir)))
	uiPrintln()
}

func emitSimulatedSystemText(line string) {
	uiPrintln()
	uiPrintln(line)
	uiPrintln()
}

func loadLocalRoomPresentation(worldPath string, roomID RoomID) localRoomPresentation {
	room := localRoomPresentation{
		Title:       string(roomID),
		AreaName:    simulatedAreaName,
		Description: []string{"You stand in a familiar place."},
	}

	filePath := filepath.Join(worldPath, fmt.Sprintf("%s.txt", roomID))
	f, err := os.Open(filePath)
	if err != nil {
		return room
	}
	defer f.Close()

	var (
		inDescBlock bool
		descRaw     []string
	)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "RoomName:") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "RoomName:"))
			if name != "" {
				room.Title = name
			}
			continue
		}

		if strings.HasPrefix(trimmed, "RoomDesc:") {
			inDescBlock = true
			descRaw = descRaw[:0]
			continue
		}

		if inDescBlock {
			if strings.EqualFold(trimmed, "End of RoomDesc") {
				inDescBlock = false
				room.Description = wrapDescriptionLines(normalizeDescription(descRaw), descWrapWidth)
				continue
			}
			descRaw = append(descRaw, trimmed)
		}
	}

	return room
}

func normalizeDescription(lines []string) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return "You stand in a familiar place."
	}
	return strings.Join(parts, " ")
}

func wrapDescriptionLines(text string, width int) []string {
	if width < 20 {
		width = 20
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{"You stand in a familiar place."}
	}

	lines := make([]string, 0, 4)
	cur := words[0]
	for _, w := range words[1:] {
		if len(cur)+1+len(w) <= width {
			cur += " " + w
			continue
		}
		lines = append(lines, cur)
		cur = w
	}
	lines = append(lines, cur)
	return lines
}

func formatSimulatedExits(exits map[string]RoomID) []string {
	if len(exits) == 0 {
		return []string{"None"}
	}

	var dirs []string
	for dir := range exits {
		dirs = append(dirs, dir)
	}
	sort.Slice(dirs, func(i, j int) bool {
		ri := exitSortRank(dirs[i])
		rj := exitSortRank(dirs[j])
		if ri != rj {
			return ri < rj
		}
		return dirs[i] < dirs[j]
	})

	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		out = append(out, formatExitDisplayName(dir))
	}
	return out
}

func exitSortRank(dir string) int {
	switch strings.ToUpper(strings.TrimSpace(dir)) {
	case "N":
		return 0
	case "NE":
		return 1
	case "E":
		return 2
	case "SE":
		return 3
	case "S":
		return 4
	case "SW":
		return 5
	case "W":
		return 6
	case "NW":
		return 7
	default:
		return 100
	}
}

func formatExitDisplayName(dir string) string {
	switch strings.ToUpper(strings.TrimSpace(dir)) {
	case "N":
		return "North"
	case "S":
		return "South"
	case "E":
		return "East"
	case "W":
		return "West"
	case "NE":
		return "Northeast"
	case "NW":
		return "Northwest"
	case "SE":
		return "Southeast"
	case "SW":
		return "Southwest"
	default:
		t := strings.TrimSpace(dir)
		if t == "" {
			return "Unknown"
		}
		return strings.ToUpper(t[:1]) + strings.ToLower(t[1:])
	}
}

func toRoomIDExits(exits map[string]wne.RoomID) map[string]RoomID {
	out := make(map[string]RoomID, len(exits))
	for dir, id := range exits {
		out[dir] = RoomID(id)
	}
	return out
}
