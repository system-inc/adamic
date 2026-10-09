package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var runtimeFeatures = []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"}

func featureSource(mask int) string {
	var source strings.Builder
	for bit, feature := range runtimeFeatures {
		if mask&(1<<bit) != 0 {
			fmt.Fprintf(&source, "#define %s 1\n", feature)
		}
	}
	return source.String() + "#include \"adamic.h\"\n"
}

func featureSymbol(mask int) string {
	name := "adamic_runtime_features"
	for bit, feature := range runtimeFeatures {
		if mask&(1<<bit) != 0 {
			name += "_" + strings.ToLower(strings.TrimPrefix(feature, "ADAMIC_"))
		}
	}
	return name
}

func requireFeatureLinkError(t *testing.T, output []byte, err error, wanted string) {
	t.Helper()
	if err == nil {
		t.Fatal("mismatched runtime linked successfully")
	}
	text := string(output)
	undefined := strings.Contains(text, "undefined reference") || strings.Contains(text, "undefined symbol") || strings.Contains(text, "Undefined symbols")
	if !undefined || !regexp.MustCompile(regexp.QuoteMeta(wanted)+`(?:[^a-zA-Z0-9_]|$)`).MatchString(text) {
		t.Fatalf("want undefined feature symbol %s, got %v\n%s", wanted, err, output)
	}
	t.Logf("link rejected %s: %s", wanted, text)
}

func featureGCFlags() []string {
	if goruntime.GOOS == "darwin" {
		return []string{"-Wl,-dead_strip"}
	}
	return []string{"-Wl,--gc-sections"}
}

func compileFeatureUnit(t *testing.T, directory, name, source string, includes ...string) string {
	return compileFeatureUnitForTarget(t, directory, name, source, Options{}, includes...)
}

func compileFeatureUnitForTarget(t *testing.T, directory, name, source string, options Options, includes ...string) string {
	t.Helper()
	path := filepath.Join(directory, name+".c")
	object := filepath.Join(directory, name+".o")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	flags := append(Flags(options), "-ffunction-sections", "-fdata-sections", "-I", directory)
	for _, include := range includes {
		flags = append(flags, "-I", include)
	}
	flags = append(flags, "-c", path, "-o", object)
	if output, err := exec.Command(compilerName(options), flags...).CombinedOutput(); err != nil {
		t.Fatalf("compiling %s before the link check: %v\n%s", name, err, output)
	}
	return object
}

func featureHeaders(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return directory
}

// No sanitizer and no layout-dependent code: only the retained header reference can reject this.
func TestRuntimeFeatureMismatchProgramFeature(t *testing.T) {
	t.Parallel()
	testRuntimeFeatureMismatch(t, 1, 0)
}

func TestRuntimeFeatureMismatchRuntimeFeature(t *testing.T) {
	t.Parallel()
	testRuntimeFeatureMismatch(t, 0, 1)
}

