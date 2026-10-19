package native

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const decodePrefixes = 92014
const decodePrefixUnitSize = 2048
const normalizePointUnitSize = 65536

type testUnitRange struct{ first, last int }

func (piece testUnitRange) name() string {
	return fmt.Sprintf("cases_%06d_%06d", piece.first, piece.last)
}
func unitRanges(total, width int) []testUnitRange {
	var pieces []testUnitRange
	for first := 0; first < total; first += width {
		pieces = append(pieces, testUnitRange{first, min(first+width, total)})
	}
	return pieces
}

// Frozen counts describe the old workloads. A missing, duplicated or overlapping
// piece fails independently of whether all the remaining observations agree.
func TestSplitUnitCoverage(t *testing.T) {
	t.Parallel()
	for _, plan := range []struct {
		name                string
		total, width, units int
	}{
		{"decode_prefixes", 92014, decodePrefixUnitSize, 45},
		{"normalize_points", 0x110000, normalizePointUnitSize, 17},
	} {
		t.Run(plan.name, func(t *testing.T) {
			pieces := unitRanges(plan.total, plan.width)
			if len(pieces) != plan.units {
				t.Fatalf("got %d units, want %d", len(pieces), plan.units)
			}
			next := 0
			for _, piece := range pieces {
				if piece.first != next || piece.last <= piece.first || piece.last > plan.total {
					t.Fatalf("invalid coverage at %+v after %d", piece, next)
				}
				next = piece.last
			}
			if next != plan.total {
				t.Fatalf("covered %d pieces, want %d", next, plan.total)
			}
		})
	}
	// Verify the actual top-level wrappers, including their helper and range index.
	// This checks gate discovery and catches a missing, duplicated or wrong wrapper.
	for _, plan := range []struct {
		file, prefix, helper, target string
		units                        int
	}{
		{"decode_ascii_test.go", "TestDecodeASCIIUnit", "runDecodeASCIIUnit", "native", 45},
		{"decode_ascii_test.go", "TestDecodeASCIIWASIUnit", "runDecodeASCIIUnit", "wasi", 45},
		{"normalize_test.go", "TestNormalizeMatchesNodePoints", "runNormalizeUnit", "", 17},
	} {
		file, err := parser.ParseFile(token.NewFileSet(), plan.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, plan.units)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, plan.prefix) {
				continue
			}
			index, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, plan.prefix))
			if err != nil || index < 0 || index >= plan.units || function.Name.Name != fmt.Sprintf("%s%02d", plan.prefix, index) {
				t.Fatalf("unexpected unit %s", function.Name.Name)
			}
			if seen[index] || len(function.Body.List) != 2 {
				t.Fatalf("invalid wrapper %s", function.Name.Name)
			}
			seen[index] = true
			parallel, ok := function.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("missing first-statement t.Parallel in %s", function.Name.Name)
			}
			parallelCall, ok := parallel.X.(*ast.CallExpr)
			if !ok {
				t.Fatalf("missing t.Parallel in %s", function.Name.Name)
			}
			selector, ok := parallelCall.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Parallel" || len(parallelCall.Args) != 0 {
				t.Fatalf("missing t.Parallel in %s", function.Name.Name)
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "t" {
				t.Fatalf("wrong parallel receiver in %s", function.Name.Name)
			}
			statement, ok := function.Body.List[1].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("invalid wrapper %s", function.Name.Name)
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok {
				t.Fatalf("invalid call %s", function.Name.Name)
			}
			helper, ok := call.Fun.(*ast.Ident)
			if !ok || helper.Name != plan.helper {
				t.Fatalf("wrong helper in %s", function.Name.Name)
			}
			wantArguments := 2
			if plan.target != "" {
				wantArguments = 3
			}
			if len(call.Args) != wantArguments {
				t.Fatalf("wrong arguments in %s", function.Name.Name)
			}
			actual, ok := call.Args[len(call.Args)-1].(*ast.BasicLit)
			if !ok || actual.Value != strconv.Itoa(index) {
				t.Fatalf("wrong range in %s", function.Name.Name)
			}
			if plan.target != "" {
				target, ok := call.Args[1].(*ast.BasicLit)
				if !ok || target.Value != strconv.Quote(plan.target) {
					t.Fatalf("wrong target in %s", function.Name.Name)
				}
			}
		}
		for index, present := range seen {
			if !present {
				t.Errorf("missing top-level %s unit %d", plan.prefix, index)
			}
		}
	}
}

// eachInParallel runs job for every index below count on four workers, the CPUs a product's unit is given, and
// returns the first failure in index order. A product's build compiles its translation units with it.
func eachInParallel(count int, job func(index int) error) error {
	failures := make([]error, count)
	jobs := make(chan int)
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				failures[index] = job(index)
			}
		}()
	}
	for index := 0; index < count; index++ {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	for _, err := range failures {
		if err != nil {
			return err
		}
	}
	return nil
}

// Unit timing is diagnostic unless the reference audit explicitly enables it.
func checkGrainBudget(t *testing.T) func() {
	t.Helper()
	started := time.Now()
	return func() {
		elapsed := time.Since(started)
		t.Logf("unit elapsed %.3fs", elapsed.Seconds())
		if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" && elapsed > 60*time.Second {
			t.Errorf("unit over 60s budget: %s", elapsed)
		}
	}
}

