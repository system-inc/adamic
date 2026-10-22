// Nothing here may be flagged: each WaitDelay goes only to files or nowhere, carries its reason, or isn't exec's.
package allowed

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

func fileOutputs() {
	command := exec.Command("sh")
	command.WaitDelay = time.Second
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	_ = command.Run()
}

func createdFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	command := exec.Command("sh")
	command.WaitDelay = time.Second
	command.Stdout, command.Stderr = file, nil
	return command.Run()
}

func nilOutputs() {
	command := exec.Command("sh")
	command.WaitDelay = time.Second
	command.Stdout = nil
	_ = command.Run()
}

func fileLiteral() {
	_ = (&exec.Cmd{Path: "/bin/true", Stdout: os.Stdout, Stderr: nil, WaitDelay: time.Second}).Run()
}

func escapeAbove() {
	var output bytes.Buffer
	command := exec.Command("sh")
	// WaitDelay: the child never starts a descendant, and a late drain here is a hang we want reported.
	command.WaitDelay = time.Second
	command.Stdout = &output
	_ = command.Run()
}

func escapeWrapped() {
	command := exec.Command("sh")
	// WaitDelay: a reason long enough to wrap
	// onto a second line above the code it explains.
	command.WaitDelay = time.Second
	_, _ = command.CombinedOutput()
}

func escapeOnTheLine() *exec.Cmd {
	command := exec.Command("sh")
	command.WaitDelay = time.Second // WaitDelay: every caller hands it *os.File outputs.
	return command
}

func escapeOnALiteral() *exec.Cmd {
	return &exec.Cmd{
		Path: "/bin/true",
		// WaitDelay: the reason, on the field.
		WaitDelay: time.Second,
	}
}

func noWaitDelay(ctx context.Context) ([]byte, error) {
	var output bytes.Buffer
	command := exec.CommandContext(ctx, "sh")
	command.Stdout = &output
	if err := command.Run(); err != nil {
		return nil, err
	}
	return exec.Command("sh").CombinedOutput()
}

func helperWithoutWaitDelay() *exec.Cmd {
	return exec.Command("sh")
}

// Two commands: the one with WaitDelay writes to files, the buffered one has none.
func anotherCommandBuffered() {
	var output bytes.Buffer
	quiet := exec.Command("sh")
	quiet.WaitDelay = time.Second
	quiet.Stdout = os.Stdout
	loud := exec.Command("sh")
	loud.Stdout = &output
	_, _ = exec.Command("sh").Output()
	_ = quiet.Run()
	_ = loud.Run()
}

// A field named WaitDelay on some other type is not exec's, whatever its neighbors are called.
type lookalike struct {
	WaitDelay time.Duration
	Stdout    io.Writer
}

func (lookalike) Output() ([]byte, error) { return nil, nil }

func notExec() lookalike {
	var output bytes.Buffer
	fake := lookalike{WaitDelay: time.Second}
	fake.WaitDelay = time.Second
	fake.Stdout = &output
	_, _ = fake.Output()
	return fake
}
