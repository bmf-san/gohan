//go:build windows

package main

// tryLockBuildFile is a no-op on Windows; always reports the lock as acquired.
func tryLockBuildFile(_ string) (func(), bool) {
	return func() {}, true
}
