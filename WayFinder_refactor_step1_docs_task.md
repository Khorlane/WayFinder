# WayFinder Refactor -- Step 1 Documentation Alignment Task

## Objective

Align project documentation to the Step 1 architecture boundary:

- WTL is the only source of inbound MUD text for the runtime pipeline.
- Runtime pipeline is: WTL -> WEG -> WNE -> WMR -> WCS.
- Simulated mode is a development harness; live mode is the future production text source.

This task is documentation-only.

------------------------------------------------------------------------

# Constraints

Follow these rules strictly:

- Do NOT modify `.go` source files.
- Do NOT modify runtime behavior.
- Do NOT move files or directories.
- Do NOT rename subsystems.

Allowed changes:

- Markdown/text documentation updates only.

------------------------------------------------------------------------

# Files To Align

Update these files so they are consistent with the Step 1 boundary:

- `docs/architecture.md` (authoritative architecture description)
- `AGENTS.md` (agent guardrails and pipeline statement)
- `README.md` (public architecture and pipeline summary)
- `docs/dev/llm_context.txt` (AI architectural guidance)
- `docs/dev/chat_context.txt` (session bootstrap context)

When relevant, ensure wording in these files does not conflict:

- `docs/discovery_model.md`
- `docs/mapper_rules.md`
- `docs/dev/ToDo.md`

------------------------------------------------------------------------

# Required Documentation Outcomes

Ensure docs clearly state:

- Five major components: WCS, WEG, WNE, WMR, WTL.
- WMP remains a panel hosted inside WCS and is not a top-level pipeline stage.
- WTL owns inbound MUD text sourcing in both modes:
  - Simulated MUD mode (development harness)
  - Live MUD mode (future network mode)
- Both modes feed equivalent raw MUD text into WEG.
- WCS still owns user command capture and presentation concerns.

------------------------------------------------------------------------

# Validation

After edits:

1. Regenerate `docs/dev/go_source_index.md` with the existing repository script.
2. Confirm generated output succeeds without errors.
3. Confirm no non-documentation files changed.

------------------------------------------------------------------------

# Completion Criteria

Step 1 documentation alignment is complete when:

- All target docs consistently describe the same runtime pipeline and component boundaries.
- No file outside documentation/context/index generation outputs is modified.
- `go_source_index.md` regeneration succeeds.
- No functional behavior changes are introduced.
