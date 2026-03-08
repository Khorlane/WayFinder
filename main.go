// *******************************************************
// Wayfinder — Discovering the world one room at a time. *
// *******************************************************

// Project architecture documentation:
// docs/architecture.md
// docs/mapper_rules.md
// docs/discovery_model.md

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	wcswin32 "WayFinder/internal/wcs/win32"
	"WayFinder/solver"
)

type RoomID string

type Room struct {
	ID     RoomID
	Placed bool
	R, C   int // Unbounded integer grid coordinates.
}

type Mapper struct {
	rooms          map[RoomID]*Room
	occ            map[[2]int]RoomID
	cur            *Room
	topo           Topology
	locks          map[lockedAdjKey]struct{} // Monotonic set of discovered adjacencies that must remain satisfied.
	debug          io.Writer
	solverProvider solver.SolverProvider
}

type lockedAdjKey struct {
	From RoomID
	To   RoomID
	Dir  string
}

type ConstraintRelation struct {
	Key           lockedAdjKey
	Locked        bool
	Enforced      bool
	AxisAligned   bool
	SameRow       bool
	SameColumn    bool
	RequiresOrder bool
	NoRoomBetween bool
}

type ConstraintSet struct {
	Discovered map[RoomID]struct{}
	Relations  []ConstraintRelation
}

func relationForKey(k lockedAdjKey) ConstraintRelation {
	rel := ConstraintRelation{
		Key:           k,
		RequiresOrder: true,
	}
	switch k.Dir {
	case "E", "W":
		rel.AxisAligned = true
		rel.SameRow = true
		rel.NoRoomBetween = true
	case "N", "S":
		rel.AxisAligned = true
		rel.SameColumn = true
		rel.NoRoomBetween = true
	}
	return rel
}

func (m *Mapper) BuildConstraintSet(extraDiscovered ...RoomID) ConstraintSet {
	cs := ConstraintSet{
		Discovered: make(map[RoomID]struct{}),
		Relations:  make([]ConstraintRelation, 0, len(m.locks)),
	}
	for id, r := range m.rooms {
		if r != nil && r.Placed {
			cs.Discovered[id] = struct{}{}
		}
	}
	for _, id := range extraDiscovered {
		if id == "" {
			continue
		}
		cs.Discovered[id] = struct{}{}
	}

	byKey := make(map[lockedAdjKey]ConstraintRelation)
	if m.topo != nil {
		for from := range cs.Discovered {
			exits := m.topo.ExitsFrom(from)
			for _, dir := range []string{"N", "E", "S", "W"} {
				to, ok := exits[dir]
				if !ok {
					continue
				}
				if _, ok := cs.Discovered[to]; !ok {
					continue
				}
				k := lockedAdjKey{From: from, To: to, Dir: dir}
				rel := relationForKey(k)
				rel.NoRoomBetween = false
				byKey[k] = rel
			}
		}
	}
	for k := range m.locks {
		rel := relationForKey(k)
		rel.Locked = true
		rel.Enforced = true
		byKey[k] = rel
	}
	for _, rel := range byKey {
		cs.Relations = append(cs.Relations, rel)
	}
	sort.Slice(cs.Relations, func(i, j int) bool {
		li := cs.Relations[i].Key
		lj := cs.Relations[j].Key
		if li.From != lj.From {
			return li.From < lj.From
		}
		if li.Dir != lj.Dir {
			return li.Dir < lj.Dir
		}
		return li.To < lj.To
	})
	return cs
}

type lockedAdjViolationError struct {
	Key       lockedAdjKey
	ExpectedR int
	ExpectedC int
}

func (e *lockedAdjViolationError) Error() string {
	return fmt.Sprintf("mapping invariant: candidate move would break locked adjacency %s -%s-> %s",
		e.Key.From, e.Key.Dir, e.Key.To)
}

type collisionError struct {
	Row      int
	Col      int
	Occupant RoomID
	Moving   RoomID
}

func (e *collisionError) Error() string {
	return fmt.Sprintf("occupancy collision at (R=%d,C=%s): moving %s into anchored room %s",
		e.Row, colName(e.Col), e.Moving, e.Occupant)
}

type roomSnapshot struct {
	Placed bool
	R      int
	C      int
}

type mapperSnapshot struct {
	rooms map[RoomID]roomSnapshot
	occ   map[[2]int]RoomID
	locks map[lockedAdjKey]struct{}
}

var uiOut io.Writer = os.Stdout

func uiPrint(a ...any) {
	fmt.Fprint(uiOut, a...)
}

func uiPrintf(format string, a ...any) {
	fmt.Fprintf(uiOut, format, a...)
}

func uiPrintln(a ...any) {
	fmt.Fprintln(uiOut, a...)
}

func setupLogging(logPath string) (func(), error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	origStdout := os.Stdout
	os.Stdout = f
	uiOut = io.MultiWriter(origStdout, f)
	fmt.Fprintf(f, "\n=== WayFinder run %s ===\n", time.Now().Format(time.RFC3339))
	return func() {
		os.Stdout = origStdout
		_ = f.Close()
	}, nil
}

func NewMapper() *Mapper {
	m := &Mapper{
		rooms:          make(map[RoomID]*Room),
		occ:            make(map[[2]int]RoomID),
		locks:          make(map[lockedAdjKey]struct{}),
		debug:          io.Discard,
		solverProvider: solver.DefaultSolverProvider,
	}
	return m
}

func (m *Mapper) BindTopology(t Topology) {
	m.topo = t
}

func (m *Mapper) SetDebugWriter(w io.Writer) {
	if w == nil {
		m.debug = io.Discard
		return
	}
	m.debug = w
}

func (m *Mapper) SetSolverProvider(p solver.SolverProvider) {
	if p == nil {
		m.solverProvider = solver.DefaultSolverProvider
		return
	}
	m.solverProvider = p
}

func (m *Mapper) debugln(a ...any) {
	fmt.Fprintln(m.debug, a...)
}

func (m *Mapper) debugf(format string, a ...any) {
	fmt.Fprintf(m.debug, format, a...)
}

func colName(c int) string { return strconv.Itoa(c) }

func cellLabel(id RoomID) string {
	// Debug-only: stable 3-hex label derived from RoomID (no new fields needed).
	// (FNV-1a, truncated to 12 bits)
	s := string(id)
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	v := int(h & 0xFFF)
	return fmt.Sprintf("%03X", v)
}

func normalizeDirName(s string) string {
	t := strings.TrimSpace(strings.ToUpper(s))
	switch t {
	case "N", "NORTH":
		return "N"
	case "S", "SOUTH":
		return "S"
	case "E", "EAST":
		return "E"
	case "W", "WEST":
		return "W"

	case "NE", "NORTHEAST", "NORTH-EAST", "NORTH EAST":
		return "NE"
	case "NW", "NORTHWEST", "NORTH-WEST", "NORTH WEST":
		return "NW"
	case "SE", "SOUTHEAST", "SOUTH-EAST", "SOUTH EAST":
		return "SE"
	case "SW", "SOUTHWEST", "SOUTH-WEST", "SOUTH WEST":
		return "SW"

	default:
		return ""
	}
}

func dirDelta(dir string) (dr, dc int, ok bool) {
	switch dir {
	case "N":
		return -1, 0, true
	case "S":
		return 1, 0, true
	case "E":
		return 0, 1, true
	case "W":
		return 0, -1, true

	case "NE":
		return -1, 1, true
	case "NW":
		return -1, -1, true
	case "SE":
		return 1, 1, true
	case "SW":
		return 1, -1, true

	default:
		return 0, 0, false
	}
}

func (m *Mapper) getRoom(id RoomID) *Room {
	r := m.rooms[id]
	if r == nil {
		r = &Room{ID: id}
		m.rooms[id] = r
	}
	return r
}

