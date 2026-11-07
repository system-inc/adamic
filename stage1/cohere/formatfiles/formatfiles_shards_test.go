package formatfiles

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/formatfiles/oracles"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// ADAMIC_TEST_SHARD=i/n selects one shard; unset runs all. The count stays fixed.
// Trees come from cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359 and the
// seeded generator. Hash tree names and mutant names, never enumeration positions.
const testThePortParsesAsGoCohereDoesShards = 32

type formatfilesPart struct {
	cases, answers string
	ids            []string
}

func formatfilesShard(key string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum64() % testThePortParsesAsGoCohereDoesShards)
}

func formatfilesSelectedShard(test *testing.T) int {
	test.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return -1
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	index, err := strconv.Atoi(fields[0])
	count, countErr := strconv.Atoi(fields[1])
	if err != nil || countErr != nil || count != testThePortParsesAsGoCohereDoesShards || index < 0 || index >= count {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/%d", value, testThePortParsesAsGoCohereDoesShards)
	}
	return index
}

func formatfilesBlocks(test *testing.T, text, marker string) []string {
	test.Helper()
	chunks := strings.Split(text, "\n"+marker)
	if !strings.HasPrefix(chunks[0], marker) {
		test.Fatalf("missing %q corpus", marker)
	}
	for index := 1; index < len(chunks); index++ {
		chunks[index] = marker + chunks[index]
	}
	for index := 0; index < len(chunks)-1; index++ {
		chunks[index] += "\n"
	}
	return chunks
}

func formatfilesPartition(test *testing.T, path, answers string) []formatfilesPart {
	test.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	header, body, ok := strings.Cut(string(contents), "\n")
	if !ok {
		test.Fatal("missing handles header")
	}
	cases := formatfilesBlocks(test, body, "tree\t")
	oracle := formatfilesBlocks(test, answers, "tree ")
	if len(cases) != len(oracle) || len(cases) == 0 {
		test.Fatal("case/answer corpus differs or empty")
	}
	parts := make([]formatfilesPart, testThePortParsesAsGoCohereDoesShards)
	for index := range parts {
		parts[index].cases = header + "\n"
	}
	expected := make(map[string]bool)
	for index, block := range cases {
		fields := strings.SplitN(block, "\t", 3)
		if len(fields) != 3 {
			test.Fatal("invalid tree record")
		}
		key := fields[1]
		if expected[key] {
			test.Fatalf("repeated tree %q", key)
		}
		expected[key] = true
		// The current tree names have no escaped control characters.
		name, _, _ := strings.Cut(oracle[index], "\n")
		if name != "tree "+key {
			test.Fatalf("tree/answer identity differs: %q / %q", key, name)
		}
		shard := formatfilesShard("tree/" + key)
		parts[shard].cases += block
		parts[shard].answers += oracle[index]
		parts[shard].ids = append(parts[shard].ids, key)
	}
	seen := make(map[string]bool)
	for _, part := range parts {
		for _, key := range part.ids {
			if !expected[key] || seen[key] {
				test.Fatalf("unexpected or repeated sharded tree %q", key)
			}
			seen[key] = true
		}
	}
	if len(seen) != len(expected) {
		test.Fatalf("union %d, live corpus %d", len(seen), len(expected))
	}
	// Every mutant retains the entire live corpus. Each mutant belongs to exactly
	// one shard by name; count its tree comparisons as distinct mode/case IDs.
	modes := map[string]bool{"baseline": true}
	for _, mutant := range mutants {
		key := "mutant/" + mutant.name
		if modes[key] {
			test.Fatalf("repeated mode %q", key)
		}
		modes[key] = true
	}

	expectedCases := make(map[string]bool)
	assignedCases := make([][]string, len(parts))
	for mode := range modes {
		for key := range expected {
			expectedCases[mode+"/"+key] = true
		}
	}
	for shard, part := range parts {
		for _, key := range part.ids {
			assignedCases[shard] = append(assignedCases[shard], "baseline/"+key)
		}
	}
	for _, mutant := range mutants {
		mode := "mutant/" + mutant.name
		shard := formatfilesShard(mode)
		for key := range expected {
			assignedCases[shard] = append(assignedCases[shard], mode+"/"+key)
		}
	}
	actualCases := make(map[string]bool)
	for _, cases := range assignedCases {
		for _, id := range cases {
			if !expectedCases[id] || actualCases[id] {
				test.Fatalf("unexpected or repeated mode/tree case %q", id)
			}
			actualCases[id] = true
		}
	}
	if len(actualCases) != len(expectedCases) {
		test.Fatalf("union %d, live enumeration %d", len(actualCases), len(expectedCases))
	}
	test.Logf(
		"union: %d trees, %d modes, %d unique mode/tree cases, %d shards",
		len(seen),
		len(modes),
		len(actualCases),
		len(parts),
	)
	return parts
}