func testRuntimeFeatureMismatch(t *testing.T, programMask, runtimeMask int) {
	t.Helper()
	library, err := RuntimeLibraryForSource("", featureSource(runtimeMask), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if runtimeMask == 0 {
		plain, err := RuntimeLibrary("", Options{})
		if err != nil || plain != library {
			t.Fatalf("plain RuntimeLibrary: %q %v", plain, err)
		}
	}
	object := compileFeatureUnit(t, t.TempDir(), "mismatch-probe", featureSource(programMask)+"int main(void) { return 0; }\n", filepath.Dir(library))
	binary := filepath.Join(t.TempDir(), "program")
	arguments := append(Flags(Options{}), featureGCFlags()...)
	arguments = append(arguments, object, "-o", binary)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	output, err := exec.Command("clang", arguments...).CombinedOutput()
	requireFeatureLinkError(t, output, err, featureSymbol(programMask))
	if err := Build(featureSource(programMask)+"int main(void) { return 0; }\n", binary, Options{}); err != nil {
		t.Fatalf("matching runtime: %v", err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil || len(output) != 0 {
		t.Fatalf("matching program: %v %q", err, output)
	}
}

// Every bit participates, both present and absent, even after section removal. Small definition
// objects keep this exhaustive link matrix independent of runtime work unrelated to the check.
func TestRuntimeFeatureSetsNative(t *testing.T) {
	t.Parallel()
	testRuntimeFeatureSets(t, Options{})
}

func TestRuntimeFeatureSetsWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("set WASI_SYSROOT to the WASI SDK sysroot")
	}
	testRuntimeFeatureSets(t, Options{Target: "wasm32-wasi"})
}

func testRuntimeFeatureSets(t *testing.T, options Options) {
	directory := featureHeaders(t)
	definition, err := runtime.ReadFile("runtime/features.c")
	if err != nil {
		t.Fatal(err)
	}
	count := 1 << len(runtimeFeatures)
	programs, definitions := make([]string, count), make([]string, count)
	for mask := range count {
		programs[mask] = compileFeatureUnitForTarget(t, directory, fmt.Sprintf("program-%d", mask), featureSource(mask)+"int main(void) { return 0; }\n", options)
		definitions[mask] = compileFeatureUnitForTarget(t, directory, fmt.Sprintf("runtime-%d", mask), featureSource(mask)+string(definition), options)
	}
	for mask := range count {
		for bit := -1; bit < len(runtimeFeatures); bit++ {
			actual := mask
			if bit >= 0 {
				actual ^= 1 << bit
			}
			arguments := append(Flags(options), featureGCFlags()...)
			arguments = append(arguments, programs[mask], definitions[actual], "-o", filepath.Join(directory, "matrix"))
			output, err := exec.Command(compilerName(options), arguments...).CombinedOutput()
			if bit < 0 {
				if err != nil {
					t.Fatalf("matching feature set %d: %v\n%s", mask, err, output)
				}
			} else {
				requireFeatureLinkError(t, output, err, featureSymbol(mask))
			}
		}
	}
	// An unused unit compiled with another set must still reject an otherwise matching program.
	unused := compileFeatureUnitForTarget(t, directory, "unused", featureSource(0)+"int unused_client(void) { return 0; }\n", options)
	arguments := append(Flags(options), featureGCFlags()...)
	arguments = append(arguments, programs[1], definitions[1], unused, "-o", filepath.Join(directory, "mixed"))
	output, err := exec.Command(compilerName(options), arguments...).CombinedOutput()
	requireFeatureLinkError(t, output, err, featureSymbol(0))
	t.Log("32 matching sets, 160 one-bit mismatches, and an unused mixed unit checked")
}

// The overlay mode reproduces split-build-flags' third mutant without sanitizers. The ordinary
// mode remains a Node comparison; neither mode executes a mismatched binary.
func TestSplitTSGoRuntimeFeatures(t *testing.T) {
	t.Parallel()
	archive := os.Getenv("ADAMIC_CLANG_TSGO_ARCHIVE")
	if archive == "" {
		t.Skip("set ADAMIC_CLANG_TSGO_ARCHIVE to a built checker archive")
	}
	loaded, err := load.Load([]string{"../oracle/testdata/arguments_length_extended.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	source, err := TSGoC(program)
	if err != nil {
		t.Fatal(err)
	}
	if !slicesContain(featureFlags(source), "-DADAMIC_CLOSURE_CONVENTION=1") {
		t.Fatal("the witness no longer enables the closure convention")
	}
	binary := filepath.Join(t.TempDir(), "program")
	err = BuildSplitTSGo(source, binary, archive, Options{Jobs: 5})
	if os.Getenv("ADAMIC_TEST_FEATURE_RUNTIME_MISMATCH") == "1" {
		mask := 0
		for bit, feature := range runtimeFeatures {
			if slicesContain(featureFlags(source), "-D"+feature+"=1") {
				mask |= 1 << bit
			}
		}
		if err == nil {
			t.Fatal("mismatched checker runtime linked successfully")
		}
		requireFeatureLinkError(t, []byte(err.Error()), err, featureSymbol(mask))
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", "../oracle/testdata/arguments_length_extended.a")
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("split checker stdout differs: native %q Node %q", got, want)
	}
}
