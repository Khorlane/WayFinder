package wne

import (
	"fmt"
	"maps"
	"slices"
)

// RoomArrival describes a server-confirmed visit and its relationship to the previous room.
type RoomArrival struct {
	Room, From RoomID
	Direction  string
	Known      bool
}

// LiveNavigation records visits independently of whether the map can place them.
type LiveNavigation struct {
	world     LiveWorld
	reported  map[RoomID]map[string]RoomID
	discovery Discovery
	current   RoomID
}

type LiveWorld interface {
	Topology
	RoomsNamed(name string) []RoomID
}

func NewLiveNavigation(world LiveWorld, discovery Discovery) *LiveNavigation {
	return &LiveNavigation{world: world, reported: make(map[RoomID]map[string]RoomID), discovery: discovery}
}

func (n *LiveNavigation) SetFallbackWorld(world LiveWorld) {
	n.world = world
}

// RecordExits replaces a room's outgoing edges with the server's complete set.
func (n *LiveNavigation) RecordExits(id RoomID, exits map[string]RoomID) bool {
	if previous, known := n.reported[id]; known && maps.Equal(previous, exits) {
		return false
	}
	n.reported[id] = maps.Clone(exits)
	return true
}

func (n *LiveNavigation) HasRoom(id RoomID) bool {
	_, reported := n.reported[id]
	return reported || n.world != nil && n.world.HasRoom(id)
}

func (n *LiveNavigation) ExitsFrom(id RoomID) map[string]RoomID {
	if exits, reported := n.reported[id]; reported {
		return maps.Clone(exits)
	}
	if n.world != nil {
		return n.world.ExitsFrom(id)
	}
	return nil
}

func (n *LiveNavigation) Neighbors(id RoomID) []RoomID {
	neighbors := make(map[RoomID]struct{})
	for _, to := range n.ExitsFrom(id) {
		neighbors[to] = struct{}{}
	}
	for from, exits := range n.reported {
		for _, to := range exits {
			if to == id {
				neighbors[from] = struct{}{}
				break
			}
		}
	}
	if n.world != nil {
		for _, from := range n.world.Neighbors(id) {
			if _, reported := n.reported[from]; reported {
				continue
			}
			for _, to := range n.world.ExitsFrom(from) {
				if to == id {
					neighbors[from] = struct{}{}
					break
				}
			}
		}
	}
	return slices.Sorted(maps.Keys(neighbors))
}

func (n *LiveNavigation) RoomsNamed(name string) []RoomID {
	if n.world != nil {
		return n.world.RoomsNamed(name)
	}
	return nil
}

// ResolveRoom matches a title-only room block against the immutable world graph.
func (n *LiveNavigation) ResolveRoom(name string, exits map[string]struct{}, direction string) (RoomID, error) {
	candidates := n.RoomsNamed(name)
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	var matches []RoomID
	for _, id := range candidates {
		if maps.EqualFunc(n.ExitsFrom(id), exits, func(RoomID, struct{}) bool { return true }) {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("No room named %s matches these exits in Rooms/", name)
	}

	currentExits := n.ExitsFrom(n.current)
	nearby := map[RoomID]bool{n.current: true}
	for _, id := range currentExits {
		nearby[id] = true
	}
	var near RoomID
	count := 0
	for _, id := range matches {
		if direction != "" && currentExits[direction] == id {
			return id, nil
		}
		if nearby[id] {
			near = id
			count++
		}
	}
	if count == 1 {
		return near, nil
	}
	return "", fmt.Errorf("Several rooms match %s. The map will resume when your location is clear", name)
}

func (n *LiveNavigation) ClearCurrentRoom() {
	n.current = ""
}

func (n *LiveNavigation) ObserveRoom(id RoomID) RoomArrival {
	arrival := RoomArrival{Room: id, From: n.current, Known: n.HasRoom(id)}
	n.current = id
	n.discovery.Discover(id)
	if arrival.Known && arrival.From != "" && arrival.From != id {
		for dir, destination := range n.ExitsFrom(arrival.From) {
			if destination == id && (arrival.Direction == "" || dir < arrival.Direction) {
				arrival.Direction = dir
			}
		}
	}
	return arrival
}
