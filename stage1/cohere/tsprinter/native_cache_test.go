package tsprinter

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// The emitted C completely describes a lowered port, including source mutants.
// Products are only executed; builds write into the private miss directory.
func cachedNative(program *ir.Program) (string, error) {
	source := native.C(program)
	options := native.Options{Sanitize: true}
	flags := append(nativeCacheFlags(options), fmt.Sprintf("source-sha256=%x", sha256.Sum256([]byte(source))))
	inputs, err := nativeCacheInputs(flags)
	if err != nil {
		return "", err
	}
	directory, err := buildcache.Get(inputs, func(directory string) error {
		return native.BuildProduct(source, directory, options)
	})
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "port"), nil
}

var nativeCacheSources struct {
	sync.Once
	inputs buildcache.Inputs
	err    error
}

// GoInputs resolves all compiler dependencies, module files, build settings and
// toolchains. Reuse that accounting rather than maintaining another source list.
func nativeCacheInputs(flags []string) (buildcache.Inputs, error) {
	nativeCacheSources.Do(func() {
		files := map[string]bool{
			"stage1/cohere/tsprinter/helpers_test.go":      true,
			"stage1/cohere/tsprinter/native_cache_test.go": true,
		}
		for _, pkg := range []string{"./internal/load", "./internal/lower", "./internal/native"} {
			inputs, err := buildcache.GoInputs("tsprinter", pkg, nil, nil)
			if err != nil {
				nativeCacheSources.err = err
				return
			}
			for _, file := range inputs.Files {
				files[file] = true
			}
			nativeCacheSources.inputs.Flags = append(nativeCacheSources.inputs.Flags, inputs.Flags...)
			nativeCacheSources.inputs.Toolchain = append(nativeCacheSources.inputs.Toolchain, inputs.Toolchain...)
		}
		for file := range files {
			nativeCacheSources.inputs.Files = append(nativeCacheSources.inputs.Files, file)
		}
		sort.Strings(nativeCacheSources.inputs.Files)
	})
	inputs := buildcache.Inputs{
		Name:      "tsprinter-sanitized-port",
		Files:     append([]string(nil), nativeCacheSources.inputs.Files...),
		Flags:     append(append([]string(nil), nativeCacheSources.inputs.Flags...), flags...),
		Toolchain: append(append([]string(nil), nativeCacheSources.inputs.Toolchain...), buildcache.Tool("clang", "--version"), buildcache.Tool("clang", "-print-search-dirs"), buildcache.Tool("llvm-ar", "--version")),
	}
	return inputs, nativeCacheSources.err
}

// These are TypeScript source mutants, not Go overlay builds. Snapshot every
// mutated file before the cache lookup, and key the imported parser/scanner too.
// A hit skips both lowering and native compilation, but still runs the oracle.
func cachedMutant(path string) (string, error) {
	options := native.Options{Sanitize: true}
	flags := append(nativeCacheFlags(options), "entry="+filepath.Base(path))
	parser, err := filepath.Abs("../../typescript/parser")
	if err != nil {
		return "", err
	}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
	if err != nil {
		return "", err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		// mutatedPort's absolute parser import has an incidental machine location.
		source := strings.ReplaceAll(string(data), parser+"/", "stage1/typescript/parser/")
		flags = append(flags, fmt.Sprintf("%s=%x", filepath.Base(file), sha256.Sum256([]byte(source))))
	}
	inputs, err := nativeCacheInputs(flags)
	if err != nil {
		return "", err
	}
	inputs.Name = "tsprinter-sanitized-mutant"
	err = filepath.WalkDir(filepath.Join(repository, "stage1/typescript"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	directory, err := buildcache.Get(inputs, func(directory string) error {
		program, err := lowerProgram(path)
		if err != nil {
			return err
		}
		return native.BuildProduct(native.C(program), directory, options)
	})
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "port"), nil
}

func nativeCacheFlags(options native.Options) []string {
	flags := native.Flags(options)
	for _, name := range []string{"PATH", "CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "LIBRARY_PATH", "COMPILER_PATH", "GCC_EXEC_PREFIX", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return flags
}
