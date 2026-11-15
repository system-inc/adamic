package adaptedtree

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The product retains apply.py's incremental refs, before any later adaptation
// can change sys.ts. This consumer reads those refs and never starts a Go process.
func TestBuiltinImportsScopeFollowsOwnerPins(t *testing.T) {
	t.Parallel()
	product := Product(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	root, err := filepath.Abs(repositoryFromLane)
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	for _, side := range []string{"before", "after"} {
		command := toolchain.CommandContextIn(ctx, filepath.Join(product, "tree"), "git",
			"archive", "refs/stage3/93-builtin-imports-"+side, "src")
		command.Env = append(command.Env, toolchain.GitEnv()...)
		data, err := command.Output()
		if err != nil {
			t.Fatal(err)
		}
		reader := tar.NewReader(bytes.NewReader(data))
		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if header.Typeflag != tar.TypeReg {
				continue
			}
			if !filepath.IsLocal(header.Name) {
				t.Fatalf("non-local snapshot path: %s", header.Name)
			}
			name := filepath.Join(scratch, side, header.Name)
			if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	proofs := filepath.Join(root, "stage3/adapt/93-builtin-imports/proofs")
	run := func(script, mode string, want string) {
		t.Helper()
		args := []string{filepath.Join(proofs, script), filepath.Join(scratch, "before"), filepath.Join(scratch, "after")}
		if mode != "" {
			args = append(args, mode)
		}
		command := toolchain.CommandContextIn(ctx, "", "node", args...)
		command.Env = append(command.Env, "NODE_PATH="+filepath.Join(root, "stage3/api/node_modules"))
		output, err := command.CombinedOutput()
		if want == "" {
			if err != nil {
				t.Fatalf("%s: %v\n%s", script, err, output)
			}
		} else {
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), want) {
				t.Fatalf("%s %s: want exit 1 naming %q, got %v\n%s", script, mode, want, err, output)
			}
		}
		t.Logf("%s %s: %s", script, mode, output)
	}
	run("check.cjs", "", "")
	run("scope-mutant.cjs", "", "scope-outside-mutant")
	run("scope-mutant.cjs", "inside", "pinned After owner differs: getNodeSystem")
	run("check.cjs", "", "")
}