func formatfilesCaseFile(test *testing.T, part formatfilesPart) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(part.cases), 0644); err != nil {
		test.Fatal(err)
	}
	return path
}

func formatfilesBinary(test *testing.T, program *ir.Program, mutant *mutant) string {
	return formatfilesBinaryWithOptions(test, program, mutant, true)
}

func formatfilesBinaryWithOptions(test *testing.T, program *ir.Program, mutant *mutant, sanitize bool) string {
	test.Helper()
	source := native.C(program)
	directory := buildcache.Product(test, formatfilesNativeInputs(
		source,
		mutant,
		sanitize,
	), func(directory string) error {
		return formatfilesBuildNative(test, source, filepath.Join(directory, "port"), sanitize)
	})
	return filepath.Join(directory, "port")
}

// formatfilesNativeInputs keys the native port built from source: the compiler that lowered and builds it
// (compilerinputs.Lowering), the port's and its gitignore slice's .ts files, and the init in
// formatfiles_prepare_test.go that runs native.Build for it.
func formatfilesNativeInputs(source string, mutant *mutant, sanitize bool) buildcache.Inputs {
	name := "formatfiles-native"
	if !sanitize {
		name += "-unsanitized"
	}
	flags := append([]string{}, native.Flags(native.Options{Sanitize: sanitize})...)
	if mutant != nil {
		name += "-" + mutant.name
		flags = append(flags, mutant.file, mutant.from, mutant.to)
	}
	flags = append(
		flags,
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		fmt.Sprintf("C-sha256=%x", sha256.Sum256([]byte(source))),
	)
	return buildcache.Inputs{
		Name: name,
		Files: append(
			compilerinputs.Lowering(),
			"stage1/cohere/formatfiles/golang.ts",
			"stage1/cohere/formatfiles/disk.ts",
			"stage1/cohere/formatfiles/enumerate.ts",
			"stage1/cohere/formatfiles/main.ts",
			"stage1/cohere/gitignore/path.ts",
			"stage1/cohere/gitignore/glob.ts",
			"stage1/cohere/gitignore/gitignore.ts",
			"stage1/cohere/formatfiles/formatfiles_prepare_test.go",
		),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}
}

// An edit to a package the compiler links must move every native product's key, and an edit to tsc's baselines, which
// nothing links, must keep it.
func TestFormatfilesCompilerProductsKeyWhatTheCompilerLinks(test *testing.T) {
	test.Parallel()
	products := []buildcache.Inputs{formatfilesNativeInputs("", nil, true), formatfilesNativeInputs("", nil, false)}
	for index := range mutants {
		products = append(products, formatfilesNativeInputs("", &mutants[index], true))
	}
	covers := func(entry, path string) bool { return entry == path || strings.HasPrefix(path, entry+"/") }
	for _, product := range products {
		linked, baselines := false, false
		for _, entry := range product.Files {
			linked = linked || covers(entry, "cohere/TypeScript/tsc/internal/checker")
			baselines = baselines || covers(entry, "cohere/TypeScript/tsc/testdata")
		}
		if !linked {
			test.Errorf(
				"%s: no Files entry covers cohere/TypeScript/tsc/internal/checker, a package the compiler links",
				product.Name,
			)
		}
		if baselines {
			test.Errorf(
				"%s: a Files entry covers cohere/TypeScript/tsc/testdata, baselines nothing links",
				product.Name,
			)
		}
	}
}

