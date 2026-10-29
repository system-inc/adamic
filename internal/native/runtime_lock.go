//go:build linux || darwin

package native

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// The lock is a file of its own for the platforms whose syscall has flock; elsewhere runtime_lock_other.go's lockRuntime
// takes none. Only cachedRuntime reaches it, never an archive's build (RuntimeRequest.Build), so it is no file
// runtimeproducts' runtimeFiles declares.

// runtimeLockWait is the longest a process waits on another building the runtime archive it wants (lockRuntime). A
// build takes up to about a minute on a loaded tree builder, and a waiter waits only on the one build holding the lock,
// so a holder past this is stuck, not slow: the waiter builds as it would have without the lock, since a lock may slow a
// build but must never hang one.
const runtimeLockWait = 5 * time.Minute

// runtimeLockPoll is how often a waiter tries the lock again. Each try is one system call, and a waiter finds its
// archive within one of these of its builder's rename.
const runtimeLockPoll = 50 * time.Millisecond

// lockRuntime takes the lock beside the runtime archive's directory, directory+".lock", which a process building it
// holds from its second look to its rename, so of the processes that miss one key together one compiles it and the rest
// find what it published (Loom's tree builders, Oct 28: hundreds of product tests on one cache built one key 22 times).
// It is beside the directory, never in it, since the directory is published whole by rename, and it reaches neither the
// archive's bytes nor its key. flock's lock goes with its holder, so a build that crashed never leaves it held. It
// returns the release. A lock it can't open or take, or one another process still holds after runtimeLockWait, it says
// so on stderr, never in the build log, and returns a release that does nothing: the caller builds as it would have
// without the lock.
func lockRuntime(directory string) func() {
	key := filepath.Base(directory)
	lock, err := os.OpenFile(directory+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "native: building runtime %s without its lock: %v\n", key, err)
		return func() {}
	}
	// A try that would block returns at once, so the wait is bounded by the deadline rather than by the holder.
	deadline := time.Now().Add(runtimeLockWait)
	for {
		err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
				lock.Close()
			}
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EINTR) {
			lock.Close()
			fmt.Fprintf(os.Stderr, "native: building runtime %s without its lock: %v\n", key, err)
			return func() {}
		}
		if time.Now().After(deadline) {
			lock.Close()
			fmt.Fprintf(os.Stderr, "native: building runtime %s without its lock: another process held it past %s\n", key, runtimeLockWait)
			return func() {}
		}
		time.Sleep(runtimeLockPoll)
	}
}
