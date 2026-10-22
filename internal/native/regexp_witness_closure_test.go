package native

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

var regexpClosureMutant = flag.String("regexp-witness-closure-mutant", "", "throwaway RegExp witness closure mutation")

// Undefined weak globals remain uses: a future optional runtime reference must
// not hide a dependency on an omitted member. Common and weak definitions count.
type regexpMemberSymbols struct {
	Defined   []string
	Undefined []string
}

type regexpClosureInventory struct {
	Members     map[string]regexpMemberSymbols
	Sources     []string
	External    []string
	MutatedWeak regexpMemberSymbols
}

// nm is already declared on loom/planner-reads's tools.txt. Keep archive member
// identity (-A) and global symbols (-g), including data, TLS, weak and common.
func regexpGlobalSymbols(output string, archive bool) (map[string]regexpMemberSymbols, error) {
	members := map[string]regexpMemberSymbols{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		member := "external"
		if archive {
			start, end := strings.LastIndex(fields[0], "["), strings.LastIndex(fields[0], "]:")
			if start < 0 || end < start {
				return nil, fmt.Errorf("nm lost archive member identity: %s", line)
			}
			member = strings.TrimSuffix(fields[0][start+1:end], ".o") + ".c"
			fields = fields[1:]
		}
		if len(fields) < 2 || len(fields[1]) != 1 {
			return nil, fmt.Errorf("unrecognized nm global: %s", line)
		}
		name := strings.SplitN(fields[0], "@", 2)[0]
		symbols := members[member]
		if fields[1] == "U" || fields[1] == "w" || fields[1] == "v" {
			symbols.Undefined = append(symbols.Undefined, name)
		} else {
			symbols.Defined = append(symbols.Defined, name)
		}
		members[member] = symbols
	}
	return members, nil
}

func regexpClosureTool(t testing.TB, tool string, args ...string) string {
	t.Helper()
	output, err := toolchain.Command(tool, args...).Output()
	if err != nil {
		t.Fatalf("%s %v: %v", tool, args, err)
	}
	return strings.TrimSpace(string(output))
}

// Discover actual system definitions, never a guessed whitelist of C function
// names. libc_nonshared is part of glibc's libc.so linker script (e.g. atexit).
// The ASan/UBSan archives provide the instrumentation the witnesses request;
// they are compiler support, not a way to forgive unresolved runtime symbols.
func regexpClosureLibraries(t testing.TB, compiler string) []string {
	t.Helper()
	var files []string
	for _, name := range []string{"libc.so.6", "libc_nonshared.a", "libm.so.6", "libpthread.so.0", "libclang_rt.asan.a", "libclang_rt.ubsan_standalone.a"} {
		path := regexpClosureTool(t, compiler, "-print-file-name="+name)
		if !filepath.IsAbs(path) {
			t.Fatalf("clang cannot locate symbol library %s: %s", name, path)
		}
		files = append(files, path)
	}
	return files
}

