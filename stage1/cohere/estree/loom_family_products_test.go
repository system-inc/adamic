package estree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Four workers use the gate's four CPUs; source, sanitizer flags and cases stay fixed.
func estreeFamilyNativeOptions() native.Options {
	return native.Options{Sanitize: true, Split: true, Jobs: 4}
}

func TestProduct_DeepMutantsNative0(test *testing.T) {
	test.Parallel()
	started := time.Now()
	deepMutantsNative(test, deepMutantsEnumeration()[0])
	test.Logf("product: %.3fs", time.Since(started).Seconds())
}

func TestProduct_SyntaxMutantNative_000(test *testing.T) {
	test.Parallel()
	syntaxMutantNativeProduct(test, syntaxMutantEnumeration()[0])
}

func TestProduct_SyntaxMutantsSetup_000(test *testing.T) {
	test.Parallel()
	syntaxMutantPreparedProduct(test, 0)
}
func TestProduct_SyntaxMutantsSetup_001(test *testing.T) {
	test.Parallel()
	syntaxMutantPreparedProduct(test, 1)
}
func TestProduct_SyntaxMutantsSetup_002(test *testing.T) {
	test.Parallel()
	syntaxMutantPreparedProduct(test, 2)
}

// Checkpoint the checked IR so the native emitter has its own cached build stage. Gob retains the concrete
// expression/statement types (ir.RegisterGob); derived private facts are recomputed by the backend. Source snapshots
// remain stable for Node execution.
func estreeFamilyLowerCheckpoint(dir string, snapshot map[string][]byte) error {
	source := filepath.Join(dir, "source")
	if createDirectoryError := os.Mkdir(source, 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	for name, data := range snapshot {
		if writeError := os.WriteFile(filepath.Join(source, name), data, 0644); writeError != nil {
			return writeError
		}
	}
	// Named placelessly: the IR's positions would otherwise carry this build directory, random per build (#tqrqx60).
	program, loadDisplayedError := load.LoadDisplayed([]string{filepath.Join(source, "main.ts")}, buildcache.Relative)
	if loadDisplayedError != nil {
		return loadDisplayedError
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return lowerError
	}
	ir.RegisterGob()
	file, createError := os.Create(filepath.Join(dir, "program.gob"))
	if createError != nil {
		return createError
	}
	encodeError := gob.NewEncoder(file).Encode(lowered)
	closeError := file.Close()
	if encodeError != nil {
		return fmt.Errorf("encode checked IR: %w", encodeError)
	}
	if closeError != nil {
		return closeError
	}
	// The source's imports name this checkout for the lowering above; the product names it <repository> (#tqrqx60).
	return buildcache.RelativeFiles(source)
}
func estreeFamilyReadCheckpoint(dir string) (*ir.Program, error) {
	ir.RegisterGob()
	file, openError := os.Open(filepath.Join(dir, "program.gob"))
	if openError != nil {
		return nil, openError
	}
	defer file.Close()
	var program ir.Program
	if decodeError := gob.NewDecoder(file).Decode(&program); decodeError != nil {
		return nil, decodeError
	}
	return &program, nil
}
func estreeFamilyCopySource(from, to string) error {
	if createDirectoryError := os.Mkdir(filepath.Join(to, "source"), 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	files, globError := filepath.Glob(filepath.Join(from, "source", "*.ts"))
	if globError != nil {
		return globError
	}
	for _, file := range files {
		data, readError := os.ReadFile(file)
		if readError != nil {
			return readError
		}
		if writeError := os.WriteFile(filepath.Join(to, "source", filepath.Base(file)), data, 0644); writeError != nil {
			return writeError
		}
	}
	return nil
}
func TestProduct_DeepMutantsIR0(test *testing.T) {
	test.Parallel()
	deepMutantsIRProduct(test, deepMutantsEnumeration()[0])
}
func TestProduct_DeepMutantsIR1(test *testing.T) {
	test.Parallel()
	deepMutantsIRProduct(test, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsIR2(test *testing.T) {
	test.Parallel()
	deepMutantsIRProduct(test, deepMutantsEnumeration()[2])
}

// Hash actual Go dependencies (including embeds and local modules), rather than
// unrelated upstream test corpora. This caches metadata only, never a build.
var estreeFamilyDependencies struct {
	once     sync.Once
	files    []string
	external []string
	err      error
}

func estreeFamilyDependencyFiles(test *testing.T) []string {
	test.Helper()
	estreeFamilyDependencies.once.Do(func() {
		repo := root(test)
		command := toolchain.CommandIn(
			repo,
			"go",
			"list",
			"-deps",
			"-json",
			"./internal/load",
			"./internal/lower",
			"./internal/native",
			"./internal/javascript",
			"github.com/system-inc/cohere/internal/format/estree",
		)
		output, outputError := command.Output()
		if outputError != nil {
			estreeFamilyDependencies.err = fmt.Errorf("Go dependencies: %w", outputError)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(output))
		files := map[string]bool{
			"go.mod":               true,
			"go.work":              true,
			"stage1/typescript":    true,
			"stage1/cohere/estree": true,
		}
		// add keys a file the build reads: by its repository path inside the tree, and outside it (go's module cache)
		// by module, version and its path from the module's directory, never by where this machine keeps the cache. A
		// Loom runner's is /tmp/loom-sources/<blob>-modcache, and an absolute path keyed the Go oracle apart from
		// Workshop's (landable-10's run).
		add := func(path string, module *estreeModule) error {
			relative, relativePathError := filepath.Rel(repo, path)
			if relativePathError != nil {
				return relativePathError
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				data, readError := os.ReadFile(path)
				if readError != nil {
					return readError
				}
				name := path
				if module != nil && module.Dir != "" {
					inside, relativePathError := filepath.Rel(module.Dir, path)
					if relativePathError != nil {
						return relativePathError
					}
					name = module.Path + "@" + module.Version + "/" + filepath.ToSlash(inside)
				}
				estreeFamilyDependencies.external = append(
					estreeFamilyDependencies.external,
					fmt.Sprintf("external=%s:%x", name, sha256.Sum256(data)),
				)
				return nil
			}
			files[filepath.ToSlash(relative)] = true
			return nil
		}
		for {
			var item struct {
				Standard                                                                   bool
				Dir                                                                        string
				GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
				Module                                                                     *estreeModule
				Error                                                                      *struct{ Err string }
			}
			if decodeError := decoder.Decode(&item); decodeError == io.EOF {
				break
			} else if decodeError != nil {
				estreeFamilyDependencies.err = decodeError
				return
			}
			if item.Error != nil {
				estreeFamilyDependencies.err = fmt.Errorf("Go dependency: %s", item.Error.Err)
				return
			}
			if item.Standard {
				continue
			}
			for _, group := range [][]string{
				item.GoFiles,
				item.CgoFiles,
				item.CFiles,
				item.CXXFiles,
				item.HFiles,
				item.SFiles,
				item.SysoFiles,
				item.EmbedFiles,
			} {
				for _, name := range group {
					if addError := add(filepath.Join(item.Dir, name), item.Module); addError != nil {
						estreeFamilyDependencies.err = addError
						return
					}
				}
			}
			if item.Module != nil && item.Module.GoMod != "" {
				if addError := add(item.Module.GoMod, item.Module); addError != nil {
					estreeFamilyDependencies.err = addError
					return
				}
				sum := filepath.Join(filepath.Dir(item.Module.GoMod), "go.sum")
				if _, statError := os.Stat(sum); statError == nil {
					if addError := add(sum, item.Module); addError != nil {
						estreeFamilyDependencies.err = addError
						return
					}
				}
			}
		}
		for file := range files {
			estreeFamilyDependencies.files = append(estreeFamilyDependencies.files, file)
		}
		sort.Strings(estreeFamilyDependencies.files)
		sort.Strings(estreeFamilyDependencies.external)
	})
	if estreeFamilyDependencies.err != nil {
		test.Fatal(estreeFamilyDependencies.err)
	}
	return estreeFamilyDependencies.files
}

func TestProduct_SyntaxMutantIR_000(test *testing.T) {
	test.Parallel()
	syntaxMutantIRProduct(test, syntaxMutantEnumeration()[0])
}
func TestProduct_SyntaxMutantIR_001(test *testing.T) {
	test.Parallel()
	syntaxMutantIRProduct(test, syntaxMutantEnumeration()[1])
}
func TestProduct_SyntaxMutantIR_002(test *testing.T) {
	test.Parallel()
	syntaxMutantIRProduct(test, syntaxMutantEnumeration()[2])
}

// estreeModule is the module go list reports a dependency in.
type estreeModule struct{ Path, Version, GoMod, Dir string }

func estreeFamilyDependencyFlags(test *testing.T) []string {
	test.Helper()
	estreeFamilyDependencyFiles(test)
	return estreeFamilyDependencies.external
}