func TestFormatfilesShardPlantedDisagreement(test *testing.T) {
	test.Parallel()
	// Route real case/answer blocks through the same partitioner as the oracle
	// comparison, then plant exactly one disagreement in the routed answers.
	path := filepath.Join(test.TempDir(), "cases.txt")
	var cases, answers strings.Builder
	cases.WriteString("handles\t.ts\n")
	for _, key := range []string{"generated 0", "generated 1", "generated 2"} {
		fmt.Fprintf(&cases, "tree\t%s\t/fixture\nhouse\t0\n", key)
		fmt.Fprintf(&answers, "tree %s\nsame\n", key)
	}
	if err := os.WriteFile(path, []byte(cases.String()), 0644); err != nil {
		test.Fatal(err)
	}
	parts := formatfilesPartition(test, path, answers.String())
	planted := "generated 1"
	caught := []int{}
	for shard, part := range parts {
		got := strings.Replace(part.answers, "tree "+planted+"\nsame\n", "tree "+planted+"\nplanted disagreement\n", 1)
		if firstDifference(got, part.answers) != "" {
			caught = append(caught, shard)
		}
	}
	owner := formatfilesShard("tree/" + planted)
	if len(caught) != 1 || caught[0] != owner {
		test.Fatalf("caught by %v, want only %d", caught, owner)
	}
	test.Logf("planted disagreement %q caught only by shard-%03d", planted, owner)
}

// The gate discovers these top-level leaves with go test -list. Each one runs
// on its own instance, and ADAMIC_TEST_SHARD=i/32 can select the same unit.
func TestThePortParsesAsGoCohereDoes_000(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 0)
}
func TestThePortParsesAsGoCohereDoes_001(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 1)
}
func TestThePortParsesAsGoCohereDoes_002(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 2)
}
func TestThePortParsesAsGoCohereDoes_003(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 3)
}
func TestThePortParsesAsGoCohereDoes_004(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 4)
}
func TestThePortParsesAsGoCohereDoes_005(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 5)
}
func TestThePortParsesAsGoCohereDoes_006(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 6)
}
func TestThePortParsesAsGoCohereDoes_007(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 7)
}
func TestThePortParsesAsGoCohereDoes_008(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 8)
}
func TestThePortParsesAsGoCohereDoes_009(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 9)
}
func TestThePortParsesAsGoCohereDoes_010(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 10)
}
func TestThePortParsesAsGoCohereDoes_011(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 11)
}
func TestThePortParsesAsGoCohereDoes_012(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 12)
}
func TestThePortParsesAsGoCohereDoes_013(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 13)
}
func TestThePortParsesAsGoCohereDoes_014(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 14)
}
func TestThePortParsesAsGoCohereDoes_015(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 15)
}
func TestThePortParsesAsGoCohereDoes_016(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 16)
}
func TestThePortParsesAsGoCohereDoes_017(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 17)
}
func TestThePortParsesAsGoCohereDoes_018(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 18)
}
func TestThePortParsesAsGoCohereDoes_019(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 19)
}
func TestThePortParsesAsGoCohereDoes_020(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 20)
}
func TestThePortParsesAsGoCohereDoes_021(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 21)
}
func TestThePortParsesAsGoCohereDoes_022(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 22)
}
func TestThePortParsesAsGoCohereDoes_023(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 23)
}
func TestThePortParsesAsGoCohereDoes_024(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 24)
}
func TestThePortParsesAsGoCohereDoes_025(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 25)
}
func TestThePortParsesAsGoCohereDoes_026(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 26)
}
func TestThePortParsesAsGoCohereDoes_027(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 27)
}
func TestThePortParsesAsGoCohereDoes_028(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 28)
}
func TestThePortParsesAsGoCohereDoes_029(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 29)
}
func TestThePortParsesAsGoCohereDoes_030(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 30)
}
func TestThePortParsesAsGoCohereDoes_031(test *testing.T) {
	test.Parallel()
	formatfilesCheckShard(test, 31)
}

