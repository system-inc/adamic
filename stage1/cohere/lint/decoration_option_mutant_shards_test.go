package lint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
)

// The decoration hyphen mutant runs on each backend in a top-level test of its own, so Loom's cut can place them apart,
// and the emitted JavaScript's lowering of the mutant port is a keyed product built ahead by
// TestProduct_DecorationOptionMutantJavaScript (#mq768wp). In one test the two runs held it at 25.1 s
// (planner/lane-recut-46's measured seconds) and 11.6 s on Lunatic, Node's run about 2 s of it and the
// in-process lowering nearly all the rest: split by backend alone, the emitted JavaScript's shard was still 13.3 to
// 17.4 s there.
const testDecorationOptionMutantShards = 2

// The mutant's edit: ESLint escapes a `-` decoration as `\x2d`, so it is one more character to skip; the mutant drops
// it from the class as the RE2 range reading did, which the `["-", "/"]` row's "-/ TODO" catches.
const decorationOptionMutantFile = "comments.ts"
const decorationOptionMutantFrom = "if(foldPoint(member) === foldPoint(point))"
const decorationOptionMutantTo = "if(member !== 45 && foldPoint(member) === foldPoint(point))"

// decorationOptionMutantSides are the backends the mutant must be caught on, one shard each.
func decorationOptionMutantSides() []string {
	return []string{"Node", "emitted JavaScript"}
}

func TestDecorationOptionMutant_000(test *testing.T) {
	test.Parallel()
	decorationOptionMutantShard(test, 0)
}
func TestDecorationOptionMutant_001(test *testing.T) {
	test.Parallel()
	decorationOptionMutantShard(test, 1)
}

// decorationOptionMutantShard runs the mutant on one backend over the generated rows, against the Go oracle's answer.
func decorationOptionMutantShard(test *testing.T, shard int) {
	test.Helper()
	path := manifest(test, generated(test))
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", path).output
	side := decorationOptionMutantSides()[shard]
	var run execution
	switch shard {
	case 0:
		run = node(
			test,
			mutant(test, decorationOptionMutantFrom, decorationOptionMutantTo, decorationOptionMutantFile),
			path,
			false,
		)
	case 1:
		run = runJavaScript(test, decorationOptionMutantJavaScript(test), path, false)
	default:
		test.Fatalf("decoration option mutant shard %d runs no backend", shard)
	}
	if err := decorationOptionMutantCheck(side, run.output, want); err != nil {
		test.Fatal(err)
	}
	test.Logf("decoration hyphen dropped from the class caught on %s: %s", side, difference(run.output, want))
}

// Not parallel: a product that isn't lean builds alone in its process.
func TestProduct_DecorationOptionMutantJavaScript(test *testing.T) {
	requireLintArtifacts(test, filepath.Dir(decorationOptionMutantJavaScript(test)), "lint.mjs")
}

// decorationOptionMutantJavaScript is the mutant port lowered to JavaScript, its lint.mjs's path. The compiler lowering
// it is linked into this test binary, where no trace sees it read, so its key names the compiler's paths
// (compilerinputs.Lowering) beside the port, whose copy the build edits, and the parser and scanner it imports.
func decorationOptionMutantJavaScript(test *testing.T) string {
	test.Helper()
	inputs := buildcache.Inputs{
		Name:  "lint-decoration-option-mutant-javascript",
		Files: append(compilerinputs.Lowering(), "stage1/cohere/lint", "stage1/typescript"),
		Flags: []string{
			"copyPort",
			decorationOptionMutantFile,
			decorationOptionMutantFrom,
			decorationOptionMutantTo,
			"registry.Generate",
			"lower.Lower",
			"javascript.JavaScript",
		},
		Toolchain: []string{runtime.Version()},
	}
	product := buildcache.Product(test, inputs, func(directory string) error {
		source, err := os.MkdirTemp("", "decoration-option-mutant-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(source)
		copyPort(test, source, decorationOptionMutantFrom, decorationOptionMutantTo, decorationOptionMutantFile)
		return emitJavaScriptTo(source, filepath.Join(directory, "lint.mjs"))
	})
	return filepath.Join(product, "lint.mjs")
}

// Both the live shards and the planted-failure test use this exact evaluator.
func decorationOptionMutantCheck(side string, got, want []byte) error {
	if bytes.Equal(got, want) {
		return fmt.Errorf("decoration hyphen mutant survived on %s", side)
	}
	return nil
}

func TestDecorationOptionMutantUnion(test *testing.T) {
	test.Parallel()
	// Read actual top-level registrations: a missing, repeated, or misnumbered wrapper must fail even if the side list
	// still names every backend.
	file, err := parser.ParseFile(token.NewFileSet(), "decoration_option_mutant_shards_test.go", nil, 0)
	if err != nil {
		test.Fatal(err)
	}
	registrations := map[int]int{}
	const prefix = "TestDecorationOptionMutant_"
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, prefix)
		shard, err := strconv.Atoi(suffix)
		if err != nil ||
			shard < 0 ||
			shard >= testDecorationOptionMutantShards ||
			suffix != fmt.Sprintf("%03d", shard) {
			test.Fatalf("invalid shard name %s", function.Name.Name)
		}
		registrations[shard]++
		calls := 0
		ast.Inspect(function.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "decorationOptionMutantShard" {
				return true
			}
			calls++
			if len(call.Args) != 2 {
				test.Fatalf("bad wrapper %s", function.Name.Name)
			}
			number, ok := call.Args[1].(*ast.BasicLit)
			if !ok || number.Value != strconv.Itoa(shard) {
				test.Fatalf("wrapper %s runs wrong side", function.Name.Name)
			}
			return true
		})
		if calls != 1 {
			test.Fatalf("wrapper %s calls runner %d times", function.Name.Name, calls)
		}
	}
	if len(decorationOptionMutantSides()) != testDecorationOptionMutantShards {
		test.Fatalf(
			"%d backends for %d shards, want one shard each",
			len(decorationOptionMutantSides()),
			testDecorationOptionMutantShards,
		)
	}
	for shard := range testDecorationOptionMutantShards {
		if registrations[shard] != 1 {
			test.Fatalf("shard %03d registered %d times", shard, registrations[shard])
		}
	}
	test.Logf(
		"union: %d backends, each exactly once across %d top-level shards",
		len(decorationOptionMutantSides()),
		len(registrations),
	)
}

// A surviving mutant, output equal to the Go answer, is planted on one backend: only that backend's check may fail,
// and a caught mutant passes on every backend.
func TestDecorationOptionMutantPlantedFailure(test *testing.T) {
	test.Parallel()
	planted := 1
	want := []byte("case 0\nfixed\tunchanged\n")
	var caught []int
	for shard, side := range decorationOptionMutantSides() {
		got := []byte("case 0\nfixed\tmutated\n")
		if err := decorationOptionMutantCheck(side, got, want); err != nil {
			test.Fatalf("caught mutant failed on %s: %v", side, err)
		}
		if shard == planted {
			got = append([]byte(nil), want...)
		}
		if decorationOptionMutantCheck(side, got, want) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != planted {
		test.Fatalf("planted surviving mutant caught by shards %v", caught)
	}
	test.Logf(
		"planted surviving mutant on %s caught by exactly shard-%03d",
		decorationOptionMutantSides()[planted],
		caught[0],
	)
}
