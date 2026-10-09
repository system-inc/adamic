package buildcache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const auditReplayKey = "_ADAMIC_BUILD_AUDIT_KEY"
const auditReplayDirectory = "_ADAMIC_BUILD_AUDIT_DIRECTORY"
const auditRebuiltExit = 79

// Callbacks cannot be serialized. A test audit replays its owning test in a separate process, using the original
// callback and stopping immediately after that callback builds the requested key. Compilation, setup and rebuild
// all belong to the audit unit's clock. No executable is copied or retained during the test's fetch.
// Keep the checkout and build environment unchanged until the gate drains this queue. Inputs are checked again
// before replay, and replay must reach the same key; a changed checkout or environment fails rather than auditing
// a different product. Tools using Get replay their executable and arguments instead of a test package.
type auditEntry struct {
	Key        string
	Inputs     Inputs
	Files      []manifestFile
	Product    string
	Root       string
	Directory  string
	Test       bool
	Executable string
	Arguments  []string
}

func auditDirectory() (string, error) {
	directory := os.Getenv("ADAMIC_BUILD_STORE_AUDITS")
	if directory == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		directory = filepath.Join(user, "adamic", "build-audits")
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	return directory, os.MkdirAll(directory, 0o700)
}

func auditEntries() ([]string, error) {
	directory, err := auditDirectory()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	return paths, nil
}

func queueAudit(key string, inputs Inputs, fetched string, testName []string) error {
	directory, err := auditDirectory()
	if err != nil {
		return err
	}
	fetched, err = filepath.Abs(fetched)
	if err != nil {
		return err
	}
	product, err := describeProduct(key, inputs.Name, fetched)
	if err != nil {
		return err
	}
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	working, err := os.Getwd()
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	entry := auditEntry{Key: key, Inputs: inputs, Files: product.Files, Product: fetched, Root: root, Directory: working,
		Executable: executable, Arguments: append([]string(nil), os.Args[1:]...), Test: flag.Lookup("test.run") != nil}
	if entry.Test {
		name := ""
		if len(testName) != 0 {
			name = testName[0]
		}
		if name == "" {
			name = callingTest()
		}
		entry.Arguments = replayTestArguments(entry.Arguments, name)
	}
	// Keep distinct snapshots of the same key: a subsequent honest fetch must not overwrite a poisoning witness.
	files, _ := json.Marshal(entry.Files)
	path := filepath.Join(directory, fmt.Sprintf("%s-%x.json", key, sha256.Sum256(files)))
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return writeQueuedEntry(path, encoded)
}

// Product supplies the full subtest name. Get used by a test supplies its top-level caller; TestMain and tools
// have no owning test, so they retain the invocation's original selection.
func callingTest() string {
	callers := make([]uintptr, 32)
	count := runtime.Callers(2, callers)
	frames := runtime.CallersFrames(callers[:count])
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "_test.go") {
			function := frame.Function
			if dot := strings.LastIndex(function, "/"); dot >= 0 {
				function = function[dot+1:]
			}
			parts := strings.Split(function, ".")
			if len(parts) > 1 && strings.HasPrefix(parts[1], "Test") && parts[1] != "TestMain" {
				return parts[1]
			}
		}
		if !more {
			return ""
		}
	}
}

func replayTestArguments(arguments []string, name string) []string {
	var kept []string
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		flagName := strings.SplitN(argument, "=", 2)[0]
		remove := flagName == "-test.timeout" || flagName == "-test.count" || flagName == "-test.paniconexit0" ||
			flagName == "-test.cpuprofile" || flagName == "-test.memprofile" || flagName == "-test.blockprofile" ||
			flagName == "-test.mutexprofile" || flagName == "-test.coverprofile" || flagName == "-test.trace" ||
			flagName == "-test.shuffle" || flagName == "-test.testlogfile" || flagName == "-test.outputdir" || flagName == "-test.gocoverdir" || (name != "" && flagName == "-test.run")
		if remove {
			if !strings.Contains(argument, "=") && index+1 < len(arguments) && !strings.HasPrefix(arguments[index+1], "-") {
				index++
			}
			continue
		}
		kept = append(kept, argument)
	}
	if name != "" {
		parts := strings.Split(name, "/")
		for index := range parts {
			parts[index] = "^" + regexp.QuoteMeta(parts[index]) + "$"
		}
		kept = append(kept, "-test.run="+strings.Join(parts, "/"))
	}
	return append(kept, "-test.timeout=0", "-test.count=1", "-test.paniconexit0=false", "-test.shuffle=off")
}

