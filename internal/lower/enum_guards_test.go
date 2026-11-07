package lower

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// The numeric IR has no flag-domain annotation. Query the actual checker-backed
// proof separately so an open numeric enum cannot hide a broken flag proof.
func enumGuardProof(test *testing.T, source string) (*lowering, []*ast.Node) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "proof.a")
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	test.Cleanup(release)
	return &lowering{program: loaded, checker: checked, result: &ir.Program{}}, loaded.Files()[0].Statements.Nodes
}

func assertEnumGuardFlag(test *testing.T, source string) {
	test.Helper()
	l, statements := enumGuardProof(test, source)
	if !l.flagEnum(l.symbol(statements[0].Name())) {
		test.Fatal("flag domain not recognized, including its literal initializers and bit 30")
	}
}

func assertEnumGuardDomains(test *testing.T) {
	test.Helper()
	source := flagDeclaration + `
function left(a: Flags, n: number): Flags { return a & n; }
function right(a: Flags, n: number): Flags { return n & a; }
function both(a: Flags, b: Flags): Flags { return a & b; }
function neither(a: number, b: number): number { return a & b; }
function orNumber(a: Flags, n: number): number { return a | n; }
function xorFlags(a: Flags, b: Flags): Flags { return a ^ b; }
`
	l, statements := enumGuardProof(test, source)
	target := l.symbol(statements[0].Name())
	if !l.flagEnum(target) {
		test.Fatal("flag domain must include bit 30")
	}
	wants := []bool{true, true, true, false, false, true}
	for index, statement := range statements[1:] {
		function := statement.AsFunctionDeclaration()
		value := function.Body.AsBlock().Statements.Nodes[0].AsReturnStatement().Expression
		if got := l.flagDomain(value, target); got != wants[index] {
			test.Errorf("%s flag domain = %v, want %v", statement.Name().Text(), got, wants[index])
		}
	}
}

func TestEnumMemberValuesAndReverseNameMatchNode(test *testing.T) {
	test.Parallel()
	source := `enum E { A = 1, B = 2, Alias = A } console.log(E.A + ' ' + E.B + ' ' + E.Alias + ' ' + E[1]);`
	program := lowersAndAgreesWithNode(test, source)
	binary := filepath.Join(test.TempDir(), "native")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, combinedOutputError := exec.CommandContext(ctx, binary).CombinedOutput()
	if combinedOutputError != nil || string(output) != "1 2 1 Alias\n" {
		test.Fatalf("native: %v, output %q", combinedOutputError, output)
	}
}

// Open numeric enum lowering bypasses these classification proofs, so changing
// them does not change emitted JavaScript. Direct proof checks are necessary to
// guard M14, M15 and M16; output comparisons alone cannot observe them.
func TestEnumFlagProofsWithoutObservableLoweringEffect(test *testing.T) {
	test.Parallel()
	assertEnumGuardFlag(test, "enum Flags { None = 0x0, A = 0x1 << 0b0, High = 0o1 << 0x1e }")
	assertEnumGuardDomains(test)
}

// JavaScript object literals overwrite duplicate keys, hiding M04 in output.
// Native objects need a unique field slot, so guard that invariant directly.
func TestEnumReverseMappingUsesSingleSlot(test *testing.T) {
	test.Parallel()
	// JavaScript overwrites duplicate keys; behavior cannot observe native field-slot uniqueness.
	program := lowersAndAgreesWithNode(test, "enum E { A = 1, B = 2, Alias = A } console.log(E[1] ?? 'missing');")
	found := false
	for _, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok || program.Locals[declaration.Local].Name != "E" {
			continue
		}
		object, ok := declaration.Value.(ir.ObjectLiteral)
		if !ok {
			test.Fatal("enum object allocation is missing")
		}
		found = true
		seen := map[string]bool{}
		for _, field := range object.Fields {
			if seen[field.Name] {
				test.Fatalf("duplicate enum field slot %q", field.Name)
			}
			seen[field.Name] = true
		}
	}
	if !found {
		test.Fatal("enum declaration is missing")
	}
}
