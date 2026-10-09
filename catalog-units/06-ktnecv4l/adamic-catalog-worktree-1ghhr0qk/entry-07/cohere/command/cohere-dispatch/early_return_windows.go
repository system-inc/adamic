package main

import "fmt"

// returnEarly is never reached on Windows, where execute always waits: a child there cannot inherit a
// descriptor beyond its standard three, so there is nowhere to hear a verdict.
func returnEarly(binaryPath string, arguments []string) error {
	return fmt.Errorf("running %s: returning early is not supported on Windows", binaryPath)
}
