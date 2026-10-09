package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A repository of its own under the test's directory, so keys hash files the test controls.
func repository(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	write(t, directory, "source/a.c", "int a;\n")
	write(t, directory, "source/b.c", "int b;\n")
	write(t, directory, "flags.txt", "-O2\n")
	return directory
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func key(t *testing.T, root string, inputs Inputs) string {
	t.Helper()
	value, err := Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

var base = Inputs{Name: "port", Files: []string{"source", "flags.txt"}, Flags: []string{"-DADAMIC"}, Toolchain: []string{"clang 20.1.8"}}

// Every input that can change a product changes its key: drop any of them from the hash and its subtest fails.
func TestEveryInputChangesTheKey(t *testing.T) {
	changed := map[string]func(t *testing.T, root string) Inputs{
		"name":         func(t *testing.T, root string) Inputs { in := base; in.Name = "port2"; return in },
		"flag":         func(t *testing.T, root string) Inputs { in := base; in.Flags = []string{"-DADAMIC_TSGO"}; return in },
		"flag order":   func(t *testing.T, root string) Inputs { in := base; in.Flags = []string{"-DADAMIC", "-O0"}; return in },
		"toolchain":    func(t *testing.T, root string) Inputs { in := base; in.Toolchain = []string{"clang 21.0.0"}; return in },
		"file list":    func(t *testing.T, root string) Inputs { in := base; in.Files = []string{"source"}; return in },
		"file content": func(t *testing.T, root string) Inputs { write(t, root, "source/a.c", "int a2;\n"); return base },
		"added file":   func(t *testing.T, root string) Inputs { write(t, root, "source/c.c", ""); return base },
		"renamed file": func(t *testing.T, root string) Inputs {
			if err := os.Rename(filepath.Join(root, "source/b.c"), filepath.Join(root, "source/b2.c")); err != nil {
				t.Fatal(err)
			}
			return base
		},
		"exec bit": func(t *testing.T, root string) Inputs {
			if err := os.Chmod(filepath.Join(root, "flags.txt"), 0o755); err != nil {
				t.Fatal(err)
			}
			return base
		},
		"empty directory": func(t *testing.T, root string) Inputs {
			if err := os.Mkdir(filepath.Join(root, "source", "empty"), 0o755); err != nil {
				t.Fatal(err)
			}
			return base
		},
	}
	for name, change := range changed {
		t.Run(name, func(t *testing.T) {
			root := repository(t)
			before := key(t, root, base)
			if after := key(t, root, change(t, root)); after == before {
				t.Fatalf("changing the %s left the key %s", name, before)
			}
		})
	}
	t.Run("nothing changed", func(t *testing.T) {
		root := repository(t)
		before := key(t, root, base)
		later := time.Now().Add(time.Hour)
		if err := os.Chtimes(filepath.Join(root, "source/a.c"), later, later); err != nil {
			t.Fatal(err)
		}
		if after := key(t, root, base); after != before {
			t.Fatalf("a touched file with the same content moved the key: %s, then %s", before, after)
		}
	})
	t.Run("outside the repository", func(t *testing.T) {
		root := repository(t)
		for _, name := range []string{"/etc/hosts", "../escape"} {
			in := base
			in.Files = []string{name}
			if _, err := Key(root, in); err == nil {
				t.Fatalf("input %s was accepted", name)
			}
		}
	})
}

// Products are cached under a test-owned directory, keyed in this repository (Get finds it from the working
// directory), with the census log captured.
func cached(t *testing.T) (string, string) {
	t.Helper()
	cache := t.TempDir()
	log := filepath.Join(t.TempDir(), "builds.log")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	t.Setenv("ADAMIC_BUILD_LOG", log)
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	return cache, log
}

var thisPackage = Inputs{Name: "buildcache test", Files: []string{"internal/buildcache/buildcache.go"}}

func TestABuildRunsOncePerKey(t *testing.T) {
	_, log := cached(t)
	var builds int
	build := func(directory string) error {
		builds++
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	}
	first := Product(t, thisPackage, build)
	second := Product(t, thisPackage, build)
	if builds != 1 || first != second {
		t.Fatalf("%d builds, products %s and %s", builds, first, second)
	}
	if content, err := os.ReadFile(filepath.Join(second, "product")); err != nil || string(content) != "built" {
		t.Fatalf("the product reads %q, %v", content, err)
	}
	lines, _ := os.ReadFile(log)
	if got := strings.Fields(strings.ReplaceAll(string(lines), "\n", " ")); len(got) != 10 || got[3] != "miss" || got[8] != "hit" || got[1] != "buildcache_test" {
		t.Fatalf("census lines: %q", lines)
	}
}

func TestAFailedBuildPublishesNothing(t *testing.T) {
	cache, _ := cached(t)
	failure := errors.New("clang failed")
	if _, err := Get(thisPackage, func(directory string) error {
		os.WriteFile(filepath.Join(directory, "half"), nil, 0o644)
		return failure
	}); !errors.Is(err, failure) {
		t.Fatalf("the build's error was lost: %v", err)
	}
	entries, _ := os.ReadDir(cache)
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("a failed build left %s in the cache", entry.Name())
		}
	}
	var built bool
	Product(t, thisPackage, func(string) error { built = true; return nil })
	if !built {
		t.Fatal("the next call reused a failed build")
	}
}

func TestUncachedModeBuildsEveryTime(t *testing.T) {
	cached(t)
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	var builds int
	build := func(string) error { builds++; return nil }
	first := Product(t, thisPackage, build)
	second := Product(t, thisPackage, build)
	if builds != 2 || first == second {
		t.Fatalf("%d builds, products %s and %s", builds, first, second)
	}
}

func TestParallelCallersBuildOnce(t *testing.T) {
	cached(t)
	var builds atomic.Int32
	var group sync.WaitGroup
	products := make([]string, 8)
	for index := range products {
		group.Add(1)
		go func() {
			defer group.Done()
			directory, err := Get(thisPackage, func(string) error {
				builds.Add(1)
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			if err != nil {
				t.Error(err)
			}
			products[index] = directory
		}()
	}
	group.Wait()
	if builds.Load() != 1 {
		t.Fatalf("%d builds for one key", builds.Load())
	}
	for _, product := range products {
		if product != products[0] {
			t.Fatalf("products differ: %v", products)
		}
	}
}

func TestToolNamesItselfOnce(t *testing.T) {
	if value := Tool("go", "version"); !strings.HasPrefix(value, "go version: go version go") {
		t.Fatalf("Tool reported %q", value)
	}
	if value := Tool("adamic-no-such-tool"); !strings.Contains(value, "adamic-no-such-tool: ") || !strings.Contains(value, "(") {
		t.Fatalf("a missing tool reported %q", value)
	}
}
