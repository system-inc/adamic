package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/executable"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"

	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestShardsAgree requires the driver's output to be byte-identical however many processes share the
// manifest (#tj6d455): one, two, and the machine's cores. Equal counts cannot see a reordered or repeated
// case, so the whole output is compared, and the count mode too.
const testShardsAgreeShards = 32

// testShardsAgreeHeavy are the rows too heavy for a hash bucket, by the key testShardsAgreeOwner hashes (the base
// name, and the options when the row has any), each with the name its tests carry. checker.ts is one row that takes
// about 24 s per output pass and 6 s per count pass alone on one core, so the bucket it hashed to ran about 118 s of
// it (#wyt8ntk); here it has a test per mode instead. A future heavy row joins by name.
var testShardsAgreeHeavy = []struct{ Key, Name string }{
	{"checker.ts", "Checker"},
}

// testShardsAgreePasses are the process counts every bucket's rows are compared at. Zero is the unit's CPUs, never the
// host's cores: on a runner's 64-core box this ran 64 shard processes in an 8-CPU share.
var testShardsAgreePasses = []struct {
	Label string
	Count int
}{{"1", 1}, {"2", 2}, {"N", 0}}

// testShardsAgreeHeavyPasses are the process counts a heavy row is compared at, by the label its tests carry: one. A
// heavy entry is one row (testShardsAgreePartition refuses another size), and shards.Run deals rows round-robin, so at
// 2 and at the unit's CPUs that row runs on shard 0 exactly as at 1 and every other shard runs none: checker.ts printed
// the same 10,244,022 bytes at 1, 2 and 8 (lint#20/20, Oct 14), and those four passes were 64 to 87 s of the unit's
// 309 to 412 (#xj8b4ge). What they added, an empty shard that prints nothing, TestShardsRunEmptyShardsPrintNothing
// checks. Its name stays outside TestShardsAgree<Name>_, which TestShardsAgree_Union reads as a heavy row's wrapper.
var testShardsAgreeHeavyPasses = []struct {
	Label string
	Count int
}{{"1", 1}}

func testShardsAgreeCount(count int) int {
	if count == 0 {
		return runtime.GOMAXPROCS(0)
	}
	return count
}

func testShardsAgree(test *testing.T, shard int) {
	test.Helper()
	if os.Getenv("ADAMIC_SHARDS_AGREE_LEAF") != "1" {
		products := testShardsAgreeReady(test)
		leaf := testShardsAgreeLeafFor(test, products, products.Buckets[shard], fmt.Sprintf("bucket-%03d", shard))
		testShardsAgreeChild(test, test.Name(), "ADAMIC_SHARDS_AGREE_LEAF=1", "ADAMIC_SHARDS_AGREE_READY="+leaf)
		return
	}
	// The parent resolved every product and wrote what this leaf reads, so the leaf looks nothing up.
	leaf := testShardsAgreeLeafRead(test)
	if leaf.Rows == 0 {
		test.Log("empty shard")
		return
	}
	test.Logf("union %d cases, shard %03d: %d cases", leaf.Union, shard, leaf.Rows)
	want, wantCount := testShardsAgreeLeafReferences(test, leaf)
	type answer struct {
		output, count []byte
		err           error
	}
	answers := make([]answer, len(testShardsAgreePasses))
	// The counts are independent passes over the same manifest, so running them at most the parent's share at a time
	// changes no comparison, only how many shard processes this test keeps running at once.
	share := make(chan struct{}, max(1, leaf.Share))
	var workers sync.WaitGroup
	for i, pass := range testShardsAgreePasses {
		workers.Add(1)
		go func(i, count int) {
			defer workers.Done()
			share <- struct{}{}
			defer func() { <-share }()
			answers[i].output, answers[i].err = shards.Run(leaf.Binary, leaf.Manifest, count, false)
			if answers[i].err == nil {
				answers[i].count, answers[i].err = shards.Run(leaf.Binary, leaf.Manifest, count, true)
			}
		}(i, testShardsAgreeCount(pass.Count))
	}
	workers.Wait()
	for i, pass := range testShardsAgreePasses {
		count := testShardsAgreeCount(pass.Count)
		got := answers[i]
		if got.err != nil {
			test.Fatal(got.err)
		}
		if diff := difference(got.output, want); diff != "" {
			test.Fatalf("%d shards: %s", count, diff)
		}
		if string(got.count) != string(wantCount) {
			test.Fatalf("%d shards count %q, want %q", count, got.count, wantCount)
		}
	}
	test.Logf("%d rows: identical at 1, 2 and %d shards, %d bytes", leaf.Rows, testShardsAgreeCount(0), len(want))
}

