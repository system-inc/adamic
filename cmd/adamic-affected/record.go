package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

type recordCheckpoint struct {
	Record    recordFile
	Root      string
	Logs      string
	Inventory map[string]string
	Plan      []string
	Failed    map[string]string
	Isolation string
}

type failedAttempt struct {
	Logs   string
	Failed map[string]string
}

// A checkpoint is evidence for this unfinished reference run, never a branch
// result cache. Resume verifies all inputs and every finished event log first.
func recordMain(root, destination string, packages []packageInfo, jobs int, mainRef string, retryFailed bool, isolation string) error {
	if jobs < 1 {
		return fmt.Errorf("-jobs must be positive")
	}
	status, err := command(root, "git", "status", "--porcelain")
	if err != nil || len(status) != 0 {
		return fmt.Errorf("record requires a clean checkout: %w", err)
	}
	commit, err := command(root, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	main, err := command(root, "git", "rev-parse", mainRef)
	if err != nil {
		return err
	}
	if !bytes.Equal(commit, main) {
		return fmt.Errorf("record requires HEAD = %s", mainRef)
	}
	if _, err := command(root, "git", "merge-base", "--is-ancestor", strings.TrimSpace(string(main)), "origin/main"); err != nil {
		return fmt.Errorf("fixed reference is not a known main ancestor: %w", err)
	}
	tools, err := toolchain(root)
	if err != nil {
		return err
	}
	submodule, err := submoduleIdentity(root)
	if err != nil {
		return err
	}
	inventory := map[string]string{}
	if err := tree(root, root, inventory); err != nil {
		return err
	}
	plan := make([]string, len(packages))
	for i, pkg := range packages {
		plan[i] = pkg.ImportPath
	}
	if isolation != "" {
		found := false
		for _, name := range plan {
			found = found || name == isolation
		}
		if !found {
			return fmt.Errorf("isolated package is not in the gate: %s", isolation)
		}
	}
	checkpointPath := destination + ".partial"
	checkpoint := recordCheckpoint{Record: recordFile{Version: formatVersion, Commit: string(bytes.TrimSpace(commit)), Toolchain: tools, Submodule: submodule, Packages: map[string]closure{}, EventHashes: map[string]string{}, Reference: destination + ".reference.jsonl"}, Root: root, Inventory: inventory, Plan: plan}
	checkpoint.Isolation = isolation
	if contents, readErr := os.ReadFile(checkpointPath); readErr == nil {
		if err := json.Unmarshal(contents, &checkpoint); err != nil {
			return fmt.Errorf("invalid checkpoint: %w", err)
		}
		if checkpoint.Record.Version != formatVersion || checkpoint.Isolation != isolation || checkpoint.Root != root || checkpoint.Record.Commit != string(bytes.TrimSpace(commit)) || !reflect.DeepEqual(checkpoint.Record.Toolchain, tools) || checkpoint.Record.Submodule != submodule || !reflect.DeepEqual(checkpoint.Inventory, inventory) || !reflect.DeepEqual(checkpoint.Plan, plan) {
			return fmt.Errorf("checkpoint identity changed; refusing to mix reference runs")
		}
		if len(checkpoint.Failed) > 0 && !retryFailed {
			return fmt.Errorf("reference had failed packages; use a new -out or explicitly -retry-failed after diagnosing them")
		}
		if err := validateFinished(root, checkpoint.Record); err != nil {
			return err
		}
		if len(checkpoint.Failed) > 0 {
			checkpoint.Record.Retries = append(checkpoint.Record.Retries, failedAttempt{Logs: checkpoint.Logs, Failed: checkpoint.Failed})
			checkpoint.Logs, err = os.MkdirTemp(destination+".logs", "retry-")
			if err != nil {
				return err
			}
			checkpoint.Failed = nil
			if err := atomicJSON(checkpointPath, checkpoint); err != nil {
				return err
			}
		} else if !checkpoint.Record.Complete && len(checkpoint.Record.Packages) < len(packages) {
			// An interrupted process may have left unfinished logs without a
			// recorded failure. Never truncate those when resuming its packages.
			checkpoint.Logs, err = os.MkdirTemp(destination+".logs", "resume-")
			if err != nil {
				return err
			}
			if err := atomicJSON(checkpointPath, checkpoint); err != nil {
				return err
			}
		}
		if checkpoint.Record.Complete {
			if _, err := os.Stat(destination); err == nil {
				return fmt.Errorf("record already complete; use a new -out for a fresh uncached reference")
			} else if !os.IsNotExist(err) {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "resume %d/%d completed packages\n", len(checkpoint.Record.Packages), len(packages))
	} else if !os.IsNotExist(readErr) {
		return readErr
	} else {
		if err := os.MkdirAll(destination+".logs", 0755); err != nil {
			return err
		}
		checkpoint.Logs, err = os.MkdirTemp(destination+".logs", "run-")
		if err != nil {
			return err
		}
		if err := atomicJSON(checkpointPath, checkpoint); err != nil {
			return err
		}
	}
	observer, err := buildNotificationObserver(checkpoint.Logs)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if pkg.ImportPath == "github.com/system-inc/adamic/internal/native" {
			if err := observerCompatible(checkpoint.Logs, observer); err != nil {
				return err
			}
			break
		}
	}
	started := time.Now()
	prior := checkpoint.Record.WallSeconds
	type task struct {
		index int
		pkg   packageInfo
	}
	type result struct {
		name  string
		value closure
		err   error
	}
	tasks := make(chan task, len(packages))
	results := make(chan result, len(packages))
	var isolated *task
	ordinary := 0
	for i, pkg := range packages {
		if _, done := checkpoint.Record.Packages[pkg.ImportPath]; !done {
			job := task{i, pkg}
			if pkg.ImportPath == isolation {
				isolated = &job
			} else {
				tasks <- job
				ordinary++
			}
		}
	}
	if isolated == nil {
		close(tasks)
	} else if ordinary == 0 {
		tasks <- *isolated
		close(tasks)
	}
	var workers sync.WaitGroup
	for i := 0; i < jobs; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range tasks {
				fmt.Fprintf(os.Stderr, "record %d/%d %s\n", job.index+1, len(packages), job.pkg.ImportPath)
				stem := filepath.Join(checkpoint.Logs, fmt.Sprintf("%03d", job.index))
				value, err := recordPackage(root, job.pkg, stem, observer, inventory)
				results <- result{job.pkg.ImportPath, value, err}
			}
		}()
	}
	go func() { workers.Wait(); close(results) }()
	var firstError error
	completed := 0
	for result := range results {
		completed++
		if result.err != nil {
			if checkpoint.Failed == nil {
				checkpoint.Failed = map[string]string{}
			}
			checkpoint.Failed[result.name] = result.err.Error()
			checkpoint.Record.WallSeconds = prior + time.Since(started).Seconds()
			if err := atomicJSON(checkpointPath, checkpoint); err != nil {
				return err
			}
			if firstError == nil {
				firstError = result.err
			}
			fmt.Fprintln(os.Stderr, result.err)
			if isolated != nil && completed == ordinary {
				tasks <- *isolated
				close(tasks)
			}
			continue
		}
		checkpoint.Record.Packages[result.name] = result.value
		if result.value.Events != "" {
			contents, err := os.ReadFile(result.value.Events)
			if err != nil {
				return err
			}
			checkpoint.Record.EventHashes[result.name] = digest(contents)
		}
		checkpoint.Record.WallSeconds = prior + time.Since(started).Seconds()
		if err := writeReference(checkpoint.Record); err != nil {
			return err
		}
		if err := atomicJSON(checkpointPath, checkpoint); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "checkpoint %d/%d %s\n", len(checkpoint.Record.Packages), len(packages), result.name)
		if isolated != nil && completed == ordinary {
			tasks <- *isolated
			close(tasks)
		}
	}
	if firstError != nil {
		return firstError
	}
	after, err := toolchain(root)
	if err != nil || !reflect.DeepEqual(tools, after) {
		return fmt.Errorf("toolchain changed during record")
	}
	afterSubmodule, err := submoduleIdentity(root)
	if err != nil || submodule != afterSubmodule {
		return fmt.Errorf("submodule changed during record")
	}
	if err := validateFinished(root, checkpoint.Record); err != nil {
		return err
	}
	status, err = command(root, "git", "status", "--porcelain")
	if err != nil || len(status) != 0 {
		return fmt.Errorf("checkout changed during record")
	}
	checkpoint.Record.Complete = true
	checkpoint.Record.WallSeconds = prior + time.Since(started).Seconds()
	if err := atomicJSON(checkpointPath, checkpoint); err != nil {
		return err
	}
	return atomicJSON(destination, checkpoint.Record)
}

