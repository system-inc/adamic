package estree

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// All paths passed to Product are repository-relative. Go dependencies include
// embedded inputs and local module files; mutant content lives in Flags because
// its private source directory is deliberately outside the repository.
func recoveryCacheInputs(t *testing.T, path string) buildcache.Inputs {
	t.Helper()
	repo := root(t)
	cmd := exec.Command("go", "list", "-deps", "-json", "./internal/load", "./internal/lower", "./internal/native", "./internal/javascript")
	cmd.Dir = repo
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	files := map[string]bool{}
	var external []string
	add := func(path string) {
		relative, err := filepath.Rel(repo, path)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			external = append(external, filepath.Base(path)+":"+fmt.Sprintf("%x", sha256.Sum256(content)))
			return
		}
		files[filepath.ToSlash(relative)] = true
	}
	for {
		var p struct {
			Standard                                                                   bool
			Dir                                                                        string
			GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
			Module                                                                     *struct{ GoMod string }
		}
		if err := decoder.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if p.Standard {
			continue
		}
		for _, group := range [][]string{p.GoFiles, p.CgoFiles, p.CFiles, p.CXXFiles, p.HFiles, p.SFiles, p.SysoFiles, p.EmbedFiles} {
			for _, name := range group {
				add(filepath.Join(p.Dir, name))
			}
		}
		if p.Module != nil && p.Module.GoMod != "" {
			add(p.Module.GoMod)
			if _, err := os.Stat(filepath.Join(filepath.Dir(p.Module.GoMod), "go.sum")); err == nil {
				add(filepath.Join(filepath.Dir(p.Module.GoMod), "go.sum"))
			}
		}
	}
	add(filepath.Join(repo, "go.mod"))
	if _, err := os.Stat(filepath.Join(repo, "go.sum")); err == nil {
		add(filepath.Join(repo, "go.sum"))
	}
	// This file is part of the builder recipe, including compiler feature selection.
	add(filepath.Join(repo, "stage1/cohere/estree/recovery_cache_test.go"))
	add(filepath.Join(repo, "stage1/cohere/estree/recovery_shards_test.go"))
	names := []string{"stage1/typescript"}
	for file := range files {
		names = append(names, file)
	}
	sourceFiles, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"load.Load", "lower.Lower", "native.C", "javascript.JavaScript"}
	for _, file := range sourceFiles {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.ReplaceAll(string(source), filepath.ToSlash(filepath.Join(repo, "stage1/typescript")), "../../typescript")
		flags = append(flags, filepath.Base(file)+":"+fmt.Sprintf("%x", sha256.Sum256([]byte(normalized))))
	}
	sort.Strings(names)
	sort.Strings(external)
	flags = append(flags, external...)
	environment, err := exec.Command("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "GOWORK", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOTOOLCHAIN").Output()
	if err != nil {
		t.Fatal(err)
	}
	flags = append(flags, string(environment))
	return buildcache.Inputs{Name: "estree-lowered-program", Files: names, Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
}

// native.Build embeds its random scratch path in sanitizer metadata. This local
// recipe compiles precisely the same runtime units and sanitizer flags using
// relative names, and maps debug paths to a fixed directory. It can therefore
// survive Product's byte-for-byte audit on another machine.
func recoveryStableNative(repo, directory, source string) error {
	work := filepath.Join(directory, "compile")
	if err := os.MkdirAll(filepath.Join(work, "runtime"), 0755); err != nil {
		return err
	}
	defer os.RemoveAll(work)
	if err := os.WriteFile(filepath.Join(work, "port.c"), []byte("#line 1 \"/adamic-estree/port.c\"\n"+source), 0644); err != nil {
		return err
	}
	var features []string
	enabled := map[string]bool{}
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		if strings.Contains(source, "#define "+feature+" 1\n") {
			features = append(features, "-D"+feature+"=1")
			enabled[feature] = true
		}
	}
	inputs, err := filepath.Glob(filepath.Join(repo, "internal/native/runtime/*"))
	if err != nil {
		return err
	}
	var units []string
	for _, input := range inputs {
		name := filepath.Base(input)
		if !strings.HasSuffix(name, ".c") && !strings.HasSuffix(name, ".h") {
			continue
		}
		if name == "regexp_replace.c" && !enabled["ADAMIC_REGEXP_REPLACE_CALLBACK"] {
			continue
		}
		if name == "node_host.c" && !enabled["ADAMIC_NODE_HOST"] {
			continue
		}
		content, err := os.ReadFile(input)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(work, "runtime", name), content, 0644); err != nil {
			return err
		}
		if strings.HasSuffix(name, ".c") {
			units = append(units, filepath.Join("runtime", name))
		}
	}
	args := native.Flags(native.Options{Sanitize: true})
	args = append(args, features...)
	args = append(args, "-ffile-prefix-map="+work+"=/adamic-estree", "-fdebug-prefix-map="+work+"=/adamic-estree", "-I", "runtime", "port.c")
	args = append(args, units...)
	args = append(args, "-o", filepath.Join(directory, "port"), "-lm")
	cmd := exec.Command("clang", args...)
	cmd.Dir = work
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sanitized native build: %w: %s", err, output)
	}
	return nil
}

