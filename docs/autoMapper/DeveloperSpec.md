# AutoMapper -- Developer Design Specification

## Purpose

This document defines the architecture and algorithm design for the
Go-based automapper prototype used in a MUD-style exploration system.

It is intended to serve as the long-term technical reference for
developers working on the automapper.

------------------------------------------------------------------------

# 1. System Overview

The automapper builds a 2D spatial representation of rooms while the
player explores the world.

The mapper does not know the full map ahead of time. Instead, it
constructs the layout dynamically as movement events occur.

Core responsibilities:

-   Place rooms relative to each other
-   Detect when a target cell is already occupied
-   Create space by shifting sections of the map
-   Maintain consistency between coordinate storage and occupancy grid

------------------------------------------------------------------------

# 2. Coordinate System

Grid size (current prototype):

Rows: 1..10\
Columns: A..J

Internal storage:

Row = 1-based integer\
Column = 0-based integer (A=0)

Example coordinate:

(R=3, C=B)

Internally stored as:

R=3\
C=1

------------------------------------------------------------------------

# 3. Data Structures

## Room

type Room struct { ID RoomID Placed bool R, C int }

Fields:

ID --- unique room identifier\
Placed --- whether the room is on the map\
R,C --- grid coordinates

------------------------------------------------------------------------

## Mapper

type Mapper struct { rooms map\[RoomID\]*Room occ map\[\[2\]int\]RoomID
cur *Room }

rooms → authoritative room coordinate database

occ → occupancy grid mapping (row,col) to room ID

cur → current player location

------------------------------------------------------------------------

# 4. Movement Pipeline

Movement is processed through:

mapper.Enter(roomID, directionMoved)

Steps:

1.  Determine expected coordinates relative to the current room
2.  If the target cell is occupied → create space
3.  Shift necessary rooms
4.  Place or align the destination room
5.  Update occupancy grid
6.  Verify invariants

------------------------------------------------------------------------

# 5. Direction Vectors

  Direction   ΔRow   ΔCol
  ----------- ------ ------
  N           -1     0
  S           +1     0
  E           0      +1
  W           0      -1
  NE          -1     +1
  NW          -1     -1
  SE          +1     +1
  SW          +1     -1

Expected location:

expR = current.R + dr\
expC = current.C + dc

------------------------------------------------------------------------

# 6. Shift Algorithm

If a coordinate is occupied, the mapper calls:

shiftWhere(predicate, dr, dc)

Where:

predicate(room) decides which rooms move\
dr,dc is the shift direction

------------------------------------------------------------------------

# 7. Deterministic Shift Requirement

Go map iteration order is random.

Therefore shifting must use four phases:

1.  Collect rooms to move
2.  Validate bounds
3.  Apply shifts
4.  Rebuild occupancy grid

Rooms are authoritative.\
Occupancy is reconstructed.

------------------------------------------------------------------------

# 8. Map Anchor

Room '1' is pinned:

(R=1, C=A)

This stabilizes the map orientation and prevents drift.

------------------------------------------------------------------------

# 9. Invariants

Invariant 1: Room 1 must remain at (1,A)

Invariant 2: occ\[(R,C)\] must equal the room ID stored in rooms

------------------------------------------------------------------------

# 10. Console Simulation

Commands:

n s e w ne nw se sw u d\
look\
map\
coords\
quit

Movement calls Enter().\
The map prints after every move.

------------------------------------------------------------------------

# 11. Example Result

Exploration produced:

1 . . 8\
2 . . 9\
3 6 7 A\
4 . . B\
5 . . C

Demonstrates corridor insertion and dynamic shifting.

------------------------------------------------------------------------

# 12. Future Enhancements

-   Move grid size into Mapper struct
-   Support larger maps
-   Highlight current player position
-   Handle loops more intelligently
-   Integrate with live MUD telnet stream

------------------------------------------------------------------------

# 13. Core Architectural Principle

Rooms are the source of truth.

The occupancy grid is derived from room coordinates.

------------------------------------------------------------------------

# Status

The automapper algorithm is now stable and capable of dynamically
placing and shifting rooms during exploration.
