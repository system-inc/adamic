package lint

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type checkerCompilation struct {
	c, javascript string
	bridge        bool
	err           error
	ready         chan struct{}
}

var checkerCompilationMutex sync.Mutex
var checkerCompilations = map[string]*checkerCompilation{}

// Follow the actual static import and export declarations, including type-only
// dependencies. Content and absolute path participate in the key: a changed
// mutant or a renamed/copied module must never reuse an earlier compilation.
func checkerSourceKey(entry string) (string, error) {
	digest := sha256.New()
	seen := map[string]bool{}
	var visit func(string) error
	visit = func(path string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if seen[absolute] {
			return nil
		}
		seen[absolute] = true
		data, err := os.ReadFile(absolute)
		if err != nil {
			return err
		}
		fmt.Fprintf(digest, "%d:%s%d:", len(absolute), absolute, len(data))
		digest.Write(data)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(filepath.ToSlash(absolute))}, string(data), core.ScriptKindTS)
		for _, statement := range file.Statements.Nodes {
			if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
				continue
			}
			specifier := statement.ModuleSpecifier()
			if specifier == nil || specifier.Text() == "adamic" {
				continue
			}
			imported := specifier.Text()
			if !filepath.IsAbs(imported) {
				imported = filepath.Join(filepath.Dir(absolute), imported)
			}
			if err := visit(imported); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(entry); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func checkerCompile(t *testing.T, directory string) *checkerCompilation {
	t.Helper()
	prepareRegistry(t, directory)
	entry := filepath.Join(directory, "main.ts")
	key, err := checkerSourceKey(entry)
	if err != nil {
		t.Fatal(err)
	}
	checkerCompilationMutex.Lock()
	built, found := checkerCompilations[key]
	if !found {
		built = &checkerCompilation{ready: make(chan struct{})}
		checkerCompilations[key] = built
	}
	checkerCompilationMutex.Unlock()
	if found {
		<-built.ready
	} else {
		started := time.Now()
		program, err := load.Load([]string{entry})
		loaded := time.Now()
		if err == nil {
			program.EnableTSGo()
			lowered, lowerErr := lower.Lower(context.Background(), program)
			err = lowerErr
			if err == nil {
				loweredAt := time.Now()
				built.bridge = native.UsesTSGo(lowered)
				if built.bridge {
					built.c, err = native.TSGoC(lowered)
				} else {
					built.c = native.C(lowered)
				}
				emittedAt := time.Now()
				if err == nil {
					built.javascript = javascript.JavaScript(lowered)
				}
				t.Logf("checker compilation phases: load=%s lower=%s C=%s JavaScript=%s", loaded.Sub(started), loweredAt.Sub(loaded), emittedAt.Sub(loweredAt), time.Since(emittedAt))
			}
		}
		built.err = err
		close(built.ready)
	}
	if built.err != nil {
		t.Fatal(built.err)
	}
	return built
}
