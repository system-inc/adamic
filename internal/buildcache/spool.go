package buildcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// The gate should run 'go run ./internal/buildcache/cmd/buildcache-publish' after run.py's test units have
// finished, before its final verdict. It is a separate gate unit, with its own clock. Keep the local cache until
// it has drained: entries refer to complete, immutable products there. run.py is deliberately not wired here.
// Run 'go run ./internal/buildcache/cmd/buildcache-publish -audit' at that same point as another gate unit.
// Keep the checkout and build environment until audits drain, so their inputs and callbacks can be replayed.
type spoolEntry struct {
	Key       string
	Name      string
	Directory string
	Namespace string
}

func spoolDirectory() (string, error) {
	directory := os.Getenv("ADAMIC_BUILD_STORE_SPOOL")
	if directory == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		directory = filepath.Join(user, "adamic", "build-spool")
	}
	return directory, os.MkdirAll(directory, 0o700)
}

func entryLock(path string) (*os.File, error) {
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		lock.Close()
		return nil, err
	}
	// Do not unlink locks: a waiting process must keep locking the same inode.
	return lock, nil
}

func spool(key, name, product string) error {
	if publishToken() == "" || os.Getenv("ADAMIC_BUILD_STORE") == "off" {
		return nil
	}
	directory, err := spoolDirectory()
	if err != nil {
		return err
	}
	product, err = filepath.Abs(product)
	if err != nil {
		return err
	}
	entry := spoolEntry{Key: key, Name: name, Directory: product, Namespace: writeNamespace()}
	path := filepath.Join(directory, entry.Namespace+"-"+key+".json")
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return writeQueuedEntry(path, encoded)
}

// Entries are immutable and become visible atomically. Writers never wait for a drainer's lock: an upload or
// rebuild already in progress must not put its work back on a test's clock. The first snapshot for a path wins.
func writeQueuedEntry(path string, encoded []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".queuing-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(encoded); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	err = os.Link(file.Name(), path)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	return err
}

func spoolEntries() ([]string, error) {
	directory, err := spoolDirectory()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	return paths, nil
}

// PublishSpool drains deferred uploads. Each entry is locked through its upload and removal. A failed upload
// leaves the entry for retry; the store's content-addressed blobs and immutable refs make retries idempotent.
func PublishSpool() error {
	paths, err := spoolEntries()
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return nil
	}
	token := publishToken()
	if token == "" || os.Getenv("ADAMIC_BUILD_STORE") == "off" {
		return errors.New("buildcache: cannot drain spool without a store write credential and enabled store")
	}
	var failures []error
	for _, path := range paths {
		if err := publishEntry(path, token); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", filepath.Base(path), err))
		}
	}
	return errors.Join(failures...)
}

func publishEntry(path, token string) error {
	lock, err := entryLock(path)
	if err != nil {
		return err
	}
	defer lock.Close()
	encoded, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} // Another drainer finished while we waited.
	if err != nil {
		return err
	}
	var entry spoolEntry
	if err = json.Unmarshal(encoded, &entry); err != nil {
		return err
	}
	if entry.Namespace != "build" && entry.Namespace != "build-candidate" {
		return fmt.Errorf("invalid refs namespace %q", entry.Namespace)
	}
	if err = publishWithToken(entry.Key, entry.Name, entry.Directory, entry.Namespace, token); err != nil {
		return err
	}
	return os.Remove(path)
}
