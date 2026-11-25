package adaptedtree

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// The function clone shares the parsed node's JSDoc array before either tag read (#7ez1s16).
// Adamic's 17:57Z ruling on #3qh3m28 keeps the array as the cache key, so the clone must
// read the original's cached @deprecated tag. Stock TypeScript 6.0.3 is the reference
// because the preceding adapters leave its tag worker and parser invalidation intact.
// Mutant: key the adapted table by the node, and the clone's tag read is empty.
func TestJSDocTagsAreSharedByClones(t *testing.T) {
	t.Parallel()
	product := Product(t)
	tags := func(compiler string) []byte {
		t.Helper()
		command := exec.Command("node", "adaptedtree/testdata/jsdoc-cache/clone-test.cjs", compiler)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("JSDoc clone probe from %s: %v\n%s", compiler, err, stderr.Bytes())
		}
		return output
	}
	stock := tags(filepath.Join(repositoryFromLane, "stage3/api/node_modules/typescript"))
	adapted := tags(filepath.Join(product, "tree/built/local/typescript.js"))
	want := []byte("[[332,\"hello\"]]\n")
	if !bytes.Equal(stock, want) || !bytes.Equal(adapted, stock) {
		t.Fatalf("JSDoc clone rows: stock %q, adapted %q, want %q from both", stock, adapted, want)
	}
	t.Logf("stock %sadapted %s", stock, adapted)
}
