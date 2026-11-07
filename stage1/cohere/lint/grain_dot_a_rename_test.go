package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testDotARenameShards = 2

var dotARenameCases = []string{"no-var-a-to-ts:original", "no-var-a-to-ts:renamed"}

func dotARenameOwner(key string) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(key))
	return int(hash.Sum64() % testDotARenameShards)
}

// The original pinned corpus is one rename and its first no-var witness.
// Builds are immutable content-addressed products, never shard-local builds.
type dotARenameProducts struct {
	Original, Changed, Oracle, OriginalNative, ChangedNative, OriginalJS, ChangedJS string
	Before, After                                                                   []byte
}

func dotARenameInputs(test *testing.T, name string, flags []string) buildcache.Inputs {
	test.Helper()
	files := append(rulesAgreeSourceInputs(test), "stage1/cohere/lint/testdata/oracle.go")
	// copyPort also copies witnesses and option sidecars; those enter its key.
	for _, descriptor := range prepareRegistry(test, packageDirectory) {
		directory := filepath.Join("stage1/cohere/lint/rules", descriptor.Slug, "testdata")
		if _, statError := os.Stat(filepath.Join(repository, directory)); statError == nil {
			files = append(files, filepath.ToSlash(directory))
		} else if !os.IsNotExist(statError) {
			test.Fatal(statError)
		}
	}
	return buildcache.Inputs{Name: "dot-a-rename-" + name, Files: files, Flags: flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), buildcache.Tool(
			"go",
			"env",
			"GOOS",
			"GOARCH",
			"CGO_ENABLED",
			"GOEXPERIMENT",
			"CC",
			"CXX",
			"CGO_CFLAGS",
			"CGO_LDFLAGS",
		), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}}
}

// Each declaration and filtered shard uses the same once-per-process recipe.
var dotARenameProductOnce [6]sync.Once
var dotARenameProductDirectories [6]string

func dotARenameProduct(test *testing.T, slot int, inputs buildcache.Inputs, recipe func(string) error) string {
	test.Helper()
	dotARenameProductOnce[slot].Do(func() {
		dotARenameProductDirectories[slot] = buildcache.Product(test, inputs, recipe)
	})
	if dotARenameProductDirectories[slot] == "" {
		test.Fatal("DotARename product preparation failed")
	}
	return dotARenameProductDirectories[slot]
}

func dotARenameSource(test *testing.T) string {
	test.Helper()
	product := dotARenameProduct(test, 0, dotARenameInputs(
		test,
		"renamed-source",
		[]string{"rules/no-var/rule.a -> rule.ts", packageDirectory},
	), func(dir string) error {
		copyPort(test, dir, "", "")
		if renameError := os.Rename(
			filepath.Join(dir, "rules/no-var/rule.a"),
			filepath.Join(dir, "rules/no-var/rule.ts"),
		); renameError != nil {
			return renameError
		}
		// The port's imports name this checkout; the product names it <repository> (#tqrqx60).
		return buildcache.RelativeFiles(dir)
	})
	// The loader and Node read the port, so callers get the copy whose imports name this checkout.
	directory, resolvedError := buildcache.Resolved(product)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	return directory
}

// dotARenameRegistered is a private copy of the renamed source as Resolved reads it, with its registry generated. The
// source carries the registry copyPort generated for rule.a, so generating it for rule.ts writes .generated/registry.ts
// and registry.go: in place, that is into Resolved's copy, which every reader of the source on this box shares, or into
// the product itself when Resolved answers with it, and a second reader would read the first one's registry (#0sqk6ce).
// Each reader generates into its own copy instead, and the source stays as it was built.
func dotARenameRegistered(test *testing.T, source string) string {
	test.Helper()
	directory := test.TempDir()
	if copyFsError := os.CopyFS(directory, os.DirFS(source)); copyFsError != nil {
		test.Fatal(copyFsError)
	}
	prepareRegistry(test, directory)
	return directory
}

