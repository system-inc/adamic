package buildcache

import (
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Product identity is the check the shared tier lands on (Kirk's ruling, Oct 9): a product built or fetched through
// the cache must be byte-identical to a fresh build. This file holds it for GoBuild, whose products are the only ones
// this package builds itself; cmd/productidentity holds it for every TestProduct_ in the repository. A flipped byte in
// a stored blob is TestAStoredProductThatDoesntCheckIsPoisoning's "a blob that isn't its hash".

// The flags GoBuild builds with, spelled out here rather than read from reproducible, so a fresh build is main's
// recipe as written and a change to GoBuild's flags shows up as a difference.
var freshFlags = []string{"-trimpath", "-ldflags=-buildid=", "-buildvcs=false"}

// freshGoBuild is 'go build' run directly at root with GoBuild's flags, nothing from buildcache, and its own empty Go
// build cache when goCache is set, so nothing cached anywhere can stand in for the build.
func freshGoBuild(t *testing.T, root, output, pkg string, goCache bool, arguments []string, environment ...string) ([]byte, float64) {
	t.Helper()
	started := time.Now()
	path := filepath.Join(t.TempDir(), output)
	command := exec.Command("go", append(append(append(append([]string{"build"}, freshFlags...), arguments...), "-o", path), pkg)...)
	command.Dir = root
	command.Env = append(os.Environ(), environment...)
	if goCache {
		command.Env = append(command.Env, "GOCACHE="+t.TempDir())
	}
	if combined, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, combined)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content, time.Since(started).Seconds()
}

// identical says whether two builds of one product are the same bytes. On macOS a cgo binary is linked by the system
// linker, which can write a different LC_UUID for the same objects, and the code signature hashes the page holding it
// (#cchsq45): a difference confined to those two is reported with its offsets and accepted there, and any other byte
// fails. Everywhere else every byte must match.
func identical(t *testing.T, name string, want, got []byte) bool {
	t.Helper()
	if bytes.Equal(want, got) {
		return true
	}
	ranges := differingRanges(want, got)
	if runtime.GOOS == "darwin" && len(want) == len(got) {
		regions := uuidAndSignature(want)
		confined := len(regions) > 0
		for _, span := range ranges {
			inside := false
			for _, region := range regions {
				inside = inside || span[0] >= region[0] && span[1] < region[1]
			}
			confined = confined && inside
		}
		if confined {
			t.Logf("%s: differs only in LC_UUID and the code signature, at %s (macOS, #cchsq45; byte-identical on Linux)", name, describeRanges(ranges, regions))
			return true
		}
	}
	t.Errorf("%s: %d and %d bytes, differing at %s", name, len(want), len(got), describeRanges(ranges, uuidAndSignature(want)))
	return false
}

func differingRanges(first, second []byte) [][2]int {
	var ranges [][2]int
	for index := 0; index < max(len(first), len(second)); index++ {
		if index < len(first) && index < len(second) && first[index] == second[index] {
			continue
		}
		if len(ranges) > 0 && ranges[len(ranges)-1][1] >= index-16 {
			ranges[len(ranges)-1][1] = index
		} else {
			ranges = append(ranges, [2]int{index, index})
		}
	}
	return ranges
}

func describeRanges(ranges [][2]int, regions map[string][2]int) string {
	var described []string
	for _, span := range ranges[:min(len(ranges), 12)] {
		label := ""
		for name, region := range regions {
			if span[0] >= region[0] && span[1] < region[1] {
				label = " (" + name + ")"
			}
		}
		described = append(described, fmt.Sprintf("0x%x-0x%x%s", span[0], span[1], label))
	}
	if len(ranges) > 12 {
		described = append(described, fmt.Sprintf("and %d more", len(ranges)-12))
	}
	return strings.Join(described, " ")
}

// uuidAndSignature is where a Mach-O's LC_UUID bytes and its code signature sit, or nothing for another format.
func uuidAndSignature(content []byte) map[string][2]int {
	file, err := macho.NewFile(bytes.NewReader(content))
	if err != nil {
		return nil
	}
	defer file.Close()
	regions := map[string][2]int{}
	offset := 32
	if file.Magic == macho.Magic32 {
		offset = 28
	}
	for _, load := range file.Loads {
		raw := load.Raw()
		switch file.ByteOrder.Uint32(raw[0:4]) {
		case 0x1b: // LC_UUID: the 16 bytes after its command and size
			regions["LC_UUID"] = [2]int{offset + 8, offset + len(raw)}
		case 0x1d: // LC_CODE_SIGNATURE: dataoff, datasize
			start := int(file.ByteOrder.Uint32(raw[8:12]))
			regions["code signature"] = [2]int{start, start + int(file.ByteOrder.Uint32(raw[12:16]))}
		}
		offset += len(raw)
	}
	return regions
}

// Every GoBuild product, built fresh, built cold through the cache into a store, and fetched warm by another machine
// from that store, is the same bytes all three ways.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestProductIdentityOfGoBuildProducts(t *testing.T) {
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, built := range []struct{ output, pkg string }{
		{"hello", "./internal/buildcache/testdata/hello"},
		{"withc", "./internal/buildcache/testdata/withc"},
	} {
		t.Run(built.output, func(t *testing.T) {
			store, log := shared(t)
			token := filepath.Join(t.TempDir(), "publish-token")
			os.WriteFile(token, []byte("gate-box-token"), 0o600)
			t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
			fresh, freshSeconds := freshGoBuild(t, root, built.output, built.pkg, true, nil)

			started := time.Now()
			cold, err := os.ReadFile(GoBuild(t, built.output, built.pkg, nil))
			if err != nil {
				t.Fatal(err)
			}
			coldSeconds := time.Since(started).Seconds()
			if len(store.writes) == 0 {
				t.Fatal("the cold build published nothing")
			}

			// Another machine: an empty cache, no credential.
			t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
			t.Setenv("ADAMIC_BUILD_STORE_TOKEN", filepath.Join(t.TempDir(), "no-token"))
			writes := len(store.writes)
			started = time.Now()
			warm, err := os.ReadFile(GoBuild(t, built.output, built.pkg, nil))
			if err != nil {
				t.Fatal(err)
			}
			warmSeconds := time.Since(started).Seconds()
			if lines, _ := os.ReadFile(log); strings.Count(string(lines), " miss ") != 1 || strings.Count(string(lines), " fetched ") != 1 || len(store.writes) != writes {
				t.Fatalf("the cold build must miss and the warm one fetch, writing nothing: census %q, writes %v", lines, store.writes[writes:])
			}

			t.Logf("product %s fresh=%x cold=%x warm=%x seconds fresh=%.2f cold=%.2f warm=%.2f",
				built.output, sha256.Sum256(fresh), sha256.Sum256(cold), sha256.Sum256(warm), freshSeconds, coldSeconds, warmSeconds)
			identical(t, built.output+" fresh and cold", fresh, cold)
			// A fetch is a copy of what was stored, so warm and cold are the same bytes on every platform.
			if !bytes.Equal(cold, warm) {
				t.Errorf("the fetched product isn't what was built: %s", describeRanges(differingRanges(cold, warm), uuidAndSignature(cold)))
			}
		})
	}
}

