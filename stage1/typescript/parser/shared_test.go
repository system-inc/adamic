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
	parserCPUChild()
	parserBuildChild()
	var err error
	parserSharedDirectory, err = os.MkdirTemp("", "adamic-parser-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	status := m.Run()
	fmt.Printf("parser longest native build child: %s; build CPU budget: %s\n", parserLongestBuild, parserBuildCPUBudget)
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
		command := exec.Command("go", "build", "-overlay="+path, "-o", binary, virtual)
		command.Dir = root
		if output, err := parserGuardOutput(command, parserBuildCPUBudget); err != nil || len(output) != 0 {
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
const parserBuildCPUBudget = 10 * time.Minute

func buildParserC(source, binary string, sanitize bool) error {
	path := binary + ".c"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		return err
	}
	defer os.Remove(path)
	request, err := json.Marshal(parserCheckerRequest{Source: path, Binary: binary, Sanitize: sanitize})
	if err != nil {
		return err
	}
	requestPath := binary + ".build.json"
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		return err
	}
	defer os.Remove(requestPath)
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command := exec.Command(executable)
	command.Env = append(os.Environ(), "ADAMIC_PARSER_C_REQUEST="+requestPath)
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		parserBuildTimingLock.Lock()
		if elapsed > parserLongestBuild {
			parserLongestBuild = elapsed
		}
		parserBuildTimingLock.Unlock()
	}()
	if output, err := parserGuardOutput(command, parserBuildCPUBudget); err != nil || len(output) != 0 {
		return fmt.Errorf("parser native build %s: %v: %s", binary, err, output)
	}
	return nil
}

// The helper's CPU limit is inherited by runtime compiler children too. The
// outer 60-minute backstop covers the complete runtime and program build.
func parserCompileC(request parserCheckerRequest) error {
	options := native.Options{Sanitize: request.Sanitize}
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		return err
	}
	flags := native.Flags(options)
	flags = append(flags, "-I", filepath.Dir(library), "-o", request.Binary, request.Source)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := parserGuardOutput(exec.Command("clang", flags...), parserBuildCPUBudget); err != nil || len(output) != 0 {
		return fmt.Errorf("parser clang build %s: %v: %s", request.Source, err, output)
	}
	return nil
}
