package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var checkerBuildMutex sync.Mutex
var checkerArchives = map[bool]string{}
var checkerBuildDirectories []string

// A single archive per instrumentation mode and test process, with its sources from the landed tree.
func checkerArchive(t *testing.T, sanitize bool) string {
	t.Helper()
	checkerBuildMutex.Lock()
	defer checkerBuildMutex.Unlock()
	if path := checkerArchives[sanitize]; path != "" {
		return path
	}
	directory, err := os.MkdirTemp("", "adamic-lint-checker-")
	if err != nil {
		t.Fatal(err)
	}
	checkerBuildDirectories = append(checkerBuildDirectories, directory)
	archive := filepath.Join(directory, "checker.a")
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive")
	command.Dir = root
	if sanitize {
		command.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	log, err := os.Create(filepath.Join(directory, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	log.Close()
	if err != nil {
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("checker archive: %v\n%s", err, data)
	}
	checkerArchives[sanitize] = archive
	return archive
}
func cleanupCheckerArchives() {
	for _, directory := range checkerBuildDirectories {
		os.RemoveAll(directory)
	}
}
