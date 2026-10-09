package buildcache

import (
	"bytes"
	"debug/macho"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Not parallel: shared changes the process build-store environment; the counter is inherited by replay.
func TestConcurrentAuditsRebuildOnce(t *testing.T) {
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	counter := os.Getenv("_ADAMIC_TEST_AUDIT_COUNTER")
	if counter == "" {
		counter = filepath.Join(t.TempDir(), "builds")
		t.Setenv("_ADAMIC_TEST_AUDIT_COUNTER", counter)
	}
	store.put(t, thisKey(t), map[string]string{"product": "built"})
	// Get has no testing.TB; the owning test must still be found and replayed correctly.
	_, err := Get(thisPackage, func(directory string) error {
		file, err := os.OpenFile(counter, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		_, err = file.WriteString("built\n")
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(counter); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rebuilt inline: %v", err)
	}
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			if err := DrainAudits(); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	content, err := os.ReadFile(counter)
	if err != nil || string(content) != "built\n" {
		t.Fatalf("concurrent audits rebuilt %q, %v", content, err)
	}
	if err := DrainAudits(); err != nil {
		t.Fatal(err)
	}
	content, err = os.ReadFile(counter)
	if err != nil || string(content) != "built\n" {
		t.Fatalf("empty retry rebuilt %q, %v", content, err)
	}
}

// Not parallel: shared and the two cache tiers are selected with process environment variables.
func TestAnHonestFetchCannotOverwriteAnAuditWitness(t *testing.T) {
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	for _, content := range []string{"wrong", "built"} {
		t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
		store.put(t, thisKey(t), map[string]string{"product": content})
		Product(t, thisPackage, func(directory string) error {
			return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
		})
	}
	entries, err := auditEntries()
	if err != nil || len(entries) != 2 {
		t.Fatalf("distinct snapshots: %v, %v", entries, err)
	}
	err = DrainAudits()
	var poisoned poisonedError
	if !errors.As(err, &poisoned) || !strings.Contains(err.Error(), "differs from a rebuild") {
		t.Fatalf("lost the wrong snapshot: %v", err)
	}
	entries, err = auditEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("audit witnesses after drain: %v, %v", entries, err)
	}
}

// Not parallel: shared changes the process build-store and audit environment.
func TestAnAuditThatNeverReachesItsBuildFails(t *testing.T) {
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	store.put(t, thisKey(t), map[string]string{"product": "built"})
	Product(t, thisPackage, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	entries, err := auditEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("audit entries: %v, %v", entries, err)
	}
	encoded, err := os.ReadFile(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	var entry auditEntry
	if err := json.Unmarshal(encoded, &entry); err != nil {
		t.Fatal(err)
	}
	entry.Arguments = []string{"-test.run=^NoSuchTest$"}
	encoded, err = json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entries[0], encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := DrainAudits(); err == nil || !strings.Contains(err.Error(), "did not complete its rebuild") {
		t.Fatalf("missing rebuild accepted: %v", err)
	}
	entries, err = auditEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("incomplete audit was dropped: %v, %v", entries, err)
	}
}

// Not parallel: shared selects the fake store and audit queue through process environment variables.
func TestQueuedMachOAuditsIgnoreOnlyLinkerMetadata(t *testing.T) {
	content := machOFixture(t)
	file, err := macho.NewFile(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	patches := map[string]int{}
	offset := 32
	for _, load := range file.Loads {
		raw := load.Raw()
		switch file.ByteOrder.Uint32(raw) {
		case 0x1b:
			patches["uuid"] = offset + 8
		case 0x1d:
			patches["signature"] = int(file.ByteOrder.Uint32(raw[8:12]))
		}
		offset += len(raw)
	}
	section := file.Section("__text")
	if section == nil || len(patches) != 2 {
		t.Fatal("fixture lacks linker metadata or __text")
	}
	patches["text"] = int(section.Offset)
	for name, patch := range patches {
		t.Run(name, func(t *testing.T) {
			store, _ := shared(t)
			t.Setenv("ADAMIC_BUILD_AUDIT", "1")
			changed := bytes.Clone(content)
			changed[patch] ^= 1
			store.put(t, thisKey(t), map[string]string{"product": string(changed)})
			directory := Product(t, thisPackage, func(directory string) error {
				return os.WriteFile(filepath.Join(directory, "product"), content, 0o644)
			})
			output, err := publisherOutput(t, "-audit")
			if name == "text" {
				if err == nil || !strings.Contains(string(output), "cache poisoning:") || !strings.Contains(string(output), "differs from a rebuild") {
					t.Fatalf("changed __TEXT audit: %v\n%s", err, output)
				}
			} else if err != nil {
				t.Fatalf("linker-only audit: %v\n%s", err, output)
			}
			after, err := os.ReadFile(filepath.Join(directory, "product"))
			if err != nil || !bytes.Equal(changed, after) {
				t.Fatalf("drain changed fetched bytes: %v", err)
			}
		})
	}
}

// Not parallel: shared changes the process store environment and the test holds queue entry locks.
func TestQueueWritesDoNotWaitForDrainers(t *testing.T) {
	store, _ := shared(t)
	token := filepath.Join(t.TempDir(), "token")
	write(t, filepath.Dir(token), filepath.Base(token), "gate-box-token")
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	directory := Product(t, thisPackage, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	spoolPaths, err := spoolEntries()
	if err != nil || len(spoolPaths) != 1 {
		t.Fatalf("spool: %v, %v", spoolPaths, err)
	}
	store.put(t, thisKey(t), map[string]string{"product": "built"})
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	fetched := Product(t, thisPackage, func(string) error { t.Fatal("rebuilt a fetch"); return nil })
	auditPaths, err := auditEntries()
	if err != nil || len(auditPaths) != 1 {
		t.Fatalf("audits: %v, %v", auditPaths, err)
	}
	for _, queue := range []struct {
		path  string
		write func() error
	}{
		{spoolPaths[0], func() error { return spool(thisKey(t), thisPackage.Name, directory) }},
		{auditPaths[0], func() error { return queueAudit(thisKey(t), thisPackage, fetched, []string{t.Name()}) }},
	} {
		lock, err := entryLock(queue.path)
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- queue.write() }()
		select {
		case err := <-done:
			lock.Close()
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			lock.Close()
			<-done
			t.Fatal("queuing waited for a drainer's store work")
		}
	}
}

// Not parallel: shared changes the process store and audit environment.
func TestAuditNormalizationErrorsStayOffTheTestsClock(t *testing.T) {
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	malformed := []byte{0xcf, 0xfa, 0xed, 0xfe}
	store.put(t, thisKey(t), map[string]string{"product": string(malformed)})
	directory := Product(t, thisPackage, func(string) error { t.Fatal("rebuilt inline"); return nil })
	content, err := os.ReadFile(filepath.Join(directory, "product"))
	if err != nil || !bytes.Equal(content, malformed) {
		t.Fatalf("fetched bytes: %x, %v", content, err)
	}
	output, err := publisherOutput(t, "-audit")
	if err == nil || !strings.Contains(string(output), thisKey(t)[:12]) {
		t.Fatalf("malformed audit was accepted: %v\n%s", err, output)
	}
	entries, err := auditEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("malformed witness lost: %v, %v", entries, err)
	}
}
