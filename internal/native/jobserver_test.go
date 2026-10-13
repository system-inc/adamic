package native

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// tokens makes a jobserver FIFO holding count tokens, names it in ADAMIC_CC_JOBSERVER for the test, and gives native a
// fresh jobserver to open it. Not parallel: the variable and the jobserver are the process's.
func tokens(t *testing.T, count int) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jobserver")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	// Opened for reading and writing, as native opens it, so filling it doesn't wait for a reader.
	fifo, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { fifo.Close() })
	if _, err := fifo.Write([]byte(strings.Repeat("+", count))); err != nil {
		t.Fatal(err)
	}
	t.Setenv(JobserverVariable, path)
	jobserver = &compilerJobserver{}
	t.Cleanup(func() { jobserver = &compilerJobserver{} })
}

// With two tokens, never more than two compiles run at once however many ask, and both tokens are used (#cyew6p3).
// Mutant: the token never taken (eight at once).
func TestCompilesHoldAJobserverTokenEach(t *testing.T) {
	tokens(t, 2)
	var running, most atomic.Int32
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			err := withCompilerToken(func() error {
				now := running.Add(1)
				for {
					seen := most.Load()
					if now <= seen || most.CompareAndSwap(seen, now) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				running.Add(-1)
				return nil
			})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if most.Load() != 2 {
		t.Fatalf("%d compiles ran at once on two tokens", most.Load())
	}
}

// A compile that fails gives its token back, so the next compile on a one-token jobserver still runs. Mutant: the token
// returned only on success (the next compile waits forever).
func TestAFailedCompileGivesItsTokenBack(t *testing.T) {
	tokens(t, 1)
	failed := errors.New("clang failed")
	if err := withCompilerToken(func() error { return failed }); !errors.Is(err, failed) {
		t.Fatalf("the compile's error came back as %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- withCompilerToken(func() error { return nil }) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the next compile is still waiting for the token the failed one held")
	}
}

// Unset, a compile runs as before; a jobserver that can't be opened fails the compile naming it, never runs it
// unbounded. Mutant: an unopenable jobserver ignored.
func TestAJobserverThatCantBeOpenedFailsTheCompile(t *testing.T) {
	t.Setenv(JobserverVariable, "")
	jobserver = &compilerJobserver{}
	t.Cleanup(func() { jobserver = &compilerJobserver{} })
	ran := false
	if err := withCompilerToken(func() error { ran = true; return nil }); err != nil || !ran {
		t.Fatalf("unset: ran %v, %v", ran, err)
	}
	missing := filepath.Join(t.TempDir(), "absent")
	t.Setenv(JobserverVariable, missing)
	jobserver = &compilerJobserver{}
	ran = false
	err := withCompilerToken(func() error { ran = true; return nil })
	if err == nil || ran || !strings.Contains(err.Error(), JobserverVariable+"="+missing) {
		t.Fatalf("a missing jobserver: ran %v, %v", ran, err)
	}
}

// The jobserver moves no key: a unit's and a runtime's keys are the same with it named and unset, since how many
// compile at once changes no byte. Mutant: the variable read into runtimeKeyParts.
func TestTheJobserverMovesNoKey(t *testing.T) {
	files := []runtimeFile{{"unit.c", []byte("int adamic_unit(void) { return 1; }\n")}}
	flags := []string{"-O2", "-std=c11"}
	keys := func() (string, string) {
		return unitKey(files, flags, "/tools/bin/clang", "clang version 20.1.8"), runtimeKey(files, flags, "/tools/bin/clang", "clang version 20.1.8", "headers")
	}
	t.Setenv(JobserverVariable, "")
	unit, runtimeKeyUnset := keys()
	t.Setenv(JobserverVariable, "/run/user/1000/loom-cc-jobserver")
	if unitSet, runtimeKeySet := keys(); unitSet != unit || runtimeKeySet != runtimeKeyUnset {
		t.Fatalf("a named jobserver moved the keys: unit %s to %s, runtime %s to %s", unit[:12], unitSet[:12], runtimeKeyUnset[:12], runtimeKeySet[:12])
	}
}
