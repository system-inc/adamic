package regex

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var checkerDirectory string
var checkerArchivePath string

func TestMain(m *testing.M) {
	code := m.Run()
	if checkerDirectory != "" {
		os.RemoveAll(checkerDirectory)
	}
	os.Exit(code)
}
func regexCheckerArchive(t *testing.T, repository string) string {
	t.Helper()
	if checkerArchivePath != "" {
		return checkerArchivePath
	}
	var err error
	checkerDirectory, err = os.MkdirTemp("", "regex-checker-")
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(checkerDirectory, "checker.a")
	command := exec.Command("go", "build", "-buildmode=c-archive", "-ldflags=-w", "-o", archive, "./bridge/tsgo/archive")
	command.Dir = repository
	command.Env = append(os.Environ(), "GOMAXPROCS=4", "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	log, err := os.Create(filepath.Join(checkerDirectory, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	log.Close()
	if err != nil {
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("sanitized checker archive: %v\n%s", err, data)
	}
	checkerArchivePath = archive
	return checkerArchivePath
}
