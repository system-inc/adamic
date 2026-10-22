// Each function here sets WaitDelay on a command whose output is buffered, here or by a caller, and must be flagged.
package flagged

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

func fieldWithBuffer() {
	var output bytes.Buffer
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `^exec\.Cmd\.WaitDelay with buffered output reports ErrWaitDelay by chance when a slow run drains late; run it with childguard\.Capture, or say why in // WaitDelay: <reason> \(#n3gk22h\)$`
	command.Stdout = &output
	_ = command.Run()
}

func stderrToBuilderBeforeDelay() {
	var errors strings.Builder
	command := exec.Command("sh")
	command.Stderr = &errors
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	_ = command.Run()
}

func bothStreamsAtOnce() {
	var output bytes.Buffer
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	command.Stdout, command.Stderr = &output, &output
	_ = command.Run()
}

// An io.Writer may hold an *os.File, but the rule reads the type the code wrote, and exec copies anything else.
func writerOfUnknownKind(writer io.Writer) {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	command.Stdout = writer
	_ = command.Run()
}

func compositeLiteral() {
	var output bytes.Buffer
	_ = (&exec.Cmd{Path: "/bin/true", Stdout: &output, WaitDelay: time.Second}).Run() // want `exec\.Cmd\.WaitDelay with buffered output`
}

func compositeLiteralBufferedLater() {
	var output bytes.Buffer
	command := &exec.Cmd{Path: "/bin/true", WaitDelay: time.Second} // want `exec\.Cmd\.WaitDelay with buffered output`
	command.Stderr = &output
	_ = command.Run()
}

func compositeLiteralDeclared() {
	var command = exec.Cmd{Path: "/bin/true", WaitDelay: time.Second} // want `exec\.Cmd\.WaitDelay with buffered output`
	_, _ = command.Output()
}

func combinedOutput() {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	_, _ = command.CombinedOutput()
}

func output() {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	_, _ = command.Output()
}

// The markdownblocks shape (#n3gk22h): the helper sets WaitDelay and returns the command, and its caller buffers it.
func returnedHelper(ctx context.Context, name string) *exec.Cmd {
	command := exec.CommandContext(ctx, name)
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	return command
}

func callerBuffersHelper(ctx context.Context) {
	command := returnedHelper(ctx, "sh")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	_ = command.Run()
}

func returnedWithError() (*exec.Cmd, error) {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	return command, nil
}

func returnedLiteral() *exec.Cmd {
	return &exec.Cmd{Path: "/bin/true", WaitDelay: time.Second} // want `exec\.Cmd\.WaitDelay with buffered output`
}

func returnedPointerToValue() *exec.Cmd {
	var command exec.Cmd
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	return &command
}

type runner struct {
	*exec.Cmd
}

func embedded() {
	var output bytes.Buffer
	run := runner{exec.Command("sh")}
	run.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	run.Stdout = &output
	_ = run.Run()
}

type job struct {
	command *exec.Cmd
}

func throughAField(work *job) {
	var output bytes.Buffer
	work.command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	work.command.Stdout = &output
	_ = work.command.Run()
}

func bufferedInAClosure() {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	run := func() ([]byte, error) { return command.CombinedOutput() }
	_, _ = run()
}

func escapeWithoutAReason() {
	command := exec.Command("sh")
	// WaitDelay:
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	_, _ = command.Output()
}

func escapeTooFarAbove() {
	command := exec.Command("sh")
	// WaitDelay: this reason is for the line just below, not for the assignment after it.
	command.Env = nil
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	_, _ = command.Output()
}
