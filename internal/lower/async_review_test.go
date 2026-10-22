package lower

import (
	"errors"
	"strings"
	"testing"
)

// Merely finding an await somewhere in the source is not a consumption proof.
// A conditional observer, an escaping reader or an intervening failure can
// leave the original rejecting task without a consumer.
func TestAsyncPromiseConsumptionPaths(t *testing.T) {
	t.Parallel()
	for _, tail := range []string{
		`const pending = fail(); console.log("done");`,
		`let pending = fail(); pending = Promise.resolve(3); await pending;`,
		`const pending = fail(); if (false) { await pending; }`,
		`const pending = fail(); const read = async (): Promise<number> => await pending; console.log(typeof read);`,
		`const pending = fail(); const other = fail(); await other; await pending;`,
	} {
		_, err := lowerSource(t, `async function fail(): Promise<number> { await Promise.resolve(); throw new Error("lost"); }`+tail)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(gap.What, "a promise not awaited in the same function") {
			t.Fatalf("unproved observer admitted: %s: %v", tail, err)
		}
	}
}
