package native

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"
)

func TestHostPromises(t *testing.T) {
	t.Parallel()
	// Not parallel: this compiles both sanitizer runtimes and executes bounded
	// subprocess controls and mutants; simultaneous gates obscure race evidence.
	if goruntime.GOOS != "linux" {
		t.Skip("host promise sanitizer evidence currently requires Linux LeakSanitizer")
	}
	logPath := filepath.Join(t.TempDir(), "host-promises.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "testdata/async_host/check.py")
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	if closeErr := log.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	output, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		t.Fatalf("host bridge checks: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}
