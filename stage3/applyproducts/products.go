// Package applyproducts owns the four successive adapted TypeScript build products.
package applyproducts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

var boundaries = []string{"10", "40", "70", "99"}

type lookup func(buildcache.Inputs, func(string) error) string

func root() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if data, err := os.ReadFile(filepath.Join(directory, "go.mod")); err == nil && strings.HasPrefix(string(data), "module github.com/system-inc/adamic\n") {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no Adamic repository above working directory")
		}
		directory = parent
	}
}

func inputs(repository string, index int) (buildcache.Inputs, error) {
	files := []string{"stage3/source.json", "stage3/apply.py", "stage3/applyproducts/products.go", "stage3/api/package.json", "stage3/api/package-lock.json"}
	entries, err := os.ReadDir(filepath.Join(repository, "stage3/adapt"))
	if err != nil {
		return buildcache.Inputs{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name()[:2] <= boundaries[index] {
			files = append(files, "stage3/adapt/"+entry.Name())
		}
	}
	flags := []string{"through=" + boundaries[index], "--ignore-scripts", "--no-audit", "--no-fund", "--userconfig=/dev/null", "--globalconfig=empty"}
	for _, name := range []string{"NODE_OPTIONS", "NODE_DISABLE_COMPILE_CACHE", "LANG", "LC_ALL", "TZ", "npm_config_registry", "NPM_CONFIG_REGISTRY", "CENSUS_TYPESCRIPT", "TSC_ADAPT_TYPESCRIPT"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return buildcache.Inputs{Name: "stage3-adapted-" + boundaries[index], Files: files, Flags: flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("node", "--version"), buildcache.Tool("npm", "--version"), buildcache.Tool("python3", "--version"), buildcache.Tool("git", "--version"), buildcache.Tool("tar", "--version")}}, nil
}

func resolve(repository string, index int, get lookup) (string, error) {
	previous := ""
	if index > 0 {
		var err error
		previous, err = resolve(repository, index-1, get)
		if err != nil {
			return "", err
		}
	}
	key, err := inputs(repository, index)
	if err != nil {
		return "", err
	}
	return get(key, func(directory string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
		defer cancel()
		command := workerCommand(ctx, repository, "stage3/apply.py", "--step-worker", directory, boundaries[index], previous)
		// npm's cache location is not an input. Each cold product uses its own cache.
		command.Env = append(os.Environ(), "npm_config_cache="+filepath.Join(directory, "npm-cache"))
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("adapt through %s: %w\n%s", boundaries[index], err, output)
		}
		return nil
	}), nil
}

func workerCommand(ctx context.Context, repository string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, "python3", arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	command.Dir = repository
	return command
}

// Product is the single recipe used by each declaration and by fetches.
func Product(t testing.TB, index int) string {
	t.Helper()
	repository, err := root()
	if err != nil {
		t.Fatal(err)
	}
	directory, err := resolve(repository, index, func(inputs buildcache.Inputs, build func(string) error) string {
		return buildcache.Product(t, inputs, build)
	})
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

// Get lets the Python CLI fetch exactly the recipe declared by the Go tests.
func Get() (string, error) {
	repository, err := root()
	if err != nil {
		return "", err
	}
	var failure error
	directory, err := resolve(repository, 3, func(inputs buildcache.Inputs, build func(string) error) string {
		if failure != nil {
			return ""
		}
		result, err := buildcache.Get(inputs, build)
		if err != nil {
			failure = err
		}
		return result
	})
	if err != nil {
		return "", err
	}
	return directory, failure
}
