package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// `adamic build <file> -o <relative path>` leaves the binary at that path under the working directory. native links
// from inside its own temporary build directory, so a relative -o used to land in there and go with it, exiting 0
// with nothing written (Language, Oct 26). The relative path climbs from the working directory into a temporary one,
// so nothing is written into the tree, and the real builder links it. Mutant: drop buildWithNativeExports's
// filepath.Abs, and native is handed the relative path.
func TestBuildWritesARelativeOutputUnderTheWorkingDirectory(test *testing.T) {
	test.Parallel()
	directory, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		test.Fatal(workingDirectoryError)
	}
	binary := filepath.Join(test.TempDir(), "program")
	relative, relativePathError := filepath.Rel(directory, binary)
	if relativePathError != nil {
		test.Fatal(relativePathError)
	}
	source, absolutePathError := filepath.Abs("testdata/build_target.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	compile := func(source, output string, options native.Options) error {
		if output != binary {
			return errors.New("native was handed -o " + output + ", want " + binary + " under the working directory")
		}
		return native.Build(source, output, options)
	}
	if code := buildWithNative(source, relative, nil, compile); code != 0 {
		test.Fatalf("adamic build -o %s exit %d", relative, code)
	}
	result, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil || string(result) != "ready\n" {
		test.Fatalf(
			"the binary a relative -o names: %v, %q; want it under the working directory, printing ready",
			combinedOutputError,
			result,
		)
	}
}

// When the working directory can't be read, a relative -o has nowhere to go: that's the machine, so exit 3, and the
// message keeps -o as it was typed. The driver removes its own working directory before run. Mutants: return 1, and
// the message names the reassigned output.
func TestBuildExitsThreeWhenTheWorkingDirectoryIsGone(test *testing.T) {
	test.Parallel()
	if runtime.GOOS == "darwin" {
		// macOS's getcwd still answers with a removed directory's old path, so Abs succeeds there; Linux's says ENOENT.
		test.Skip("macOS reads a removed working directory")
	}
	source, absolutePathError := filepath.Abs("testdata/build_target.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := filepath.Join(test.TempDir(), "gone")
	if createDirectoryError := os.Mkdir(directory, 0o700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		os.Args[0],
		"-test.run=^TestRemovedWorkingDirectoryDriver$",
		"--",
		"build",
		source,
		"-o",
		"program",
	)
	command.Dir = directory
	command.Env = append(os.Environ(), "ADAMIC_REMOVED_DIRECTORY_TEST_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	code := 0
	if runError := command.Run(); runError != nil {
		var exit *exec.ExitError
		if !errors.As(runError, &exit) {
			test.Fatal(runError)
		}
		code = exit.ExitCode()
	}
	if code != 3 || !strings.HasPrefix(diagnostic.String(), "adamic: -o program: ") {
		test.Fatalf(
			"got exit %d, stderr %q; want exit 3 and stderr beginning %q",
			code,
			diagnostic.String(),
			"adamic: -o program: ",
		)
	}
}

// Not parallel: a driver process, which removes its own working directory and exits with run's status.
func TestRemovedWorkingDirectoryDriver(test *testing.T) {
	if os.Getenv("ADAMIC_REMOVED_DIRECTORY_TEST_DRIVER") != "1" {
		return
	}
	directory, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		test.Fatal(workingDirectoryError)
	}
	if removeError := os.Remove(directory); removeError != nil {
		test.Fatal(removeError)
	}
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	test.Fatal("missing driver arguments")
}
