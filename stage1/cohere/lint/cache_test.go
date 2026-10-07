package lint

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Keys name and length-prefix every component. Tests can remove one component
// in a subprocess, without adding a production cache-bypass exception.
var lintKeyDrop string

func lintKey(kind string, components map[string]string) string {
	hash := sha256.New()
	part := func(s string) { fmt.Fprintf(hash, "%d:", len(s)); hash.Write([]byte(s)) }
	part("adamic-lint-v1")
	part(kind)
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if kind+"/"+name != lintKeyDrop {
			part(name)
			part(components[name])
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func lintBytes(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// This repository can have no go.sum when every dependency is replaced locally.
func lintOptionalBytes(t *testing.T, path string) string {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "absent"
	}
	if err != nil {
		t.Fatal(err)
	}
	return "present:" + string(data)
}
func lintCacheRoot(t *testing.T, kind string) string {
	t.Helper()
	root, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "adamic", "lint", kind)
}

var lintLocks sync.Map

func lintLock(key string) func() {
	value, _ := lintLocks.LoadOrStore(key, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	return lock.Unlock
}

// Publish a complete file, with a digest checked on every read. A torn or
// corrupt observation is a miss. Cross-process writers publish identical bytes.
func lintPublish(t *testing.T, path string, data []byte, mode fs.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".publish-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Chmod(mode); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(f.Name(), path); err != nil {
		t.Fatal(err)
	}
}

type lintEnvelope struct {
	Key, Digest string
	Data        []byte
}

func lintCachedBytes(t *testing.T, kind, key string, run func() []byte) []byte {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return run()
	}
	path := filepath.Join(lintCacheRoot(t, kind), key+".json")
	unlock := lintLock(path)
	defer unlock()
	if data, err := os.ReadFile(path); err == nil {
		var entry lintEnvelope
		if json.Unmarshal(data, &entry) == nil && entry.Key == key && entry.Digest == lintKey("digest", map[string]string{"bytes": string(entry.Data)}) {
			return entry.Data
		}
	}
	data := run()
	if !t.Failed() {
		encoded, err := json.Marshal(lintEnvelope{key, lintKey("digest", map[string]string{"bytes": string(data)}), data})
		if err != nil {
			t.Fatal(err)
		}
		lintPublish(t, path, encoded, 0600)
	}
	return data
}
func lintArtifact(t *testing.T, kind, key string, run func() string) string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return run()
	}
	// Artifact and its integrity envelope are one atomic directory publication.
	root := lintCacheRoot(t, kind)
	directory := filepath.Join(root, key)
	path := filepath.Join(directory, "artifact")
	unlock := lintLock(directory)
	defer unlock()
	valid := func() bool {
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		digest, err := os.ReadFile(filepath.Join(directory, "digest"))
		return err == nil && string(digest) == lintKey("digest", map[string]string{"bytes": string(data)})
	}
	if valid() {
		return path
	}
	produced := run()
	data := []byte(lintBytes(t, produced))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	temporary, err := os.MkdirTemp(root, ".artifact-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(temporary)
	if err := os.WriteFile(filepath.Join(temporary, "artifact"), data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(temporary, "digest"), []byte(lintKey("digest", map[string]string{"bytes": string(data)})), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temporary, directory); err != nil {
		if !valid() { // A damaged entry must never be trusted. Use the fresh artifact.
			return produced
		}
	}
	return path
}
func lintTool(t *testing.T, name string, args ...string) string {
	return string(executeUncached(t, "", name, args...).output)
}
func lintContext(t *testing.T) string {
	// Runtime flags, runner code, harness code and environment can all affect an
	// observation. The bypass control itself does not change program semantics.
	parts := map[string]string{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "go": lintTool(t, "go", "version"), "uid": fmt.Sprint(os.Geteuid()), "executable": lintExecutable(t)}
	var stack syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_STACK, &stack); err != nil {
		t.Fatal(err)
	}
	parts["stack"] = fmt.Sprint(stack.Cur, stack.Max)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	registryFiles, err := filepath.Glob("registry/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, registryFiles...)
	for _, file := range append(files, filepath.Join(repository, "oracle/node.mjs"), filepath.Join(repository, "oracle/adamic.mjs")) {
		parts[file] = lintBytes(t, file)
	}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if name != "ADAMIC_GATE_UNCACHED" && name != "XDG_CACHE_HOME" && name != "COHERE_DOCS_CAPTURE" && name != "ADAMIC_LINT_CACHE_PROBE" && name != "ADAMIC_LINT_CACHE_DROP" {
			parts["env/"+name] = entry
		}
	}
	return lintKey("context", parts)
}
func lintCohere(t *testing.T) string {
	root := filepath.Join(repository, "cohere")
	parts := map[string]string{"HEAD": string(executeUncached(t, root, "git", "rev-parse", "HEAD").output), "diff": string(executeUncached(t, root, "git", "diff", "--binary", "HEAD", "--").output), "go": lintTool(t, "go", "version"), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "context": lintContext(t)}
	// Include untracked files and nested submodules as well as tracked edits.
	files := strings.Split(string(executeUncached(t, root, "git", "ls-files", "--others", "--exclude-standard", "-z").output), "\x00")
	for _, file := range files {
		if file != "" {
			parts["untracked/"+file] = lintBytes(t, filepath.Join(root, file))
		}
	}
	parts["submodules"] = string(executeUncached(t, root, "git", "submodule", "status", "--recursive").output)
	nested := executeUncached(t, root, "git", "submodule", "foreach", "--quiet", "--recursive", "pwd").output
	for _, module := range strings.Split(strings.TrimSpace(string(nested)), "\n") {
		if module == "" {
			continue
		}
		parts[module+"/HEAD"] = string(executeUncached(t, module, "git", "rev-parse", "HEAD").output)
		parts[module+"/diff"] = string(executeUncached(t, module, "git", "diff", "--binary", "HEAD", "--").output)
		unknown := executeUncached(t, module, "git", "ls-files", "--others", "--exclude-standard", "-z").output
		for _, file := range strings.Split(string(unknown), "\x00") {
			if file != "" {
				parts[module+"/untracked/"+file] = lintBytes(t, filepath.Join(module, file))
			}
		}
	}
	return lintKey("cohere", parts)
}
func lintOracleKey(t *testing.T, root string) string {
	prepareRegistry(t, root)
	parts := map[string]string{"cohere": lintCohere(t), "overlay": lintBytes(t, "testdata/oracle.go"), "registry": lintBytes(t, filepath.Join(root, ".generated/registry.go"))}
	for _, d := range prepareRegistry(t, root) {
		parts["adapter/"+d.Slug] = lintBytes(t, filepath.Join(root, "rules", d.Slug, "oracle.go"))
	}
	// Every available adapter participates, including ones not registered in this
	// private oracle. Full and restricted oracle builds share the same contract.
	for _, d := range prepareRegistry(t, ".") {
		parts["available-adapter/"+d.Slug] = lintBytes(t, filepath.Join("rules", d.Slug, "oracle.go"))
	}
	return lintKey("oracle", parts)
}

