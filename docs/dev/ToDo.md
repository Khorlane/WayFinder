# ToDo

- [ ] WF-001 Add a minimal runtime switch to use either current local world navigation (existing proven flow) or a live telnet connection to `www.holyquest.org:7373` via `github.com/reiver/go-telnet`.
- [ ] WF-002 When the temporary `Rooms/` dataset is no longer needed, remove the entire `Rooms/` folder from the project.
- [x] WF-003 Reduce `main.go` scope by extracting mapper core, world loading, and CLI harness into focused internal packages; keep `main.go` as bootstrap/composition wiring only. (commit: a5f7625)
- [x] WF-004 Refactor package/folder layout to architecture-aligned top-level subsystems (`wcs/`, `wne/`, `wmr/`, `solver/`) and remove temporary `internal/harness` naming. (commit: f616a65)
- [x] WF-005 Local dev harness now emits simulated HolyQuest-style room output after room entry and movement attempts, while still showing the discovered-room map with `@` for current room. (commit: 6f9ce48)
- [x] WF-006 Refactor `wmr/runtime.go` to isolate the local/dev harness: move `Run`, `LoadWorld`, `parseRoomFileIntoWorld`, and other local-mode helpers into `wmr/local_mode.go`, leaving `runtime.go` responsible only for mapper/runtime behavior. Simulated HolyQuest-style output in `local_mud_output.go` must remain unchanged. (commit: 3809f64)
- [x] WF-007 ToDo/maintenance hygiene: corrected WF-005 commit tag and added `docs/dev/MAINTENANCE.md` guidance for updating `(commit: pending|TBD)` to exact commit hashes when closing WF items. (commit: f819d22)
- [x] WF-008 Upgrade `docs/dev/scripts/generate_go_source_index.go` to prepend a full repository tree section to `docs/dev/go_source_index.md`, excluding full `Rooms/` expansion by showing only first and last room files (sorted) plus a summarized count. (commit: pending)
