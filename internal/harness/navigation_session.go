package harness

import "fmt"

// Navigator is the adapter-facing contract used by CLI/MUD frontends.
type Navigator interface {
	CurrentRoom() RoomID
	CurrentExits() map[string]RoomID
	Move(dir string) error
	Mapper() *Mapper
	Discovery() *DiscoveryState
}

// NavigationSession is the adapter boundary between an event source (CLI/MUD)
// and the mapping core. It owns current room/discovery progression.
type NavigationSession struct {
	world     *World
	mapper    *Mapper
	discovery *DiscoveryState
	cur       RoomID
}

func NewNavigationSession(world *World, mapper *Mapper, discovery *DiscoveryState, start RoomID) (*NavigationSession, error) {
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
	mapper.BindTopology(world)
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

func (s *NavigationSession) Mapper() *Mapper {
	return s.mapper
}

func (s *NavigationSession) Discovery() *DiscoveryState {
	return s.discovery
}