// Selected and full builds validate and render through the same registry API.
func lintRegistry(t *testing.T, root string) ([]registry.Descriptor, error) {
	marker, err := os.ReadFile(filepath.Join(root, ".selected-rule"))
	if os.IsNotExist(err) {
		return registry.Generate(root)
	}
	if err != nil {
		return nil, err
	}
	return registry.Generate(root, string(marker))
}

var lintOracles sync.Map

func goOracleFrom(t *testing.T, sourceRoot string) string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return goOracleFromUncached(t, sourceRoot)
	}
	key := lintOracleKey(t, sourceRoot)
	path := lintArtifact(t, "oracle", key, func() string {
		path := goOracleFromUncached(t, sourceRoot)
		if lintOracleKey(t, sourceRoot) != key {
			t.Fatal("oracle inputs changed during build; retry")
		}
		return path
	})
	lintOracles.Store(path, key)
	return path
}
func lintManifestKey(t *testing.T, path string) string {
	data := lintBytes(t, path)
	parts := map[string]string{"manifest": data}
	for _, row := range strings.Split(data, "\n") {
		if row != "" {
			file, _, _ := strings.Cut(row, "\t")
			parts["source/"+file] = lintBytes(t, file)
		}
	}
	return lintKey("manifest", parts)
}

// Timing samples must execute processes, never measure an observation replay.
func lintObservationUncached() bool {
	return os.Getenv("ADAMIC_GATE_UNCACHED") == "1" || os.Getenv("ADAMIC_LINT_BENCH") == "1"
}

