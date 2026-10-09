package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestShardsAgree requires the driver's output to be byte-identical however many processes share the
// manifest (#tj6d455): one, two, and the machine's cores. Equal counts cannot see a reordered or repeated
// case, so the whole output is compared, and the count mode too.
const testShardsAgreeShards = 32

func testShardsAgree(t *testing.T, shard int) {
	t.Helper()
	testShardsAgreeGroupDeadline(t)
	if os.Getenv("ADAMIC_SHARDS_AGREE_LEAF") != "1" {
		products := testShardsAgreeReady(t)
		testShardsAgreeChild(t, t.Name(), "ADAMIC_SHARDS_AGREE_LEAF=1", "ADAMIC_SHARDS_AGREE_READY="+products.Root)
		return
	}
	products := testShardsAgreeRead(t, os.Getenv("ADAMIC_SHARDS_AGREE_READY"))
	// The parent's lookup has completed before this child's 90-second deadline.
	rows := products.Buckets[shard]
	if len(rows) == 0 {
		t.Log("empty shard")
		return
	}
	path := manifest(t, rows)
	binary := products.Binary
	t.Logf("union %d cases, shard %03d: %d cases", len(products.Rows), shard, len(rows))
	want := testShardsAgreeExecute(t, "", binary, "--manifest", path)
	wantCount := testShardsAgreeExecute(t, "", binary, "--manifest", path, "--count")
	counts := []int{1, 2, runtime.NumCPU()}
	type answer struct {
		output, count []byte
		err           error
	}
	answers := make([]answer, len(counts))
	var workers sync.WaitGroup
	for i, count := range counts {
		workers.Add(1)
		go func(i, count int) {
			defer workers.Done()
			answers[i].output, answers[i].err = shards.Run(binary, path, count, false)
			if answers[i].err == nil {
				answers[i].count, answers[i].err = shards.Run(binary, path, count, true)
			}
		}(i, count)
	}
	workers.Wait()
	for i, count := range counts {
		got := answers[i]
		if got.err != nil {
			t.Fatal(got.err)
		}
		if diff := difference(got.output, want.output); diff != "" {
			t.Fatalf("%d shards: %s", count, diff)
		}
		if string(got.count) != string(wantCount.output) {
			t.Fatalf("%d shards count %q, want %q", count, got.count, wantCount.output)
		}
	}
	t.Logf("%d rows: identical at 1, 2 and %d shards, %d bytes", len(rows), runtime.NumCPU(), len(want.output))
}

func testShardsAgreeRows(t *testing.T, rows []string) []string {
	for _, row := range testShardsAgreeUpstream(t) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	// The compiler files are the corpus with large files, where shards differ most in what they hold, so
	// the test refuses to run without them rather than passing on a smaller corpus.
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout: TestShardsAgree needs its compiler files")
	}
	matches, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("no compiler files under %s", source)
	}
	rows = append(rows, matches...)
	return rows
}

