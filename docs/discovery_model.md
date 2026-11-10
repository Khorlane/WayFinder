# Wayfinder Discovery Model

## Purpose

Wayfinder reveals the map as the player explores the world.

The system separates **world knowledge** from **player discovery**.

------------------------------------------------------------------------

## Authoritative Data

Two structures represent truth:

### World Topology

The graph of known rooms and directed exits.

    roomID → direction → neighborRoomID

The simulator loads the full graph from its room files. In live play the graph
grows from HolyQuest's room reports. Each report with exit destinations supplies
the complete outgoing edges of that room. A destination ID in an exit does not
mark the destination discovered.

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

The map panel displays only the discovered portion of the world.

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
