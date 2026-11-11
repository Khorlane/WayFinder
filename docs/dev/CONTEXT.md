# WayFinder development context

Start with [AGENTS.md](../../AGENTS.md) for repository rules and the [README](../../README.md) for build and launch commands. The design lives in [Architecture](../architecture.md), [Discovery model](../discovery_model.md), and [Mapper rules](../mapper_rules.md). Use [ToDo.md](ToDo.md) for open work.

## Current behavior

- **Live play.** The default Win32 app connects to HolyQuest. WTL owns the TCP connection and sends commands with CRLF. WEG decodes ANSI foreground colors and identifies password prompts. WCS displays the scrolling transcript and handles input masking, command echo, and session history.
- **Local simulation.** WTL loads the local `Rooms/` dataset and runs the console harness. The simulated gateway interprets commands, WNE tracks navigation and discovery, and WMR computes the discovered-room layout shown in the console. Room-file parsing in WTL belongs to the simulator harness.
- **Live mapping.** WTL negotiates GMCP. WEG decodes `HolyQuest.Room`, and WNE builds the session's topology from room IDs and complete outgoing exit destinations. WMR projects confirmed visits, and the Map Panel inside WCS draws rooms and their aligned connections. WTL loads `Rooms/` on demand for older GMCP messages that omit destinations or for text-based room discovery.

## Where to start in the code

| Area | Files |
| --- | --- |
| Startup and mode selection | [main.go](../../main.go) |
| Win32 shell, live controls, and map drawing | [shell_windows.go](../../wcs/win32/shell_windows.go), [live_windows.go](../../wcs/win32/live_windows.go), [map_windows.go](../../wcs/win32/map_windows.go) |
| Live transport and local harness | [live_mode.go](../../wtl/live_mode.go), [telnet_stream.go](../../wtl/telnet_stream.go), [simulated_mode.go](../../wtl/simulated_mode.go) |
| Game text and events | [live_gateway.go](../../weg/live_gateway.go), [gmcp.go](../../weg/gmcp.go), [room_stream.go](../../weg/room_stream.go), [ansi.go](../../weg/ansi.go), [simulated_gateway.go](../../weg/simulated_gateway.go) |
| Navigation and mapping | [live_navigation.go](../../wne/live_navigation.go), [navigation_session.go](../../wne/navigation_session.go), [live_layout.go](../../wmr/live_layout.go), [runtime.go](../../wmr/runtime.go), [solver.go](../../wmr/solver.go) |

## Handoff

- Use `rg --files -g '*.go'` to find current source files and Git history to recover earlier decisions.
- Update this page when a runtime mode or subsystem boundary changes. Keep open work in [ToDo.md](ToDo.md).
- Live sessions append GMCP statuses, Telnet negotiation bytes, validated room IDs, exit counts, and mapper placement and rebuild diagnostics to `WayFinder-gmcp.log` beside the executable.
- `WayFinder-mapper.jsonl` appends a `view` record from the same snapshot the Map Panel receives on login, room changes, and updates to the current room's exits. Match rebuild records by PID and attempt number to find the previous layout, rejected candidates, and the applied layout or failure.
- A failed rebuild opens the map editor and pauses command input. The capture records correction attempts and any validation error under the failed rebuild's PID and attempt number.
- **Edit map** opens the same editor during normal play once rooms are visible. Drag a box from empty space to select rooms. Ctrl-drag adds another box, and Ctrl-click adds or removes an individual room. Drag a selected room to move the group along both axes, or Shift-drag a row or column. **Save** applies a valid layout and stays in the editor. **Cancel** resets the draft to the last accepted layout, or to the opening draft if nothing has been saved. **Exit** attempts a save and resumes play on success.
- For the current workflow, build the requested change and report the executable path. The maintainer is handling testing for now.
