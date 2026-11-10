# WayFinder backlog

- [ ] **WF-001** Evaluate whether the live client needs Telnet negotiation beyond the current TCP text session. The live connection and mode selection shipped in `3d670a7`.
- [ ] **WF-002** Retire the local `Rooms/` dataset once the simulator no longer needs it.

## Live mapping

- [ ] Have WEG recognize room entries and exits in [HolyQuest output](examples/holyquest-room-output.txt), then send navigation events to WNE.
- [ ] Display WMR's discovered-room layout in the Win32 Map Panel.

Record the implementation commit when closing a `WF-*` item. Earlier completed items are in `git show 96b2e26:docs/dev/ToDo.md`.
