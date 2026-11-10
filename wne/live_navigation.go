package wne

import (
	"fmt"
	"maps"
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
	discovery Discovery
	current   RoomID
}

type LiveWorld interface {
	World
	RoomsNamed(name string) []RoomID
}

func NewLiveNavigation(world LiveWorld, discovery Discovery) *LiveNavigation {
	return &LiveNavigation{world: world, discovery: discovery}
}

// ResolveRoom matches a title-only room block against the immutable world graph.
func (n *LiveNavigation) ResolveRoom(name string, exits map[string]struct{}, direction string) (RoomID, error) {
	candidates := n.world.RoomsNamed(name)
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	var matches []RoomID
	for _, id := range candidates {
		if maps.EqualFunc(n.world.ExitsFrom(id), exits, func(RoomID, struct{}) bool { return true }) {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("No room named %s matches these exits in Rooms/", name)
	}

	currentExits := n.world.ExitsFrom(n.current)
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
	arrival := RoomArrival{Room: id, From: n.current, Known: n.world.HasRoom(id)}
	n.current = id
	n.discovery.Discover(id)
	if arrival.Known && arrival.From != "" && arrival.From != id {
		for dir, destination := range n.world.ExitsFrom(arrival.From) {
			if destination == id && (arrival.Direction == "" || dir < arrival.Direction) {
				arrival.Direction = dir
			}
		}
	}
	return arrival
}
