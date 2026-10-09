package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func TestParseIntMapUsesIndexRadix(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "console.log(['10','10','10'].map(Number.parseInt).join(','));")
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name != "library_map_parseInt" {
			continue
		}
		if len(function.Parameters) != 3 || len(function.Body) != 1 {
			t.Fatalf("unexpected callback: %#v", function)
		}
		ret, ok := function.Body[0].(ir.Return)
		if !ok {
			t.Fatalf("callback body: %T", function.Body[0])
		}
		call, ok := ret.Value.(ir.NumberCall)
		if !ok || call.Function != "parseInt" || len(call.Arguments) != 2 {
			t.Fatalf("parseInt call: %#v", ret.Value)
		}
		want := []ir.Expression{ir.Read{Local: function.Parameters[0], Of: ir.String}, ir.Read{Local: function.Parameters[1], Of: ir.Number}}
		if !reflect.DeepEqual(call.Arguments, want) {
			t.Fatalf("parseInt must receive element then index radix: got %#v, want %#v", call.Arguments, want)
		}
		return
	}
	t.Fatal("missing parseInt map callback")
}

// Each execution owns its working directory so source and lowered JavaScript
// see the same empty filesystem, without sharing state with parallel tests.
func fsLoweringAgreesWithNode(t *testing.T, source, wantStdout string, wantExit int) {
	t.Helper()
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(path string) (string, int) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", runner, path)
		command.Dir = t.TempDir()
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		exit := 0
		if err != nil {
			var failure *exec.ExitError
			if !errors.As(err, &failure) || ctx.Err() != nil {
				t.Fatalf("Node failed to run: %v; stderr %s", err, stderr.Bytes())
			}
			exit = failure.ExitCode()
		}
		t.Logf("Node %s: exit %d, stdout %q", filepath.Base(path), exit, stdout)
		return string(stdout), exit
	}
	oracle, oracleExit := run(path)
	if oracle != wantStdout || oracleExit != wantExit {
		t.Fatalf("source Node: exit %d stdout %q, want exit %d stdout %q", oracleExit, oracle, wantExit, wantStdout)
	}
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	got, exit := run(generated)
	if got != oracle || exit != oracleExit {
		t.Fatalf("JavaScript backend: exit %d stdout %q; source Node: exit %d stdout %q", exit, got, oracleExit, oracle)
	}
}

func TestFSOpenStringFlagsLower(t *testing.T) {
	t.Parallel()
	fsLoweringAgreesWithNode(t, `import {writeFileSync,openSync,closeSync,readFileSync,unlinkSync} from 'node:fs';
 writeFileSync('x','opened bytes'); const fd=openSync('x','r'); closeSync(fd);
 console.log(readFileSync('x','utf8')); unlinkSync('x');`, "opened bytes\n", 0)
}

func TestFSRemoveDefaultRetryDelayLowers(t *testing.T) {
	t.Parallel()
	fsLoweringAgreesWithNode(t, `import {writeFileSync,existsSync,rmSync} from 'node:fs';
 writeFileSync('x','remove me'); console.log(existsSync('x')?'present':'missing');
 rmSync('x',{retryDelay:100}); console.log(existsSync('x')?'present':'missing');`, "present\nmissing\n", 0)
}

func TestFSExistsOperation(t *testing.T) {
	t.Parallel()
	fsLoweringAgreesWithNode(t, `import {existsSync,writeFileSync,unlinkSync} from 'node:fs';
 console.log(existsSync('x')?'present':'missing'); writeFileSync('x','exists');
 console.log(existsSync('x')?'present':'missing'); unlinkSync('x');
 console.log(existsSync('x')?'present':'missing');`, "missing\npresent\nmissing\n", 0)
}

func TestFSStatThrowsByDefault(t *testing.T) {
	t.Parallel()
	// Missing-file behavior distinguishes throwIfNoEntry's default: Node exits
	// before the final print; returning undefined instead continues and exits zero.
	fsLoweringAgreesWithNode(t, `import {existsSync,statSync} from 'node:fs';
 console.log(existsSync('missing')?'present':'missing'); statSync('missing');
 console.log('stat returned instead of throwing');`, "missing\n", 1)
}