// testShardsAgreeHeavyPass compares one heavy row's answer at one process count, in one mode, with its reference.
func testShardsAgreeHeavyPass(test *testing.T, name, label string, countOnly bool) {
	test.Helper()
	heavy := -1
	for index, entry := range testShardsAgreeHeavy {
		if entry.Name == name {
			heavy = index
		}
	}
	count := -1
	for _, pass := range testShardsAgreeHeavyPasses {
		if pass.Label == label {
			count = pass.Count
		}
	}
	if heavy < 0 || count < 0 {
		test.Fatalf("no heavy row %q or pass %q", name, label)
	}
	if os.Getenv("ADAMIC_SHARDS_AGREE_LEAF") != "1" {
		products := testShardsAgreeReady(test)
		leaf := testShardsAgreeLeafFor(test, products, products.Heavy[heavy], "heavy-"+name)
		testShardsAgreeChild(test, test.Name(), "ADAMIC_SHARDS_AGREE_LEAF=1", "ADAMIC_SHARDS_AGREE_READY="+leaf)
		return
	}
	leaf := testShardsAgreeLeafRead(test)
	count = testShardsAgreeCount(count)
	test.Logf("union %d cases, heavy %s: %d cases", leaf.Union, name, leaf.Rows)
	want, wantCount := testShardsAgreeLeafReferences(test, leaf)
	got, err := shards.Run(leaf.Binary, leaf.Manifest, count, countOnly)
	if err != nil {
		test.Fatal(err)
	}
	if countOnly {
		if string(got) != string(wantCount) {
			test.Fatalf("%d shards count %q, want %q", count, got, wantCount)
		}
		test.Logf("%d rows: count identical at %d shards, %s", leaf.Rows, count, got)
		return
	}
	if diff := difference(got, want); diff != "" {
		test.Fatalf("%d shards: %s", count, diff)
	}
	test.Logf("%d rows: output identical at %d shards, %d bytes", leaf.Rows, count, len(want))
}

// testShardsAgreeLeaf is what a leaf child reads, every path resolved by the parent that launched it.
type testShardsAgreeLeaf struct {
	Binary, Manifest, Output, Count string
	Rows, Union, Share              int
}

// testShardsAgreeLeafFor writes rows' manifest and their references, read back on this machine, beside a leaf file
// for the child, and returns the leaf file's path. The child would otherwise look up the prepared product and every
// product its rows name again, about 8 s per leaf on one core (#wyt8ntk).
func testShardsAgreeLeafFor(test *testing.T, products testShardsAgreeProducts, rows []string, label string) string {
	test.Helper()
	directory := test.TempDir()
	leaf := testShardsAgreeLeaf{
		Binary: products.Binary,
		Rows:   len(rows),
		Union:  len(products.Rows),
		Share:  testguard.Share(),
	}
	if len(rows) > 0 {
		leaf.Manifest = manifest(test, rows)
		references := testShardsAgreeReferencesReady(test)
		for _, file := range []struct {
			extension string
			path      *string
		}{{".output", &leaf.Output}, {".count", &leaf.Count}} {
			data, err := os.ReadFile(filepath.Join(references, label+file.extension))
			if err != nil {
				test.Fatal(err)
			}
			*file.path = filepath.Join(directory, label+file.extension)
			if err = os.WriteFile(*file.path, []byte(buildcache.Absolute(string(data))), 0644); err != nil {
				test.Fatal(err)
			}
		}
	}
	data, err := json.Marshal(leaf)
	if err != nil {
		test.Fatal(err)
	}
	path := filepath.Join(directory, "leaf.json")
	if err = os.WriteFile(path, data, 0644); err != nil {
		test.Fatal(err)
	}
	return path
}

