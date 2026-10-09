package lint

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var lintOracleOnce sync.Once
var lintOraclePath string

// All live-port consumers and the declaration use this one overlay recipe and key.
func lintGoOracleProduct(t *testing.T) string {
	t.Helper()
	lintOracleOnce.Do(func() {
		prepareRegistry(t, packageDirectory)
		files := []string{"stage1/cohere/lint/testdata/oracle.go", "stage1/cohere/lint/oracle_product_test.go", "stage1/cohere/lint/registry", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum"}
		for _, file := range portFiles(t) {
			files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
		}
		inputs := buildcache.Inputs{
			Name: "lint-go-oracle", Files: files,
			Flags:     []string{"build", "-trimpath", "-ldflags=-buildid=", "overlay:live-registry-and-rule-oracles"},
			Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(directory string) error {
			started := time.Now()
			defer func() { t.Logf("cold build Go oracle=%s", time.Since(started)) }()
			sourceRoot := packageDirectory
			root, err := filepath.Abs(filepath.Join(repository, "cohere"))
			if err != nil {
				return err
			}
			side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
			if err != nil {
				return err
			}
			descriptors, err := registry.Generate(sourceRoot)
			if err != nil {
				return err
			}
			replacements := map[string]string{}
			var virtualFiles []string
			var failure error
			add := func(name, source string) {
				virtual := filepath.Join(root, "adamic_lint_"+name+".go")
				absolute, err := filepath.Abs(source)
				if err != nil {
					failure = err
					return
				}
				replacements[virtual] = absolute
				virtualFiles = append(virtualFiles, virtual)
			}
			add("oracle", side)
			add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
			for _, d := range descriptors {
				add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
			}
			if failure != nil {
				return failure
			}
			overlay, err := json.Marshal(map[string]any{"Replace": replacements})
			if err != nil {
				return err
			}
			path := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			binary := filepath.Join(directory, "oracle")
			args := append([]string{"build", "-trimpath", "-ldflags=-buildid=", "-overlay=" + path, "-o", binary}, virtualFiles...)
			command := exec.CommandContext(t.Context(), "go", args...)
			command.Dir = root
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("lint Go oracle: %w\n%s", err, output)
			}
			return nil
		})
		lintOraclePath = filepath.Join(directory, "oracle")
	})
	if lintOraclePath == "" {
		t.Fatal("lint Go oracle build failed earlier in this process")
	}
	return lintOraclePath
}
func TestProduct_LintGoOracle(t *testing.T) {
	t.Parallel()
	lintGoOracleProduct(t)
}
