package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var repositoryCorpusRoots = []string{
	"bench", "bridge", "cmd", "internal", "stage1", "stage3/drivers/tsc/corpus",
}

func trackedRootFiles(t *testing.T, checkout, root string) ([]string, error) {
	t.Helper()
	info, err := os.Stat(filepath.Join(checkout, root))
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("corpus root %s: missing directory (%v)", root, err)
	}
	command := bounded(t, "git", "-C", checkout, "ls-files", "-z", "--", root)
	data, err := output(command)
	if err != nil {
		return nil, fmt.Errorf("corpus root %s: git ls-files: %w", root, err)
	}
	var files []string
	for _, path := range strings.Split(string(data), "\x00") {
		if strings.HasSuffix(path, ".ts") {
			files = append(files, filepath.Join(checkout, path))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("corpus root %s: no git-tracked .ts files", root)
	}
	t.Logf("corpus root %s: %d git-tracked .ts files", root, len(files))
	return files, nil
}

func printerCorpusFiles(t *testing.T) []string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to a TypeScript 6.0.3 source checkout at 050880ce; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	command := bounded(t, "git", "-C", source, "rev-parse", "HEAD")
	if data, err := output(command); err != nil || strings.TrimSpace(string(data)) != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("TypeScript pin: %q %v", data, err)
	}
	var files []string
	for _, path := range repositoryCorpusRoots {
		listed, err := trackedRootFiles(t, root, path)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, listed...)
	}
	listed, err := trackedRootFiles(t, source, "src/compiler")
	if err != nil {
		t.Fatal(err)
	}
	return append(files, listed...)
}

func TestTrackedCorpusRoots(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if result := execute(t, nil, "git", "-C", directory, "init", "-q"); result.exitCode != 0 {
		t.Fatal(string(result.stderr))
	}
	if err := os.Mkdir(filepath.Join(directory, "named"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("x;\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if result := execute(t, nil, "git", "-C", directory, "add", "named/tracked.ts", "outside.ts"); result.exitCode != 0 {
		t.Fatal(string(result.stderr))
	}
	files, err := trackedRootFiles(t, directory, "named")
	if err != nil || len(files) != 1 || files[0] != filepath.Join(directory, "named/tracked.ts") {
		t.Fatalf("named tracked corpus: %v %v", files, err)
	}
	if _, err := trackedRootFiles(t, directory, "missing"); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing root was not named: %v", err)
	}
	if err := os.Mkdir(filepath.Join(directory, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := trackedRootFiles(t, directory, "empty"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty root was not named: %v", err)
	}
}
