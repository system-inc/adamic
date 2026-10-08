package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Healthy builds belong to this run, not a test's TempDir. Mutant directories
// always build independently so sharing cannot conceal their changes.
var parserSharedDirectory string
var parserPackageDirectory, _ = filepath.Abs(".")
var parserSharedValues sync.Map
var parserBuildTimingLock sync.Mutex
var parserLongestBuild time.Duration

type parserSharedValue struct {
	once sync.Once
	path string
	err  error
}

func TestMain(m *testing.M) {
	parserBuildChild()
	var err error
	parserSharedDirectory, err = os.MkdirTemp("", "adamic-parser-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	status := m.Run()
	fmt.Printf("parser longest native build child: %s; build hang guard: %s\n", parserLongestBuild, parserBuildDeadline)
	if err := os.RemoveAll(parserSharedDirectory); err != nil {
		fmt.Fprintln(os.Stderr, err)
		status = 1
	}
	os.Exit(status)
}

func parserShared(t *testing.T, key string, makeValue func() (string, error)) string {
	t.Helper()
	stored, _ := parserSharedValues.LoadOrStore(key, &parserSharedValue{})
	value := stored.(*parserSharedValue)
	value.once.Do(func() { value.path, value.err = makeValue() })
	if value.err != nil {
		t.Fatal(value.err)
	}
	return value.path
}

func sharedParserOracle(t *testing.T) string {
	t.Helper()
	return parserShared(t, "oracle", func() (string, error) {
		root, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
		if err != nil {
			return "", err
		}
		side, err := filepath.Abs("testdata/oracle.go")
		if err != nil {
			return "", err
		}
		virtual := filepath.Join(root, "adamic_parser_oracle.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return "", err
		}
		path := filepath.Join(parserSharedDirectory, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return "", err
		}
		binary := filepath.Join(parserSharedDirectory, "oracle")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, "go", "build", "-overlay="+path, "-o", binary, virtual)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil || len(output) != 0 {
			return "", fmt.Errorf("oracle build: %v: %s", err, output)
		}
		return binary, nil
	})
}

func sharedParserPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	return parserShared(t, fmt.Sprintf("native-%t", sanitize), func() (string, error) {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return "", err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return "", err
		}
		binary := filepath.Join(parserSharedDirectory, fmt.Sprintf("native-%t", sanitize))
		if err := buildParserC(native.C(lowered), binary, sanitize); err != nil {
			return "", err
		}
		return binary, nil
	})
}

// Use native's normal flags and runtime, including both sanitizers. Clang is a
// child too, so its deadline is a hang guard, not the package limit.
const parserBuildDeadline = 4 * time.Minute

func buildParserC(source, binary string, sanitize bool) error {
	options := native.Options{Sanitize: sanitize}
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		return err
	}
	path := binary + ".c"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		return err
	}
	defer os.Remove(path)
	flags := native.Flags(options)
	flags = append(flags, "-I", filepath.Dir(library), "-o", binary, path)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	ctx, cancel := context.WithTimeout(context.Background(), parserBuildDeadline)
	defer cancel()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		parserBuildTimingLock.Lock()
		if elapsed > parserLongestBuild {
			parserLongestBuild = elapsed
		}
		parserBuildTimingLock.Unlock()
	}()
	command := exec.CommandContext(ctx, "clang", flags...)
	if output, err := command.CombinedOutput(); err != nil || len(output) != 0 {
		return fmt.Errorf("parser clang build: %v: %s", err, output)
	}
	return nil
}
