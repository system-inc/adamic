package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// forgetInputs drops the requests this process keyed, for a test that changes the tree or go's environment under the
// process, which a unit never does.
func forgetInputs() {
	computedInputs.Range(func(identity, _ any) bool { computedInputs.Delete(identity); return true })
}

// GoBuild's key names every file the build compiles in this repository, the module files and a cgo header reached
// through -I, and the product runs.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestGoBuildKeysWhatTheBuildCompilesAndRuns(t *testing.T) {
	_, log := cached(t)
	inputs, err := GoInputs("withc", "./internal/buildcache/testdata/withc", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"internal/buildcache/testdata/withc/main.go", "internal/buildcache/testdata/include", "go.mod"} {
		if !slices.Contains(inputs.Files, want) {
			t.Fatalf("%s isn't in the key's files: %v", want, inputs.Files)
		}
	}
	if slices.Contains(inputs.Files, "internal/buildcache/testdata/hello/main.go") {
		t.Fatalf("a package the build doesn't compile is in its key: %v", inputs.Files)
	}
	// A module the main module replaces with a directory of the tree is keyed by its go.mod, though the build compiles
	// none of it (#c2chyab). Mutant: keyInputs without the replacements. Holds where the cohere submodule is checked out.
	if root, err := repositoryRoot(); err == nil && len(existing(filepath.Join(root, "cohere/TypeScript-shim/ast"), "go.mod")) == 1 &&
		!slices.Contains(inputs.Files, "cohere/TypeScript-shim/ast/go.mod") {
		t.Fatalf("a locally replaced module's go.mod isn't in the key's files: %v", inputs.Files)
	}
	if !slices.ContainsFunc(inputs.Flags, func(flag string) bool { return strings.HasPrefix(flag, "CGO_CFLAGS=") }) ||
		!slices.ContainsFunc(inputs.Toolchain, func(tool string) bool { return strings.HasPrefix(tool, "go version: ") }) {
		t.Fatalf("flags %v, toolchain %v", inputs.Flags, inputs.Toolchain)
	}
	// withc reaches testdata/include, outside its own directory, so the key names the package and that directory, and
	// the build lays a header into withc that moves with it (outsideHeaderPlan).
	if want := outsideHeaders + `{"package":"internal/buildcache/testdata/withc","includes":["internal/buildcache/testdata/include"]}`; !slices.Contains(inputs.Flags, want) {
		t.Fatalf("the key doesn't name withc's header outside it, %s: %v", want, inputs.Flags)
	}
	binary := GoBuild(t, "withc", "./internal/buildcache/testdata/withc", nil)
	// GoBuild built it with the reproducible flags, as its recorded inputs say.
	if described, err := os.ReadFile(filepath.Dir(binary) + ".inputs"); err != nil || !strings.Contains(string(described), "flag arguments -trimpath -ldflags=-buildid=") {
		t.Fatalf("the product's inputs: %q, %v", described, err)
	}
	output, err := exec.Command(binary).Output()
	if err != nil || strings.TrimSpace(string(output)) != "hello from a header" {
		t.Fatalf("the product printed %q, %v", output, err)
	}
	// Built once: the second call is a hit.
	GoBuild(t, "withc", "./internal/buildcache/testdata/withc", nil)
	if lines, _ := os.ReadFile(log); strings.Count(string(lines), " miss ") != 1 || strings.Count(string(lines), " hit ") != 1 {
		t.Fatalf("census: %q", lines)
	}
}

