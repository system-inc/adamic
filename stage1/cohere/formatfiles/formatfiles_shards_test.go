package formatfiles

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
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

func formatfilesSelectedShard(t *testing.T) int {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return -1
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	index, err := strconv.Atoi(fields[0])
	count, countErr := strconv.Atoi(fields[1])
	if err != nil || countErr != nil || count != testThePortParsesAsGoCohereDoesShards || index < 0 || index >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/%d", value, testThePortParsesAsGoCohereDoesShards)
	}
	return index
}

func formatfilesBlocks(t *testing.T, text, marker string) []string {
	t.Helper()
	chunks := strings.Split(text, "\n"+marker)
	if !strings.HasPrefix(chunks[0], marker) {
		t.Fatalf("missing %q corpus", marker)
	}
	for index := 1; index < len(chunks); index++ {
		chunks[index] = marker + chunks[index]
	}
	for index := 0; index < len(chunks)-1; index++ {
		chunks[index] += "\n"
	}
	return chunks
}

func formatfilesPartition(t *testing.T, path, answers string) []formatfilesPart {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	header, body, ok := strings.Cut(string(contents), "\n")
	if !ok {
		t.Fatal("missing handles header")
	}
	cases := formatfilesBlocks(t, body, "tree\t")
	oracle := formatfilesBlocks(t, answers, "tree ")
	if len(cases) != len(oracle) || len(cases) == 0 {
		t.Fatal("case/answer corpus differs or empty")
	}
	parts := make([]formatfilesPart, testThePortParsesAsGoCohereDoesShards)
	for index := range parts {
		parts[index].cases = header + "\n"
	}
	expected := make(map[string]bool)
	for index, block := range cases {
		fields := strings.SplitN(block, "\t", 3)
		if len(fields) != 3 {
			t.Fatal("invalid tree record")
		}
		key := fields[1]
		if expected[key] {
			t.Fatalf("repeated tree %q", key)
		}
		expected[key] = true
		// The current tree names have no escaped control characters.
		name, _, _ := strings.Cut(oracle[index], "\n")
		if name != "tree "+key {
			t.Fatalf("tree/answer identity differs: %q / %q", key, name)
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
				t.Fatalf("unexpected or repeated sharded tree %q", key)
			}
			seen[key] = true
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("union %d, live corpus %d", len(seen), len(expected))
	}
	// Every mutant retains the entire live corpus. Each mutant belongs to exactly
	// one shard by name; count its tree comparisons as distinct mode/case IDs.
	modes := map[string]bool{"baseline": true}
	for _, mutant := range mutants {
		key := "mutant/" + mutant.name
		if modes[key] {
			t.Fatalf("repeated mode %q", key)
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
				t.Fatalf("unexpected or repeated mode/tree case %q", id)
			}
			actualCases[id] = true
		}
	}
	if len(actualCases) != len(expectedCases) {
		t.Fatalf("union %d, live enumeration %d", len(actualCases), len(expectedCases))
	}
	t.Logf("union: %d trees, %d modes, %d unique mode/tree cases, %d shards", len(seen), len(modes), len(actualCases), len(parts))
	return parts
}

func formatfilesCaseFile(t *testing.T, part formatfilesPart) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(part.cases), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func formatfilesBinary(t *testing.T, program *ir.Program, mutant *mutant) string {
	return formatfilesBinaryWithOptions(t, program, mutant, true)
}

