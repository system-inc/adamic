package buildcache

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Artifact names a required file relative to the product directory. Executable files
// are probed with Arguments and Stdin, and must exit with ExitCode (normally zero).
type Artifact struct {
	Name       string
	Executable bool
	Arguments  []string
	Stdin      string
	ExitCode   int
}

// A product test runs traced, and under --pin-traced the probe shares one vCPU with
// the tracer, so a large oracle's cold start needs room the box can't promise in 5 s.
const artifactProbeDeadline = 30 * time.Second

// RequireArtifacts proves that a builder returned usable products. Probes should be
// --help-like calls: whatever a probe reads joins the traced product's read set.
func RequireArtifacts(test testing.TB, directory string, artifacts ...Artifact) {
	test.Helper()
	for _, artifact := range artifacts {
		if requireArtifactError := requireArtifact(directory, artifact); requireArtifactError != nil {
			test.Fatalf("artifact %s: %v", artifact.Name, requireArtifactError)
		}
	}
}

func requireArtifact(directory string, artifact Artifact) error {
	if directory == "" || artifact.Name == "" || !filepath.IsLocal(artifact.Name) {
		return fmt.Errorf("directory and local artifact name are required")
	}
	path := filepath.Join(directory, artifact.Name)
	info, statError := os.Stat(path)
	if statError != nil {
		return statError
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("must be a non-empty regular file")
	}
	if !artifact.Executable {
		return nil
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), artifactProbeDeadline)
	defer cancel()
	command := exec.CommandContext(timeoutContext, path, artifact.Arguments...)
	command.Dir = directory
	command.Stdin = strings.NewReader(artifact.Stdin)
	if Tracing() {
		// A traced build watches its processes with ptrace, which LeakSanitizer refuses to run under, so a sanitized
		// product's probe would fail only because it is traced. The probe proves the product runs; the tests that
		// run it untraced still check for leaks (#n1g5604, #yzeez0d).
		command.Env = withoutLeakDetection(os.Environ())
	}
	// Each stream is read to its end, the probe's process group killed once it exits (childguard.Capture), so a child
	// that keeps an output pipe open can't hold the read past the probe, and no timer after the exit makes a slow drain
	// red by chance, as a WaitDelay did (#n3gk22h, adamiclint's waitdelay).
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		return fmt.Errorf("probe: %w", timeoutContext.Err())
	}
	if captureError != nil {
		if exit, ok := captureError.(*exec.ExitError); !ok || exit.ExitCode() != artifact.ExitCode {
			return fmt.Errorf("probe: %w: %s", captureError, output)
		}
	} else if artifact.ExitCode != 0 {
		return fmt.Errorf("probe exited 0, want %d", artifact.ExitCode)
	}
	return nil
}

// withoutLeakDetection returns environment with LeakSanitizer turned off, keeping every other ASAN_OPTIONS setting;
// a later detect_leaks wins over an earlier one.
func withoutLeakDetection(environment []string) []string {
	result := make([]string, 0, len(environment)+1)
	found := false
	for _, entry := range environment {
		if value, ok := strings.CutPrefix(entry, "ASAN_OPTIONS="); ok {
			found = true
			if value != "" {
				entry += ":"
			}
			entry += "detect_leaks=0"
		}
		result = append(result, entry)
	}
	if !found {
		result = append(result, "ASAN_OPTIONS=detect_leaks=0")
	}
	return result
}
