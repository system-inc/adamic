package estree

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A copied source tree must keep its product address, while a surviving-mutant
// candidate must get a different one. This catches accidental use of private
// paths in a key and accidental omission of mutable source content.
func TestRecoveryCacheInputKeys(test *testing.T) {
	test.Parallel()
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	const before = "(scanner.flags & 1) === 0 &&\n        (token"
	identical := mutantPort(test, "sourceLookahead.ts", before, before)
	mutant := mutantPort(test, "sourceLookahead.ts", before, "(scanner.flags & 1) >= 0 &&\n        (token")
	keys := make([]string, 3)
	for i, path := range []string{main, identical, mutant} {
		var keyError error
		keys[i], keyError = buildcache.Key(root(test), recoveryCacheInputs(test, path))
		if keyError != nil {
			test.Fatal(keyError)
		}
	}
	if keys[0] != keys[1] {
		test.Fatal("identical source in a private directory changed the product key")
	}
	if keys[0] == keys[2] {
		test.Fatal("changed mutant source did not change the product key")
	}
}
