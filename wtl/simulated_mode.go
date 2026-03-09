package wtl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	wcswin32 "WayFinder/wcs/win32"
	"WayFinder/weg"
	"WayFinder/wmr"
	"WayFinder/wne"
)

const (
	simulatedAreaName = "LocalDevArea"
	simulatedPromptHP = 100
	simulatedPromptMP = 100
	descWrapWidth     = 72
)

var uiOut io.Writer = os.Stdout

func uiPrint(a ...any) {
	fmt.Fprint(uiOut, a...)
}

func uiPrintf(format string, a ...any) {
	fmt.Fprintf(uiOut, format, a...)
}

func uiPrintln(a ...any) {
	fmt.Fprintln(uiOut, a...)
}

func setupLogging(logPath string) (func(), error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	origStdout := os.Stdout
	os.Stdout = f
	uiOut = io.MultiWriter(origStdout, f)
	wmr.SetUIOutput(uiOut)
	fmt.Fprintf(f, "\n=== WayFinder run %s ===\n", time.Now().Format(time.RFC3339))
	return func() {
		os.Stdout = origStdout
		uiOut = os.Stdout
		wmr.SetUIOutput(nil)
		_ = f.Close()
	}, nil
}

type World struct {
	exits     map[wmr.RoomID]map[string]wmr.RoomID // room -> dir -> neighbor
	neighbors map[wmr.RoomID]map[wmr.RoomID]struct{}
}

func (w *World) ExitsFrom(roomID wmr.RoomID) map[string]wmr.RoomID {
	if w == nil || w.exits == nil {
		return map[string]wmr.RoomID{}
	}
	ex := w.exits[roomID]
	out := make(map[string]wmr.RoomID, len(ex))
	for d, id := range ex {
		out[d] = id
	}
	return out
}