func recordPackage(root string, pkg packageInfo, stem, observer string, initial map[string]string) (closure, error) {
	static, err := staticInputs(root, pkg.ImportPath)
	if err != nil {
		return closure{}, err
	}
	value := closure{Static: static, Observed: map[string]string{}}
	dependencies, err := list(root, "-deps", "-test", pkg.ImportPath)
	if err != nil {
		return value, err
	}
	for _, dependency := range dependencies {
		if inside(root, dependency.Dir) && len(dependency.CgoFiles) > 0 {
			value.Uncertain = append(value.Uncertain, "repository cgo inputs may include headers outside Go's file list")
		}
	}
	binary := stem + ".test"
	if err := logged(root, stem+".build.log", "go", "test", "-c", "-o", binary, pkg.ImportPath); err != nil {
		return value, err
	}
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return value, nil
	} else if err != nil {
		return value, err
	}
	value.Events = stem + ".jsonl"
	if err := tracedRun(root, pkg, binary, stem, &value, observer); err != nil {
		return value, err
	}
	for path, after := range value.Observed {
		before, present := initial[path]
		if !present {
			before = "missing"
		}
		if before != after {
			return value, fmt.Errorf("observed input changed during %s: %s", pkg.ImportPath, path)
		}
	}
	after, err := staticInputs(root, pkg.ImportPath)
	if err != nil || !reflect.DeepEqual(static, after) {
		return value, fmt.Errorf("static inputs changed during %s", pkg.ImportPath)
	}
	return value, nil
}

