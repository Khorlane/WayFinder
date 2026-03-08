package wne

import "testing"

type testWorld struct {
	exits map[RoomID]map[string]RoomID
}

func (w *testWorld) ExitsFrom(roomID RoomID) map[string]RoomID {
	ex := w.exits[roomID]
	out := make(map[string]RoomID, len(ex))
	for dir, id := range ex {
		out[dir] = id
	}
	return out
}

func (w *testWorld) Neighbors(roomID RoomID) []RoomID {
	ex := w.exits[roomID]
	out := make([]RoomID, 0, len(ex))
	for _, id := range ex {
		out = append(out, id)
	}
	return out
}

func (w *testWorld) HasRoom(roomID RoomID) bool {
	_, ok := w.exits[roomID]
	return ok
}

type testMapper struct {
	bound    Topology
	enterLog []RoomID
}

func (m *testMapper) BindTopology(t Topology) {
	m.bound = t
}

func (m *testMapper) Enter(id RoomID, _ string) error {
	m.enterLog = append(m.enterLog, id)
	return nil
}

type testDiscovery struct {
	seen map[RoomID]struct{}
}

func (d *testDiscovery) Discover(roomID RoomID) {
	d.seen[roomID] = struct{}{}
}

func (d *testDiscovery) IsDiscovered(roomID RoomID) bool {
	_, ok := d.seen[roomID]
	return ok
}

func buildWorld() *testWorld {
	return &testWorld{
		exits: map[RoomID]map[string]RoomID{
			"A": {"E": "B"},
			"B": {"W": "A", "S": "C"},
			"C": {"N": "B"},
		},
	}
}

func TestNavigationSessionStartAndMove(t *testing.T) {
	w := buildWorld()
	m := &testMapper{}
	d := &testDiscovery{seen: make(map[RoomID]struct{})}

	s, err := NewNavigationSession(w, m, d, "A")
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if got := s.CurrentRoom(); got != "A" {
		t.Fatalf("current room = %s, want A", got)
	}
	if !d.IsDiscovered("A") {
		t.Fatalf("start room should be discovered")
	}

	if err := s.Move("E"); err != nil {
		t.Fatalf("move E failed: %v", err)
	}
	if got := s.CurrentRoom(); got != "B" {
		t.Fatalf("current room = %s, want B", got)
	}
	if !d.IsDiscovered("B") {
		t.Fatalf("moved room should be discovered")
	}
}

func TestNavigationSessionNoExit(t *testing.T) {
	w := buildWorld()
	m := &testMapper{}
	d := &testDiscovery{seen: make(map[RoomID]struct{})}

	s, err := NewNavigationSession(w, m, d, "A")
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := s.Move("N"); err == nil {
		t.Fatalf("expected no-exit error")
	}
	if got := s.CurrentRoom(); got != "A" {
		t.Fatalf("current room = %s, want A after failed move", got)
	}
}
