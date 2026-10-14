package native

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"
)

// Each subprocess builds its controls and mutants in a private temporary directory.
func asyncHostPromises(t *testing.T, sanitizer string) {
	if goruntime.GOOS != "linux" {
		t.Skip("host promise sanitizer evidence currently requires Linux LeakSanitizer")
	}
	logPath := filepath.Join(t.TempDir(), "host-promises.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "testdata/async_host/check.py", "--sanitizer", sanitizer)
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

func TestAsyncHostPromisesAddress(t *testing.T) {
	t.Parallel()
	asyncHostPromises(t, "address,undefined")
}

func TestAsyncHostPromisesThread(t *testing.T) {
	t.Parallel()
	asyncHostPromises(t, "thread")
}