// This product only compiles omitted members and the throwaway changed weak.c.
// Listed objects come from the two existing witness products. Product-order
// explicitly waits for both, so it cannot quietly become their second builder.
func regexpWitnessClosureInventory(t testing.TB) string {
	t.Helper()
	if goruntime.GOOS != "linux" {
		t.Skip("RegExp witness symbol closure is held on Linux")
	}
	options := Options{Sanitize: true}
	files, flags, compiler, version, err := runtimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	request, err := runtimeRequest(files, flags, compiler, version, options)
	if err != nil {
		t.Fatal(err)
	}
	libraries := regexpClosureLibraries(t, compiler)
	inputs := buildcache.Inputs{
		Name:      "regexp-witness-closure-inventory",
		Files:     []string{"internal/native/regexp_witness_closure_test.go", "internal/native/regexp_test.go", "internal/native/library.go", "internal/native/jobserver.go", "internal/native/native.go", "internal/native/target.go", "internal/native/runtime"},
		Flags:     slices.Clone(request.Portable),
		Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("nm", "--version")},
	}
	for _, path := range libraries {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("symbol library %s %x", filepath.Base(path), sha256.Sum256(data)))
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		inventory := regexpClosureInventory{Members: map[string]regexpMemberSymbols{}}
		scanArchive := func(path string) map[string]regexpMemberSymbols {
			t.Helper()
			symbols, err := regexpGlobalSymbols(regexpClosureTool(t, "nm", "-A", "-g", "-P", path), true)
			if err != nil {
				t.Fatal(err)
			}
			// nm prints no rows for a genuinely empty object. Keep its member
			// identity too, so an empty C unit is not mistaken for a stale name.
			for _, object := range strings.Fields(regexpClosureTool(t, request.Archiver, "-t", path)) {
				name := strings.TrimSuffix(object, ".o") + ".c"
				if _, exists := symbols[name]; !exists {
					symbols[name] = regexpMemberSymbols{}
				}
			}
			return symbols
		}
		for part := range 2 {
			archive := regexWitnessRuntime(t, regexWitnessRequest(t, options, part), part)
			for name, symbols := range scanArchive(filepath.Join(archive, "runtime.a")) {
				inventory.Members[name] = symbols
			}
		}
		all, err := readRuntime(runtime, "runtime")
		if err != nil {
			return err
		}
		for _, file := range all {
			if strings.HasSuffix(file.name, ".c") {
				inventory.Sources = append(inventory.Sources, file.name)
			}
		}
		var omitted, mutated []runtimeFile
		for _, file := range files {
			if strings.HasSuffix(file.name, ".h") {
				omitted = append(omitted, file)
				mutated = append(mutated, file)
			}
			if strings.HasSuffix(file.name, ".c") && !regexWitnessRuntimeSources[file.name] {
				omitted = append(omitted, file)
			}
			if file.name == "weak.c" {
				// New external call, compiled from a copy only. normalize.c is not listed.
				changed := slices.Clone(file.contents)
				changed = append(changed, []byte("\nadamic_string *adamic_regexp_closure_probe(const adamic_string *s) { return adamic_string_normalize(s, s); }\n")...)
				mutated = append(mutated, runtimeFile{file.name, changed})
			}
		}
		build := func(subdirectory string, snapshot []runtimeFile) (string, error) {
			request, err := runtimeRequest(snapshot, flags, compiler, version, options)
			if err != nil {
				return "", err
			}
			root := filepath.Join(directory, subdirectory)
			if err := os.MkdirAll(root, 0700); err != nil {
				return "", err
			}
			if err := request.Build(root, 4, nil); err != nil {
				return "", err
			}
			return filepath.Join(root, "runtime.a"), nil
		}
		if slices.ContainsFunc(omitted, func(file runtimeFile) bool { return strings.HasSuffix(file.name, ".c") }) {
			archive, err := build("omitted", omitted)
			if err != nil {
				return err
			}
			for name, symbols := range scanArchive(archive) {
				inventory.Members[name] = symbols
			}
		}
		archive, err := build("call-mutant", mutated)
		if err != nil {
			return err
		}
		inventory.MutatedWeak = scanArchive(archive)["weak.c"]
		var librarySymbols []map[string]regexpMemberSymbols
		for _, path := range libraries {
			args := []string{"-g", "-P", "--defined-only", path}
			if strings.HasSuffix(path, ".so.6") || strings.HasSuffix(path, ".so.0") {
				args = append([]string{"-D"}, args...)
			}
			// -A gives static archives member labels, unlike dynamic shared libraries.
			archive := strings.HasSuffix(path, ".a")
			if archive {
				args = append([]string{"-A"}, args...)
			}
			symbols, err := regexpGlobalSymbols(regexpClosureTool(t, "nm", args...), archive)
			if err != nil {
				return err
			}
			librarySymbols = append(librarySymbols, symbols)
		}
		data, err := regexpWitnessClosureBytes(inventory, librarySymbols)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "symbols.json"), data, 0600)
	})
}

