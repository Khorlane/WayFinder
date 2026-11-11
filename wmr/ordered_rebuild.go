package wmr

import (
	"errors"
	"fmt"
	"sort"
)

type rebuildRelation struct {
	from, to int
	delta    [2]int
	reserve  bool
}

// rebuildOrdered assigns rows and columns from the enforced relationships.
// Separate axis groups can move independently, even if they shared an old coordinate.
func (s *ConstraintSolver) rebuildOrdered(cs SolverConstraintSet, discovered map[SolverRoomID]struct{}, oldPos map[SolverRoomID][2]int, fromID, enterID SolverRoomID, validate func(map[SolverRoomID][2]int) error) (map[SolverRoomID][2]int, map[[2]int]SolverRoomID, error) {
	ids := make([]SolverRoomID, 0, len(discovered))
	for id := range discovered {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	index := make(map[SolverRoomID]int, len(ids))
	for i, id := range ids {
		index[id] = i
	}

	var relations []rebuildRelation
	var established []rebuildRelation
	for _, constraint := range cs.Relations {
		if !constraint.Enforced || !constraint.RequiresOrder {
			continue
		}
		key := constraint.Key
		from, fromKnown := index[key.From]
		to, toKnown := index[key.To]
		dr, dc, valid := s.ctx.DirDelta(key.Dir)
		if !fromKnown || !toKnown || !valid {
			return nil, nil, fmt.Errorf("invalid rebuild relation %s -%s-> %s", key.From, key.Dir, key.To)
		}
		relation := rebuildRelation{from: from, to: to, delta: [2]int{dr, dc}, reserve: constraint.NoRoomBetween}
		relations = append(relations, relation)
		if constraint.Locked {
			established = append(established, relation)
		}
	}

	groups := [2][]int{rebuildAxisGroups(len(ids), relations, 0), rebuildAxisGroups(len(ids), relations, 1)}
	var options [2][][]int
	for axis := range groups {
		preferred := make([]int, len(ids))
		for i, id := range ids {
			preferred[i] = oldPos[id][axis]
		}
		var err error
		options[axis], err = orderRebuildAxis(groups[axis], groups[1-axis], relations, preferred, axis)
		if err != nil {
			return nil, nil, fmt.Errorf("order rebuild %s: %w", [2]string{"rows", "columns"}[axis], err)
		}
	}

	anchor, base := fromID, oldPos[fromID]
	if _, pinned := index["1"]; pinned {
		anchor, base = "1", [2]int{1, 0}
	}
	anchorIndex := index[anchor]
	var best map[SolverRoomID][2]int
	var bestOcc map[[2]int]SolverRoomID
	var bestScore rebuildScore
	consider := func(coords map[SolverRoomID][2]int, occ map[[2]int]SolverRoomID) error {
		if err := validate(coords); err != nil {
			return err
		}
		score := scoreRebuild(coords, oldPos)
		if best == nil || score.betterThan(bestScore) {
			best, bestOcc, bestScore = coords, occ, score
		}
		return nil
	}
	repairCrossing := func(coords map[SolverRoomID][2]int, crossing *connectionCrossingError) {
		for side, link := range crossing.links {
			opposite := crossing.links[1-side]
			for axis, laneGroups := range groups {
				left, right := coords[opposite[0]][axis], coords[opposite[1]][axis]
				// Try positions on either side of the other connection's endpoints.
				targets := [...]int{left - 1, left, left + 1, right - 1, right, right + 1}
				for _, endpoint := range link {
					group := laneGroups[index[endpoint]]
					for _, target := range targets {
						shift := target - coords[endpoint][axis]
						if shift == 0 {
							continue
						}
						candidate, occupied := shiftRebuildLane(ids, coords, laneGroups, group, axis, shift)
						if candidate != nil {
							consider(candidate, occupied)
						}
					}
				}
			}
		}
	}
	var rejected map[SolverRoomID][2]int
	var rejectedErr error
	var rejectedScore rebuildScore
	var collision map[SolverRoomID][2]int
	existingGroups := [2][]int{rebuildAxisGroups(len(ids), established, 0), rebuildAxisGroups(len(ids), established, 1)}
	// Closing a loop can require moving the source lane toward the destination.
	// Try both endpoint groups against the established map before wider searches.
	for axis, laneGroups := range existingGroups {
		for _, endpoints := range [][2]SolverRoomID{{fromID, enterID}, {enterID, fromID}} {
			moving, movingKnown := index[endpoints[0]]
			fixed, fixedKnown := index[endpoints[1]]
			oldMoving, movingPlaced := oldPos[endpoints[0]]
			oldFixed, fixedPlaced := oldPos[endpoints[1]]
			if !movingKnown || !fixedKnown || !movingPlaced || !fixedPlaced || laneGroups[moving] == laneGroups[fixed] {
				continue
			}
			delta := oldFixed[axis] - oldMoving[axis]
			if delta == 0 {
				continue
			}
			coords, occ := shiftRebuildLane(ids, oldPos, laneGroups, laneGroups[moving], axis, delta)
			if coords == nil {
				continue
			}
			var crossing *connectionCrossingError
			if err := consider(coords, occ); errors.As(err, &crossing) {
				repairCrossing(coords, crossing)
			}
		}
	}
	budget := max(32, min(4096, 200000/(len(ids)*len(ids))))
	for _, rows := range options[0] {
		for _, cols := range options[1] {
			if budget == 0 {
				break
			}
			budget--
			rowOffset := base[0] - rows[groups[0][anchorIndex]]
			colOffset := base[1] - cols[groups[1][anchorIndex]]
			coords := make(map[SolverRoomID][2]int, len(ids))
			occ := make(map[[2]int]SolverRoomID, len(ids))
			for i, id := range ids {
				cell := [2]int{rowOffset + rows[groups[0][i]], colOffset + cols[groups[1][i]]}
				coords[id], occ[cell] = cell, id
			}
			if len(occ) != len(ids) {
				if s.ctx.ReportRejected != nil && collision == nil {
					collision = coords
				}
				continue
			}
			score := scoreRebuild(coords, oldPos)
			if best != nil && !score.betterThan(bestScore) {
				continue
			}
			if err := validate(coords); err != nil {
				if s.ctx.ReportRejected != nil && (rejected == nil || score.betterThan(rejectedScore)) {
					rejected, rejectedErr, rejectedScore = coords, err, score
				}
				continue
			}
			best, bestOcc, bestScore = coords, occ, score
		}
	}
	if s.ctx.ReportRejected != nil {
		if rejected != nil {
			s.ctx.ReportRejected("compact search", rejected, rejectedErr)
		} else if collision != nil {
			s.ctx.ReportRejected("compact search", collision, fmt.Errorf("candidate placed multiple rooms in the same cell"))
		}
	}
	if best == nil {
		return nil, nil, fmt.Errorf("no crossing-free layout found within the rebuild search")
	}
	return best, bestOcc, nil
}

func shiftRebuildLane(ids []SolverRoomID, source map[SolverRoomID][2]int, groups []int, group, axis, delta int) (map[SolverRoomID][2]int, map[[2]int]SolverRoomID) {
	if len(source) != len(ids) {
		return nil, nil
	}
	coords := make(map[SolverRoomID][2]int, len(ids))
	occ := make(map[[2]int]SolverRoomID, len(ids))
	for i, id := range ids {
		cell := source[id]
		if groups[i] == group {
			cell[axis] += delta
		}
		coords[id], occ[cell] = cell, id
	}
	if len(occ) != len(ids) {
		return nil, nil
	}
	if cell, pinned := coords["1"]; pinned && cell != [2]int{1, 0} {
		return nil, nil
	}
	return coords, occ
}

type rebuildScore struct{ area, movement int }

func (s rebuildScore) betterThan(other rebuildScore) bool {
	return s.area < other.area || s.area == other.area && s.movement < other.movement
}

func scoreRebuild(coords, oldPos map[SolverRoomID][2]int) rebuildScore {
	var low, high [2]int
	first := true
	movement := 0
	for id, cell := range coords {
		for axis := range cell {
			if first || cell[axis] < low[axis] {
				low[axis] = cell[axis]
			}
			if first || cell[axis] > high[axis] {
				high[axis] = cell[axis]
			}
			if old, ok := oldPos[id]; ok {
				delta := cell[axis] - old[axis]
				movement += max(delta, -delta)
			}
		}
		first = false
	}
	return rebuildScore{area: (high[0] - low[0] + 1) * (high[1] - low[1] + 1), movement: movement}
}

func rebuildAxisGroups(count int, relations []rebuildRelation, axis int) []int {
	parents := make([]int, count)
	for i := range parents {
		parents[i] = i
	}
	root := func(i int) int {
		for parents[i] != i {
			parents[i] = parents[parents[i]]
			i = parents[i]
		}
		return i
	}
	for _, relation := range relations {
		if relation.delta[axis] == 0 {
			parents[root(relation.to)] = root(relation.from)
		}
	}
	for i := range parents {
		parents[i] = root(i)
	}
	return parents
}

func orderRebuildAxis(groups, lanes []int, relations []rebuildRelation, preferred []int, axis int) ([][]int, error) {
	n := len(groups)
	after := make([][]int, n)
	before := make([][]int, n)
	indegree := make([]int, n)
	reserved := make([][][2]int, n)
	for _, relation := range relations {
		delta := relation.delta[axis]
		if delta == 0 {
			continue
		}
		first, last := groups[relation.from], groups[relation.to]
		if delta < 0 {
			first, last = last, first
		}
		if first == last {
			return nil, fmt.Errorf("a locked pair requires both alignment and strict ordering")
		}
		after[first] = append(after[first], last)
		before[last] = append(before[last], first)
		indegree[last]++
		if relation.reserve {
			for i, group := range groups {
				if lanes[i] == lanes[relation.from] && group != first && group != last {
					reserved[group] = append(reserved[group], [2]int{first, last})
				}
			}
		}
	}

	var candidates []int
	ranks := make([]int, n)
	for i := range groups {
		ranks[i] = -1
		if groups[i] == i {
			candidates = append(candidates, i)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return preferred[candidates[i]] < preferred[candidates[j]] })

	// While a locked interval is open, a room on its lane must wait outside it.
	// Bound the search so an incompatible layout cannot stall live output.
	const maxAttempts = 20000
	var options [][]int
	search := func(share bool, limit int) {
		attempts := 0
		var place func(depth, lastRank, lastGroup int)
		place = func(depth, lastRank, lastGroup int) {
			if depth == len(candidates) {
				options = append(options, append([]int(nil), ranks...))
				return
			}
			for _, group := range candidates {
				if attempts >= maxAttempts || len(options) >= limit {
					return
				}
				if ranks[group] >= 0 || indegree[group] != 0 {
					continue
				}
				blocked := false
				for _, pair := range reserved[group] {
					if ranks[pair[0]] >= 0 && ranks[pair[1]] < 0 {
						blocked = true
						break
					}
				}
				if blocked {
					continue
				}
				firstRank := lastRank + 1
				if share && lastRank >= 0 && lastGroup < group {
					firstRank = lastRank
					for _, pred := range before[group] {
						firstRank = max(firstRank, ranks[pred]+1)
					}
				}
				for rank := firstRank; rank <= lastRank+1 && attempts < maxAttempts && len(options) < limit; rank++ {
					attempts++
					ranks[group] = rank
					for _, next := range after[group] {
						indegree[next]--
					}
					place(depth+1, rank, group)
					for _, next := range after[group] {
						indegree[next]++
					}
					ranks[group] = -1
				}
			}
		}
		place(0, -1, -1)
	}
	search(false, 1)
	search(true, 64)
	if len(options) == 0 {
		return nil, fmt.Errorf("no order found for the locked lanes within the rebuild search")
	}
	sort.SliceStable(options, func(i, j int) bool {
		return maxAxisRank(options[i]) < maxAxisRank(options[j])
	})
	return options, nil
}

func maxAxisRank(ranks []int) int {
	maxRank := 0
	for _, rank := range ranks {
		maxRank = max(maxRank, rank)
	}
	return maxRank
}
