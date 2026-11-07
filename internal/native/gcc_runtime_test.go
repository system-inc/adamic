package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/buildcache"
)

// Every translation unit is a separate product and test, including optional
// runtime features. The inventory test makes adding a C file require an owner.
func gccRuntimeUnit(test *testing.T, name string) string {
	return gccRuntimeProduct(test, name, "")
}

func gccRuntimeProduct(test *testing.T, name, mutant string) string {
	test.Helper()
	compiler := gccAuditCompiler(test)
	files, flags, _, _, gccAuditRuntimeInputsError := gccAuditRuntimeInputs(Options{Count: true, ProgramRegion: true})
	if gccAuditRuntimeInputsError != nil {
		test.Fatal(gccAuditRuntimeInputsError)
	}
	flags = slices.DeleteFunc(flags, func(s string) bool {
		return s == "-Wcast-function-type-strict" ||
			s == "-Wno-self-assign" ||
			s == "-Werror" ||
			strings.HasPrefix(s, "-Werror=")
	})
	flags = append(
		flags,
		"-Werror=array-bounds",
		"-Werror=maybe-uninitialized",
		"-Werror=stringop-overflow",
		"-Werror=infinite-recursion",
	)
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: "gcc-runtime-" + name + mutant,
		Files: []string{
			"go.mod",
			"internal/native/runtime",
			"bridge/tsgo/tsgo.h",
			"internal/native/gcc_runtime_test.go",
			"internal/native/gcc_units_test.go",
			"internal/native/gcc_invariants_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/target.go",
		},
		Flags:     portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{buildcache.Tool(compiler, "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		found := false
		for _, file := range files {
			if file.name == name {
				found = true
				if mutant != "" {
					text := string(file.contents)
					before, after := "", ""
					if mutant == "recursion" {
						before = "double adamic_math_sin(double x) {"
						after = before + "\n return adamic_math_sin(x);"
					} else {
						spec := gccInvariants[mutant]
						before = spec.guard
						after = "/* guard removed */"
					}
					if strings.Count(text, before) != 1 {
						return fmt.Errorf("GCC mutant seam missing: %s", mutant)
					}
					file.contents = []byte(strings.Replace(text, before, after, 1))
				}
			}
			if strings.HasSuffix(file.name, ".h") || file.name == name {
				if writeError := os.WriteFile(
					filepath.Join(directory, file.name),
					file.contents,
					0600,
				); writeError != nil {
					return writeError
				}
			}
		}
		if !found {
			return fmt.Errorf("runtime unit disappeared: %s", name)
		}
		args := append(
			append(slices.Clone(flags), prefixMaps(directory, "/adamic-runtime")...),
			"-c",
			name,
			"-o",
			"unit.o",
		)
		command, release := boundedrun.Command(45*time.Second, compiler, args...)
		defer release()
		command.Dir = directory
		output, combinedOutputError := command.CombinedOutput()
		if mutant == "" {
			if combinedOutputError != nil {
				return fmt.Errorf("GCC %s: %w\n%s", name, combinedOutputError, output)
			}
		} else {
			diagnostic := gccMutantDiagnostic(mutant)
			if combinedOutputError == nil || !strings.Contains(string(output), diagnostic) {
				return fmt.Errorf("GCC mutant %s survived: %v\n%s", mutant, combinedOutputError, output)
			}
		}
		return os.WriteFile(filepath.Join(directory, "diagnostics"), output, 0600)
	})
}

func checkGCCRuntimeUnit(test *testing.T, name string) {
	test.Helper()
	directory := gccRuntimeUnit(test, name)
	diagnostics, readError := os.ReadFile(filepath.Join(directory, "diagnostics"))
	if readError != nil {
		test.Fatal(readError)
	}
	if len(diagnostics) != 0 {
		test.Logf("GCC advisory warnings for %s:\n%s", name, diagnostics)
	}
}
func TestGCCRuntimeInventory(test *testing.T) {
	test.Parallel()
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		test.Fatal(readRuntimeError)
	}
	var names []string
	for _, file := range files {
		if strings.HasSuffix(file.name, ".c") {
			names = append(names, file.name)
		}
	}
	if !slices.Equal(names, gccRuntimeUnits) {
		test.Fatalf(
			"give each runtime C file its own audit product and test: got %v; pinned %v",
			names,
			gccRuntimeUnits,
		)
	}
}