// These products include implementation, runtime, rule and toolchain inputs. No
// artifact belongs to a shard's TempDir, so parallel leaves can safely reuse it.
func testShardsAgreeInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	files := []string{"internal", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/static_single_assignment", "cohere/mutation_aliasing", "cohere/go.mod", "cohere/go.sum", "go.mod", "stage1/cohere/lint/shared_test.go", "stage1/cohere/lint/registry", "stage1/cohere/lint/testdata/oracle.go"}
	if name == "oracle" || name == "capture" {
		files = append(files, "stage1/cohere/lint/shards_agree_split_test.go")
	}
	for _, path := range portFiles(t) {
		files = append(files, "stage1/cohere/lint/"+filepath.ToSlash(path))
	}
	return buildcache.Inputs{Name: "shards-agree-" + name, Files: files, Flags: append(native.Flags(native.Options{}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")), Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
}
func testShardsAgreeOracle(t *testing.T) string {
	d := buildcache.Product(t, testShardsAgreeInputs(t, "oracle"), func(d string) error { _, err := testShardsAgreeGoOracleIn(packageDirectory, d); return err })
	return filepath.Join(d, "oracle")
}
func testShardsAgreeUpstream(t *testing.T) []string {
	d := buildcache.Product(t, testShardsAgreeInputs(t, "capture"), func(d string) error {
		rows, err := testShardsAgreeCaptureUpstream(packageDirectory, d)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(d, "rows"), []byte(strings.Join(rows, "\n")), 0644)
	})
	// Product atomically renames its scratch directory; rewrite stored paths.
	data, err := os.ReadFile(filepath.Join(d, "rows"))
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(string(data), "\n")
	for i, row := range rows {
		fields := strings.SplitN(row, "\t", 2)
		at := strings.Index(fields[0], "/case-")
		if at < 0 {
			t.Fatalf("capture path: %q", row)
		}
		fields[0] = d + fields[0][at:]
		rows[i] = strings.Join(fields, "\t")
	}
	return rows
}
func testShardsAgreeBinary(t *testing.T, directory string) string {
	c := buildcache.Product(t, testShardsAgreeInputs(t, "lowered"), func(d string) error {
		prepareRegistry(t, directory)
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(d, "main.c"), []byte(native.C(lowered)), 0644)
	})
	inputs := testShardsAgreeInputs(t, "native")
	// The lowered input is already content-addressed by all its source inputs.
	d := buildcache.Product(t, inputs, func(d string) error {
		data, err := os.ReadFile(filepath.Join(c, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(d, "scanner"), native.Options{})
	})
	return filepath.Join(d, "scanner")
}
func testShardsAgreeOwner(t *testing.T, row string) int {
	t.Helper()
	fields := strings.SplitN(row, "\t", 2)
	// Temporary capture/generated roots are deliberately absent from this key.
	key := filepath.Base(fields[0])
	if len(fields) > 1 {
		key += "\t" + fields[1]
	}
	data, err := os.ReadFile(fields[0])
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(append([]byte(key+"\x00"), data...))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testShardsAgreeShards)
}
func testShardsAgreePartition(t *testing.T, rows []string) [][]string {
	buckets := make([][]string, testShardsAgreeShards)
	visits := make([]int, len(rows))
	for shard := range buckets {
		for i, row := range rows {
			if testShardsAgreeOwner(t, row) == shard {
				buckets[shard] = append(buckets[shard], row)
				visits[i]++
			}
		}
	}
	for i, n := range visits {
		if n != 1 {
			t.Fatalf("case %d visited %d times", i, n)
		}
	}
	return buckets
}
func TestShardsAgree_Union(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("shards_agree_split_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < testShardsAgreeShards; i++ {
		name := fmt.Sprintf("func TestShardsAgree_%03d(", i)
		if bytes.Count(data, []byte(name)) != 1 {
			t.Fatalf("enumeration missing/duplicate %s", name)
		}
	}
	if len(regexp.MustCompile(`(?m)^func TestShardsAgree_[0-9]{3}\(`).FindAll(data, -1)) != testShardsAgreeShards {
		t.Fatal("shard enumeration differs from constant")
	}
	products := testShardsAgreeReady(t)
	rows := products.Rows
	buckets := testShardsAgreePartition(t, rows)
	// Plant one output disagreement and use the same byte oracle as every leaf.
	planted := rows[0]
	caught := []int{}
	for shard, bucket := range buckets {
		for _, row := range bucket {
			want := []byte(row)
			got := append([]byte(nil), want...)
			if row == planted {
				got = append(got, '!')
			}
			if difference(got, want) != "" {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 || caught[0] != testShardsAgreeOwner(t, planted) {
		t.Fatalf("planted disagreement caught by %v", caught)
	}
	t.Logf("union %d cases across %d shards", len(rows), testShardsAgreeShards)
	t.Logf("planted disagreement caught exactly once by TestShardsAgree_%03d", caught[0])
}
func TestShardsAgree_000(t *testing.T) { t.Parallel(); testShardsAgree(t, 0) }
func TestShardsAgree_001(t *testing.T) { t.Parallel(); testShardsAgree(t, 1) }
func TestShardsAgree_002(t *testing.T) { t.Parallel(); testShardsAgree(t, 2) }
func TestShardsAgree_003(t *testing.T) { t.Parallel(); testShardsAgree(t, 3) }
func TestShardsAgree_004(t *testing.T) { t.Parallel(); testShardsAgree(t, 4) }
func TestShardsAgree_005(t *testing.T) { t.Parallel(); testShardsAgree(t, 5) }
func TestShardsAgree_006(t *testing.T) { t.Parallel(); testShardsAgree(t, 6) }
func TestShardsAgree_007(t *testing.T) { t.Parallel(); testShardsAgree(t, 7) }
func TestShardsAgree_008(t *testing.T) { t.Parallel(); testShardsAgree(t, 8) }
func TestShardsAgree_009(t *testing.T) { t.Parallel(); testShardsAgree(t, 9) }
func TestShardsAgree_010(t *testing.T) { t.Parallel(); testShardsAgree(t, 10) }
func TestShardsAgree_011(t *testing.T) { t.Parallel(); testShardsAgree(t, 11) }
func TestShardsAgree_012(t *testing.T) { t.Parallel(); testShardsAgree(t, 12) }
func TestShardsAgree_013(t *testing.T) { t.Parallel(); testShardsAgree(t, 13) }
func TestShardsAgree_014(t *testing.T) { t.Parallel(); testShardsAgree(t, 14) }
func TestShardsAgree_015(t *testing.T) { t.Parallel(); testShardsAgree(t, 15) }

// Setup publishes an immutable corpus and products for separately filtered leaves.
// Not parallel: publishes shared products before the parallel leaves are released.
func TestShardsAgree_Setup(t *testing.T) {
	if os.Getenv("ADAMIC_SHARDS_AGREE_SETUP") != "1" {
		testShardsAgreeChild(t, t.Name(), "ADAMIC_SHARDS_AGREE_SETUP=1")
		return
	}
	testShardsAgreeGroupDeadline(t)
	started := time.Now()
	inputs := testShardsAgreePreparedInputs(t)
	d := buildcache.Product(t, inputs, func(d string) error {
		binary := testShardsAgreeBinary(t, packageDirectory)
		oracle := testShardsAgreeOracle(t)
		generatedRows := generated(t)
		generatedPaths := map[string]bool{}
		for _, row := range generatedRows {
			generatedPaths[strings.SplitN(row, "\t", 2)[0]] = true
		}
		rows := testShardsAgreeRows(t, generatedRows)
		if err := os.Mkdir(filepath.Join(d, "generated"), 0755); err != nil {
			return err
		}
		// generated(t) owns temporary sources: copy them into this persistent product.
		copies := map[string]string{}
		for i, row := range rows {
			fields := strings.SplitN(row, "\t", 2)
			// Copy only the exact generated paths; corpus basenames may overlap.
			base := filepath.Base(fields[0])
			if generatedPaths[fields[0]] {
				original := fields[0]
				target, ok := copies[original]
				if !ok {
					target = filepath.Join(d, "generated", base)
					data, err := os.ReadFile(original)
					if err != nil {
						return err
					}
					if err = os.WriteFile(target, data, 0644); err != nil {
						return err
					}
					copies[original] = target
				}
				fields[0] = target
				rows[i] = strings.Join(fields, "\t")
			}
		}
		buckets := testShardsAgreePartition(t, rows)
		for i := range buckets {
			buckets[i] = testShardsAgreeRecoveryRows(t, oracle, buckets[i])
		}
		products := testShardsAgreeProducts{Root: d, Binary: binary, Rows: rows, Buckets: buckets}
		data, err := json.Marshal(products)
		if err != nil {
			return err
		}
		// Scratch paths are rewritten on read after Product publishes by rename.
		return os.WriteFile(filepath.Join(d, "prepared.json"), data, 0644)
	})
	t.Logf("TestShardsAgree (setup): %.3fs; %s", time.Since(started).Seconds(), d)
}

type testShardsAgreeProducts struct {
	Root    string
	Binary  string
	Rows    []string
	Buckets [][]string
}

func testShardsAgreePreparedInputs(t *testing.T) buildcache.Inputs {
	inputs := testShardsAgreeInputs(t, "prepared-v1")
	inputs.Files = append(inputs.Files, "stage1/cohere/lint/lint_test.go", "stage1/cohere/lint/shards_agree_split_test.go")
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to pinned TypeScript checkout")
	}
	paths, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no compiler files under %s", source)
	}
	hash := sha256.New()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(hash, "%s %d\n", path, len(data))
		hash.Write(data)
	}
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("compiler corpus %x", hash.Sum(nil)))
	return inputs
}
func testShardsAgreeReady(t *testing.T) testShardsAgreeProducts {
	t.Helper()
	// A miss is a failure, never a lazy build in a leaf. Run Setup separately first.
	d := buildcache.Product(t, testShardsAgreePreparedInputs(t), func(string) error {
		return fmt.Errorf("shared setup missing: run TestShardsAgree_Setup before selecting leaves")
	})
	return testShardsAgreeRead(t, d)
}
func testShardsAgreeRead(t *testing.T, d string) testShardsAgreeProducts {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(d, "prepared.json"))
	if err != nil {
		t.Fatal(err)
	}
	var products testShardsAgreeProducts
	if err = json.Unmarshal(data, &products); err != nil {
		t.Fatal(err)
	}
	rewrite := func(row string) string {
		fields := strings.SplitN(row, "\t", 2)
		if strings.HasPrefix(fields[0], products.Root+string(filepath.Separator)) {
			fields[0] = d + strings.TrimPrefix(fields[0], products.Root)
		}
		return strings.Join(fields, "\t")
	}
	for i, row := range products.Rows {
		products.Rows[i] = rewrite(row)
	}
	for i := range products.Buckets {
		for j, row := range products.Buckets[i] {
			products.Buckets[i][j] = rewrite(row)
		}
	}
	products.Root = d
	return products
}
func testShardsAgreeChild(t *testing.T, name string, markers ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=0", "-test.v")
	command.Env = append(append(os.Environ(), markers...), "ADAMIC_SHARDS_AGREE_BOUND=1")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	output, err := command.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Fatalf("%s: %v (deadline: %v)", name, err, ctx.Err())
	}
}