// The arguments and the environment change the key; an -overlay, which reads files outside the repository, is refused.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestGoBuildKeysArgumentsAndEnvironment(t *testing.T) {
	cached(t)
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	keyOf := func(arguments []string, environment ...string) string {
		t.Helper()
		inputs, err := GoInputs("hello", "./internal/buildcache/testdata/hello", arguments, environment)
		if err != nil {
			t.Fatal(err)
		}
		return key(t, root, inputs)
	}
	plain := keyOf(nil)
	for name, other := range map[string]string{
		"an argument":     keyOf([]string{"-trimpath"}),
		"the environment": keyOf(nil, "CGO_CFLAGS=-O1 -fsanitize=address"),
		"the compiler":    keyOf(nil, "CC=/usr/bin/no-such-cc"),
	} {
		if other == plain {
			t.Fatalf("%s left the key %s", name, plain)
		}
	}
	// An -overlay whose files are all inside the repository is keyed by its map and their content, not by where the
	// overlay file itself sits (@system_adamic, Oct 9 04:57Z); one reading a file outside the repository is refused.
	overlay := func(replacement string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "overlay.json")
		declared, _ := json.Marshal(map[string]map[string]string{"Replace": {
			filepath.Join(root, "internal/buildcache/testdata/hello/main.go"): replacement}})
		if err := os.WriteFile(path, declared, 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	one := filepath.Join(root, "internal/buildcache/testdata/overlay/one.go.txt")
	two := filepath.Join(root, "internal/buildcache/testdata/overlay/two.go.txt")
	first, again, other := keyOf([]string{"-overlay=" + overlay(one)}), keyOf([]string{"-overlay", overlay(one)}), keyOf([]string{"-overlay=" + overlay(two)})
	if first != again {
		t.Fatalf("the same overlay at two temporary paths keyed %s and %s", first, again)
	}
	if first == other || first == plain {
		t.Fatalf("a replacement's content didn't move the key: %s, %s, plain %s", first, other, plain)
	}
	// The replacement is an input file, hashed by content like any other, not just a name in the map.
	inputs, err := GoInputs("hello", "./internal/buildcache/testdata/hello", []string{"-overlay=" + overlay(one)}, nil)
	if err != nil || !slices.Contains(inputs.Files, "internal/buildcache/testdata/overlay/one.go.txt") {
		t.Fatalf("the overlay's replacement isn't among the key's files: %v %v", inputs.Files, err)
	}
	outside := filepath.Join(t.TempDir(), "main.go")
	os.WriteFile(outside, []byte("package main\n\nfunc main() {}\n"), 0o644)
	if _, err := GoInputs("hello", "./internal/buildcache/testdata/hello", []string{"-overlay=" + overlay(outside)}, nil); err == nil ||
		!strings.Contains(err.Error(), "outside the repository") {
		t.Fatalf("an overlay reading a file outside the repository was keyed: %v", err)
	}
}

