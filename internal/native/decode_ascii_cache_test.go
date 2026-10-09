package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Decoder setup uses the same content keys as the test cache, but never waits
// indefinitely on another process. Publish only completed products by rename.
func decodeCacheGet(ctx context.Context, t *testing.T, repository string) func(buildcache.Inputs, func(string) error) (string, error) {
	return func(in buildcache.Inputs, build func(string) error) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		root := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
		if root == "" {
			user, err := os.UserCacheDir()
			if err != nil {
				return "", err
			}
			root = filepath.Join(user, "adamic-build")
		}
		if err := os.MkdirAll(root, 0755); err != nil {
			return "", err
		}
		if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
			directory, err := os.MkdirTemp(root, ".decode-uncached-")
			if err != nil {
				return "", err
			}
			t.Cleanup(func() { os.RemoveAll(directory) })
			return directory, build(directory)
		}
		key, err := buildcache.Key(repository, in)
		if err != nil {
			return "", err
		}
		product := filepath.Join(root, key)
		if _, err := os.Stat(product); err == nil {
			return product, nil
		}
		lock, err := os.OpenFile(product+".lock", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return "", err
		}
		defer lock.Close()
		started := time.Now()
		for {
			err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			if err == nil {
				break
			}
			if !errors.Is(err, syscall.EWOULDBLOCK) {
				return "", err
			}
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("cache lock %s waited %.3fs: %w", in.Name, time.Since(started).Seconds(), ctx.Err())
			case <-time.After(10 * time.Millisecond):
			}
		}
		defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		t.Logf("cache lock %s wait %.3fs", in.Name, time.Since(started).Seconds())
		if _, err := os.Stat(product); err == nil {
			return product, nil
		}
		directory, err := os.MkdirTemp(root, ".decode-building-")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(directory)
		if err := build(directory); err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if err := os.Rename(directory, product); err != nil {
			return "", err
		}
		return product, nil
	}
}

func buildDecodeObjects(ctx context.Context, cache *testBuildCache, repository, compiler, target string, flags, sources []string, snapshot []runtimeFile, contents [][]byte) (string, error) {
	inputs := append([][]byte(nil), contents...)
	inputs = append(inputs, []byte(compiler), []byte(strings.Join(flags, "\x00")))
	entry, err := cache.Tree("decode-parallel-"+target, inputs, func(destination string) error {
		// Headers affect every object; C source content affects only its own object.
		var headers [][]byte
		for _, file := range snapshot {
			if !strings.HasSuffix(file.name, ".c") {
				headers = append(headers, []byte(file.name), file.contents)
			}
		}
		files := append([]string(nil), sources...)
		files = append(files, filepath.Join(repository, "internal/native/decode_ascii/probe.c"), filepath.Join(repository, "internal/native/decode_ascii/baseline.c"))
		objects := make([]string, len(files))
		failures := make([]error, len(files))
		jobs := make(chan int)
		var workers sync.WaitGroup
		// One owner builds the binary; siblings wait without launching more compilers.
		for worker := 0; worker < 4; worker++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for index := range jobs {
					object := filepath.Join(destination, filepath.Base(files[index])+".o")
					objects[index] = object
					compileFlags := make([]string, 0, len(flags))
					for _, flag := range flags {
						if !strings.HasPrefix(flag, "-Wl,") {
							compileFlags = append(compileFlags, flag)
						}
					}
					args := append(compileFlags, "-c", files[index], "-o", object)
					command := exec.CommandContext(ctx, compiler, args...)
					command.Dir = repository
					_, failures[index] = cache.Command(command, destination, headers...)
				}
			}()
		}
		for index := range files {
			jobs <- index
		}
		close(jobs)
		workers.Wait()
		for index, err := range failures {
			if err != nil {
				return fmt.Errorf("compile %s: %w", files[index], err)
			}
		}
		args := append(append([]string(nil), flags...), objects...)
		args = append(args, "-lm", "-o", filepath.Join(destination, "decode-"+target))
		command := exec.CommandContext(ctx, compiler, args...)
		command.Dir = repository
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("decoder link: %w\n%s", err, output)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return filepath.Join(entry, "decode-"+target), nil
}

// Hold the same flock as a sibling process: deadline expiry must return an
// identifiable lock wait error instead of hanging until the test kill.
func TestDecodeCacheLockDeadline(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	in := buildcache.Inputs{Name: t.TempDir()}
	key, err := buildcache.Key(repository, in)
	if err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		t.Skip("cache disabled")
	}
	if root == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			t.Fatal(err)
		}
		root = filepath.Join(user, "adamic-build")
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, key+".lock")
	lock, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err = decodeCacheGet(ctx, t, repository)(in, func(string) error { t.Error("built while sibling held lock"); return nil })
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "cache lock") || !strings.Contains(err.Error(), "waited") {
		t.Fatalf("missing bounded lock diagnostic: %v", err)
	}
}
