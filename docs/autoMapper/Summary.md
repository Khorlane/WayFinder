Summarize this entire conversation in detailed bullet points for a new, blank session. Focus on:
1. Objective: The primary goal of the session. 
2. Context: Key information/decisions agreed upon. 
3. Key Takeaways/Insights: What we learned. 
4. Open Issues/Next Steps: What needs to be done next. 
Use clear Markdown formatting. Do not omit crucial details. 
Produce a downloadable *.md file named: AutoMapper.md

# AutoMapper Session Summary

## 1. Objective

-   Build and debug a **Go-based automapper prototype** for a MUD-style
    navigation system.
-   The mapper should:
    -   Maintain a **10x10 grid**.
    -   Place rooms dynamically as the player moves.
    -   Shift existing rooms when necessary to create space.
    -   Keep the map **consistent and deterministic**.
-   The environment simulates movement via **console input/output**,
    using a world file rather than live telnet input.

------------------------------------------------------------------------

## 2. Context

### World Representation

-   The world is defined in `world.txt` using a simple format:

        Room 1 S(2) E(8)
        Room 2 N(1) S(3)
        ...

-   Parsed into:

    ``` go
    map[RoomID]map[string]RoomID
    ```

-   Directions supported:

    -   N S E W
    -   NE NW SE SW
    -   U D (ignored in current grid model)

### Mapper Data Structures

Primary structures:

``` go
type Room struct {
    ID     RoomID
    Placed bool
    R, C   int
}

type Mapper struct {
    rooms map[RoomID]*Room
    occ   map[[2]int]RoomID
    cur   *Room
}
```

Key concepts:

-   `rooms` = authoritative database of rooms and coordinates
-   `occ` = occupancy grid mapping `(row,col)` → room id
-   `cur` = current player location

Grid properties:

-   Rows: **1--10**
-   Columns: **A--J** (stored internally as 0--9)

### Console Simulation

The program simulates player movement with commands:

    n s e w ne nw se sw u d
    look
    map
    coords
    quit

Movement triggers:

    mapper.Enter(nextRoom, directionMoved)

The map prints **after every move**.

### Starting Conditions

-   Start room (`1`) is always anchored at:

```{=html}
<!-- -->
```
    R=1, C=A

This is enforced by a **pinRoom1() invariant check**.

------------------------------------------------------------------------

## 3. Key Takeaways / Insights

### 3.1 Occupancy Drift Bug

Observed failure:

    INVARIANT FAIL: occ(1,A)=0 (present=false), expected '1'

Meaning:

-   Room coordinates remained correct
-   But `occ` grid lost entries

Root cause:

-   `shiftWhere()` moved rooms but did not reliably rebuild the
    occupancy grid.

Fix:

    Rebuild m.occ from scratch after every shift.

------------------------------------------------------------------------

### 3.2 Missing setOcc() Logic

Another failure source was the placement routine.

Correct behavior of `setOcc()`:

1.  Remove old occupancy if the room was already placed.
2.  Update coordinates.
3.  Insert the new occupancy entry.
4.  Prevent collisions.

This ensured `rooms` and `occ` remain synchronized.

------------------------------------------------------------------------

### 3.3 Non‑Deterministic Shifts (Critical Discovery)

Major mapping bug occurred when creating space for new rooms.

Example scenario:

    A -> W -> 7

The mapper had to shift rooms to create a hole.

Original implementation:

    for r := range m.rooms {
        if pred(r) { shift }
    }

Problem:

-   Go maps iterate in **random order**
-   Predicate depended on `from.C`
-   `from.C` changed during iteration
-   Result: inconsistent shifts (rooms sometimes moved, sometimes
    didn't)

### Correct Solution

Make `shiftWhere()` a **4-phase operation**:

1.  Collect rooms to move
2.  Validate bounds
3.  Apply shifts
4.  Rebuild occupancy

This guarantees deterministic behavior.

------------------------------------------------------------------------

### 3.4 Correct Automapper Behavior Demonstrated

Final successful map after exploration:

    1 . . 8
    2 . . 9
    3 6 7 A
    4 . . B
    5 . . C

This shows:

-   Corridor from `3 → 6 → 7 → A`
-   East column `8 9 A B C`
-   South column `1 2 3 4 5`
-   Dynamic shifting successfully created the necessary space.

------------------------------------------------------------------------

### 3.5 Key Architectural Principle

The system now uses:

**rooms = authoritative coordinates**

**occ = derived structure rebuilt as needed**

This prevents state divergence.

------------------------------------------------------------------------

## 4. Open Issues / Next Steps

### 4.1 Move Grid Size Into Mapper Struct

Currently grid bounds are duplicated in multiple places.

Recommended change:

``` go
type Mapper struct {
    rooms map[RoomID]*Room
    occ   map[[2]int]RoomID
    cur   *Room
    maxR  int
    maxC  int
}
```

Initialize:

    maxR = 10
    maxC = 10

Then use `m.maxR/m.maxC` everywhere.

------------------------------------------------------------------------

### 4.2 Automapper Stability Testing

Test more complex scenarios:

-   multiple east/west corridor merges
-   loops
-   diagonal paths
-   returning to already placed rooms

Goal: ensure **global shifts never break invariants**.

------------------------------------------------------------------------

### 4.3 Visual Improvements

Possible enhancements:

-   Center the current room in the display
-   Highlight current location
-   Show explored vs unexplored exits
-   Allow dynamic map size

------------------------------------------------------------------------

### 4.4 Future Integration Target

The current prototype simulates a MUD locally.

Future goal:

-   integrate mapper with **telnet client stream**
-   parse movement messages
-   update map live during gameplay.

------------------------------------------------------------------------

## Final Status

The automapper prototype now:

-   loads world topology
-   navigates via console commands
-   dynamically places rooms
-   shifts the map when necessary
-   preserves deterministic layout
-   maintains strict invariants between `rooms` and `occ`.

The core automapper algorithm is now **functionally stable**.
