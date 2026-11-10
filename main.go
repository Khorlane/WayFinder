package main

import (
	"fmt"
	"os"

	"WayFinder/wcs/win32"
	"WayFinder/wtl"
)

const liveGameAddress = "129.212.241.43:7373"

// Main boots the current runtime entrypoint at the WTL boundary.
// Architectural pipeline target is WTL -> WEG -> WNE -> WMR -> WCS.
func main() {
	if len(os.Args) == 1 {
		win32.RunLiveWCS(liveGameAddress, wtl.PlayLive)
		return
	}
	if os.Args[1] == "--play" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run . --play <host:port>")
			os.Exit(2)
		}
		win32.RunLiveWCS(os.Args[2], wtl.PlayLive)
		return
	}
	if os.Args[1] == "--live" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run . --live <host:port>")
			os.Exit(2)
		}
		if err := wtl.StreamLiveOutput(os.Args[2], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	os.Exit(wtl.Run(os.Args))
}