func dotARenameLowered(test *testing.T, renamed bool) string {
	test.Helper()
	name, directory, slot := "original", packageDirectory, 2
	if renamed {
		name, directory, slot = "renamed", dotARenameSource(test), 3
	}
	inputs := dotARenameInputs(test, name+"-lowered", []string{"native.C", "javascript.JavaScript", name})
	return dotARenameProduct(test, slot, inputs, func(dir string) error {
		port := directory
		if renamed {
			port = dotARenameRegistered(test, directory)
		} else {
			prepareRegistry(test, port)
		}
		phase := time.Now()
		program, loadError := load.Load([]string{filepath.Join(port, "main.ts")})
		if loadError != nil {
			return loadError
		}
		test.Logf("%s load: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		result, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		test.Logf("%s lower: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		if writeError := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(native.C(result)), 0644); writeError != nil {
			return writeError
		}
		test.Logf("%s native emit: %.3fs", name, time.Since(phase).Seconds())
		phase = time.Now()
		defer func() { test.Logf("%s JavaScript emit: %.3fs", name, time.Since(phase).Seconds()) }()
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(result)), 0644)
	})
}

func dotARenameNative(test *testing.T, renamed bool) string {
	test.Helper()
	name, slot := "original", 4
	if renamed {
		name, slot = "renamed", 5
	}
	lowered := dotARenameLowered(test, renamed)
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	inputs := dotARenameInputs(test, name+"-native", append(native.Flags(options), name, "Split=true", "Jobs=4"))
	return dotARenameProduct(test, slot, inputs, func(dir string) error {
		data, readError := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if readError != nil {
			return readError
		}
		started := time.Now()
		defer func() { test.Logf("%s clang: %.3fs", name, time.Since(started).Seconds()) }()
		return native.Build(string(data), filepath.Join(dir, "scanner"), options)
	})
}

// Every product the shards read is declared, so Workshop's loom build-tree builds each once and a shard only reads
// them. C emission is over the 60 s a test may take, never the 600 s a product may (Loom's KindCeilings): measured
// Oct 10 inside the shards at 97% busy, lowering 71 s and the sanitized native 86 to 152 s.
func TestProduct_DotARenameSource(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, dotARenameSource(test), "main.ts", "rules/no-var/rule.ts")
	test.Logf("%s: %.3fs", test.Name(), time.Since(started).Seconds())
}

func TestProduct_DotARenameOriginalLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, dotARenameLowered(test, false), "lint.c", "lint.mjs")
}

func TestProduct_DotARenameRenamedLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, dotARenameLowered(test, true), "lint.c", "lint.mjs")
}

func TestProduct_DotARenameOriginalNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, filepath.Join(dotARenameNative(test, false), "scanner"))
}

func TestProduct_DotARenameRenamedNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, filepath.Join(dotARenameNative(test, true), "scanner"))
}

// dotARenameFiles is every file under directory by its slash path, as Absolute reads it back, so a product and the
// copy Resolved made of it read alike.
func dotARenameFiles(test *testing.T, directory string) map[string]string {
	test.Helper()
	root, productRootError := buildcache.ProductRoot(directory)
	if productRootError != nil {
		test.Fatal(productRootError)
	}
	files := map[string]string{}
	walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil || entry.IsDir() {
			return pathError
		}
		data, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		relative, relativePathError := filepath.Rel(root, path)
		files[filepath.ToSlash(relative)] = buildcache.Absolute(string(data))
		return relativePathError
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return files
}

// Two readers of the renamed source, one after the other, each generate the registry into a copy of their own: the
// second copies the source with the registry it was built with, not the first reader's, and after both, the directory
// Resolved answered still holds the product's bytes exactly (#0sqk6ce). Generating in place fails it on either
// answer Resolved gives: a copy gains registry.go and differs from the product, and the product itself gains it too.
func TestDotARenameReadersRegisterPrivately(test *testing.T) {
	test.Parallel()
	source := dotARenameSource(test)
	built := dotARenameFiles(test, dotARenameProductDirectories[0])
	const registryTypeScript, registryGo = ".generated/registry.ts", ".generated/registry.go"
	if _, found := built[registryTypeScript]; !found {
		test.Fatalf("the renamed source was built without %s", registryTypeScript)
	}
	if _, found := built[registryGo]; found {
		test.Fatalf("the renamed source was built with %s", registryGo)
	}
	first := dotARenameFiles(test, dotARenameRegistered(test, source))
	// Generating for rule.ts changes the registry the source carries, so a reader generating in place would show.
	if first[registryTypeScript] == built[registryTypeScript] || first[registryGo] == "" {
		test.Fatalf("the first reader's registry is the source's as built, so a shared registry can't show here")
	}
	if seen := dotARenameFiles(test, source)[registryTypeScript]; seen != built[registryTypeScript] {
		test.Fatalf("after the first reader, the source's %s is not the one it was built with", registryTypeScript)
	}
	second := dotARenameFiles(test, dotARenameRegistered(test, source))
	if second[registryTypeScript] != first[registryTypeScript] || second[registryGo] != first[registryGo] {
		test.Fatal("the second reader generated a different registry from the same source")
	}
	resolved := dotARenameFiles(test, source)
	for name, content := range built {
		if resolved[name] != content {
			test.Errorf(
				"after both readers, %s in Resolved's %s differs from the product's",
				name,
				filepath.Base(source),
			)
		}
	}
	for name := range resolved {
		if _, found := built[name]; !found {
			test.Errorf(
				"after both readers, Resolved's %s holds %s, which the product doesn't",
				filepath.Base(source),
				name,
			)
		}
	}
	test.Logf(
		"two readers registered privately; Resolved's %s holds the product's %d files exactly",
		filepath.Base(source),
		len(built),
	)
}