func validateFinished(root string, record recordFile) error {
	if record.Packages == nil || record.EventHashes == nil {
		return fmt.Errorf("checkpoint missing evidence maps")
	}
	for name, value := range record.Packages {
		current, err := staticInputs(root, name)
		if err != nil || !reflect.DeepEqual(value.Static, current) || changed(root, value.Observed) {
			return fmt.Errorf("inputs changed for completed reference package %s", name)
		}
		if value.Events != "" {
			contents, err := os.ReadFile(value.Events)
			if err != nil || digest(contents) != record.EventHashes[name] {
				return fmt.Errorf("reference log changed for %s", name)
			}
		}
	}
	return nil
}

func atomicJSON(path string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return atomicFile(path, func(file *os.File) error { _, err := file.Write(contents); return err })
}
func atomicFile(path string, write func(*os.File) error) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".affected-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := write(file); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
func writeReference(record recordFile) error {
	return atomicFile(record.Reference, func(file *os.File) error {
		names := make([]string, 0, len(record.Packages))
		for name := range record.Packages {
			names = append(names, name)
		}
		// Sorting fixes concatenation order without altering any raw test event.
		sort.Strings(names)
		for _, name := range names {
			path := record.Packages[name].Events
			if path == "" {
				continue
			}
			input, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(file, input)
			closeErr := input.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
		return nil
	})
}
