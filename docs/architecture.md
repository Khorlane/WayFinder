# WayFinder Architecture

## System Overview

WayFinder is a MUD navigation client and automapper written in Go.
The system discovers rooms dynamically and constructs a spatial map as the
player moves through the world.

The architecture separates UI, parsing, navigation, mapping, and solver logic
so each subsystem has a clearly defined responsibility.

---

## System Model

WayFinder consists of five primary subsystems plus solver support:

- **WCS** — WayFinder Client Shell
- **WEG** — WayFinder Event Gateway
- **WNE** — WayFinder Navigation Engine
- **WMR** — WayFinder Mapping Runtime
- **WTL** — WayFinder Telnet Layer
- **solver** — constraint/rebuild support used by WMR

Runtime flow:

**WTL → WEG → WNE → WMR → WCS**

Important:

- **WMP is hosted inside WCS as a UI panel**
- **WMP is not part of the runtime pipeline**
- **WMR computes layout but does not render UI**

---

## Core Subsystems

## WCS — WayFinder Client Shell

The application UI shell.

Current implementation uses a **Win32 native window**.

Responsibilities:

- create the main application window
- display MUD output
- capture command input
- host UI panels
- route UI events and text to runtime components

WCS is responsible for presentation and user interaction.
WCS does not own navigation logic, topology logic, or map layout rules.

---

## WMP — WayFinder Map Panel

A UI panel hosted inside WCS.

Responsibilities:

- render the map produced by WMR
- display the current player position
- display discovered rooms
- present map output to the user

Important:

- WMP performs rendering only
- WMP contains no parsing logic
- WMP contains no navigation logic
- WMP contains no topology or placement logic
- WMP is **not part of the runtime pipeline**

WMP is a visual surface, not an authority on map state.

---

## WEG — WayFinder Event Gateway

Responsible for parsing raw MUD output and converting it into normalized events
used by the navigation system.

Example events:

- room entered
- exits discovered
- movement failure

Responsibilities:

- accept raw game text from WCS/telnet input
- identify navigation-relevant information
- emit normalized events for WNE consumption

WEG does not own navigation session state or map layout.

---

## WNE — WayFinder Navigation Engine

Maintains the navigation session.

Responsibilities:

- track the current room
- expose movement operations
- maintain discovered topology
- coordinate discovery updates

WNE owns navigation state and discovered-world state, but does not render UI and
does not compute spatial layout.

---

## WMR — WayFinder Mapping Runtime

Maintains the internal map model and layout.

Responsibilities:

- incremental room placement
- enforce spatial constraints
- maintain adjacency relationships
- invoke solver support when layout rebuild is required
- produce the room layout consumed by WMP

Important:

WMR computes layout but **does not render UI**.

WMR is the spatial authority for the discovered map layout.

---

## solver — Constraint / Rebuild Support

Solver support is used by WMR when incremental placement cannot preserve all
required constraints directly.

Responsibilities:

- validate constraint sets
- compute rebuild results
- support recovery from placement conflicts

The solver is not a UI subsystem and not a top-level runtime stage.
It is a support component used by WMR.

---

## Runtime Pipeline

Primary runtime flow:

**WTL → WEG → WNE → WMR → WCS**

Rendering occurs inside WCS using the WMP panel.

Interpretation of the flow:

1. **WTL** receives raw MUD text (from simulated mode or live telnet)
2. **WEG** parses raw game text into normalized events
3. **WNE** updates navigation/discovery state
4. **WMR** updates spatial map layout
5. **WCS** displays output and renders the map using WMP

This preserves a clean split between UI, parsing, navigation, and mapping.

---

## Development Modes

## Local Mode

Local simulation environment used during development.

Characteristics:

- world loaded from local room files
- simulated MUD output generated locally
- supports local navigation/testing without live telnet

Primary implementation files:

- `wmr/local_mode.go`
- `wmr/local_mud_output.go`

---

## Live Mode

Real MUD connection via telnet.

Planned integration library:

- `github.com/reiver/go-telnet`

Live mode should preserve the same architectural split:
WCS handles UI, WEG parses text, WNE manages navigation state, and WMR manages
map layout.

---

## Architectural Guardrails

These rules define the intended structure of the system:

1. Keep **WCS** as the UI shell.
2. Keep **WMP** as a rendering panel inside WCS.
3. Keep **WEG** as the parsing/event-normalization layer.
4. Keep **WNE** as the navigation/discovery authority.
5. Keep **WMR** as the mapping/layout authority.
6. Keep solver logic as WMR support, not as a UI or pipeline stage.
7. Do not merge rendering into WMR.
8. Do not move navigation state into WCS or WMP.
9. Do not introduce alternate architectural layers unless explicitly requested.

---

## Current Status Snapshot

Current state at a high level:

- mapping runtime implemented
- solver implemented
- navigation session implemented
- local development harness implemented
- telnet integration not yet complete
- WEG parser not yet complete
- WCS event wiring not yet complete
- WMP map rendering not yet complete

This architecture document describes the intended structure that current and
future work should follow.
