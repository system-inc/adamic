package main

import "time"

// On Windows the dispatcher always waits for the engine (takeVerdictFile), so no run writes the table
// after its caller has moved on and there is nothing to keep in order.

func holdTableLock(directory string) func() {
	return func() {}
}

func holdTableReadLock(directory string, bound time.Duration) (func(), bool) {
	return func() {}, true
}
