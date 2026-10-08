// Run a gathered source slice through the shared leak check without changing
// compiler options or the production runtime.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: check-slice.go ENTRY [ARGUMENT ...]")
	}
	source, err := load.Load([]string{os.Args[1]})
	must(err)
	program, err := lower.Lower(context.Background(), source)
	must(err)
	code := native.C(program)
	dir, err := os.MkdirTemp("", "scout-33-slice-")
	must(err)
	defer os.RemoveAll(dir)
	binary, counted := filepath.Join(dir, "sanitized"), filepath.Join(dir, "counted")
	must(native.Build(code, binary, native.Options{Sanitize: true}))
	execute := func(environment []string, name string, arguments ...string) leakcheck.Run {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, name, arguments...)
		command.Env = append(os.Environ(), environment...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if _, ok := err.(*exec.ExitError); !ok {
			must(err)
		}
		return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
	}
	report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: counted,
		Arguments: func() []string { return os.Args[2:] }, Execute: execute})
	must(err)
	if report != "" {
		panic(report)
	}
	fmt.Println("shared leak check: PASS")
	must(native.Build(code, counted, native.Options{Count: true}))
	// Exercise the counted predicate on Linux as well as macOS.
	command := exec.Command(counted, os.Args[2:]...)
	output, err := command.CombinedOutput()
	must(err)
	if report := leakcheck.Unbalanced(leakcheck.Run{Stderr: output}); report != "" {
		panic(report)
	}
	fmt.Printf("counted balance: PASS\n%s", output)
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
