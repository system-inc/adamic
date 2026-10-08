package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIncompleteCheckerAgrees(t *testing.T) {
	t.Parallel()
	compilerManifest(t)
	directory, _ := filepath.Abs(".")
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	artifacts := os.Getenv("ADAMIC_RECOVERY_ARTIFACTS")
	if artifacts == "" {
		artifacts = t.TempDir()
	}
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		t.Fatal(err)
	}
	for _, executable := range []struct{ name, path string }{{"oracle", oracle}, {"sanitized-native", binary}} {
		data, err := os.ReadFile(executable.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifacts, executable.name), data, 0755); err != nil {
			t.Fatal(err)
		}
	}
	sourcePath := filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/checker.ts")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	spans := execute(t, "", oracle, "--token-spans", sourcePath)
	fields := strings.Fields(string(spans.output))
	for _, mode := range []string{"cut", "duplicate"} {
		index := 189457
		if mode == "duplicate" {
			index = 196255
		}
		start, _ := strconv.Atoi(fields[(index-1)*2])
		end, _ := strconv.Atoi(fields[(index-1)*2+1])
		input := source[:end]
		if mode == "duplicate" {
			input = append(append(append([]byte(nil), source[:end]...), ' '), source[start:]...)
		}
		path := filepath.Join(artifacts, fmt.Sprintf("checker.ts-%s-%d.ts", mode, index))
		if err := os.WriteFile(path, input, 0644); err != nil {
			t.Fatal(err)
		}
		want, err := recoveryRunLimit(t, incompleteDeadline(len(input)), path+".go", oracle, path, "--whole", "--recovery")
		if err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole", "--recovery"}},
			{"native", binary, []string{path, "--whole", "--recovery"}},
		} {
			started := time.Now()
			got, err := recoveryRunLimit(t, incompleteDeadline(len(input)), path+"."+side.name, side.command, side.args...)
			t.Logf("%s %s: %d input bytes, %d output bytes, %s, %v", side.name, mode, len(input), len(got), time.Since(started), err)
			t.Logf("%s deadline: %s", mode, incompleteDeadline(len(input)))
			if err != nil {
				t.Error(err)
			} else if !bytes.Equal(got, want) {
				t.Errorf("%s %s: %s", side.name, mode, difference(got, want))
			}
		}
		if mode == "cut" {
			var workers sync.WaitGroup
			for worker := 0; worker < 4; worker++ {
				workers.Add(1)
				go func(worker int) {
					defer workers.Done()
					started := time.Now()
					got, err := recoveryRunLimit(t, incompleteDeadline(len(input)), fmt.Sprintf("%s.concurrent-%d", path, worker), binary, path, "--whole", "--recovery")
					t.Logf("four-native concurrency worker %d: %s, %v", worker, time.Since(started), err)
					if err != nil {
						t.Error(err)
					} else if !bytes.Equal(got, want) {
						t.Error("concurrent output differs")
					}
				}(worker)
			}
			workers.Wait()
		}
	}
}
