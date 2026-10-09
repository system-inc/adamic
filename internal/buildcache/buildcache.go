// Package buildcache is the one home for build products tests need: the checker archive, stage 0, Go oracle
// binaries, a port's native build. Each product is addressed by the hash of everything that can change it (its
// files by content, its flags, its toolchain), built once per hash and reused, so a test's time starts after the
// fetch (@system_adamic's ruling on the 30 s rule, Oct 8).
//
// A miss builds into an empty private directory and publishes it whole by rename, so a product on disk is
// always complete and a failed build leaves nothing. ADAMIC_BUILD_CACHE=off builds every time into a fresh
// directory: the uncached proof mode, which is what lands main. The cache is local to the machine
// (ADAMIC_BUILD_CACHE_DIR, or the user cache directory's adamic-build), and a miss there fetches from the shared
// store before it builds (store.go).
package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Inputs are everything that can change a product. Name is one per build function. Files are repository-relative
// files and directories, hashed by content (a directory by every name and file under it, sorted). Flags are every
// flag and environment value the build reads. Toolchain names the tools, from Tool or runtime.Version().
type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}

// Product returns the directory holding the product built from inputs, failing the test if it can't be built.
func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {
	t.Helper()
	directory, line, err := get(inputs, build)
	if line != "" {
		t.Log(line)
	}
	if err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	return directory
}

// Get is Product for TestMain and tools.
func Get(inputs Inputs, build func(directory string) error) (string, error) {
	directory, _, err := get(inputs, build)
	return directory, err
}

func get(inputs Inputs, build func(directory string) error) (string, string, error) {
	started := time.Now()
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		directory, err := os.MkdirTemp("", "adamic-build-")
		if err != nil {
			return "", "", err
		}
		if err = build(directory); err != nil {
			return "", "", err
		}
		// Main's own gate runs uncached and is the one writer of trusted refs: what it built from main's sources is
		// published for everyone to read (store.go).
		if trusted() {
			if root, err := repositoryRoot(); err == nil {
				if key, err := Key(root, inputs); err == nil {
					if err = publish(key, inputs.Name, directory); err != nil {
						note("publish %s %s failed: %v", inputs.Name, key[:12], err)
					}
				}
			}
		}
		return directory, record(inputs.Name, "uncached", "off", started), nil
	}
	root, err := repositoryRoot()
	if err != nil {
		return "", "", err
	}
	key, err := Key(root, inputs)
	if err != nil {
		return "", "", err
	}
	cache, err := cacheDirectory()
	if err != nil {
		return "", "", err
	}
	product := filepath.Join(cache, key)
	if _, err = os.Stat(product); err == nil {
		return product, record(inputs.Name, key, "hit", started), nil
	}
	// One builder per key across every process on the machine: the rest wait, then find it built.
	lock, err := os.OpenFile(product+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return "", "", err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return "", "", err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if _, err = os.Stat(product); err == nil {
		return product, record(inputs.Name, key, "hit", started), nil
	}
	scratch, err := os.MkdirTemp(cache, ".building-"+key[:12]+"-")
	if err != nil {
		return "", "", err
	}
	outcome := "miss"
	if os.Getenv("ADAMIC_BUILD_STORE") != "off" {
		switch err = fetch(key, scratch); {
		case err == nil:
			outcome = "fetched"
			if auditing() {
				if err = audit(key, inputs.Name, scratch, build); err != nil {
					os.RemoveAll(scratch)
					return "", "", err
				}
				outcome = "audited"
			}
		case errors.Is(err, errNotStored):
			// Not stored, or the store unreachable: build here, from an empty directory again.
			note("store %s %s: %v", inputs.Name, key[:12], err)
			os.RemoveAll(scratch)
			if scratch, err = os.MkdirTemp(cache, ".building-"+key[:12]+"-"); err != nil {
				return "", "", err
			}
		default:
			os.RemoveAll(scratch)
			return "", "", err
		}
	}
	if outcome == "miss" {
		if err = build(scratch); err != nil {
			os.RemoveAll(scratch)
			return "", "", err
		}
	}
	if err = os.WriteFile(product+".inputs", []byte(describe(inputs)), 0o644); err != nil {
		os.RemoveAll(scratch)
		return "", "", err
	}
	if err = os.Rename(scratch, product); err != nil {
		os.RemoveAll(scratch)
		return "", "", err
	}
	if outcome == "miss" {
		// Uploads are off the test's clock. The product must have its permanent name before it is spooled.
		if err = spool(key, inputs.Name, product); err != nil {
			note("spool %s %s failed: %v", inputs.Name, key[:12], err)
		}
	}
	return product, record(inputs.Name, key, outcome, started), nil
}

