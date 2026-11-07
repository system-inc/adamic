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
	log, createError := os.Create(logPath)
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command("python3", "testdata/async_host/check.py", "--sanitizer", sanitizer)
	command.Stdout, command.Stderr = log, log
	runError := command.Run()
	if closeError := log.Close(); closeError != nil {
		test.Fatal(closeError)
	}
	output, readError := os.ReadFile(logPath)
	if readError != nil {
		test.Fatal(readError)
	}
	if runError != nil {
		test.Fatalf("host bridge checks: %v\n%s", runError, output)
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
