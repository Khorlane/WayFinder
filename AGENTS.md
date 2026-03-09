# WayFinder agent instructions

## Project architecture

WayFinder consists of five components:

- WayFinder Client Shell (WCS)
- WayFinder Event Gateway (WEG)
- WayFinder Navigation Engine (WNE)
- WayFinder Map Renderer (WMR)
- WayFinder Telnet Layer (WTL)

Pipeline:

WTL → WEG → WNE → WMR → WCS

## Hard rules

- Preserve the architecture above.
- Do not move navigation logic into UI code.
- Do not move parsing logic into the navigation engine.
- Do not move rendering logic into the navigation engine.
- Keep world topology and discovery state authoritative.
- Treat the grid/layout as a rebuildable projection.

## Navigation model

- A room becomes discovered when entered.
- If a room is discovered, its exits are known.
- There are no hidden exits in this world design.
- Undiscovered rooms are not placed on the map.

## Spatial model

- Use ordered directional constraints, not rigid one-cell adjacency.
- Preserve directional ordering.
- Gaps are allowed.
- No room may appear between locked ordered pairs on the same axis.

## Repo conventions

- Keep `main.go` in the project root unless explicitly asked otherwise.
- Keep architecture docs in `docs/`.
- Keep AI/developer context in `docs/dev/`.
- Prefer small, surgical changes over broad refactors.
- Do not rename major components without explicit instruction.

## Before making changes

Read these files first when relevant:

- `README.md`
- `docs/architecture.md`
- `docs/discovery_model.md`
- `docs/mapper_rules.md`
- `docs/dev/llm_context.txt`

## Validation

Before finalizing code changes:

- run relevant tests
- preserve existing behavior unless the task explicitly changes behavior
- avoid introducing architectural drift