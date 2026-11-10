# WayFinder development context

Start with [AGENTS.md](../../AGENTS.md) for repository rules and the [README](../../README.md) for build and launch commands. The design lives in [Architecture](../architecture.md), [Discovery model](../discovery_model.md), and [Mapper rules](../mapper_rules.md). Use [ToDo.md](ToDo.md) for open work.

## Current behavior

- **Live play.** The default Win32 app connects to HolyQuest. WTL owns the TCP connection and sends commands with CRLF. WEG decodes ANSI foreground colors and identifies password prompts. WCS displays the scrolling transcript and handles input masking, command echo, and session history.
- **Local simulation.** WTL loads the local `Rooms/` dataset and runs the console harness. The simulated gateway interprets commands, WNE tracks navigation and discovery, and WMR computes the discovered-room layout shown in the console. Room-file parsing in WTL belongs to the simulator harness.
- **Live mapping.** WTL loads the known topology from `Rooms/` and negotiates GMCP. WEG uses the server's `HolyQuest.Room` IDs when GMCP is enabled. On connections without GMCP, WEG recognizes room titles followed by `Exits:`, and WNE resolves them using the known exits and location context. WMR projects confirmed visits, and the Map Panel inside WCS draws rooms and their aligned connections. Live play works without the dataset, with a status message in the panel.

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
- Live sessions append GMCP statuses, Telnet negotiation bytes, and validated room IDs to `WayFinder-gmcp.log` beside the executable. The trace records protocol metadata only.
- For the current workflow, build the requested change and report the executable path. The maintainer is handling testing for now.