func gccMutantDiagnostic(mutant string) string {
	switch mutant {
	case "DTOA":
		return "-Werror=array-bounds"
	case "Reduction":
		return "-Werror=maybe-uninitialized"
	case "Map":
		return "-Werror=stringop-overflow"
	case "recursion":
		return "-Werror=infinite-recursion"
	}
	panic(mutant)
}
func checkGCCDiagnosticMutant(test *testing.T, name, mutant string) {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join(gccRuntimeProduct(test, name, mutant), "diagnostics"))
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.Contains(string(data), gccMutantDiagnostic(mutant)) {
		test.Fatalf("mutant not caught: %s", data)
	}
	test.Logf("%s rejected: %s", mutant, gccMutantDiagnostic(mutant))
}
func TestProduct_GCCDiagnosticDTOAMutant(test *testing.T) {
	test.Parallel()
	gccRuntimeProduct(test, "dtoa.c", "DTOA")
}
func TestProduct_GCCDiagnosticReductionMutant(test *testing.T) {
	test.Parallel()
	gccRuntimeProduct(test, "ieee754.c", "Reduction")
}
func TestProduct_GCCDiagnosticMapMutant(test *testing.T) {
	test.Parallel()
	gccRuntimeProduct(test, "map.c", "Map")
}
func TestProduct_GCCDiagnosticRecursionMutant(test *testing.T) {
	test.Parallel()
	gccRuntimeProduct(test, "ieee754.c", "recursion")
}
func TestGCCDiagnosticDTOAMutant(test *testing.T) {
	test.Parallel()
	checkGCCDiagnosticMutant(test, "dtoa.c", "DTOA")
}
func TestGCCDiagnosticReductionMutant(test *testing.T) {
	test.Parallel()
	checkGCCDiagnosticMutant(test, "ieee754.c", "Reduction")
}
func TestGCCDiagnosticMapMutant(test *testing.T) {
	test.Parallel()
	checkGCCDiagnosticMutant(test, "map.c", "Map")
}
func TestGCCDiagnosticRecursionMutant(test *testing.T) {
	test.Parallel()
	checkGCCDiagnosticMutant(test, "ieee754.c", "recursion")
}

// Match the existing portability tests' executable lookup and version query,
// and require GNU GCC: clang under the gcc name is not this audit.
func gccAuditCompiler(test *testing.T) string {
	test.Helper()
	if goruntime.GOOS != "linux" {
		test.Skip("GNU gcc audits run on Linux; macOS gcc is clang")
	}
	compiler, lookPathError := exec.LookPath("gcc")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	command, release := boundedrun.Command(5*time.Second, compiler, "--version")
	defer release()
	version, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil || !strings.Contains(string(version), "Free Software Foundation") {
		test.Fatalf("GNU GCC required: %v\n%s", combinedOutputError, version)
	}
	return compiler
}

// Cover compiler-only regex and checker bridge bodies as well as emitter features.
// The bridge ABI header is keyed and copied with the runtime, just as BuildTSGo does.
func gccAuditRuntimeInputs(options Options) ([]runtimeFile, []string, string, string, error) {
	var features []string
	for _, feature := range runtimeFeatures {
		features = append(features, "-D"+feature+"=1")
	}
	features = append(features, "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-DADAMIC_TSGO=1")
	files, flags, compiler, version, runtimeInputsError := runtimeInputs("", options, features)
	if runtimeInputsError != nil {
		return nil, nil, "", "", runtimeInputsError
	}
	header, readError := os.ReadFile("../../bridge/tsgo/tsgo.h")
	if readError != nil {
		return nil, nil, "", "", readError
	}
	files = append(files, runtimeFile{"tsgo.h", header})
	return files, flags, compiler, version, nil
}
