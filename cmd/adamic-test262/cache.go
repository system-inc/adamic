package main

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Cache observations, never verdicts: decide still compares both executions on every hit.
type resultCache struct {
	directory   string
	bypass      bool // Set per runner before observations; workers share this policy.
	locks       sync.Map
	nodeContext string
}

type recordedExecution struct {
	Stdout, Stderr []byte
	Exit           int
	Signal         string
	TimedOut       bool
}

func recordExecution(value execution) recordedExecution {
	return recordedExecution{[]byte(value.Stdout), []byte(value.Stderr), value.Exit, value.Signal, value.TimedOut}
}
func (value recordedExecution) execution() execution {
	return execution{string(value.Stdout), string(value.Stderr), value.Exit, value.Signal, value.TimedOut}
}

type cacheEnvelope struct {
	Key, Digest string
	Value       json.RawMessage
}

func cacheKey(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(hash, "%d:", len(part))
		hash.Write([]byte(part))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func compilerResultKey(program, compiler, command, context string) string {
	return cacheKey("test262-compiler-v1", program, compiler, command, context)
}

func nodeResultKey(program, version, adaptation, command, context string) string {
	return cacheKey("test262-node-v1", program, version, adaptation, command, context)
}
func nativeResultKey(code, library, command, context string) string {
	return cacheKey("test262-native-v1", code, library, command, context)
}

func (cache *resultCache) observe(key string, execute func() (execution, bool)) execution {
	if cache == nil || cache.bypass {
		value, _ := execute()
		return value
	}
	return cache.reuse(key, execute)
}

// Tests exercise disk reuse independently of the integration gate's environment bypass.
func (cache *resultCache) reuse(key string, execute func() (execution, bool)) execution {
	value, _ := cache.locks.LoadOrStore(key, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	path := filepath.Join(cache.directory, key+".json")
	if contents, err := os.ReadFile(path); err == nil {
		var envelope cacheEnvelope
		var recorded recordedExecution
		if json.Unmarshal(contents, &envelope) == nil && envelope.Key == key && envelope.Digest == cacheKey(string(envelope.Value)) && json.Unmarshal(envelope.Value, &recorded) == nil {
			return recorded.execution()
		}
	}
	result, reusable := execute()
	// Deadlines and failed starts are host observations, not reproducible program answers.
	if !reusable || result.TimedOut || (result.Exit == -1 && result.Signal == "") {
		return result
	}
	encoded, err := json.Marshal(recordExecution(result))
	if err != nil {
		return result
	}
	contents, err := json.Marshal(cacheEnvelope{key, cacheKey(string(encoded)), encoded})
	if err != nil || os.MkdirAll(cache.directory, 0700) != nil {
		return result
	}
	temporary, err := os.CreateTemp(cache.directory, ".result-")
	if err != nil {
		return result
	}
	defer os.Remove(temporary.Name())
	_, writeError := temporary.Write(contents)
	closeError := temporary.Close()
	if writeError == nil && closeError == nil {
		_ = os.Rename(temporary.Name(), path)
	}
	return result
}

// resultEnvironment names every inherited variable that can change an observation, and the runner's
// children (the compiler worker or adamic c, clang, the native program and Node) inherit these and
// passedEnvironment's, nothing else (childEnvironment). So what a child can read is what the key names:
// a variable not listed here never reaches one, and adding a variable a child should see means keying
// it. ASAN_OPTIONS and UBSAN_OPTIONS are absent on purpose: the engine sets both for the native run,
// and nativeEnvironment is already part of the native command's key.
var resultEnvironment = []string{
	// The compiler is a Go process, and GODEBUG switches runtime and standard library behavior.
	"GODEBUG",
	// clang finds the linker through PATH, and the sanitizers find the symbolizer whose frames are
	// part of a crash's stderr through it. Builder's resolver will name these tools instead.
	"PATH",
	// clang reads these for the headers, libraries, helper programs and options of a link.
	"CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "COMPILER_PATH", "CCC_OVERRIDE_OPTIONS",
	// On macOS /usr/bin/clang is xcrun's shim: these choose the toolchain and SDK behind it.
	"DEVELOPER_DIR", "TOOLCHAINS", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET",
	// The dynamic loader can put other code in clang, Node or the native program.
	"LD_PRELOAD", "LD_LIBRARY_PATH", "DYLD_INSERT_LIBRARIES", "DYLD_LIBRARY_PATH", "DYLD_FALLBACK_LIBRARY_PATH", "DYLD_FRAMEWORK_PATH",
	// A sanitizer report is symbolized by the program these name.
	"ASAN_SYMBOLIZER_PATH", "UBSAN_SYMBOLIZER_PATH",
	// Node's options, debug channels, warnings and colors change what a Node run prints or does.
	"NODE_OPTIONS", "NODE_DEBUG", "NODE_DEBUG_NATIVE", "NODE_NO_WARNINGS", "NODE_PENDING_DEPRECATION", "NODE_REDIRECT_WARNINGS",
	"FORCE_COLOR", "NO_COLOR", "NODE_DISABLE_COLORS",
	// Node's Date and Intl answer in the time zone and locale these choose, from the ICU data named.
	"TZ", "LC_ALL", "LC_MESSAGES", "LANG", "NODE_ICU_DATA",
}

// passedEnvironment is what a child inherits beyond resultEnvironment, unkeyed: where clang and the
// compiler write scratch files they remove, and where a coverage build of the runner, run as its own
// compiler worker, writes its counters. Neither reaches what a child prints or how it exits.
var passedEnvironment = []string{"TMPDIR", "GOCOVERDIR"}

// keyedEnvironment is each resultEnvironment variable lookup finds set, with its name: unset and set to
// empty are different to some readers, so only a set variable is keyed.
func keyedEnvironment(lookup func(string) (string, bool)) []string {
	var keyed []string
	for _, name := range resultEnvironment {
		if value, set := lookup(name); set {
			keyed = append(keyed, name+"="+value)
		}
	}
	return keyed
}

// childEnvironment is the environment every child runs with: the variables the key names, as the
// runner has them, then passedEnvironment's, then extra, whose later values win.
func childEnvironment(extra ...string) []string {
	environment := keyedEnvironment(os.LookupEnv)
	for _, name := range passedEnvironment {
		if value, set := os.LookupEnv(name); set {
			environment = append(environment, name+"="+value)
		}
	}
	return append(environment, extra...)
}

// runnerTools are the programs the runner runs by name, each resolved once through toolchain.Path (a
// declared tools directory's before PATH's, and only it in a sandbox), so the context hashes the very
// program every observation runs.
type runnerTools struct{ node, clang string }

func resolveTools() (runnerTools, error) {
	node, err := toolchain.Path("node")
	if err != nil {
		return runnerTools{}, err
	}
	clang, err := toolchain.Path("clang")
	if err != nil {
		return runnerTools{}, err
	}
	return runnerTools{node: node, clang: clang}, nil
}

// prepareCache derives the execution context every observation is keyed under. root is the checkout
// under test, as prepareMode gets it, and tools are the programs its runs execute.
func prepareCache(root string, tools runnerTools) (*resultCache, string, string, error) {
	directory, err := os.UserCacheDir()
	if err != nil {
		return nil, "", "", err
	}
	versionCommand := exec.Command(tools.node, "--version")
	versionCommand.Env = childEnvironment()
	version, err := versionCommand.CombinedOutput()
	if err != nil {
		return nil, "", "", fmt.Errorf("node --version: %w: %s", err, version)
	}
	parts := []string{"test262-context-v2", runtime.GOOS, runtime.GOARCH, fmt.Sprint(os.Geteuid())}
	// Relative paths in the keyed commands (an explicit -root or -work, and the scratch compiler under
	// it) resolve from the working directory, so where the runner stands in the checkout is part of
	// what they name. Where the checkout sits on the machine is not part of the context. The commands
	// still name this machine's places (node's and clang's paths found on PATH, the runtime library
	// and scratch compiler under the cache and work directories), so today observations are reused on
	// the machine that made them, from any copy of the tree there.
	place, err := checkoutPlace(root)
	if err != nil {
		return nil, "", "", err
	}
	parts = append(parts, place)
	var stack syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_STACK, &stack); err != nil {
		return nil, "", "", err
	}
	parts = append(parts, fmt.Sprint(stack.Cur, stack.Max))
	parts = append(parts, keyedEnvironment(os.LookupEnv)...)
	// Node's harness is embedded source plus Go build settings, not the linked compiler.
	// Lowering and the embedded runtime can change without changing Node's execution.
	nodeParts := append([]string{}, parts...)
	nodeParts = append(nodeParts, nodeHarnessIdentity())
	if info, ok := debug.ReadBuildInfo(); ok {
		nodeParts = append(nodeParts, info.GoVersion)
		for _, setting := range info.Settings {
			if !strings.HasPrefix(setting.Key, "vcs") {
				nodeParts = append(nodeParts, setting.Key, setting.Value)
			}
		}
	}
	// The full executable remains a conservative compiler and native identity.
	executable, err := os.Executable()
	if err != nil {
		return nil, "", "", err
	}
	for _, path := range []string{executable, tools.node, tools.clang} {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, "", "", err
		}
		// The runner's location is not input to successful compilation or native execution.
		// go run allocates a new executable path even for identical built bytes.
		if path == executable {
			parts = append(parts, "runner", cacheKey(string(contents)))
		} else {
			parts = append(parts, path, cacheKey(string(contents)))
		}
		if path == tools.node {
			nodeParts = append(nodeParts, path, cacheKey(string(contents)))
		}
	}
	return &resultCache{directory: filepath.Join(directory, "adamic", "test262"), bypass: os.Getenv("ADAMIC_GATE_UNCACHED") == "1", nodeContext: cacheKey(nodeParts...)}, string(version), cacheKey(parts...), nil
}