// Oracle answers are build products too. Use stable relative input names so an
// audit's path fields and diagnostics do not depend on t.TempDir or the checkout.
func recoveryAnswer(t *testing.T, oracle, list, mode string) []byte {
	t.Helper()
	if mode != "--manifest" && mode != "--audit" {
		t.Fatalf("unsupported recovery oracle mode %q", mode)
	}
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	manifestBytes, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	flags := []string{mode, fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary))}
	for _, path := range strings.Split(string(manifestBytes), "\n") {
		if path == "" {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, string(content))
		flags = append(flags, string(content))
	}
	inputs := buildcache.Inputs{Name: "estree-oracle-answer", Files: []string{"stage1/cohere/estree/recovery_cache_test.go"}, Flags: flags, Toolchain: []string{runtime.Version()}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		work := filepath.Join(directory, "run")
		if err := os.MkdirAll(work, 0755); err != nil {
			return err
		}
		defer os.RemoveAll(work)
		var names []string
		for i, source := range sources {
			name := fmt.Sprintf("input-%06d.ts", i)
			if err := os.WriteFile(filepath.Join(work, name), []byte(source), 0644); err != nil {
				return err
			}
			names = append(names, name)
		}
		if err := os.WriteFile(filepath.Join(work, "manifest"), []byte(strings.Join(names, "\n")+"\n"), 0644); err != nil {
			return err
		}
		args := []string{mode, "manifest"}
		if mode == "--audit" {
			args = append(args, "answers")
		}
		command := exec.Command(oracle, args...)
		command.Dir = work
		var stderr strings.Builder
		command.Stderr = &stderr
		cold := time.Now()
		answer, err := command.Output()
		t.Logf("BUILD estree-oracle-answer cold wall=%.6fs", time.Since(cold).Seconds())
		if err != nil || stderr.Len() != 0 {
			return fmt.Errorf("oracle answer: %v: %s", err, &stderr)
		}
		return os.WriteFile(filepath.Join(directory, "answer"), answer, 0644)
	})
	answer, err := os.ReadFile(filepath.Join(directory, "answer"))
	if err != nil {
		t.Fatal(err)
	}
	return answer
}

const testRecoveryNativeRecipeShards = 4

// ADAMIC_TEST_SHARD=i/n selects shards by index modulo n; unset runs all.
// The fixture checks the compiler recipe, including leak detection, rather than
// merely checking that its command line still contains sanitizer switches.
func TestRecoveryNativeRecipe(t *testing.T) {
	t.Parallel()
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	source := `#include "adamic.h"
#include <limits.h>
#include <stdlib.h>
static void * volatile lost;
__attribute__((noinline)) static void leak(void) {
 void *p=malloc(123); *(volatile char *)p=1; lost=p; lost=NULL;
}
int main(int argc,char **argv) {
 if(argc<2) return 0;
 if(argv[1][0]=='a') {char *p=malloc(4);volatile int offset=argc+8;p[offset]=1;free(p);}
 if(argv[1][0]=='u') {volatile int x=INT_MAX;return x+argc;}
 if(argv[1][0]=='l') leak();
 return 0;
}
`
	inputs := buildcache.Inputs{
		Name:      "estree-native-recipe-proof-a",
		Files:     []string{"internal/native/runtime", "internal/native/native.go", "internal/native/library.go", "stage1/cohere/estree/recovery_cache_test.go"},
		Flags:     append(native.Flags(native.Options{Sanitize: true}), source),
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("getconf", "GNU_LIBC_VERSION")},
	}
	first := recoveryProduct(t, setup, inputs, func(dir string) error { return recoveryStableNative(root(t), dir, source) })
	inputs.Name = "estree-native-recipe-proof-b"
	second := recoveryProduct(t, setup, inputs, func(dir string) error { return recoveryStableNative(root(t), dir, source) })
	a, err := os.ReadFile(filepath.Join(first, "port"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(second, "port"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("native products differ across independent build directories")
	}
	cases := recoveryCases("native-recipe", []string{"normal", "address", "undefined", "leak"})
	shards := partitionRecovery(t, cases, testRecoveryNativeRecipeShards)
	runRecoveryShards(t, shards, func(t *testing.T, _ int, cases []recoveryCase) {
		for _, c := range cases {
			var args []string
			if c.source != "normal" {
				args = []string{c.source[:1]}
			}
			cmd := exec.Command(filepath.Join(first, "port"), args...)
			cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=23", "UBSAN_OPTIONS=halt_on_error=1")
			output, err := cmd.CombinedOutput()
			if c.source == "normal" {
				if err != nil || len(output) != 0 {
					t.Fatalf("positive fixture: %v: %s", err, output)
				}
				continue
			}
			diagnostic := map[string]string{"address": "AddressSanitizer: heap-buffer-overflow", "undefined": "runtime error: signed integer overflow", "leak": "LeakSanitizer: detected memory leaks"}[c.source]
			if err == nil || !strings.Contains(string(output), diagnostic) {
				t.Fatalf("%s sanitizer failed to catch fixture: %v: %s", c.source, err, output)
			}
		}
	})
}
