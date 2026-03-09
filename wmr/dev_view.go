package wmr

import "sort"

type discoveredView interface {
	IsDiscovered(RoomID) bool
}

type worldView interface {
	ExitsFrom(RoomID) map[string]RoomID
}

func (m *Mapper) PrintGrid10x10() {
	m.PrintGrid10x10Discovered(nil)
}

func (m *Mapper) PrintGrid10x10Discovered(discovery discoveredView) {
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

func (m *Mapper) PrintRoomsDiscovered(world worldView, discovery discoveredView) {
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
				exits := visibleExits(world.ExitsFrom(id), discovery)
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

func visibleExits(exits map[string]RoomID, discovery discoveredView) map[string]RoomID {
	visible := make(map[string]RoomID)
	for dir, neighbor := range exits {
		if discovery.IsDiscovered(neighbor) {
			visible[dir] = neighbor
		}
	}
	return visible
}
