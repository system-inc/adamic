package buildcache

import (
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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

// freshGoBuild is 'go build' run directly at root with GoBuild's flags and nothing from buildcache, through goCache,
// a copy of the fresh Go cache template (freshGoCache) that holds nothing of the repository, so nothing the cold side
// cached can stand in for the build.
func freshGoBuild(t *testing.T, root, output, pkg, goCache string, arguments []string, environment ...string) ([]byte, float64) {
	t.Helper()
	started := time.Now()
	path := filepath.Join(t.TempDir(), output)
	command := exec.Command("go", append(append(append(append([]string{"build"}, freshFlags...), arguments...), "-o", path), pkg)...)
	command.Dir = root
	command.Env = append(append(os.Environ(), environment...), "GOCACHE="+goCache)
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
	cache := freshGoCache(t)
	for _, built := range []struct{ output, pkg string }{
		{"hello", "./internal/buildcache/testdata/hello"},
		// withc (cgo) comes back with #c91x2tn: on Linux its fresh and cold links differed in the ELF build id and about 442
		// bytes of the external link's debug info (Oct 9 16:34Z). GoBuild has no callers yet, so no gate product is affected.
	} {
		goCache := cache()
		t.Run(built.output, func(t *testing.T) {
			// Built for this machine alone (ADAMIC_BUILD_STORE=off), keyed by GoInputs' declared files: a fetched
			// product's bytes are its manifest's, checked by hash, so the store's half is the Settle tests'.
			cached(t)
			fresh, freshSeconds := freshGoBuild(t, root, built.output, built.pkg, goCache, nil)

			started := time.Now()
			cold, err := os.ReadFile(GoBuild(t, built.output, built.pkg, nil))
			if err != nil {
				t.Fatal(err)
			}
			coldSeconds := time.Since(started).Seconds()
			t.Logf("product %s fresh=%x cold=%x seconds fresh=%.2f cold=%.2f", built.output, sha256.Sum256(fresh), sha256.Sum256(cold), freshSeconds, coldSeconds)
			identical(t, built.output+" fresh and cold", fresh, cold)
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
	forgetInputs()
	root.once = sync.Once{}
	root.directory = ""
	root.err = nil
}

// freshGoCache gives each fresh build a copy of the fresh Go cache template, so it compiles every package of the
// repository from source without paying for the standard library each time. It fetches the template, so a test calls
// it before cached(t) points the build cache at the test and before goModule moves the repository.
func freshGoCache(t *testing.T) func() string {
	t.Helper()
	template := freshGoCacheTemplate(t)
	return func() string {
		t.Helper()
		directory := filepath.Join(t.TempDir(), "go-build")
		if err := os.CopyFS(directory, os.DirFS(template)); err != nil {
			t.Fatal(err)
		}
		return directory
	}
}

// The fresh Go cache template is built ahead, as a product, so no identity test pays for compiling the standard library
// (7.6 s of CPU on a Mac, about a quarter of each identity test's own work when every test built its own).
func TestProduct_FreshGoCache(t *testing.T) {
	t.Parallel()
	freshGoCacheTemplate(t)
}

// freshGoCacheTemplate is a Go build cache holding fmt and runtime/cgo with their dependencies, compiled with a fresh
// build's flags from an empty cache and holding nothing of any repository: never seeded from the caller's own cache, so
// an object the cold side's go build reuses can't be in the fresh build too and agree with it. Its build reads no file
// of the tree; its key is the build's arguments, go's environment as GoBuild keys it, and the Go and C toolchains.
func freshGoCacheTemplate(t testing.TB) string {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	environment := []string{"GOFLAGS=", "GOWORK=off"}
	settings, values, err := goSettings(root, environment, goEnvironment)
	if err != nil {
		t.Fatal(err)
	}
	toolchain := []string{Tool("go", "version")}
	if compiler := strings.Fields(values["CC"]); len(compiler) > 0 {
		toolchain = append(toolchain, Tool(compiler[0], "--version"))
	}
	arguments := append(append([]string{"build"}, freshFlags...), "fmt", "runtime/cgo")
	directory := Product(t, Inputs{Name: "fresh go cache", Flags: append([]string{"go " + strings.Join(arguments, " ")}, settings...), Toolchain: toolchain}, func(directory string) error {
		cache := filepath.Join(directory, "go-build")
		command := exec.Command("go", arguments...)
		// The product's own directory, outside every module, so the build reads nothing of the tree.
		command.Dir = directory
		command.Env = append(append(os.Environ(), environment...), "GOCACHE="+cache)
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go %s: %v\n%s", strings.Join(arguments, " "), err, combined)
		}
		return reproducibleGoCache(cache)
	})
	return filepath.Join(directory, "go-build")
}

// reproducibleGoCache makes a Go build cache the same bytes from every build of the same packages, as a product must
// be. Go writes each index entry ('v1 <action> <output> <size> <time>', the '-a' files) with the nanosecond it wrote
// it, and trim.txt with when it last trimmed; every output file is already the same build to build under -trimpath
// (three empty-cache builds of fmt and runtime/cgo, Oct 11). Neither time decides a hit (go reads an entry's time only
// for a cached test result, and trims by the files' modification times), so each entry's time is written as 0 and
// trim.txt, which go writes again at its next trim, is left out.
func reproducibleGoCache(cache string) error {
	if err := os.Remove(filepath.Join(cache, "trim.txt")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return filepath.WalkDir(cache, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "-a") {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// The time is the entry's last field, 20 characters wide, before the newline.
		timed := len(content) - len("\n") - 20
		if fields := strings.Fields(string(content)); len(fields) != 5 || fields[0] != "v1" || timed < 1 || content[timed-1] != ' ' || content[len(content)-1] != '\n' {
			return fmt.Errorf("%s isn't a Go build cache entry: %q", path, content)
		}
		return os.WriteFile(path, fmt.Appendf(content[:timed:timed], "%20d\n", 0), 0o644)
	})
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
// Each fresh build goes through its own copy of the fresh Go cache template, which cache makes.
func changeEachInput(t *testing.T, cache func() string, directory, output, pkg string, arguments, environment *[]string, steps []inputChange) {
	t.Helper()
	// Each change is built for this machine alone through one cache kept across the steps, keyed by GoInputs' declared
	// files: the tripwire a measured read set is held against, so a declared key that ignores an input still fails here.
	_, log := cached(t)
	cold := t.TempDir()
	var previous []byte
	for _, step := range steps {
		step.change()
		fresh, _ := freshGoBuild(t, directory, output, pkg, cache(), *arguments, *environment...)
		if got := run(t, filepath.Join(t.TempDir(), output), fresh); got != step.output {
			t.Fatalf("after changing %s a fresh build printed %q, want %q: the test's own expectation is wrong", step.input, got, step.output)
		}
		if previous != nil && !changedOutsideTheLink(previous, fresh) {
			t.Fatalf("changing %s didn't change the product, so this step proves nothing", step.input)
		}
		previous = fresh
		for _, way := range []string{"cold"} {
			t.Setenv("ADAMIC_BUILD_CACHE_DIR", cold)
			before, _ := os.ReadFile(log)
			// Each step changed the tree under this process, which a unit's never does: key it afresh.
			forgetInputs()
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
// step fails, naming it. everyInputSteps is the sequence, in the order its changes apply to everyInputModule.
func everyInputSteps(t *testing.T, directory string, arguments, environment *[]string) []inputChange {
	return []inputChange{
		{"nothing (the first build)", func() {}, "hello 42 first 9 plain"},
		{"the package's Go file", replaceIn(t, directory, "app/main.go", "fmt.Println(", `fmt.Println("go",`), "go hello 42 first 9 plain"},
		{"a dependency's Go file", replaceIn(t, directory, "library/library.go", "index < 3", "index < 4"), "go hello 42 first 16 plain"},
		{"the package's header", replaceIn(t, directory, "app/greeting.h", "hello", "howdy"), "go howdy 42 first 16 plain"},
		{"a C file", replaceIn(t, directory, "app/answer.c", "42", "43"), "go howdy 43 first 16 plain"},
		{"an embedded file", replaceIn(t, directory, "app/message.txt", "first", "second"), "go howdy 43 second 16 plain"},
		{"go.mod", replaceIn(t, directory, "go.mod", "go 1.21", "go 1.22"), "go howdy 43 second 6 plain"},
		// -gcflags changes the code and no file go list reports, so only the key's arguments can tell.
		{"an argument that changes no file", func() { *arguments = []string{"-gcflags=-N -l"} }, "go howdy 43 second 6 plain"},
		{"an argument that changes the files", func() { *arguments = []string{"-gcflags=-N -l", "-tags=fancy"} }, "go howdy 43 second 6 fancy"},
		{"the environment", func() { *environment = []string{"CGO_CFLAGS=-O2 -g -DANSWER_OFFSET=100"} }, "go howdy 143 second 6 fancy"},
	}
}

// The sequence runs as two top-level tests, each its own Loom atom: the ten steps in one test were 31 s alone on one
// core of a Loom box. Each part names its test and its first and last step; TestEveryInputStepRunsInExactlyOnePart
// holds the parts to the whole sequence.
var everyInputParts = []struct{ test, first, last string }{
	{"TestProductIdentityAfterEveryCompiledFileChanges", "nothing (the first build)", "a C file"},
	{"TestProductIdentityAfterEveryOtherInputChanges", "an embedded file", "the environment"},
}

// Not parallel: changes the working directory (t.Chdir) and points the build cache and the store at this test through t.Setenv.
func TestProductIdentityAfterEveryCompiledFileChanges(t *testing.T) {
	changeEveryInputPart(t)
}

// Not parallel: changes the working directory (t.Chdir) and points the build cache and the store at this test through t.Setenv.
func TestProductIdentityAfterEveryOtherInputChanges(t *testing.T) {
	changeEveryInputPart(t)
}

// changeEveryInputPart runs the steps of the part named for this test. A part that starts later in the sequence first
// makes every earlier change, unbuilt, and builds that tree as its own first step, expecting what the step before its
// first printed: its first change then has a product from before it, which a key that ignores the change finds as a
// hit, and a fresh build to differ from, exactly as in the whole sequence.
func changeEveryInputPart(t *testing.T) {
	t.Helper()
	cache := freshGoCache(t)
	directory := goModule(t, everyInputModule)
	var arguments, environment []string
	steps := everyInputSteps(t, directory, &arguments, &environment)
	first, last := everyInputPart(t, steps, t.Name())
	part := steps[first : last+1]
	if first > 0 {
		for _, step := range steps[:first] {
			step.change()
		}
		part = append([]inputChange{{"nothing since " + steps[first-1].input + " (this part's first build)", func() {}, steps[first-1].output}}, part...)
	}
	changeEachInput(t, cache, directory, "app", "./app", &arguments, &environment, part)
}

// everyInputPart is where in steps the part a test runs starts and ends.
func everyInputPart(t *testing.T, steps []inputChange, test string) (int, int) {
	t.Helper()
	at := func(input string) int {
		index := slices.IndexFunc(steps, func(step inputChange) bool { return step.input == input })
		if index < 0 {
			t.Fatalf("no step changes %q", input)
		}
		return index
	}
	for _, part := range everyInputParts {
		if part.test == test {
			return at(part.first), at(part.last)
		}
	}
	t.Fatalf("no part of everyInputParts is %s's", test)
	return 0, 0
}

// Every step of everyInputSteps runs in exactly one part, in order: the parts, one after another, are the whole
// sequence. It builds nothing.
func TestEveryInputStepRunsInExactlyOnePart(t *testing.T) {
	t.Parallel()
	var arguments, environment []string
	steps := everyInputSteps(t, t.TempDir(), &arguments, &environment)
	next := 0
	for _, part := range everyInputParts {
		first, last := everyInputPart(t, steps, part.test)
		if first != next || last < first {
			t.Fatalf("%s runs steps %d to %d, want it to start at %d (%q): a step would run in no part or in two", part.test, first, last, next, steps[min(next, len(steps)-1)].input)
		}
		next = last + 1
	}
	if next != len(steps) {
		t.Fatalf("the parts end at step %d of %d: %q runs in no part", next, len(steps), steps[next].input)
	}
}

// everyInputModule is the repository everyInputSteps changes, one input at a time.
var everyInputModule = map[string]string{
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
}

// GoBuild keys a header reached through '#cgo CFLAGS: -I' outside the package's directory, so the product must follow
// it too. Go's own build cache doesn't see such a header (its action ID hashes the package's directory, not the
// headers the C compiler opens elsewhere), so go build alone, run again after the header changes, links the object
// built from the old one.
// Not parallel: changes the working directory (t.Chdir) and points the build cache and the store at this test through t.Setenv.
func TestProductIdentityAfterAHeaderOutsideThePackageChanges(t *testing.T) {
	cache := freshGoCache(t)
	directory := goModule(t, map[string]string{
		"go.mod":                   "module github.com/system-inc/adamic\n\ngo 1.22\n",
		"include/outside.h":        "#define OUTSIDE \"before\"\n",
		"reached/main.go":          "package main\n\n// #cgo CFLAGS: -I${SRCDIR}/../include\n// #include \"outside.h\"\n// static const char *outside(void) { return OUTSIDE; }\nimport \"C\"\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"reached\", C.GoString(C.outside())) }\n",
		"include/unrelated-note.h": "/* in the directory the key hashes, read by nothing */\n",
	})
	var arguments, environment []string
	changeEachInput(t, cache, directory, "reached", "./reached", &arguments, &environment, []inputChange{
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
