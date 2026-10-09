package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

func TestRuntimeKeyIncludesEveryInput(t *testing.T) {
	t.Parallel()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	flags := Flags(Options{Sanitize: true, Count: true, slabs: true, cpu: "haswell"})
	key := runtimeKey(files, flags, "clang", "version 1")
	for index, flag := range flags {
		changed := append([]string{}, flags[:index]...)
		changed = append(changed, flags[index+1:]...)
		if runtimeKey(files, changed, "clang", "version 1") == key {
			t.Errorf("omitting %s did not change the key", flag)
		}
	}
	for index, file := range files {
		changed := append([]runtimeFile{}, files...)
		changed[index].contents = append([]byte{}, file.contents...)
		changed[index].contents[0] ^= 1
		if runtimeKey(changed, flags, "clang", "version 1") == key {
			t.Errorf("changing %s did not change the key", file.name)
		}
	}
	if runtimeKey(files, flags, "clang", "version 2") == key {
		t.Error("compiler version missing from key")
	}
	if runtimeKey(files, flags, "other-clang", "version 1") == key {
		t.Error("compiler path missing from key")
	}
	changed := append([]string{}, flags...)
	changed[0], changed[1] = changed[1], changed[0]
	if runtimeKey(files, changed, "clang", "version 1") == key {
		t.Error("flag order missing from key")
	}
}

// Inputs that join to the same bytes must still key apart, or one runtime is served another's
// library. Each pair below keeps the count of parts the same, so only the boundaries differ.
func TestRuntimeKeyKeepsBoundaries(t *testing.T) {
	t.Parallel()
	files := []runtimeFile{{"a.c", []byte("int a;\n")}}
	if runtimeKey(files, []string{"a", "bc"}, "clang", "version 1") == runtimeKey(files, []string{"ab", "c"}, "clang", "version 1") {
		t.Error(`flags ["a" "bc"] and ["ab" "c"] share a key`)
	}
	if runtimeKey(files, nil, "clang", "version 1") == runtimeKey(files, nil, "clangv", "ersion 1") {
		t.Error(`compiler "clang" with version "version 1" and "clangv" with "ersion 1" share a key`)
	}
	split := []runtimeFile{{"a.c", []byte("bc")}}
	moved := []runtimeFile{{"a.cb", []byte("c")}}
	if runtimeKey(split, nil, "clang", "version 1") == runtimeKey(moved, nil, "clang", "version 1") {
		t.Error(`file "a.c" holding "bc" and "a.cb" holding "c" share a key`)
	}
}

// Count is checked in the linked runtime itself, not merely in a comparison of keys. Build an
// uncounted library first: a key missing ADAMIC_COUNT would reuse it and silently lose the report.
func TestRuntimeCacheKeepsCountFlags(t *testing.T) {
	t.Parallel()
	const source = "int main(void) { return 0; }\n"
	for _, count := range []bool{false, true, false} {
		binary := filepath.Join(t.TempDir(), "main")
		if err := Build(source, binary, Options{Count: count}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("run: %v\n%s", err, output)
		}
		want := ""
		if count {
			want = "adamic: counts: allocations 0 frees 0 retains 0 releases 0 peak 0 regions 0\n"
		}
		if string(output) != want {
			t.Errorf("count %v: got %q, want %q", count, output, want)
		}
	}
}

func smallRuntime(t *testing.T, sources fstest.MapFS, cache string, compiler string) string {
	t.Helper()
	files, err := readRuntime(sources, ".")
	if err != nil {
		t.Fatal(err)
	}
	library, err := cachedRuntime(files, Flags(Options{}), compiler, "test compiler", cache)
	if err != nil {
		t.Fatal(err)
	}
	return library
}

