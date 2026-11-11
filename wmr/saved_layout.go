package wmr

import (
	"fmt"
	"sort"
)

// SavedLayout stores the grid projection and its locked connections.
// The live server establishes the current room on each connection.
type SavedLayout struct {
	Rooms map[SolverRoomID][2]int `json:"rooms"`
	Locks []SolverLockedAdjKey    `json:"locks"`
}

func (m *Mapper) SaveLayout() SavedLayout {
	layout := SavedLayout{Rooms: m.placementCoordinates()}
	for key := range m.locks {
		layout.Locks = append(layout.Locks, SolverLockedAdjKey{
			From: SolverRoomID(key.From), To: SolverRoomID(key.To), Dir: key.Dir,
		})
	}
	sort.Slice(layout.Locks, func(i, j int) bool {
		left, right := layout.Locks[i], layout.Locks[j]
		if left.From != right.From {
			return left.From < right.From
		}
		if left.Dir != right.Dir {
			return left.Dir < right.Dir
		}
		return left.To < right.To
	})
	return layout
}

// RestoreLayout validates a saved projection against the restored discovery and
// topology, then installs it with no current room.
func (m *Mapper) RestoreLayout(layout SavedLayout, discovery discoveredView) error {
	restored := NewMapper()
	for id, cell := range layout.Rooms {
		room := RoomID(id)
		if room == "" || !discovery.IsDiscovered(room) {
			return fmt.Errorf("saved layout contains unvisited room %q", id)
		}
		if other, occupied := restored.occ[cell]; occupied {
			return fmt.Errorf("saved rooms %s and %s overlap", other, room)
		}
		restored.setOcc(room, cell[0], cell[1])
	}
	for _, lock := range layout.Locks {
		from, to := RoomID(lock.From), RoomID(lock.To)
		_, _, planar := dirDelta(lock.Dir)
		if !planar || m.topo == nil || m.topo.ExitsFrom(from)[lock.Dir] != to {
			return fmt.Errorf("saved connection %s -%s-> %s is missing from topology", from, lock.Dir, to)
		}
		restored.locks[lockedAdjKey{From: from, To: to, Dir: lock.Dir}] = struct{}{}
	}
	constraints := restored.toSolverConstraintSet(restored.BuildConstraintSet())
	if err := restored.solver().ValidateConstraintSet(constraints, func(id SolverRoomID) (int, int, bool) {
		cell, ok := layout.Rooms[id]
		return cell[0], cell[1], ok
	}); err != nil {
		return err
	}
	if err := validateConnections(constraints, layout.Rooms); err != nil {
		return err
	}
	m.rooms, m.occ, m.locks, m.cur = restored.rooms, restored.occ, restored.locks, nil
	return nil
}