func formatfilesBinaryWithOptions(t *testing.T, program *ir.Program, mutant *mutant, sanitize bool) string {
	t.Helper()
	source := native.C(program)
	name := "formatfiles-native"
	if !sanitize {
		name += "-unsanitized"
	}
	flags := append([]string{}, native.Flags(native.Options{Sanitize: sanitize})...)
	if mutant != nil {
		name += "-" + mutant.name
		flags = append(flags, mutant.file, mutant.from, mutant.to)
	}
	flags = append(flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), fmt.Sprintf("C-sha256=%x", sha256.Sum256([]byte(source))))
	inputs := buildcache.Inputs{Name: name,
		Files: []string{"stage1/cohere/formatfiles/golang.ts", "stage1/cohere/formatfiles/disk.ts", "stage1/cohere/formatfiles/enumerate.ts", "stage1/cohere/formatfiles/main.ts", "stage1/cohere/gitignore/path.ts", "stage1/cohere/gitignore/glob.ts", "stage1/cohere/gitignore/gitignore.ts", "internal", "go.mod", "cohere/TypeScript/tsc", "cohere/TypeScript-shim"},
		Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		return formatfilesBuildNative(t, source, filepath.Join(directory, "port"), sanitize)
	})
	return filepath.Join(directory, "port")
}

func TestFormatfilesShardPlantedDisagreement(t *testing.T) {
	t.Parallel()
	// Route real case/answer blocks through the same partitioner as the oracle
	// comparison, then plant exactly one disagreement in the routed answers.
	path := filepath.Join(t.TempDir(), "cases.txt")
	var cases, answers strings.Builder
	cases.WriteString("handles\t.ts\n")
	for _, key := range []string{"generated 0", "generated 1", "generated 2"} {
		fmt.Fprintf(&cases, "tree\t%s\t/fixture\nhouse\t0\n", key)
		fmt.Fprintf(&answers, "tree %s\nsame\n", key)
	}
	if err := os.WriteFile(path, []byte(cases.String()), 0644); err != nil {
		t.Fatal(err)
	}
	parts := formatfilesPartition(t, path, answers.String())
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
		t.Fatalf("caught by %v, want only %d", caught, owner)
	}
	t.Logf("planted disagreement %q caught only by shard-%03d", planted, owner)
}

// The gate discovers these top-level leaves with go test -list. Each one runs
// on its own instance, and ADAMIC_TEST_SHARD=i/32 can select the same unit.
func TestThePortParsesAsGoCohereDoes_000(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 0)
}
func TestThePortParsesAsGoCohereDoes_001(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 1)
}
func TestThePortParsesAsGoCohereDoes_002(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 2)
}
func TestThePortParsesAsGoCohereDoes_003(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 3)
}
func TestThePortParsesAsGoCohereDoes_004(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 4)
}
func TestThePortParsesAsGoCohereDoes_005(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 5)
}
func TestThePortParsesAsGoCohereDoes_006(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 6)
}
func TestThePortParsesAsGoCohereDoes_007(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 7)
}
func TestThePortParsesAsGoCohereDoes_008(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 8)
}
func TestThePortParsesAsGoCohereDoes_009(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 9)
}
func TestThePortParsesAsGoCohereDoes_010(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 10)
}
func TestThePortParsesAsGoCohereDoes_011(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 11)
}
func TestThePortParsesAsGoCohereDoes_012(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 12)
}
func TestThePortParsesAsGoCohereDoes_013(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 13)
}
func TestThePortParsesAsGoCohereDoes_014(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 14)
}
func TestThePortParsesAsGoCohereDoes_015(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 15)
}
func TestThePortParsesAsGoCohereDoes_016(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 16)
}
func TestThePortParsesAsGoCohereDoes_017(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 17)
}
func TestThePortParsesAsGoCohereDoes_018(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 18)
}
func TestThePortParsesAsGoCohereDoes_019(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 19)
}
func TestThePortParsesAsGoCohereDoes_020(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 20)
}
func TestThePortParsesAsGoCohereDoes_021(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 21)
}
func TestThePortParsesAsGoCohereDoes_022(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 22)
}
func TestThePortParsesAsGoCohereDoes_023(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 23)
}
func TestThePortParsesAsGoCohereDoes_024(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 24)
}
func TestThePortParsesAsGoCohereDoes_025(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 25)
}
func TestThePortParsesAsGoCohereDoes_026(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 26)
}
func TestThePortParsesAsGoCohereDoes_027(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 27)
}
func TestThePortParsesAsGoCohereDoes_028(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 28)
}
func TestThePortParsesAsGoCohereDoes_029(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 29)
}
func TestThePortParsesAsGoCohereDoes_030(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 30)
}
func TestThePortParsesAsGoCohereDoes_031(t *testing.T) {
	t.Parallel()
	formatfilesCheckShard(t, 31)
}

