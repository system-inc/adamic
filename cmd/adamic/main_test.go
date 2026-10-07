package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildAcceptsRootsAndProject(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs("../../internal/lower/testdata/multi_root")
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range []bool{false, true} {
		arguments := []string{"build"}
		if project {
			arguments = append(arguments, "--project", filepath.Join(directory, "tsconfig.json"))
		} else {
			for _, name := range []string{"third.a", "first.a", "second.a"} {
				arguments = append(arguments, filepath.Join(directory, name))
			}
		}
		binary := filepath.Join(t.TempDir(), "program")
		arguments = append(arguments, "-o", binary, "--sanitize")
		if code := run(arguments); code != 0 {
			t.Fatalf("%v: exit %d", arguments, code)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		output, err := exec.CommandContext(ctx, binary).CombinedOutput()
		cancel()
		if err != nil || string(output) != "leaf\nthird leaf\nshared leaf\nfirst shared\nsecond shared\n" {
			t.Fatalf("%v: %v, %q", arguments, err, output)
		}
	}
}
