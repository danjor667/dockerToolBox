// Command dockertoolbox is the entry point for DockerToolBox.
//
// v0.1 ships as a standalone binary. The command tree lives in
// internal/commands so a Docker CLI plugin entry point can be added
// later without moving any command logic.
package main

import (
	"fmt"
	"os"

	"dockerToolBox/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