func TestThePortParsesAsGoCohereDoesUnion(t *testing.T) {
	t.Parallel()
	shared := formatfilesReady(t)
	parts := formatfilesPartition(t, shared.casesPath, shared.answers)
	if len(parts) != testThePortParsesAsGoCohereDoesShards {
		t.Fatal("enumerated shard count differs")
	}
	// Check the written-out top-level table too: missing or repeated leaves are
	// a failure even when their data partitions themselves are complete.
	source, err := os.ReadFile("formatfiles_shards_test.go")
	if err != nil {
		t.Fatal(err)
	}

	table, err := parser.ParseFile(token.NewFileSet(), "formatfiles_shards_test.go", source, 0)
	if err != nil {
		t.Fatal(err)
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
			t.Fatalf("invalid leaf declaration %q", fn.Name.Name)
		}
		if len(fn.Body.List) != 2 {
			t.Fatalf("unexpected leaf body %s", fn.Name.Name)
		}
		statement, ok := fn.Body.List[1].(*ast.ExprStmt)
		if !ok {
			t.Fatal("missing shard call")
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Fatal("invalid shard call")
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "formatfilesCheckShard" {
			t.Fatal("wrong shard runner")
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Value != strconv.Itoa(index) {
			t.Fatalf("leaf %d calls the wrong shard", index)
		}
		names[index] = true
	}

	if len(names) != len(parts) {
		t.Fatalf("%d leaves, %d partitions", len(names), len(parts))
	}
}

func formatfilesAskedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_FORMATFILES_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 400
	if value := os.Getenv("COHERE_FORMATFILES_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	directory := formatfilesSetupDirectory(t)
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	// The trees live here, where the port walks them after cohere's side has laid them out; the
	// directory is the test's own, so it outlives cohere's side and goes when the test ends.
	scratch := filepath.Join(directory, "trees")
	if err := os.Mkdir(scratch, 0o755); err != nil {
		t.Fatal(err)
	}
	formatfilesCohereSide(t, map[string]any{"scratch": scratch, "seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath})
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_FORMATFILES_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated trees", seed, generated)
	return casesPath, string(answers)
}

func formatfilesCohereSide(t *testing.T, request map[string]any) {
	t.Helper()
	directory := t.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "formatfiles")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}

	inputs := buildcache.Inputs{Name: "formatfiles-go-oracle",
		Files:     []string{"cohere/go.mod", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "stage1/cohere/formatfiles/testdata/cohere_side_test.go"},
		Flags:     []string{"go test -c -overlay ./internal/format/formatfiles", buildcache.Tool("go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "GOFLAGS")},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("go", "version")}}
	product := buildcache.Product(t, inputs, func(directory string) error {
		command := formatfilesCommand(t, "go", "test", "-c", "-overlay="+overlayPath, "-o", filepath.Join(directory, "oracle"), "./internal/format/formatfiles")
		command.Dir = cohere
		output, err := combinedOutput(command)
		if err != nil {
			return fmt.Errorf("build Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	command := formatfilesCommand(t, filepath.Join(product, "oracle"), "-test.run=^TestAdamicPortCases$", "-test.timeout=90s")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
}

// Bound child processes without an external timeout executable. Killing the
// process group also terminates compilers launched by the Go oracle build.
func formatfilesCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	parent := context.Background()
	if formatfilesSetupContext != nil {
		parent = formatfilesSetupContext
	}
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}