func (m *Mapper) clearOcc(id RoomID) {
	room := m.rooms[id]
	if room == nil || !room.Placed {
		return
	}
	delete(m.occ, [2]int{room.R, room.C})
}

func (m *Mapper) setOcc(id RoomID, r, c int) {
	rm := m.getRoom(id)

	// If it was already placed somewhere else, remove the old occupancy entry.
	if rm.Placed {
		delete(m.occ, [2]int{rm.R, rm.C})
	}

	// If target cell is occupied by some *other* room, that's a bug in caller logic.
	if occ, ok := m.occ[[2]int{r, c}]; ok && occ != id {
		panic(fmt.Sprintf("setOcc collision at (R=%d,C=%s): trying to place %s but %s is there",
			r, colName(c), string(id), string(occ)))
	}

	rm.R, rm.C = r, c
	rm.Placed = true
	m.occ[[2]int{r, c}] = id

	m.refreshLockedAdjacencies()
}

func edgeAlignedAndOrdered(fromR, fromC, toR, toC int, dir string) bool {
	switch dir {
	case "N":
		return toC == fromC && toR < fromR
	case "S":
		return toC == fromC && toR > fromR
	case "E":
		return toR == fromR && toC > fromC
	case "W":
		return toR == fromR && toC < fromC
	case "NE":
		return toR < fromR && toC > fromC
	case "NW":
		return toR < fromR && toC < fromC
	case "SE":
		return toR > fromR && toC > fromC
	case "SW":
		return toR > fromR && toC < fromC
	default:
		return false
	}
}

func roomBetweenAxis(fromR, fromC, toR, toC, r, c int) bool {
	if fromR == toR && r == fromR {
		lo, hi := fromC, toC
		if lo > hi {
			lo, hi = hi, lo
		}
		return c > lo && c < hi
	}
	if fromC == toC && c == fromC {
		lo, hi := fromR, toR
		if lo > hi {
			lo, hi = hi, lo
		}
		return r > lo && r < hi
	}
	return false
}

func (m *Mapper) noRoomBetweenAxis(coordAfter func(RoomID) (int, int, bool), fromID, toID RoomID, fromR, fromC, toR, toC int) bool {
	// "Between" is defined only for same-row or same-column constraints.
	if fromR != toR && fromC != toC {
		return true
	}
	for id, rm := range m.rooms {
		if rm == nil || !rm.Placed || id == fromID || id == toID {
			continue
		}
		r, c, ok := coordAfter(id)
		if !ok {
			continue
		}
		if roomBetweenAxis(fromR, fromC, toR, toC, r, c) {
			return false
		}
	}
	return true
}

func (m *Mapper) refreshLockedAdjacencies() {
	if m.topo == nil {
		return
	}
	for from, fromRm := range m.rooms {
		if fromRm == nil || !fromRm.Placed {
			continue
		}
		exits := m.topo.ExitsFrom(from)
		for dir, to := range exits {
			toRm := m.rooms[to]
			if toRm == nil || !toRm.Placed {
				continue
			}
			if edgeAlignedAndOrdered(fromRm.R, fromRm.C, toRm.R, toRm.C, dir) &&
				m.noRoomBetweenAxis(
					func(id RoomID) (int, int, bool) {
						rm := m.rooms[id]
						if rm == nil || !rm.Placed {
							return 0, 0, false
						}
						return rm.R, rm.C, true
					},
					from, to, fromRm.R, fromRm.C, toRm.R, toRm.C,
				) {
				m.locks[lockedAdjKey{From: from, To: to, Dir: dir}] = struct{}{}
			}
		}
	}
}

func (m *Mapper) validateLockedAdjacencies(coordAfter func(RoomID) (int, int, bool)) error {
	return m.validateConstraintSet(m.BuildConstraintSet(), coordAfter)
}

func (m *Mapper) validateConstraintSet(cs ConstraintSet, coordAfter func(RoomID) (int, int, bool)) error {
	solverCoord := func(id solver.RoomID) (int, int, bool) {
		return coordAfter(RoomID(id))
	}
	err := m.solver().ValidateConstraintSet(m.toSolverConstraintSet(cs), solverCoord)
	var sLockErr *solver.LockedAdjViolationError
	if errors.As(err, &sLockErr) {
		return &lockedAdjViolationError{
			Key: lockedAdjKey{
				From: RoomID(sLockErr.Key.From),
				To:   RoomID(sLockErr.Key.To),
				Dir:  sLockErr.Key.Dir,
			},
			ExpectedR: sLockErr.ExpectedR,
			ExpectedC: sLockErr.ExpectedC,
		}
	}
	return err
}

func (m *Mapper) solverContext() solver.SolverContext {
	rooms := make(map[solver.RoomID]solver.SolverRoomState, len(m.rooms))
	for id, r := range m.rooms {
		if r == nil {
			continue
		}
		rooms[solver.RoomID(id)] = solver.SolverRoomState{
			Placed: r.Placed,
			R:      r.R,
			C:      r.C,
		}
	}
	return solver.SolverContext{
		Rooms: rooms,
		NoRoomBetweenAxis: func(coordAfter func(solver.RoomID) (int, int, bool), fromID, toID solver.RoomID, fromR, fromC, toR, toC int) bool {
			mainCoord := func(id RoomID) (int, int, bool) {
				return coordAfter(solver.RoomID(id))
			}
			return m.noRoomBetweenAxis(mainCoord, RoomID(fromID), RoomID(toID), fromR, fromC, toR, toC)
		},
		EdgeAlignedAndOrdered: edgeAlignedAndOrdered,
		DirDelta:              dirDelta,
		ColName:               colName,
		Debugln:               m.debugln,
		Debugf:                m.debugf,
	}
}

func (m *Mapper) solver() solver.SolverEngine {
	provider := m.solverProvider
	if provider == nil {
		provider = solver.DefaultSolverProvider
	}
	return provider(m.solverContext())
}

func (m *Mapper) toSolverConstraintSet(cs ConstraintSet) solver.ConstraintSet {
	out := solver.ConstraintSet{
		Discovered: make(map[solver.RoomID]struct{}, len(cs.Discovered)),
		Relations:  make([]solver.ConstraintRelation, 0, len(cs.Relations)),
	}
	for id := range cs.Discovered {
		out.Discovered[solver.RoomID(id)] = struct{}{}
	}
	for _, rel := range cs.Relations {
		out.Relations = append(out.Relations, solver.ConstraintRelation{
			Key: solver.LockedAdjKey{
				From: solver.RoomID(rel.Key.From),
				To:   solver.RoomID(rel.Key.To),
				Dir:  rel.Key.Dir,
			},
			Locked:        rel.Locked,
			Enforced:      rel.Enforced,
			AxisAligned:   rel.AxisAligned,
			SameRow:       rel.SameRow,
			SameColumn:    rel.SameColumn,
			RequiresOrder: rel.RequiresOrder,
			NoRoomBetween: rel.NoRoomBetween,
		})
	}
	return out
}

