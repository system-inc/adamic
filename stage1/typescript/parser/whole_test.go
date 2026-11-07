package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wholeNode(test *testing.T, directory, manifest string, count bool) execution {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	args := []string{
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		manifest,
		"--whole",
	}
	if count {
		args = append(args, "--count")
	}
	return execute(test, "", "node", args...)
}

// The whole-tree comparison over the pinned compiler runs in compilerExpressionsPartition's shards (each file owned by
// its base name's hash), so no one test parses all 77 files on Go, Node and sanitized native: 26.4 s alone on one core
// as one test (Workshop, Oct 11).
func wholeCompilerShard(test *testing.T, index int) {
	test.Helper()
	shards, planted, _ := compilerExpressionsEnumeration(test)
	paths := shards[index]
	manifest := filepath.Join(test.TempDir(), "compiler.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(paths, "\n")+"\n"), 0644); err != nil {
		test.Fatal(err)
	}
	oracle := goOracle(test)
	want := execute(test, "", oracle, "--manifest", manifest, "--whole")
	for caseIndex, path := range paths {
		if path != planted {
			continue
		}
		marker := []byte(fmt.Sprintf("case %d\n", caseIndex))
		offset := bytes.Index(want.output, marker) + len(marker)
		if offset < len(marker) || offset >= len(want.output) {
			test.Fatal("planted case absent from oracle output")
		}
		mutant := append([]byte(nil), want.output...)
		mutant[offset] ^= 1
		if difference(mutant, want.output) == "" {
			test.Fatal("planted disagreement survived")
		}
		test.Logf("planted disagreement caught by shard-%03d for %s", index, filepath.Base(planted))
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		test.Fatal(err)
	}
	got := wholeNode(test, directory, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("Node: %s", diff)
	}
	got = execute(test, "", buildPort(test, directory, true), "--manifest", manifest, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("native: %s", diff)
	}
	test.Logf("shard-%03d: %d whole compiler files, %d identical whole-tree bytes", index, len(paths), len(want.output))
}

func TestWholeCompilerAgrees_000(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 0)
}

func TestWholeCompilerAgrees_001(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 1)
}

func TestWholeCompilerAgrees_002(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 2)
}

func TestWholeCompilerAgrees_003(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 3)
}

func TestWholeCompilerAgrees_004(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 4)
}

func TestWholeCompilerAgrees_005(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 5)
}

func TestWholeCompilerAgrees_006(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 6)
}

func TestWholeCompilerAgrees_007(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 7)
}

func TestWholeCompilerAgrees_008(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 8)
}

func TestWholeCompilerAgrees_009(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 9)
}

func TestWholeCompilerAgrees_010(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 10)
}

func TestWholeCompilerAgrees_011(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 11)
}

func TestWholeCompilerAgrees_012(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 12)
}

func TestWholeCompilerAgrees_013(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 13)
}

func TestWholeCompilerAgrees_014(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 14)
}

func TestWholeCompilerAgrees_015(test *testing.T) {
	test.Parallel()
	wholeCompilerShard(test, 15)
}

func wholeManifest(test *testing.T, cases []string) string {
	test.Helper()
	directory := test.TempDir()
	var manifest strings.Builder
	for i, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("whole-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			test.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		test.Fatal(err)
	}
	return path
}
