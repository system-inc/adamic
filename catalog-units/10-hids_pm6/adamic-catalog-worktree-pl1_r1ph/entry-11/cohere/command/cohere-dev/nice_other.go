//go:build !unix

package main

// lowerPriority does nothing where there is no nice value. The machines the house develops on are Unix.
func lowerPriority(int) error {
	return nil
}
