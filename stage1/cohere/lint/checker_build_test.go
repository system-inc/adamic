package lint

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var checkerBuildMutex sync.Mutex
var checkerArchives = map[bool]string{}
var checkerBinaryMutex sync.Mutex

type checkerBuiltBinary struct {
	path  string
	err   error
	ready chan struct{}
}

var checkerBinaries = map[string]*checkerBuiltBinary{}
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
	command.Env = append(os.Environ(), "GOMAXPROCS=4")
	if sanitize {
		command.Env = append(command.Env, "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
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

// Identical generated C and build options share one binary for this test process.
// Mutants change the source hash, so each still compiles and runs independently.
func checkerBinary(t *testing.T, source, archive string, sanitize bool) string {
	t.Helper()
	key := fmt.Sprintf("%x:%s:%t", sha256.Sum256([]byte(source)), archive, sanitize)
	checkerBinaryMutex.Lock()
	built, found := checkerBinaries[key]
	if !found {
		built = &checkerBuiltBinary{ready: make(chan struct{})}
		checkerBinaries[key] = built
	}
	checkerBinaryMutex.Unlock()
	if found {
		<-built.ready
	} else {
		// Share identical builds while allowing the existing bounded mutant workers
		// to compile different sources independently. Publish failures before t.Fatal.
		built.path, built.err = compileCheckerBinary(source, archive, sanitize)
		close(built.ready)
	}
	if built.err != nil {
		t.Fatal(built.err)
	}
	return built.path
}

func compileCheckerBinary(source, archive string, sanitize bool) (string, error) {
	directory, err := os.MkdirTemp("", "adamic-lint-native-")
	if err != nil {
		return "", err
	}
	checkerBuildMutex.Lock()
	checkerBuildDirectories = append(checkerBuildDirectories, directory)
	checkerBuildMutex.Unlock()
	path := filepath.Join(directory, "scanner")
	if archive != "" {
		err = buildCheckerWithRuntime(source, path, archive, native.Options{Sanitize: sanitize})
	} else {
		err = native.Build(source, path, native.Options{Sanitize: sanitize})
	}
	return path, err
}

func cleanupCheckerArchives() {
	for _, directory := range checkerBuildDirectories {
		os.RemoveAll(directory)
	}
}

// Use the landed split builder's content-keyed runtime and object caches.
// One clang worker per build preserves TestMutants' existing four-build bound.
func buildCheckerWithRuntime(source, output, archive string, options native.Options) error {
	options.Jobs = 1
	return native.BuildSplitTSGo(source, output, archive, options)
}
