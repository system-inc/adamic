package estree

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A copied source tree must keep its product address, while a surviving-mutant
// candidate must get a different one. This catches accidental use of private
// paths in a key and accidental omission of mutable source content.
func TestRecoveryCacheInputKeys(t *testing.T) {
	t.Parallel()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	const before = "(scanner.flags & 1) === 0 &&\n        (token"
	identical := mutantPort(t, "sourceLookahead.ts", before, before)
	mutant := mutantPort(t, "sourceLookahead.ts", before, "(scanner.flags & 1) >= 0 &&\n        (token")
	keys := make([]string, 3)
	for i, path := range []string{main, identical, mutant} {
		keys[i], err = buildcache.Key(root(t), recoveryCacheInputs(t, path))
		if err != nil {
			t.Fatal(err)
		}
	}
	if keys[0] != keys[1] {
		t.Fatal("identical source in a private directory changed the product key")
	}
	if keys[0] == keys[2] {
		t.Fatal("changed mutant source did not change the product key")
	}
}
