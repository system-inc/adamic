package estree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
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

func syntaxMutantProductInputs(t *testing.T, item syntaxMutant) buildcache.Inputs {
	t.Helper()
	return buildcache.Inputs{
		Name:  "estree-syntax-mutant-lowered-" + item.name,
		Files: []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere", "go.mod"},
		Flags: []string{item.file, item.from, item.to, "repository=" + root(t), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

func syntaxMutantLoweredProduct(t *testing.T, item syntaxMutant) string {
	t.Helper()
	return buildcache.Product(t, syntaxMutantProductInputs(t, item), func(dir string) error {
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
}

func syntaxMutantNativeProduct(t *testing.T, item syntaxMutant) string {
	t.Helper()
	lowered := syntaxMutantLoweredProduct(t, item)
	inputs := syntaxMutantProductInputs(t, item)
	inputs.Name = "estree-syntax-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "port.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
}

func syntaxMutantProducts(t *testing.T, item syntaxMutant) (string, string) {
	t.Helper()
	lowered := syntaxMutantLoweredProduct(t, item)
	product := syntaxMutantNativeProduct(t, item)
	return filepath.Join(lowered, "main.ts"), filepath.Join(product, "port")
}

type syntaxMutantPrepared struct {
	Main, Binary string
	Sources      []string
	Want         []byte
}

// Preparation is shared within a process, including independently selected shards.
var syntaxMutantsPrepareOnce sync.Once
var syntaxMutantsPrepared []syntaxMutantPrepared

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

// Both the build-phase unit and shard preparation use this exact oracle recipe.
func syntaxMutantOracleProduct(t *testing.T) string {
	t.Helper()
	inputs := syntaxMutantSetupInputs(t)
	inputs.Name = "syntax-mutants-go-oracle-v1"
	return buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(goOracle(t))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
	})
}

func prepareSyntaxMutants(t *testing.T) {
	t.Helper()
	dir := buildcache.Product(t, syntaxMutantSetupInputs(t), func(dir string) error {
		items := syntaxMutantEnumeration()
		prepared := make([]syntaxMutantPrepared, len(items))
		// GoBuild is absent on this base: retain the original overlay Go build.
		var oracle string
		oracleReady := make(chan struct{})
		go func() {
			defer close(oracleReady)
			oracle = filepath.Join(syntaxMutantOracleProduct(t), "oracle")
		}()
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
				prepared[i] = syntaxMutantPrepared{Main: main, Binary: binary, Sources: sources, Want: want}
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
}
func syntaxMutantReady(t *testing.T) []syntaxMutantPrepared {
	t.Helper()
	syntaxMutantsPrepareOnce.Do(func() { prepareSyntaxMutants(t) })
	if len(syntaxMutantsPrepared) != len(syntaxMutantEnumeration()) {
		t.Fatal("syntax mutant preparation failed")
	}
	return syntaxMutantsPrepared
}

// The setup manifest is itself a product; shards still prepare it independently.
func TestProduct_SyntaxMutantsSetup(t *testing.T) {
	t.Parallel()
	syntaxMutantReady(t)
}

func TestProduct_SyntaxMutantsGoOracle(t *testing.T) {
	t.Parallel()
	syntaxMutantOracleProduct(t)
}

func TestProduct_SyntaxMutantLowered_000(t *testing.T) {
	t.Parallel()
	syntaxMutantLoweredProduct(t, syntaxMutantEnumeration()[0])
}

func TestProduct_SyntaxMutantLowered_001(t *testing.T) {
	t.Parallel()
	syntaxMutantLoweredProduct(t, syntaxMutantEnumeration()[1])
}

func TestProduct_SyntaxMutantLowered_002(t *testing.T) {
	t.Parallel()
	syntaxMutantLoweredProduct(t, syntaxMutantEnumeration()[2])
}

func TestProduct_SyntaxMutantNative_001(t *testing.T) {
	t.Parallel()
	syntaxMutantNativeProduct(t, syntaxMutantEnumeration()[1])
}

func TestProduct_SyntaxMutantNative_002(t *testing.T) {
	t.Parallel()
	syntaxMutantNativeProduct(t, syntaxMutantEnumeration()[2])
}

func runSyntaxMutantShard(t *testing.T, shard int) {
	prepared := syntaxMutantReady(t)[shard]
	started := time.Now()
	defer func() { t.Logf("shard work: %.3fs", time.Since(started).Seconds()) }()
	cpu := syntaxMutantCPU()
	defer func() { t.Logf("CPU: %.3fs", (syntaxMutantCPU() - cpu).Seconds()) }()
	item := syntaxMutantEnumeration()[shard]
	list := manifest(t, prepared.Sources)
	if shard != 0 && !strings.Contains(string(prepared.Want), `"status":"error"`) {
		t.Fatal(string(prepared.Want))
	}
	for name, got := range map[string][]byte{"Node": onNode(t, prepared.Main, "--manifest", list), "native": execute(t, "", prepared.Binary, "--manifest", list)} {
		if os.Getenv("ADAMIC_SYNTAX_MUTANTS_PLANT") == "1" && shard == 0 && name == "native" {
			wantCases := syntaxMutantCanonicalCases(t, prepared.Want, len(prepared.Sources))
			gotCases := syntaxMutantCanonicalCases(t, got, len(prepared.Sources))
			witness := -1
			for i, source := range prepared.Sources {
				if strings.Contains(source, "[P in A]") {
					if witness != -1 {
						t.Fatal("duplicate planted witness")
					}
					witness = i
				}
			}
			if witness < 0 {
				t.Fatal("missing planted witness")
			}
			if wantCases[witness] == gotCases[witness] {
				t.Fatal("planted witness already agrees")
			}
			gotCases[witness] = wantCases[witness]
			got = []byte(strings.Join(gotCases, ""))
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
	syntaxMutantReady(t)
	items := syntaxMutantEnumeration()
	if len(syntaxMutantRunners) != testSyntaxMutantsShards || len(items) != testSyntaxMutantsShards {
		t.Fatal("shard count mismatch")
	}
	seen := map[string]int{}
	count := 0
	for i, runner := range syntaxMutantRunners {
		name := runtime.FuncForPC(reflect.ValueOf(runner).Pointer()).Name()
		if !strings.HasSuffix(name, fmt.Sprintf(".TestSyntaxMutants_%03d", i)) {
			t.Fatalf("unexpected shard runner %s", name)
		}
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

// Canonical manifests terminate each case with a stripped-source line. Source
// newlines are escaped, so boundaries preserve every original comparison byte.
func syntaxMutantCanonicalCases(t *testing.T, data []byte, count int) []string {
	t.Helper()
	var cases []string
	var current strings.Builder
	for _, line := range strings.SplitAfter(string(data), "\n") {
		current.WriteString(line)
		if strings.HasPrefix(line, "stripped ") {
			cases = append(cases, current.String())
			current.Reset()
		}
	}
	if current.Len() != 0 || len(cases) != count {
		t.Fatalf("canonical cases %d, want %d; trailing bytes %d", len(cases), count, current.Len())
	}
	return cases
}
func syntaxMutantCPU() time.Duration {
	var self, children syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &self) != nil || syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children) != nil {
		return 0
	}
	return time.Duration(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec)*time.Second + time.Duration(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)*time.Microsecond
}