func (m *Mapper) shiftWhere(pred func(*Room) bool, dr, dc int) error {
	if dr == 0 && dc == 0 {
		return nil
	}

	// Phase 1: select rooms to move (predicate evaluated BEFORE any moves).
	var move []*Room
	block := make(map[RoomID]struct{})
	for _, r := range m.rooms {
		if r == nil || !r.Placed || !pred(r) {
			continue
		}
		move = append(move, r)
		block[r.ID] = struct{}{}
	}
	if len(move) == 0 {
		return nil
	}

	m.refreshLockedAdjacencies()

	coordAfter := func(id RoomID) (int, int, bool) {
		rm := m.rooms[id]
		if rm == nil || !rm.Placed {
			return 0, 0, false
		}
		if _, inBlock := block[id]; inBlock {
			return rm.R + dr, rm.C + dc, true
		}
		return rm.R, rm.C, true
	}

	// Validate occupancy collisions against rooms that will stay anchored.
	for _, r := range move {
		nr, nc := r.R+dr, r.C+dc
		if occ, ok := m.occ[[2]int{nr, nc}]; ok {
			if _, inBlock := block[occ]; !inBlock {
				return &collisionError{
					Row:      nr,
					Col:      nc,
					Occupant: occ,
					Moving:   r.ID,
				}
			}
		}
	}

	// Any candidate shift must preserve every locked discovered adjacency.
	if err := m.validateConstraintSet(m.BuildConstraintSet(), coordAfter); err != nil {
		return err
	}

	// Phase 2: apply shifts.
	for _, r := range move {
		r.R += dr
		r.C += dc
	}

	// Phase 3: rebuild occupancy from scratch.
	m.occ = make(map[[2]int]RoomID)
	for _, r := range m.rooms {
		if r == nil || !r.Placed {
			continue
		}
		key := [2]int{r.R, r.C}
		if prev, ok := m.occ[key]; ok && prev != r.ID {
			return fmt.Errorf("occupancy collision at (R=%d,C=%s): %s vs %s",
				r.R, colName(r.C), string(prev), string(r.ID))
		}
		m.occ[key] = r.ID
	}

	m.refreshLockedAdjacencies()

	return nil
}

func blockKey(block map[RoomID]struct{}) string {
	ids := make([]string, 0, len(block))
	for id := range block {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func formatBlockRooms(block map[RoomID]struct{}) string {
	ids := make([]string, 0, len(block))
	for id := range block {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	return "[" + strings.Join(ids, " ") + "]"
}

func cloneBlock(block map[RoomID]struct{}) map[RoomID]struct{} {
	out := make(map[RoomID]struct{}, len(block))
	for id := range block {
		out[id] = struct{}{}
	}
	return out
}

func (m *Mapper) printRejection(err error) {
	var lockErr *lockedAdjViolationError
	var collErr *collisionError
	switch {
	case errors.As(err, &lockErr):
		m.debugln("REJECTED")
		m.debugln("reason=locked_adjacency")
		m.debugf("adjacency=%s -%s-> %s\n", lockErr.Key.From, lockErr.Key.Dir, lockErr.Key.To)
		m.debugf("expectedRelation=(%d,%d)\n", lockErr.ExpectedR, lockErr.ExpectedC)
	case errors.As(err, &collErr):
		m.debugln("REJECTED")
		m.debugln("reason=collision")
		m.debugf("cell=(%d,%d)\n", collErr.Row, collErr.Col)
		m.debugf("occupant=%s\n", collErr.Occupant)
	default:
		m.debugln("REJECTED")
		m.debugf("reason=%v\n", err)
	}
}

func (m *Mapper) destinationCandidateBlocks(targetID, anchorID RoomID, allowAnchor bool) ([]map[RoomID]struct{}, error) {
	if m.topo == nil {
		return nil, fmt.Errorf("mapping error: topology not bound")
	}
	target := m.rooms[targetID]
	if target == nil || !target.Placed {
		return nil, fmt.Errorf("mapping error: target %s is not placed", targetID)
	}
	if targetID == anchorID {
		return nil, fmt.Errorf("mapping error: target %s is the anchor room", targetID)
	}

	reachable := make(map[RoomID]struct{})
	q := []RoomID{targetID}
	reachable[targetID] = struct{}{}

	for len(q) > 0 {
		id := q[0]
		q = q[1:]
		for _, nb := range m.topo.Neighbors(id) {
			if !allowAnchor && nb == anchorID {
				continue
			}
			rm := m.rooms[nb]
			if rm == nil || !rm.Placed {
				continue
			}
			if _, seen := reachable[nb]; seen {
				continue
			}
			reachable[nb] = struct{}{}
			q = append(q, nb)
		}
	}

	const maxQueuedCandidates = 1024
	const maxReturnedCandidates = 512

	seen := make(map[string]struct{})
	var queue []map[RoomID]struct{}

	start := map[RoomID]struct{}{targetID: {}}
	queue = append(queue, start)
	seen[blockKey(start)] = struct{}{}

	var out []map[RoomID]struct{}
	for i := 0; i < len(queue) && len(out) < maxReturnedCandidates; i++ {
		cur := queue[i]
		out = append(out, cur)

		frontierSet := make(map[RoomID]struct{})
		for id := range cur {
			for _, nb := range m.topo.Neighbors(id) {
				if !allowAnchor && nb == anchorID {
					continue
				}
				if _, ok := reachable[nb]; !ok {
					continue
				}
				if _, in := cur[nb]; in {
					continue
				}
				frontierSet[nb] = struct{}{}
			}
		}

		var frontier []string
		for id := range frontierSet {
			frontier = append(frontier, string(id))
		}
		sort.Strings(frontier)

		for _, s := range frontier {
			if len(queue) >= maxQueuedCandidates {
				break
			}
			nb := RoomID(s)
			next := cloneBlock(cur)
			next[nb] = struct{}{}
			k := blockKey(next)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			queue = append(queue, next)
		}
	}

	// Always include the maximal destination-side component.
	full := make(map[RoomID]struct{}, len(reachable))
	for id := range reachable {
		full[id] = struct{}{}
	}
	// If the search cap trimmed it, force-add full component as last resort.
	if _, ok := seen[blockKey(full)]; !ok {
		out = append(out, full)
	}

	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) < len(out[j])
		}
		return blockKey(out[i]) < blockKey(out[j])
	})
	return out, nil
}

func (m *Mapper) moveDestinationWithCandidates(targetID, anchorID RoomID, drMove, dcMove, expR, expC int) error {
	target := m.rooms[targetID]
	curR, curC := 0, 0
	if target != nil && target.Placed {
		curR, curC = target.R, target.C
	}
	anchor := m.rooms[anchorID]
	if anchor == nil || !anchor.Placed {
		return fmt.Errorf("mapping error: current anchor %s is not placed", anchorID)
	}
	// Preserve move relation even if an anchor-crossing candidate shifts the current side.
	relDr, relDc := expR-anchor.R, expC-anchor.C

	m.debugln("CORRECTION START")
	m.debugf("currentRoom=%s\n", anchorID)
	m.debugf("destinationRoom=%s\n", targetID)
	m.debugf("expectedCoords=(%d,%d)\n", expR, expC)
	m.debugf("currentCoords=(%d,%d)\n", curR, curC)

	var failures []string
	tried := 0
	seen := make(map[string]struct{})

	trySet := func(candidates []map[RoomID]struct{}, phase string, requireAnchor bool) error {
		for _, block := range candidates {
			if requireAnchor {
				if _, ok := block[anchorID]; !ok {
					continue
				}
			}
			k := blockKey(block)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			tried++

			m.debugln("CANDIDATE BLOCK")
			m.debugf("blockIndex=%d\n", tried)
			m.debugf("phase=%s\n", phase)
			m.debugf("rooms=%s\n", formatBlockRooms(block))
			m.debugln("TRY MOVE")
			m.debugf("delta=(%d,%d)\n", drMove, dcMove)

			if err := m.moveBlock(block, drMove, dcMove); err != nil {
				m.printRejection(err)
				failures = append(failures, fmt.Sprintf("candidate %d (size=%d): %v", tried, len(block), err))
				continue
			}

			anchorAfter := m.rooms[anchorID]
			targetAfter := m.rooms[targetID]
			if anchorAfter == nil || !anchorAfter.Placed || targetAfter == nil || !targetAfter.Placed {
				return fmt.Errorf("mapping invariant: accepted candidate left anchor/target unplaced")
			}
			if targetAfter.R != anchorAfter.R+relDr || targetAfter.C != anchorAfter.C+relDc {
				return fmt.Errorf("mapping invariant: accepted candidate did not satisfy move adjacency %s -> %s",
					anchorID, targetID)
			}

			m.debugln("ACCEPTED")
			m.debugf("blockIndex=%d\n", tried)
			m.debugf("delta=(%d,%d)\n", drMove, dcMove)
			return nil
		}
		return fmt.Errorf("no candidate accepted in phase %s", phase)
	}

	destOnly, err := m.destinationCandidateBlocks(targetID, anchorID, false)
	if err != nil {
		return err
	}
	if err := trySet(destOnly, "destination_only", false); err == nil {
		return nil
	}

	crossing, err := m.destinationCandidateBlocks(targetID, anchorID, true)
	if err != nil {
		return err
	}
	if err := trySet(crossing, "anchor_crossing", true); err == nil {
		return nil
	}

	m.debugln("CORRECTION FAILED")
	m.debugf("candidatesTried=%d\n", tried)
	if tried == 0 {
		return fmt.Errorf("mapping error: no correction candidates generated for %s", targetID)
	}
	return fmt.Errorf("mapping error: no valid correction candidate for %s (tried %d): %s",
		targetID, tried, strings.Join(failures, " | "))
}

