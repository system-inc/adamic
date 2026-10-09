package estree

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMutantEmittedProductSnapshotIdentity(t *testing.T) {
	t.Parallel()
	makeSource := func(directory, text string) string {
		path := filepath.Join(directory, "main.ts")
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	first := makeSource(t.TempDir(), "console.log('first');")
	relocated := makeSource(t.TempDir(), "console.log('first');")
	original := mutantEmittedProduct(t, first)
	if got := mutantEmittedProduct(t, relocated); got != original {
		t.Fatalf("identical relocated snapshots must share the emitted product: %s != %s", got, original)
	}
	makeSource(filepath.Dir(relocated), "console.log('changed');")
	changed := mutantEmittedProduct(t, relocated)
	if changed == original {
		t.Fatal("changed snapshot reused the old emitted product")
	}
	before, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(changed)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("changed snapshot emitted the old JavaScript")
	}
}