func runLibrary(t *testing.T, library string, want string) {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte("#include <stdio.h>\nint answer(void);\nint main(void) { printf(\"%d\\n\", answer()); return 0; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "main")
	arguments := append([]string{"-o", binary, source}, RuntimeLinkFlags(library)...)
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, output)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Change the actual input after caching, and link both libraries: this catches hashing filenames
// alone, omitting headers, and compiling something other than the bytes that were hashed.
func TestRuntimeCacheRebuildsChangedSources(t *testing.T) {
	t.Parallel()
	cache := t.TempDir()
	sources := fstest.MapFS{
		"answer.c": &fstest.MapFile{Data: []byte("#include \"answer.h\"\nint answer(void) { return ANSWER + 0; }\n")},
		"answer.h": &fstest.MapFile{Data: []byte("#define ANSWER 1\n")},
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	first := smallRuntime(t, sources, cache, compiler)
	sources["answer.c"].Data = []byte("#include \"answer.h\"\nint answer(void) { return ANSWER + 1; }\n")
	second := smallRuntime(t, sources, cache, compiler)
	if first == second {
		t.Fatal("a changed runtime source reused the cached library")
	}
	sources["answer.h"].Data = []byte("#define ANSWER 3\n")
	third := smallRuntime(t, sources, cache, compiler)
	if second == third {
		t.Fatal("a changed runtime header reused the cached library")
	}
	runLibrary(t, first, "1\n")
	runLibrary(t, second, "2\n")
	runLibrary(t, third, "4\n")
}

// Many callers begin on an empty cache together. A compiler wrapper records actual compilations,
// so the test requires one build as well as an intact archive every caller can link.
func TestRuntimeCacheConcurrentBuilders(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(directory, "clang")
	log := filepath.Join(directory, "compiled")
	// Quote paths for /bin/sh without relying on the shell's environment or global test settings.
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
	script := fmt.Sprintf("#!/bin/sh\nif [ \"$1\" != '--version' ]; then echo compiled >> %s; fi\nexec %s \"$@\"\n", quote(log), quote(compiler))
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	files := []runtimeFile{{"answer.c", []byte("int answer(void) { return 42; }\n")}}
	cache := filepath.Join(directory, "cache")
	const callers = 32
	libraries := make([]string, callers)
	failures := make([]error, callers)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for index := range callers {
		workers.Go(func() {
			<-start
			libraries[index], failures[index] = cachedRuntime(files, Flags(Options{}), wrapper, "test compiler", cache)
		})
	}
	close(start)
	workers.Wait()
	for index, failure := range failures {
		if failure != nil {
			t.Fatalf("caller %d: %v", index, failure)
		}
		if libraries[index] != libraries[0] {
			t.Fatal("callers got different libraries")
		}
	}
	compiled, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(compiled) != "compiled\n" {
		t.Fatalf("want one compilation, got %q", compiled)
	}
	for _, library := range libraries {
		runLibrary(t, library, "42\n")
	}
	entries, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("want one published entry with no temporary builds, got %d", len(entries))
	}
}

// Processes do not share runtimeBuilds. Every one must see a complete archive and headers even
// when several independent builds race to publish the same directory.
func TestRuntimeCacheConcurrentProcesses(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_RUNTIME_CACHE"
	files := []runtimeFile{
		{"answer.c", []byte("#include \"answer.h\"\nint answer(void) { return ANSWER; }\n")},
		{"answer.h", []byte("#define ANSWER 42\n")},
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	if cache := os.Getenv(variable); cache != "" {
		library, err := cachedRuntime(files, Flags(Options{}), compiler, "process test", cache)
		if err != nil {
			t.Fatal(err)
		}
		header, err := os.ReadFile(filepath.Join(filepath.Dir(library), "answer.h"))
		if err != nil || string(header) != "#define ANSWER 42\n" {
			t.Fatalf("published header: %q, %v", header, err)
		}
		runLibrary(t, library, "42\n")
		return
	}
	cache := filepath.Join(t.TempDir(), "cache")
	const callers = 8
	outputs := make([][]byte, callers)
	failures := make([]error, callers)
	var workers sync.WaitGroup
	for index := range callers {
		workers.Go(func() {
			command := exec.Command(os.Args[0], "-test.run=^TestRuntimeCacheConcurrentProcesses$")
			command.Env = append(os.Environ(), variable+"="+cache)
			outputs[index], failures[index] = command.CombinedOutput()
		})
	}
	workers.Wait()
	for index, failure := range failures {
		if failure != nil {
			t.Errorf("process %d: %v\n%s", index, failure, outputs[index])
		}
	}
	entries, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("want one published entry with no temporary builds, got %d", len(entries))
	}
}