// inRepository points repositoryRoot at directory for this test, and back after it.
func inRepository(t *testing.T, directory string) {
	t.Chdir(directory)
	resetRepositoryRoot()
	t.Cleanup(resetRepositoryRoot)
}

func resetRepositoryRoot() {
	root.once = sync.Once{}
	root.directory = ""
	root.err = nil
}

// freshGoCache is a Go build cache holding the standard library and runtime/cgo and nothing of the repository under
// test, built once per test from a module of its own: each fresh build gets a copy, so it compiles every package of
// the repository from source without paying for the standard library each time.
func freshGoCache(t *testing.T) func() string {
	t.Helper()
	template := t.TempDir()
	module := t.TempDir()
	write(t, module, "go.mod", "module example.com/template\n\ngo 1.21\n")
	write(t, module, "main.go", "package main\n\n// int one(void) { return 1; }\nimport \"C\"\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(C.one()) }\n")
	command := exec.Command("go", append(append([]string{"build"}, freshFlags...), "-o", filepath.Join(module, "template"), ".")...)
	command.Dir = module
	command.Env = append(os.Environ(), "GOCACHE="+template, "GOWORK=off", "GOFLAGS=")
	if combined, err := command.CombinedOutput(); err != nil {
		t.Fatalf("the template Go build cache: %v\n%s", err, combined)
	}
	return func() string {
		t.Helper()
		directory := filepath.Join(t.TempDir(), "go-build")
		if err := os.CopyFS(directory, os.DirFS(template)); err != nil {
			t.Fatal(err)
		}
		return directory
	}
}

// A repository of its own for the stale-product tests, keyed and built from the test's working directory.
func goModule(t *testing.T, files map[string]string) string {
	t.Helper()
	directory := t.TempDir()
	for name, content := range files {
		write(t, directory, name, content)
	}
	inRepository(t, directory)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	return directory
}

type inputChange struct {
	input  string
	change func()
	output string
}

