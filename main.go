package main

import (
	"os"

	"WayFinder/wtl"
)

// Main boots the current runtime entrypoint at the WTL boundary.
// Architectural pipeline target is WTL -> WEG -> WNE -> WMR -> WCS.
func main() {
	os.Exit(wtl.Run(os.Args))
}
