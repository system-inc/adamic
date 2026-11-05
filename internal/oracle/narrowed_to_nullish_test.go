package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A read the checker narrowed to undefined or null, of a binding or a property a call may write, stops
// as not yet, in .a and .ts (#r0vzaaa). The checker keeps the narrowing across a call it doesn't look
// inside, and a consumer that asks it for the read's type folds the read to that value, so both backends
// printed undefined where Node prints 4. Each stops at the read: the callee writes the module binding
// (n01), the null twin (n03), a let a closure the call runs writes (n04), a parameter the same way (n08),
// a module binding another module's function writes, read in the importing module under its own check
// (n05), a module binding its own setter writes, called from the module it imports (n06,
// n06_module_setter/back calling setLater), and ??= in the callee (n07, #aw65psm's g05). A write of
// another form counts as one: ++ (n09, which writes nothing on this run, so Node prints undefined and
// the stop is the rule's conservative answer), an array or object destructuring target (n10, n11) and a
// for-of head (n12), each in a callee declared below the read. There's no compound read here: tsc
// refuses `s += 'a'` on a binding narrowed to undefined (TS2322), so it never reaches lowering. A
// property narrowed the same way stops on every read, since nothing cheap proves no alias writes it: a
// class field a callee writes (f01), an object literal's field (f02), a template read of it (f03) and
// the null twin (f04). Their controls lower and agree with Node on release native, sanitized native and
// the JavaScript backend: a callee that doesn't write the binding (n02), a local no other function
// writes, read narrowed after a call (c01), a binding narrowed to number before the call that writes it
// (c02, narrowed.go's own road), a field the callee writes but nothing narrowed (c03), and an imported
// binding read after the call that sets it, which the checker doesn't narrow in the importing module
// (n14).
// Mutant: narrowedNullishReadGap's binding test forced to let every binding through, and every binding
// stop here lowers instead, printing undefined where Node prints 4; its property test the same, and
// f01 to f04 lower.
func TestNarrowedToNullishReadsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/narrowed_to_nullish"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q: Node prints %q", stop, onNode(t, absolute).stdout)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			truth := onNode(t, absolute)
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 40 {
		t.Fatalf("found %d programs, want 40", programs)
	}
}