func dotARenameSetup(test *testing.T, renamed bool) dotARenameProducts {
	test.Helper()
	started := time.Now()
	defer func() { test.Logf("TestDotARename (setup): %.3f s", time.Since(started).Seconds()) }()
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	p := dotARenameProducts{Original: directory}
	var readError error
	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		p.Oracle = lintGoOracleProduct(test)
	}()
	go func() {
		defer workers.Done()
		// Node generates the port's registry before it runs it, so it reads a private copy, never the source's.
		p.Changed = dotARenameRegistered(test, dotARenameSource(test))
		p.Before, readError = os.ReadFile(filepath.Join(directory, "rules/no-var/rule.a"))
		if readError != nil {
			return
		}
		p.After, readError = os.ReadFile(filepath.Join(p.Changed, "rules/no-var/rule.ts"))
		if readError != nil {
			return
		}
	}()
	go func() {
		defer workers.Done()
		if renamed {
			p.ChangedNative = filepath.Join(dotARenameNative(test, true), "scanner")
			p.ChangedJS = filepath.Join(dotARenameLowered(test, true), "lint.mjs")
		} else {
			p.OriginalNative = filepath.Join(dotARenameNative(test, false), "scanner")
			p.OriginalJS = filepath.Join(dotARenameLowered(test, false), "lint.mjs")
		}
	}()
	workers.Wait()
	if readError != nil {
		test.Fatal(readError)
	}
	if test.Failed() {
		test.Fatal("DotARename preparation failed")
	}
	return p
}

func dotARenameUnion(test *testing.T) {
	test.Helper()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "grain_dot_a_rename_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok &&
			strings.HasPrefix(fn.Name.Name, "TestDotARename_") &&
			fn.Name.Name != "TestDotARename_Setup" {
			names[fn.Name.Name] = true
		}
	}
	if len(names) != testDotARenameShards {
		test.Fatalf("enumerated %d shards, declared %d", len(names), testDotARenameShards)
	}
	for i := 0; i < testDotARenameShards; i++ {
		if !names[fmt.Sprintf("TestDotARename_%03d", i)] {
			test.Fatalf("missing shard %d", i)
		}
	}
	// The pinned original enumeration has exactly this one semantic case.
	cases := dotARenameCases
	owners := make(map[string]int)
	for shard := 0; shard < testDotARenameShards; shard++ {
		for _, key := range cases {
			if dotARenameOwner(key) == shard {
				owners[key]++
			}
		}
	}
	for _, key := range cases {
		if owners[key] != 1 {
			test.Fatalf("case %s has %d owners", key, owners[key])
		}
	}
	test.Logf("union: %d/%d cases exactly once", len(owners), len(cases))
}

var dotARenameSetupOnce [2]sync.Once
var dotARenamePrepared [2]*dotARenameProducts

// Filtered shards prepare their own real products before the case deadline.
func dotARenameReady(test *testing.T, renamed bool) dotARenameProducts {
	test.Helper()
	slot := 0
	if renamed {
		slot = 1
	}
	dotARenameSetupOnce[slot].Do(func() {
		products := dotARenameSetup(test, renamed)
		dotARenamePrepared[slot] = &products
	})
	if dotARenamePrepared[slot] == nil {
		test.Fatal("DotARename setup did not complete")
	}
	return *dotARenamePrepared[slot]
}

