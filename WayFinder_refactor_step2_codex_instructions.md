# WayFinder Refactor – Step 2 Instructions for Codex

## Objective
Isolate the **Simulated MUD development harness** under the **WTL boundary** without changing behavior.

At the end of this step, the design should clearly reflect:

- **WTL is the only source of inbound MUD text**
- **WTL has two operating modes**
  - Simulated MUD
  - Live MUD (future, not implemented yet)
- The current room-file loading, simulated navigation, and simulated MUD text generation are treated as **WTL simulated-mode responsibilities**
- Downstream architectural intent remains:

**WTL → WEG → WNE → WMR → WCS**

This step is still a **controlled refactor pass**, not a redesign.

---

## Constraints

Follow these rules strictly:

- Do **NOT** create new major components.
- Do **NOT** introduce package proliferation.
- Keep the repository structure **as flat as practical**.
- Prefer **moving or re-homing existing responsibilities** over inventing new abstractions.
- Do **NOT** implement live telnet connectivity.
- Do **NOT** change observable behavior.
- Do **NOT** change the current simulated-mode user experience.
- Do **NOT** move mapping ownership into WTL.
- Do **NOT** move parsing ownership into WTL.
- Do **NOT** move navigation authority into WTL.

This step is about **isolating the simulator as an input source**, not changing who owns navigation, mapping, or parsing.

---

## Architectural Target for This Step

### WTL responsibilities after Step 2
WTL should own the **input-source side** of the simulator:

- starting simulated mode
- loading the local room dataset as simulator backing data
- executing simulated movement mechanics needed to generate fake MUD output
- producing simulated MUD-style text
- exposing that text through the same conceptual boundary that future live mode will use

### Responsibilities that must remain outside WTL
- **WEG**: interpretation/parsing of MUD text
- **WNE**: authoritative discovered navigation state
- **WMR**: spatial map projection/layout
- **WCS**: UI shell and rendering surface

---

## Tasks

## 1. Identify simulator-specific code currently outside WTL

Review current simulator/local-mode code and isolate the parts that are simulator/input-harness responsibilities, including things such as:

- local room/world loading used only for simulation
- simulated movement execution used to produce fake room transitions
- simulated MUD output generation
- simulator startup/run path

The goal is to distinguish:
- **simulator input generation**
from
- **core navigation/mapping responsibilities**

---

## 2. Re-home simulator responsibilities under WTL

Move the simulator/input-harness responsibilities into the **existing WTL area**.

Guidance:
- Reuse the existing `wtl/` area.
- Keep structure flat.
- Avoid nested folders unless absolutely necessary.
- Prefer a small number of files with clear responsibility.

Expected conceptual result:
- **WTL(simulated)** becomes the owner of the current local simulated MUD source
- WMR should no longer appear to own the simulator harness
- WMR should remain the mapping runtime only

Important:
This is a **responsibility isolation move**, not a behavior rewrite.

---

## 3. Preserve current simulated behavior

After the move:
- local testing should still work
- room-file-backed simulation should still work
- simulated MUD text output should still look the same from the user’s perspective

Any interfaces/adapters needed to preserve behavior should be kept minimal and practical.

---

## 4. Keep core ownership boundaries intact

After this step, the codebase should better reflect:

- WTL = source of simulated MUD text
- WEG = parser/interpreter
- WNE = navigation/discovery authority
- WMR = mapping/layout authority
- WCS = shell/presentation

Do not let the simulator refactor accidentally:
- move topology authority into WTL
- move projection logic into WTL
- move parser logic into WMR or WCS
- create duplicate ownership of room/discovery truth

---

## 5. Update documentation/context only where required

Update only the docs/context files that must change because the simulator ownership has moved.

Examples:
- `docs/dev/chat_context.txt`
- `docs/dev/llm_context.txt`
- `README.md` or `docs/architecture.md` only if a current implementation note is now inaccurate

Do not do broad doc rewriting.

---

## 6. Regenerate source index

After the refactor:
- regenerate `docs/dev/go_source_index.md`

This file will be used for architectural verification in the next review pass.

---

## Completion Criteria

Step 2 is complete when:

- The simulator/local-mode harness is clearly isolated under **WTL**
- The repository remains flat and controlled
- No extra architectural layers were introduced
- Simulated mode behavior still works
- WMR no longer appears to own simulator/input-source responsibilities
- `docs/dev/go_source_index.md` is regenerated
- The project still reflects the five-component model:
  **WCS, WEG, WNE, WMR, WTL**

## Anti-goals

Do **not** do any of the following in this step:

- implement live telnet
- redesign WEG
- redesign WNE
- redesign WMR
- introduce a sixth component
- create deep package trees
- perform speculative cleanup unrelated to simulator isolation
