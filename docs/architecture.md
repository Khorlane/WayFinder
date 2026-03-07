
# WayFinder Architecture

## Overview

WayFinder is a MUD navigation system that builds a readable spatial projection
of **discovered rooms** while a player explores a world.

The system separates:

- **World topology** – authoritative structure of rooms and exits
- **Discovery state** – rooms the player has encountered
- **Spatial projection** – a visual layout computed for readability

The layout is not authoritative. It is a **projection** that may be rebuilt or
repacked when new rooms are discovered.

WayFinder is implemented as a modular system where each component has a clearly
defined responsibility.

---

# System Component Model

WayFinder is composed of five primary components.

```
WayFinder
│
├─ WayFinder Client Shell      (WCS)
├─ WayFinder Event Gateway     (WEG)
├─ WayFinder Navigation Engine (WNE)
├─ WayFinder Map Renderer      (WMR)
└─ WayFinder Map Panel         (WMP)
```

Each component performs a specific role in the navigation pipeline.

---

# Runtime Data Flow

```
Live MUD Server
        │
        ▼
WayFinder Client Shell (WCS)
        │
        ▼
WayFinder Event Gateway (WEG)
        │
        ▼
WayFinder Navigation Engine (WNE)
        │
        ▼
WayFinder Map Renderer (WMR)
        │
        ▼
WayFinder Map Panel (WMP)
```

This pipeline ensures clean separation between networking, parsing, navigation
logic, and UI rendering.

---

# Component Responsibilities

## WayFinder Client Shell (WCS)

The Client Shell is the host application.

Responsibilities:

- Telnet transport using `github.com/reiver/go-telnet`
- UI framework provided by `github.com/fyne-io/fyne`
- Command input
- Display of MUD text output
- Window management
- Routing events between components

The shell is responsible for the **overall application lifecycle**.

The shell does not perform navigation or map layout logic.

---

## WayFinder Event Gateway (WEG)

The Event Gateway converts raw MUD output into normalized navigation events.

Responsibilities:

- Parse room text
- Detect exits
- Detect player movement
- Emit structured events

Example events:

```
EnterRoom(roomID)
ExitsSeen(roomID, exits)
PlayerCommand(command)
```

The gateway isolates fragile text parsing from the rest of the system.

---

## WayFinder Navigation Engine (WNE)

The Navigation Engine is the core mapping system.

Responsibilities:

- Maintain discovery state
- Maintain spatial layout of discovered rooms
- Enforce layout constraints
- Repack or rebuild layouts when conflicts occur

The engine does **not** perform rendering.

It produces an internal representation of the map state.

---

## WayFinder Map Renderer (WMR)

The Map Renderer converts navigation state into a visual representation.

Responsibilities:

- Convert room coordinates into renderable structures
- Preserve spatial readability
- Apply spacing and layout rules
- Generate a map model for display

The renderer does not compute navigation logic.

---

## WayFinder Map Panel (WMP)

The Map Panel is the UI component that displays the map.

Responsibilities:

- Display the rendered map
- Handle resizing
- Refresh when map state changes
- Provide future support for zoom and panning

The panel reads renderer output but does not modify navigation state.

---

# Discovery Model

WayFinder builds the map **incrementally** as the player explores.

Rules:

- The system only lays out **discovered rooms**
- Undiscovered rooms are not placed on the grid
- Discovery occurs when the player enters a room

When a room is discovered:

```
room → discovered
exits → revealed
```

There are **no hidden exits** in the world design.

---

# Spatial Layout Model

The map is a **readable projection**, not a rigid geometric model.

Constraints use **ordered alignment rules**.

Examples:

- Rooms may remain on the same row with gaps between them.
- Rooms may remain on the same column with gaps between them.
- Relative direction is preserved.

Example valid vertical relationships:

```
A
.
B
```

Example valid horizontal relationships:

```
A ... B
```

Rules:

- ordering must remain correct
- gaps are allowed
- rooms cannot be inserted between locked pairs

---

# Layout Rebuild Strategy

When a new discovery creates a layout conflict:

1. Attempt local adjustments
2. If conflict remains, rebuild layout from discovered topology
3. Preserve ordering constraints during rebuild

The rebuild uses the **discovered subgraph only**.

Undiscovered world topology is not used for layout.

---

# Architectural Principles

WayFinder follows strict separation of responsibilities.

- Networking is isolated from navigation logic
- Parsing is isolated from navigation logic
- Navigation logic contains no UI code
- Rendering contains no navigation logic
- UI contains no mapping logic

This separation keeps the system maintainable and safe to evolve.

---

# Future Expansion

Possible future components:

- persistent map storage
- multiple render styles
- zoomable graphical maps
- trigger/automation systems

The current architecture supports these without modifying the navigation core.
