//go:build !linux && !darwin

package native

// lockRuntime takes no lock off Linux and macOS (a wasm32-wasi host, whose syscall has no flock): runtimeBuilds' mutex
// alone serializes a key's builders, within one process, as it did before runtime_lock.go's lock, and processes asking
// for one key at once may each build it, the first rename published.
func lockRuntime(directory string) func() {
	return func() {}
}