// A runner asks for the product Workshop built (#t37sw0f): the same tree checked out at two absolute paths, built with
// two -p shares (the core count each machine computes into GOFLAGS, and one passed as an argument) and pointed at by
// three workspace files (found in the tree, named in the tree, and one outside it using the same directory), keys
// byte-identically, and its recorded inputs are byte-identical too, so the key's equality is the inputs' and not an
// accident. A source byte, a build tag and the Go release each move it.
// Not parallel: changes the working directory (t.Chdir) and the environment (t.Setenv).
func TestGoBuildKeysTheSameTreeTheSameFromAnyMachine(t *testing.T) {
	cached(t)
	tree := map[string]string{
		"go.mod":        "module github.com/system-inc/adamic\n\ngo 1.21\n",
		"go.work":       "go 1.21\n\nuse .\n",
		"app/main.go":   "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\", tagged) }\n",
		"app/plain.go":  "//go:build !fancy\n\npackage main\n\nconst tagged = \"plain\"\n",
		"app/fancy.go":  "//go:build fancy\n\npackage main\n\nconst tagged = \"fancy\"\n",
		"unrelated.txt": "read by nothing\n",
	}
	base := t.TempDir()
	checkout := func(path string) string {
		directory := filepath.Join(base, path)
		for name, content := range tree {
			write(t, directory, name, content)
		}
		return directory
	}
	workshop, runner := checkout("home/ahra/work/adamic"), checkout("srv/runner/7/elsewhere/adamic")
	outside := filepath.Join(base, "loom", "go.work")
	write(t, base, "loom/go.work", "go 1.21\n\nuse "+runner+"\n")
	type machine struct{ directory, workspace, flags string }
	keyed := func(on machine, arguments ...string) (string, string) {
		t.Helper()
		inRepository(t, on.directory)
		t.Setenv("GOWORK", on.workspace)
		t.Setenv("GOFLAGS", on.flags)
		inputs, err := GoInputs("app", "./app", reproducible(arguments), nil)
		if err != nil {
			t.Fatal(err)
		}
		root, err := repositoryRoot()
		if err != nil {
			t.Fatal(err)
		}
		return key(t, root, inputs), describe(root, inputs)
	}
	reference, described := keyed(machine{workshop, filepath.Join(workshop, "go.work"), "-buildvcs=false -trimpath -p=12"})
	for name, other := range map[string]machine{
		"another checkout, its go.work found, another -p share first": {runner, "", "-p=3 -buildvcs=false -trimpath"},
		"another checkout, a workspace file outside it using it":      {runner, outside, "-buildvcs=false -p=7 -trimpath"},
	} {
		if key, description := keyed(other); key != reference || description != described {
			t.Errorf("%s keyed %s, Workshop %s; inputs\n%s\nagainst Workshop's\n%s", name, key, reference, description, described)
		}
	}
	plain := machine{runner, "", "-buildvcs=false -trimpath"}
	if key, _ := keyed(plain, "-p", "4"); key != reference {
		t.Errorf("a -p 4 argument moved the key to %s from %s", key, reference)
	}
	if strings.Contains(described, base) || strings.Contains(described, "-p=") {
		t.Errorf("the recorded inputs name a machine:\n%s", described)
	}

	// What changes the product moves the key: a build tag, a source byte, the Go release.
	if key, _ := keyed(plain, "-tags=fancy"); key == reference {
		t.Error("-tags=fancy left the key the same")
	}
	write(t, runner, "app/main.go", strings.Replace(tree["app/main.go"], "hello", "howdy", 1))
	if key, _ := keyed(plain); key == reference {
		t.Error("a changed source byte left the key the same")
	}
	write(t, runner, "app/main.go", tree["app/main.go"])
	if key, _ := keyed(plain); key != reference {
		t.Fatalf("the source put back keyed %s, not %s", key, reference)
	}
	// Another release, as a go that reports go1.99.0 and otherwise is this one (the release keys twice: Tool's go
	// version and GOTOOLCHAIN's selection, which is go env's GOVERSION). The key runs the go toolchain.Path finds, a
	// tools directory's before PATH's, so the fake stands in a tools directory of the test's own.
	real, err := toolchain.Path("go")
	if err != nil {
		t.Fatal(err)
	}
	standIn := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(standIn, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(standIn, "go")
	os.WriteFile(fake, []byte("#!/bin/sh\ncase \"$1\" in\nversion) echo 'go version go1.99.0 fake/arch'; exit 0 ;;\n"+
		"env) "+real+" \"$@\" | sed -E 's/\"GOVERSION\": \"[^\"]*\"/\"GOVERSION\": \"go1.99.0\"/'; exit 0 ;;\nesac\nexec "+real+" \"$@\"\n"), 0o755)
	t.Setenv("ADAMIC_TOOLS", filepath.Dir(standIn))
	forget := func() { tools.Range(func(name, _ any) bool { tools.Delete(name); return true }) }
	forget()
	t.Cleanup(forget)
	if key, description := keyed(plain); key == reference || !strings.Contains(description, "GOTOOLCHAIN selects go1.99.0") {
		t.Errorf("another Go release keyed %s against %s:\n%s", key, reference, description)
	}
}

// GoBuild builds the same bytes from any checkout path, so a product another machine published audits clean: the cgo
// package, copied into two module directories at different paths and built with GoBuild's flags, is byte-identical.
func TestGoBuildIsReproducibleAcrossCheckoutPathsAndCommits(t *testing.T) {
	t.Parallel()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	var products [][]byte
	for _, name := range []string{"one", "elsewhere/two"} {
		module := filepath.Join(t.TempDir(), name)
		for _, file := range []string{"withc/main.go", "include/greeting.h"} {
			content, err := os.ReadFile(filepath.Join(root, "internal/buildcache/testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			os.MkdirAll(filepath.Join(module, filepath.Dir(file)), 0o755)
			os.WriteFile(filepath.Join(module, file), content, 0o644)
		}
		os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.com/reproducible\n\ngo 1.21\n"), 0o644)
		// Each checkout is its own repository at its own commit, as two gate trees are: Go stamps a binary built in a
		// repository with the commit unless -buildvcs=false.
		for _, git := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "commit " + name}} {
			if output, err := toolchain.GitCommand(append([]string{"-C", module}, git...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", git, err, output)
			}
		}
		command := toolchain.CommandIn(module, "go", append(append([]string{"build"}, reproducible(nil)...), "-o", filepath.Join(module, "out"), "./withc")...)
		command.Env = append(command.Env, "GOWORK=off", "GOFLAGS=")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, output)
		}
		content, err := os.ReadFile(filepath.Join(module, "out"))
		if err != nil {
			t.Fatal(err)
		}
		products = append(products, content)
	}
	// On macOS the system linker may write a different LC_UUID for the same objects (#cchsq45): identical accepts and
	// logs a difference confined to it and the code signature there, and nowhere else.
	if !identical(t, "two checkout paths at two commits", products[0], products[1]) {
		t.Fatalf("two checkout paths at two commits built %x and %x", sha256.Sum256(products[0]), sha256.Sum256(products[1]))
	}
	arguments := reproducible([]string{"-buildmode=c-archive"})
	if !slices.Equal(arguments, []string{"-trimpath", "-ldflags=-buildid=", "-buildvcs=false", "-buildmode=c-archive"}) {
		t.Fatalf("arguments %v", arguments)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("a caller's own -ldflags was accepted")
		}
	}()
	reproducible([]string{"-ldflags=-s"})
}

