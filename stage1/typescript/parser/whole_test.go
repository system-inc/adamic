package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wholeNode(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest, "--whole"}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}

// The whole-tree comparison over the pinned compiler runs in compilerExpressionsPartition's shards (each file owned by
// its base name's hash), so no one test parses all 77 files on Go, Node and sanitized native: 26.4 s alone on one core
// as one test (Workshop, Oct 11).
func wholeCompilerShard(t *testing.T, index int) {
	t.Helper()
	shards, planted, _ := compilerExpressionsEnumeration(t)
	paths := shards[index]
	manifest := filepath.Join(t.TempDir(), "compiler.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(paths, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	for caseIndex, path := range paths {
		if path != planted {
			continue
		}
		marker := []byte(fmt.Sprintf("case %d\n", caseIndex))
		offset := bytes.Index(want.output, marker) + len(marker)
		if offset < len(marker) || offset >= len(want.output) {
			t.Fatal("planted case absent from oracle output")
		}
		mutant := append([]byte(nil), want.output...)
		mutant[offset] ^= 1
		if difference(mutant, want.output) == "" {
			t.Fatal("planted disagreement survived")
		}
		t.Logf("planted disagreement caught by shard-%03d for %s", index, filepath.Base(planted))
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	got := wholeNode(t, directory, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}
	t.Logf("shard-%03d: %d whole compiler files, %d identical whole-tree bytes", index, len(paths), len(want.output))
}

func TestWholeCompilerAgrees_000(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 0)
}

func TestWholeCompilerAgrees_001(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 1)
}

func TestWholeCompilerAgrees_002(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 2)
}

func TestWholeCompilerAgrees_003(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 3)
}

func TestWholeCompilerAgrees_004(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 4)
}

func TestWholeCompilerAgrees_005(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 5)
}

func TestWholeCompilerAgrees_006(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 6)
}

func TestWholeCompilerAgrees_007(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 7)
}

func TestWholeCompilerAgrees_008(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 8)
}

func TestWholeCompilerAgrees_009(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 9)
}

func TestWholeCompilerAgrees_010(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 10)
}

func TestWholeCompilerAgrees_011(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 11)
}

func TestWholeCompilerAgrees_012(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 12)
}

func TestWholeCompilerAgrees_013(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 13)
}

func TestWholeCompilerAgrees_014(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 14)
}

func TestWholeCompilerAgrees_015(t *testing.T) {
	t.Parallel()
	wholeCompilerShard(t, 15)
}

func TestWholeCompilerAgreesUnion(t *testing.T) {
	t.Parallel()
	functions := []func(*testing.T){TestWholeCompilerAgrees_000, TestWholeCompilerAgrees_001, TestWholeCompilerAgrees_002, TestWholeCompilerAgrees_003, TestWholeCompilerAgrees_004, TestWholeCompilerAgrees_005, TestWholeCompilerAgrees_006, TestWholeCompilerAgrees_007, TestWholeCompilerAgrees_008, TestWholeCompilerAgrees_009, TestWholeCompilerAgrees_010, TestWholeCompilerAgrees_011, TestWholeCompilerAgrees_012, TestWholeCompilerAgrees_013, TestWholeCompilerAgrees_014, TestWholeCompilerAgrees_015}
	if len(functions) != testCompilerExpressionsAgreeShards {
		t.Fatal("top-level shard enumeration differs")
	}
	shards, planted, files := compilerExpressionsEnumeration(t)
	count, catches, caught := 0, 0, -1
	for index, shard := range shards {
		for _, path := range shard {
			count++
			want := []byte("case 0\nSourceFile\n")
			got := append([]byte(nil), want...)
			if path == planted {
				got[len("case 0\n")] ^= 1
			}
			if difference(got, want) != "" {
				catches++
				caught = index
			}
		}
	}
	if count != files || catches != 1 {
		t.Fatalf("union=%d/%d planted catches=%d", count, files, catches)
	}
	t.Logf("union=%d each exactly once; planted disagreement %s caught by shard-%03d", count, filepath.Base(planted), caught)
}

func wholeManifest(t *testing.T, cases []string) string {
	t.Helper()
	directory := t.TempDir()
	var manifest strings.Builder
	for i, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("whole-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