func (w *World) Neighbors(roomID wmr.RoomID) []wmr.RoomID {
	if w == nil || w.neighbors == nil {
		return nil
	}
	set := w.neighbors[roomID]
	out := make([]wmr.RoomID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (w *World) HasRoom(roomID wmr.RoomID) bool {
	if w == nil || w.exits == nil {
		return false
	}
	_, ok := w.exits[roomID]
	return ok
}

func (w *World) ensureRoom(id wmr.RoomID) {
	if w.exits[id] == nil {
		w.exits[id] = make(map[string]wmr.RoomID)
	}
	if w.neighbors[id] == nil {
		w.neighbors[id] = make(map[wmr.RoomID]struct{})
	}
}

func (w *World) addExit(from wmr.RoomID, dir string, to wmr.RoomID) {
	w.ensureRoom(from)
	w.ensureRoom(to)
	w.exits[from][dir] = to
	w.neighbors[from][to] = struct{}{}
	w.neighbors[to][from] = struct{}{}
}

type DiscoveryState struct {
	discoveredRooms map[wmr.RoomID]struct{}
}

func NewDiscoveryState() *DiscoveryState {
	return &DiscoveryState{
		discoveredRooms: make(map[wmr.RoomID]struct{}),
	}
}

func (d *DiscoveryState) Discover(roomID wmr.RoomID) {
	d.discoveredRooms[roomID] = struct{}{}
}

func (d *DiscoveryState) IsDiscovered(roomID wmr.RoomID) bool {
	if d == nil {
		return false
	}
	_, ok := d.discoveredRooms[roomID]
	return ok
}

func discoveredRoomIDs(discovery *DiscoveryState) []wmr.RoomID {
	ids := make([]wmr.RoomID, 0, len(discovery.discoveredRooms))
	for id := range discovery.discoveredRooms {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func toWNERoomID(id wmr.RoomID) wne.RoomID   { return wne.RoomID(id) }
func fromWNERoomID(id wne.RoomID) wmr.RoomID { return wmr.RoomID(id) }

type wneTopologyAdapter struct {
	t wne.Topology
}

func (a wneTopologyAdapter) ExitsFrom(roomID wmr.RoomID) map[string]wmr.RoomID {
	ex := a.t.ExitsFrom(toWNERoomID(roomID))
	out := make(map[string]wmr.RoomID, len(ex))
	for dir, id := range ex {
		out[dir] = fromWNERoomID(id)
	}
	return out
}

func (a wneTopologyAdapter) Neighbors(roomID wmr.RoomID) []wmr.RoomID {
	neighbors := a.t.Neighbors(toWNERoomID(roomID))
	out := make([]wmr.RoomID, 0, len(neighbors))
	for _, id := range neighbors {
		out = append(out, fromWNERoomID(id))
	}
	return out
}

func (a wneTopologyAdapter) HasRoom(roomID wmr.RoomID) bool {
	return a.t.HasRoom(toWNERoomID(roomID))
}

type wneWorldAdapter struct {
	w *World
}

func (a wneWorldAdapter) ExitsFrom(roomID wne.RoomID) map[string]wne.RoomID {
	ex := a.w.ExitsFrom(fromWNERoomID(roomID))
	out := make(map[string]wne.RoomID, len(ex))
	for dir, id := range ex {
		out[dir] = toWNERoomID(id)
	}
	return out
}

func (a wneWorldAdapter) Neighbors(roomID wne.RoomID) []wne.RoomID {
	neighbors := a.w.Neighbors(fromWNERoomID(roomID))
	out := make([]wne.RoomID, 0, len(neighbors))
	for _, id := range neighbors {
		out = append(out, toWNERoomID(id))
	}
	return out
}

func (a wneWorldAdapter) HasRoom(roomID wne.RoomID) bool {
	return a.w.HasRoom(fromWNERoomID(roomID))
}

type wneMapperAdapter struct {
	m *wmr.Mapper
}

func (a wneMapperAdapter) BindTopology(t wne.Topology) {
	a.m.BindTopology(wneTopologyAdapter{t: t})
}

func (a wneMapperAdapter) Enter(id wne.RoomID, dirMoved string) error {
	return a.m.Enter(fromWNERoomID(id), dirMoved)
}

type wneDiscoveryAdapter struct {
	d *DiscoveryState
}

func (a wneDiscoveryAdapter) Discover(roomID wne.RoomID) {
	a.d.Discover(fromWNERoomID(roomID))
}

func LoadWorld(path string) (*World, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !st.IsDir() {
		return nil, fmt.Errorf("expected Rooms directory path, got file: %s", path)
	}

	w := &World{
		exits:     make(map[wmr.RoomID]map[string]wmr.RoomID),
		neighbors: make(map[wmr.RoomID]map[wmr.RoomID]struct{}),
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}
		full := filepath.Join(path, name)
		if err := parseRoomFileIntoWorld(w, full); err != nil {
			return nil, err
		}
	}

	return w, nil
}

func parseRoomFileIntoWorld(w *World, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var roomID wmr.RoomID
	var curExitName string
	var curExitTo wmr.RoomID

	flushExit := func() error {
		if curExitName == "" {
			return nil
		}
		dir := normalizeDirName(curExitName)
		if dir == "" {
			return fmt.Errorf("%s: unsupported ExitName %q (only North/South/East/West)", filePath, curExitName)
		}
		if roomID == "" {
			return fmt.Errorf("%s: ExitName %q seen before RoomId", filePath, curExitName)
		}
		if curExitTo == "" {
			return fmt.Errorf("%s: ExitName %q missing ExitToRoomId", filePath, curExitName)
		}
		w.addExit(roomID, dir, curExitTo)
		curExitName = ""
		curExitTo = ""
		return nil
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "RoomId:") {
			roomID = wmr.RoomID(strings.TrimSpace(strings.TrimPrefix(line, "RoomId:")))
			continue
		}

		if strings.HasPrefix(line, "ExitName:") {
			if err := flushExit(); err != nil {
				return err
			}
			curExitName = strings.TrimSpace(strings.TrimPrefix(line, "ExitName:"))
			continue
		}

		if strings.HasPrefix(line, "ExitToRoomId:") {
			curExitTo = wmr.RoomID(strings.TrimSpace(strings.TrimPrefix(line, "ExitToRoomId:")))
			continue
		}

		if strings.EqualFold(line, "End of Exits") {
			if err := flushExit(); err != nil {
				return err
			}
			continue
		}

		if strings.EqualFold(line, "End of Room") {
			if err := flushExit(); err != nil {
				return err
			}
			break
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if roomID == "" {
		return fmt.Errorf("%s: missing RoomId:", filePath)
	}
	w.ensureRoom(roomID)
	return nil
}

type localRoomPresentation struct {
	Title       string
	AreaName    string
	Description []string
	Contents    []string
}

func emitLocalPrompt() {
	uiPrintf("%dH %dM > ", simulatedPromptHP, simulatedPromptMP)
}

func emitSimulatedRoomOutput(worldPath string, roomID wmr.RoomID, exits map[string]wmr.RoomID) {
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

func loadLocalRoomPresentation(worldPath string, roomID wmr.RoomID) localRoomPresentation {
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

func formatSimulatedExits(exits map[string]wmr.RoomID) []string {
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

func toRoomIDExits(exits map[string]wne.RoomID) map[string]wmr.RoomID {
	out := make(map[string]wmr.RoomID, len(exits))
	for dir, id := range exits {
		out[dir] = wmr.RoomID(id)
	}
	return out
}

func normalizeDirName(s string) string {
	t := strings.TrimSpace(strings.ToUpper(s))
	switch t {
	case "N", "NORTH":
		return "N"
	case "S", "SOUTH":
		return "S"
	case "E", "EAST":
		return "E"
	case "W", "WEST":
		return "W"
	case "NE", "NORTHEAST", "NORTH-EAST", "NORTH EAST":
		return "NE"
	case "NW", "NORTHWEST", "NORTH-WEST", "NORTH WEST":
		return "NW"
	case "SE", "SOUTHEAST", "SOUTH-EAST", "SOUTH EAST":
		return "SE"
	case "SW", "SOUTHWEST", "SOUTH-WEST", "SOUTH WEST":
		return "SW"
	default:
		return ""
	}
}

// Run executes WTL simulated mode and emits local MUD-style text.
func Run(args []string) int {
	cleanupLog, logErr := setupLogging("log.txt")
	if logErr != nil {
		fmt.Fprintln(os.Stderr, "Logging setup error:", logErr)
	} else {
		defer cleanupLog()
	}

	if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
		uiPrintln("Usage: go run . <Rooms directory path>")
		return 2
	}
	worldPath := args[1]

	world, err := LoadWorld(worldPath)
	if err != nil {
		uiPrintln("LoadWorld error:", err)
		return 1
	}

	mapper := wmr.NewMapper()
	mapper.SetDebugWriter(os.Stdout)
	discovery := NewDiscoveryState()

	start := wne.RoomID("JesseSquare8")
	session, err := wne.NewNavigationSession(
		wneWorldAdapter{w: world},
		wneMapperAdapter{m: mapper},
		wneDiscoveryAdapter{d: discovery},
		start,
	)
	if err != nil {
		uiPrintln(err)
		return 1
	}
	gateway := weg.NewSimulatedGateway(
		session,
		mapper,
		world,
		discovery,
		func() []wmr.RoomID { return discoveredRoomIDs(discovery) },
	)

	uiPrintln("Commands: n s e w ne nw se sw | look | map | coords | show | gui | quit")
	emitSimulatedRoomOutput(worldPath, wmr.RoomID(session.CurrentRoom()), toRoomIDExits(session.CurrentExits()))
	gateway.PrintMap()

	in := bufio.NewReader(os.Stdin)
	for {
		emitLocalPrompt()
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			uiPrintln()
			continue
		}

		result := gateway.IngestRawText("SIMCMD " + line)
		switch result.Kind {
		case weg.KindQuit:
			uiPrintln()
			return 0
		case weg.KindGUI:
			uiPrintln()
			wcswin32.RunWCS()
		case weg.KindLook:
			uiPrintln()
			emitSimulatedRoomOutput(worldPath, result.CurrentRoom, result.CurrentExits)
			gateway.PrintMap()
		case weg.KindMap:
			uiPrintln()
			gateway.PrintMap()
		case weg.KindCoords:
			uiPrintln()
			gateway.PrintCoords()
		case weg.KindShow:
			uiPrintln()
			uiPrintln("Discovered rooms:")
			if len(result.Discovered) == 0 {
				uiPrintln("(none)")
				break
			}
			for _, id := range result.Discovered {
				uiPrintln(string(id))
			}
		case weg.KindMoveFail:
			uiPrintln()
			if result.MoveErr != nil && result.MoveErr.Error() == "no exit that way" {
				emitSimulatedMoveFailure(result.Direction)
				continue
			}
			emitSimulatedSystemText(fmt.Sprintf("System: %v", result.MoveErr))
		case weg.KindMoveOK:
			uiPrintln()
			emitSimulatedRoomOutput(worldPath, result.CurrentRoom, result.CurrentExits)
			gateway.PrintMap()
		default:
			uiPrintln()
			emitSimulatedSystemText("Huh?")
		}
	}
}
