package ir

import "strings"

// Throws distinguishes JavaScript's catchable failed read from an Adamic type invariant panic.
// Defined messages for reads carry the same TypeError prefix as Node's uncaught error text.
func (defined Defined) Throws() bool {
	return strings.HasPrefix(defined.Message, "TypeError: ")
}
