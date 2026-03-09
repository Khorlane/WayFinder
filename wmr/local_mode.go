package wmr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	wcswin32 "WayFinder/wcs/win32"
	"WayFinder/wne"
)

func (m *Mapper) PrintGrid10x10() {
	m.PrintGrid10x10Discovered(nil)
}

func (m *Mapper) PrintGrid10x10Discovered(discovery *DiscoveryState) {
	first := true
	r1, r2 := 0, 0
	c1, c2 := 0, 0

	for rc, id := range m.occ {
		if discovery != nil && !discovery.IsDiscovered(id) {
			continue
		}
		r, c := rc[0], rc[1]
		if first {
			r1, r2 = r, r
			c1, c2 = c, c
			first = false
			continue
		}
		if r < r1 {
			r1 = r
		}
		if r > r2 {
			r2 = r
		}
		if c < c1 {
			c1 = c
		}
		if c > c2 {
			c2 = c
		}
	}

	if first {
		uiPrintln("(map empty)")
		return
	}

	uiPrintf("R\\C   ")
	for c := c1; c <= c2; c++ {
		uiPrintf("%-4s", colName(c))
	}
	uiPrintln()

	for r := r1; r <= r2; r++ {
		uiPrintf("%-5d", r)
		for c := c1; c <= c2; c++ {
			if id, ok := m.occ[[2]int{r, c}]; ok {
				if discovery != nil && !discovery.IsDiscovered(id) {
					uiPrintf("%-4s", ".")
					continue
				}
				label := cellLabel(id)
				if m.cur != nil && id == m.cur.ID && len(label) >= 2 {
					label = label[:1] + "@" + label[2:]
				}
				uiPrintf("%-4s", label)
			} else {
				uiPrintf("%-4s", ".")
			}
		}
		uiPrintln()
	}
}

func (m *Mapper) PrintRooms() {
	m.PrintRoomsDiscovered(nil, nil)
}

func (m *Mapper) PrintRoomsDiscovered(world *World, discovery *DiscoveryState) {
	var ids []string
	for id := range m.rooms {
		if discovery != nil && !discovery.IsDiscovered(id) {
			continue
		}
		ids = append(ids, string(id))
	}
	sort.Strings(ids)

	uiPrintln("ROOM COORDINATES")
	for _, s := range ids {
		id := RoomID(s)
		rm := m.rooms[id]
		if rm.Placed {
			uiPrintf("%s  (R=%d,C=%s)  cell=%s", s, rm.R, colName(rm.C), cellLabel(id))
			if world != nil && discovery != nil {
				exits := visibleExits(world.exits[id], discovery)
				if len(exits) == 0 {
					uiPrintf("  exits=(none)")
				} else {
					var dirs []string
					for d := range exits {
						dirs = append(dirs, d)
					}
					sort.Strings(dirs)
					uiPrintf("  exits=")
					for i, d := range dirs {
						if i > 0 {
							uiPrint(" ")
						}
						uiPrintf("%s(%s)", d, exits[d])
					}
				}
			}
			uiPrintln()
		} else {
			uiPrintf("%s  (unplaced)\n", s)
		}
	}
}

type World struct {
	exits     map[RoomID]map[string]RoomID // room -> dir -> neighbor
	neighbors map[RoomID]map[RoomID]struct{}
}

func (w *World) ExitsFrom(roomID RoomID) map[string]RoomID {
	if w == nil || w.exits == nil {
		return map[string]RoomID{}
	}
	ex := w.exits[roomID]
	out := make(map[string]RoomID, len(ex))
	for d, id := range ex {
		out[d] = id
	}
	return out
}

