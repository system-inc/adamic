package estree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Port products include the lowered C, emitted JS, source bundle and sanitized
// binary. GoInputs supplies every compiler/checker dependency and embedded runtime
// input; the bundle flags also cover a mutant's exact replacement bytes. The
// independent oracle is a separate GoBuild product (oracle_product_test.go).
func estreePortProduct(t *testing.T, main string, sanitize bool) (string, string) {
	t.Helper()
	repository := root(t)
	inputs, err := buildcache.GoInputs("estree-port", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "estree port C/native/JS"
	inputs.Flags = append(inputs.Flags, "product format 2")
	inputs.Files = append(inputs.Files, "stage1/typescript", "stage1/cohere/estree/products_test.go")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	options := native.Options{Sanitize: sanitize}
	inputs.Flags = append(inputs.Flags, native.Flags(options)...)
	inputs.Flags = append(inputs.Flags, "compile all runtime translation units", "-ffile-prefix-map=<product>=/adamic-estree-product", "-ffile-prefix-map=<repository>=/adamic-source", "-fdebug-compilation-dir=/adamic-estree-product")
	for _, name := range []string{"CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "LIBRARY_PATH", "COMPILER_PATH", "SDKROOT", "SOURCE_DATE_EPOCH"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(main), "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	bundle := make(map[string][]byte)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(data), "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(repository, "stage1/typescript"))+"/")
		name := filepath.Base(path)
		bundle[name] = []byte(text)
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("source %s %x", name, sha256.Sum256([]byte(text))))
	}
	if _, ok := bundle[filepath.Base(main)]; !ok {
		t.Fatal("main missing from source bundle")
	}
	inputs.Flags = append(inputs.Flags, "main="+filepath.Base(main))
	sort.Strings(inputs.Files)
	directory := buildcache.Product(t, inputs, func(directory string) error {
		sources := filepath.Join(directory, "sources")
		if err := os.Mkdir(sources, 0755); err != nil {
			return err
		}
		names := make([]string, 0, len(bundle))
		for name := range bundle {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(sources, name), bundle[name], 0644); err != nil {
				return err
			}
		}
		program, err := load.Load([]string{filepath.Join(sources, filepath.Base(main))})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		source := native.C(lowered)
		if err := os.WriteFile(filepath.Join(directory, "port.c"), []byte(source), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "port.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
			return err
		}
		return estreeCompileProduct(repository, directory, source, options)
	})
	return filepath.Join(directory, "port"), filepath.Join(directory, "port.mjs")
}

// Use the backend's flags and every runtime translation unit it would put in its
// whole-archive link, with the same source feature defines. Normalize build paths
// in DWARF and sanitizer locations: native.Build uses random temporary paths,
// which otherwise make a fetched-product audit differ despite identical inputs.
func estreeCompileProduct(repository, directory, source string, options native.Options) error {
	runtimeDirectory := filepath.Join(directory, "runtime")
	if err := os.Mkdir(runtimeDirectory, 0755); err != nil {
		return err
	}
	flags := native.Flags(options)
	flags = append(flags, "-ffile-prefix-map="+directory+"=/adamic-estree-product", "-ffile-prefix-map="+repository+"=/adamic-source", "-fdebug-compilation-dir=/adamic-estree-product")
	features := make(map[string]bool)
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		if strings.Contains(source, "#define "+feature+" 1\n") {
			features[feature] = true
			flags = append(flags, "-D"+feature+"=1")
		}
	}
	files, err := filepath.Glob(filepath.Join(repository, "internal/native/runtime", "*"))
	if err != nil {
		return err
	}
	arguments := append(flags, "-I", "runtime", "-o", "port", "port.c")
	for _, path := range files {
		name := filepath.Base(path)
		if !strings.HasSuffix(name, ".c") && !strings.HasSuffix(name, ".h") {
			continue
		}
		if name == "regexp_replace.c" && !features["ADAMIC_REGEXP_REPLACE_CALLBACK"] {
			continue
		}
		if name == "node_host.c" && !features["ADAMIC_NODE_HOST"] {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(runtimeDirectory, name)
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
		if strings.HasSuffix(name, ".c") {
			arguments = append(arguments, filepath.Join("runtime", name))
		}
	}
	arguments = append(arguments, "-lm")
	command := exec.Command("clang", arguments...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("clang port product: %w\n%s", err, output)
	}
	return nil
}
