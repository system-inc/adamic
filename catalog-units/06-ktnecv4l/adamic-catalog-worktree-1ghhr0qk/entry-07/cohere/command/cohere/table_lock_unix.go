//go:build !windows

package main

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// holdTableLock takes the table's lock exclusively, waiting for any other writer, and returns its release.
// It returns a release that does nothing when the lock cannot be taken, since writing without it is what
// every run did before it existed.
//
// flock belongs to the open file, so the kernel drops it when the holder exits however it exits, and a
// killed writer can never leave the next run waiting on a lock nobody holds.
func holdTableLock(directory string) func() {
	file, err := os.OpenFile(tableLockPath(directory), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return func() {}
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return func() {}
	}
	return func() {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
	}
}

// holdTableReadLock takes the table's lock shared, waiting up to bound for a writer to finish, and returns
// its release and whether it was taken. Past bound it gives up and returns a release that does nothing, and
// the run reads without it, where the worst outcome is a miss, never a stale hit.
//
// The lock file is created here as a writer would, so a reader racing the very first write in a directory
// holds the same lock that write is about to take, rather than finding no file and reading around it. With no
// directory at all there is nothing to read, and no lock is taken.
func holdTableReadLock(directory string, bound time.Duration) (func(), bool) {
	file, err := os.OpenFile(tableLockPath(directory), os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return func() {}, true
	}
	deadline := time.Now().Add(bound)
	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				file.Close()
			}, true
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) || time.Now().After(deadline) {
			file.Close()
			return func() {}, false
		}
		time.Sleep(2 * time.Millisecond)
	}
}