// changeEachInput applies each change in turn and requires GoBuild to give, cold through one cache kept across the
// steps and warm from the store into an empty cache, exactly what a fresh build of the changed tree gives: the same
// output and the same bytes. A key that ignores the input finds the product from before the change in either place.
func changeEachInput(t *testing.T, directory, output, pkg string, arguments, environment *[]string, steps []inputChange) {
	t.Helper()
	store, log := shared(t)
	token := filepath.Join(t.TempDir(), "publish-token")
	os.WriteFile(token, []byte("gate-box-token"), 0o600)
	cold := t.TempDir()
	cache := freshGoCache(t)
	var previous []byte
	for _, step := range steps {
		step.change()
		fresh, _ := freshGoBuild(t, directory, output, pkg, false, *arguments, append(*environment, "GOCACHE="+cache())...)
		if got := run(t, filepath.Join(t.TempDir(), output), fresh); got != step.output {
			t.Fatalf("after changing %s a fresh build printed %q, want %q: the test's own expectation is wrong", step.input, got, step.output)
		}
		if previous != nil && !changedOutsideTheLink(previous, fresh) {
			t.Fatalf("changing %s didn't change the product, so this step proves nothing", step.input)
		}
		previous = fresh
		for _, way := range []string{"cold", "warm"} {
			if way == "cold" {
				t.Setenv("ADAMIC_BUILD_CACHE_DIR", cold)
				t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
			} else {
				t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
				t.Setenv("ADAMIC_BUILD_STORE_TOKEN", filepath.Join(t.TempDir(), "no-token"))
			}
			before, _ := os.ReadFile(log)
			binary := GoBuild(t, output, pkg, *arguments, *environment...)
			after, _ := os.ReadFile(log)
			// The census line this build wrote: 'build <name> <key12> <outcome> <seconds>', after any note about the store.
			var outcome []string
			for _, line := range strings.Split(string(after[len(before):]), "\n") {
				if fields := strings.Fields(line); len(fields) == 5 && fields[0] == "build" {
					outcome = fields
				}
			}
			content, err := os.ReadFile(binary)
			if err != nil {
				t.Fatal(err)
			}
			got := run(t, binary, nil)
			stale := got != step.output || !identical(t, fmt.Sprintf("after changing %s, %s and fresh", step.input, way), fresh, content)
			switch {
			case !stale:
			case way == "cold" && len(outcome) == 5 && outcome[3] == "hit":
				t.Fatalf("after changing %s, the cold product printed %q, a fresh build %q: the key ignores %s, so the product from before the change was a hit", step.input, got, step.output, step.input)
			case way == "cold":
				t.Fatalf("after changing %s, the cold product printed %q, a fresh build %q, though its key moved (%v): go build itself built a stale product, and it was published under an honest key", step.input, got, step.output, outcome)
			default:
				t.Fatalf("after changing %s, the warm product printed %q, a fresh build %q (%v): the key ignores %s, or the store served another product", step.input, got, step.output, outcome, step.input)
			}
		}
	}
	if len(store.objects) == 0 {
		t.Fatal("nothing was published, so no warm build was fetched")
	}
}

