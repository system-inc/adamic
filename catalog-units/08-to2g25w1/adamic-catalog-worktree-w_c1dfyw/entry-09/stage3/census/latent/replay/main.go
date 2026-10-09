// Replay builds the guarded measurement worker with the census overlay.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	started := time.Now()
	code := run()
	fmt.Fprintf(os.Stderr, "replay total (including overlay and Go build): %s\n", time.Since(started))
	os.Exit(code)
}

func run() int {
	repository, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	scratch, err := os.MkdirTemp("", "latent-replay-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(scratch)
	prepare := exec.Command("python3", filepath.Join(repository, "stage3/census/latent/make_overlay.py"), repository, scratch)
	prepare.Stderr = os.Stderr
	if err := prepare.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	arguments := []string{"run", "-buildvcs=false", "-overlay=" + filepath.Join(scratch, "overlay.json"), "./stage3/census/latent/replay/worker"}
	worker := exec.Command("go", append(arguments, os.Args[1:]...)...)
	worker.Stdout, worker.Stderr = os.Stdout, os.Stderr
	if err := worker.Run(); err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			return failure.ExitCode()
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