// A Go oracle a product test builds by hand, go test -c over a module with a harness its overlay lays in (css's
// css-printer-oracle), is the same bytes from any checkout path with ReproducibleGoFlags, and isn't without them, so the
// flags are what makes it so (#j8wf2w4: every hand-rolled oracle held its checkout's path). GOENV=off keeps a machine's
// own go env file, which may set -trimpath itself (Kirk's Mac does), out of both builds.
func TestAHandBuiltOracleIsTheSameFromAnyCheckoutPathOnlyWithTheReproducibleFlags(t *testing.T) {
	t.Parallel()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	build := func(flags []string) [][]byte {
		var products [][]byte
		for _, name := range []string{"one", "elsewhere/two"} {
			module := filepath.Join(t.TempDir(), name)
			for _, file := range []string{"module/go.mod", "module/greet/greet.go", "module/greet/greet_test.go", "module-side/side_test.go.txt"} {
				content, err := os.ReadFile(filepath.Join(root, "internal/buildcache/testdata", file))
				if err != nil {
					t.Fatal(err)
				}
				os.MkdirAll(filepath.Join(module, filepath.Dir(file)), 0o755)
				os.WriteFile(filepath.Join(module, file), content, 0o644)
			}
			overlay := filepath.Join(module, "overlay.json")
			declared, _ := json.Marshal(map[string]map[string]string{"Replace": {filepath.Join(module, "module/greet/adamic_side_test.go"): filepath.Join(module, "module-side/side_test.go.txt")}})
			os.WriteFile(overlay, declared, 0o644)
			command := toolchain.CommandIn(filepath.Join(module, "module"), "go", append(append(append([]string{"test"}, flags...), "-c", "-overlay="+overlay, "-o", filepath.Join(module, "oracle")), "./greet")...)
			command.Env = append(command.Env, "GOENV=off", "GOWORK=off", "GOFLAGS=")
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("%v\n%s", err, output)
			}
			content, err := os.ReadFile(filepath.Join(module, "oracle"))
			if err != nil {
				t.Fatal(err)
			}
			products = append(products, content)
		}
		return products
	}
	if products := build(ReproducibleGoFlags()); !identical(t, "a hand-built oracle at two checkout paths", products[0], products[1]) {
		t.Fatalf("a hand-built oracle at two checkout paths built %x and %x", sha256.Sum256(products[0]), sha256.Sum256(products[1]))
	}
	if products := build(nil); bytes.Equal(products[0], products[1]) {
		t.Fatal("without the reproducible flags, the oracle at two checkout paths was the same bytes: this test can't tell the flags did it")
	}
}

// GoTest builds a module's test binary as a product, in the module's own directory, with a harness the overlay adds
// (as a port lays its Go oracle over cohere). Its key names the package's test files, the module's go.mod and the
// harness's content, never the harness's virtual path, which isn't on disk; the binary runs the harness; and a second
// ask is a hit.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestGoTestKeysAModulesTestBinary(t *testing.T) {
	_, log := cached(t)
	forgetInputs()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	module := "internal/buildcache/testdata/module"
	virtual := filepath.Join(root, module, "greet/adamic_side_test.go")
	side := filepath.Join(root, "internal/buildcache/testdata/module-side/side_test.go.txt")
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	declared, _ := json.Marshal(map[string]map[string]string{"Replace": {virtual: side}})
	if err := os.WriteFile(overlay, declared, 0o644); err != nil {
		t.Fatal(err)
	}
	request := goRequest{verb: "test -c", module: module, output: "greet.test", pkg: "./greet", arguments: reproducible([]string{"-overlay=" + overlay}), environment: []string{"GOWORK=off"}}
	inputs, err := request.inputs()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{module + "/greet/greet.go", module + "/greet/greet_test.go", module + "/go.mod", "internal/buildcache/testdata/module-side/side_test.go.txt"} {
		if !slices.Contains(inputs.Files, want) {
			t.Fatalf("%s isn't in the key's files: %v", want, inputs.Files)
		}
	}
	if slices.Contains(inputs.Files, module+"/greet/adamic_side_test.go") {
		t.Fatalf("the harness's virtual path, which isn't on disk, is keyed as a file: %v", inputs.Files)
	}
	binary := GoTest(t, module, "greet.test", "./greet", []string{"-overlay=" + overlay}, "GOWORK=off")
	output, err := exec.Command(binary, "-test.run=^TestSide$", "-test.v").Output()
	if err != nil || !strings.Contains(string(output), "side says hello") {
		t.Fatalf("the test binary ran %q, %v", output, err)
	}
	GoTest(t, module, "greet.test", "./greet", []string{"-overlay=" + overlay}, "GOWORK=off")
	if lines, _ := os.ReadFile(log); strings.Count(string(lines), " miss ") != 1 || strings.Count(string(lines), " hit ") != 1 {
		t.Fatalf("census: %q", lines)
	}
	// Without the harness it's another test binary.
	plain, err := goRequest{verb: "test -c", module: module, output: "greet.test", pkg: "./greet", arguments: reproducible(nil), environment: []string{"GOWORK=off"}}.inputs()
	if err != nil {
		t.Fatal(err)
	}
	if key(t, root, plain) == key(t, root, inputs) {
		t.Fatal("the harness left the key as it was")
	}
}

