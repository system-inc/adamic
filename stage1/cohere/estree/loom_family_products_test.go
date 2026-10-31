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

func TestProduct_DeepMutantsNative0(t *testing.T) {
	t.Parallel()
	started := time.Now()
	deepMutantsNative(t, deepMutantsEnumeration()[0])
	t.Logf("product: %.3fs", time.Since(started).Seconds())
}

func TestProduct_SyntaxMutantNative_000(t *testing.T) {
	t.Parallel()
	syntaxMutantNativeProduct(t, syntaxMutantEnumeration()[0])
}

func TestProduct_SyntaxMutantsSetup_000(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 0)
}
func TestProduct_SyntaxMutantsSetup_001(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 1)
}
func TestProduct_SyntaxMutantsSetup_002(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 2)
}

// Checkpoint the checked IR so the native emitter has its own cached build stage. Gob retains the concrete
// expression/statement types (ir.RegisterGob); derived private facts are recomputed by the backend. Source snapshots
// remain stable for Node execution.
func estreeFamilyLowerCheckpoint(dir string, snapshot map[string][]byte) error {
	source := filepath.Join(dir, "source")
	if err := os.Mkdir(source, 0755); err != nil {
		return err
	}
	for name, data := range snapshot {
		if err := os.WriteFile(filepath.Join(source, name), data, 0644); err != nil {
			return err
		}
	}
	// Named placelessly: the IR's positions would otherwise carry this build directory, random per build (#tqrqx60).
	program, err := load.LoadDisplayed([]string{filepath.Join(source, "main.ts")}, buildcache.Relative)
	if err != nil {
		return err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	ir.RegisterGob()
	file, err := os.Create(filepath.Join(dir, "program.gob"))
	if err != nil {
		return err
	}
	err = gob.NewEncoder(file).Encode(lowered)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("encode checked IR: %w", err)
	}
	if closeErr != nil {
		return closeErr
	}
	// The source's imports name this checkout for the lowering above; the product names it <repository> (#tqrqx60).
	return buildcache.RelativeFiles(source)
}
func estreeFamilyReadCheckpoint(dir string) (*ir.Program, error) {
	ir.RegisterGob()
	file, err := os.Open(filepath.Join(dir, "program.gob"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var program ir.Program
	if err := gob.NewDecoder(file).Decode(&program); err != nil {
		return nil, err
	}
	return &program, nil
}
func estreeFamilyCopySource(from, to string) error {
	if err := os.Mkdir(filepath.Join(to, "source"), 0755); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(from, "source", "*.ts"))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(to, "source", filepath.Base(file)), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
func TestProduct_DeepMutantsIR0(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[0])
}
func TestProduct_DeepMutantsIR1(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsIR2(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[2])
}

// Hash actual Go dependencies (including embeds and local modules), rather than
// unrelated upstream test corpora. This caches metadata only, never a build.
var estreeFamilyDependencies struct {
	once     sync.Once
	files    []string
	external []string
	err      error
}

func estreeFamilyDependencyFiles(t *testing.T) []string {
	t.Helper()
	estreeFamilyDependencies.once.Do(func() {
		repo := root(t)
		command := toolchain.CommandIn(repo, "go", "list", "-deps", "-json", "./internal/load", "./internal/lower", "./internal/native", "./internal/javascript", "github.com/system-inc/cohere/internal/format/estree")
		output, err := command.Output()
		if err != nil {
			estreeFamilyDependencies.err = fmt.Errorf("Go dependencies: %w", err)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(output))
		files := map[string]bool{"go.mod": true, "go.work": true, "stage1/typescript": true, "stage1/cohere/estree": true}
		// add keys a file the build reads: by its repository path inside the tree, and outside it (go's module cache) by
		// module, version and its path from the module's directory, never by where this machine keeps the cache. A Loom
		// runner's is /tmp/loom-sources/<blob>-modcache, and an absolute path keyed the Go oracle apart from Workshop's
		// (landable-10's run).
		add := func(path string, module *estreeModule) error {
			relative, err := filepath.Rel(repo, path)
			if err != nil {
				return err
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				name := path
				if module != nil && module.Dir != "" {
					inside, err := filepath.Rel(module.Dir, path)
					if err != nil {
						return err
					}
					name = module.Path + "@" + module.Version + "/" + filepath.ToSlash(inside)
				}
				estreeFamilyDependencies.external = append(estreeFamilyDependencies.external, fmt.Sprintf("external=%s:%x", name, sha256.Sum256(data)))
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
			if err := decoder.Decode(&item); err == io.EOF {
				break
			} else if err != nil {
				estreeFamilyDependencies.err = err
				return
			}
			if item.Error != nil {
				estreeFamilyDependencies.err = fmt.Errorf("Go dependency: %s", item.Error.Err)
				return
			}
			if item.Standard {
				continue
			}
			for _, group := range [][]string{item.GoFiles, item.CgoFiles, item.CFiles, item.CXXFiles, item.HFiles, item.SFiles, item.SysoFiles, item.EmbedFiles} {
				for _, name := range group {
					if err := add(filepath.Join(item.Dir, name), item.Module); err != nil {
						estreeFamilyDependencies.err = err
						return
					}
				}
			}
			if item.Module != nil && item.Module.GoMod != "" {
				if err := add(item.Module.GoMod, item.Module); err != nil {
					estreeFamilyDependencies.err = err
					return
				}
				sum := filepath.Join(filepath.Dir(item.Module.GoMod), "go.sum")
				if _, err := os.Stat(sum); err == nil {
					if err := add(sum, item.Module); err != nil {
						estreeFamilyDependencies.err = err
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
		t.Fatal(estreeFamilyDependencies.err)
	}
	return estreeFamilyDependencies.files
}

func TestProduct_SyntaxMutantIR_000(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[0])
}
func TestProduct_SyntaxMutantIR_001(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[1])
}
func TestProduct_SyntaxMutantIR_002(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[2])
}

// estreeModule is the module go list reports a dependency in.
type estreeModule struct{ Path, Version, GoMod, Dir string }

func estreeFamilyDependencyFlags(t *testing.T) []string {
	t.Helper()
	estreeFamilyDependencyFiles(t)
	return estreeFamilyDependencies.external
}