func execute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	if lintObservationUncached() {
		return executeUncached(t, directory, name, args...)
	}
	if key, ok := lintOracles.Load(name); ok && len(args) >= 2 && args[0] == "--manifest" {
		started := time.Now()
		data := lintCachedBytes(t, "go-output", lintKey("go-output", map[string]string{"oracle": key.(string), "manifest": lintManifestKey(t, args[1]), "arguments": strings.Join(args[2:], "\x00")}), func() []byte { return executeUncached(t, directory, name, args...).output })
		return execution{data, time.Since(started)}
	}
	return executeUncached(t, directory, name, args...)
}

// Parse actual import declarations, never text in comments or string literals.
// Relative names inside the snapshot are stable across temporary directories.
var lintImports sync.Map

func lintModules(t *testing.T, entry string) string {
	root, err := filepath.Abs(filepath.Dir(entry))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string]string{}
	seen := map[string]bool{}
	var visit func(string)
	visit = func(path string) {
		path, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if seen[path] {
			return
		}
		seen[path] = true
		data := lintBytes(t, path)
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			relative = path
		}
		parts[relative] = data
		if filepath.Ext(path) == ".mjs" {
			return
		}

		inputKey := lintKey("imports", map[string]string{"bytes": data})
		value, known := lintImports.Load(inputKey)
		var imports []string
		if known {
			imports = value.([]string)
		} else {
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			for _, statement := range program.Files()[0].Statements.Nodes {
				if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
					continue
				}
				if specifier := statement.ModuleSpecifier(); specifier != nil {
					imports = append(imports, specifier.Text())
				}
			}
			lintImports.Store(inputKey, imports)
		}
		for _, name := range imports {
			if filepath.IsAbs(name) {
				visit(name)
			} else if strings.HasPrefix(name, ".") {
				visit(filepath.Join(filepath.Dir(path), name))
			} else if name != "adamic" {
				t.Fatalf("cache cannot identify module %q", name)
			}
		}
	}
	visit(entry)
	return lintKey("modules", parts)
}

var lintExecutableIdentity sync.Once
var lintExecutableKey string

func lintExecutable(t *testing.T) string {
	lintExecutableIdentity.Do(func() { lintExecutableKey = lintTool(t, "go", "tool", "buildid", os.Args[0]) })
	return lintExecutableKey
}

func lintCompiler(t *testing.T) string {
	// The executable identity prevents a long-lived worker built from old compiler
	// sources from publishing under a newly edited compiler's source-only key.
	identity := lintExecutable(t)
	parts := map[string]string{"go.sum": lintOptionalBytes(t, filepath.Join(repository, "go.sum")), "go.mod": lintBytes(t, filepath.Join(repository, "go.mod")), "executable": identity, "cohere": lintCohere(t)}
	err := filepath.WalkDir(filepath.Join(repository, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			parts[path] = "symlink:" + target
		} else if !entry.IsDir() {
			parts[path] = "file:" + lintBytes(t, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return lintKey("compiler", parts)
}
func lintPortOptions(directory string, sanitize bool) native.Options {
	options := native.Options{Sanitize: sanitize}
	if _, err := os.Stat(filepath.Join(directory, ".selected-rule")); err == nil && os.Getenv("ADAMIC_GATE_UNCACHED") != "1" {
		options.Split = true
		options.Jobs = runtime.GOMAXPROCS(0)
	}
	return options
}

func lintPortKey(t *testing.T, directory string, sanitize bool) string {
	options := lintPortOptions(directory, sanitize)
	prepareRegistry(t, directory)
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		t.Fatal(err)
	}
	return lintKey("port", map[string]string{"modules": lintModules(t, filepath.Join(directory, "main.ts")), "compiler": lintCompiler(t), "runtime": filepath.Base(filepath.Dir(library)), "context": lintContext(t), "split": fmt.Sprint(options.Split), "jobs": fmt.Sprint(options.Jobs), "flags": strings.Join(native.Flags(options), "\x00")})
}
func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return buildPortUncached(t, directory, sanitize)
	}
	key := lintPortKey(t, directory, sanitize)
	return lintArtifact(t, "port", key, func() string {
		path := buildPortOptionsUncached(t, directory, lintPortOptions(directory, sanitize))
		if lintPortKey(t, directory, sanitize) != key {
			t.Fatal("port inputs changed during build; retry")
		}
		return path
	})
}
func node(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	if lintObservationUncached() {
		return nodeUncached(t, directory, manifest, count)
	}
	prepareRegistry(t, directory)
	started := time.Now()
	key := lintKey("node", map[string]string{"modules": lintModules(t, filepath.Join(directory, "main.ts")), "node": lintTool(t, "node", "--version"), "manifest": lintManifestKey(t, manifest), "count": fmt.Sprint(count), "context": lintContext(t)})
	data := lintCachedBytes(t, "node", key, func() []byte { return nodeUncached(t, directory, manifest, count).output })
	return execution{data, time.Since(started)}
}
func emittedJavaScript(t *testing.T, directory string) string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return emittedJavaScriptUncached(t, directory)
	}
	prepareRegistry(t, directory)
	key := lintKey("javascript-build", map[string]string{"modules": lintModules(t, filepath.Join(directory, "main.ts")), "compiler": lintCompiler(t), "context": lintContext(t)})
	return lintArtifact(t, "javascript-build", key, func() string {
		path := emittedJavaScriptUncached(t, directory)
		changed := lintKey("javascript-build", map[string]string{"modules": lintModules(t, filepath.Join(directory, "main.ts")), "compiler": lintCompiler(t), "context": lintContext(t)})
		if changed != key {
			t.Fatal("JavaScript inputs changed during build; retry")
		}
		return path
	})
}
func runJavaScript(t *testing.T, module, manifest string, count bool) execution {
	t.Helper()
	if lintObservationUncached() {
		return runJavaScriptUncached(t, module, manifest, count)
	}
	started := time.Now()
	key := lintKey("javascript", map[string]string{"modules": lintBytes(t, module), "node": lintTool(t, "node", "--version"), "manifest": lintManifestKey(t, manifest), "count": fmt.Sprint(count), "context": lintContext(t)})
	data := lintCachedBytes(t, "javascript", key, func() []byte { return runJavaScriptUncached(t, module, manifest, count).output })
	return execution{data, time.Since(started)}
}