// Key is the product's address: a hash of every input, each length-prefixed so no two inputs run together.
func Key(root string, inputs Inputs) (string, error) {
	hash := sha256.New()
	field := func(kind, value string) {
		fmt.Fprintf(hash, "%s %d\n%s\n", kind, len(value), value)
	}
	field("buildcache", "v1")
	field("name", inputs.Name)
	for _, name := range inputs.Files {
		if err := hashPath(hash, root, name, field); err != nil {
			return "", err
		}
	}
	for _, flag := range inputs.Flags {
		field("flag", flag)
	}
	for _, tool := range inputs.Toolchain {
		field("tool", tool)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hashPath(hash io.Writer, root, name string, field func(kind, value string)) error {
	if filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..") {
		return fmt.Errorf("input %q must be inside the repository, relative to its root", name)
	}
	// WalkDir visits in lexical order, so the same tree always hashes the same.
	return filepath.WalkDir(filepath.Join(root, name), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			field("directory", relative)
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			field("symlink", relative+" -> "+target)
		case info.Mode().IsRegular():
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(content)
			field("file", fmt.Sprintf("%s %t %x", relative, info.Mode()&0o111 != 0, sum))
		default:
			return fmt.Errorf("input %s is neither a file, a directory nor a symlink", relative)
		}
		return nil
	})
}

var tools sync.Map

// Tool names a tool by its own report, such as Tool("clang", "--version"), running it once per process. A tool
// that can't run is named by its error, so the build that needs it fails rather than reusing another's product.
func Tool(name string, arguments ...string) string {
	command := strings.Join(append([]string{name}, arguments...), " ")
	if value, ok := tools.Load(command); ok {
		return value.(string)
	}
	output, err := exec.Command(name, arguments...).CombinedOutput()
	value := command + ": " + strings.TrimSpace(string(output))
	if err != nil {
		value += " (" + err.Error() + ")"
	}
	tools.Store(command, value)
	return value
}

func cacheDirectory() (string, error) {
	directory := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
	if directory == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		directory = filepath.Join(user, "adamic-build")
	}
	return directory, os.MkdirAll(directory, 0o755)
}

var root struct {
	once      sync.Once
	directory string
	err       error
}

// repositoryRoot is the nearest directory above the working directory holding this module's go.mod.
func repositoryRoot() (string, error) {
	root.once.Do(func() {
		directory, err := os.Getwd()
		if err != nil {
			root.err = err
			return
		}
		for {
			content, err := os.ReadFile(filepath.Join(directory, "go.mod"))
			if err == nil && strings.HasPrefix(string(content), "module github.com/system-inc/adamic\n") {
				root.directory = directory
				return
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				root.err = errors.New("buildcache: no adamic go.mod above the working directory")
				return
			}
			directory = parent
		}
	})
	return root.directory, root.err
}

func describe(inputs Inputs) string {
	lines := []string{"name " + inputs.Name}
	for _, name := range inputs.Files {
		lines = append(lines, "file "+name)
	}
	for _, flag := range inputs.Flags {
		lines = append(lines, "flag "+flag)
	}
	for _, tool := range inputs.Toolchain {
		lines = append(lines, "tool "+tool)
	}
	return strings.Join(lines, "\n") + "\n"
}

// note appends a line about the store to $ADAMIC_BUILD_LOG when set, beside the census lines.
func note(format string, arguments ...any) {
	if path := os.Getenv("ADAMIC_BUILD_LOG"); path != "" {
		if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			fmt.Fprintf(file, format+"\n", arguments...)
			file.Close()
		}
	}
}

// record is the product's census line, 'build <name> <key12> hit|fetched|audited|miss|off <seconds>', also appended to
// $ADAMIC_BUILD_LOG when set, so the gate can count every build as its own unit.
func record(name, key, outcome string, started time.Time) string {
	line := fmt.Sprintf("build %s %s %s %.2f", strings.ReplaceAll(name, " ", "_"), key[:min(12, len(key))], outcome, time.Since(started).Seconds())
	if path := os.Getenv("ADAMIC_BUILD_LOG"); path != "" {
		if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			file.WriteString(line + "\n")
			file.Close()
		}
	}
	return line
}
