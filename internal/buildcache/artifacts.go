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

// RequireArtifacts proves that a builder returned usable products.
func RequireArtifacts(t testing.TB, directory string, artifacts ...Artifact) {
	t.Helper()
	for _, artifact := range artifacts {
		if err := requireArtifact(directory, artifact); err != nil {
			t.Fatalf("artifact %s: %v", artifact.Name, err)
		}
	}
}

func requireArtifact(directory string, artifact Artifact) error {
	if directory == "" || artifact.Name == "" || !filepath.IsLocal(artifact.Name) {
		return fmt.Errorf("directory and local artifact name are required")
	}
	path := filepath.Join(directory, artifact.Name)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("must be a non-empty regular file")
	}
	if !artifact.Executable {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, path, artifact.Arguments...)
	command.Dir = directory
	command.Stdin = strings.NewReader(artifact.Stdin)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("probe: %w", ctx.Err())
	}
	if err != nil {
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != artifact.ExitCode {
			return fmt.Errorf("probe: %w: %s", err, output)
		}
	} else if artifact.ExitCode != 0 {
		return fmt.Errorf("probe exited 0, want %d", artifact.ExitCode)
	}
	return nil
}
