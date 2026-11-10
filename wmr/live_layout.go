package wmr

import "sort"

// Layout is a detached snapshot for a view of discovered rooms.
type Layout struct {
	Rooms   []Room
	Links   [][2]RoomID
	Current RoomID
}

// PlaceArrival projects a server-confirmed visit using the existing mapper.
func (m *Mapper) PlaceArrival(id, from RoomID, direction string) error {
	room := m.getRoom(id)
	if anchor := m.rooms[from]; direction != "" && anchor != nil && anchor.Placed {
		return m.placeFrom(room, anchor, direction)
	}
	if room.Placed {
		m.cur = room
		return nil
	}
	if len(m.occ) == 0 {
		return m.Enter(id, "START")
	}
	if m.topo != nil {
		// A return, recall, or teleport can land beside a different placed room.
		exits := m.topo.ExitsFrom(id)
		compass := [...]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
		for _, neighbor := range m.topo.Neighbors(id) {
			anchor := m.rooms[neighbor]
			if anchor == nil || !anchor.Placed {
				continue
			}
			incoming := m.topo.ExitsFrom(neighbor)
			for i, dir := range compass {
				if incoming[dir] == id {
					return m.placeFrom(room, anchor, dir)
				}
				if exits[dir] == neighbor {
					// The opposite compass direction gives the same spatial relation.
					return m.placeFrom(room, anchor, compass[(i+len(compass)/2)%len(compass)])
				}
			}
		}
	}
	// Leave a gap beside the existing map when there is no placed neighbor.
	rightmost := 0
	for cell := range m.occ {
		rightmost = max(rightmost, cell[1])
	}
	m.setOcc(id, 1, rightmost+2)
	m.cur = room
	return nil
}

func (m *Mapper) placeFrom(room, anchor *Room, dir string) error {
	before := m.captureSnapshot()
	previous := m.cur
	m.cur = anchor
	err := m.Enter(room.ID, dir)
	if err == nil {
		err = m.validateLockedAdjacencies(func(id RoomID) (int, int, bool) {
			if placed := m.rooms[id]; placed != nil && placed.Placed {
				return placed.R, placed.C, true
			}
			return 0, 0, false
		})
	}
	if err != nil {
		m.restoreSnapshot(before)
		m.cur = previous
	}
	return err
}

func (m *Mapper) SnapshotLayout(discovery discoveredView, current RoomID) Layout {
	layout := Layout{Current: current}
	visible := make(map[RoomID]bool, len(m.occ))
	for _, id := range m.occ {
		if discovery.IsDiscovered(id) {
			visible[id] = true
			layout.Rooms = append(layout.Rooms, *m.rooms[id])
		}
	}
	sort.Slice(layout.Rooms, func(i, j int) bool { return layout.Rooms[i].ID < layout.Rooms[j].ID })
	links := make(map[[2]RoomID]bool)
	for lock := range m.locks {
		if !visible[lock.From] || !visible[lock.To] {
			continue
		}
		link := [2]RoomID{min(lock.From, lock.To), max(lock.From, lock.To)}
		if !links[link] {
			links[link] = true
			layout.Links = append(layout.Links, link)
		}
	}
	return layout
}
