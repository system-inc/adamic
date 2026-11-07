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
	check func(test *testing.T, ctx context.Context, products pilotProducts, truth run)
}{
	{"sanitized native", func(test *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(
			test,
			malformedEventsExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized),
			truth,
		)
	}},
	{"source Node", func(test *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(test, malformedEventsNode(test, ctx, products.main), truth)
	}},
	{"JavaScript backend", func(test *testing.T, ctx context.Context, products pilotProducts, truth run) {
		identifierScalarsAgree(test, malformedEventsNode(test, ctx, products.backend), truth)
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
func identifierScalarsGo(test *testing.T, ctx context.Context, products pilotProducts) run {
	test.Helper()
	identifierScalarsTruth.once.Do(func() {
		truth := malformedEventsExecute(test, ctx, nil, products.goBinary)
		clean(test, "actual Go NormalizeIdentifier", truth)
		identifierScalarsTruth.truth, identifierScalarsTruth.ready = truth, true
	})
	if !identifierScalarsTruth.ready {
		test.Fatal("actual Go NormalizeIdentifier failed earlier in this process")
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

func identifierScalarsAgree(test *testing.T, side, truth run) {
	test.Helper()
	if identifierScalarsDisagreementError := identifierScalarsDisagreement(
		side,
		truth,
	); identifierScalarsDisagreementError != nil {
		test.Fatalf("identifier scalar oracle: %v", identifierScalarsDisagreementError)
	}
}

func identifierScalarsLeaks(test *testing.T, ctx context.Context, products pilotProducts, truth run) {
	test.Helper()
	if runtime.GOOS == "darwin" {
		report := malformedEventsExecute(test, ctx, nil, "leaks", "--atExit", "--", products.release)
		if report.exitCode != 0 {
			test.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	} else {
		report := malformedEventsExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized)
		if report.exitCode != 0 {
			test.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	}
}

func identifierScalarsMutant(test *testing.T, ctx context.Context, products pilotProducts, truth run) {
	test.Helper()
	scratch := test.TempDir()
	if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for _, file := range []string{
		"identifier.ts",
		"identifierCaseKeys.ts",
		"identifierCaseValues.ts",
		"testdata/identifier_probe.ts",
	} {
		data, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
		}
		if file == "identifierCaseKeys.ts" {
			if !strings.Contains(string(data), "65,") {
				test.Fatal("case table mutation anchor")
			}
			data = []byte(strings.Replace(string(data), "65,", "64,", 1))
		}
		write(test, filepath.Join(scratch, file), data)
	}
	mutant := malformedEventsNode(test, ctx, filepath.Join(scratch, "testdata/identifier_probe.ts"))
	clean(test, "case table mutant", mutant)
	if bytes.Equal(mutant.stdout, truth.stdout) {
		test.Fatal("case table mutant survived")
	}
	test.Logf(
		"output-only case-table mutant caught at byte%d",
		firstDifference(string(mutant.stdout), string(truth.stdout)),
	)
}

func identifierScalarsShard(test *testing.T, index int) {
	test.Helper()
	products := identifierProductsReady(test)
	configureMarkdownMemory(test)
	markdownMemory.acquire(4)
	defer markdownMemory.release(4)
	ctx, finish := malformedEventsDeadline(test, test.Name())
	defer finish()
	truth := identifierScalarsGo(test, ctx, products)
	identifierScalarsSides[index].check(test, ctx, products, truth)
	test.Logf(
		"%s: all1112064 Unicode scalars against actual Go normalization; Go Unicode%s",
		identifierScalarsSides[index].name,
		unicode.Version,
	)
}

// identifierScalarsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired shard
// can't leave a side unchecked while the count still agrees.
func identifierScalarsEnumeration(test *testing.T) []int {
	test.Helper()
	if len(identifierScalarsSides) != identifierScalarsShards {
		test.Fatalf(
			"%d identifier sides, but identifierScalarsShards is %d",
			len(identifierScalarsSides),
			identifierScalarsShards,
		)
	}
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "identifier_independent_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestMdastIdentifierScalars_") {
			continue
		}
		shard, parseError := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestMdastIdentifierScalars_"))
		if parseError != nil {
			test.Fatal(parseError)
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
				test.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				test.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, parseError := strconv.Atoi(literal.Value); parseError != nil || index != shard {
				test.Fatalf("%s runs side %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			test.Fatalf("%s runs %d sides, want one", function.Name.Name, calls)
		}
		plan = append(plan, shard)
	}
	return plan
}

func TestMdastIdentifierScalarsUnion(test *testing.T) {
	test.Parallel()
	visits := make([]int, len(identifierScalarsSides))
	for _, index := range identifierScalarsEnumeration(test) {
		if index < 0 || index >= len(visits) {
			test.Fatalf("TestMdastIdentifierScalars_%03d names no side", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			test.Fatalf(
				"side %d (%s) run by %d top-level tests, want one",
				index,
				identifierScalarsSides[index].name,
				count,
			)
		}
	}
	test.Logf("union: all %d identifier sides run, each by exactly one top-level test", len(visits))
}

func TestMdastIdentifierScalarsPlantedFailure(test *testing.T) {
	test.Parallel()
	// Plant one wrong byte in one comparing side's answer. Exactly its owner must reject it; every other comparing
	// side, given Go's own bytes, stays accepted.
	truth := run{stdout: []byte("Go identifier bytes\n")}
	var caught []int
	for _, shard := range identifierScalarsEnumeration(test) {
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
		test.Fatalf("planted disagreement caught by %v, want [1]", caught)
	}
	test.Log("planted disagreement caught exactly by TestMdastIdentifierScalars_001")
}

func TestMdastIdentifierScalars_000(test *testing.T) {
	test.Parallel()
	identifierScalarsShard(test, 0)
}
func TestMdastIdentifierScalars_001(test *testing.T) {
	test.Parallel()
	identifierScalarsShard(test, 1)
}
func TestMdastIdentifierScalars_002(test *testing.T) {
	test.Parallel()
	identifierScalarsShard(test, 2)
}
func TestMdastIdentifierScalars_003(test *testing.T) {
	test.Parallel()
	identifierScalarsShard(test, 3)
}
func TestMdastIdentifierScalars_004(test *testing.T) {
	test.Parallel()
	identifierScalarsShard(test, 4)
}
