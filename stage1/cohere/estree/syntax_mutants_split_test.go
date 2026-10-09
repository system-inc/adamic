package estree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testSyntaxMutantsShards = 3

type syntaxMutant struct{ name, file, from, to, source string }

func syntaxMutantEnumeration() []syntaxMutant {
	return []syntaxMutant{
		{"mapped-constraint", "convert.ts", "this.set(result, 'constraint', this.converted(this.child(parameter, 1)));", "this.set(result, 'constraint', absent());", ""},
		{"erasure-precedence", "sourceBinary.ts", "if(nextRank > lastRank ||", "if(false && nextRank > lastRank ||", "1+1 as number *2;"},
		{"reference-pragma", "pipeline.ts", "if(reference !== '')", "if(false)", "/// <reference path='missingquote.ts />\nx;"},
	}
}

func syntaxMutantProducts(t *testing.T, item syntaxMutant) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:  "estree-syntax-mutant-lowered-" + item.name,
		Files: []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Flags: []string{item.file, item.from, item.to, "repository=" + root(t), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		main := mutantPort(t, item.file, item.from, item.to)
		program, err := load.Load([]string{main})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		// Source remains usable after the building test's temporary directories disappear.
		files, err := filepath.Glob(filepath.Join(filepath.Dir(main), "*.ts"))
		if err != nil {
			return err
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, filepath.Base(file)), data, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	inputs.Name = "estree-syntax-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(lowered, "main.ts"), filepath.Join(product, "port")
}

type syntaxMutantPrepared struct {
	Main, Binary, Manifest string
	Want                   []byte
}

var syntaxMutantsPrepared []syntaxMutantPrepared
var syntaxMutantsPreparedDirectory string

