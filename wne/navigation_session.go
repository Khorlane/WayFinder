package wne

import "fmt"

type RoomID string

type Topology interface {
	ExitsFrom(roomID RoomID) map[string]RoomID
	Neighbors(roomID RoomID) []RoomID
	HasRoom(roomID RoomID) bool
}

type World interface {
	ExitsFrom(roomID RoomID) map[string]RoomID
	HasRoom(roomID RoomID) bool
}

type Mapper interface {
	BindTopology(t Topology)
	Enter(id RoomID, dirMoved string) error
}

type Discovery interface {
	Discover(roomID RoomID)
}

// Navigator is the adapter-facing contract used by CLI/MUD frontends.
type Navigator interface {
	CurrentRoom() RoomID
	CurrentExits() map[string]RoomID
	Move(dir string) error
}

// NavigationSession is the adapter boundary between an event source (CLI/MUD)
// and the mapping core. It owns current room/discovery progression.
// WNE maintains discovered navigation topology and consumes already-parsed
// movement intent/events from upstream layers (for example WEG).
type NavigationSession struct {
	world     World
	mapper    Mapper
	discovery Discovery
	cur       RoomID
}

func NewNavigationSession(world World, mapper Mapper, discovery Discovery, start RoomID) (*NavigationSession, error) {
	if world == nil {
		return nil, fmt.Errorf("world is nil")
	}
	if mapper == nil {
		return nil, fmt.Errorf("mapper is nil")
	}
	if discovery == nil {
		return nil, fmt.Errorf("discovery is nil")
	}
	if !world.HasRoom(start) {
		return nil, fmt.Errorf("start room %s not found in world", start)
	}

	topo, ok := world.(Topology)
	if !ok {
		return nil, fmt.Errorf("world does not satisfy topology contract")
	}
	mapper.BindTopology(topo)
	if err := mapper.Enter(start, "START"); err != nil {
		return nil, fmt.Errorf("mapper start error: %w", err)
	}
	discovery.Discover(start)
	return &NavigationSession{
		world:     world,
		mapper:    mapper,
		discovery: discovery,
		cur:       start,
	}, nil
}

func (s *NavigationSession) CurrentRoom() RoomID {
	return s.cur
}

func (s *NavigationSession) CurrentExits() map[string]RoomID {
	return s.world.ExitsFrom(s.cur)
}

func (s *NavigationSession) Move(dir string) error {
	next, ok := s.world.ExitsFrom(s.cur)[dir]
	if !ok {
		return fmt.Errorf("no exit that way")
	}
	if err := s.mapper.Enter(next, dir); err != nil {
		return err
	}
	s.discovery.Discover(next)
	s.cur = next
	return nil
}
