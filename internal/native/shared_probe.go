package native

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildSharedProbe compiles a test's host-side injector without linking the
// Adamic runtime. Its TLS and heap are the target process's, never a second
// runtime in an LD_PRELOAD object. Compiler work uses the normal jobserver.
func BuildSharedProbe(source, output string) error {
	directory, err := os.MkdirTemp("", "adamic-shared-probe-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	if err := os.WriteFile(filepath.Join(directory, "probe.c"), []byte(source), 0600); err != nil {
		return err
	}
	server, err := defaultJobserver()
	if err != nil {
		return err
	}
	flags := append(Flags(Options{}), "-fPIC", "-shared", "probe.c", "-ldl", "-o", output)
	command := clangIn(directory, compilerName(Options{}), flags...)
	var result []byte
	if err := server.Hold(func() error { var err error; result, err = command.CombinedOutput(); return err }); err != nil {
		return fmt.Errorf("native: shared probe: %w\n%s", err, result)
	}
	return nil
}
