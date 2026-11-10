# WayFinder

WayFinder is a native Windows client for the HolyQuest MUD.

## Quick start

To build from source, use Windows and Go 1.23.4 or later. Run these commands
from the project directory.

```powershell
go build -ldflags=-H=windowsgui -o WayFinder.exe .
.\WayFinder.exe
```

WayFinder connects to HolyQuest at `129.212.241.43:7373` and displays the color
transcript in the Output View. Type in the Input Console and press **Enter** to
send a command. Use **Up** and **Down** to recall commands from the current
session. Close the window to disconnect.

Place the HolyQuest `Rooms/` dataset beside `WayFinder.exe` or in the working
directory to enable the Map Panel. The panel reveals rooms as you visit them
and highlights your current location. When HolyQuest enables GMCP, the panel
uses the room IDs sent by the server.

## Other ways to run

- `go run .` opens the live Windows client from source.
- `go run . --play <host:port>` opens the client and connects to a different server.
- `go run . --live <host:port>` streams server output to the terminal in read-only mode.
- `go run . .\Rooms` runs the local simulator with a room dataset in `Rooms/`.

## Project documentation

- [Architecture](docs/architecture.md)
- [Discovery model](docs/discovery_model.md)
- [Mapper rules](docs/mapper_rules.md)

## License

[The Unlicense](LICENSE)