func (w *World) Neighbors(roomID RoomID) []RoomID {
	if w == nil || w.neighbors == nil {
		return nil
	}
	set := w.neighbors[roomID]
	out := make([]RoomID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (w *World) HasRoom(roomID RoomID) bool {
	if w == nil || w.exits == nil {
		return false
	}
	_, ok := w.exits[roomID]
	return ok
}

type DiscoveryState struct {
	discoveredRooms map[RoomID]struct{}
}

func NewDiscoveryState() *DiscoveryState {
	return &DiscoveryState{
		discoveredRooms: make(map[RoomID]struct{}),
	}
}

func (d *DiscoveryState) Discover(roomID RoomID) {
	d.discoveredRooms[roomID] = struct{}{}
}

func (d *DiscoveryState) IsDiscovered(roomID RoomID) bool {
	if d == nil {
		return false
	}
	_, ok := d.discoveredRooms[roomID]
	return ok
}

func visibleExits(exits map[string]RoomID, discovery *DiscoveryState) map[string]RoomID {
	visible := make(map[string]RoomID)
	for dir, neighbor := range exits {
		if discovery.IsDiscovered(neighbor) {
			visible[dir] = neighbor
		}
	}
	return visible
}

func discoveredRoomIDs(discovery *DiscoveryState) []RoomID {
	ids := make([]RoomID, 0, len(discovery.discoveredRooms))
	for id := range discovery.discoveredRooms {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func toWNERoomID(id RoomID) wne.RoomID   { return wne.RoomID(id) }
func fromWNERoomID(id wne.RoomID) RoomID { return RoomID(id) }

type wneTopologyAdapter struct {
	t wne.Topology
}

func (a wneTopologyAdapter) ExitsFrom(roomID RoomID) map[string]RoomID {
	ex := a.t.ExitsFrom(toWNERoomID(roomID))
	out := make(map[string]RoomID, len(ex))
	for dir, id := range ex {
		out[dir] = fromWNERoomID(id)
	}
	return out
}

func (a wneTopologyAdapter) Neighbors(roomID RoomID) []RoomID {
	neighbors := a.t.Neighbors(toWNERoomID(roomID))
	out := make([]RoomID, 0, len(neighbors))
	for _, id := range neighbors {
		out = append(out, fromWNERoomID(id))
	}
	return out
}

func (a wneTopologyAdapter) HasRoom(roomID RoomID) bool {
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
	m *Mapper
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
		exits:     make(map[RoomID]map[string]RoomID),
		neighbors: make(map[RoomID]map[RoomID]struct{}),
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

func (w *World) ensureRoom(id RoomID) {
	if w.exits[id] == nil {
		w.exits[id] = make(map[string]RoomID)
	}
	if w.neighbors[id] == nil {
		w.neighbors[id] = make(map[RoomID]struct{})
	}
}

func (w *World) addExit(from RoomID, dir string, to RoomID) {
	w.ensureRoom(from)
	w.ensureRoom(to)
	w.exits[from][dir] = to
	w.neighbors[from][to] = struct{}{}
	w.neighbors[to][from] = struct{}{}
}

// parseRoomFileIntoWorld is local simulated-mode harness parsing only.
// In the target runtime architecture, raw inbound MUD text parsing belongs to
// WEG after WTL provides the text stream.
func parseRoomFileIntoWorld(w *World, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var roomID RoomID
	var curExitName string
	var curExitTo RoomID

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
			roomID = RoomID(strings.TrimSpace(strings.TrimPrefix(line, "RoomId:")))
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
			curExitTo = RoomID(strings.TrimSpace(strings.TrimPrefix(line, "ExitToRoomId:")))
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
	// Ensure room exists in exits map even if it has no exits.
	w.ensureRoom(roomID)
	return nil
}

// Run executes the local simulated development harness.
// This path is non-authoritative and exists to exercise WNE/WMR behavior until
// WTL live mode is wired as the primary inbound text source.
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

	mapper := NewMapper()
	mapper.SetDebugWriter(os.Stdout)
	discovery := NewDiscoveryState()

	// Fixed start room for the harness.
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

	uiPrintln("Commands: n s e w ne nw se sw | look | map | coords | show | gui | quit")
	emitSimulatedRoomOutput(worldPath, RoomID(session.CurrentRoom()), toRoomIDExits(session.CurrentExits()))
	mapper.PrintGrid10x10Discovered(discovery)

	in := bufio.NewReader(os.Stdin)
	for {
		emitLocalPrompt()
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			uiPrintln()
			continue
		}

		switch line {
		case "quit", "exit":
			uiPrintln()
			return 0
		case "gui":
			uiPrintln()
			wcswin32.RunWCS()
		case "look":
			uiPrintln()
			emitSimulatedRoomOutput(worldPath, RoomID(session.CurrentRoom()), toRoomIDExits(session.CurrentExits()))
			mapper.PrintGrid10x10Discovered(discovery)
		case "map":
			uiPrintln()
			mapper.PrintGrid10x10Discovered(discovery)
		case "coords":
			uiPrintln()
			mapper.PrintRoomsDiscovered(world, discovery)
		case "show":
			uiPrintln()
			ids := discoveredRoomIDs(discovery)
			uiPrintln("Discovered rooms:")
			if len(ids) == 0 {
				uiPrintln("(none)")
				break
			}
			for _, id := range ids {
				uiPrintln(string(id))
			}
		default:
			// movement: only N/S/E/W
			dir := normalizeDirName(line)
			if dir == "" {
				uiPrintln()
				emitSimulatedSystemText("Huh?")
				continue
			}
			if err := session.Move(dir); err != nil {
				if err.Error() == "no exit that way" {
					uiPrintln()
					emitSimulatedMoveFailure(dir)
					continue
				}
				uiPrintln()
				emitSimulatedSystemText(fmt.Sprintf("System: %v", err))
				continue
			}
			uiPrintln()
			emitSimulatedRoomOutput(worldPath, RoomID(session.CurrentRoom()), toRoomIDExits(session.CurrentExits()))
			mapper.PrintGrid10x10Discovered(discovery)
		}
	}
}