func testShardsAgreeLeafRead(test *testing.T) testShardsAgreeLeaf {
	test.Helper()
	data, err := os.ReadFile(os.Getenv("ADAMIC_SHARDS_AGREE_READY"))
	if err != nil {
		test.Fatal(err)
	}
	var leaf testShardsAgreeLeaf
	if err = json.Unmarshal(data, &leaf); err != nil {
		test.Fatal(err)
	}
	return leaf
}

func testShardsAgreeLeafReferences(test *testing.T, leaf testShardsAgreeLeaf) ([]byte, []byte) {
	test.Helper()
	output, err := os.ReadFile(leaf.Output)
	if err != nil {
		test.Fatal(err)
	}
	count, err := os.ReadFile(leaf.Count)
	if err != nil {
		test.Fatal(err)
	}
	return output, count
}

func testShardsAgreeRows(test *testing.T, rows []string) []string {
	for _, row := range testShardsAgreeUpstream(test) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	// The compiler files are the corpus with large files, where shards differ most in what they hold, so
	// the test refuses to run without them rather than passing on a smaller corpus.
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal(
			"set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout: TestShardsAgree needs its compiler files",
		)
	}
	var matches []string
	var err error
	buildcache.Keying(func() { matches, err = filepath.Glob(filepath.Join(source, "src/compiler/*.ts")) })
	if err != nil {
		test.Fatal(err)
	}
	if len(matches) == 0 {
		test.Fatalf("no compiler files under %s", source)
	}
	rows = append(rows, matches...)
	return rows
}

