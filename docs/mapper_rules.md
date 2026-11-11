# Wayfinder Mapper Rules

## Layout Model

The map grid is a **visual projection of the discovered subgraph**.

It is not authoritative and may be rebuilt when necessary.

------------------------------------------------------------------------

## Ordered Alignment Constraints

Directional relationships are enforced as **axis‑aligned ordering
constraints**.

Example relationships:

-   North/South → same column
-   East/West → same row

Strict ordering must be preserved.

Example:

    D3B
    ECE

ECE must always remain **below** D3B.

------------------------------------------------------------------------

## Gap Tolerance

Rooms do not need to be adjacent.

These are all valid:

    D3B
    ECE

    D3B
    .
    ECE

    D3B
    .
    .
    ECE

Gaps are allowed to maintain readable layouts.

------------------------------------------------------------------------

## Lane Reservation

Once two rooms form a locked relationship:

-   no other room may be placed between them on that axis.

Example:

    ECE ... 061

A new room cannot appear between ECE and 061.

Drawn connections may meet at a shared room. A connection crossing another
connection or passing through an unrelated room makes the layout invalid.

------------------------------------------------------------------------

## Layout Rebuild

If incremental placement cannot satisfy constraints:

1.  Gather discovered rooms
2.  Recompute layout
3.  Preserve directional ordering
4.  Allow gaps where needed

This avoids complex incremental repair logic.

Validate each candidate against the traversed direction, every locked lane,
and the connection geometry before replacing the layout. During a rebuild,
try row and column orders that allow independent groups to share coordinates.
Among valid candidates, prefer the smallest footprint and use movement from
the previous layout to break ties.

------------------------------------------------------------------------

## Debug Rendering

Rooms are displayed using stable hash labels for alignment debugging.

Example:

    4E6  9FB  A15

This visualization helps verify spatial relationships during
development.
