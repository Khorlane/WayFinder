package main

import (
	"fmt"
	"sort"
)

type SolverContext struct {
	Rooms             map[RoomID]SolverRoomState
	NoRoomBetweenAxis func(func(RoomID) (int, int, bool), RoomID, RoomID, int, int, int, int) bool
	Debugln           func(...any)
	Debugf            func(string, ...any)
}

type SolverEngine interface {
	ValidateConstraintSet(cs ConstraintSet, coordAfter func(RoomID) (int, int, bool)) error
	ComputeRebuildResult(cs ConstraintSet, enterID, fromID RoomID, dirMoved string) (RebuildResult, error)
}

type SolverProvider func(SolverContext) SolverEngine

type SolverRoomState struct {
	Placed bool
	R      int
	C      int
}

type RebuildRoomState struct {
	Placed bool
	R      int
	C      int
}

type RebuildResult struct {
	Rooms   map[RoomID]RebuildRoomState
	Occ     map[[2]int]RoomID
	Current RoomID
}

type ConstraintSolver struct {
	ctx SolverContext
}

func NewConstraintSolver(ctx SolverContext) *ConstraintSolver {
	return &ConstraintSolver{ctx: ctx}
}

var DefaultSolverProvider SolverProvider = func(ctx SolverContext) SolverEngine {
	return NewConstraintSolver(ctx)
}

var _ SolverEngine = (*ConstraintSolver)(nil)

func (s *ConstraintSolver) ValidateConstraintSet(cs ConstraintSet, coordAfter func(RoomID) (int, int, bool)) error {
	for _, rel := range cs.Relations {
		if !rel.Enforced {
			continue
		}
		k := rel.Key
		fromR, fromC, okFrom := coordAfter(k.From)
		toR, toC, okTo := coordAfter(k.To)
		if !okFrom || !okTo {
			return fmt.Errorf("mapping invariant: locked adjacency references unplaced room: %s -%s-> %s", k.From, k.Dir, k.To)
		}
		if rel.RequiresOrder && !edgeAlignedAndOrdered(fromR, fromC, toR, toC, k.Dir) {
			expDr, expDc, _ := dirDelta(k.Dir)
			return &lockedAdjViolationError{
				Key:       k,
				ExpectedR: expDr,
				ExpectedC: expDc,
			}
		}
		if rel.NoRoomBetween &&
			!s.ctx.NoRoomBetweenAxis(coordAfter, k.From, k.To, fromR, fromC, toR, toC) {
			expDr, expDc, _ := dirDelta(k.Dir)
			return &lockedAdjViolationError{
				Key:       k,
				ExpectedR: expDr,
				ExpectedC: expDc,
			}
		}
	}
	return nil
}

