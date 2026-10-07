package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNodeProcessCwdBarrier(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_cached_cwd.a")
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.py"))
	if err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(t.TempDir(), "plain.mjs")
	// This control uses Node's own synchronous file API, without the Adamic runtime.
	if err := os.WriteFile(plain, []byte(`import { readFileSync } from 'node:fs';
const first = process.cwd();
process.stdout.write('ready\n');
if (readFileSync(process.argv[2], 'utf8') !== 'continue') throw new Error('invalid barrier');
console.log('cached ' + (process.cwd() === first));
`), 0600); err != nil {
		t.Fatal(err)
	}
	observe := func(harness string, command ...string) run {
		result := execute(t, "python3", append([]string{harness}, command...)...)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Errorf("cwd barrier driver failed: %d %q", result.exitCode, result.stderr)
			return result
		}
		var output struct {
			Stdout, Stderr []byte
			ExitCode       int
		}
		if err := json.Unmarshal(result.stdout, &output); err != nil {
			t.Error(err)
			return result
		}
		return run{stdout: output.Stdout, stderr: output.Stderr, exitCode: output.ExitCode}
	}
	truth := observe(driver, "node", plain)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\ncached true\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected plain Node control: %+v", truth)
	}
	for _, backend := range []struct {
		name    string
		command []string
	}{
		{"plain Node", []string{"node", plain}},
		{"source Node", []string{"node", "--disable-warning=ExperimentalWarning", runner, path}},
		{"native", []string{binary}},
		{"JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", runner, script}},
	} {
		t.Run(backend.name, func(t *testing.T) {
			var done sync.WaitGroup
			for index := 0; index < 24; index++ {
				done.Add(1)
				go func() {
					defer done.Done()
					if difference := disagreement(truth, observe(driver, backend.command...)); difference != "" {
						t.Error(difference)
					}
				}()
			}
			done.Wait()
			t.Log("24 concurrent observations compared with independent plain Node")
		})
	}
	data, err := os.ReadFile(driver)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "base64.b64encode(b''.join(output))"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("driver truncation mutant anchor changed")
	}
	mutant := filepath.Join(t.TempDir(), "truncated-driver.py")
	if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), anchor, "base64.b64encode(output[0])", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, path}, {binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
		bad := observe(mutant, command...)
		if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
			t.Fatal("driver read-ahead-loss mutant did not fail only Node stdout comparison")
		}
	}
	t.Log("clean driver truncation mutant caught only by Node stdout on source and both backends")
}
