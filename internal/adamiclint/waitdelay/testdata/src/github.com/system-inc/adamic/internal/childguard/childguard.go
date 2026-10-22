// The package that holds the cure is never flagged, even for the shape the rule flags everywhere else. Run,
// CombinedOutput and Capture stand in for the real ones, with their signatures, for the fixtures that call them.
package childguard

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type Options struct {
	Unbounded bool
}

func Run(cmd *exec.Cmd, options Options) error { return cmd.Run() }

func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) { return cmd.CombinedOutput() }

func Capture(ctx context.Context, cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	return nil, nil, cmd.Run()
}

func buffered() []byte {
	var output bytes.Buffer
	command := exec.Command("sh")
	command.WaitDelay = time.Second
	command.Stdout = &output
	_ = command.Run()
	return output.Bytes()
}
