// Package testbuildcache is the temporary native-test product store. Its public
// API matches internal/buildcache so callers can change one import when it lands.
package testbuildcache

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Inputs struct {
	Name                    string
	Files, Flags, Toolchain []string
}

func Get(inputs Inputs, build func(string) error) (string, error) {
	h := sha256.New()
	field := func(value string) { fmt.Fprintf(h, "%d:%s", len(value), value) }
	field(inputs.Name)
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("no go.mod above working directory")
		}
		root = parent
	}
	for _, name := range inputs.Files {
		if filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..") {
			return "", fmt.Errorf("input must be repository-relative: %s", name)
		}
		err := filepath.WalkDir(filepath.Join(root, name), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			field(relative)
			if entry.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			field(fmt.Sprintf("%x", sha256.Sum256(data)))
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	for _, value := range inputs.Flags {
		field(value)
	}
	for _, value := range inputs.Toolchain {
		field(value)
	}
	key := fmt.Sprintf("%x", h.Sum(nil))
	cache := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
	if cache == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		cache = filepath.Join(base, "adamic-build")
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", err
	}
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		scratch, err := os.MkdirTemp("", "native-test-build-")
		if err != nil {
			return "", err
		}
		if err := build(scratch); err != nil {
			os.RemoveAll(scratch)
			return "", err
		}
		return scratch, nil
	}
	product := filepath.Join(cache, key)
	if _, err := os.Stat(product); err == nil {
		return product, nil
	}
	lock, err := os.OpenFile(product+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return "", err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return "", err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if _, err := os.Stat(product); err == nil {
		return product, nil
	}
	scratch, err := os.MkdirTemp(cache, ".building-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(scratch)
	started := time.Now()
	if err := build(scratch); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(scratch, "complete"), []byte(fmt.Sprintf("%s %s %.3f\n", inputs.Name, key, time.Since(started).Seconds())), 0644); err != nil {
		return "", err
	}
	if err := os.Rename(scratch, product); err != nil {
		return "", err
	}
	if path := os.Getenv("ADAMIC_BUILD_LOG"); path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return "", err
		}
		_, err = fmt.Fprintf(file, "build %s %s miss %.3f\n", inputs.Name, key[:12], time.Since(started).Seconds())
		file.Close()
		if err != nil {
			return "", err
		}
	}
	return product, nil
}

var tools sync.Map

func Tool(name string, arguments ...string) string {
	key := strings.Join(append([]string{name}, arguments...), " ")
	if value, ok := tools.Load(key); ok {
		return value.(string)
	}
	data, err := exec.Command(name, arguments...).CombinedOutput()
	value := key + ": " + strings.TrimSpace(string(data))
	if err != nil {
		value += " " + err.Error()
	}
	tools.Store(key, value)
	return value
}
