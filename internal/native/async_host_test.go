package native

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"
)

// Each subprocess builds its controls and mutants in a private temporary directory.
func asyncHostPromises(test *testing.T, sanitizer string) {
	if goruntime.GOOS != "linux" {
		test.Skip("host promise sanitizer evidence currently requires Linux LeakSanitizer")
	}
	logPath := filepath.Join(test.TempDir(), "host-promises.log")
	log, err := os.Create(logPath)
	if err != nil {
		test.Fatal(err)
	}
	command := exec.Command("python3", "testdata/async_host/check.py", "--sanitizer", sanitizer)
	command.Stdout, command.Stderr = log, log
	err = command.Run()
	if closeErr := log.Close(); closeErr != nil {
		test.Fatal(closeErr)
	}
	output, readErr := os.ReadFile(logPath)
	if readErr != nil {
		test.Fatal(readErr)
	}
	if err != nil {
		test.Fatalf("host bridge checks: %v\n%s", err, output)
	}
	test.Logf("%s", output)
}

func TestAsyncHostPromisesAddress(test *testing.T) {
	test.Parallel()
	asyncHostPromises(test, "address,undefined")
}

func TestAsyncHostPromisesThread(test *testing.T) {
	test.Parallel()
	asyncHostPromises(test, "thread")
}
