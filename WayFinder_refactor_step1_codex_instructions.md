# WayFinder Refactor -- Step 1 Instructions for Codex

## Objective

Establish the architectural boundary that **WTL is the only source of
inbound MUD text for the runtime pipeline**.

At this stage **no behavioral changes and no structural changes** should
be made to the system.\
The goal is to **lock in the architectural rule clearly in the codebase**
so future refactors follow the correct design.

The authoritative pipeline must be:

WTL → WEG → WNE → WMR → WCS

WTL is the **only source of inbound MUD text for the runtime pipeline**.

Two future operating modes must be recognized:

1.  **Simulated MUD mode** -- development harness (current room files
    and navigation simulator)
2.  **Live MUD mode** -- real MUD server connection (not implemented
    yet)

Both modes must eventually produce the **same MUD text stream** consumed
by the rest of the system.

The simulator must be treated strictly as a **development harness**, not
as an authoritative world model.

------------------------------------------------------------------------

# Constraints

Follow these rules strictly:

-   Do **NOT** create new packages.
-   Do **NOT** restructure directories.
-   Do **NOT** rename subsystems.
-   Do **NOT** change runtime behavior.
-   Do **NOT** move files.

The repository structure should remain unchanged.

Only non-functional architectural clarification comments should be added.

------------------------------------------------------------------------

# Tasks

## 1. Add Clear Responsibility Notes

Add high‑level comments describing the responsibility
of each subsystem:

WTL\
Handles the **source of MUD text**.\
Future modes: simulated or live.

WEG\
Parses raw MUD text and produces structured WayFinder events.

WNE\
Maintains discovered navigation topology.

WMR\
Computes spatial map projection from the discovered topology.

WCS\
Displays output, captures commands, and renders the map.

------------------------------------------------------------------------

## 2. Clarify the Simulator Role

Where the simulator or local-mode logic exists, add a brief comment
stating:

-   This logic represents **Simulated MUD mode**
-   It exists only as a **development harness**
-   In the finished system, **WTL live mode will replace this as the
    primary source**

No code movement or redesign should happen yet.

------------------------------------------------------------------------

# Completion Criteria

Step 1 is complete when:

-   WTL is documented as the **only inbound MUD text boundary**.
-   Simulated mode is described as a **development harness**.
-   The repository structure is unchanged.
-   No functional behavior changes.

No functional behavior should change.
