package main

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"os"
	"os/exec"
	"time"
)

func main() {
	switch os.Args[1] {
	case "M07":
		cmd := exec.Command("sh", "-c", "echo ready; sleep 0.1; echo done")
		out, err := childguard.CombinedOutput(cmd, childguard.Options{FirstOutput: time.Second, Ceiling: 2 * time.Second})
		fmt.Printf("output=%q error=%v\n", out, err)
	case "M10":
		var out bytes.Buffer
		cmd := exec.Command("sh", "-c", "printf ready")
		cmd.Stdout = &out
		err := childguard.Run(cmd, childguard.Options{})
		fmt.Printf("stdout_restored=%t output=%q error=%v\n", cmd.Stdout == &out, out.String(), err)
	case "M18":
		cmd := exec.Command("sh", "-c", "echo ready; sleep 1")
		err := childguard.Run(cmd, childguard.Options{Stall: 50 * time.Millisecond, Ceiling: time.Second})
		guard, ok := err.(*childguard.Error)
		if ok {
			fmt.Printf("reason=%s load=%q\n", guard.Reason, guard.Load)
		} else {
			fmt.Printf("error=%v\n", err)
		}
	}
}
