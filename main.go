package main

import (
	"os"

	"WayFinder/wmr"
)

// Main boots the current local development runtime path.
// Architectural pipeline target is WTL -> WEG -> WNE -> WMR -> WCS.
func main() {
	os.Exit(wmr.Run(os.Args))
}
