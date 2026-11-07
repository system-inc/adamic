package yaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every program a test of this package lowers or compiles is a product, built ahead by its TestProduct_ twin, so a
// test only reads it and no test's clock holds a lowering or a clang run (#9wyp4tt). A build is lowered once into
// <stem>.c and <stem>.mjs, then compiled by clang into <stem> with its options; a mutant is its mutated sources, their
// emitted JavaScript and their native binary in one product.

// yamlPortInputs are a build's declared inputs: the compiler (compilerinputs.Lowering), this package's TypeScript and
// gaps/ programs, and this file, which holds the recipe.
func yamlPortInputs(test *testing.T, name string) buildcache.Inputs {
	test.Helper()
	files := append(compilerinputs.Lowering(), "stage1/cohere/yaml/port_products_test.go")
	for _, pattern := range []string{"*.ts", "gaps/*.ts"} {
		sources, globError := filepath.Glob(pattern)
		if globError != nil {
			test.Fatal(globError)
		}
		for _, source := range sources {
			files = append(files, "stage1/cohere/yaml/"+source)
		}
	}
	return buildcache.Inputs{
		Name:      name,
		Files:     files,
		Toolchain: []string{runtime.Version()},
		Flags:     []string{"GOFLAGS=" + os.Getenv("GOFLAGS"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH},
	}
}

type yamlProductValue struct {
	once      sync.Once
	directory string
}

var yamlProducts sync.Map

// yamlProduct runs build once per process for each key, and fails every later caller when that build failed.
func yamlProduct(test *testing.T, key string, build func() string) string {
	test.Helper()
	stored, _ := yamlProducts.LoadOrStore(key, &yamlProductValue{})
	value := stored.(*yamlProductValue)
	value.once.Do(func() { value.directory = build() })
	if value.directory == "" {
		test.Fatalf("product %s failed earlier in this process", key)
	}
	return value.directory
}

// yamlBuild is entries of this package (paths relative to it) lowered together, and compiled with options.
type yamlBuild struct {
	name    string
	entries []string
	options native.Options
}

var yamlSanitized = native.Options{Sanitize: true}

func yamlStem(entry string) string {
	return strings.TrimSuffix(filepath.Base(entry), ".ts")
}

// lowered is the product holding <stem>.c and <stem>.mjs for every entry.
func (build yamlBuild) lowered(test *testing.T) string {
	test.Helper()
	recipe := "lower+native.C+javascript.JavaScript " + strings.Join(build.entries, " ")
	return yamlProduct(test, "lowered "+build.name+" "+recipe, func() string {
		inputs := yamlPortInputs(test, "yaml-"+build.name+"-lowered")
		inputs.Flags = append(inputs.Flags, recipe)
		return buildcache.Product(test, inputs, func(directory string) error {
			for _, entry := range build.entries {
				if yamlLowerError := yamlLower(directory, entry); yamlLowerError != nil {
					return yamlLowerError
				}
			}
			return nil
		})
	})
}

// native is the product holding the binary <stem> for every entry, compiled from lowered's C with build.options.
func (build yamlBuild) native(test *testing.T) string {
	test.Helper()
	recipe := "lower+native.C+javascript.JavaScript " + strings.Join(build.entries, " ")
	// Build branches on options Flags doesn't spell, Split and Request among them, so the options are keyed whole.
	flags := append(native.Flags(build.options), fmt.Sprintf("options=%+v", build.options))
	return yamlProduct(test, "native "+build.name+" "+recipe+" "+strings.Join(flags, " "), func() string {
		lowered := build.lowered(test)
		inputs := yamlPortInputs(test, "yaml-"+build.name+"-native")
		inputs.Flags = append(inputs.Flags, recipe)
		inputs.Flags = append(inputs.Flags, flags...)
		inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		return buildcache.Product(test, inputs, func(directory string) error {
			for _, entry := range build.entries {
				source, readError := os.ReadFile(filepath.Join(lowered, yamlStem(entry)+".c"))
				if readError != nil {
					return readError
				}
				if buildError := native.Build(
					string(source),
					filepath.Join(directory, yamlStem(entry)),
					build.options,
				); buildError != nil {
					return fmt.Errorf("%s: %w", entry, buildError)
				}
			}
			return nil
		})
	})
}

// binary is entry's native binary.
func (build yamlBuild) binary(test *testing.T, entry string) string {
	test.Helper()
	return filepath.Join(build.native(test), yamlStem(entry))
}

// emitted is entry's emitted JavaScript.
func (build yamlBuild) emitted(test *testing.T, entry string) string {
	test.Helper()
	return filepath.Join(build.lowered(test), yamlStem(entry)+".mjs")
}

// yamlLower lowers entry, and writes its emitted JavaScript and its C into directory.
func yamlLower(directory, entry string) (yamlLowerError error) {
	path, yamlLowerError := filepath.Abs(entry)
	if yamlLowerError != nil {
		return yamlLowerError
	}
	program, yamlLowerError := load.Load([]string{path})
	if yamlLowerError != nil {
		return fmt.Errorf("%s: %w", entry, yamlLowerError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	lowered, yamlLowerError := lower.Lower(timeoutContext, program)
	if yamlLowerError != nil {
		return fmt.Errorf("%s: %w", entry, yamlLowerError)
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, yamlStem(entry)+".mjs"),
		[]byte(javascript.JavaScript(lowered)),
		0644,
	); writeError != nil {
		return writeError
	}
	// A program stage 0 lowers but can't emit as C is the compiler's red, named here, not a panic that takes the rest
	// of the package down with it.
	defer func() {
		if recovered := recover(); recovered != nil {
			yamlLowerError = fmt.Errorf("%s lowers, but emitting its C panics: %v", entry, recovered)
		}
	}()
	return os.WriteFile(filepath.Join(directory, yamlStem(entry)+".c"), []byte(native.C(lowered)), 0644)
}

// The port builds every comparison runs: each MatchGo entry under the sanitizers, the closed gaps, the speed probes and
// the shared-slice program both ways.
var (
	yamlFormatBuild     = yamlBuild{"format", []string{"main.ts"}, yamlSanitized}
	yamlComposeBuild    = yamlBuild{"compose", []string{"compose_main.ts"}, yamlSanitized}
	yamlUnistBuild      = yamlBuild{"unist", []string{"unist_main.ts"}, yamlSanitized}
	yamlCSTBuild        = yamlBuild{"cst", []string{"cst_main.ts"}, yamlSanitized}
	yamlPropsBuild      = yamlBuild{"props", []string{"props_main.ts"}, yamlSanitized}
	yamlSchemaBuild     = yamlBuild{"schema", []string{"schema_main.ts"}, yamlSanitized}
	yamlScalarsBuild    = yamlBuild{"scalars", []string{"scalar_main.ts"}, yamlSanitized}
	yamlWidthBuild      = yamlBuild{"width", []string{"width_main.ts"}, yamlSanitized}
	yamlClosedGapsBuild = yamlBuild{"closed-gaps", []string{
		"gaps/structuralPosition.ts",
		"gaps/multiplePush.ts",
		"gaps/stringPresence.ts",
		"gaps/valuePresence.ts",
		"gaps/assignmentValue.ts",
		"gaps/dynamicCase.ts",
		"gaps/negativeCase.ts",
		"gaps/stringFallback.ts",
		"gaps/valueConjunction.ts",
		"gaps/prefixIncrement.ts",
		"gaps/emptyAlternative.ts",
	}, yamlSanitized}
	yamlSpeedProbesBuild = yamlBuild{
		"speed-probes",
		[]string{"gaps/stringUnitScan.ts", "gaps/arenaNodeRead.ts", "gaps/numericMapLookup.ts"},
		yamlSanitized,
	}
	yamlSharedSliceBuild          = yamlBuild{"shared-slice", []string{"gaps/sharedSliceAppend.ts"}, native.Options{}}
	yamlSharedSliceSanitizedBuild = yamlBuild{"shared-slice", []string{"gaps/sharedSliceAppend.ts"}, yamlSanitized}
)

func TestProduct_YAMLFormatLowered(test *testing.T)  { test.Parallel(); yamlFormatBuild.lowered(test) }
func TestProduct_YAMLFormatNative(test *testing.T)   { test.Parallel(); yamlFormatBuild.native(test) }
func TestProduct_YAMLComposeLowered(test *testing.T) { test.Parallel(); yamlComposeBuild.lowered(test) }
func TestProduct_YAMLComposeNative(test *testing.T)  { test.Parallel(); yamlComposeBuild.native(test) }
func TestProduct_YAMLUnistLowered(test *testing.T)   { test.Parallel(); yamlUnistBuild.lowered(test) }
func TestProduct_YAMLUnistNative(test *testing.T)    { test.Parallel(); yamlUnistBuild.native(test) }
func TestProduct_YAMLCSTLowered(test *testing.T)     { test.Parallel(); yamlCSTBuild.lowered(test) }
func TestProduct_YAMLCSTNative(test *testing.T)      { test.Parallel(); yamlCSTBuild.native(test) }
func TestProduct_YAMLPropsLowered(test *testing.T)   { test.Parallel(); yamlPropsBuild.lowered(test) }
func TestProduct_YAMLPropsNative(test *testing.T)    { test.Parallel(); yamlPropsBuild.native(test) }
func TestProduct_YAMLSchemaLowered(test *testing.T)  { test.Parallel(); yamlSchemaBuild.lowered(test) }
func TestProduct_YAMLSchemaNative(test *testing.T)   { test.Parallel(); yamlSchemaBuild.native(test) }
func TestProduct_YamlScalarsLowered(test *testing.T) { test.Parallel(); yamlScalarsBuild.lowered(test) }
func TestProduct_YamlScalarsNative(test *testing.T)  { test.Parallel(); yamlScalarsBuild.native(test) }
func TestProduct_YAMLWidthLowered(test *testing.T)   { test.Parallel(); yamlWidthBuild.lowered(test) }
func TestProduct_YAMLWidthNative(test *testing.T)    { test.Parallel(); yamlWidthBuild.native(test) }
func TestProduct_YAMLClosedGapsLowered(test *testing.T) {
	test.Parallel()
	yamlClosedGapsBuild.lowered(test)
}
func TestProduct_YAMLClosedGapsNative(test *testing.T) {
	test.Parallel()
	yamlClosedGapsBuild.native(test)
}
func TestProduct_YAMLSpeedProbesLowered(test *testing.T) {
	test.Parallel()
	yamlSpeedProbesBuild.lowered(test)
}
func TestProduct_YAMLSpeedProbesNative(test *testing.T) {
	test.Parallel()
	yamlSpeedProbesBuild.native(test)
}
func TestProduct_YAMLSharedSliceLowered(test *testing.T) {
	test.Parallel()
	yamlSharedSliceBuild.lowered(test)
}
func TestProduct_YAMLSharedSliceNative(test *testing.T) {
	test.Parallel()
	yamlSharedSliceBuild.native(test)
}
func TestProduct_YAMLSharedSliceSanitized(test *testing.T) {
	test.Parallel()
	yamlSharedSliceSanitizedBuild.native(test)
}

// yamlMutation replaces from, which must occur once where it occurs, with to.
type yamlMutation struct{ name, from, to string }

// yamlMutants are one family's mutations of the program at entry. files are the sources copied, every *.ts of this
// package when nil. mutated is the file each site lives in; when empty, every copied file holding a site is mutated.
type yamlMutants struct {
	family    string
	entry     string
	files     []string
	mutated   string
	mutations []yamlMutation
}

// product is mutant index built: its mutated sources under sources/, their emitted JavaScript as mutant.mjs and their
// release binary as mutant, the build every mutant test made inside its own clock before.
func (mutants yamlMutants) product(test *testing.T, index int) string {
	test.Helper()
	mutation := mutants.mutations[index]
	files := mutants.files
	if files == nil {
		var globError error
		if files, globError = filepath.Glob("*.ts"); globError != nil {
			test.Fatal(globError)
		}
	}
	options := native.Options{}
	recipe := []string{
		"entry=" + mutants.entry,
		"files=" + strings.Join(files, " "),
		"mutated=" + mutants.mutated,
		"name=" + mutation.name,
		"from=" + mutation.from,
		"to=" + mutation.to,
	}
	return yamlProduct(test, "mutant "+mutants.family+"\n"+strings.Join(recipe, "\n"), func() string {
		inputs := yamlPortInputs(test, "yaml-"+mutants.family+"-mutant")
		inputs.Flags = append(inputs.Flags, recipe...)
		inputs.Flags = append(inputs.Flags, native.Flags(options)...)
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("options=%+v", options))
		inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		return buildcache.Product(test, inputs, func(directory string) error {
			sources := filepath.Join(directory, "sources")
			if createDirectoryError := os.Mkdir(sources, 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			for _, file := range files {
				source, readError := os.ReadFile(file)
				if readError != nil {
					return readError
				}
				if file == mutants.mutated ||
					(mutants.mutated == "" && strings.Contains(string(source), mutation.from)) {
					if strings.Count(string(source), mutation.from) != 1 {
						return fmt.Errorf("%s: mutation site must occur once", mutation.name)
					}
					source = []byte(strings.Replace(string(source), mutation.from, mutation.to, 1))
				}
				if writeError := os.WriteFile(filepath.Join(sources, file), source, 0644); writeError != nil {
					return writeError
				}
			}
			program, loadError := load.Load([]string{filepath.Join(sources, mutants.entry)})
			if loadError != nil {
				return loadError
			}
			timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			lowered, lowerError := lower.Lower(timeoutContext, program)
			if lowerError != nil {
				return lowerError
			}
			if writeError := os.WriteFile(
				filepath.Join(directory, "mutant.mjs"),
				[]byte(javascript.JavaScript(lowered)),
				0644,
			); writeError != nil {
				return writeError
			}
			return native.Build(native.C(lowered), filepath.Join(directory, "mutant"), options)
		})
	})
}