// A go product's plan holds its request's -overlay as the map it declares, never the file (#mmpt615, Mentat's M1): the
// test writes the file before the build, in a directory of its own, and run writes its own overlay inside the window,
// so go reads no file of the test's there that the window would count as an input. Two plans of one request are the
// same, so the plan is a function of the request and what its key holds. Mutant: the request's -overlay kept among the
// plan's arguments.
func TestAGoRequestsOverlayIsItsPlansMap(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	replacement := filepath.Join(root, "internal/buildcache/testdata/module/greet/greet.go")
	if err := os.WriteFile(overlay, []byte(`{"Replace":{"greet/extra.go":"`+replacement+`"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	request := goRequest{verb: "test -c", module: "internal/buildcache/testdata/module", output: "greet.test", pkg: "./greet"}
	for _, arguments := range [][]string{{"-overlay=" + overlay}, {"-trimpath", "-overlay", overlay}} {
		request.arguments = arguments
		parts, err := request.planned(root, Inputs{})
		if err != nil {
			t.Fatalf("arguments %q: %v", arguments, err)
		}
		if slices.ContainsFunc(parts.command.Arguments, func(argument string) bool { return strings.HasPrefix(argument, "-overlay") || argument == overlay }) {
			t.Errorf("arguments %q: the plan runs go with the test's overlay file: %q", arguments, parts.command.Arguments)
		}
		if want := map[string]string{"greet/extra.go": replacement}; !maps.Equal(parts.replace, want) {
			t.Errorf("arguments %q: the plan's overlay is %q, want %q", arguments, parts.replace, want)
		}
		if again := request.plan(root, Inputs{}); !reflect.DeepEqual(request.plan(root, Inputs{}), again) {
			t.Errorf("arguments %q: two plans of one request differ:\n%+v\n%+v", arguments, request.plan(root, Inputs{}), again)
		}
	}
	request.arguments = []string{"-trimpath"}
	if parts, err := request.planned(root, Inputs{}); err != nil || parts.replace != nil {
		t.Errorf("a request without an overlay plans %q, %v", parts.replace, err)
	}
}

// A process keys a request once, and a planner's index keys it with no go at all (#pc0jvv4): what a tree build records
// (ADAMIC_BUILD_INPUTS_RECORD) a unit reads (ADAMIC_BUILD_INPUTS), and with neither it keys with go as before.
// Not parallel: points the build cache and the store at this test, and PATH and ADAMIC_TOOLS, through t.Setenv.
func TestARequestIsKeyedOnceAndCanBeKeyedAhead(t *testing.T) {
	cached(t)
	forgetInputs()
	record := filepath.Join(t.TempDir(), "inputs.jsonl")
	t.Setenv("ADAMIC_BUILD_INPUTS_RECORD", record)
	first, err := GoInputs("hello", "./internal/buildcache/testdata/hello", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again, err := GoInputs("hello", "./internal/buildcache/testdata/hello", nil, nil); err != nil || !slices.Equal(again.Files, first.Files) {
		t.Fatalf("asked again: %v, %v", again, err)
	}
	if lines, _ := os.ReadFile(record); strings.Count(string(lines), "\n") != 1 {
		t.Fatalf("a request keyed twice in one process was recorded %d times", strings.Count(string(lines), "\n"))
	}
	t.Setenv("ADAMIC_BUILD_INPUTS_RECORD", "")
	forgetInputs()
	// No go anywhere toolchain.Path looks: no tools directory, and nothing on PATH.
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("PATH", t.TempDir())
	if _, err := GoInputs("hello", "./internal/buildcache/testdata/hello", nil, nil); err == nil {
		t.Fatal("keyed with no go and no index")
	}
	t.Setenv("ADAMIC_BUILD_INPUTS", record)
	ahead, err := GoInputs("hello", "./internal/buildcache/testdata/hello", nil, nil)
	if err != nil {
		t.Fatalf("the index didn't key the request: %v", err)
	}
	if !slices.Equal(ahead.Files, first.Files) || !slices.Equal(ahead.Flags, first.Flags) || ahead.Name != first.Name {
		t.Fatalf("the index keyed %+v, go keyed %+v", ahead, first)
	}
}

// A test binary keys the same from two checkouts of one tree at different paths, as GoBuild's do: the module, its
// test file and the harness laid over it, with the overlay file at another temporary path in each.
// Not parallel: changes the working directory (t.Chdir) and the environment (t.Setenv).
func TestGoTestKeysTheSameTreeTheSameFromAnyPath(t *testing.T) {
	cached(t)
	tree := map[string]string{
		"go.mod":                       "module github.com/system-inc/adamic\n\ngo 1.21\n",
		"cohere/go.mod":                "module example.com/cohere\n\ngo 1.21\n",
		"cohere/format/format.go":      "package format\n\nfunc Format(s string) string { return \"[\" + s + \"]\" }\n",
		"cohere/format/format_test.go": "package format\n\nimport \"testing\"\n\nfunc TestFormat(t *testing.T) { _ = Format(\"x\") }\n",
		"port/side_test.go.txt":        "package format\n\nimport \"testing\"\n\nfunc TestSide(t *testing.T) { t.Log(Format(\"side\")) }\n",
	}
	keyed := func(directory string) (string, Inputs) {
		t.Helper()
		for name, content := range tree {
			write(t, directory, name, content)
		}
		inRepository(t, directory)
		overlay := filepath.Join(t.TempDir(), "overlay.json")
		declared, _ := json.Marshal(map[string]map[string]string{"Replace": {
			filepath.Join(directory, "cohere/format/adamic_side_test.go"): filepath.Join(directory, "port/side_test.go.txt")}})
		if err := os.WriteFile(overlay, declared, 0o644); err != nil {
			t.Fatal(err)
		}
		inputs, err := goRequest{verb: "test -c", module: "cohere", output: "format.test", pkg: "./format",
			arguments: reproducible([]string{"-overlay=" + overlay}), environment: []string{"GOWORK=off"}}.inputs()
		if err != nil {
			t.Fatal(err)
		}
		return key(t, directory, inputs), inputs
	}
	base := t.TempDir()
	one, first := keyed(filepath.Join(base, "work", "adamic"))
	two, second := keyed(filepath.Join(base, "elsewhere", "srv", "7", "adamic"))
	if one != two {
		t.Fatalf("one tree at two paths keyed %s and %s:\n%+v\n%+v", one, two, first, second)
	}
	if !slices.Contains(first.Files, "cohere/format/format_test.go") || !slices.Contains(first.Files, "port/side_test.go.txt") {
		t.Fatalf("the key's files: %v", first.Files)
	}
}

// GoBuildIn builds in a module of its own a main its overlay synthesizes, which isn't on disk: keyed through the
// overlay, it runs, and its name differs from a root build's.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestGoBuildInBuildsASynthesizedMainInAModule(t *testing.T) {
	cached(t)
	forgetInputs()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	module := "internal/buildcache/testdata/module"
	virtual := filepath.Join(root, module, "cmd/greeter/main.go")
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	declared, _ := json.Marshal(map[string]map[string]string{"Replace": {virtual: filepath.Join(root, "internal/buildcache/testdata/module-side/main.go.txt")}})
	if err := os.WriteFile(overlay, declared, 0o644); err != nil {
		t.Fatal(err)
	}
	binary := GoBuildIn(t, module, "greeter", virtual, []string{"-overlay=" + overlay}, "GOWORK=off")
	output, err := exec.Command(binary).Output()
	if err != nil || strings.TrimSpace(string(output)) != "synthesized hello" {
		t.Fatalf("the synthesized main printed %q, %v", output, err)
	}
	inputs, err := goRequest{verb: "build", module: module, output: "greeter", pkg: virtual, arguments: reproducible([]string{"-overlay=" + overlay}), environment: []string{"GOWORK=off"}}.inputs()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(inputs.Name, "go build "+module+" ") || !slices.Contains(inputs.Files, module+"/greet/greet.go") || slices.Contains(inputs.Files, module+"/cmd/greeter/main.go") {
		t.Fatalf("name %q, files %v", inputs.Name, inputs.Files)
	}
}

// A mutation is planted into the product and keyed by what it plants (#ebc67r2): into a package's source or into an
// overlay's replacement, never into the tree. The binary runs the mutant, the unmutated product is another, a changed
// after-string is another again, the same mutation asked twice is a hit, and a mutation that would plant nothing or
// plant into a file the build doesn't read is refused.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestGoTestMutatedPlantsAndKeysEachMutation(t *testing.T) {
	_, log := cached(t)
	forgetInputs()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	module := "internal/buildcache/testdata/module"
	source := module + "/greet/greet.go"
	before, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		t.Fatal(err)
	}
	sideFile := "internal/buildcache/testdata/module-side/side_test.go.txt"
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	declared, _ := json.Marshal(map[string]map[string]string{"Replace": {filepath.Join(root, module, "greet/adamic_side_test.go"): filepath.Join(root, sideFile)}})
	if err := os.WriteFile(overlay, declared, 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-overlay=" + overlay}
	run := func(binary, test string) string {
		output, _ := exec.Command(binary, "-test.run=^"+test+"$", "-test.v").CombinedOutput()
		return string(output)
	}
	howdy := []Mutation{{File: source, Before: `return "hello"`, After: `return "howdy"`}}
	mutant := GoTestMutated(t, module, "greet.test", "./greet", arguments, howdy, "GOWORK=off")
	if output := run(mutant, "TestGreeting"); !strings.Contains(output, "FAIL") || !strings.Contains(output, "howdy") {
		t.Fatalf("the mutant's TestGreeting ran %q, want it failing on howdy", output)
	}
	if output := run(mutant, "TestSide"); !strings.Contains(output, "side says howdy") {
		t.Fatalf("the mutant's harness ran %q", output)
	}
	if after, _ := os.ReadFile(filepath.Join(root, source)); string(after) != string(before) {
		t.Fatal("planting a mutation changed the tree")
	}
	if output := run(GoTest(t, module, "greet.test", "./greet", arguments, "GOWORK=off"), "TestGreeting"); !strings.Contains(output, "PASS") {
		t.Fatalf("the unmutated TestGreeting ran %q", output)
	}
	said := []Mutation{{File: sideFile, Before: `"side says"`, After: `"side said"`}}
	if output := run(GoTestMutated(t, module, "greet.test", "./greet", arguments, said, "GOWORK=off"), "TestSide"); !strings.Contains(output, "side said hello") {
		t.Fatalf("a mutated overlay replacement ran %q", output)
	}
	GoTestMutated(t, module, "greet.test", "./greet", arguments, howdy, "GOWORK=off")
	if lines, _ := os.ReadFile(log); strings.Count(string(lines), " miss ") != 3 || strings.Count(string(lines), " hit ") != 1 {
		t.Fatalf("census: %q", lines)
	}

	keyed := func(mutations []Mutation) (Inputs, error) {
		return goRequest{verb: "test -c", module: module, output: "greet.test", pkg: "./greet", arguments: reproducible(arguments), mutations: mutations, environment: []string{"GOWORK=off"}}.inputs()
	}
	keyOf := func(mutations []Mutation) string {
		t.Helper()
		inputs, err := keyed(mutations)
		if err != nil {
			t.Fatal(err)
		}
		return key(t, root, inputs)
	}
	hiya := []Mutation{{File: source, Before: `return "hello"`, After: `return "hiya"`}}
	if keyOf(howdy) == keyOf(hiya) {
		t.Fatal("a changed after-string left the key as it was")
	}
	if keyOf(howdy) == keyOf(nil) {
		t.Fatal("the mutation left the key as it was")
	}
	for _, refused := range [][]Mutation{
		{{File: source, Before: `return "goodbye"`, After: `return "x"`}},
		{{File: source, Before: "", After: "x"}},
		{{File: "internal/buildcache/testdata/hello/main.go", Before: "hello", After: "x"}},
	} {
		if _, err := keyed(refused); err == nil {
			t.Errorf("%+v was keyed, want it refused", refused)
		}
	}
}

// GoExports' export data is the same at every commit that changes nothing it compiles: a test main is package main,
// and go stamps the commit into every one it compiles unless -buildvcs=false, so its export data's build id moved at
// each commit under an unmoved key (landable-8, internal/ir: 4 of 335 files, each a test main's). Built uncached here
// at two commits a note apart. Mutant: GoExports' go list without -buildvcs=false.
// Not parallel: inRepository changes the working directory and repositoryRoot.
func TestGoExportsAreTheSameAtEveryCommit(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.22\n")
	write(t, root, "tool/main.go", "package main\n\nfunc main() {}\n")
	write(t, root, "tool/main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestNothing(t *testing.T) {}\n")
	write(t, root, "notes.txt", "one\n")
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-q", "-m", "one")
	inRepository(t, root)
	cached(t)
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	// GoExports itself, whatever the environment a gate gives it.
	t.Setenv("GOFLAGS", "")
	testMain := func() []byte {
		t.Helper()
		directory := GoExports(t, "./tool")
		content, err := os.ReadFile(filepath.Join(directory, "listing.json"))
		if err != nil {
			t.Fatal(err)
		}
		var listing []ListedPackage
		if err := json.Unmarshal(content, &listing); err != nil {
			t.Fatal(err)
		}
		for _, listed := range listing {
			if listed.ImportPath == "github.com/system-inc/adamic/tool.test" && listed.Export != "" {
				data, err := os.ReadFile(filepath.Join(directory, listed.Export))
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
		}
		t.Fatalf("the listing holds no test main's export data: %s", content)
		return nil
	}
	first := testMain()
	write(t, root, "notes.txt", "two\n")
	gitIn(t, root, "commit", "-q", "-am", "two")
	if second := testMain(); !slices.Equal(first, second) {
		t.Fatal("a test main's export data differs between two commits that change nothing it compiles")
	}
}

// GoDirectExports keeps the named package, its tests and what they import directly, never the rest of -deps, while
// GoExports keeps all of it (#mtvmz1w), and the two never share a key. tool imports lib/b, whose signature returns
// lib/c's type, and tool's test imports lib/d. Mutants that each fail it: GoDirectExports keeping all of -deps; a test
// file's import dropped; GoDirectExports' verb GoExports' own, so one product would answer both.
// Not parallel: inRepository changes the working directory and repositoryRoot.
func TestGoDirectExportsKeepOnlyWhatTheNamedPackagesImport(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.22\n")
	write(t, root, "lib/c/c.go", "package c\n\ntype T struct{ Field int }\n")
	write(t, root, "lib/b/b.go", "package b\n\nimport \"github.com/system-inc/adamic/lib/c\"\n\nfunc F() c.T { return c.T{} }\n")
	write(t, root, "lib/d/d.go", "package d\n\nconst D = 1\n")
	write(t, root, "tool/main.go", "package main\n\nimport \"github.com/system-inc/adamic/lib/b\"\n\nfunc main() { _ = b.F().Field }\n")
	write(t, root, "tool/main_test.go", "package main\n\nimport (\n\t\"testing\"\n\n\t\"github.com/system-inc/adamic/lib/d\"\n)\n\nfunc TestNothing(t *testing.T) { _ = d.D }\n")
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-q", "-m", "one")
	inRepository(t, root)
	cached(t)
	t.Setenv("GOFLAGS", "")
	listed := func(directory string) []string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(directory, "listing.json"))
		if err != nil {
			t.Fatal(err)
		}
		var listing []ListedPackage
		if err := json.Unmarshal(content, &listing); err != nil {
			t.Fatal(err)
		}
		paths := []string{}
		for _, one := range listing {
			paths = append(paths, strings.TrimPrefix(one.ImportPath, "github.com/system-inc/adamic/"))
		}
		return paths
	}
	direct, whole := GoDirectExports(t, "./tool"), GoExports(t, "./tool")
	if direct == whole {
		t.Fatalf("GoDirectExports and GoExports answered one product, %s", direct)
	}
	kept, all := listed(direct), listed(whole)
	if !slices.Contains(kept, "lib/b") || !slices.Contains(kept, "lib/d") || slices.Contains(kept, "lib/c") {
		t.Fatalf("GoDirectExports listed %v; want tool's direct imports lib/b and lib/d, never lib/c", kept)
	}
	if !slices.Contains(all, "lib/c") {
		t.Fatalf("GoExports listed %v; want all of -deps, lib/c included", all)
	}
}

// A tool written a moment ago is asked again while it is busy (ETXTBSY: a descriptor open for writing, as a process
// forked by a parallel test holds one until it execs), never named by the error and its path. Mutant: toolValue
// running the tool once.
func TestABusyToolIsAskedAgain(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("exec refuses a file open for writing on Linux; elsewhere a tool is never busy this way")
	}
	tool := filepath.Join(t.TempDir(), "busytool")
	writer, err := os.OpenFile(tool, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString("#!/bin/sh\necho 'busytool 1.0'\n"); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		writer.Close()
	}()
	if got := Tool(tool, "--version"); got != "busytool --version: busytool 1.0" {
		t.Fatalf("a tool busy for a moment reported %q", got)
	}
}
