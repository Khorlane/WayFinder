package wmr

import "fmt"

type connectionCrossingError struct {
	links [2][2]SolverRoomID
}

func (e *connectionCrossingError) Error() string {
	return fmt.Sprintf("connections %s–%s and %s–%s cross", e.links[0][0], e.links[0][1], e.links[1][0], e.links[1][1])
}

func onSegment(a, b, p [2]int) bool {
	return orientation(a, b, p) == 0 &&
		p[0] >= min(a[0], b[0]) && p[0] <= max(a[0], b[0]) &&
		p[1] >= min(a[1], b[1]) && p[1] <= max(a[1], b[1])
}

func orientation(a, b, p [2]int) int64 {
	return int64(b[0]-a[0])*int64(p[1]-a[1]) - int64(b[1]-a[1])*int64(p[0]-a[0])
}

func segmentsIntersect(a, b, c, d [2]int) bool {
	if max(min(a[0], b[0]), min(c[0], d[0])) > min(max(a[0], b[0]), max(c[0], d[0])) ||
		max(min(a[1], b[1]), min(c[1], d[1])) > min(max(a[1], b[1]), max(c[1], d[1])) {
		return false
	}
	p, q := orientation(a, b, c), orientation(a, b, d)
	r, s := orientation(c, d, a), orientation(c, d, b)
	return (p <= 0 && q >= 0 || q <= 0 && p >= 0) && (r <= 0 && s >= 0 || s <= 0 && r >= 0)
}

func sharesEndpoint[T comparable](a, b, c, d T) bool {
	return a == c || a == d || b == c || b == d
}

func validateConnections(cs SolverConstraintSet, coords map[SolverRoomID][2]int) error {
	var links [][2]SolverRoomID
	seen := make(map[[2]SolverRoomID]bool)
	for _, rel := range cs.Relations {
		key := rel.Key
		a, aOK := coords[key.From]
		b, bOK := coords[key.To]
		if !aOK || !bOK || !edgeAlignedAndOrdered(a[0], a[1], b[0], b[1], key.Dir) {
			continue
		}
		link := [2]SolverRoomID{min(key.From, key.To), max(key.From, key.To)}
		if seen[link] {
			continue
		}
		blocked := false
		for id, p := range coords {
			if id == key.From || id == key.To || !onSegment(a, b, p) {
				continue
			}
			// A blocked cardinal exit is drawn only as a tick until its lane is clear.
			if !rel.Enforced && rel.AxisAligned {
				blocked = true
				break
			}
			return fmt.Errorf("connection %s -%s-> %s passes through room %s", key.From, key.Dir, key.To, id)
		}
		if blocked {
			continue
		}
		for _, other := range links {
			if !sharesEndpoint(link[0], link[1], other[0], other[1]) && segmentsIntersect(a, b, coords[other[0]], coords[other[1]]) {
				return &connectionCrossingError{links: [2][2]SolverRoomID{link, other}}
			}
		}
		seen[link] = true
		links = append(links, link)
	}
	return nil
}

func (m *Mapper) canLockConnection(key lockedAdjKey, from, to *Room) bool {
	a, b := [2]int{from.R, from.C}, [2]int{to.R, to.C}
	for id, room := range m.rooms {
		if room != nil && room.Placed && id != key.From && id != key.To && onSegment(a, b, [2]int{room.R, room.C}) {
			return false
		}
	}
	for other := range m.locks {
		if sharesEndpoint(key.From, key.To, other.From, other.To) {
			continue
		}
		left, right := m.rooms[other.From], m.rooms[other.To]
		if left != nil && left.Placed && right != nil && right.Placed &&
			segmentsIntersect(a, b, [2]int{left.R, left.C}, [2]int{right.R, right.C}) {
			return false
		}
	}
	return true
}