func syntaxMutantSetupInputs(t *testing.T) buildcache.Inputs {
	return buildcache.Inputs{Name: "syntax-mutants-setup-v1", Files: []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod", "go.work"}, Flags: []string{"root=" + root(t), "sanitize=true", "split=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
}
func syntaxMutantRead(t *testing.T, dir string) []syntaxMutantPrepared {
	data, err := os.ReadFile(filepath.Join(dir, "ready.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result []syntaxMutantPrepared
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != len(syntaxMutantEnumeration()) {
		t.Fatal("incomplete setup")
	}
	return result
}

// Not parallel: publishes shared build products before parallel shards are released.
func TestSyntaxMutants_Setup(t *testing.T) {
	started := time.Now()
	dir := buildcache.Product(t, syntaxMutantSetupInputs(t), func(dir string) error {
		items := syntaxMutantEnumeration()
		prepared := make([]syntaxMutantPrepared, len(items))
		// GoBuild is absent on this base: retain the original overlay Go build.
		var oracle string
		oracleReady := make(chan struct{})
		go func() { oracle = goOracle(t); close(oracleReady) }()
		var workers sync.WaitGroup
		for i, item := range items {
			workers.Add(1)
			go func(i int, item syntaxMutant) {
				defer workers.Done()
				main, binary := syntaxMutantProducts(t, item)
				sources := []string{item.source}
				if i == 0 {
					sources = syntaxGrammar()
				}
				list := manifest(t, sources)
				args := []string{"--manifest", list}
				if i != 0 {
					args = []string{"--audit", list, t.TempDir()}
				}
				<-oracleReady
				want := execute(t, "", oracle, args...)
				// Preserve source filenames in the oracle output by publishing the original manifest and inputs.
				paths, err := os.ReadFile(list)
				if err != nil {
					t.Error(err)
					return
				}
				persistent := filepath.Join(dir, fmt.Sprintf("cases-%d", i))
				if err := os.MkdirAll(persistent, 0755); err != nil {
					t.Error(err)
					return
				}
				// Oracle canonical output contains no temporary directory names; input basenames are kept.
				names := strings.Fields(string(paths))
				var published []string
				for _, name := range names {
					data, err := os.ReadFile(name)
					if err != nil {
						t.Error(err)
						return
					}
					path := filepath.Join(persistent, filepath.Base(name))
					if err := os.WriteFile(path, data, 0644); err != nil {
						t.Error(err)
						return
					}
					published = append(published, path)
				}
				manifestPath := filepath.Join(persistent, "manifest")
				if err := os.WriteFile(manifestPath, []byte(strings.Join(published, "\n")+"\n"), 0644); err != nil {
					t.Error(err)
					return
				}
				prepared[i] = syntaxMutantPrepared{main, binary, manifestPath, want}
			}(i, item)
		}
		workers.Wait()
		if t.Failed() {
			return fmt.Errorf("setup failed")
		}
		data, err := json.Marshal(prepared)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "ready.json"), data, 0644)
	})
	syntaxMutantsPrepared = syntaxMutantRead(t, dir)
	syntaxMutantsPreparedDirectory = dir
	t.Logf("TestSyntaxMutants (setup): %.3fs", time.Since(started).Seconds())
}
func syntaxMutantReady(t *testing.T) []syntaxMutantPrepared {
	if dir := os.Getenv("ADAMIC_SYNTAX_MUTANTS_READY"); dir != "" {
		return syntaxMutantRead(t, dir)
	}
	if syntaxMutantsPrepared != nil {
		return syntaxMutantsPrepared
	}
	dir := buildcache.Product(t, syntaxMutantSetupInputs(t), func(string) error { return fmt.Errorf("run TestSyntaxMutants_Setup first; shards never build") })
	return syntaxMutantRead(t, dir)
}
func runSyntaxMutantShard(t *testing.T, shard int) {
	item := syntaxMutantEnumeration()[shard]
	prepared := syntaxMutantReady(t)[shard]
	if shard != 0 && !strings.Contains(string(prepared.Want), `"status":"error"`) {
		t.Fatal(string(prepared.Want))
	}
	for name, got := range map[string][]byte{"Node": onNode(t, prepared.Main, "--manifest", prepared.Manifest), "native": execute(t, "", prepared.Binary, "--manifest", prepared.Manifest)} {
		if os.Getenv("ADAMIC_SYNTAX_MUTANTS_PLANT") == "1" && shard == 0 && name == "native" {
			got = prepared.Want
		}
		if shard == 0 {
			if diff := firstDifference(prepared.Want, got); diff == "" {
				t.Fatal(name + " mutant survived")
			} else {
				t.Log(name + ": " + diff)
			}
		} else {
			if !strings.Contains(string(got), "0 Program ") {
				t.Fatal(name + " control did not accept")
			}
			t.Log(name + ": disabled check accepts Go-refused input; acceptance oracle catches it")
		}
	}
	t.Log(item.name)
}
func TestSyntaxMutants_000(t *testing.T) { t.Parallel(); runSyntaxMutantShard(t, 0) }
func TestSyntaxMutants_001(t *testing.T) { t.Parallel(); runSyntaxMutantShard(t, 1) }
func TestSyntaxMutants_002(t *testing.T) { t.Parallel(); runSyntaxMutantShard(t, 2) }

var syntaxMutantRunners = [...]func(*testing.T){TestSyntaxMutants_000, TestSyntaxMutants_001, TestSyntaxMutants_002}

func TestSyntaxMutantsUnion(t *testing.T) {
	t.Parallel()
	items := syntaxMutantEnumeration()
	if len(syntaxMutantRunners) != testSyntaxMutantsShards || len(items) != testSyntaxMutantsShards {
		t.Fatal("shard count mismatch")
	}
	seen := map[string]int{}
	count := 0
	for i := range syntaxMutantRunners {
		item := items[i]
		sources := []string{item.source}
		if i == 0 {
			sources = syntaxGrammar()
		}
		for _, source := range sources {
			seen[item.name+"\x00"+source]++
			count++
		}
	}
	for _, item := range items {
		sources := []string{item.source}
		if item.name == "mapped-constraint" {
			sources = syntaxGrammar()
		}
		for _, source := range sources {
			if seen[item.name+"\x00"+source] != 1 {
				t.Fatal("case coverage mismatch")
			}
		}
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	caught := []int{}
	for i := range syntaxMutantRunners {
		name := fmt.Sprintf("TestSyntaxMutants_%03d", i)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := exec.CommandContext(ctx, executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_SYNTAX_MUTANTS_PLANT=1")
		if syntaxMutantsPreparedDirectory != "" {
			command.Env = append(command.Env, "ADAMIC_SYNTAX_MUTANTS_READY="+syntaxMutantsPreparedDirectory)
		}
		output, err := command.CombinedOutput()
		contextErr := ctx.Err()
		cancel()
		if contextErr != nil {
			t.Fatalf("cooked: %s exceeded 90s", name)
		}
		if err != nil {
			if !strings.Contains(string(output), "mutant survived") || !strings.Contains(string(output), "--- FAIL: "+name) {
				t.Fatalf("unexpected failure: %v\n%s", err, output)
			}
			caught = append(caught, i)
		}
	}
	if len(caught) != 1 || caught[0] != 0 {
		t.Fatalf("planted failure caught by %v", caught)
	}
	t.Logf("union: %d cases exactly once; planted survivor caught only by TestSyntaxMutants_000", count)
}
