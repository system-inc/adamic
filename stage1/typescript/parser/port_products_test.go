package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every port a test builds natively through buildPort is a product, and each has its TestProduct_ here, so the tree
// build makes it ahead and a runner, which only reads the cache, finds it. The control's sanitized build is
// TestProduct_WholeMutantsNative_Control's.

type jsxNameBoundaryChange struct{ name, from, to, control string }

// jsxNameBoundaryChanges are TestJsxNameBoundaryRejections' two permissive mutants of jsx.ts, each with the name
// boundary case it must accept.
func jsxNameBoundaryChanges() []jsxNameBoundaryChange {
	return []jsxNameBoundaryChange{
		{"private name accepted", "if(this.parser.kind() !== 'Identifier' && !this.parser.kind().endsWith('Keyword'))", "if(this.parser.kind() !== 'Identifier' && !this.parser.kind().endsWith('Keyword') && this.parser.kind() !== 'PrivateIdentifier')", "private member"},
		{"escaped name accepted", "if((this.parser.scanner.flags & 1032) !== 0)", "if((this.parser.scanner.flags & 0) !== 0)", "escaped tag"},
	}
}

type jsxScannerChange struct{ name, from, to string }

// jsxScannerChanges are TestJsxScannerMutants' two mutants of the scanner's JSX payloads.
func jsxScannerChanges() []jsxScannerChange {
	return []jsxScannerChange{
		{"raw attribute value", "this.value = this.text.slice(start, this.pos);\n        if(this.pos === this.text.length)", "this.value = this.text.slice(start, this.pos) + '!';\n        if(this.pos === this.text.length)"},
		{"JSX name payload", "this.value += this.text.slice(start, this.pos);", "this.value += this.text.slice(start, this.pos) + '!';"},
	}
}

// jsxScannerMutant copies the parser unchanged and gives it its own scanner module, with change planted in it.
func jsxScannerMutant(t *testing.T, change jsxScannerChange) string {
	t.Helper()
	mutant := copyPort(t, "", "", "")
	scannerDirectory := filepath.Join(mutant, "scanner")
	if err := os.Mkdir(scannerDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	originalScanner, _ := filepath.Abs("../scanner/scanner.ts")
	for _, file := range []string{"scanner.ts", "characters.ts", "tokens.ts"} {
		data, err := os.ReadFile(filepath.Join("../scanner", file))
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if file == "scanner.ts" {
			if strings.Count(source, change.from) != 1 {
				t.Fatal("scanner mutant must change exactly one site")
			}
			source = strings.Replace(source, change.from, change.to, 1)
		}
		if err := os.WriteFile(filepath.Join(scannerDirectory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range portFiles {
		path := filepath.Join(mutant, file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := strings.ReplaceAll(string(data), originalScanner, filepath.Join(scannerDirectory, "scanner.ts"))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return mutant
}

// countTreeMutant is the port with countTree starting from zero: every printed AST byte stays, and only a node-count
// check can catch it. TestNodeCountCheckCatchesMutant and TestWholeCountCheckCatchesMutant share it.
func countTreeMutant(t *testing.T) string {
	t.Helper()
	from := "export function countTree(nodes: readonly ParseNode[], root: number): number {\n    const node = nodes[root] ?? panic('missing parse node');\n    let count = 1;"
	return copyPort(t, "nodes.ts", from, strings.Replace(from, "let count = 1;", "let count = 0;", 1))
}

func TestProduct_ParserPortRelease(t *testing.T) {
	t.Parallel()
	buildPort(t, compilerExpressionsProductDirectory(t), false)
}
func TestProduct_JsxNameBoundaryNative_000(t *testing.T) {
	t.Parallel()
	change := jsxNameBoundaryChanges()[0]
	buildPort(t, copyPort(t, "jsx.ts", change.from, change.to), true)
}
func TestProduct_JsxNameBoundaryNative_001(t *testing.T) {
	t.Parallel()
	change := jsxNameBoundaryChanges()[1]
	buildPort(t, copyPort(t, "jsx.ts", change.from, change.to), true)
}
func TestProduct_JsxScannerNative_000(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxScannerMutant(t, jsxScannerChanges()[0]), true)
}
func TestProduct_JsxScannerNative_001(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxScannerMutant(t, jsxScannerChanges()[1]), true)
}
func TestProduct_CountTreeNative(t *testing.T) {
	t.Parallel()
	buildPort(t, countTreeMutant(t), true)
}
