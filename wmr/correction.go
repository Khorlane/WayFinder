package wmr

import (
	"errors"
	"fmt"
	"slices"
)

type rebuildFailure struct {
	err         error
	attempt     uint64
	from, to    RoomID
	direction   string
	constraints SolverConstraintSet
}

func (f *rebuildFailure) Error() string { return f.err.Error() }

// Correction is a detached proposal for editing the live layout.
type Correction struct {
	Draft  Layout
	Submit func([]Room) (Layout, error)
	Close  func()
}

// LayoutEditing exposes automatic recovery and manual editing for a map view.
type LayoutEditing struct {
	Failure *Correction
	Begin   func() *Correction
}

// NewCorrection prepares a draft and applies it to the mapper only after validation.
func (m *Mapper) NewCorrection(layout Layout, discovery discoveredView, placementErr error, write func(RebuildCapture) error) *Correction {
	var failure *rebuildFailure
	if !errors.As(placementErr, &failure) {
		return nil
	}
	draft := Layout{
		Rooms:          slices.Clone(layout.Rooms),
		Links:          slices.Clone(layout.Links),
		ExitDirections: make(map[RoomID][]string, len(layout.ExitDirections)+1),
		Current:        failure.to,
	}
	for id, directions := range layout.ExitDirections {
		draft.ExitDirections[id] = slices.Clone(directions)
	}
	dr, dc, _ := dirDelta(failure.direction)
	if !slices.ContainsFunc(draft.Rooms, func(room Room) bool { return room.ID == failure.to }) {
		anchor := m.rooms[failure.from]
		cell := [2]int{anchor.R + dr, anchor.C + dc}
		for m.occ[cell] != "" {
			cell[0] += dr
			cell[1] += dc
		}
		draft.Rooms = append(draft.Rooms, Room{ID: failure.to, Placed: true, R: cell[0], C: cell[1]})
		if m.topo != nil {
			for dir := range m.topo.ExitsFrom(failure.to) {
				draft.ExitDirections[failure.to] = append(draft.ExitDirections[failure.to], dir)
			}
			slices.Sort(draft.ExitDirections[failure.to])
		}
	}
	link := [2]RoomID{min(failure.from, failure.to), max(failure.from, failure.to)}
	if !slices.Contains(draft.Links, link) {
		draft.Links = append(draft.Links, link)
	}

	constraints := failure.constraints
	move := requiredMove(SolverRoomID(failure.from), SolverRoomID(failure.to), failure.direction, dr, dc)
	constraints.Relations = append(slices.Clone(constraints.Relations), move)
	return m.newEdit(draft, discovery, failure.to, constraints, RebuildCapture{
		Attempt: failure.attempt, Stage: "correction", From: failure.from, To: failure.to, Direction: failure.direction,
	}, write)
}

// NewMapEdit prepares a draft of the current map using its existing constraints.
func (m *Mapper) NewMapEdit(layout Layout, discovery discoveredView, write func(RebuildCapture) error) *Correction {
	var current RoomID
	if m.cur != nil {
		current = m.cur.ID
	}
	return m.newEdit(layout, discovery, current, m.toSolverConstraintSet(m.BuildConstraintSet()), RebuildCapture{
		Stage: "edit", From: layout.Current, To: layout.Current,
	}, write)
}

func (m *Mapper) newEdit(draft Layout, discovery discoveredView, current RoomID, constraints SolverConstraintSet, capture RebuildCapture, write func(RebuildCapture) error) *Correction {
	reference := NewMapper()
	for _, room := range draft.Rooms {
		reference.getRoom(room.ID)
	}
	solver := reference.solver()
	submit := func(rooms []Room) (Layout, error) {
		coords := make(map[SolverRoomID][2]int, len(rooms))
		for _, room := range rooms {
			coords[SolverRoomID(room.ID)] = [2]int{room.R, room.C}
		}
		validationErr := validateCorrection(reference, solver, constraints, coords)
		entry := capture
		entry.Current, entry.Rooms, entry.Connections = draft.Current, coords, constraints.Relations
		entry.Links, entry.ExitDirections = draft.Links, draft.ExitDirections
		if validationErr != nil {
			entry.Error = validationErr.Error()
		}
		var applied Layout
		if validationErr == nil {
			result := RebuildResult{
				Rooms:   make(map[SolverRoomID]RebuildRoomState, len(coords)),
				Occ:     make(map[[2]int]SolverRoomID, len(coords)),
				Current: SolverRoomID(current),
			}
			for id, cell := range coords {
				result.Rooms[id] = RebuildRoomState{Placed: true, R: cell[0], C: cell[1]}
				result.Occ[cell] = id
			}
			m.applyRebuildResult(result)
			applied = m.SnapshotLayout(discovery, draft.Current)
			entry.Links, entry.ExitDirections = applied.Links, applied.ExitDirections
		}
		if write != nil {
			_ = write(entry)
		}
		return applied, validationErr
	}
	return &Correction{Draft: draft, Submit: submit}
}

func validateCorrection(reference *Mapper, solver SolverEngine, constraints SolverConstraintSet, coords map[SolverRoomID][2]int) error {
	if len(coords) != len(reference.rooms) {
		return errors.New("The correction must place all rooms from this attempt.")
	}
	occupied := make(map[[2]int]RoomID, len(coords))
	for id := range reference.rooms {
		cell, ok := coords[SolverRoomID(id)]
		if !ok {
			return fmt.Errorf("The correction is missing room %s.", id)
		}
		if other := occupied[cell]; other != "" {
			return fmt.Errorf("Rooms %s and %s overlap.", other, id)
		}
		occupied[cell] = id
	}
	if cell, pinned := coords["1"]; pinned && cell != [2]int{1, 0} {
		return errors.New("Room 1 must stay at its starting position.")
	}
	if err := solver.ValidateConstraintSet(constraints, func(id SolverRoomID) (int, int, bool) {
		cell, placed := coords[id]
		return cell[0], cell[1], placed
	}); err != nil {
		return err
	}
	return validateConnections(constraints, coords)
}
