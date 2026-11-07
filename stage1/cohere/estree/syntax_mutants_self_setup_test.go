package estree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
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
	"github.com/system-inc/adamic/internal/native"
)

const testSyntaxMutantsShards = 3

type syntaxMutant struct{ name, file, from, to, source string }

func syntaxMutantEnumeration() []syntaxMutant {
	return []syntaxMutant{
		{
			"mapped-constraint",
			"convert.ts",
			"this.set(result, 'constraint', this.converted(this.child(parameter, 1)));",
			"this.set(result, 'constraint', absent());",
			"",
		},
		{
			"erasure-precedence",
			"sourceBinary.ts",
			"if(nextRank > lastRank ||",
			"if(false && nextRank > lastRank ||",
			"1+1 as number *2;",
		},
		{
			"reference-pragma",
			"pipeline.ts",
			"if(reference !== '')",
			"if(false)",
			"/// <reference path='missingquote.ts />\nx;",
		},
	}
}

func syntaxMutantProductInputs(test *testing.T, item syntaxMutant) buildcache.Inputs {
	test.Helper()
	inputs := buildcache.Inputs{
		Name:  "estree-syntax-mutant-lowered-" + item.name,
		Files: estreeFamilyDependencyFiles(test),
		Flags: []string{
			item.file,
			item.from,
			item.to,
			"repository=" + root(test),
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	inputs.Flags = append(inputs.Flags, estreeFamilyDependencyFlags(test)...)
	return inputs
}

func syntaxMutantIRProduct(test *testing.T, item syntaxMutant) string {
	test.Helper()
	inputs := syntaxMutantProductInputs(test, item)
	// v2: the IR's maps are sorted and its positions placeless, so a v1 program.gob isn't one this reads.
	inputs.Name = "estree-syntax-mutant-ir-v2-" + item.name
	return buildcache.Product(test, inputs, func(dir string) error {
		main := mutantPort(test, item.file, item.from, item.to)
		files, globError := filepath.Glob(filepath.Join(filepath.Dir(main), "*.ts"))
		if globError != nil {
			return globError
		}
		snapshot := map[string][]byte{}
		for _, file := range files {
			data, readError := os.ReadFile(file)
			if readError != nil {
				return readError
			}
			snapshot[filepath.Base(file)] = data
		}
		return estreeFamilyLowerCheckpoint(dir, snapshot)
	})
}
func syntaxMutantLoweredProduct(test *testing.T, item syntaxMutant) string {
	test.Helper()
	checkpoint := syntaxMutantIRProduct(test, item)
	inputs := syntaxMutantProductInputs(test, item)
	inputs.Name = "estree-syntax-mutant-emitted-" + item.name
	return buildcache.Product(test, inputs, func(dir string) error {
		program, estreeFamilyReadCheckpointError := estreeFamilyReadCheckpoint(checkpoint)
		if estreeFamilyReadCheckpointError != nil {
			return estreeFamilyReadCheckpointError
		}
		files, globError := filepath.Glob(filepath.Join(checkpoint, "source", "*.ts"))
		if globError != nil {
			return globError
		}
		for _, file := range files {
			data, readError := os.ReadFile(file)
			if readError != nil {
				return readError
			}
			if writeError := os.WriteFile(filepath.Join(dir, filepath.Base(file)), data, 0644); writeError != nil {
				return writeError
			}
		}
		return os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(program)), 0644)
	})
}

