// Commands handed to childguard. This package sits under adamic's path so it may import internal/childguard.
package childguardcalls

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	elsewhere "lookalike/childguard"
)

// childguard.Run wraps every output in a writer of its own, so exec copies through pipes whatever the outputs are.
func handedToRun() {
	command := exec.Command("true")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	command.Stdout = os.Stdout
	_ = childguard.Run(command, childguard.Options{})
}

// childguard.CombinedOutput gives the command a bytes.Buffer: the stage 1 oracles' shape.
func handedToCombinedOutput(ctx context.Context) ([]byte, error) {
	command := exec.CommandContext(ctx, "true")
	command.WaitDelay = time.Second // want `exec\.Cmd\.WaitDelay with buffered output`
	return childguard.CombinedOutput(command, childguard.Options{Unbounded: true})
}

func literalHandedToRun() {
	_ = childguard.Run(&exec.Cmd{Path: "/bin/true", WaitDelay: time.Second}, childguard.Options{}) // want `exec\.Cmd\.WaitDelay with buffered output`
}

// childguard.Capture gives the command *os.File pipes, so exec copies nothing and WaitDelay only bounds the wait.
func handedToCapture(ctx context.Context) ([]byte, []byte, error) {
	command := exec.CommandContext(ctx, "true")
	command.WaitDelay = time.Second
	return childguard.Capture(ctx, command)
}

// The rule knows childguard by its package path, not by the name a call is spelled with.
func handedToALookalike() {
	command := exec.Command("true")
	command.WaitDelay = time.Second
	command.Stdout = os.Stdout
	_ = elsewhere.Run(command, elsewhere.Options{})
	_, _ = elsewhere.CombinedOutput(command, elsewhere.Options{})
}

// A command without WaitDelay handed to childguard is fine.
func withoutWaitDelay() {
	_ = childguard.Run(exec.Command("true"), childguard.Options{})
}
