package lower

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// The numeric IR has no flag-domain annotation. Query the actual checker-backed
// proof separately so an open numeric enum cannot hide a broken flag proof.
func enumGuardProof(t *testing.T, source string) (*lowering, []*ast.Node) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proof.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	t.Cleanup(release)
	return &lowering{program: loaded, checker: checked, result: &ir.Program{}}, loaded.Files()[0].Statements.Nodes
}

func assertEnumGuardFlag(t *testing.T, source string) {
	t.Helper()
	l, statements := enumGuardProof(t, source)
	if !l.flagEnum(l.symbol(statements[0].Name())) {
		t.Fatal("flag domain not recognized, including its literal initializers and bit 30")
	}
}

func assertEnumGuardDomains(t *testing.T) {
	t.Helper()
	source := flagDeclaration + `
function left(a: Flags, n: number): Flags { return a & n; }
function right(a: Flags, n: number): Flags { return n & a; }
function both(a: Flags, b: Flags): Flags { return a & b; }
function neither(a: number, b: number): number { return a & b; }
function orNumber(a: Flags, n: number): number { return a | n; }
function xorFlags(a: Flags, b: Flags): Flags { return a ^ b; }
`
	l, statements := enumGuardProof(t, source)
	target := l.symbol(statements[0].Name())
	if !l.flagEnum(target) {
		t.Fatal("flag domain must include bit 30")
	}
	wants := []bool{true, true, true, false, false, true}
	for index, statement := range statements[1:] {
		function := statement.AsFunctionDeclaration()
		value := function.Body.AsBlock().Statements.Nodes[0].AsReturnStatement().Expression
		if got := l.flagDomain(value, target); got != wants[index] {
			t.Errorf("%s flag domain = %v, want %v", statement.Name().Text(), got, wants[index])
		}
	}
}

func enumGuardNodeOutput(t *testing.T, source string, program *ir.Program, want string, buildNative bool) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	run := func(name, executable string, arguments ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, arguments...)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: error %v, stdout %q, stderr %q", name, err, stdout, stderr.Bytes())
		}
		return stdout
	}
	truth := run("source Node", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(truth) != want {
		t.Fatalf("source Node stdout = %q, want %q", truth, want)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := run("JavaScript backend", "node", "--disable-warning=ExperimentalWarning", runner, generated); !bytes.Equal(got, truth) {
		t.Errorf("JavaScript backend stdout = %q, source Node = %q", got, truth)
	}
	if buildNative {
		binary := filepath.Join(directory, "native")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		if got := run("native backend with ASan/UBSan", binary); !bytes.Equal(got, truth) {
			t.Errorf("native backend stdout = %q, source Node = %q", got, truth)
		}
	}
}

func TestEnumMemberValuesAndReverseNameMatchNode(t *testing.T) {
	t.Parallel()
	source := `enum E { A = 1, B = 2, Alias = A } console.log(E.A + ' ' + E.B + ' ' + E.Alias + ' ' + E[1]);`
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if program == nil {
		t.Fatal("lowering returned empty IR")
	}
	enumGuardNodeOutput(t, source, program, "1 2 1 Alias\n", true)
}

// Open numeric enum lowering bypasses these classification proofs, so changing
// them does not change emitted JavaScript. Direct proof checks are necessary to
// guard M14, M15 and M16; output comparisons alone cannot observe them.
func TestEnumFlagProofsWithoutObservableLoweringEffect(t *testing.T) {
	t.Parallel()
	assertEnumGuardFlag(t, "enum Flags { None = 0x0, A = 0x1 << 0b0, High = 0o1 << 0x1e }")
	assertEnumGuardDomains(t)
}

// JavaScript object literals overwrite duplicate keys, hiding M04 in output.
// Native objects need a unique field slot, so guard that invariant directly.
func TestEnumReverseMappingUsesSingleSlot(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "enum E { A = 1, B = 2, Alias = A } console.log(E[1] ?? 'missing');")
	if err != nil {
		t.Fatal(err)
	}
	if program == nil {
		t.Fatal("lowering returned empty IR")
	}
	found := false
	for _, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok || program.Locals[declaration.Local].Name != "E" {
			continue
		}
		object, ok := declaration.Value.(ir.ObjectLiteral)
		if !ok {
			t.Fatal("enum object allocation is missing")
		}
		found = true
		seen := map[string]bool{}
		for _, field := range object.Fields {
			if seen[field.Name] {
				t.Fatalf("duplicate enum field slot %q", field.Name)
			}
			seen[field.Name] = true
		}
	}
	if !found {
		t.Fatal("enum declaration is missing")
	}
}
