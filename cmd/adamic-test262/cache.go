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
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/system-inc/adamic/internal/boundedrun"
)

// Cache observations, never verdicts: decide still compares both executions on every hit.
type resultCache struct {
	directory   string
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
	if cache == nil || os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
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

func prepareCache() (*resultCache, string, string, error) {
	directory, err := os.UserCacheDir()
	if err != nil {
		return nil, "", "", err
	}
	command, release := boundedrun.Command(boundedrun.Probe, "node", "--version")
	defer release()
	version, err := command.CombinedOutput()
	if err != nil {
		return nil, "", "", fmt.Errorf("node --version: %w: %s", err, version)
	}
	parts := []string{"test262-context-v1", runtime.GOOS, runtime.GOARCH, fmt.Sprint(os.Geteuid())}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", "", err
	}
	parts = append(parts, cwd)
	var stack syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_STACK, &stack); err != nil {
		return nil, "", "", err
	}
	parts = append(parts, fmt.Sprint(stack.Cur, stack.Max))
	environment := os.Environ()
	sort.Strings(environment)
	for _, variable := range environment {
		if !strings.HasPrefix(variable, "ADAMIC_GATE_UNCACHED=") {
			parts = append(parts, variable)
		}
	}
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
	for _, name := range []string{executable, "node", "clang"} {
		path := name
		if name != executable {
			path, err = exec.LookPath(name)
			if err != nil {
				return nil, "", "", err
			}
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, "", "", err
		}
		// The runner's location is not input to successful compilation or native execution.
		// go run allocates a new executable path even for identical built bytes.
		if name == executable {
			parts = append(parts, "runner", cacheKey(string(contents)))
		} else {
			parts = append(parts, path, cacheKey(string(contents)))
		}
		if name == "node" {
			nodeParts = append(nodeParts, path, cacheKey(string(contents)))
		}
	}
	return &resultCache{directory: filepath.Join(directory, "adamic", "test262"), nodeContext: cacheKey(nodeParts...)}, string(version), cacheKey(parts...), nil
}

// Embedding the actual built sources avoids trusting mutable checkout files at run time.
// Compiler implementation and tests cannot affect the Node adaptation/capture path.
//
//go:embed *.go
var runnerSources embed.FS

func nodeHarnessIdentity() string {
	return nodeHarnessWithHelperIdentity(runnerSources, boundedrun.Identity())
}

func nodeHarnessWithHelperIdentity(sources fs.FS, helper string) string {
	return cacheKey(nodeHarnessSourceIdentity(sources), helper)
}

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
