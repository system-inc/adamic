package typeaware

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
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

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "typeaware-products-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	productDirectory = directory
	code := m.Run()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func sharedProduct(key string, build func(string) (string, error)) (string, error) {
	stored, _ := products.LoadOrStore(key, &product{})
	value := stored.(*product)
	value.once.Do(func() {
		directory, err := os.MkdirTemp(productDirectory, "product-")
		if err != nil {
			value.err = err
			return
		}
		value.value, value.err = build(directory)
	})
	return value.value, value.err
}

// Shared commands write their logs beside the product. Failed builders return
// their diagnostics to every waiter instead of leaving an empty cached path.
func buildProduct(t *testing.T, name string, command *exec.Cmd, directory string) error {
	t.Helper()
	stdout, err := os.Create(filepath.Join(directory, "build.stdout"))
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(directory, "build.stderr"))
	if err != nil {
		return err
	}
	defer stderr.Close()
	command.Stdout, command.Stderr = stdout, stderr
	started := time.Now()
	err = command.Run()
	t.Logf("phase command %s %.6fs", name, time.Since(started).Seconds())
	if err != nil {
		out, _ := os.ReadFile(stdout.Name())
		report, _ := os.ReadFile(stderr.Name())
		return fmt.Errorf("%s: %w\n%s\n%s", name, err, out, report)
	}
	return nil
}

func (h *harness) stage0() string {
	h.t.Helper()
	path, err := sharedProduct("stage0", func(directory string) (string, error) {
		binary := filepath.Join(directory, "adamic")
		command := exec.Command("go", "build", "-o", binary, "./cmd/adamic")
		command.Dir = h.repository
		return binary, buildProduct(h.t, "stage0", command, directory)
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return path
}

// A child has its own logs, sources and counter. Sharing h.t or h.next would
// race as soon as parallel subtests start.
func (h *harness) child(t *testing.T) *harness {
	t.Helper()
	directory := filepath.Join(h.directory, filepath.Base(t.Name()))
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, repository: h.repository, directory: directory, parallel: true}
}

// Check publication, key isolation and error propagation under the concurrency
// used by the mutants. These are build products, never cached oracle verdicts.
func TestSharedProductPublication(t *testing.T) {
	t.Parallel()
	key := t.Name() + t.TempDir()
	var lock sync.Mutex
	builds := 0
	var wait sync.WaitGroup
	answers := make(chan string, 16)
	for range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			answer, err := sharedProduct(key+" good", func(directory string) (string, error) {
				lock.Lock()
				builds++
				lock.Unlock()
				path := filepath.Join(directory, "ready")
				return path, os.WriteFile(path, []byte("complete"), 0644)
			})
			if err != nil {
				t.Error(err)
				return
			}
			data, err := os.ReadFile(answer)
			if err != nil || string(data) != "complete" {
				t.Errorf("incomplete product: %q, %v", data, err)
			}
			answers <- answer
		}()
	}
	wait.Wait()
	close(answers)
	if builds != 1 {
		t.Fatalf("built shared product %d times", builds)
	}
	var first string
	for answer := range answers {
		if first == "" {
			first = answer
		}
		if answer != first {
			t.Fatal("waiters received different products")
		}
	}
	other, err := sharedProduct(key+" other", func(directory string) (string, error) { return directory, nil })
	if err != nil || other == first {
		t.Fatalf("different keys shared a product: %s, %v", other, err)
	}
	for range 2 {
		_, err := sharedProduct(key+" failed", func(_ string) (string, error) { return "", fmt.Errorf("broken build") })
		if err == nil || err.Error() != "broken build" {
			t.Fatalf("lost build failure: %v", err)
		}
	}
}

// testing's parent elapsed time excludes time waiting for parallel children.
// Cleanup runs after all children, so this records the complete suite wall time.
// It is an observation, not a deadline or a performance assertion.
func traceGroup(t *testing.T) {
	t.Helper()
	started := time.Now()
	t.Cleanup(func() { t.Logf("phase group %s %.6fs", t.Name(), time.Since(started).Seconds()) })
}