// These products include implementation, runtime, rule and toolchain inputs. No
// artifact belongs to a shard's TempDir, so parallel leaves can safely reuse it.
func testShardsAgreeInputs(test *testing.T, name string) buildcache.Inputs {
	test.Helper()
	files := []string{
		"internal",
		"cohere/internal",
		"cohere/TypeScript/tsc",
		"cohere/TypeScript-shim",
		"cohere/static_single_assignment",
		"cohere/mutation_aliasing",
		"cohere/go.mod",
		"cohere/go.sum",
		"go.mod",
		"stage1/cohere/lint/shared_test.go",
		"stage1/cohere/lint/registry",
		"stage1/cohere/lint/testdata/oracle.go",
	}
	if name == "capture" {
		files = append(files, "stage1/cohere/lint/shards_agree_split_test.go")
	}
	for _, path := range portFiles(test) {
		files = append(files, "stage1/cohere/lint/"+filepath.ToSlash(path))
	}
	tools := []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
	if name == "capture" || strings.HasPrefix(name, "prepared") {
		// The capture runs cohere's go test with toolchain's go, and the prepared corpus holds its rows. go's settings
		// are asked here, under the root's go.work; the capture runs in cohere, under cohere's, so both workspaces are
		// inputs too.
		tools = append(tools, lintGoSettings())
		files = append(files, "go.work", "cohere/go.work", "cohere/go.work.sum")
	}
	return buildcache.Inputs{
		Name:      "shards-agree-" + name,
		Files:     files,
		Flags:     append(native.Flags(native.Options{}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
		Toolchain: tools,
	}
}
func testShardsAgreeUpstream(test *testing.T) []string {
	test.Helper()
	return oracles.ShardsAgreeCapture(test)
}
func testShardsAgreeLowered(test *testing.T, directory string) string {
	return buildcache.Product(test, testShardsAgreeInputs(test, "lowered"), func(d string) error {
		prepareRegistry(test, directory)
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
}
func testShardsAgreeBinary(test *testing.T, directory string) string {
	c := testShardsAgreeLowered(test, directory)
	inputs := testShardsAgreeInputs(test, "native")
	// The lowered input is already content-addressed by all its source inputs.
	d := buildcache.Product(test, inputs, func(d string) error {
		data, err := os.ReadFile(filepath.Join(c, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(d, "scanner"), native.Options{})
	})
	return filepath.Join(d, "scanner")
}

// testShardsAgreeKey names a row by its file's base name and its options: temporary capture/generated roots are
// deliberately absent from this key.
func testShardsAgreeKey(row string) string {
	fields := strings.SplitN(row, "\t", 2)
	key := filepath.Base(fields[0])
	if len(fields) > 1 {
		key += "\t" + fields[1]
	}
	return key
}
func testShardsAgreeOwner(test *testing.T, row string) int {
	test.Helper()
	data, err := os.ReadFile(strings.SplitN(row, "\t", 2)[0])
	if err != nil {
		test.Fatal(err)
	}
	return testShardsAgreeOwnerBytes(row, data)
}

// testShardsAgreeOwnerBytes is the existing row hash over the exact bytes
// already included in the prepared product's compiler-corpus input hash.
func testShardsAgreeOwnerBytes(row string, data []byte) int {
	sum := sha256.Sum256(append([]byte(testShardsAgreeKey(row)+"\x00"), data...))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testShardsAgreeShards)
}

// testShardsAgreeHeavyOwner is the index in testShardsAgreeHeavy of the entry naming row, or -1 for a bucket's row.
func testShardsAgreeHeavyOwner(row string) int {
	for index, entry := range testShardsAgreeHeavy {
		if testShardsAgreeKey(row) == entry.Key {
			return index
		}
	}
	return -1
}

// testShardsAgreePartition places every row exactly once: in the heavy entry naming it, else in its hash bucket.
// A heavy entry naming no row fails, so the list never outlives the row it was made for.
func testShardsAgreePartition(
	test *testing.T,
	rows []string,
	sources ...map[string][]byte,
) (buckets, heavy [][]string) {
	owner := func(row string) int {
		if len(sources) > 0 {
			if data, ok := sources[0][strings.SplitN(row, "\t", 2)[0]]; ok {
				return testShardsAgreeOwnerBytes(row, data)
			}
		}
		return testShardsAgreeOwner(test, row)
	}
	buckets = make([][]string, testShardsAgreeShards)
	heavy = make([][]string, len(testShardsAgreeHeavy))
	visits := make([]int, len(rows))
	for index, entry := range testShardsAgreeHeavy {
		for i, row := range rows {
			if testShardsAgreeHeavyOwner(row) == index {
				heavy[index] = append(heavy[index], row)
				visits[i]++
			}
		}
		if len(heavy[index]) != 1 {
			test.Fatalf(
				"heavy row %q names %d rows of the corpus, want one: a heavy entry is compared at one process",
				entry.Key,
				len(heavy[index]),
			)
		}
	}
	for shard := range buckets {
		for i, row := range rows {
			if testShardsAgreeHeavyOwner(row) < 0 && owner(row) == shard {
				buckets[shard] = append(buckets[shard], row)
				visits[i]++
			}
		}
	}
	for i, n := range visits {
		if n != 1 {
			test.Fatalf("case %d visited %d times", i, n)
		}
	}
	return buckets, heavy
}
func TestShardsAgree_000(test *testing.T) { test.Parallel(); testShardsAgree(test, 0) }
func TestShardsAgree_001(test *testing.T) { test.Parallel(); testShardsAgree(test, 1) }
func TestShardsAgree_002(test *testing.T) { test.Parallel(); testShardsAgree(test, 2) }
func TestShardsAgree_003(test *testing.T) { test.Parallel(); testShardsAgree(test, 3) }
func TestShardsAgree_004(test *testing.T) { test.Parallel(); testShardsAgree(test, 4) }
func TestShardsAgree_005(test *testing.T) { test.Parallel(); testShardsAgree(test, 5) }
func TestShardsAgree_006(test *testing.T) { test.Parallel(); testShardsAgree(test, 6) }
func TestShardsAgree_007(test *testing.T) { test.Parallel(); testShardsAgree(test, 7) }
func TestShardsAgree_008(test *testing.T) { test.Parallel(); testShardsAgree(test, 8) }
func TestShardsAgree_009(test *testing.T) { test.Parallel(); testShardsAgree(test, 9) }
func TestShardsAgree_010(test *testing.T) { test.Parallel(); testShardsAgree(test, 10) }
func TestShardsAgree_011(test *testing.T) { test.Parallel(); testShardsAgree(test, 11) }
func TestShardsAgree_012(test *testing.T) { test.Parallel(); testShardsAgree(test, 12) }
func TestShardsAgree_013(test *testing.T) { test.Parallel(); testShardsAgree(test, 13) }
func TestShardsAgree_014(test *testing.T) { test.Parallel(); testShardsAgree(test, 14) }
func TestShardsAgree_015(test *testing.T) { test.Parallel(); testShardsAgree(test, 15) }

// Setup fetches the same products as independently selected shards.

func testShardsAgreePrepare(test *testing.T) string {
	test.Helper()
	started := time.Now()
	inputs, sources := testShardsAgreePreparedInputs(test)
	d := buildcache.Product(test, inputs, func(d string) error {
		binary := testShardsAgreeBinary(test, packageDirectory)
		oracle := lintGoOracleProduct(test)
		generatedRows := generated(test)
		generatedPaths := map[string]bool{}
		for _, row := range generatedRows {
			generatedPaths[strings.SplitN(row, "\t", 2)[0]] = true
		}
		rows := testShardsAgreeRows(test, generatedRows)
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
		buckets, heavy := testShardsAgreePartition(test, rows, sources)
		for i := range buckets {
			buckets[i] = testShardsAgreeRecoveryRows(test, oracle, buckets[i], sources)
		}
		for i := range heavy {
			heavy[i] = testShardsAgreeRecoveryRows(test, oracle, heavy[i], sources)
		}
		products := testShardsAgreeProducts{Root: d, Binary: binary, Rows: rows, Buckets: buckets, Heavy: heavy}
		data, err := json.Marshal(products)
		if err != nil {
			return err
		}
		// This product, the scanner, the capture's cases and TypeScript's sources are each named by a root the reader
		// supplies: Product publishes this scratch directory by rename, and a runner holds all four elsewhere
		// (#tqrqx60).
		return os.WriteFile(filepath.Join(d, "prepared.json"), []byte(buildcache.Relative(string(data))), 0644)
	})
	test.Logf("TestShardsAgree (setup): %.3fs; %s", time.Since(started).Seconds(), d)
	return d
}

// Heavy holds each testShardsAgreeHeavy entry's rows, in the list's order.
type testShardsAgreeProducts struct {
	Root    string
	Binary  string
	Rows    []string
	Buckets [][]string
	Heavy   [][]string
}

func testShardsAgreePreparedInputs(test *testing.T) (buildcache.Inputs, map[string][]byte) {
	sources := map[string][]byte{}
	inputs := testShardsAgreeInputs(test, "prepared-v1")
	inputs.Files = append(
		inputs.Files,
		"stage1/cohere/lint/lint_test.go",
		"stage1/cohere/lint/shards_agree_split_test.go",
	)
	// The rows hold the Go oracle's classifications, so the oracle's bytes key them: its build arguments live in
	// oracle_product_test.go, which these files don't name.
	oracle, err := os.ReadFile(lintGoOracleProduct(test))
	if err != nil {
		test.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(oracle)))
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to pinned TypeScript checkout")
	}
	var paths []string
	buildcache.Keying(func() { paths, err = filepath.Glob(filepath.Join(source, "src/compiler/*.ts")) })
	if err != nil {
		test.Fatal(err)
	}
	if len(paths) == 0 {
		test.Fatalf("no compiler files under %s", source)
	}
	hash := sha256.New()
	for _, path := range paths {
		var data []byte
		buildcache.Keying(func() { data, err = os.ReadFile(path) })
		if err != nil {
			test.Fatal(err)
		}
		sources[path] = data
		// Named inside the TypeScript checkout, not by where this machine keeps it (#t37sw0f).
		relative, err := filepath.Rel(source, path)
		if err != nil {
			test.Fatal(err)
		}
		fmt.Fprintf(hash, "%s %d\n", filepath.ToSlash(relative), len(data))
		hash.Write(data)
	}
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("compiler corpus %x", hash.Sum(nil)))
	return inputs, sources
}

var testShardsAgreePreparedOnce sync.Once
var testShardsAgreePreparedRoot string
var testShardsAgreePrepared *testShardsAgreeProducts

// testShardsAgreeReady is the prepared product read back on this machine, looked up and read once per process.
func testShardsAgreeReady(test *testing.T) testShardsAgreeProducts {
	test.Helper()
	testShardsAgreePreparedOnce.Do(func() {
		testShardsAgreePreparedRoot = testShardsAgreePrepare(test)
		products := testShardsAgreeRead(test, testShardsAgreePreparedRoot)
		testShardsAgreePrepared = &products
	})
	if testShardsAgreePrepared == nil {
		test.Fatal("the prepared product failed earlier in this process")
	}
	return *testShardsAgreePrepared
}

var testShardsAgreeReferencesOnce sync.Once
var testShardsAgreeReferencesRoot string

func testShardsAgreeReferencesReady(test *testing.T) string {
	test.Helper()
	testShardsAgreeReferencesOnce.Do(func() { testShardsAgreeReferencesRoot = testShardsAgreeReferences(test) })
	if testShardsAgreeReferencesRoot == "" {
		test.Fatal("the references product failed earlier in this process")
	}
	return testShardsAgreeReferencesRoot
}

// testShardsAgreeReferences is each bucket's and each heavy entry's answer from one process, output and count: the
// same scanner over the same rows the leaves run, computed once when products are filled rather than in every unit
// (#wyt8ntk). Every comparison with it still runs live in a leaf. Its declared inputs are the prepared product's,
// which hold the scanner's and the capture's, and its build reads the prepared rows, the scanner and every row's file,
// so a trace keys it on all three products.
func testShardsAgreeReferences(test *testing.T) string {
	test.Helper()
	testShardsAgreeReady(test)
	root := testShardsAgreePreparedRoot
	inputs, sources := testShardsAgreePreparedInputs(test)
	inputs.Name = "shards-agree-references-v1"
	inputs.Flags = append(inputs.Flags, "reads shards-agree-prepared-v1 shards-agree-native lint-rules-agree-capture")
	return buildcache.Product(test, inputs, func(d string) error {
		products := testShardsAgreeRead(test, root)
		type job struct {
			label string
			rows  []string
		}
		var jobs []job
		for shard, rows := range products.Buckets {
			if len(rows) > 0 {
				jobs = append(jobs, job{fmt.Sprintf("bucket-%03d", shard), rows})
			}
		}
		for index, entry := range testShardsAgreeHeavy {
			jobs = append(jobs, job{"heavy-" + entry.Name, products.Heavy[index]})
		}
		// The build phase owns its CPUs: one scanner per CPU, the heaviest row first so it never runs last alone.
		sort.SliceStable(jobs, func(i, j int) bool {
			return strings.HasPrefix(jobs[i].label, "heavy-") && !strings.HasPrefix(jobs[j].label, "heavy-")
		})
		failures := make([]error, 2*len(jobs))
		share := make(chan struct{}, runtime.GOMAXPROCS(0))
		var workers sync.WaitGroup
		for i, job := range jobs {
			copied, restore := testShardsAgreeSourceSnapshots(test, job.rows, sources)
			path := manifest(test, copied)
			for m, mode := range []struct {
				extension string
				arguments []string
			}{{".output", nil}, {".count", []string{"--count"}}} {
				workers.Add(1)
				go func() {
					defer workers.Done()
					share <- struct{}{}
					defer func() { <-share }()
					output, err := testShardsAgreeRun(
						"",
						nil,
						products.Binary,
						append([]string{"--manifest", path}, mode.arguments...)...,
					)
					if err == nil {
						// The answer names the rows' files, which this product names by the roots a reader supplies
						// (#tqrqx60); a text Absolute would not read back as it was is refused rather than compared
						// wrongly.
						output = restore(output)
						relative := buildcache.Relative(string(output))
						if buildcache.Absolute(relative) != string(output) {
							err = fmt.Errorf("%s%s does not read back through Absolute", job.label, mode.extension)
						} else {
							err = os.WriteFile(filepath.Join(d, job.label+mode.extension), []byte(relative), 0644)
						}
					}
					failures[2*i+m] = err
				}()
			}
		}
		workers.Wait()
		for _, err := range failures {
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func testShardsAgreeRead(test *testing.T, d string) testShardsAgreeProducts {
	test.Helper()
	data, err := os.ReadFile(filepath.Join(d, "prepared.json"))
	if err != nil {
		test.Fatal(err)
	}
	var products testShardsAgreeProducts
	if err = json.Unmarshal([]byte(buildcache.Absolute(string(data))), &products); err != nil {
		test.Fatal(err)
	}
	products.Root = d
	return products
}

func TestShardsAgree_016(test *testing.T) { test.Parallel(); testShardsAgree(test, 16) }

func TestShardsAgree_017(test *testing.T) { test.Parallel(); testShardsAgree(test, 17) }

func TestShardsAgree_018(test *testing.T) { test.Parallel(); testShardsAgree(test, 18) }

func TestShardsAgree_019(test *testing.T) { test.Parallel(); testShardsAgree(test, 19) }

func TestShardsAgree_020(test *testing.T) { test.Parallel(); testShardsAgree(test, 20) }

func TestShardsAgree_021(test *testing.T) { test.Parallel(); testShardsAgree(test, 21) }

func TestShardsAgree_022(test *testing.T) { test.Parallel(); testShardsAgree(test, 22) }

func TestShardsAgree_023(test *testing.T) { test.Parallel(); testShardsAgree(test, 23) }

func TestShardsAgree_024(test *testing.T) { test.Parallel(); testShardsAgree(test, 24) }

func TestShardsAgree_025(test *testing.T) { test.Parallel(); testShardsAgree(test, 25) }

func TestShardsAgree_026(test *testing.T) { test.Parallel(); testShardsAgree(test, 26) }

func TestShardsAgree_027(test *testing.T) { test.Parallel(); testShardsAgree(test, 27) }

func TestShardsAgree_028(test *testing.T) { test.Parallel(); testShardsAgree(test, 28) }

func TestShardsAgree_029(test *testing.T) { test.Parallel(); testShardsAgree(test, 29) }

func TestShardsAgree_030(test *testing.T) { test.Parallel(); testShardsAgree(test, 30) }

func TestShardsAgree_031(test *testing.T) { test.Parallel(); testShardsAgree(test, 31) }

// Each heavy row has a test per heavy pass and mode, named for it; TestShardsAgree_Union reads these lines.

func TestShardsAgreeChecker_1_Output(test *testing.T) {
	test.Parallel()
	testShardsAgreeHeavyPass(test, "Checker", "1", false)
}

func TestShardsAgreeChecker_1_Count(test *testing.T) {
	test.Parallel()
	testShardsAgreeHeavyPass(test, "Checker", "1", true)
}

// An empty shard prints nothing, and its count is zero: one bucket row in a manifest of its own, run at one process and
// at the unit's CPUs (at least two), so every shard but 0 has no row, gives the same output and count. A planted empty
// shard that prints a byte (the real binary behind a wrapper that writes one first for any shard but 0) fails both.
func TestShardsRunEmptyShardsPrintNothing(test *testing.T) {
	test.Parallel()
	products := testShardsAgreeReady(test)
	row := ""
	for _, bucket := range products.Buckets {
		if len(bucket) > 0 {
			row = bucket[0]
			break
		}
	}
	if row == "" {
		test.Fatal("no bucket holds a row")
	}
	path := manifest(test, []string{row})
	count := max(2, testShardsAgreeCount(0))
	planted := filepath.Join(test.TempDir(), "planted")
	script := fmt.Sprintf(
		"#!/bin/sh\ncase \"$*\" in *\"--shard 0/\"*) ;; *) printf x ;; esac\nexec '%s' \"$@\"\n",
		products.Binary,
	)
	if err := executable.Write(planted, []byte(script)); err != nil {
		test.Fatal(err)
	}
	for _, countOnly := range []bool{false, true} {
		one, err := shards.Run(products.Binary, path, 1, countOnly)
		if err != nil {
			test.Fatal(err)
		}
		many, err := shards.Run(products.Binary, path, count, countOnly)
		if err != nil {
			test.Fatal(err)
		}
		if !bytes.Equal(one, many) {
			test.Fatalf("count only %t: %d shards over one row printed %q, one shard %q", countOnly, count, many, one)
		}
		if _, err := shards.Run(planted, path, count, countOnly); err == nil {
			test.Fatalf("count only %t: an empty shard that prints a byte passed", countOnly)
		}
	}
	test.Logf("one row at 1 and %d shards: the same output and count; a planted empty shard's byte fails both", count)
}

// Product recipes run without a test-side deadline. The build phase owns their ceiling.
// testShardsAgreeRun runs name in directory. A bare name is a tool (the capture's go): the one internal/toolchain
// resolves, under its environment plus environment, with PWD set to directory. A path is a program a test built (the
// port, the oracle), run as named with this process's environment.
func testShardsAgreeRun(directory string, environment []string, name string, args ...string) ([]byte, error) {
	var command *exec.Cmd
	if strings.ContainsRune(name, filepath.Separator) {
		command = exec.Command(name, args...)
		command.Dir = directory
		// An explicit environment loses the PWD os/exec sets from Dir, and the go command trusts PWD over the
		// real working directory, so a stale one resolves the module through the wrong path.
		command.Env = append(os.Environ(), environment...)
		if directory != "" {
			command.Env = append(command.Env, "PWD="+directory)
		}
	} else {
		command = toolchain.CommandIn(directory, name, args...)
		command.Env = append(command.Env, environment...)
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

func testShardsAgreeRecoveryRows(test *testing.T, oracle string, rows []string, sources map[string][]byte) []string {
	test.Helper()
	if len(rows) == 0 {
		return rows
	}
	copied, _ := testShardsAgreeSourceSnapshots(test, rows, sources)
	output, err := testShardsAgreeRun("", nil, oracle, "--manifest", manifest(test, copied), "--diagnostics")
	if err != nil {
		test.Fatal(err)
	}
	answer := execution{output: output}
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		test.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
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

// testShardsAgreeSourceSnapshots gives children the exact external bytes hashed
// by PreparedInputs, without reading an unkeyable checkout during a build. Rows
// stored in the products keep their original paths and ownership; outputs are
// restored before Relative serializes them. Other source roots stay unchanged.
func testShardsAgreeSourceSnapshots(
	test *testing.T,
	rows []string,
	sources map[string][]byte,
) ([]string, func([]byte) []byte) {
	test.Helper()
	directory := test.TempDir()
	copied := append([]string{}, rows...)
	type replacement struct{ from, to []byte }
	var replacements []replacement
	paths := map[string]string{}
	for i, row := range rows {
		fields := strings.Split(row, "\t")
		data, ok := sources[fields[0]]
		if !ok {
			continue
		}
		target, ok := paths[fields[0]]
		if !ok {
			target = filepath.Join(directory, fmt.Sprint(i), filepath.Base(fields[0]))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				test.Fatal(err)
			}
			if err := os.WriteFile(target, data, 0644); err != nil {
				test.Fatal(err)
			}
			paths[fields[0]] = target
			replacements = append(replacements, replacement{[]byte(target), []byte(fields[0])})
		}
		fields[0] = target
		copied[i] = strings.Join(fields, "\t")
	}
	return copied, func(output []byte) []byte {
		for _, r := range replacements {
			output = bytes.ReplaceAll(output, r.from, r.to)
		}
		return output
	}
}

func TestShardsAgree_SourceSnapshotsMatchGo(test *testing.T) {
	test.Parallel()
	_, sources := testShardsAgreePreparedInputs(test)
	root := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	rows := []string{filepath.Join(root, "src/compiler/checker.ts"), filepath.Join(root, "src/compiler/binder.ts")}
	for _, path := range rows {
		data, ok := sources[path]
		if !ok {
			test.Fatalf("missing snapshot source %s", path)
		}
		if testShardsAgreeOwner(test, path) != testShardsAgreeOwnerBytes(path, data) {
			test.Fatalf("source snapshot changes ownership for %s", path)
		}
	}
	copied, restore := testShardsAgreeSourceSnapshots(test, rows, sources)
	oracle := lintGoOracleProduct(test)
	for _, args := range [][]string{nil, {"--count"}} {
		want, err := testShardsAgreeRun(
			"",
			nil,
			oracle,
			append([]string{"--manifest", manifest(test, rows)}, args...)...,
		)
		if err != nil {
			test.Fatal(err)
		}
		got, err := testShardsAgreeRun(
			"",
			nil,
			oracle,
			append([]string{"--manifest", manifest(test, copied)}, args...)...,
		)
		if err != nil {
			test.Fatal(err)
		}
		if !bytes.Equal(restore(got), want) {
			test.Fatalf("source snapshot differs from Go for %v", args)
		}
	}
}
