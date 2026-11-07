package typeaware

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Products live for the package, never for the first test that requests them.
// Main does not yet carry internal/native/testbuildcache. This small keyed
// store follows lint's shared helper; builders return errors, never call Fatal
// inside Once, and publish only completed immutable products.
var productDirectory string
var products sync.Map

type product struct {
	once  sync.Once
	value string
	err   error
}

// Not parallel: package-wide build product directory.
func TestMain(suite *testing.M) {
	directory, temporaryDirectoryError := os.MkdirTemp("", "typeaware-products-")
	if temporaryDirectoryError != nil {
		fmt.Fprintln(os.Stderr, temporaryDirectoryError)
		os.Exit(1)
	}
	productDirectory = directory
	code := suite.Run()
	if removeError := os.RemoveAll(directory); removeError != nil {
		fmt.Fprintln(os.Stderr, removeError)
		code = 1
	}
	os.Exit(code)
}

func sharedProduct(key string, build func(string) (string, error)) (string, error) {
	stored, _ := products.LoadOrStore(key, &product{})
	value := stored.(*product)
	value.once.Do(func() {
		directory, temporaryDirectoryError := os.MkdirTemp(productDirectory, "product-")
		if temporaryDirectoryError != nil {
			value.err = temporaryDirectoryError
			return
		}
		value.value, value.err = build(directory)
	})
	return value.value, value.err
}

// Shared commands write their logs beside the product. Failed builders return
// their diagnostics to every waiter instead of leaving an empty cached path.
func buildProduct(test *testing.T, name string, command *exec.Cmd, directory string) error {
	test.Helper()
	stdout, createError := os.Create(filepath.Join(directory, "build.stdout"))
	if createError != nil {
		return createError
	}
	defer stdout.Close()
	stderr, createError := os.Create(filepath.Join(directory, "build.stderr"))
	if createError != nil {
		return createError
	}
	defer stderr.Close()
	command.Stdout, command.Stderr = stdout, stderr
	started := time.Now()
	typeAwareRunBuildCommandError := typeAwareRunBuildCommand(command)
	test.Logf("phase command %s %.6fs", name, time.Since(started).Seconds())
	if typeAwareRunBuildCommandError != nil {
		out, _ := os.ReadFile(stdout.Name())
		report, _ := os.ReadFile(stderr.Name())
		return fmt.Errorf("%s: %w\n%s\n%s", name, typeAwareRunBuildCommandError, out, report)
	}
	return nil
}

func (local *harness) stage0() string {
	local.t.Helper()
	path, cachedTypeAwareProductError := cachedTypeAwareProduct("stage0", func(directory string) (string, error) {
		binary := filepath.Join(directory, "adamic")
		command := toolchain.CommandIn(
			local.repository,
			"go",
			append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-o", binary, "./cmd/adamic")...,
		)
		return binary, buildProduct(local.t, "stage0", command, directory)
	})
	if cachedTypeAwareProductError != nil {
		local.t.Fatal(cachedTypeAwareProductError)
	}
	return path
}

// A child has its own logs, sources and counter. Sharing h.t or h.next would
// race as soon as parallel subtests start.
func (local *harness) child(test *testing.T) *harness {
	test.Helper()
	directory := filepath.Join(local.directory, filepath.Base(test.Name()))
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	return &harness{t: test, repository: local.repository, directory: directory, parallel: true}
}

// Check publication, key isolation and error propagation under the concurrency
// used by the mutants. These are build products, never cached oracle verdicts.
func TestSharedProductPublication(test *testing.T) {
	test.Parallel()
	key := test.Name() + test.TempDir()
	var lock sync.Mutex
	builds := 0
	var wait sync.WaitGroup
	answers := make(chan string, 16)
	for range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			answer, sharedProductError := sharedProduct(key+" good", func(directory string) (string, error) {
				lock.Lock()
				builds++
				lock.Unlock()
				path := filepath.Join(directory, "ready")
				return path, os.WriteFile(path, []byte("complete"), 0644)
			})
			if sharedProductError != nil {
				test.Error(sharedProductError)
				return
			}
			data, readError := os.ReadFile(answer)
			if readError != nil || string(data) != "complete" {
				test.Errorf("incomplete product: %q, %v", data, readError)
			}
			answers <- answer
		}()
	}
	wait.Wait()
	close(answers)
	if builds != 1 {
		test.Fatalf("built shared product %d times", builds)
	}
	var first string
	for answer := range answers {
		if first == "" {
			first = answer
		}
		if answer != first {
			test.Fatal("waiters received different products")
		}
	}
	other, sharedProductError := sharedProduct(
		key+" other",
		func(directory string) (string, error) { return directory, nil },
	)
	if sharedProductError != nil || other == first {
		test.Fatalf("different keys shared a product: %s, %v", other, sharedProductError)
	}
	for range 2 {
		_, sharedProductError := sharedProduct(
			key+" failed",
			func(_ string) (string, error) { return "", fmt.Errorf("broken build") },
		)
		if sharedProductError == nil || sharedProductError.Error() != "broken build" {
			test.Fatalf("lost build failure: %v", sharedProductError)
		}
	}
}

// testing's parent elapsed time excludes time waiting for parallel children.
// Cleanup runs after all children, so this records the complete suite wall time.
// It is an observation, not a deadline or a performance assertion.
func traceGroup(test *testing.T) {
	test.Helper()
	started := time.Now()
	test.Cleanup(func() { test.Logf("phase group %s %.6fs", test.Name(), time.Since(started).Seconds()) })
}
