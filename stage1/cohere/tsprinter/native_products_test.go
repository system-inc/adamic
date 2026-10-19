package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/native"
)

// Every sanitized build the closed gaps' tests, TestClassInterfaceMethodGap's workaround and TestNumberConstructor run
// is a keyed product built ahead by a TestProduct_ of its own, so none of them compiles C (#tze35sy, Loom's build law):
// four clang builds were made inside the tests.

func TestProduct_TSPrinterGapsSanitized(t *testing.T) {
	t.Parallel()
	for _, source := range []sanitizedSource{prefixUpdateValueSource, optionalBooleanFunctionSource, classInterfaceWorkaroundSource} {
		requirePort(t, sourceSanitized(t, source))
	}
}

func TestProduct_TSPrinterNumberConstructorSanitized(t *testing.T) {
	t.Parallel()
	requirePort(t, sourceSanitized(t, numberConstructorSource))
}

func requirePort(t *testing.T, directory string) {
	t.Helper()
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port"})
}

// A sanitizedSource is a program a test runs natively: the file at path, relative to this package, written as entry
// (load reads a .a apart from a .ts) with each replacement made in order.
type sanitizedSource struct {
	path         string
	entry        string
	replacements []replacement
}

// A replacement takes the from text, which must stand exactly once, to to.
type replacement struct{ from, to string }

var (
	prefixUpdateValueSource       = sanitizedSource{path: "gaps/prefixUpdateValue.ts", entry: "prefixUpdateValue.ts"}
	optionalBooleanFunctionSource = sanitizedSource{path: "gaps/optionalBooleanFunction.ts.txt", entry: "gap.a"}
	numberConstructorSource       = sanitizedSource{path: "testdata/numberConstructor.ts", entry: "numberConstructor.ts"}
	// The gap's class method made a safe explicit callback, the normal run its gap check must reject.
	classInterfaceWorkaroundSource = sanitizedSource{path: "gaps/classInterfaceMethod.ts.txt", entry: "workaround.ts", replacements: []replacement{
		{"read(): number", "readValue(): number"},
		{"console.log(`${new Consumer(new Box()).run()}`);", "const box = new Box();\nconsole.log(`${new Consumer({read: () => box.readValue()}).run()}`);"},
	}}
)

// rewritten is text with each replacement made in order, refusing one that doesn't change exactly one place.
func rewritten(text string, replacements []replacement) (string, error) {
	for _, made := range replacements {
		if strings.Count(text, made.from) != 1 {
			return "", fmt.Errorf("the replacement must change exactly one place: %q", made.from)
		}
		text = strings.Replace(text, made.from, made.to, 1)
	}
	return text, nil
}

// sourceSanitized is source written into the product, lowered in this process and built by clang under the
// sanitizers, keyed by the compiler's paths, the source file, its entry name, the replacements and the native build.
func sourceSanitized(t *testing.T, source sanitizedSource) string {
	t.Helper()
	inputs := sanitizedInputs("tsprinter "+source.path+" as "+source.entry+" native sanitized", "stage1/cohere/tsprinter/"+source.path)
	inputs.Flags = append(inputs.Flags, "entry="+source.entry)
	for _, made := range source.replacements {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("replacement=%q=>%q", made.from, made.to))
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		contents, err := os.ReadFile(source.path)
		if err != nil {
			return err
		}
		text, err := rewritten(string(contents), source.replacements)
		if err != nil {
			return err
		}
		sources := filepath.Join(directory, "source")
		if err := os.Mkdir(sources, 0o755); err != nil {
			return err
		}
		entry := filepath.Join(sources, source.entry)
		if err := os.WriteFile(entry, []byte(text), 0o644); err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// sanitizedInputs keys a sanitized build lowered in this process: the compiler is linked into this test binary, where
// no trace sees it read, so its paths are named (compilerinputs.Lowering) beside the sources and this file, with the
// native options and flags, the environment native.Build and clang read, and clang's version.
func sanitizedInputs(name string, sources ...string) buildcache.Inputs {
	options := native.Options{Sanitize: true}
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	files := append(compilerinputs.Lowering(), sources...)
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/tsprinter/native_products_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// buildSanitized lowers the program at entry and builds it under the sanitizers as directory/port.
func buildSanitized(entry, directory string) error {
	program, err := lowerProgram(entry)
	if err != nil {
		return err
	}
	return native.Build(native.C(program), filepath.Join(directory, "port"), native.Options{Sanitize: true})
}

// sanitizedRun runs a sanitized build with leak detection off, as the oracle does; leaks is its own run.
func sanitizedRun(t *testing.T, binary string, arguments ...string) oracles.Run {
	t.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...)
}
