# Wayfinder Discovery Model

## Purpose

Wayfinder reveals the map as the player explores the world.

The system separates **world knowledge** from **player discovery**.

------------------------------------------------------------------------

## Authoritative Data

Two structures represent truth:

### World Topology

The full graph of rooms and exits.

    roomID → direction → neighborRoomID

This is immutable after loading.

------------------------------------------------------------------------

### Discovery State

Tracks rooms the player has visited.

    discoveredRooms = set of RoomIDs

When the player enters a room:

    Enter(roomID)
    → mark room discovered

------------------------------------------------------------------------

## Exit Visibility

If a room is discovered, its exits are automatically known.

There are **no hidden exits** in this world design.

Therefore no separate discovered-edge structure is required.

------------------------------------------------------------------------

## Map Rendering

Renderer displays only the discovered portion of the world.

Display: - discovered rooms - their known exits

Hide: - undiscovered rooms - edges leading to undiscovered rooms

------------------------------------------------------------------------

## Layout Projection

The grid layout is generated only for discovered rooms.

Undiscovered rooms are not pre‑placed.

When new discoveries introduce layout conflicts, the mapper may rebuild
the layout for all discovered rooms.

------------------------------------------------------------------------

## Design Goals

-   deterministic behavior
-   readable navigation layout
-   stable directional relationships
-   fog‑of‑war exploration
