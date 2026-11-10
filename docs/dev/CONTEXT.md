# WayFinder development context

Start with [AGENTS.md](../../AGENTS.md) for repository rules and the [README](../../README.md) for build and launch commands. The design lives in [Architecture](../architecture.md), [Discovery model](../discovery_model.md), and [Mapper rules](../mapper_rules.md). Use [ToDo.md](ToDo.md) for open work.

## Current behavior

- **Live play.** The default Win32 app connects to HolyQuest. WTL owns the TCP connection and sends commands with CRLF. WEG decodes ANSI foreground colors and identifies password prompts. WCS displays the scrolling transcript and handles input masking, command echo, and session history. Live play currently ends at text presentation. Room discovery and map layout run in the simulator.
- **Local simulation.** WTL loads the local `Rooms/` dataset and runs the console harness. The simulated gateway interprets commands, WNE tracks navigation and discovery, and WMR computes the discovered-room layout shown in the console. Room-file parsing in WTL belongs to the simulator harness.
- **Live mapping boundary.** The intended navigation flow is WTL → WEG → WNE → WMR → WCS. The Map Panel belongs inside WCS and draws WMR's layout. The [room-output example](examples/holyquest-room-output.txt) shows the room headers, exits, and movement prompts that a live gateway can use.

## Where to start in the code

| Area | Files |
| --- | --- |
| Startup and mode selection | [main.go](../../main.go) |
| Win32 shell and live controls | [shell_windows.go](../../wcs/win32/shell_windows.go), [live_windows.go](../../wcs/win32/live_windows.go) |
| Live transport and local harness | [live_mode.go](../../wtl/live_mode.go), [simulated_mode.go](../../wtl/simulated_mode.go) |
| Game text and simulated commands | [live_gateway.go](../../weg/live_gateway.go), [ansi.go](../../weg/ansi.go), [simulated_gateway.go](../../weg/simulated_gateway.go) |
| Navigation and mapping | [navigation_session.go](../../wne/navigation_session.go), [runtime.go](../../wmr/runtime.go), [solver.go](../../wmr/solver.go) |

## Handoff

- Use `rg --files -g '*.go'` to find current source files and Git history to recover earlier decisions.
- Update this page when a runtime mode or subsystem boundary changes. Keep open work in [ToDo.md](ToDo.md).
- For the current workflow, build the requested change and report the executable path. The maintainer is handling testing for now.
