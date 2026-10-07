package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
)

// Cache observations, never verdicts: decide still compares both executions on every hit.
type resultCache struct {
	directory string
	locks     sync.Map
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
	version, err := exec.Command("node", "--version").CombinedOutput()
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
	// The executable covers the runner's own harness, adaptation and capture implementation.
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
		parts = append(parts, path, cacheKey(string(contents)))
	}
	return &resultCache{directory: filepath.Join(directory, "adamic", "test262")}, string(version), cacheKey(parts...), nil
}
