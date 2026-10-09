package ir

import "strings"

// Throws distinguishes JavaScript's catchable failed read from an Adamic type invariant panic.
// Defined messages for reads carry the same TypeError prefix as Node's uncaught error text.
func (d Defined) Throws() bool {
	return strings.HasPrefix(d.Message, "TypeError: ")
}