func (m *Mapper) validateBlockMove(block map[RoomID]struct{}, drMove, dcMove int) error {
	// Occupancy collisions against anchored side.
	for id := range block {
		rm := m.rooms[id]
		if rm == nil || !rm.Placed {
			return fmt.Errorf("mapping error: room %s in movable block is not placed", id)
		}
		nr, nc := rm.R+drMove, rm.C+dcMove
		if occ, ok := m.occ[[2]int{nr, nc}]; ok {
			if _, inBlock := block[occ]; !inBlock {
				return &collisionError{
					Row:      nr,
					Col:      nc,
					Occupant: occ,
					Moving:   id,
				}
			}
		}
	}

	coordAfter := func(id RoomID) (int, int, bool) {
		rm := m.rooms[id]
		if rm == nil || !rm.Placed {
			return 0, 0, false
		}
		if _, inBlock := block[id]; inBlock {
			return rm.R + drMove, rm.C + dcMove, true
		}
		return rm.R, rm.C, true
	}

	// Check cross-boundary topology constraints after the move.
	if m.topo != nil {
		for from, fromRm := range m.rooms {
			if fromRm == nil || !fromRm.Placed {
				continue
			}
			exits := m.topo.ExitsFrom(from)
			fromR, fromC, okFrom := coordAfter(from)
			if !okFrom {
				continue
			}
			_, fromInBlock := block[from]

			for dir, to := range exits {
				toR, toC, okTo := coordAfter(to)
				if !okTo {
					continue
				}
				_, toInBlock := block[to]
				if fromInBlock == toInBlock {
					continue
				}

				drEdge, dcEdge, ok := dirDelta(dir)
				if !ok {
					return fmt.Errorf("mapping error: unsupported direction %q on edge %s -> %s", dir, from, to)
				}
				expToR, expToC := fromR+drEdge, fromC+dcEdge
				if toR != expToR || toC != expToC {
					return fmt.Errorf("mapping error: moving block would contradict known adjacency %s -%s-> %s", from, dir, to)
				}
			}
		}
	}

	return nil
}

func (m *Mapper) moveBlock(block map[RoomID]struct{}, drMove, dcMove int) error {
	if drMove == 0 && dcMove == 0 {
		return nil
	}
	if len(block) == 0 {
		return fmt.Errorf("mapping error: empty movable block")
	}

	if err := m.validateBlockMove(block, drMove, dcMove); err != nil {
		return err
	}
	return m.shiftWhere(func(r *Room) bool {
		_, ok := block[r.ID]
		return ok
	}, drMove, dcMove)
}

func (m *Mapper) captureSnapshot() mapperSnapshot {
	s := mapperSnapshot{
		rooms: make(map[RoomID]roomSnapshot, len(m.rooms)),
		occ:   make(map[[2]int]RoomID, len(m.occ)),
		locks: make(map[lockedAdjKey]struct{}, len(m.locks)),
	}
	for id, r := range m.rooms {
		if r == nil {
			continue
		}
		s.rooms[id] = roomSnapshot{
			Placed: r.Placed,
			R:      r.R,
			C:      r.C,
		}
	}
	for k, v := range m.occ {
		s.occ[k] = v
	}
	for k := range m.locks {
		s.locks[k] = struct{}{}
	}
	return s
}

func (m *Mapper) restoreSnapshot(s mapperSnapshot) {
	for id, rs := range s.rooms {
		r := m.getRoom(id)
		r.Placed = rs.Placed
		r.R = rs.R
		r.C = rs.C
	}
	m.occ = make(map[[2]int]RoomID, len(s.occ))
	for k, v := range s.occ {
		m.occ[k] = v
	}
	m.locks = make(map[lockedAdjKey]struct{}, len(s.locks))
	for k := range s.locks {
		m.locks[k] = struct{}{}
	}
}

func (m *Mapper) stateSignature() string {
	var parts []string
	for id, r := range m.rooms {
		if r == nil || !r.Placed {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s@%d,%d", id, r.R, r.C))
	}
	sort.Strings(parts)

	var lockParts []string
	for k := range m.locks {
		lockParts = append(lockParts, fmt.Sprintf("%s-%s-%s", k.From, k.Dir, k.To))
	}
	sort.Strings(lockParts)
	return strings.Join(parts, "|") + "||" + strings.Join(lockParts, "|")
}

func (m *Mapper) holeOpenNow(fromID RoomID, drHole, dcHole int) bool {
	from := m.rooms[fromID]
	if from == nil || !from.Placed {
		return false
	}
	holeR, holeC := from.R+drHole, from.C+dcHole
	_, occ := m.occ[[2]int{holeR, holeC}]
	return !occ
}

type plannedMove struct {
	Phase string
	Block map[RoomID]struct{}
	Dr    int
	Dc    int
}

func smallestBlocks(candidates []map[RoomID]struct{}, maxCount int, requireID RoomID) []map[RoomID]struct{} {
	var out []map[RoomID]struct{}
	for _, b := range candidates {
		if requireID != "" {
			if _, ok := b[requireID]; !ok {
				continue
			}
		}
		out = append(out, b)
		if len(out) >= maxCount {
			break
		}
	}
	return out
}

func planningBlocks(candidates []map[RoomID]struct{}, maxSmall int, requireID RoomID) []map[RoomID]struct{} {
	small := smallestBlocks(candidates, maxSmall, requireID)
	seen := make(map[string]struct{}, len(small))
	out := make([]map[RoomID]struct{}, 0, len(small)+3)
	for _, b := range small {
		k := blockKey(b)
		seen[k] = struct{}{}
		out = append(out, b)
	}

	// Ensure planner also tries some larger options.
	for i := len(candidates) - 1; i >= 0; i-- {
		b := candidates[i]
		if requireID != "" {
			if _, ok := b[requireID]; !ok {
				continue
			}
		}
		k := blockKey(b)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, b)
		break
	}
	for _, frac := range []float64{0.66, 0.5} {
		idx := int(float64(len(candidates)-1) * frac)
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		b := candidates[idx]
		if requireID != "" {
			if _, ok := b[requireID]; !ok {
				continue
			}
		}
		k := blockKey(b)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, b)
	}
	return out
}

func blockHasID(block map[RoomID]struct{}, id RoomID) bool {
	_, ok := block[id]
	return ok
}

