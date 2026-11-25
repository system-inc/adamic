package adaptedtree

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// A class with two @implements tags keeps both in the declaration the adapted compiler emits (#4w889k2).
// 99-readonly-family3-implements-types turns getImplementsTypes' push into a fresh array, and a writer that keeps only
// the last type (`resolvedImplementsTypes = [implementsType]`) checks clean under stock TypeScript and passes every
// upstream test: the lane's oracle never appends twice in one call (J-tzasvq2, 192 calls, none with two). Stock
// TypeScript 6.0.3, stage3/api's, is the reference, so the adapted declaration is held to its bytes, and the reference
// itself must name both. Mutant: that writer as adapt.cjs's after text, red (the adapted declaration reads
// `class Both implements B`).
func TestTwoImplementsTagsKeepBothInTheDeclaration(t *testing.T) {
	t.Parallel()
	product := Product(t)
	declaration := func(compiler string) []byte {
		t.Helper()
		command := exec.Command("node", "adaptedtree/testdata/implements/declaration.cjs", compiler, "adaptedtree/testdata/implements/fixture.js")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("declaration from %s: %v\n%s", compiler, err, stderr.Bytes())
		}
		return output
	}
	stock := declaration(filepath.Join(repositoryFromLane, "stage3/api/node_modules/typescript"))
	if !bytes.Contains(stock, []byte("class Both implements A, B {")) {
		t.Fatalf("stock TypeScript's declaration names both implements types nowhere, so the fixture proves nothing:\n%s", stock)
	}
	adapted := declaration(filepath.Join(product, "tree/built/local/typescript.js"))
	if !bytes.Equal(adapted, stock) {
		t.Fatalf("the adapted compiler's declaration differs from stock TypeScript's.\nadapted:\n%s\nstock:\n%s", adapted, stock)
	}
}