func (s *ConstraintSolver) ComputeRebuildResult(cs ConstraintSet, enterID, fromID RoomID, dirMoved string) (RebuildResult, error) {
	rooms := make(map[RoomID]SolverRoomState, len(s.ctx.Rooms)+2)
	for id, rs := range s.ctx.Rooms {
		rooms[id] = rs
	}
	if _, ok := rooms[enterID]; !ok {
		rooms[enterID] = SolverRoomState{}
	}
	if _, ok := rooms[fromID]; !ok {
		rooms[fromID] = SolverRoomState{}
	}

	drMove, dcMove, ok := dirDelta(dirMoved)
	if !ok {
		return RebuildResult{}, fmt.Errorf("mapping error: rebuild got unsupported direction %q", dirMoved)
	}

	discovered := make(map[RoomID]struct{}, len(cs.Discovered)+2)
	for id := range cs.Discovered {
		discovered[id] = struct{}{}
	}
	if len(discovered) == 0 {
		for id, r := range rooms {
			if r.Placed {
				discovered[id] = struct{}{}
			}
		}
	}
	discovered[enterID] = struct{}{}
	discovered[fromID] = struct{}{}

	type edge struct {
		to RoomID
		dr int
		dc int
	}
	adj := make(map[RoomID][]edge)
	for _, rel := range cs.Relations {
		k := rel.Key
		if _, ok := discovered[k.From]; !ok {
			continue
		}
		if _, ok := discovered[k.To]; !ok {
			continue
		}
		dr, dc, ok := dirDelta(k.Dir)
		if !ok {
			continue
		}
		adj[k.From] = append(adj[k.From], edge{to: k.To, dr: dr, dc: dc})
	}
	for from := range adj {
		sort.Slice(adj[from], func(i, j int) bool {
			if adj[from][i].dr != adj[from][j].dr {
				return adj[from][i].dr < adj[from][j].dr
			}
			if adj[from][i].dc != adj[from][j].dc {
				return adj[from][i].dc < adj[from][j].dc
			}
			return adj[from][i].to < adj[from][j].to
		})
	}
	neighbors := func(id RoomID) []edge {
		return adj[id]
	}

	oldPos := make(map[RoomID][2]int)
	for id, r := range rooms {
		if !r.Placed {
			continue
		}
		if _, ok := discovered[id]; !ok {
			continue
		}
		oldPos[id] = [2]int{r.R, r.C}
	}

	affected := make(map[RoomID]struct{})
	type qItem struct {
		id    RoomID
		depth int
	}
	q := []qItem{{id: fromID, depth: 0}, {id: enterID, depth: 0}}
	for _, it := range q {
		affected[it.id] = struct{}{}
	}
	const localDepth = 3
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if cur.depth >= localDepth {
			continue
		}
		for _, e := range neighbors(cur.id) {
			if _, ok := affected[e.to]; ok {
				continue
			}
			affected[e.to] = struct{}{}
			q = append(q, qItem{id: e.to, depth: cur.depth + 1})
		}
	}

	layoutApprox := func(anchorOutside bool) (map[RoomID][2]int, map[[2]int]RoomID, error) {
		coords := make(map[RoomID][2]int)
		occ := make(map[[2]int]RoomID)
		anchored := make(map[RoomID]struct{})

		if anchorOutside {
			for id := range discovered {
				if _, inAffected := affected[id]; inAffected {
					continue
				}
				p, ok := oldPos[id]
				if !ok {
					continue
				}
				coords[id] = p
				occ[[2]int{p[0], p[1]}] = id
				anchored[id] = struct{}{}
			}
		}

		placeNear := func(id RoomID, wantR, wantC int) error {
			if _, placed := coords[id]; placed {
				return nil
			}
			tryCell := func(r, c int) bool {
				key := [2]int{r, c}
				if cur, taken := occ[key]; taken {
					if cur != id {
						return false
					}
				}
				coords[id] = [2]int{r, c}
				occ[key] = id
				return true
			}
			if tryCell(wantR, wantC) {
				return nil
			}
			maxRadius := len(discovered)*4 + 32
			for radius := 1; radius <= maxRadius; radius++ {
				for dr := -radius; dr <= radius; dr++ {
					drAbs := dr
					if drAbs < 0 {
						drAbs = -drAbs
					}
					dc := radius - drAbs
					cands := [][2]int{{wantR + dr, wantC + dc}}
					if dc != 0 {
						cands = append(cands, [2]int{wantR + dr, wantC - dc})
					}
					for _, rc := range cands {
						if tryCell(rc[0], rc[1]) {
							return nil
						}
					}
				}
			}
			return fmt.Errorf("mapping error: unable to place %s near (R=%d,C=%s)", id, wantR, colName(wantC))
		}

		fromBase := [2]int{0, 0}
		if p, ok := oldPos[fromID]; ok {
			fromBase = p
		}
		if _, ok := anchored[fromID]; ok {
			delete(anchored, fromID)
			delete(occ, coords[fromID])
			delete(coords, fromID)
		}
		if err := placeNear(fromID, fromBase[0], fromBase[1]); err != nil {
			return nil, nil, err
		}

		fromRC := coords[fromID]
		enterR, enterC := fromRC[0]+drMove, fromRC[1]+dcMove
		enterTarget := [2]int{enterR, enterC}
		if occID, taken := occ[enterTarget]; taken && occID != enterID {
			if _, anchoredOcc := anchored[occID]; anchoredOcc {
				return nil, nil, fmt.Errorf("mapping error: local repack blocked by anchored room %s at target (R=%d,C=%s)",
					occID, enterR, colName(enterC))
			}
			delete(coords, occID)
			delete(occ, enterTarget)
		}
		coords[enterID] = enterTarget
		occ[enterTarget] = enterID

		bfs := []RoomID{fromID, enterID}
		seen := map[RoomID]struct{}{fromID: {}, enterID: {}}
		for len(bfs) > 0 {
			id := bfs[0]
			bfs = bfs[1:]
			base, ok := coords[id]
			if !ok {
				continue
			}
			for _, e := range neighbors(id) {
				if _, has := coords[e.to]; !has {
					if err := placeNear(e.to, base[0]+e.dr, base[1]+e.dc); err != nil {
						return nil, nil, err
					}
				}
				if _, done := seen[e.to]; !done {
					seen[e.to] = struct{}{}
					bfs = append(bfs, e.to)
				}
			}
		}

		var ids []string
		for id := range discovered {
			ids = append(ids, string(id))
		}
		sort.Strings(ids)
		for _, sid := range ids {
			id := RoomID(sid)
			if _, ok := coords[id]; ok {
				continue
			}
			want := [2]int{0, 0}
			if p, ok := oldPos[id]; ok {
				want = p
			}
			if err := placeNear(id, want[0], want[1]); err != nil {
				return nil, nil, err
			}
		}
		return coords, occ, nil
	}

	s.ctx.Debugln("REPACK START")
	s.ctx.Debugln("mode=local")
	coords, newOcc, err := layoutApprox(true)
	if err != nil {
		s.ctx.Debugln("REPACK FALLBACK")
		s.ctx.Debugf("reason=%v\n", err)
		s.ctx.Debugln("mode=full_rebuild")
		coords, newOcc, err = layoutApprox(false)
		if err != nil {
			return RebuildResult{}, err
		}
	}

	fromRC, okFrom := coords[fromID]
	enterRC, okEnter := coords[enterID]
	if !okFrom || !okEnter {
		return RebuildResult{}, fmt.Errorf("mapping error: repack failed to place from/enter rooms")
	}
	if enterRC[0] != fromRC[0]+drMove || enterRC[1] != fromRC[1]+dcMove {
		return RebuildResult{}, fmt.Errorf("mapping error: repack did not satisfy immediate move %s -%s-> %s", fromID, dirMoved, enterID)
	}

	roomStates := make(map[RoomID]RebuildRoomState, len(rooms))
	for id := range rooms {
		if rc, ok := coords[id]; ok {
			roomStates[id] = RebuildRoomState{Placed: true, R: rc[0], C: rc[1]}
		} else {
			roomStates[id] = RebuildRoomState{Placed: false}
		}
	}

	s.ctx.Debugln("REPACK APPLIED")
	s.ctx.Debugf("from=%s enter=%s dir=%s\n", fromID, enterID, dirMoved)
	return RebuildResult{
		Rooms:   roomStates,
		Occ:     newOcc,
		Current: enterID,
	}, nil
}
