package dispatch

import (
	"errors"
	"os/exec"
)

// ErrNoToolchain reports that no Go toolchain is available to build with.
//
// It is a named error rather than a message because the dispatcher reports it as the reason a local
// build was not attempted, alongside the separate failure of not finding a shipped binary. The two
// call for different fixes — install Go, or fix the install — and a reader needs to be told which.
var ErrNoToolchain = errors.New("no Go toolchain found on PATH")

// HasToolchain reports whether a Go toolchain is available.
//
// This decides which half of the dispatcher runs: with a toolchain the rules on disk are the truth
// and the rebuild cache serves them, without one the shipped platform binary is the only answer.
// Finding the shipped binary belongs to `internal/release`, which resolves it out of the installed
// npm platform package rather than from a directory beside the source.
func HasToolchain() bool {
	_, err := exec.LookPath("go")
	return err == nil
}
