package main

import "testing"

func testWorld() *World {
	return &World{
		exits: map[RoomID]map[string]RoomID{
			"A": {"E": "B"},
			"B": {"W": "A", "S": "C"},
			"C": {"N": "B"},
		},
	}
}

func TestNavigationSessionStartAndMove(t *testing.T) {
	w := testWorld()
	m := NewMapper()
	d := NewDiscoveryState()

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
	w := testWorld()
	m := NewMapper()
	d := NewDiscoveryState()

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
