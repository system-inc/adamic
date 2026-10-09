package main

import (
	"context"
	"os/exec"
)

// Build and enumeration are setup: the unit runner owns their deadline.
func inventoryEngineSetupCommand(name string, arguments ...string) *exec.Cmd {
	return exec.CommandContext(context.Background(), name, arguments...)
}