// Keep serialization independent of map iteration order, so one inventory key
// always makes the same symbols.json bytes (#6r7v4ye).
func regexpWitnessClosureBytes(inventory regexpClosureInventory, libraries []map[string]regexpMemberSymbols) ([]byte, error) {
	for _, library := range libraries {
		for _, member := range library {
			inventory.External = append(inventory.External, member.Defined...)
		}
	}
	// GNU ld synthesizes these ASan boundaries, rather than runtime C sources.
	inventory.External = append(inventory.External, "__start_asan_globals", "__stop_asan_globals")
	slices.Sort(inventory.External)
	inventory.External = slices.Compact(inventory.External)
	return json.Marshal(inventory)
}

func TestRegExpWitnessClosureBytesDeterministic(t *testing.T) {
	t.Parallel()
	libraries := []map[string]regexpMemberSymbols{{}, {}}
	for member := range 8 {
		symbols := regexpMemberSymbols{Undefined: []string{"not_a_definition"}}
		for symbol := range 8 {
			symbols.Defined = append(symbols.Defined, fmt.Sprintf("symbol_%02d", member*8+symbol))
		}
		// Real libraries can define the same symbol in several members.
		symbols.Defined = append(symbols.Defined, "symbol_00")
		libraries[member/4][fmt.Sprintf("member_%d.o", member)] = symbols
	}
	inventory := regexpClosureInventory{
		Members: map[string]regexpMemberSymbols{"weak.c": {Undefined: []string{"symbol_00"}}},
		Sources: []string{"weak.c"},
	}
	var first []byte
	for run := range 20 {
		data, err := regexpWitnessClosureBytes(inventory, libraries)
		if err != nil {
			t.Fatal(err)
		}
		if run == 0 {
			first = data
		} else if !bytes.Equal(first, data) {
			t.Fatalf("symbols.json bytes differ between serialization 1 and %d", run+1)
		}
	}
	var decoded regexpClosureInventory
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.External) != 66 || !slices.IsSorted(decoded.External) || slices.Contains(decoded.External, "not_a_definition") {
		t.Fatalf("expected 64 unique definitions and two ASan boundaries, sorted: %v", decoded.External)
	}
	if !slices.Equal(decoded.Sources, inventory.Sources) || !slices.Equal(decoded.Members["weak.c"].Undefined, inventory.Members["weak.c"].Undefined) {
		t.Fatal("serialization lost runtime inventory fields")
	}
}

func TestProduct_RegExpWitnessClosureInventory(t *testing.T) {
	t.Parallel()
	regexpWitnessClosureInventory(t)
}

func regexpWitnessClosureGaps(selected map[string]bool, inventory regexpClosureInventory) []string {
	providers := map[string][]string{}
	for name, symbols := range inventory.Members {
		for _, symbol := range symbols.Defined {
			providers[symbol] = append(providers[symbol], name)
		}
	}
	var gaps []string
	for name, included := range selected {
		if !included {
			continue
		}
		if !slices.Contains(inventory.Sources, name) {
			gaps = append(gaps, fmt.Sprintf("%s is listed in regexWitnessRuntimeSources but no runtime source defines it", name))
			continue
		}
		symbols, ok := inventory.Members[name]
		if !ok {
			gaps = append(gaps, fmt.Sprintf("%s has no object under regexWitnessRequest's options", name))
			continue
		}
		for _, symbol := range symbols.Undefined {
			owners := providers[symbol]
			if slices.ContainsFunc(owners, func(owner string) bool { return selected[owner] }) {
				continue
			}
			if slices.Contains(inventory.External, symbol) {
				continue
			}
			if len(owners) != 0 {
				slices.Sort(owners)
				gaps = append(gaps, fmt.Sprintf("%s calls %s, defined in %s, which regexWitnessRuntimeSources leaves out", name, symbol, strings.Join(owners, ", ")))
			} else {
				gaps = append(gaps, fmt.Sprintf("%s calls %s, which no runtime source or linked system library defines", name, symbol))
			}
		}
	}
	slices.Sort(gaps)
	return slices.Compact(gaps)
}