func TestShardsAgree_SetupRequired(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestShardsAgree_000$", "-test.timeout=0", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+t.TempDir(), "ADAMIC_SHARDS_AGREE_LEAF=0", "ADAMIC_SHARDS_AGREE_BOUND=1")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("missing-setup probe exceeded deadline: %v", ctx.Err())
	}
	if err == nil || !bytes.Contains(output, []byte("shared setup missing")) {
		t.Fatalf("leaf must refuse an unprepared cache: %v\n%s", err, output)
	}
	if bytes.Contains(output, []byte("build shards-agree-oracle")) || bytes.Contains(output, []byte("build shards-agree-lowered")) {
		t.Fatalf("leaf attempted shared builds: %s", output)
	}
}

func TestShardsAgree_016(t *testing.T) { t.Parallel(); testShardsAgree(t, 16) }

func TestShardsAgree_017(t *testing.T) { t.Parallel(); testShardsAgree(t, 17) }

func TestShardsAgree_018(t *testing.T) { t.Parallel(); testShardsAgree(t, 18) }

func TestShardsAgree_019(t *testing.T) { t.Parallel(); testShardsAgree(t, 19) }

func TestShardsAgree_020(t *testing.T) { t.Parallel(); testShardsAgree(t, 20) }

func TestShardsAgree_021(t *testing.T) { t.Parallel(); testShardsAgree(t, 21) }

