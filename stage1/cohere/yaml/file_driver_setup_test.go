package yaml

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"
)

// The gate selects each leaf in a fresh process. Run the setup test before m.Run
// so neither a shard's test clock nor its child deadline includes shared builds.
// The setup child also supports uncached proof runs: its products remain valid
// until the parent has finished every selected leaf.
func TestMain(m *testing.M) {
	flag.Parse()
	if os.Getenv("ADAMIC_FILE_DRIVER_SETUP_CHILD") == "1" {
		os.Exit(m.Run())
	}
	pattern, err := regexp.Compile(flag.Lookup("test.run").Value.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	selected := pattern.MatchString("TestFileDriver_Setup") || pattern.MatchString("TestFileDriverUnion")
	for shard := 0; shard < testFileDriverShards; shard++ {
		selected = selected || pattern.MatchString(fmt.Sprintf("TestFileDriver_%03d", shard))
	}
	// Listing discovers leaves without preparing their products.
	if !selected || flag.Lookup("test.list").Value.String() != "" {
		os.Exit(m.Run())
	}
	directory, err := os.MkdirTemp("", "yaml-file-driver-state-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	manifest := filepath.Join(directory, "state.json")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFileDriver_Setup$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_FILE_DRIVER_SETUP_CHILD=1", "ADAMIC_FILE_DRIVER_STATE="+manifest)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = time.Second
	err = command.Run()
	cancel()
	if err == nil {
		err = fileDriverReadState(manifest)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "file-driver setup:", err)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	status := m.Run()
	os.RemoveAll(directory)
	os.Exit(status)
}

type fileDriverManifest struct {
	Inputs                         []string
	Expected                       [][]byte
	Binary, Emitted, Runner, Entry string
}

func fileDriverPublish(t *testing.T, state *fileDriverState) {
	t.Helper()
	path := os.Getenv("ADAMIC_FILE_DRIVER_STATE")
	if path == "" {
		t.Fatal("setup must run through the TestMain preflight")
	}
	data, err := json.Marshal(fileDriverManifest{state.inputs, state.expected, state.binary, state.emitted, state.runner, state.entry})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func fileDriverReadState(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var manifest fileDriverManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	fileDriverShared = &fileDriverState{manifest.Inputs, manifest.Expected, manifest.Binary, manifest.Emitted, manifest.Runner, manifest.Entry}
	return nil
}
