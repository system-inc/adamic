package overlaytext

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// The carry skips these bodies and replays their first-ask module-read effects.
// Fingerprint normalized AST bodies: parse without comments and format with an
// empty FileSet, so comments and source positions are not part of the contract.
var carryReplayBodies = map[string]string{
	"entriesProgramHasRecords":  "d08d63eabd0fb9f08c9750de9bce09b724ecbaf5ec5c3b52dfdd19117100157a",
	"entriesProgramRecordsWalk": "852371609dd722f09b94d30b04116ede41a67f0ab6d89f9390c69fa236f85fa9",
	"entriesRecordLiteral":      "f15d4f56c50e7b98d231f5ee967b76caee5bc9c66d50bd7e7143aa7d328ada98",
	"entriesRecordLiteralWalk":  "ee77f34596203a624dbeded6e2576020d94eb6c0208b3df2a879508d350ab68a",
}

func carryReplayBodyGaps(source []byte) []string {
	const proof = "carry replay was proven against the old bodies; re-prove it with the subset byte-identity run in stage3/census/latent/evidence/carry-facts/report.md"
	file, err := parser.ParseFile(token.NewFileSet(), "entries_records.go", source, 0)
	if err != nil {
		return []string{fmt.Sprintf("entries_records.go: %v; %s", err, proof)}
	}
	actual := map[string]string{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Body == nil {
			continue
		}
		pointer, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		owner, ok := pointer.X.(*ast.Ident)
		if !ok || owner.Name != "lowering" {
			continue
		}
		if _, ok := carryReplayBodies[fn.Name.Name]; !ok {
			continue
		}
		var body bytes.Buffer
		if err := format.Node(&body, token.NewFileSet(), fn.Body); err != nil {
			return []string{fmt.Sprintf("entries_records.go: %s: %v; %s", fn.Name.Name, err, proof)}
		}
		actual[fn.Name.Name] = fmt.Sprintf("%x", sha256.Sum256(body.Bytes()))
	}
	names := []string{}
	for name := range carryReplayBodies {
		names = append(names, name)
	}
	sort.Strings(names)
	gaps := []string{}
	for _, name := range names {
		if got := actual[name]; got != carryReplayBodies[name] {
			gaps = append(gaps, fmt.Sprintf(
				"entries_records.go: %s body changed: %s (got %s, want %s)",
				name,
				proof,
				got,
				carryReplayBodies[name],
			))
		}
	}
	return gaps
}

func TestLatentCarryFactsReplayBodiesMatchMain(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile(filepath.Join("../../../../internal/lower", "entries_records.go"))
	if err != nil {
		test.Fatal(err)
	}
	for _, gap := range carryReplayBodyGaps(source) {
		test.Error(gap)
	}
}
