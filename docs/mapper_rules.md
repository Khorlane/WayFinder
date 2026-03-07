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

------------------------------------------------------------------------

## Layout Rebuild

If incremental placement cannot satisfy constraints:

1.  Gather discovered rooms
2.  Recompute layout
3.  Preserve directional ordering
4.  Allow gaps where needed

This avoids complex incremental repair logic.

------------------------------------------------------------------------

## Debug Rendering

Rooms are displayed using stable hash labels for alignment debugging.

Example:

    4E6  9FB  A15

This visualization helps verify spatial relationships during
development.
