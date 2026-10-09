//go:build unix

package main

import (
	"errors"
	"os"
	"syscall"
)

// filesLock says whether a lock on a file holds here: it does, through flock.
const filesLock = true

// tryLockFile takes an exclusive lock on the file without waiting, and reports whether it got it. The
// kernel drops the lock when the process exits, however it exits, so a slot is never held by a run that
// is gone.
func tryLockFile(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	return err == nil, err
}

func unlockFile(file *os.File) {
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