func TestRegExpWitnessRuntimeClosure(t *testing.T) {
	t.Parallel()
	directory := regexpWitnessClosureInventory(t)
	data, err := os.ReadFile(filepath.Join(directory, "symbols.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory regexpClosureInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{}
	for name, included := range regexWitnessRuntimeSources {
		selected[name] = included
	}
	switch *regexpClosureMutant {
	case "":
	case "drop-dtoa":
		delete(selected, "dtoa.c")
	case "call-normalize":
		inventory.Members["weak.c"] = inventory.MutatedWeak
	default:
		t.Fatalf("unknown closure mutant %q", *regexpClosureMutant)
	}
	for _, gap := range regexpWitnessClosureGaps(selected, inventory) {
		t.Error(gap)
	}
	t.Logf("checked %d listed sources against %d runtime object members", len(selected), len(inventory.Members))
}

func regexpWitnessClosureMutant(t *testing.T, name string) {
	t.Helper()
	if goruntime.GOOS != "linux" {
		t.Skip("RegExp witness symbol closure is held on Linux")
	}
	data, err := os.ReadFile("testdata/regexp-witness-closure-mutants.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Name, Test, Diagnostic string }
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, one := range cases {
		if one.Name != name {
			continue
		}
		command := exec.Command(os.Args[0], "-test.run=^"+one.Test+"$", "-test.count=1", "-test.timeout=15s", "-test.v", "-regexp-witness-closure-mutant="+name)
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), one.Diagnostic) || !strings.Contains(string(output), "--- FAIL: "+one.Test) {
			t.Fatalf("closure mutant %s must fail its own test by name: %v\n%s", name, err, output)
		}
		t.Logf("%s", output)
		return
	}
	t.Fatalf("no closure mutant evidence for %s", name)
}

func TestRegExpWitnessClosureDropDTOAMutant(t *testing.T) {
	t.Parallel()
	regexpWitnessClosureMutant(t, "drop-dtoa")
}

func TestRegExpWitnessClosureNewCallMutant(t *testing.T) {
	t.Parallel()
	regexpWitnessClosureMutant(t, "call-normalize")
}

func TestRegExpWitnessClosureStaleSource(t *testing.T) {
	t.Parallel()
	gaps := regexpWitnessClosureGaps(map[string]bool{"stale.c": true}, regexpClosureInventory{})
	if len(gaps) != 1 || gaps[0] != "stale.c is listed in regexWitnessRuntimeSources but no runtime source defines it" {
		t.Fatalf("stale source survived: %v", gaps)
	}
}

func TestRegExpWitnessGlobalSymbols(t *testing.T) {
	t.Parallel()
	symbols, err := regexpGlobalSymbols("/cache/runtime.a[weak.o]: global T 0 10\n/cache/runtime.a[weak.o]: shared C 8 8\n/cache/runtime.a[weak.o]: call U\n/cache/runtime.a[weak.o]: optional w\n", true)
	if err != nil {
		t.Fatal(err)
	}
	weak := symbols["weak.c"]
	if len(symbols) != 1 || strings.Join(weak.Defined, ",") != "global,shared" || strings.Join(weak.Undefined, ",") != "call,optional" {
		t.Fatalf("lost globals or member provenance: %+v", symbols)
	}
	dynamic, err := regexpGlobalSymbols("malloc@@GLIBC_2.2.5 T 0 10\n", false)
	if err != nil || strings.Join(dynamic["external"].Defined, ",") != "malloc" {
		t.Fatalf("lost versioned definition: %v %+v", err, dynamic)
	}
	if _, err := regexpGlobalSymbols("global T 0 10\n", true); err == nil {
		t.Fatal("accepted globals without archive member provenance")
	}
}