const randomRegexCases = 10000
const randomRegexUnitSize = 250

var splitCacheModes = []string{"0", "1"}
var decodeTargets = []string{"native", "wasi"}

type testShard struct{ index, count int }

func parseTestShard(value string) (testShard, error) {
	if value == "" {
		return testShard{0, 1}, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return testShard{}, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n with 0 <= i < n: %q", value)
	}
	index, firstError := strconv.Atoi(parts[0])
	count, secondError := strconv.Atoi(parts[1])
	if firstError != nil || secondError != nil || count < 1 || index < 0 || index >= count {
		return testShard{}, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n with 0 <= i < n: %q", value)
	}
	return testShard{index, count}, nil
}
func currentTestShard(t *testing.T) testShard {
	t.Helper()
	shard, err := parseTestShard(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	return shard
}
func (s testShard) owns(piece int) bool { return piece%s.count == s.index }

func TestShardSelection(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"-1/2", "0/0", "2/2", "one/2", "0/2/3", "/2"} {
		if _, err := parseTestShard(value); err == nil {
			t.Errorf("invalid shard accepted: %q", value)
		}
	}
	if shard, err := parseTestShard(""); err != nil || shard != (testShard{0, 1}) {
		t.Fatalf("default shard: %+v %v", shard, err)
	}
}

func runRegexPartitions(t *testing.T, cases []regexCase) {
	t.Helper()
	if len(cases) != randomRegexCases {
		t.Fatalf("random regex corpus: got %d, want %d", len(cases), randomRegexCases)
	}
	shard := currentTestShard(t)
	for index, piece := range unitRanges(len(cases), randomRegexUnitSize) {
		if !shard.owns(index) {
			continue
		}
		t.Run(piece.name(), func(t *testing.T) { runRegexCases(t, cases[piece.first:piece.last]) })
	}
}

func TestRetainedSplitCoverage(t *testing.T) {
	t.Parallel()
	if randomRegexCases != 10000 || randomRegexUnitSize != 250 || len(unitRanges(randomRegexCases, randomRegexUnitSize)) != 40 {
		t.Fatal("random regex units must partition all 10000 cases into 40 units of 250")
	}
	t.Run("shard_union", func(t *testing.T) {
		for _, pieces := range []int{90, 40, 18, 39, 2, 6} {
			for count := 1; count <= 100; count++ {
				seen := make([]int, pieces)
				for index := 0; index < count; index++ {
					shard, err := parseTestShard(fmt.Sprintf("%d/%d", index, count))
					if err != nil {
						t.Fatal(err)
					}
					for piece := 0; piece < pieces; piece++ {
						if shard.owns(piece) {
							seen[piece]++
						}
					}
				}
				for piece, owners := range seen {
					if owners != 1 {
						t.Fatalf("piece %d of %d has %d owners across %d shards", piece, pieces, owners, count)
					}
				}
			}
		}
	})
	t.Run("decode_targets", func(t *testing.T) {
		if strings.Join(decodeTargets, ",") != "native,wasi" {
			t.Fatal("native and WASI decoder targets must both run")
		}
	})
	t.Run("WASI", func(t *testing.T) {
		if len(wasiFixtures) != 38 {
			t.Fatalf("got %d WASI fixtures, want 38", len(wasiFixtures))
		}
		want := "efadf45774749479cdcaf1073bbd3dca6964f56681d12da0fa702141cd022190"
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(wasiFixtures, "\n")))); got != want {
			t.Fatalf("WASI fixture membership changed: %s", got)
		}
	})
	t.Run("split_cache", func(t *testing.T) {
		if strings.Join(splitCacheModes, ",") != "0,1" {
			t.Fatal("cached and bypass builds must both run")
		}
	})
	t.Run("record_mutants", func(t *testing.T) {
		if len(recordMutationCases) != 6 {
			t.Fatalf("got %d record mutants, want 6", len(recordMutationCases))
		}
		seen := map[string]bool{}
		for _, mutant := range recordMutationCases {
			if seen[mutant.name] {
				t.Fatalf("duplicate record mutant %s", mutant.name)
			}
			seen[mutant.name] = true
		}
	})
}