func TestThePortParsesAsGoCohereDoesUnion(test *testing.T) {
	test.Parallel()
	shared := formatfilesReady(test)
	parts := formatfilesPartition(test, shared.casesPath, shared.answers)
	if len(parts) != testThePortParsesAsGoCohereDoesShards {
		test.Fatal("enumerated shard count differs")
	}
	// Check the written-out top-level table too: missing or repeated leaves are
	// a failure even when their data partitions themselves are complete.
	source, err := os.ReadFile("formatfiles_shards_test.go")
	if err != nil {
		test.Fatal(err)
	}

	table, err := parser.ParseFile(token.NewFileSet(), "formatfiles_shards_test.go", source, 0)
	if err != nil {
		test.Fatal(err)
	}
	names := make(map[int]bool)
	for _, declaration := range table.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "TestThePortParsesAsGoCohereDoes_Setup" {
			continue
		}
		suffix, ok := strings.CutPrefix(fn.Name.Name, "TestThePortParsesAsGoCohereDoes_")
		if !ok {
			continue
		}
		index, err := strconv.Atoi(suffix)
		if err != nil || index < 0 || index >= len(parts) || names[index] {
			test.Fatalf("invalid leaf declaration %q", fn.Name.Name)
		}
		if len(fn.Body.List) != 2 {
			test.Fatalf("unexpected leaf body %s", fn.Name.Name)
		}
		statement, ok := fn.Body.List[1].(*ast.ExprStmt)
		if !ok {
			test.Fatal("missing shard call")
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			test.Fatal("invalid shard call")
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "formatfilesCheckShard" {
			test.Fatal("wrong shard runner")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Value != strconv.Itoa(index) {
			test.Fatalf("leaf %d calls the wrong shard", index)
		}
		names[index] = true
	}

	if len(names) != len(parts) {
		test.Fatalf("%d leaves, %d partitions", len(names), len(parts))
	}
}

func formatfilesAskedCases(test *testing.T) (string, string) {
	test.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_FORMATFILES_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			test.Fatal(err)
		}
	}
	generated := 400
	if value := os.Getenv("COHERE_FORMATFILES_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			test.Fatal(err)
		}
	}
	directory := formatfilesSetupDirectory(test)
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	// The trees live here, where the port walks them after cohere's side has laid them out; the
	// directory is the test's own, so it outlives cohere's side and goes when the test ends.
	scratch := filepath.Join(directory, "trees")
	if err := os.Mkdir(scratch, 0o755); err != nil {
		test.Fatal(err)
	}
	formatfilesCohereSide(test, map[string]any{
		"scratch":   scratch,
		"seed":      seed,
		"generated": generated,
		"cases":     casesPath,
		"answers":   answersPath,
	})
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		test.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_FORMATFILES_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			test.Fatal(err)
		}
	}
	test.Logf("seed %d, %d generated trees", seed, generated)
	return casesPath, string(answers)
}

func formatfilesCohereSide(test *testing.T, request map[string]any) {
	test.Helper()
	product, packageDirectory := formatfilesOracle(test)
	directory := test.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, err := json.Marshal(request)
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(requestPath, encoded, 0o644); err != nil {
		test.Fatal(err)
	}
	command := formatfilesCommand(test, product, "-test.run=^TestAdamicPortCases$", "-test.timeout=0")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	// Read through childguard.Capture for formatfilesExecute's reason: a WaitDelay could fail a run that had finished.
	// Capture has no stall guard, and this run loses nothing by it: without -test.v cohere's side prints nothing
	// until it ends, so only childguard's 30-minute first-output window could have fired, and the unit's own ceiling
	// (Loom kills a test unit at 90 s and a product unit at 600 s) ends a hung run sooner.
	if stdout, stderr, err := childguard.Capture(test.Context(), command); err != nil {
		test.Fatalf("cohere's side: %v\n%s%s", err, stdout, stderr)
	}
}

// formatfilesOracle is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's
// formatfiles package, a product (buildcache.GoTest) both build-phase units and shard preparation share, and the
// directory go test would run it in.
func formatfilesOracle(test *testing.T) (string, string) {
	test.Helper()
	return oracles.GoOracle(test)
}

// Setup children carry no deadline; the gate bounds the whole unit.
func formatfilesCommand(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	return formatfilesContextCommand(test.Context(), name, arguments...)
}
