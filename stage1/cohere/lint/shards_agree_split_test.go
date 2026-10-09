package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestShardsAgree requires the driver's output to be byte-identical however many processes share the
// manifest (#tj6d455): one, two, and the machine's cores. Equal counts cannot see a reordered or repeated
// case, so the whole output is compared, and the count mode too.
const testShardsAgreeShards = 16

func testShardsAgree(t *testing.T, shard int) {
	t.Helper()
	started := time.Now()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := testShardsAgreeOracle(t)
	rows := testShardsAgreeRows(t)
	buckets := testShardsAgreePartition(t, rows)
	selected := buckets[shard]
	t.Logf("union %d cases, shard %03d: %d cases", len(rows), shard, len(selected))
	rows = selected
	path := manifest(t, recoveryRows(t, oracle, rows))
	binary := testShardsAgreeBinary(t, directory)
	t.Logf("TestShardsAgree (setup): %.3fs", time.Since(started).Seconds())
	want := execute(t, "", binary, "--manifest", path)
	wantCount := execute(t, "", binary, "--manifest", path, "--count")
	for _, count := range []int{1, 2, runtime.NumCPU()} {
		got, err := shards.Run(binary, path, count, false)
		if err != nil {
			t.Fatal(err)
		}
		if diff := difference(got, want.output); diff != "" {
			t.Fatalf("%d shards: %s", count, diff)
		}
		gotCount, err := shards.Run(binary, path, count, true)
		if err != nil {
			t.Fatal(err)
		}
		if string(gotCount) != string(wantCount.output) {
			t.Fatalf("%d shards count %q, want %q", count, gotCount, wantCount.output)
		}
	}
	t.Logf("%d rows: identical at 1, 2 and %d shards, %d bytes", len(rows), runtime.NumCPU(), len(want.output))
}

func testShardsAgreeRows(t *testing.T) []string {
	rows := generated(t)
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
func testShardsAgreeInputs(name string) buildcache.Inputs {
	return buildcache.Inputs{Name: "shards-agree-" + name, Files: []string{"stage1/cohere/lint", "internal", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/static_single_assignment", "cohere/mutation_aliasing", "cohere/go.mod", "cohere/go.sum", "go.mod"}, Flags: append(native.Flags(native.Options{}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")), Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
}
func testShardsAgreeOracle(t *testing.T) string {
	d := buildcache.Product(t, testShardsAgreeInputs("oracle"), func(d string) error { _, err := goOracleIn(packageDirectory, d); return err })
	return filepath.Join(d, "oracle")
}
func testShardsAgreeUpstream(t *testing.T) []string {
	d := buildcache.Product(t, testShardsAgreeInputs("capture"), func(d string) error {
		rows, err := captureUpstream(packageDirectory, d)
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
	c := buildcache.Product(t, testShardsAgreeInputs("lowered"), func(d string) error {
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
	inputs := testShardsAgreeInputs("native")
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
	if bytes.Count(data, []byte("func TestShardsAgree_0")) != testShardsAgreeShards {
		t.Fatal("shard enumeration differs from constant")
	}
	rows := testShardsAgreeRows(t)
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