func TestRetainedTopLevelCoverage(t *testing.T) {
	t.Parallel()
	for _, plan := range []struct {
		file, prefix, helper string
		count                int
	}{
		{"regexp_test.go", "TestRegExpBytecodeRandomNodeShard", "runRegExpBytecodeRandomNodeShard", 40},
		{"record_test.go", "TestRecordMutantsUnit", "runRecordMutantUnit", 6},
		{"units_tsgo_test.go", "TestSplitTSGoAgreesUnit", "runSplitTSGoAgreesUnit", 2},
		{"wasm_test.go", "TestWASIUnit", "runWASIUnit", 39},
	} {
		file, err := parser.ParseFile(token.NewFileSet(), plan.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, plan.count)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, plan.prefix) {
				continue
			}
			index, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, plan.prefix))
			if err != nil || index < 0 || index >= plan.count || seen[index] {
				t.Fatalf("invalid unit %s", function.Name.Name)
			}
			seen[index] = true
			statement := function.Body.List[len(function.Body.List)-1].(*ast.ExprStmt)
			call := statement.X.(*ast.CallExpr)
			helper, ok := call.Fun.(*ast.Ident)
			if !ok || helper.Name != plan.helper || len(call.Args) != 2 {
				t.Fatalf("wrong helper in %s", function.Name.Name)
			}
			unit, ok := call.Args[1].(*ast.BasicLit)
			if !ok || unit.Value != strconv.Itoa(index) {
				t.Fatalf("wrong unit in %s", function.Name.Name)
			}
		}
		for index, present := range seen {
			if !present {
				t.Errorf("missing %s%02d", plan.prefix, index)
			}
		}
	}
}

// checkRangesTile holds a table of ranges, each its own top-level test, to cases 0 to total: every
// case in exactly one range, none left out, none swept twice, none past the end. A row is named by
// its bounds, first and last both included, in format ("%X to %X"), so a name can't drift from what
// its test sweeps.
func checkRangesTile(t *testing.T, format string, total int, rows map[string]testUnitRange) {
	t.Helper()
	pieces := make([]testUnitRange, 0, len(rows))
	for name, piece := range rows {
		if piece.last <= piece.first {
			t.Errorf("range %q is empty: %d to %d", name, piece.first, piece.last)
		}
		if want := fmt.Sprintf(format, piece.first, piece.last-1); name != want {
			t.Errorf("range %q sweeps %s", name, want)
		}
		pieces = append(pieces, piece)
	}
	sort.Slice(pieces, func(left, right int) bool { return pieces[left].first < pieces[right].first })
	next := 0
	for _, piece := range pieces {
		if piece.first != next {
			t.Errorf("range %#x to %#x leaves a gap or overlaps: the ranges before it end at %#x", piece.first, piece.last, next)
		}
		next = max(next, piece.last)
	}
	if next != total {
		t.Errorf("ranges end at %#x, want %#x", next, total)
	}
}

// checkNamedWrappers holds a table whose rows are each their own top-level test to those tests. The
// row's name, not its place, picks its test: "BMP low halves" is prefix+"BMPLowHalves", whose body
// is t.Parallel() and then helper(t, "BMP low halves"). Every function in file whose name starts
// with prefix must be one of those, so a row without its test, a test for a row that isn't there,
// or one that runs another row's case fails here, before anything is built.
func checkNamedWrappers[Row any](t *testing.T, file, prefix, helper string, rows map[string]Row) {
	t.Helper()
	if len(rows) == 0 {
		t.Fatalf("%s has no rows", prefix)
	}
	want := make(map[string]string, len(rows))
	for row := range rows {
		name := prefix
		for _, word := range strings.Fields(row) {
			name += strings.ToUpper(word[:1]) + word[1:]
		}
		if earlier, ok := want[name]; ok {
			t.Fatalf("rows %q and %q both name %s", earlier, row, name)
		}
		want[name] = row
	}
	tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := make(map[string]bool, len(rows))
	for _, declaration := range tree.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		row, ok := want[function.Name.Name]
		if !ok {
			t.Errorf("%s names no row of its table", function.Name.Name)
			continue
		}
		found[function.Name.Name] = true
		if len(function.Body.List) != 2 {
			t.Errorf("%s must be t.Parallel() then %s(t, %q)", function.Name.Name, helper, row)
			continue
		}
		parallel, ok := function.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Errorf("%s must call t.Parallel() first", function.Name.Name)
			continue
		}
		parallelCall, ok := parallel.X.(*ast.CallExpr)
		if !ok {
			t.Errorf("%s must call t.Parallel() first", function.Name.Name)
			continue
		}
		selector, ok := parallelCall.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Parallel" || len(parallelCall.Args) != 0 {
			t.Errorf("%s must call t.Parallel() first", function.Name.Name)
			continue
		}
		if receiver, ok := selector.X.(*ast.Ident); !ok || receiver.Name != "t" {
			t.Errorf("%s must call t.Parallel() first", function.Name.Name)
			continue
		}
		statement, ok := function.Body.List[1].(*ast.ExprStmt)
		if !ok {
			t.Errorf("%s must then call %s(t, %q)", function.Name.Name, helper, row)
			continue
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Errorf("%s must then call %s(t, %q)", function.Name.Name, helper, row)
			continue
		}
		called, isHelper := call.Fun.(*ast.Ident)
		argument, isT := call.Args[0].(*ast.Ident)
		label, isLiteral := call.Args[1].(*ast.BasicLit)
		named := ""
		if isLiteral && label.Kind == token.STRING {
			named, _ = strconv.Unquote(label.Value)
		}
		if !isHelper || called.Name != helper || !isT || argument.Name != "t" || named != row {
			t.Errorf("%s must then call %s(t, %q)", function.Name.Name, helper, row)
		}
	}
	for name, row := range want {
		if !found[name] {
			t.Errorf("row %q has no top-level test %s", row, name)
		}
	}
}