// replayBuild is entered only by the audit subprocess. Its completion marker and private exit code make it
// impossible for a test that simply passes without reaching the requested callback to count as an honest audit.
func replayBuild(inputs Inputs, build func(string) error) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	key, err := Key(root, inputs)
	if err != nil {
		return err
	}
	if key != os.Getenv(auditReplayKey) {
		return nil
	}
	directory := os.Getenv(auditReplayDirectory)
	if directory == "" {
		return errors.New("audit replay has no rebuild directory")
	}
	if err := build(directory); err != nil {
		return err
	}
	if err := os.WriteFile(directory+".complete", []byte(key), 0o600); err != nil {
		return err
	}
	os.Exit(auditRebuiltExit)
	return nil
}

// DrainAudits rebuilds and compares queued snapshots. A mismatch is a poisoning error naming the key, and stays
// queued for diagnosis. Successful entries are removed under their per-entry lock; concurrent drainers recheck
// after locking. Queuing another snapshot cannot replace a witness that a drainer is comparing.
func DrainAudits() error {
	paths, err := auditEntries()
	if err != nil {
		return err
	}
	var failures []error
	for _, path := range paths {
		if err := drainAudit(path); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", filepath.Base(path), err))
		}
	}
	return errors.Join(failures...)
}

func drainAudit(path string) error {
	lock, err := entryLock(path)
	if err != nil {
		return err
	}
	defer lock.Close()
	encoded, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var entry auditEntry
	if err := json.Unmarshal(encoded, &entry); err != nil {
		return err
	}
	key, err := Key(entry.Root, entry.Inputs)
	if err != nil {
		return err
	}
	if key != entry.Key {
		return fmt.Errorf("audit %s: checkout inputs changed (now %s)", entry.Key, key)
	}
	stored, err := describeProduct(entry.Key, entry.Inputs.Name, entry.Product)
	if err != nil {
		return err
	}
	before, _ := json.Marshal(entry.Files)
	now, _ := json.Marshal(stored.Files)
	if !bytes.Equal(before, now) {
		return fmt.Errorf("audit %s: fetched snapshot changed in the local cache", entry.Key)
	}
	got, err := describeProductBytes(entry.Key, entry.Inputs.Name, entry.Product, true)
	if err != nil {
		return err
	}
	workspace, err := os.MkdirTemp(filepath.Dir(path), ".audit-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)
	rebuilt := filepath.Join(workspace, "product")
	if err := os.Mkdir(rebuilt, 0o755); err != nil {
		return err
	}
	temporary := filepath.Join(workspace, "temporary")
	if err := os.Mkdir(temporary, 0o700); err != nil {
		return err
	}
	// Both compilation and callback execution fit within one audit unit's budget, never a test's timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
	defer cancel()
	executable := entry.Executable
	if entry.Test {
		executable = filepath.Join(workspace, "runner")
		command := exec.CommandContext(ctx, "go", "test", "-c", "-o", executable, ".")
		command.Dir = entry.Directory
		command.Env = append(os.Environ(), "TMPDIR="+temporary)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("compile audit %s: %w\n%s", entry.Key, err, output)
		}
	}
	command := exec.CommandContext(ctx, executable, entry.Arguments...)
	command.Dir = entry.Directory
	command.Env = append(os.Environ(), auditReplayKey+"="+entry.Key, auditReplayDirectory+"="+rebuilt, "TMPDIR="+temporary)
	output, err := command.CombinedOutput()
	var exited *exec.ExitError
	marker, markerErr := os.ReadFile(rebuilt + ".complete")
	if !errors.As(err, &exited) || exited.ExitCode() != auditRebuiltExit || markerErr != nil || string(marker) != entry.Key {
		return fmt.Errorf("audit %s did not complete its rebuild: %v (marker: %v)\n%s", entry.Key, err, markerErr, output)
	}
	want, err := describeProductBytes(entry.Key, entry.Inputs.Name, rebuilt, true)
	if err != nil {
		return err
	}
	if err := compareAudit(entry.Key, entry.Inputs.Name, got.Files, want.Files); err != nil {
		return err
	}
	return os.Remove(path)
}