// run runs a product and returns what it printed; with content, it writes the product to path first.
func run(t *testing.T, path string, content []byte) string {
	t.Helper()
	if content != nil {
		if err := os.WriteFile(path, content, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	output, err := exec.Command(path).Output()
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return strings.TrimSpace(string(output))
}

func replaceIn(t *testing.T, directory, name, old, new string) func() {
	return func() {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil || !bytes.Contains(content, []byte(old)) {
			t.Fatalf("%s has no %q: %v", name, old, err)
		}
		write(t, directory, name, strings.Replace(string(content), old, new, 1))
	}
}

// A key that leaves out one of a GoBuild's inputs serves a stale product: each input go build reads from inside the
// package's own directories is changed in turn, and the product must follow. Drop that input from GoInputs and its
// step fails, naming it.
// Not parallel: changes the working directory (t.Chdir) and points the build cache and the store at this test through t.Setenv.
func TestProductIdentityAfterEveryInputChanges(t *testing.T) {
	directory := goModule(t, map[string]string{
		"go.mod":          "module github.com/system-inc/adamic\n\ngo 1.21\n",
		"app/greeting.h":  "#define GREETING \"hello\"\n",
		"app/message.txt": "first",
		"app/plain.go":    "//go:build !fancy\n\npackage main\n\nconst tagged = \"plain\"\n",
		"app/fancy.go":    "//go:build fancy\n\npackage main\n\nconst tagged = \"fancy\"\n",
		"app/answer.c":    "#ifndef ANSWER_OFFSET\n#define ANSWER_OFFSET 0\n#endif\nint answer(void) { return 42 + ANSWER_OFFSET; }\n",
		"library/library.go": `package library

// Sum adds what each closure captured: 0+1+2 when each iteration has its own variable (go 1.22 on), 3+3+3 before.
func Sum() int {
	var captured []func() int
	for index := 0; index < 3; index++ {
		captured = append(captured, func() int { return index })
	}
	total := 0
	for _, value := range captured {
		total += value()
	}
	return total
}
`,
		"app/main.go": `package main

/*
#include "greeting.h"
static const char *greeting(void) { return GREETING; }
int answer(void);
*/
import "C"

import (
	_ "embed"
	"fmt"

	"github.com/system-inc/adamic/library"
)

//go:embed message.txt
var message string

func main() { fmt.Println(C.GoString(C.greeting()), C.answer(), message, library.Sum(), tagged) }
`,
	})
	var arguments, environment []string
	changeEachInput(t, directory, "app", "./app", &arguments, &environment, []inputChange{
		{"nothing (the first build)", func() {}, "hello 42 first 9 plain"},
		{"the package's Go file", replaceIn(t, directory, "app/main.go", "fmt.Println(", `fmt.Println("go",`), "go hello 42 first 9 plain"},
		{"a dependency's Go file", replaceIn(t, directory, "library/library.go", "index < 3", "index < 4"), "go hello 42 first 16 plain"},
		{"the package's header", replaceIn(t, directory, "app/greeting.h", "hello", "howdy"), "go howdy 42 first 16 plain"},
		{"a C file", replaceIn(t, directory, "app/answer.c", "42", "43"), "go howdy 43 first 16 plain"},
		{"an embedded file", replaceIn(t, directory, "app/message.txt", "first", "second"), "go howdy 43 second 16 plain"},
		{"go.mod", replaceIn(t, directory, "go.mod", "go 1.21", "go 1.22"), "go howdy 43 second 6 plain"},
		// -gcflags changes the code and no file go list reports, so only the key's arguments can tell.
		{"an argument that changes no file", func() { arguments = []string{"-gcflags=-N -l"} }, "go howdy 43 second 6 plain"},
		{"an argument that changes the files", func() { arguments = []string{"-gcflags=-N -l", "-tags=fancy"} }, "go howdy 43 second 6 fancy"},
		{"the environment", func() { environment = []string{"CGO_CFLAGS=-O2 -g -DANSWER_OFFSET=100"} }, "go howdy 143 second 6 fancy"},
	})
}

// GoBuild keys a header reached through '#cgo CFLAGS: -I' outside the package's directory, so the product must follow
// it too. Go's own build cache doesn't see such a header (its action ID hashes the package's directory, not the
// headers the C compiler opens elsewhere), so go build alone, run again after the header changes, links the object
// built from the old one.
// Not parallel: changes the working directory (t.Chdir) and points the build cache and the store at this test through t.Setenv.
func TestProductIdentityAfterAHeaderOutsideThePackageChanges(t *testing.T) {
	directory := goModule(t, map[string]string{
		"go.mod":                   "module github.com/system-inc/adamic\n\ngo 1.22\n",
		"include/outside.h":        "#define OUTSIDE \"before\"\n",
		"reached/main.go":          "package main\n\n// #cgo CFLAGS: -I${SRCDIR}/../include\n// #include \"outside.h\"\n// static const char *outside(void) { return OUTSIDE; }\nimport \"C\"\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"reached\", C.GoString(C.outside())) }\n",
		"include/unrelated-note.h": "/* in the directory the key hashes, read by nothing */\n",
	})
	var arguments, environment []string
	changeEachInput(t, directory, "reached", "./reached", &arguments, &environment, []inputChange{
		{"nothing (the first build)", func() {}, "reached before"},
		{"a header reached through #cgo CFLAGS: -I", replaceIn(t, directory, "include/outside.h", "before", "after!"), "reached after!"},
		{"it back", replaceIn(t, directory, "include/outside.h", "after!", "before"), "reached before"},
	})
}

// changedOutsideTheLink says whether two builds differ in more than the bytes a macOS link can vary on its own.
func changedOutsideTheLink(before, after []byte) bool {
	if len(before) != len(after) {
		return true
	}
	regions := uuidAndSignature(before)
	for _, span := range differingRanges(before, after) {
		inside := false
		for _, region := range regions {
			inside = inside || span[0] >= region[0] && span[1] < region[1]
		}
		if !inside {
			return true
		}
	}
	return false
}