// Captures are serialized as source bytes and manifest fields, never temporary
// absolute paths. Restore the original script filename before the comparison.
type lintCaptured struct{ Name, Source, Fields string }

func lintCapture(t *testing.T, sourceRoot string, d registry.Descriptor) []string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return upstreamSelection(t, sourceRoot, &d)
	}
	key := lintKey("capture", map[string]string{"cohere": lintCohere(t), "overlay": lintBytes(t, filepath.Join(repository, "cohere/internal/lint/testing/rule_testing.go")), "package": d.UpstreamPackage, "filter": "^" + d.UpstreamTest, "rule": d.Name})
	data := lintCachedBytes(t, "capture", key, func() []byte {
		var records []lintCaptured
		for _, row := range upstreamSelection(t, sourceRoot, &d) {
			path, fields, _ := strings.Cut(row, "\t")
			records = append(records, lintCaptured{filepath.Base(path), lintBytes(t, path), fields})
		}
		encoded, err := json.Marshal(records)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	})
	var records []lintCaptured
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	var rows []string
	root := t.TempDir()
	for i, record := range records {
		path := filepath.Join(root, fmt.Sprint(i), record.Name)
		lintPublish(t, path, []byte(record.Source), 0600)
		rows = append(rows, path+"\t"+record.Fields)
	}
	return rows
}

func lintFullCapture(t *testing.T, root string) []string {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		return upstreamSelection(t, root, nil)
	}
	descriptors := prepareRegistry(t, root)
	encoded, err := json.Marshal(descriptors)
	if err != nil {
		t.Fatal(err)
	}
	key := lintKey("capture", map[string]string{"cohere": lintCohere(t), "overlay": lintBytes(t, filepath.Join(repository, "cohere/internal/lint/testing/rule_testing.go")), "package": "all registered packages", "filter": string(encoded), "rule": "all"})
	data := lintCachedBytes(t, "capture", key, func() []byte { return lintCaptureObservation(t, upstreamSelection(t, root, nil)) })
	var records []lintCaptured
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	var rows []string
	directory := t.TempDir()
	for i, record := range records {
		path := filepath.Join(directory, fmt.Sprint(i), record.Name)
		lintPublish(t, path, []byte(record.Source), 0600)
		rows = append(rows, path+"\t"+record.Fields)
	}
	if len(rows) < 150 {
		t.Fatalf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows
}

func lintCaptureObservation(t *testing.T, rows []string) []byte {
	var records []lintCaptured
	for _, row := range rows {
		path, fields, _ := strings.Cut(row, "\t")
		records = append(records, lintCaptured{filepath.Base(path), lintBytes(t, path), fields})
	}
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
