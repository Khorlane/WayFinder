# ToDo

- [ ] WF-001 Add a minimal runtime switch to use either current local world navigation (existing proven flow) or a live telnet connection to `www.holyquest.org:7373` via `github.com/reiver/go-telnet`.
- [ ] WF-002 When the temporary `Rooms/` dataset is no longer needed, remove the entire `Rooms/` folder from the project.
- [x] WF-003 Reduce `main.go` scope by extracting mapper core, world loading, and CLI harness into focused internal packages; keep `main.go` as bootstrap/composition wiring only. (commit: a5f7625)
