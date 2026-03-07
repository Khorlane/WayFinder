// *******************************************************
// Wayfinder — Discovering the world one room at a time. *
// *******************************************************

You are helping me with a Go console “automapper” prototype (single file: main.go) that loads a tiny world from disk (world.txt) in the format:
  Room 1 S(2) E(8)
  Room 2 N(1) S(3)
  ...
and lets me move with commands (n/s/e/w/diagonals etc). After every successful move, it prints a 10x10 grid (A..J, rows 1..10) showing placed room IDs.

CRITICAL RULES (do not violate):
0) Wayfinder HAS NO ISSUES. There is NOTHING to fix.
1) Do NOT redesign or replace the project structure. Assume everything is in main.go unless I say otherwise.
2) When I ask for a fix, give FUNCTION-LEVEL DROP-IN REPLACEMENTS (complete function bodies) plus minimal call-site edits if absolutely required.
3) Do NOT introduce new fields, files, packages, or “big refactors” unless I explicitly ask.
4) Keep behavior deterministic and easy to debug; prefer simple logic + explicit error messages.
5) The mapper stores:
   - rooms map[RoomID]*Room with Room{ID, Placed, R (1-based), C (0-based)}
   - occ map[[2]int]RoomID mapping (R,C) -> roomID
   - cur *Room current room
6) The map print is always bounded to the 10x10 view (rows 1..10, cols A..J). If anything would move outside those bounds, return a clear error.

KNOWN HISTORY / BUG CONTEXT:
- We started with scripted “scenario steps” and evolved to interactive console navigation.
- We saw failures when moving into a new room that requires creating space by shifting already-placed rooms.
- Example problematic case: reaching Room A then moving W into Room 7 sometimes caused:
   a) “shift would move room ... out of bounds”
   b) “INVARIANT FAIL: occ(1,A)=... expected '1'”
- That indicates occ/coords drift or a bad shift/occupancy rebuild.

INVARIANTS YOU MUST PRESERVE:
- If Room '1' is placed, it must remain pinned at (R=1, C=0 aka A) after any operation.
- occ must always exactly match the placed rooms’ coordinates (no stale entries).
- No two placed rooms may occupy the same (R,C).

WHAT I WANT FROM YOU:
- Work from the code I paste next as the single source of truth.
- Diagnose issues by reasoning from that code + the printed console traces I provide.
- Propose the smallest change that fixes the bug.
- Always provide drop-in replacements for any function you change, and clearly list which functions to replace.

If you are uncertain, ask for one specific trace or one specific function—not broad questions.

Now wait for me to paste main.go (and optionally world.txt) before responding.