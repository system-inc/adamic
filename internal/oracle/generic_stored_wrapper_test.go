package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A generic body that stores a wrapper, host.create = () => original(), is lowered per instance with
// its T read from the call's own instantiation when T stands only inside an anonymous object's
// property, which inference by reference and signature never reaches (tsc's executeCommandLine.ts:984,
// the Seats scout's reduction, #vks2w2y). wrapper is the reduction, wrapper_union the same with T a
// boxed union, and wrapper_two two instances of one body; each agrees with Node natively, sanitized
// and on the JavaScript backend, and leaks nothing.
func TestGenericStoredWrapperAgrees(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"wrapper.ts", "wrapper_union.ts", "wrapper_two.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, "internal/oracle/testdata/generic_stored_wrapper/"+name)
		})
	}
}

// widened is the generic-body-rigid twin (docs/rulings.md, the generic body): T read from the call
// must not loosen T's rigidity inside the body, so host seen as { create: () => string | number } and
// written a concrete arrow through that view stays refused at compile time, as the invariant-mutable widening
// at the view's declaration (2:50).
func TestGenericStoredWrapperWidenedIsRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_stored_wrapper/widened.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("widened.ts: got %v, want a *lower.Refused", err)
	}
	if !strings.HasSuffix(refused.Where, "widened.ts:2:50") || !strings.HasPrefix(refused.What, "a value of type { create: () => T; } seen as { create: () => string | number; }") {
		t.Fatalf("widened.ts: Refused at %s: %q, want the widening at 2:50", refused.Where, refused.What)
	}
}
