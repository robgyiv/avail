package main

import (
	"os"

	// Embed the IANA timezone database so zone lookups work on systems that ship
	// without one (minimal containers, for instance). Both the configured timezone
	// and the TZID values in .ics files are resolved through time.LoadLocation.
	_ "time/tzdata"

	"github.com/robgyiv/avail/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
