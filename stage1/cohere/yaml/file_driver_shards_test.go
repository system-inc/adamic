package yaml

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testFileDriverShards = 8

func fileDriverPartition(inputs []string) [][]int {
	shards := make([][]int, testFileDriverShards)
	for index, input := range inputs {
		h := fnv.New64a()
		_, _ = h.Write([]byte(input))
		shard := int(h.Sum64() % testFileDriverShards)
		shards[shard] = append(shards[shard], index)
	}
	return shards
}

func fileDriverCheckUnion(t *testing.T, shards [][]int, count int) {
	t.Helper()
	if len(shards) != testFileDriverShards {
		t.Fatal("shard enumeration disagrees with constant")
	}
	seen := make([]int, count)
	for _, indices := range shards {
		for _, index := range indices {
			if index < 0 || index >= count {
				t.Fatalf("invalid case %d", index)
			}
			seen[index]++
		}
	}
	for index, visits := range seen {
		if visits != 1 {
			t.Fatalf("case %d visited %d times", index, visits)
		}
	}
}

func fileDriverBuildFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"go.mod", "go.sum", "cohere/go.mod", "cohere/TypeScript/tsc/go.mod"}
	// Include the transitive implementation inputs without hashing test evidence.
	for _, directory := range []string{"stage1/cohere/yaml", "internal", "cohere"} {
		err := filepath.WalkDir(filepath.Join(repository, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			extension := filepath.Ext(path)
			if strings.HasSuffix(path, "_test.go") || (extension != ".go" && extension != ".ts" && extension != ".c" && extension != ".h") {
				return nil
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			files = append(files, relative)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func testFileDriver(t *testing.T) {
	started := time.Now()
	cases, files, _ := formatCases(t)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[:files]
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, text := range []string{"", "  \t\n", "\ufeff", "\ufeffa: b\n", "|+", "|+\n", "|+ # header\n", ">+", "a: |+", "a: |+\n", "a: |+\n  b\n\n", "{a: b, c: [x,y]}\n", "---\n...\n", "key: 'x\\y'\n"} {
		inputs = append(inputs, "0\t"+escape.Replace(text))
	}
	path := filepath.Join(t.TempDir(), "driver-cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(inputs, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := bytes.Split(bytes.TrimSuffix(goFormat(t, path), []byte("\n")), []byte("\n"))
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	productInputs := buildcache.Inputs{
		Name: "yaml-file-driver-lowered", Files: fileDriverBuildFiles(t),
		Flags:     []string{"lower+native.C+javascript.JavaScript"},
		Toolchain: []string{runtime.Version()},
	}
	loweredProduct := buildcache.Product(t, productInputs, func(directory string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "format.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "format.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	})
	productInputs.Name = "yaml-file-driver-native"
	productInputs.Flags = append(native.Flags(native.Options{Sanitize: true}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	productInputs.Toolchain = append(productInputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeProduct := buildcache.Product(t, productInputs, func(directory string) error {
		source, err := os.ReadFile(filepath.Join(loweredProduct, "format.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(directory, "format"), native.Options{Sanitize: true})
	})
	binary := filepath.Join(nativeProduct, "format")
	emitted := filepath.Join(loweredProduct, "format.mjs")
	runner := filepath.Join(root, "oracle/node.mjs")
	if len(expected) != len(inputs) {
		t.Fatalf("oracle returned %d answers for %d inputs", len(expected), len(inputs))
	}
	shards := fileDriverPartition(inputs)
	fileDriverCheckUnion(t, shards, len(inputs))
	plantedCaught := make(chan int, testFileDriverShards)
	t.Cleanup(func() {
		close(plantedCaught)
		var caught []int
		for shard := range plantedCaught {
			caught = append(caught, shard)
		}
		if len(caught) != 1 {
			t.Fatalf("planted disagreement caught by %v", caught)
		}
		t.Logf("union %d cases exactly once; planted native stdout case 0 caught by shard-%03d", len(inputs), caught[0])
	})
	t.Logf("TestFileDriver (setup): %.3fs", time.Since(started).Seconds())
	for shard, indices := range shards {
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			started := time.Now()
			defer func() { t.Logf("shard time %.3fs", time.Since(started).Seconds()) }()
			for _, index := range indices {
				input := inputs[index]
				answer := string(expected[index])
				if !strings.HasPrefix(answer, "ok\t") {
					t.Fatalf("driver control %d invalid: %s", index, answer)
				}
				wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(answer, "ok\t")))
				file := filepath.Join(t.TempDir(), "input.yaml")
				if err := os.WriteFile(file, []byte(unescapeCase(input)), 0644); err != nil {
					t.Fatal(err)
				}
				for _, side := range []struct {
					name string
					out  []byte
				}{{"native", run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, file)}, {"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, file)}, {"emitted JavaScript", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, file)}} {
					// Prove the real byte oracle rejects a planted stdout disagreement.
					if index == 0 && side.name == "native" {
						planted := append(append([]byte(nil), side.out...), '!')
						if bytes.Equal(planted, wanted) {
							t.Fatal("missed planted native stdout disagreement")
						}
						plantedCaught <- shard
					}
					if !bytes.Equal(side.out, wanted) {
						t.Fatalf("%s file %d: %s", side.name, index, firstDifference(side.out, wanted))
					}
				}
			}
		})
	}
	t.Logf("%d repository files and %d direct stdout controls byte-identical to Go on all three port executions", files, len(inputs)-files)
}
