//go:build !darwin && !linux && !windows

package main

import (
	"fmt"
	"runtime"
)

// availableMemory has no reader on the platforms cohere does not ship for, so their runs keep Go's
// default collector rather than run with collection off and nothing to bound it.
func availableMemory() (uint64, error) {
	return 0, fmt.Errorf("no reader on %s", runtime.GOOS)
}