// checkoutPlace is the working directory relative to the checkout root, in slash form: the same in
// every copy of the tree, and different wherever a relative path would name something else. Both
// sides lose their symbolic links first, since the working directory may be reported either way
// (macOS's /var is /private/var).
func checkoutPlace(root string) (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	workingDirectory, err = filepath.EvalSymlinks(workingDirectory)
	if err != nil {
		return "", err
	}
	checkout, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	checkout, err = filepath.EvalSymlinks(checkout)
	if err != nil {
		return "", err
	}
	place, err := filepath.Rel(checkout, workingDirectory)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(place), nil
}

// Embedding the actual built sources avoids trusting mutable checkout files at run time.
// Compiler implementation and tests cannot affect the Node adaptation/capture path.
//
//go:embed *.go
var runnerSources embed.FS

func nodeHarnessIdentity() string { return nodeHarnessSourceIdentity(runnerSources) }

func nodeHarnessSourceIdentity(sources fs.FS) string {
	entries, err := fs.ReadDir(sources, ".")
	if err != nil {
		panic(err)
	}
	parts := []string{"test262-node-harness-v2"}
	for _, entry := range entries {
		name := entry.Name()
		if name == "compiler.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		contents, err := fs.ReadFile(sources, name)
		if err != nil {
			panic(err)
		}
		parts = append(parts, name, string(contents))
	}
	return cacheKey(parts...)
}

// Binary identity covers executed compiler behavior. A source fingerprint additionally
// invalidates observations for library-worker edits that Go can elide from linked code,
// such as same-length comment changes. It only adds misses; it does not replace the binary.
func loweringSourceIdentity(root string) (string, error) {
	directory := filepath.Join(root, "internal", "lower")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	parts := []string{"test262-lowering-source-v1"}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return "", err
		}
		parts = append(parts, name, string(contents))
	}
	return cacheKey(parts...), nil
}
