//go:build !windows

package main

import (
	"os"
	"syscall"
)

// tryLockBuildFile attempts to acquire an exclusive non-blocking lock on lockPath.
// Returns an unlock func and true if the lock was acquired (or skipped due to an
// open error), or nil and false if another process already holds the lock.
func tryLockBuildFile(lockPath string) (func(), bool) {
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return func() {}, true
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = lockFile.Close()
		return nil, false
	}
	return func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		_ = lockFile.Close()
	}, true
}
