// Package lane declares the shared upstream build and its first manifest shard.
package lane

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/stage3/applyproducts"
)

var phases = []string{"install", "compiler", "harness"}

func repositoryRoot(t testing.TB) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func laneInputs(t testing.TB, root string, index int) buildcache.Inputs {
	t.Helper()
	// artifact.input_key deliberately binds the broad stage3 source set. Installed
	// parser bytes are pinned by node_modules.sha256; runner markers are not inputs.
	files := []string{"stage3/lane/products.go", "stage3/lane/products.py", "stage3/lane/artifact.py", "stage3/lane/node_modules.sha256"}
	err := filepath.WalkDir(filepath.Join(root, "stage3"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if entry.IsDir() {
			if name == "stage3/lane" || name == "stage3/oracle" || entry.Name() == "__pycache__" || name == "stage3/api/node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, name)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	pin, err := os.ReadFile(filepath.Join(root, "stage3/lane/node_modules.sha256"))
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"phase=" + phases[index], "upstream-packages=" + strings.TrimSpace(string(pin)), "npm=offline"}
	for _, name := range []string{"NODE_OPTIONS", "NODE_DISABLE_COMPILE_CACHE", "LANG", "LC_ALL", "TZ"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return buildcache.Inputs{Name: "stage3-lane-" + phases[index], Files: files, Flags: flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("node", "--version"), buildcache.Tool("npm", "--version"),
			buildcache.Tool("python3", "--version"), buildcache.Tool("git", "--version"), buildcache.Tool("tar", "--version"), buildcache.Tool("gzip", "--version")}}
}

func boundedCommand(ctx context.Context, root string, program string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, program, arguments...)
	command.Dir = root
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 3 * time.Second
	return command
}

func fetchLane(t testing.TB, index int) string {
	t.Helper()
	return buildcache.Product(t, laneInputs(t, repositoryRoot(t), index), func(string) error {
		return fmt.Errorf("prepare TestProduct_Stage3Lane%s before fetching the lane artifact", phases[index])
	})
}

func laneProduct(t testing.TB, index int) string {
	t.Helper()
	root := repositoryRoot(t)
	var previous string
	if index == 0 {
		previous = applyproducts.Fetch(t)
	} else {
		previous = fetchLane(t, index-1)
	}
	return buildcache.Product(t, laneInputs(t, root, index), func(directory string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 58*time.Second)
		defer cancel()
		command := boundedCommand(ctx, root, "python3", "-B", "stage3/lane/products.py", directory, phases[index], previous)
		output, err := command.CombinedOutput()
		fmt.Fprintf(os.Stderr, "%s", output)
		if err != nil {
			return fmt.Errorf("lane %s: %w: %s", phases[index], err, output)
		}
		return nil
	})
}
