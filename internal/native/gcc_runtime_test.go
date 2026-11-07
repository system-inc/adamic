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
	files, flags, _, _, err := gccAuditRuntimeInputs(Options{Count: true, ProgramRegion: true})
	if err != nil {
		test.Fatal(err)
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
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		test.Fatal(err)
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
				if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0600); err != nil {
					return err
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
		output, err := command.CombinedOutput()
		if mutant == "" {
			if err != nil {
				return fmt.Errorf("GCC %s: %w\n%s", name, err, output)
			}
		} else {
			diagnostic := gccMutantDiagnostic(mutant)
			if err == nil || !strings.Contains(string(output), diagnostic) {
				return fmt.Errorf("GCC mutant %s survived: %v\n%s", mutant, err, output)
			}
		}
		return os.WriteFile(filepath.Join(directory, "diagnostics"), output, 0600)
	})
}

func checkGCCRuntimeUnit(test *testing.T, name string) {
	test.Helper()
	directory := gccRuntimeUnit(test, name)
	diagnostics, err := os.ReadFile(filepath.Join(directory, "diagnostics"))
	if err != nil {
		test.Fatal(err)
	}
	if len(diagnostics) != 0 {
		test.Logf("GCC advisory warnings for %s:\n%s", name, diagnostics)
	}
}
func TestGCCRuntimeInventory(test *testing.T) {
	test.Parallel()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		test.Fatal(err)
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
	data, err := os.ReadFile(filepath.Join(gccRuntimeProduct(test, name, mutant), "diagnostics"))
	if err != nil {
		test.Fatal(err)
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
	compiler, err := exec.LookPath("gcc")
	if err != nil {
		test.Fatal(err)
	}
	command, release := boundedrun.Command(5*time.Second, compiler, "--version")
	defer release()
	version, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(version), "Free Software Foundation") {
		test.Fatalf("GNU GCC required: %v\n%s", err, version)
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
	files, flags, compiler, version, err := runtimeInputs("", options, features)
	if err != nil {
		return nil, nil, "", "", err
	}
	header, err := os.ReadFile("../../bridge/tsgo/tsgo.h")
	if err != nil {
		return nil, nil, "", "", err
	}
	files = append(files, runtimeFile{"tsgo.h", header})
	return files, flags, compiler, version, nil
}