func TestShardsAgree_022(t *testing.T) { t.Parallel(); testShardsAgree(t, 22) }

func TestShardsAgree_023(t *testing.T) { t.Parallel(); testShardsAgree(t, 23) }

func TestShardsAgree_024(t *testing.T) { t.Parallel(); testShardsAgree(t, 24) }

func TestShardsAgree_025(t *testing.T) { t.Parallel(); testShardsAgree(t, 25) }

func TestShardsAgree_026(t *testing.T) { t.Parallel(); testShardsAgree(t, 26) }

func TestShardsAgree_027(t *testing.T) { t.Parallel(); testShardsAgree(t, 27) }

func TestShardsAgree_028(t *testing.T) { t.Parallel(); testShardsAgree(t, 28) }

func TestShardsAgree_029(t *testing.T) { t.Parallel(); testShardsAgree(t, 29) }

func TestShardsAgree_030(t *testing.T) { t.Parallel(); testShardsAgree(t, 30) }

func TestShardsAgree_031(t *testing.T) { t.Parallel(); testShardsAgree(t, 31) }

// Keep every subprocess in the outer CommandContext process group. The package
// run/execute helpers install separate testguard groups, so using them here
// would let compiler or scanner descendants escape the 90-second cancellation.
func testShardsAgreeRun(directory string, environment []string, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Inherit the outer bounded group; a new group would escape its cancellation.
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	// An explicit environment loses the PWD os/exec sets from Dir, and the go command trusts PWD over the
	// real working directory, so a stale one resolves the module through the wrong path.
	command.Env = append(os.Environ(), environment...)
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	output, err := os.CreateTemp(sharedDirectory, "stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func testShardsAgreeExecute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Inherit the outer bounded group; a new group would escape its cancellation.
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func testShardsAgreeRecoveryRows(t *testing.T, oracle string, rows []string) []string {
	t.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := testShardsAgreeExecute(t, "", oracle, "--manifest", manifest(t, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		t.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
	}
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row
		if flags[index] != "1" {
			continue
		}
		fields := strings.Split(row, "\t")
		for len(fields) < 7 {
			fields = append(fields, "")
		}
		if fields[6] == "" {
			fields[6] = "recovery"
		}
		result[index] = strings.Join(fields, "\t")
	}
	return result
}

func testShardsAgreeGoOracleIn(sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	if _, err := testShardsAgreeRun(root, nil, "go", args...); err != nil {
		return "", err
	}
	return binary, nil
}

// captureUpstream runs cohere's own tests for sourceRoot's rules under a capture overlay and writes each
// unique asserted case as a file under directory, returning one manifest row per case. Every Run is
// captured, including tests that assert repair fields directly; the overlay changes no rule. The capture's
// destination is passed in the subprocess's environment, so no test's process environment changes.
func testShardsAgreeCaptureUpstream(sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := testShardsAgreeRun(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0"); err != nil {
			return nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}

// Go's outer test alarm can exit the controller before CommandContext cancels.
// The child independently kills its inherited group at the same 90-second
// bound, including compilers and scanners, even after that controller is gone.
func testShardsAgreeGroupDeadline(t *testing.T) {
	t.Helper()
	if os.Getenv("ADAMIC_SHARDS_AGREE_BOUND") != "1" {
		return
	}
	group := syscall.Getpgrp()
	timer := time.AfterFunc(90*time.Second, func() { _ = syscall.Kill(-group, syscall.SIGKILL) })
	t.Cleanup(func() { timer.Stop() })
}
