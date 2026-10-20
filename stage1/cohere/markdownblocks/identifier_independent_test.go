package markdownblocks

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode"
)

// identifierScalarsSides are the checks over all 1112064 Unicode scalars, each its own top-level test so Loom's
// planner can cut between them; in one test the six full-range children ran one after another. Go's answer is made
// once per process and every side reads it.
var identifierScalarsSides = []struct {
	name  string
	check func(t *testing.T, ctx context.Context, products pilotProducts, truth run)
}{
	{"sanitized native", func(t *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(t, malformedEventsExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized), truth)
	}},
	{"source Node", func(t *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(t, malformedEventsNode(t, ctx, products.main), truth)
	}},
	{"JavaScript backend", func(t *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(t, malformedEventsNode(t, ctx, products.backend), truth)
	}},
	{"leaks", identifierScalarsLeaks},
	{"case table mutant", identifierScalarsMutant},
}

const identifierScalarsShards = 5

var identifierScalarsTruth struct {
	once  sync.Once
	truth run
	ready bool
}

// identifierScalarsGo is actual Go NormalizeIdentifier over every scalar, run once per process.
func identifierScalarsGo(t *testing.T, ctx context.Context, products pilotProducts) run {
	t.Helper()
	identifierScalarsTruth.once.Do(func() {
		truth := malformedEventsExecute(t, ctx, nil, products.goBinary)
		clean(t, "actual Go NormalizeIdentifier", truth)
		identifierScalarsTruth.truth, identifierScalarsTruth.ready = truth, true
	})
	if !identifierScalarsTruth.ready {
		t.Fatal("actual Go NormalizeIdentifier failed earlier in this process")
	}
	return identifierScalarsTruth.truth
}

// identifierScalarsDisagreement is a side's verdict against Go: a clean exit with Go's exact bytes, or why not.
func identifierScalarsDisagreement(side, truth run) error {
	if side.exitCode != 0 || len(side.stderr) > 0 {
		return fmt.Errorf("exit %d stderr %s", side.exitCode, side.stderr)
	}
	if !bytes.Equal(side.stdout, truth.stdout) {
		return fmt.Errorf("first difference at byte %d", firstDifference(string(side.stdout), string(truth.stdout)))
	}
	return nil
}

func identifierScalarsAgree(t *testing.T, side, truth run) {
	t.Helper()
	if err := identifierScalarsDisagreement(side, truth); err != nil {
		t.Fatalf("identifier scalar oracle: %v", err)
	}
}

func identifierScalarsLeaks(t *testing.T, ctx context.Context, products pilotProducts, truth run) {
	t.Helper()
	if runtime.GOOS == "darwin" {
		report := malformedEventsExecute(t, ctx, nil, "leaks", "--atExit", "--", products.release)
		if report.exitCode != 0 {
			t.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	} else {
		report := malformedEventsExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized)
		if report.exitCode != 0 {
			t.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	}
}

func identifierScalarsMutant(t *testing.T, ctx context.Context, products pilotProducts, truth run) {
	t.Helper()
	scratch := t.TempDir()
	if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
		t.Fatal(e)
	}
	for _, file := range []string{"identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "testdata/identifier_probe.ts"} {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if file == "identifierCaseKeys.ts" {
			if !strings.Contains(string(data), "65,") {
				t.Fatal("case table mutation anchor")
			}
			data = []byte(strings.Replace(string(data), "65,", "64,", 1))
		}
		write(t, filepath.Join(scratch, file), data)
	}
	mutant := malformedEventsNode(t, ctx, filepath.Join(scratch, "testdata/identifier_probe.ts"))
	clean(t, "case table mutant", mutant)
	if bytes.Equal(mutant.stdout, truth.stdout) {
		t.Fatal("case table mutant survived")
	}
	t.Logf("output-only case-table mutant caught at byte%d", firstDifference(string(mutant.stdout), string(truth.stdout)))
}

func identifierScalarsShard(t *testing.T, index int) {
	t.Helper()
	products := identifierProductsReady(t)
	configureMarkdownMemory(t)
	markdownMemory.acquire(4)
	defer markdownMemory.release(4)
	ctx, finish := malformedEventsDeadline(t, t.Name())
	defer finish()
	truth := identifierScalarsGo(t, ctx, products)
	identifierScalarsSides[index].check(t, ctx, products, truth)
	t.Logf("%s: all1112064 Unicode scalars against actual Go normalization; Go Unicode%s", identifierScalarsSides[index].name, unicode.Version)
}

// identifierScalarsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired shard
// can't leave a side unchecked while the count still agrees.
func identifierScalarsEnumeration(t *testing.T) []int {
	t.Helper()
	if len(identifierScalarsSides) != identifierScalarsShards {
		t.Fatalf("%d identifier sides, but identifierScalarsShards is %d", len(identifierScalarsSides), identifierScalarsShards)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "identifier_independent_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestMdastIdentifierScalars_") {
			continue
		}
		shard, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestMdastIdentifierScalars_"))
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "identifierScalarsShard" {
				return true
			}
			if len(call.Args) != 2 {
				t.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, err := strconv.Atoi(literal.Value); err != nil || index != shard {
				t.Fatalf("%s runs side %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			t.Fatalf("%s runs %d sides, want one", function.Name.Name, calls)
		}
		plan = append(plan, shard)
	}
	return plan
}

func TestMdastIdentifierScalarsUnion(t *testing.T) {
	t.Parallel()
	visits := make([]int, len(identifierScalarsSides))
	for _, index := range identifierScalarsEnumeration(t) {
		if index < 0 || index >= len(visits) {
			t.Fatalf("TestMdastIdentifierScalars_%03d names no side", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			t.Fatalf("side %d (%s) run by %d top-level tests, want one", index, identifierScalarsSides[index].name, count)
		}
	}
	t.Logf("union: all %d identifier sides run, each by exactly one top-level test", len(visits))
}

func TestMdastIdentifierScalarsPlantedFailure(t *testing.T) {
	t.Parallel()
	// Plant one wrong byte in one comparing side's answer. Exactly its owner must reject it; every other comparing
	// side, given Go's own bytes, stays accepted.
	truth := run{stdout: []byte("Go identifier bytes\n")}
	var caught []int
	for _, shard := range identifierScalarsEnumeration(t) {
		if shard > 2 {
			continue
		}
		side := run{stdout: bytes.Clone(truth.stdout)}
		if shard == 1 {
			side.stdout[0] = 'g'
		}
		if identifierScalarsDisagreement(side, truth) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 1 {
		t.Fatalf("planted disagreement caught by %v, want [1]", caught)
	}
	t.Log("planted disagreement caught exactly by TestMdastIdentifierScalars_001")
}

func TestMdastIdentifierScalars_000(t *testing.T) { t.Parallel(); identifierScalarsShard(t, 0) }
func TestMdastIdentifierScalars_001(t *testing.T) { t.Parallel(); identifierScalarsShard(t, 1) }
func TestMdastIdentifierScalars_002(t *testing.T) { t.Parallel(); identifierScalarsShard(t, 2) }
func TestMdastIdentifierScalars_003(t *testing.T) { t.Parallel(); identifierScalarsShard(t, 3) }
func TestMdastIdentifierScalars_004(t *testing.T) { t.Parallel(); identifierScalarsShard(t, 4) }
