# WayFinder Refactor – Step 4 Instructions for Codex

## Objective
Remove remaining **cross-boundary leakage** so each of the five WayFinder components more cleanly reflects its intended ownership.

By the end of this step, the codebase should still behave the same, but the structure should more clearly express:

- **WTL** = source of MUD text (simulated now, live later)
- **WEG** = interpretation of inbound MUD text into WayFinder events/state transitions
- **WNE** = authoritative navigation/discovery owner
- **WMR** = mapping/layout projection owner
- **WCS** = UI shell / presentation

This step is about **tightening boundaries**, not changing features.

---

## Constraints

- Do NOT create new packages.
- Keep repository layout flat.
- Do NOT add major abstractions unless necessary.
- Do NOT change user-visible behavior.
- Do NOT implement live telnet yet.
- Do NOT redesign the map solver or navigation model.
- Prefer small, surgical refactors.

---

## Primary Goal

Identify places where one component still knows too much about another component’s internals and reduce that leakage.

Examples of leakage to look for:
- WTL carrying navigation or discovery authority instead of just simulation/input duties
- WEG owning presentation/debug view concerns that belong elsewhere
- WMR exposing simulator-specific assumptions
- WCS depending on core internals rather than receiving finished output/state
- duplicated direction normalization or duplicated interpretation logic spread across layers

Do not chase every tiny duplication. Focus on **meaningful architectural leakage**.

---

## Tasks

### 1. Review the new Step 2 / Step 3 structure

Inspect:
- `wtl/simulated_mode.go`
- `weg/simulated_gateway.go`
- `wne/navigation_session.go`
- `wmr/runtime.go`
- `wmr/dev_view.go`
- `main.go`

Look for boundary violations or awkward ownership after the previous passes.

---

### 2. Tighten ownership where leakage remains

Apply only targeted refactors that improve ownership clarity.

Examples of acceptable changes:
- move helper logic to the component that truly owns it
- reduce simulator-specific knowledge outside WTL
- reduce parser/event-normalization logic outside WEG
- remove obvious duplicated interpretation helpers if they blur ownership
- simplify adapter boundaries if they currently expose too much internal structure

Examples of unacceptable changes:
- broad redesign
- speculative cleanup
- creating a utility package
- moving core responsibilities between the five major components

---

### 3. Preserve WNE and WMR authority

After this step:
- WNE should still be the owner of discovered navigation state
- WMR should still be the owner of spatial projection/layout
- WEG should not absorb mapping logic
- WTL should not absorb navigation authority

If a helper exists in the wrong place, move only that helper, not the whole subsystem responsibility.

---

### 4. Keep debug/dev support practical

If debug or developer-view helpers remain necessary, keep them lightweight and owned by the most appropriate existing component.

Do not introduce a new “debug” subsystem.

---

### 5. Minimal docs/context updates only if necessary

Update docs/context only where the current implementation description becomes inaccurate after the refactor.

Likely candidates:
- `docs/dev/chat_context.txt`
- `docs/dev/llm_context.txt`

Avoid broad documentation edits.

---

### 6. Regenerate source index

After the pass:
- regenerate `docs/dev/go_source_index.md`

This will be used for verification.

---

## Completion Criteria

Step 4 is complete when:

- The five-component ownership model is clearer in the codebase
- Remaining cross-boundary leakage is reduced
- No extra packages or components were introduced
- Behavior remains unchanged
- `docs/dev/go_source_index.md` is regenerated

## Reminder

This is the **last cleanup pass before preparing a clean slot for future Live MUD mode**.

Do not implement Live MUD here.
