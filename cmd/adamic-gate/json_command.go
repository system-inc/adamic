package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/boundedrun"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func stderrPath(path string) string { return strings.TrimSuffix(path, filepath.Ext(path)) + ".stderr" }

// Keep toolchain diagnostics out of the strict JSON evidence stream.
func runJSONCommand(cmd *boundedrun.Cmd, path string) error {
	stdout, err := os.Create(path)
	if err != nil {
		return err
	}
	stderr, err := os.Create(stderrPath(path))
	if err != nil {
		return errors.Join(err, stdout.Close())
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	runErr := cmd.Run()
	closeErr := errors.Join(stdout.Close(), stderr.Close())
	if runErr != nil {
		diagnostic, _ := os.ReadFile(stderrPath(path))
		runErr = fmt.Errorf("%s: %w; stderr in %s\n%s", path, runErr, stderrPath(path), diagnostic)
	}
	return errors.Join(runErr, closeErr)
}

// A bounded wait can return while a pipe copier finishes. Snapshot stderr safely.
type commandBuffer struct {
	mutex sync.Mutex
	bytes.Buffer
}

func (b *commandBuffer) Write(data []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.Buffer.Write(data)
}
func (b *commandBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.Buffer.String()
}
