package weg

import (
	"strings"

	"WayFinder/wmr"
	"WayFinder/wne"
)

const (
	KindUnknown  = "unknown"
	KindQuit     = "quit"
	KindGUI      = "gui"
	KindLook     = "look"
	KindMap      = "map"
	KindCoords   = "coords"
	KindShow     = "show"
	KindMoveOK   = "move_ok"
	KindMoveFail = "move_fail"
)

type discoveredView interface {
	IsDiscovered(wmr.RoomID) bool
}

type worldView interface {
	ExitsFrom(wmr.RoomID) map[string]wmr.RoomID
}

type Result struct {
	Kind         string
	Direction    string
	CurrentRoom  wmr.RoomID
	CurrentExits map[string]wmr.RoomID
	Discovered   []wmr.RoomID
	MoveErr      error
}

type SimulatedGateway struct {
	nav           wne.Navigator
	mapper        *wmr.Mapper
	world         worldView
	discovery     discoveredView
	discoveredIDs func() []wmr.RoomID
}

func NewSimulatedGateway(nav wne.Navigator, mapper *wmr.Mapper, world worldView, discovery discoveredView, discoveredIDs func() []wmr.RoomID) *SimulatedGateway {
	return &SimulatedGateway{
		nav:           nav,
		mapper:        mapper,
		world:         world,
		discovery:     discovery,
		discoveredIDs: discoveredIDs,
	}
}

func (g *SimulatedGateway) IngestRawText(raw string) Result {
	cmd := strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToUpper(cmd), "SIMCMD ") {
		cmd = strings.TrimSpace(cmd[len("SIMCMD "):])
	}
	cmd = strings.TrimSpace(strings.ToLower(cmd))

	switch cmd {
	case "quit", "exit":
		return Result{Kind: KindQuit}
	case "gui":
		return Result{Kind: KindGUI}
	case "look":
		return g.snapshot(KindLook, "")
	case "map":
		return Result{Kind: KindMap}
	case "coords":
		return Result{Kind: KindCoords}
	case "show":
		r := Result{Kind: KindShow}
		if g.discoveredIDs != nil {
			r.Discovered = g.discoveredIDs()
		}
		return r
	}

	dir := normalizeDirName(cmd)
	if dir == "" {
		return Result{Kind: KindUnknown}
	}

	if err := g.nav.Move(dir); err != nil {
		return Result{
			Kind:      KindMoveFail,
			Direction: dir,
			MoveErr:   err,
		}
	}
	return g.snapshot(KindMoveOK, dir)
}

func (g *SimulatedGateway) PrintMap() {
	g.mapper.PrintGrid10x10Discovered(g.discovery)
}

func (g *SimulatedGateway) PrintCoords() {
	g.mapper.PrintRoomsDiscovered(g.world, g.discovery)
}

func (g *SimulatedGateway) snapshot(kind, dir string) Result {
	return Result{
		Kind:         kind,
		Direction:    dir,
		CurrentRoom:  wmr.RoomID(g.nav.CurrentRoom()),
		CurrentExits: toRoomIDExits(g.nav.CurrentExits()),
	}
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

func toRoomIDExits(exits map[string]wne.RoomID) map[string]wmr.RoomID {
	out := make(map[string]wmr.RoomID, len(exits))
	for dir, id := range exits {
		out[dir] = wmr.RoomID(id)
	}
	return out
}
