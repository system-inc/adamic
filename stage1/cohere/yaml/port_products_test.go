package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Every program a test of this package lowers or compiles is a product, built ahead by its TestProduct_ twin, so a
// test only reads it and no test's clock holds a lowering or a clang run (#9wyp4tt). A build is lowered once into
// <stem>.c and <stem>.mjs, then compiled by clang into <stem> with its options; a mutant is its mutated sources, their
// emitted JavaScript and their native binary in one product.

var yamlPortFiles struct {
	once  sync.Once
	files []string
}

// yamlPortInputs are a build's declared inputs: the compiler's implementation, this package's TypeScript, and the Go of
// cohere and bridge/tsgo that the loader links, with the files those embed (TestYamlPortKeyCoversEmbeds), without test
// evidence. The walk runs once per process.
func yamlPortInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	yamlPortFiles.once.Do(func() {
		files := []string{"go.mod", "go.work", "stage1/cohere/yaml/port_products_test.go"}
		for _, directory := range []string{"internal", "stage1/cohere/yaml", "cohere", "bridge/tsgo"} {
			err := filepath.WalkDir(filepath.Join(repository, directory), func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					switch entry.Name() {
					case ".git", "node_modules", "testdata", "audit", "performance", "evidence", "logs":
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				switch filepath.Ext(path) {
				case ".go", ".ts", ".c", ".h", ".mod", ".sum", ".gz":
				case ".json":
					// cohere/policy embeds its rule messages and naming tables, which the lowering links through rule_runner.
					if !strings.HasPrefix(relative, filepath.Join("cohere", "policy")+string(filepath.Separator)) {
						return nil
					}
				default:
					return nil
				}
				files = append(files, relative)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		yamlPortFiles.files = files
	})
	if len(yamlPortFiles.files) == 0 {
		t.Fatal("build input listing failed earlier in this process")
	}
	return buildcache.Inputs{Name: name, Files: append([]string(nil), yamlPortFiles.files...), Toolchain: []string{runtime.Version()}, Flags: []string{"GOFLAGS=" + os.Getenv("GOFLAGS"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}}
}

// TestYamlPortKeyCoversEmbeds holds the walk to every file go list says the build's packages embed. Embedded text
// reaches a build's output, so an embed the walk dropped would leave its products keyed as before after an edit, and a
// stale product passes without checking anything.
func TestYamlPortKeyCoversEmbeds(t *testing.T) {
	t.Parallel()
	walked := map[string]bool{}
	for _, file := range yamlPortInputs(t, "yaml-embeds").Files {
		walked[file] = true
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// toolchain's go, under its environment, in the repository root with PWD set to it, so go.work resolves there.
	var output, errorOutput bytes.Buffer
	command := toolchain.CommandIn(root, "go", "list", "-deps", "-json", "./internal/load", "./internal/lower", "./internal/javascript", "./internal/native")
	command.Stdout, command.Stderr = &output, &errorOutput
	if err := childguard.Run(command, childguard.Options{}); err != nil {
		t.Fatalf("go list: %v\n%s", err, errorOutput.Bytes())
	}
	decoder := json.NewDecoder(&output)
	checked := 0
	for {
		var listed struct {
			ImportPath string
			Dir        string
			EmbedFiles []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		for _, name := range listed.EmbedFiles {
			relative, err := filepath.Rel(root, filepath.Join(listed.Dir, name))
			if err != nil {
				t.Fatal(err)
			}
			// A module outside the repository is pinned by its go.sum line, which the walk holds.
			if strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				continue
			}
			checked++
			if !walked[relative] {
				t.Errorf("%s embeds %s, which the products' key doesn't read", listed.ImportPath, relative)
			}
		}
	}
	if checked == 0 {
		t.Fatal("go list reported no embedded file inside the repository, so this check held nothing")
	}
	t.Logf("%d embedded files checked against the products' key", checked)
}

type yamlProductValue struct {
	once      sync.Once
	directory string
}

var yamlProducts sync.Map

// yamlProduct runs build once per process for each key, and fails every later caller when that build failed.
func yamlProduct(t *testing.T, key string, build func() string) string {
	t.Helper()
	stored, _ := yamlProducts.LoadOrStore(key, &yamlProductValue{})
	value := stored.(*yamlProductValue)
	value.once.Do(func() { value.directory = build() })
	if value.directory == "" {
		t.Fatalf("product %s failed earlier in this process", key)
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
func (build yamlBuild) lowered(t *testing.T) string {
	t.Helper()
	recipe := "lower+native.C+javascript.JavaScript " + strings.Join(build.entries, " ")
	return yamlProduct(t, "lowered "+build.name+" "+recipe, func() string {
		inputs := yamlPortInputs(t, "yaml-"+build.name+"-lowered")
		inputs.Flags = append(inputs.Flags, recipe)
		return buildcache.Product(t, inputs, func(directory string) error {
			for _, entry := range build.entries {
				if err := yamlLower(directory, entry); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

// native is the product holding the binary <stem> for every entry, compiled from lowered's C with build.options.
func (build yamlBuild) native(t *testing.T) string {
	t.Helper()
	recipe := "lower+native.C+javascript.JavaScript " + strings.Join(build.entries, " ")
	flags := native.Flags(build.options)
	return yamlProduct(t, "native "+build.name+" "+recipe+" "+strings.Join(flags, " "), func() string {
		lowered := build.lowered(t)
		inputs := yamlPortInputs(t, "yaml-"+build.name+"-native")
		inputs.Flags = append(inputs.Flags, recipe)
		inputs.Flags = append(inputs.Flags, flags...)
		inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		return buildcache.Product(t, inputs, func(directory string) error {
			for _, entry := range build.entries {
				source, err := os.ReadFile(filepath.Join(lowered, yamlStem(entry)+".c"))
				if err != nil {
					return err
				}
				if err := native.Build(string(source), filepath.Join(directory, yamlStem(entry)), build.options); err != nil {
					return fmt.Errorf("%s: %w", entry, err)
				}
			}
			return nil
		})
	})
}

// binary is entry's native binary.
func (build yamlBuild) binary(t *testing.T, entry string) string {
	t.Helper()
	return filepath.Join(build.native(t), yamlStem(entry))
}

// emitted is entry's emitted JavaScript.
func (build yamlBuild) emitted(t *testing.T, entry string) string {
	t.Helper()
	return filepath.Join(build.lowered(t), yamlStem(entry)+".mjs")
}

// yamlLower lowers entry, and writes its emitted JavaScript and its C into directory.
func yamlLower(directory, entry string) (err error) {
	path, err := filepath.Abs(entry)
	if err != nil {
		return err
	}
	program, err := load.Load([]string{path})
	if err != nil {
		return fmt.Errorf("%s: %w", entry, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	lowered, err := lower.Lower(ctx, program)
	if err != nil {
		return fmt.Errorf("%s: %w", entry, err)
	}
	if err := os.WriteFile(filepath.Join(directory, yamlStem(entry)+".mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		return err
	}
	// A program stage 0 lowers but can't emit as C is the compiler's red, named here, not a panic that takes the rest of
	// the package down with it.
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%s lowers, but emitting its C panics: %v", entry, recovered)
		}
	}()
	return os.WriteFile(filepath.Join(directory, yamlStem(entry)+".c"), []byte(native.C(lowered)), 0644)
}

// The port builds every comparison runs: each MatchGo entry under the sanitizers, the closed gaps, the speed probes and
// the shared-slice program both ways.
var (
	yamlFormatBuild               = yamlBuild{"format", []string{"main.ts"}, yamlSanitized}
	yamlComposeBuild              = yamlBuild{"compose", []string{"compose_main.ts"}, yamlSanitized}
	yamlUnistBuild                = yamlBuild{"unist", []string{"unist_main.ts"}, yamlSanitized}
	yamlCSTBuild                  = yamlBuild{"cst", []string{"cst_main.ts"}, yamlSanitized}
	yamlPropsBuild                = yamlBuild{"props", []string{"props_main.ts"}, yamlSanitized}
	yamlSchemaBuild               = yamlBuild{"schema", []string{"schema_main.ts"}, yamlSanitized}
	yamlScalarsBuild              = yamlBuild{"scalars", []string{"scalar_main.ts"}, yamlSanitized}
	yamlWidthBuild                = yamlBuild{"width", []string{"width_main.ts"}, yamlSanitized}
	yamlClosedGapsBuild           = yamlBuild{"closed-gaps", []string{"gaps/stringPresence.ts", "gaps/valuePresence.ts", "gaps/assignmentValue.ts", "gaps/dynamicCase.ts", "gaps/negativeCase.ts", "gaps/stringFallback.ts", "gaps/valueConjunction.ts", "gaps/prefixIncrement.ts", "gaps/emptyAlternative.ts"}, yamlSanitized}
	yamlSpeedProbesBuild          = yamlBuild{"speed-probes", []string{"gaps/stringUnitScan.ts", "gaps/arenaNodeRead.ts", "gaps/numericMapLookup.ts"}, yamlSanitized}
	yamlSharedSliceBuild          = yamlBuild{"shared-slice", []string{"gaps/sharedSliceAppend.ts"}, native.Options{}}
	yamlSharedSliceSanitizedBuild = yamlBuild{"shared-slice", []string{"gaps/sharedSliceAppend.ts"}, yamlSanitized}
)

func TestProduct_YAMLFormatLowered(t *testing.T)      { t.Parallel(); yamlFormatBuild.lowered(t) }
func TestProduct_YAMLFormatNative(t *testing.T)       { t.Parallel(); yamlFormatBuild.native(t) }
func TestProduct_YAMLComposeLowered(t *testing.T)     { t.Parallel(); yamlComposeBuild.lowered(t) }
func TestProduct_YAMLComposeNative(t *testing.T)      { t.Parallel(); yamlComposeBuild.native(t) }
func TestProduct_YAMLUnistLowered(t *testing.T)       { t.Parallel(); yamlUnistBuild.lowered(t) }
func TestProduct_YAMLUnistNative(t *testing.T)        { t.Parallel(); yamlUnistBuild.native(t) }
func TestProduct_YAMLCSTLowered(t *testing.T)         { t.Parallel(); yamlCSTBuild.lowered(t) }
func TestProduct_YAMLCSTNative(t *testing.T)          { t.Parallel(); yamlCSTBuild.native(t) }
func TestProduct_YAMLPropsLowered(t *testing.T)       { t.Parallel(); yamlPropsBuild.lowered(t) }
func TestProduct_YAMLPropsNative(t *testing.T)        { t.Parallel(); yamlPropsBuild.native(t) }
func TestProduct_YAMLSchemaLowered(t *testing.T)      { t.Parallel(); yamlSchemaBuild.lowered(t) }
func TestProduct_YAMLSchemaNative(t *testing.T)       { t.Parallel(); yamlSchemaBuild.native(t) }
func TestProduct_YamlScalarsLowered(t *testing.T)     { t.Parallel(); yamlScalarsBuild.lowered(t) }
func TestProduct_YamlScalarsNative(t *testing.T)      { t.Parallel(); yamlScalarsBuild.native(t) }
func TestProduct_YAMLWidthLowered(t *testing.T)       { t.Parallel(); yamlWidthBuild.lowered(t) }
func TestProduct_YAMLWidthNative(t *testing.T)        { t.Parallel(); yamlWidthBuild.native(t) }
func TestProduct_YAMLClosedGapsLowered(t *testing.T)  { t.Parallel(); yamlClosedGapsBuild.lowered(t) }
func TestProduct_YAMLClosedGapsNative(t *testing.T)   { t.Parallel(); yamlClosedGapsBuild.native(t) }
func TestProduct_YAMLSpeedProbesLowered(t *testing.T) { t.Parallel(); yamlSpeedProbesBuild.lowered(t) }
func TestProduct_YAMLSpeedProbesNative(t *testing.T)  { t.Parallel(); yamlSpeedProbesBuild.native(t) }
func TestProduct_YAMLSharedSliceLowered(t *testing.T) { t.Parallel(); yamlSharedSliceBuild.lowered(t) }
func TestProduct_YAMLSharedSliceNative(t *testing.T)  { t.Parallel(); yamlSharedSliceBuild.native(t) }
func TestProduct_YAMLSharedSliceSanitized(t *testing.T) {
	t.Parallel()
	yamlSharedSliceSanitizedBuild.native(t)
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
func (mutants yamlMutants) product(t *testing.T, index int) string {
	t.Helper()
	mutation := mutants.mutations[index]
	files := mutants.files
	if files == nil {
		var err error
		if files, err = filepath.Glob("*.ts"); err != nil {
			t.Fatal(err)
		}
	}
	options := native.Options{}
	recipe := []string{"entry=" + mutants.entry, "files=" + strings.Join(files, " "), "mutated=" + mutants.mutated, "name=" + mutation.name, "from=" + mutation.from, "to=" + mutation.to}
	return yamlProduct(t, "mutant "+mutants.family+"\n"+strings.Join(recipe, "\n"), func() string {
		inputs := yamlPortInputs(t, "yaml-"+mutants.family+"-mutant")
		inputs.Flags = append(inputs.Flags, recipe...)
		inputs.Flags = append(inputs.Flags, native.Flags(options)...)
		inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		return buildcache.Product(t, inputs, func(directory string) error {
			sources := filepath.Join(directory, "sources")
			if err := os.Mkdir(sources, 0755); err != nil {
				return err
			}
			for _, file := range files {
				source, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				if file == mutants.mutated || (mutants.mutated == "" && strings.Contains(string(source), mutation.from)) {
					if strings.Count(string(source), mutation.from) != 1 {
						return fmt.Errorf("%s: mutation site must occur once", mutation.name)
					}
					source = []byte(strings.Replace(string(source), mutation.from, mutation.to, 1))
				}
				if err := os.WriteFile(filepath.Join(sources, file), source, 0644); err != nil {
					return err
				}
			}
			program, err := load.Load([]string{filepath.Join(sources, mutants.entry)})
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			lowered, err := lower.Lower(ctx, program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, "mutant.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
				return err
			}
			return native.Build(native.C(lowered), filepath.Join(directory, "mutant"), options)
		})
	})
}
