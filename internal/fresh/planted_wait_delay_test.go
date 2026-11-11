package fresh

import (
	"bytes"
	"os/exec"
	"testing"
	"time"
)

// A scratch plant for the adamiclint and cohere parity read (#57t6rwv): a buffered command with a WaitDelay, which
// both rules must flag. It never lands.
func TestPlantedWaitDelay(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	command := exec.Command("sh", "-c", "true")
	command.Stdout = &output
	command.WaitDelay = time.Second
	_ = command.Run()
}