// dotARenameChild's output is read through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a five-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF five seconds
// after the child exited, which a probe counted as a caught shard with its output cut short (#n3gk22h's shape). Each
// caller's 10 s deadline ends a hung child.
func dotARenameChild(callerContext context.Context, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, os.Args[0], args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

type dotARenameProbe struct {
	Key      string
	Products dotARenameProducts
}

func TestDotARenameUnion(test *testing.T) {
	test.Parallel()
	dotARenameUnion(test)
}

// Both variants must equal the same Go oracle: this preserves the original
// rename-equivalence assertion without building both ports in one cold unit.
func dotARenameRunShard(test *testing.T, shard int) {
	test.Helper()
	dotARenameUnion(test)
	var probe dotARenameProbe
	supplied := os.Getenv("ADAMIC_DOT_A_RENAME_PROBE")
	if supplied != "" {
		data, readError := os.ReadFile(supplied)
		if readError != nil {
			test.Fatal(readError)
		}
		if decodeError := json.Unmarshal(data, &probe); decodeError != nil {
			test.Fatal(decodeError)
		}
	}
	for _, key := range dotARenameCases {
		if dotARenameOwner(key) != shard || (supplied != "" && probe.Key != key) {
			continue
		}
		renamed := strings.HasSuffix(key, ":renamed")
		p := probe.Products
		if supplied == "" {
			p = dotARenameReady(test, renamed)
		}
		started := time.Now()
		deadline := time.AfterFunc(90*time.Second, func() { panic("P0: DotARename shard exceeded 90s") })
		defer deadline.Stop()
		defer func() { test.Logf("%s (own work): %.3fs", key, time.Since(started).Seconds()) }()
		if !bytes.Equal(p.Before, p.After) {
			test.Fatal("rename changed module bytes")
		}
		path := manifest(test, []string{ownedWitnesses(test, p.Original, "no-var")[0] + "\tno-var"})
		directory, binary, module := p.Original, p.OriginalNative, p.OriginalJS
		if renamed {
			directory, binary, module = p.Changed, p.ChangedNative, p.ChangedJS
		}
		want := compareWithJavaScript(test, p.Oracle, binary, directory, path, module)
		test.Logf("%s: .a/.ts variant matches the shared Go oracle on all three runtimes (%d bytes)", key, len(want))
		if supplied == "" {
			dotARenamePlant(test, key, p, module)
		}
		if time.Since(started) >= 60*time.Second {
			test.Fatal("DotARename own work over 60s; split smaller")
		}
	}
}

func dotARenamePlant(test *testing.T, key string, p dotARenameProducts, module string) {
	test.Helper()
	data, readError := os.ReadFile(module)
	if readError != nil {
		test.Fatal(readError)
	}
	planted := filepath.Join(test.TempDir(), "planted.mjs")
	marker := "planted dot-a rename disagreement"
	if writeError := os.WriteFile(
		planted,
		append(data, []byte(fmt.Sprintf("\nconsole.log(%q);\n", marker))...),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if strings.HasSuffix(key, ":renamed") {
		p.ChangedJS = planted
	} else {
		p.OriginalJS = planted
	}
	data, encodeError := json.Marshal(dotARenameProbe{Key: key, Products: p})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	products := filepath.Join(test.TempDir(), "products.json")
	if writeError := os.WriteFile(products, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	caught := 0
	for shard := 0; shard < testDotARenameShards; shard++ {
		name := fmt.Sprintf("TestDotARename_%03d", shard)
		timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		command := dotARenameChild(timeoutContext, "-test.run=^"+name+"$", "-test.v", "-test.timeout=10s")
		command.Env = append(os.Environ(), "ADAMIC_DOT_A_RENAME_PROBE="+products)
		stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
		output := append(stdout, stderr...)
		expired := timeoutContext.Err()
		cancel()
		if expired != nil {
			test.Fatalf("planted probe exceeded its deadline: %v\n%s", expired, output)
		}
		if captureError != nil {
			if bytes.Count(output, []byte("--- FAIL: "+name)) != 1 ||
				!bytes.Contains(output, []byte("emitted JavaScript:")) ||
				!bytes.Contains(output, []byte(marker)) {
				test.Fatalf("wrong planted failure: %v\n%s", captureError, output)
			}
			caught++
		}
	}
	if caught != 1 {
		test.Fatalf("planted disagreement caught %d times", caught)
	}
	test.Logf("planted disagreement caught exactly by TestDotARename_%03d", dotARenameOwner(key))
}

func TestDotARename_000(test *testing.T) {
	test.Parallel()
	dotARenameRunShard(test, 0)
}

func TestDotARename_001(test *testing.T) {
	test.Parallel()
	dotARenameRunShard(test, 1)
}
