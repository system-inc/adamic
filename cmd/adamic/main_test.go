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
			arguments = append(arguments, "--project", filepath.Join(directory, "tsconfig.json"), "--entry", filepath.Join(directory, "third.a"))
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
		want := "leaf\nthird leaf\nshared leaf\nfirst shared\nsecond shared\n"
		if project {
			want = "leaf\nthird leaf\n"
		}
		if err != nil || string(output) != want {
			t.Fatalf("%v: %v, %q", arguments, err, output)
		}
	}
}

func TestProjectBuildRequiresExplicitEntry(t *testing.T) {
	t.Parallel()
	if code := run([]string{"build", "--project", "../../internal/lower/testdata/project_entry/tsconfig.json", "-o", filepath.Join(t.TempDir(), "program")}); code != 2 {
		t.Fatalf("project build without --entry: exit %d, want usage refusal", code)
	}
}
