package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const tableLayoutSetupManifestEnv = "ADAMIC_TABLE_LAYOUT_SETUP_MANIFEST"
const tableLayoutSetupDeadlineEnv = "ADAMIC_TABLE_LAYOUT_SETUP_DEADLINE"

type tableLayoutManifest struct {
	Main, List, Document, Sanitized, Release, Backend string
	Inputs                                            []auditInput
	Corpus                                            []bool
	Elapsed                                           time.Duration
}

// Preparation runs before TestMain and m.Run, including for an isolated gate
// selector. The child owns the discoverable setup test; no shard builds or waits
// on a lazy shared-state lock. Listing tests must never fetch build products.
func init() {
	if os.Getenv(tableLayoutSetupManifestEnv) != "" {
		return
	}
	pattern := "."
	for index, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-test.list") {
			return
		}
		if strings.HasPrefix(arg, "-test.run=") {
			pattern = strings.TrimPrefix(arg, "-test.run=")
		}
		if arg == "-test.run" && index+2 < len(os.Args) {
			pattern = os.Args[index+2]
		}
	}
	pattern = strings.Split(pattern, "/")[0]
	selector, err := regexp.Compile(pattern)
	if err != nil {
		return
	} // testing reports the invalid selector itself.
	selected := selector.MatchString("TestMarkdownTableLayout")
	for shard := 0; shard < testMarkdownTableLayoutShards; shard++ {
		selected = selected || selector.MatchString(fmt.Sprintf("TestMarkdownTableLayout_%03d", shard))
	}
	if !selected {
		return
	}
	if err := tableLayoutPrepare(); err != nil {
		panic(err)
	}
}

func tableLayoutPrepare() error {
	manifest, err := os.CreateTemp("", "adamic-table-setup-*.json")
	if err != nil {
		return err
	}
	path := manifest.Name()
	if err := manifest.Close(); err != nil {
		os.Remove(path)
		return err
	}
	defer os.Remove(path)
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := tableLayoutCommand(ctx, binary, "-test.run=^TestMarkdownTableLayout_Setup$", "-test.timeout=0", "-test.v")
	deadline, _ := ctx.Deadline()
	command.Env = append(os.Environ(), tableLayoutSetupManifestEnv+"="+path, tableLayoutSetupDeadlineEnv+"="+deadline.Format(time.RFC3339Nano))
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("cooked: TestMarkdownTableLayout_Setup exceeded 90s: %w\n%s", ctx.Err(), output)
	}
	if err != nil {
		return fmt.Errorf("table shared setup: %w\n%s", err, output)
	}
	encoded, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data tableLayoutManifest
	if err := json.Unmarshal(encoded, &data); err != nil {
		return err
	}
	if len(data.Inputs) == 0 || len(data.Corpus) != len(data.Inputs) {
		return fmt.Errorf("table setup manifest lost its corpus")
	}
	for index := range data.Inputs {
		data.Inputs[index].Corpus = data.Corpus[index]
	}
	tableLayoutSharedProducts = tableLayoutProducts{main: data.Main, list: data.List, document: data.Document, sanitized: data.Sanitized, release: data.Release, backend: data.Backend}
	tableLayoutInputs = data.Inputs
	tableLayoutSetupElapsed = data.Elapsed
	return nil
}

func tableLayoutCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func tableLayoutExecute(ctx context.Context, environment []string, name string, args ...string) (run, error) {
	command := tableLayoutCommand(ctx, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		return run{}, fmt.Errorf("cooked: table layout case deadline exceeded: %w", ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, fmt.Errorf("running %s: %w", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
}

func tableLayoutNode(ctx context.Context, path string, args ...string) (run, error) {
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		return run{}, err
	}
	return tableLayoutExecute(ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...)
}