func syntaxMutantNativeProduct(test *testing.T, item syntaxMutant) string {
	test.Helper()
	lowered := syntaxMutantLoweredProduct(test, item)
	inputs := syntaxMutantProductInputs(test, item)
	inputs.Name = "estree-syntax-mutant-native-" + item.name
	inputs.Flags = append(inputs.Flags, native.Flags(estreeFamilyNativeOptions())...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return buildcache.Product(test, inputs, func(dir string) error {
		data, readError := os.ReadFile(filepath.Join(lowered, "port.c"))
		if readError != nil {
			return readError
		}
		return native.Build(string(data), filepath.Join(dir, "port"), estreeFamilyNativeOptions())
	})
}

func syntaxMutantProducts(test *testing.T, item syntaxMutant) (string, string) {
	test.Helper()
	lowered := syntaxMutantLoweredProduct(test, item)
	product := syntaxMutantNativeProduct(test, item)
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

const syntaxMutantsPreparedEnv = "ADAMIC_SYNTAX_MUTANTS_PREPARED"

func syntaxMutantSetupInputs(test *testing.T) buildcache.Inputs {
	inputs := buildcache.Inputs{
		Name:  "syntax-mutants-setup-v1",
		Files: estreeFamilyDependencyFiles(test),
		Flags: []string{
			"root=" + root(test),
			"sanitize=true",
			"split=true",
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
			"GOFLAGS=" + os.Getenv("GOFLAGS"),
		},
		Toolchain: []string{
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
			buildcache.Tool("go", "version"),
			buildcache.Tool("clang", "--version"),
		},
	}
	inputs.Flags = append(inputs.Flags, estreeFamilyDependencyFlags(test)...)
	return inputs
}

// Both the build-phase unit and shard preparation use this exact oracle recipe.
func syntaxMutantOracleProduct(test *testing.T) string {
	test.Helper()
	return oracles.SyntaxMutantsOracle(test)
}

func syntaxMutantPreparedProduct(test *testing.T, index int) string {
	test.Helper()
	item := syntaxMutantEnumeration()[index]
	inputs := syntaxMutantSetupInputs(test)
	// v3: Want names its sources' directory as <sources>, so a v2 product, which named a temporary directory, isn't
	// one this reads.
	inputs.Name = "syntax-mutant-ready-v3-" + item.name
	return buildcache.Product(test, inputs, func(dir string) error {
		main, binary := syntaxMutantProducts(test, item)
		oracle := filepath.Join(syntaxMutantOracleProduct(test), "oracle")
		sources := []string{item.source}
		if index == 0 {
			sources = syntaxGrammar()
		}
		list := manifest(test, sources)
		args := []string{"--manifest", list}
		if index != 0 {
			args = []string{"--audit", list, test.TempDir()}
		}
		// The oracle names each source by its path, and the sources sit in a temporary directory of this build, random
		// per build, so Want names that directory <sources> (landable-16c's tree, Oct 11: syntax-mutant-ready-v2
		// products differed build to build, the path hidden from relocatable in Want's base64). No reader compares the
		// paths.
		want := bytes.ReplaceAll(execute(test, "", oracle, args...), []byte(filepath.Dir(list)), []byte("<sources>"))
		// Main and Binary are other products, named by their keys (#tqrqx60).
		data, encodeError := json.Marshal(syntaxMutantPrepared{
			Main:    buildcache.Relative(main),
			Binary:  buildcache.Relative(binary),
			Sources: sources,
			Want:    want,
		})
		if encodeError != nil {
			return encodeError
		}
		return os.WriteFile(filepath.Join(dir, "ready.json"), data, 0644)
	})
}

func prepareSyntaxMutants(test *testing.T) {
	test.Helper()
	// The union's planted-survivor children are handed their parent's prepared mutants, so they fetch nothing.
	if descriptor := os.Getenv(syntaxMutantsPreparedEnv); descriptor != "" {
		data, readError := os.ReadFile(descriptor)
		if readError != nil {
			test.Fatal(readError)
		}
		if decodeError := json.Unmarshal(data, &syntaxMutantsPrepared); decodeError != nil {
			test.Fatal(decodeError)
		}
		return
	}
	for index := range syntaxMutantEnumeration() {
		dir := syntaxMutantPreparedProduct(test, index)
		data, readError := os.ReadFile(filepath.Join(dir, "ready.json"))
		if readError != nil {
			test.Fatal(readError)
		}
		var prepared syntaxMutantPrepared
		if decodeError := json.Unmarshal(data, &prepared); decodeError != nil {
			test.Fatal(decodeError)
		}
		prepared.Main, prepared.Binary = buildcache.Absolute(prepared.Main), buildcache.Absolute(prepared.Binary)
		prepared.Main = resolvedPort(test, filepath.Dir(prepared.Main), filepath.Base(prepared.Main))
		syntaxMutantsPrepared = append(syntaxMutantsPrepared, prepared)
	}
}
func syntaxMutantReady(test *testing.T) []syntaxMutantPrepared {
	test.Helper()
	syntaxMutantsPrepareOnce.Do(func() { prepareSyntaxMutants(test) })
	if len(syntaxMutantsPrepared) != len(syntaxMutantEnumeration()) {
		test.Fatal("syntax mutant preparation failed")
	}
	return syntaxMutantsPrepared
}

func TestProduct_EstreeFindsSyntaxMutantsGoOracle(test *testing.T) {
	test.Parallel()
	syntaxMutantOracleProduct(test)
}

func TestProduct_SyntaxMutantLowered_000(test *testing.T) {
	test.Parallel()
	syntaxMutantLoweredProduct(test, syntaxMutantEnumeration()[0])
}

func TestProduct_SyntaxMutantLowered_001(test *testing.T) {
	test.Parallel()
	syntaxMutantLoweredProduct(test, syntaxMutantEnumeration()[1])
}

func TestProduct_SyntaxMutantLowered_002(test *testing.T) {
	test.Parallel()
	syntaxMutantLoweredProduct(test, syntaxMutantEnumeration()[2])
}

func TestProduct_SyntaxMutantNative_001(test *testing.T) {
	test.Parallel()
	syntaxMutantNativeProduct(test, syntaxMutantEnumeration()[1])
}

func TestProduct_SyntaxMutantNative_002(test *testing.T) {
	test.Parallel()
	syntaxMutantNativeProduct(test, syntaxMutantEnumeration()[2])
}

func runSyntaxMutantShard(test *testing.T, shard int) {
	setup := time.Now()
	prepared := syntaxMutantReady(test)[shard]
	test.Logf("setup: %.3fs", time.Since(setup).Seconds())
	started := time.Now()
	defer func() { test.Logf("shard work: %.3fs", time.Since(started).Seconds()) }()
	cpu := syntaxMutantCPU()
	defer func() { test.Logf("CPU: %.3fs", (syntaxMutantCPU() - cpu).Seconds()) }()
	item := syntaxMutantEnumeration()[shard]
	list := manifest(test, prepared.Sources)
	if shard != 0 && !strings.Contains(string(prepared.Want), `"status":"error"`) {
		test.Fatal(string(prepared.Want))
	}
	for name, got := range map[string][]byte{
		"Node":   onNode(test, prepared.Main, "--manifest", list),
		"native": execute(test, "", prepared.Binary, "--manifest", list),
	} {
		if os.Getenv("ADAMIC_SYNTAX_MUTANTS_PLANT") == "1" && shard == 0 && name == "native" {
			wantCases := syntaxMutantCanonicalCases(test, prepared.Want, len(prepared.Sources))
			gotCases := syntaxMutantCanonicalCases(test, got, len(prepared.Sources))
			witness := -1
			for i, source := range prepared.Sources {
				if strings.Contains(source, "[P in A]") {
					if witness != -1 {
						test.Fatal("duplicate planted witness")
					}
					witness = i
				}
			}
			if witness < 0 {
				test.Fatal("missing planted witness")
			}
			if wantCases[witness] == gotCases[witness] {
				test.Fatal("planted witness already agrees")
			}
			gotCases[witness] = wantCases[witness]
			got = []byte(strings.Join(gotCases, ""))
		}
		if shard == 0 {
			if diff := firstDifference(prepared.Want, got); diff == "" {
				test.Fatal(name + " mutant survived")
			} else {
				test.Log(name + ": " + diff)
			}
		} else {
			if !strings.Contains(string(got), "0 Program ") {
				test.Fatal(name + " control did not accept")
			}
			test.Log(name + ": disabled check accepts Go-refused input; acceptance oracle catches it")
		}
	}
	test.Log(item.name)
}
func TestSyntaxMutants_000(test *testing.T) { test.Parallel(); runSyntaxMutantShard(test, 0) }
func TestSyntaxMutants_001(test *testing.T) { test.Parallel(); runSyntaxMutantShard(test, 1) }
func TestSyntaxMutants_002(test *testing.T) { test.Parallel(); runSyntaxMutantShard(test, 2) }

var syntaxMutantRunners = [...]func(*testing.T){TestSyntaxMutants_000, TestSyntaxMutants_001, TestSyntaxMutants_002}

func TestSyntaxMutantsUnion(test *testing.T) {
	test.Parallel()
	setup := time.Now()
	prepared := syntaxMutantReady(test)
	test.Logf("setup: %.3fs", time.Since(setup).Seconds())
	work := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(work).Seconds()) }()
	items := syntaxMutantEnumeration()
	if len(syntaxMutantRunners) != testSyntaxMutantsShards || len(items) != testSyntaxMutantsShards {
		test.Fatal("shard count mismatch")
	}
	seen := map[string]int{}
	count := 0
	for i, runner := range syntaxMutantRunners {
		name := runtime.FuncForPC(reflect.ValueOf(runner).Pointer()).Name()
		if !strings.HasSuffix(name, fmt.Sprintf(".TestSyntaxMutants_%03d", i)) {
			test.Fatalf("unexpected shard runner %s", name)
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
				test.Fatal("case coverage mismatch")
			}
		}
	}
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	encoded, encodeError := json.Marshal(prepared)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	descriptor := filepath.Join(test.TempDir(), "prepared.json")
	if writeError := os.WriteFile(descriptor, encoded, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	caught := []int{}
	for i := range syntaxMutantRunners {
		name := fmt.Sprintf("TestSyntaxMutants_%03d", i)
		command := exec.Command(executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_SYNTAX_MUTANTS_PLANT=1", syntaxMutantsPreparedEnv+"="+descriptor)
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil {
			if !strings.Contains(string(output), "mutant survived") ||
				!strings.Contains(string(output), "--- FAIL: "+name) {
				test.Fatalf("unexpected failure: %v\n%s", combinedOutputError, output)
			}
			caught = append(caught, i)
		}
	}
	if len(caught) != 1 || caught[0] != 0 {
		test.Fatalf("planted failure caught by %v", caught)
	}
	test.Logf("union: %d cases exactly once; planted survivor caught only by TestSyntaxMutants_000", count)
}

// Canonical manifests terminate each case with a stripped-source line. Source
// newlines are escaped, so boundaries preserve every original comparison byte.
func syntaxMutantCanonicalCases(test *testing.T, data []byte, count int) []string {
	test.Helper()
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
		test.Fatalf("canonical cases %d, want %d; trailing bytes %d", len(cases), count, current.Len())
	}
	return cases
}
func syntaxMutantCPU() time.Duration {
	var self, children syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &self) != nil ||
		syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children) != nil {
		return 0
	}
	return time.Duration(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec)*time.Second +
		time.Duration(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)*time.Microsecond
}
