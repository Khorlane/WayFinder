package wmr

// RebuildCapture records a live map view or a rebuild diagnostic.
// Room coordinates are [row, column].
type RebuildCapture struct {
	Attempt        uint64                     `json:"attempt"`
	Stage          string                     `json:"stage"`
	Candidate      string                     `json:"candidate,omitempty"`
	From           RoomID                     `json:"from"`
	To             RoomID                     `json:"to"`
	Direction      string                     `json:"direction"`
	Current        RoomID                     `json:"current"`
	Rooms          map[SolverRoomID][2]int    `json:"rooms,omitempty"`
	Connections    []SolverConstraintRelation `json:"connections,omitempty"`
	Links          [][2]RoomID                `json:"links,omitempty"`
	ExitDirections map[RoomID][]string        `json:"exitDirections,omitempty"`
	Title          string                     `json:"title,omitempty"`
	Error          string                     `json:"error,omitempty"`
}

func (m *Mapper) SetRebuildCapture(capture func(RebuildCapture)) {
	m.rebuildCapture = capture
}

func (m *Mapper) placementCoordinates() map[SolverRoomID][2]int {
	coords := make(map[SolverRoomID][2]int, len(m.occ))
	for cell, id := range m.occ {
		coords[SolverRoomID(id)] = cell
	}
	return coords
}