func plannerDeltas(drHole, dcHole int) [][2]int {
	deltas := [][2]int{
		{drHole, dcHole},
		{-drHole, -dcHole},
	}
	// Add orthogonal setup moves so multi-step search can reposition before opening a hole.
	p1 := [2]int{dcHole, -drHole}
	p2 := [2]int{-dcHole, drHole}
	if p1 != [2]int{0, 0} {
		deltas = append(deltas, p1)
	}
	if p2 != [2]int{0, 0} && p2 != p1 {
		deltas = append(deltas, p2)
	}

	seen := make(map[[2]int]struct{})
	var out [][2]int
	for _, d := range deltas {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out
}

func (m *Mapper) planMakeRoomMultiStepDepth(fromID, blockerID RoomID, drHole, dcHole, maxDepth, maxActionsPerPhase, maxNodes int, deadline time.Time) (bool, int, bool, bool, error) {
	type node struct {
		snap mapperSnapshot
		path []plannedMove
	}

	root := m.captureSnapshot()
	visited := map[string]int{m.stateSignature(): 0}
	queue := []node{{snap: root, path: nil}}
	explored := 0
	cutoff := false
	timeout := false

	for len(queue) > 0 {
		if time.Now().After(deadline) {
			timeout = true
			break
		}
		cur := queue[0]
		queue = queue[1:]
		explored++

		m.restoreSnapshot(cur.snap)
		if m.holeOpenNow(fromID, drHole, dcHole) && len(cur.path) > 0 {
			m.debugln("MAKE ROOM PLAN ACCEPTED")
			m.debugf("steps=%d\n", len(cur.path))
			for i, st := range cur.path {
				m.debugf("step=%d phase=%s delta=(%d,%d) rooms=%s\n",
					i+1, st.Phase, st.Dr, st.Dc, formatBlockRooms(st.Block))
			}
			return true, explored, cutoff, timeout, nil
		}
		if len(cur.path) >= maxDepth {
			continue
		}

		var actions []plannedMove
		dest, err := m.destinationCandidateBlocks(blockerID, fromID, false)
		if err != nil {
			return false, explored, cutoff, timeout, err
		}
		src, err := m.destinationCandidateBlocks(fromID, blockerID, false)
		if err != nil {
			return false, explored, cutoff, timeout, err
		}
		destCross, err := m.destinationCandidateBlocks(blockerID, fromID, true)
		if err != nil {
			return false, explored, cutoff, timeout, err
		}
		srcCross, err := m.destinationCandidateBlocks(fromID, blockerID, true)
		if err != nil {
			return false, explored, cutoff, timeout, err
		}

		destSet := planningBlocks(dest, maxActionsPerPhase, "")
		srcSet := planningBlocks(src, maxActionsPerPhase, fromID)
		destCrossSet := planningBlocks(destCross, maxActionsPerPhase, fromID)
		srcCrossSet := planningBlocks(srcCross, maxActionsPerPhase, blockerID)

		// At root, use full candidate sets to avoid missing a legal first move due to sampling.
		if len(cur.path) == 0 {
			destSet = dest
			srcSet = src
			destCrossSet = destCross
			srcCrossSet = srcCross
		}

		seenActions := make(map[string]struct{})
		addAction := func(phase string, block map[RoomID]struct{}, drAct, dcAct int) {
			k := fmt.Sprintf("%s|%d|%d|%s", phase, drAct, dcAct, blockKey(block))
			if _, dup := seenActions[k]; dup {
				return
			}
			seenActions[k] = struct{}{}
			actions = append(actions, plannedMove{
				Phase: phase,
				Block: block,
				Dr:    drAct,
				Dc:    dcAct,
			})
		}
		for _, d := range plannerDeltas(drHole, dcHole) {
			for _, b := range destSet {
				addAction(fmt.Sprintf("destination_only_delta_%d_%d", d[0], d[1]), b, d[0], d[1])
			}
			for _, b := range srcSet {
				addAction(fmt.Sprintf("source_side_delta_%d_%d", d[0], d[1]), b, d[0], d[1])
			}
			for _, b := range destCrossSet {
				if !blockHasID(b, fromID) {
					continue
				}
				addAction(fmt.Sprintf("destination_cross_delta_%d_%d", d[0], d[1]), b, d[0], d[1])
			}
			for _, b := range srcCrossSet {
				if !blockHasID(b, blockerID) {
					continue
				}
				addAction(fmt.Sprintf("source_cross_delta_%d_%d", d[0], d[1]), b, d[0], d[1])
			}
		}
		sort.Slice(actions, func(i, j int) bool {
			if len(actions[i].Block) != len(actions[j].Block) {
				return len(actions[i].Block) < len(actions[j].Block)
			}
			return blockKey(actions[i].Block) < blockKey(actions[j].Block)
		})

		base := m.captureSnapshot()
		for _, a := range actions {
			m.restoreSnapshot(base)
			if err := m.moveBlock(a.Block, a.Dr, a.Dc); err != nil {
				continue
			}
			if !m.holeOpenNow(fromID, drHole, dcHole) && len(cur.path)+1 >= maxDepth {
				continue
			}
			sig := m.stateSignature()
			depth := len(cur.path) + 1
			if seenDepth, ok := visited[sig]; ok && seenDepth <= depth {
				continue
			}
			visited[sig] = depth

			childPath := append(append([]plannedMove(nil), cur.path...), plannedMove{
				Phase: a.Phase,
				Block: cloneBlock(a.Block),
				Dr:    a.Dr,
				Dc:    a.Dc,
			})
			queue = append(queue, node{
				snap: m.captureSnapshot(),
				path: childPath,
			})
			if len(queue) > maxNodes {
				cutoff = true
				break
			}
		}
		if cutoff {
			break
		}
	}
	m.restoreSnapshot(root)
	return false, explored, cutoff, timeout, nil
}

func (m *Mapper) planMakeRoomMultiStep(fromID, blockerID RoomID, drHole, dcHole, maxDepth int) (bool, error) {
	const maxActionsPerPhase = 10
	const maxNodesPerDepth = 3000
	const maxTotalNodes = 20000
	const maxWall = 3 * time.Second

	deadline := time.Now().Add(maxWall)
	totalExplored := 0

	m.debugln("MAKE ROOM PLAN SEARCH")
	m.debugf("maxDepth=%d maxActionsPerPhase=%d maxNodesPerDepth=%d maxTotalNodes=%d maxWallMs=%d\n",
		maxDepth, maxActionsPerPhase, maxNodesPerDepth, maxTotalNodes, maxWall.Milliseconds())

	for depth := 1; depth <= maxDepth; depth++ {
		remaining := maxTotalNodes - totalExplored
		if remaining <= 0 {
			m.debugln("MAKE ROOM PLAN FAILED")
			m.debugf("reason=budget_exhausted exploredTotal=%d\n", totalExplored)
			return false, nil
		}
		nodesThisDepth := maxNodesPerDepth
		if remaining < nodesThisDepth {
			nodesThisDepth = remaining
		}

		ok, explored, cutoff, timeout, err := m.planMakeRoomMultiStepDepth(
			fromID, blockerID, drHole, dcHole, depth, maxActionsPerPhase, nodesThisDepth, deadline,
		)
		if err != nil {
			return false, err
		}
		totalExplored += explored
		m.debugf("MAKE ROOM PLAN DEPTH depth=%d explored=%d cutoff=%v timeout=%v\n",
			depth, explored, cutoff, timeout)
		if ok {
			return true, nil
		}
		if timeout {
			m.debugln("MAKE ROOM PLAN FAILED")
			m.debugf("reason=timeout exploredTotal=%d\n", totalExplored)
			return false, nil
		}
	}

	m.debugln("MAKE ROOM PLAN FAILED")
	m.debugf("reason=no_plan_within_bound exploredTotal=%d\n", totalExplored)
	return false, nil
}

func (m *Mapper) validateHoleOpens(block map[RoomID]struct{}, fromID RoomID, drMove, dcMove, drHole, dcHole int) error {
	coordAfter := func(id RoomID) (int, int, bool) {
		rm := m.rooms[id]
		if rm == nil || !rm.Placed {
			return 0, 0, false
		}
		if _, inBlock := block[id]; inBlock {
			return rm.R + drMove, rm.C + dcMove, true
		}
		return rm.R, rm.C, true
	}

	fromR, fromC, ok := coordAfter(fromID)
	if !ok {
		return fmt.Errorf("mapping error: makeRoom source %s is not placed", fromID)
	}
	holeR, holeC := fromR+drHole, fromC+dcHole

	for id, rm := range m.rooms {
		if rm == nil || !rm.Placed {
			continue
		}
		r, c, ok := coordAfter(id)
		if !ok {
			continue
		}
		if r == holeR && c == holeC {
			return &collisionError{
				Row:      holeR,
				Col:      holeC,
				Occupant: id,
				Moving:   fromID,
			}
		}
	}
	return nil
}

// Make a hole adjacent to 'from' in direction 'dir' by moving destination-side candidates.
func (m *Mapper) makeRoom(from *Room, dir string, blocker RoomID) error {
	m.debugln("MAKE ROOM START")
	m.debugf("fromRoom=%s\n", from.ID)
	m.debugf("direction=%s\n", dir)
	m.debugf("blocker=%s\n", blocker)

	dr, dc, ok := dirDelta(dir)
	if !ok {
		return fmt.Errorf("HUH? makeRoom got unsupported dir %q", dir)
	}

	var failures []string
	tried := 0
	seen := make(map[string]struct{})

	trySet := func(candidates []map[RoomID]struct{}, phase string, requireAnchor bool, moveDr, moveDc int) error {
		for _, block := range candidates {
			if requireAnchor {
				if _, ok := block[from.ID]; !ok {
					continue
				}
			}
			k := blockKey(block)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			tried++

			m.debugln("MAKE ROOM CANDIDATE")
			m.debugf("blockIndex=%d\n", tried)
			m.debugf("phase=%s\n", phase)
			m.debugf("rooms=%s\n", formatBlockRooms(block))
			m.debugln("MAKE ROOM TRY")
			m.debugf("fromRoom=%s\n", from.ID)
			m.debugf("direction=%s\n", dir)
			m.debugln("strategy=destination_candidate")
			m.debugf("delta=(%d,%d)\n", moveDr, moveDc)

			if err := m.validateHoleOpens(block, from.ID, moveDr, moveDc, dr, dc); err != nil {
				m.printRejection(err)
				failures = append(failures, fmt.Sprintf("candidate %d (size=%d): %v", tried, len(block), err))
				continue
			}

			if err := m.moveBlock(block, moveDr, moveDc); err != nil {
				m.printRejection(err)
				failures = append(failures, fmt.Sprintf("candidate %d (size=%d): %v", tried, len(block), err))
				continue
			}
			m.debugln("MAKE ROOM APPLIED")
			m.debugln("strategy=destination_candidate")
			m.debugf("blockIndex=%d\n", tried)
			m.debugf("delta=(%d,%d)\n", moveDr, moveDc)
			return nil
		}
		return fmt.Errorf("no candidate accepted in phase %s", phase)
	}

	destOnly, err := m.destinationCandidateBlocks(blocker, from.ID, false)
	if err != nil {
		return err
	}
	if err := trySet(destOnly, "destination_only", false, dr, dc); err == nil {
		return nil
	}

	// Two-sided fallback: move minimal source-side branch away from blocker.
	sourceSide, err := m.destinationCandidateBlocks(from.ID, blocker, false)
	if err != nil {
		return err
	}
	if err := trySet(sourceSide, "source_side", true, -dr, -dc); err == nil {
		return nil
	}

	// Bounded multi-step search: shortest plan first, preferring smaller blocks.
	if ok, err := m.planMakeRoomMultiStep(from.ID, blocker, dr, dc, 6); err != nil {
		return err
	} else if ok {
		return nil
	}

	// Legacy broad shifts retained as a final fallback.
	tryPlane := func(name string, pred func(*Room) bool, drShift, dcShift int) error {
		m.debugln("MAKE ROOM TRY")
		m.debugf("fromRoom=%s\n", from.ID)
		m.debugf("direction=%s\n", dir)
		m.debugf("strategy=%s\n", name)
		m.debugf("delta=(%d,%d)\n", drShift, dcShift)
		if err := m.shiftWhere(pred, drShift, dcShift); err != nil {
			m.printRejection(err)
			failures = append(failures, fmt.Sprintf("%s: %v", name, err))
			return err
		}
		m.debugln("MAKE ROOM APPLIED")
		m.debugf("strategy=%s\n", name)
		m.debugf("delta=(%d,%d)\n", drShift, dcShift)
		return nil
	}

	switch dir {
	case "E":
		if err := tryPlane("east_strict", func(r *Room) bool { return r.C > from.C }, 0, 1); err == nil {
			return nil
		}
		if err := tryPlane("east_inclusive", func(r *Room) bool { return r.C >= from.C }, 0, 1); err == nil {
			return nil
		}
	case "W":
		if err := tryPlane("west_strict", func(r *Room) bool { return r.C < from.C }, 0, -1); err == nil {
			return nil
		}
		if err := tryPlane("west_fallback_east_inclusive", func(r *Room) bool { return r.C >= from.C }, 0, 1); err == nil {
			return nil
		}
	case "S":
		if err := tryPlane("south_strict", func(r *Room) bool { return r.R > from.R }, 1, 0); err == nil {
			return nil
		}
		if err := tryPlane("south_fallback_north_inclusive", func(r *Room) bool { return r.R <= from.R }, -1, 0); err == nil {
			return nil
		}
	case "N":
		if err := tryPlane("north_strict", func(r *Room) bool { return r.R < from.R }, -1, 0); err == nil {
			return nil
		}
		if err := tryPlane("north_fallback_south_inclusive", func(r *Room) bool { return r.R >= from.R }, 1, 0); err == nil {
			return nil
		}
	}

	return fmt.Errorf("mapping error: makeRoom failed for blocker %s after %d candidate attempts: %s",
		blocker, tried+2, strings.Join(failures, " | "))
}

func (m *Mapper) rebuildDiscoveredLayout(cs ConstraintSet, enterID, fromID RoomID, dirMoved string) error {
	result, err := m.solver().ComputeRebuildResult(
		m.toSolverConstraintSet(cs),
		solver.RoomID(enterID),
		solver.RoomID(fromID),
		dirMoved,
	)
	if err != nil {
		return err
	}
	for id, r := range m.rooms {
		if r == nil {
			continue
		}
		rs, ok := result.Rooms[solver.RoomID(id)]
		if !ok {
			r.Placed = false
			continue
		}
		r.Placed = rs.Placed
		r.R = rs.R
		r.C = rs.C
	}
	m.occ = make(map[[2]int]RoomID, len(result.Occ))
	for k, v := range result.Occ {
		m.occ[k] = RoomID(v)
	}
	m.refreshLockedAdjacencies()
	m.cur = m.rooms[RoomID(result.Current)]
	return nil
}

// Enter applies one "arrival" event: ENTER id FROM dirMoved.
func (m *Mapper) Enter(id RoomID, dirMoved string) error {
	if err := m.enterIncremental(id, dirMoved); err == nil {
		return nil
	} else {
		if dirMoved == "START" {
			return err
		}
		if m.cur == nil || !m.cur.Placed {
			return err
		}
		cs := m.BuildConstraintSet(id, m.cur.ID)
		if rbErr := m.rebuildDiscoveredLayout(cs, id, m.cur.ID, dirMoved); rbErr != nil {
			return fmt.Errorf("%v; rebuild failed: %v", err, rbErr)
		}
		return nil
	}
}

func (m *Mapper) enterIncremental(id RoomID, dirMoved string) error {
	enter := m.getRoom(id)

	// helper: keep room '1' anchored at (R=1,C=0) after any global shift
	pinRoom1 := func() error {
		anchor := m.rooms[RoomID("1")]
		if anchor == nil || !anchor.Placed {
			return nil
		}
		dr := 1 - anchor.R
		dc := 0 - anchor.C
		if dr == 0 && dc == 0 {
			return nil
		}
		return m.shiftWhere(func(r *Room) bool { return r.Placed }, dr, dc)
	}

	if dirMoved == "START" {
		// fixed start: (R=1, C=0)
		if enter.Placed {
			m.clearOcc(enter.ID)
		}
		if occ, ok := m.occ[[2]int{1, 0}]; ok && occ != enter.ID {
			if err := m.shiftWhere(func(r *Room) bool { return r.Placed }, 0, 1); err != nil {
				return err
			}
		}
		m.setOcc(enter.ID, 1, 0)
		m.cur = enter
		return nil
	}

	if m.cur == nil || !m.cur.Placed {
		return fmt.Errorf("current room not set/placed")
	}

	dr, dc, ok := dirDelta(dirMoved)
	if !ok {
		return fmt.Errorf("HUH? bad direction %q", dirMoved)
	}

	expR, expC := m.cur.R+dr, m.cur.C+dc

	if enter.Placed {
		m.debugln("CORRECTION CHECK")
		m.debugf("currentRoom=%s\n", m.cur.ID)
		m.debugf("destinationRoom=%s\n", enter.ID)
		m.debugf("expectedCoords=(%d,%d)\n", expR, expC)
		m.debugf("currentCoords=(%d,%d)\n", enter.R, enter.C)

		coordNow := func(id RoomID) (int, int, bool) {
			rm := m.rooms[id]
			if rm == nil || !rm.Placed {
				return 0, 0, false
			}
			return rm.R, rm.C, true
		}
		if edgeAlignedAndOrdered(m.cur.R, m.cur.C, enter.R, enter.C, dirMoved) &&
			m.noRoomBetweenAxis(coordNow, m.cur.ID, enter.ID, m.cur.R, m.cur.C, enter.R, enter.C) {
			m.debugln("CORRECTION SKIPPED")
			m.debugln("reason=already_ordered")
			m.cur = enter
			return nil
		}

		// Room already exists: if misaligned, try destination-side candidate blocks.
		if enter.R != expR || enter.C != expC {
			drMove := expR - enter.R
			dcMove := expC - enter.C
			if err := m.moveDestinationWithCandidates(enter.ID, m.cur.ID, drMove, dcMove, expR, expC); err != nil {
				return err
			}
		} else {
			m.debugln("CORRECTION SKIPPED")
			m.debugln("reason=already_aligned")
		}
		m.cur = enter
		return nil
	}

	// New room: place it.
	if occ, ok := m.occ[[2]int{expR, expC}]; ok && occ != enter.ID {
		// Try one more time if something changed.
		if err := m.makeRoom(m.cur, dirMoved, occ); err != nil {
			return err
		}
		expR, expC = m.cur.R+dr, m.cur.C+dc
	}
	if occ, ok := m.occ[[2]int{expR, expC}]; ok && occ != enter.ID {
		return fmt.Errorf("mapping error: target cell still occupied after makeRoom at (R=%d,C=%s) by %s",
			expR, colName(expC), occ)
	}

	m.setOcc(enter.ID, expR, expC)
	m.cur = enter

	// If any earlier operations shifted the whole layout, pinning is a no-op; but safe to call.
	if err := pinRoom1(); err != nil {
		return err
	}

	// invariant: if room 1 is placed, it must be in occ at (1,0)
	if a := m.rooms[RoomID("1")]; a != nil && a.Placed {
		if a.R != 1 || a.C != 0 {
			return fmt.Errorf("INVARIANT FAIL: room 1 not pinned, now at (R=%d,C=%s)", a.R, colName(a.C))
		}
		if occ, ok := m.occ[[2]int{1, 0}]; !ok || occ != RoomID("1") {
			return fmt.Errorf("INVARIANT FAIL: occ(1,0)=%v (present=%v), expected '1'", occ, ok)
		}
	}
	return nil
}

func (m *Mapper) PrintGrid10x10() {
	m.PrintGrid10x10Discovered(nil)
}

func (m *Mapper) PrintGrid10x10Discovered(discovery *DiscoveryState) {
	first := true
	r1, r2 := 0, 0
	c1, c2 := 0, 0

	for rc, id := range m.occ {
		if discovery != nil && !discovery.IsDiscovered(id) {
			continue
		}
		r, c := rc[0], rc[1]
		if first {
			r1, r2 = r, r
			c1, c2 = c, c
			first = false
			continue
		}
		if r < r1 {
			r1 = r
		}
		if r > r2 {
			r2 = r
		}
		if c < c1 {
			c1 = c
		}
		if c > c2 {
			c2 = c
		}
	}

	if first {
		uiPrintln("(map empty)")
		return
	}

	uiPrintf("R\\C   ")
	for c := c1; c <= c2; c++ {
		uiPrintf("%-4s", colName(c))
	}
	uiPrintln()

	for r := r1; r <= r2; r++ {
		uiPrintf("%-5d", r)
		for c := c1; c <= c2; c++ {
			if id, ok := m.occ[[2]int{r, c}]; ok {
				if discovery != nil && !discovery.IsDiscovered(id) {
					uiPrintf("%-4s", ".")
					continue
				}
				label := cellLabel(id)
				if m.cur != nil && id == m.cur.ID && len(label) >= 2 {
					label = label[:1] + "@" + label[2:]
				}
				uiPrintf("%-4s", label)
			} else {
				uiPrintf("%-4s", ".")
			}
		}
		uiPrintln()
	}
}

func (m *Mapper) PrintRooms() {
	m.PrintRoomsDiscovered(nil, nil)
}

func (m *Mapper) PrintRoomsDiscovered(world *World, discovery *DiscoveryState) {
	var ids []string
	for id := range m.rooms {
		if discovery != nil && !discovery.IsDiscovered(id) {
			continue
		}
		ids = append(ids, string(id))
	}
	sort.Strings(ids)

	uiPrintln("ROOM COORDINATES")
	for _, s := range ids {
		id := RoomID(s)
		rm := m.rooms[id]
		if rm.Placed {
			uiPrintf("%s  (R=%d,C=%s)  cell=%s", s, rm.R, colName(rm.C), cellLabel(id))
			if world != nil && discovery != nil {
				exits := visibleExits(world.exits[id], discovery)
				if len(exits) == 0 {
					uiPrintf("  exits=(none)")
				} else {
					var dirs []string
					for d := range exits {
						dirs = append(dirs, d)
					}
					sort.Strings(dirs)
					uiPrintf("  exits=")
					for i, d := range dirs {
						if i > 0 {
							uiPrint(" ")
						}
						uiPrintf("%s(%s)", d, exits[d])
					}
				}
			}
			uiPrintln()
		} else {
			uiPrintf("%s  (unplaced)\n", s)
		}
	}
}

// ---------------- World loading ----------------

type Topology interface {
	ExitsFrom(roomID RoomID) map[string]RoomID
	Neighbors(roomID RoomID) []RoomID
	HasRoom(roomID RoomID) bool
}

type World struct {
	exits     map[RoomID]map[string]RoomID // room -> dir -> neighbor
	neighbors map[RoomID]map[RoomID]struct{}
}

func (w *World) ExitsFrom(roomID RoomID) map[string]RoomID {
	if w == nil || w.exits == nil {
		return map[string]RoomID{}
	}
	ex := w.exits[roomID]
	out := make(map[string]RoomID, len(ex))
	for d, id := range ex {
		out[d] = id
	}
	return out
}

func (w *World) Neighbors(roomID RoomID) []RoomID {
	if w == nil || w.neighbors == nil {
		return nil
	}
	set := w.neighbors[roomID]
	out := make([]RoomID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (w *World) HasRoom(roomID RoomID) bool {
	if w == nil || w.exits == nil {
		return false
	}
	_, ok := w.exits[roomID]
	return ok
}

type DiscoveryState struct {
	discoveredRooms map[RoomID]struct{}
}

func NewDiscoveryState() *DiscoveryState {
	return &DiscoveryState{
		discoveredRooms: make(map[RoomID]struct{}),
	}
}

func (d *DiscoveryState) Discover(roomID RoomID) {
	d.discoveredRooms[roomID] = struct{}{}
}

func (d *DiscoveryState) IsDiscovered(roomID RoomID) bool {
	if d == nil {
		return false
	}
	_, ok := d.discoveredRooms[roomID]
	return ok
}

func visibleExits(exits map[string]RoomID, discovery *DiscoveryState) map[string]RoomID {
	visible := make(map[string]RoomID)
	for dir, neighbor := range exits {
		if discovery.IsDiscovered(neighbor) {
			visible[dir] = neighbor
		}
	}
	return visible
}

func discoveredRoomIDs(discovery *DiscoveryState) []RoomID {
	ids := make([]RoomID, 0, len(discovery.discoveredRooms))
	for id := range discovery.discoveredRooms {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func LoadWorld(path string) (*World, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !st.IsDir() {
		return nil, fmt.Errorf("expected Rooms directory path, got file: %s", path)
	}

	w := &World{
		exits:     make(map[RoomID]map[string]RoomID),
		neighbors: make(map[RoomID]map[RoomID]struct{}),
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}
		full := filepath.Join(path, name)
		if err := parseRoomFileIntoWorld(w, full); err != nil {
			return nil, err
		}
	}

	return w, nil
}

func (w *World) ensureRoom(id RoomID) {
	if w.exits[id] == nil {
		w.exits[id] = make(map[string]RoomID)
	}
	if w.neighbors[id] == nil {
		w.neighbors[id] = make(map[RoomID]struct{})
	}
}

func (w *World) addExit(from RoomID, dir string, to RoomID) {
	w.ensureRoom(from)
	w.ensureRoom(to)
	w.exits[from][dir] = to
	w.neighbors[from][to] = struct{}{}
	w.neighbors[to][from] = struct{}{}
}

func parseRoomFileIntoWorld(w *World, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var roomID RoomID
	var curExitName string
	var curExitTo RoomID

	flushExit := func() error {
		if curExitName == "" {
			return nil
		}
		dir := normalizeDirName(curExitName)
		if dir == "" {
			return fmt.Errorf("%s: unsupported ExitName %q (only North/South/East/West)", filePath, curExitName)
		}
		if roomID == "" {
			return fmt.Errorf("%s: ExitName %q seen before RoomId", filePath, curExitName)
		}
		if curExitTo == "" {
			return fmt.Errorf("%s: ExitName %q missing ExitToRoomId", filePath, curExitName)
		}
		w.addExit(roomID, dir, curExitTo)
		curExitName = ""
		curExitTo = ""
		return nil
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "RoomId:") {
			roomID = RoomID(strings.TrimSpace(strings.TrimPrefix(line, "RoomId:")))
			continue
		}

		if strings.HasPrefix(line, "ExitName:") {
			if err := flushExit(); err != nil {
				return err
			}
			curExitName = strings.TrimSpace(strings.TrimPrefix(line, "ExitName:"))
			continue
		}

		if strings.HasPrefix(line, "ExitToRoomId:") {
			curExitTo = RoomID(strings.TrimSpace(strings.TrimPrefix(line, "ExitToRoomId:")))
			continue
		}

		if strings.EqualFold(line, "End of Exits") {
			if err := flushExit(); err != nil {
				return err
			}
			continue
		}

		if strings.EqualFold(line, "End of Room") {
			if err := flushExit(); err != nil {
				return err
			}
			break
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if roomID == "" {
		return fmt.Errorf("%s: missing RoomId:", filePath)
	}
	// Ensure room exists in exits map even if it has no exits.
	w.ensureRoom(roomID)
	return nil
}

func main() {
	cleanupLog, logErr := setupLogging("log.txt")
	if logErr != nil {
		fmt.Fprintln(os.Stderr, "Logging setup error:", logErr)
	} else {
		defer cleanupLog()
	}

	if len(os.Args) < 2 || strings.TrimSpace(os.Args[1]) == "" {
		uiPrintln("Usage: go run . <Rooms directory path>")
		os.Exit(2)
	}
	worldPath := os.Args[1]

	world, err := LoadWorld(worldPath)
	if err != nil {
		uiPrintln("LoadWorld error:", err)
		os.Exit(1)
	}

	mapper := NewMapper()
	mapper.SetDebugWriter(os.Stdout)
	discovery := NewDiscoveryState()

	// Fixed start room for the harness.
	start := RoomID("JesseSquare8")
	session, err := NewNavigationSession(world, mapper, discovery, start)
	if err != nil {
		uiPrintln(err)
		os.Exit(1)
	}

	uiPrintln("Commands: n s e w ne nw se sw | look | map | coords | show | gui | quit")
	uiPrintln()

	in := bufio.NewReader(os.Stdin)
	for {
		cur := session.CurrentRoom()
		uiPrintln()
		uiPrintf("%s\n", string(cur))
		// Always show true exits from the current room in the room prompt.
		ex := session.CurrentExits()
		if len(ex) == 0 {
			uiPrintln("Exits: (none)")
		} else {
			var dirs []string
			for d := range ex {
				dirs = append(dirs, d)
			}
			sort.Strings(dirs)
			uiPrint("Exits: ")
			for i, d := range dirs {
				if i > 0 {
					uiPrint(" ")
				}
				uiPrintf("%s(%s)", d, string(ex[d]))
			}
			uiPrintln()
		}
		uiPrint("> ")
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			continue
		}

		switch line {
		case "quit", "exit":
			return
		case "gui":
			wcswin32.RunWCS()
		case "look":
			uiPrintf("You are in Room %s\n", string(cur))
			// "look" should also show true exits for the current room.
			ex := session.CurrentExits()
			if len(ex) == 0 {
				uiPrintln("Exits: (none)")
				break
			}
			var dirs []string
			for d := range ex {
				dirs = append(dirs, d)
			}
			sort.Strings(dirs)
			uiPrint("Exits: ")
			for i, d := range dirs {
				if i > 0 {
					uiPrint(" ")
				}
				uiPrintf("%s(%s)", d, string(ex[d]))
			}
			uiPrintln()
		case "map":
			session.Mapper().PrintGrid10x10Discovered(session.Discovery())
		case "coords":
			session.Mapper().PrintRoomsDiscovered(world, session.Discovery())
		case "show":
			ids := discoveredRoomIDs(session.Discovery())
			uiPrintln("Discovered rooms:")
			if len(ids) == 0 {
				uiPrintln("(none)")
				break
			}
			for _, id := range ids {
				uiPrintln(string(id))
			}
		default:
			// movement: only N/S/E/W
			dir := normalizeDirName(line)
			if dir == "" {
				uiPrintln("HUH?")
				continue
			}
			if err := session.Move(dir); err != nil {
				if err.Error() == "no exit that way" {
					uiPrintln("No exit that way.")
					continue
				}
				uiPrintln("Mapper error:", err)
				continue
			}
			// Show map after every move
			session.Mapper().PrintGrid10x10Discovered(session.Discovery())
		}
	}
}
