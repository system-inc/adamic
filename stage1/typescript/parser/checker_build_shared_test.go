package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

type parserCheckerRequest struct {
	Source, Binary, Archive string
	Sanitize                bool
}

// The checker builder drives several clang commands. A private compiler proxy
// gives each one a guard without changing the parent process's environment.
func parserBuildChild() {
	if real := os.Getenv("ADAMIC_PARSER_COMPILER"); real != "" && filepath.Base(os.Args[0]) == "clang" {
		ctx, cancel := context.WithTimeout(context.Background(), parserBuildDeadline)
		defer cancel()
		command := exec.CommandContext(ctx, real, os.Args[1:]...)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if path := os.Getenv("ADAMIC_PARSER_CHECKER_REQUEST"); path != "" {
		var request parserCheckerRequest
		data, err := os.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(data, &request)
		}
		if err == nil {
			data, err = os.ReadFile(request.Source)
		}
		if err == nil {
			err = native.BuildSplitTSGo(string(data), request.Binary, request.Archive, native.Options{Sanitize: request.Sanitize, Jobs: 1})
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func parserCompilerEnvironment(t *testing.T) []string {
	t.Helper()
	directory := parserShared(t, "compiler-guard", func() (string, error) {
		directory := filepath.Join(parserSharedDirectory, "compiler-guard")
		if err := os.Mkdir(directory, 0755); err != nil {
			return "", err
		}
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		if err := os.Symlink(executable, filepath.Join(directory, "clang")); err != nil {
			return "", err
		}
		return directory, nil
	})
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	return append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"), "ADAMIC_PARSER_COMPILER="+compiler)
}

func parserCheckerArchive(t *testing.T, sanitize bool) string {
	t.Helper()
	environment := parserCompilerEnvironment(t)
	return parserShared(t, fmt.Sprintf("checker-archive-%t", sanitize), func() (string, error) {
		archive := filepath.Join(parserSharedDirectory, fmt.Sprintf("checker-%t.a", sanitize))
		root, err := filepath.Abs(repository)
		if err != nil {
			return "", err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, "go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive")
		command.Dir = root
		command.Env = append(environment, "GOMAXPROCS=4")
		if sanitize {
			command.Env = append(command.Env, "CC="+filepath.Join(parserSharedDirectory, "compiler-guard", "clang"), "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
		}
		if output, err := command.CombinedOutput(); err != nil || len(output) != 0 {
			return "", fmt.Errorf("checker archive: %v: %s", err, output)
		}
		return archive, nil
	})
}

func buildParserChecker(t *testing.T, source, binary string, sanitize bool) {
	t.Helper()
	archive := parserCheckerArchive(t, sanitize)
	request := parserCheckerRequest{Source: binary + ".c", Binary: binary, Archive: archive, Sanitize: sanitize}
	if err := os.WriteFile(request.Source, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(request.Source)
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	path := binary + ".json"
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), parserBuildDeadline)
	defer cancel()
	command := exec.CommandContext(ctx, executable)
	command.Env = append(parserCompilerEnvironment(t), "ADAMIC_PARSER_CHECKER_REQUEST="+path)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	parserBuildTimingLock.Lock()
	if elapsed > parserLongestBuild {
		parserLongestBuild = elapsed
	}
	parserBuildTimingLock.Unlock()
	if err != nil || len(output) != 0 {
		t.Fatalf("checker native build: %v: %s", err, output)
	}
	t.Logf("checker native build: %s; one clang worker; guard %s", elapsed, parserBuildDeadline)
}
