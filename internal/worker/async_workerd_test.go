package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func workerdAsyncProduct(t *testing.T, build bool) string {
	t.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = "workerd async handler and independent twin"
	inputs.Files = append(inputs.Files, "internal/worker/async_workerd_test.go", "workers/async/handler.a", "workers/async/twin/worker.ts")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("node", "--version"))
	if !build {
		return findWorkerProduct(t, inputs)
	}
	compiler := workerCompilerProduct(t, nil, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		command := exec.CommandContext(t.Context(), compiler, "worker", "workers/async/handler.a", "--out", filepath.Join(directory, "worker"))
		command.Dir = root
		if report, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("async Worker build: %w\n%s", err, report)
		}
		twin := filepath.Join(directory, "twin")
		if err := os.Mkdir(twin, 0755); err != nil {
			return err
		}
		command = exec.CommandContext(t.Context(), "node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; process.stdout.write(stripTypeScriptTypes(readFileSync(process.argv[1],'utf8')));`, filepath.Join(root, "workers/async/twin/worker.ts"))
		source, err := command.Output()
		if err != nil {
			return fmt.Errorf("async twin build: %w", err)
		}
		return os.WriteFile(filepath.Join(twin, "worker.mjs"), source, 0644)
	})
}
func TestProduct_WorkerdAsync(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	workerdAsyncProduct(t, true)
}
func TestWorkerdAsyncHandler(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	product := workerdAsyncProduct(t, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"worker", "twin"} {
		server, err := startWorkerd(t, filepath.Join(product, name), "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		command := exec.CommandContext(t.Context(), "node", filepath.Join(root, "workers/replay.mjs"), "--url", "http://"+server.address, filepath.Join(root, "workers/async/corpus/requests.jsonl"), "--compare", filepath.Join(root, "workers/async/corpus/responses.jsonl"))
		if report, err := command.CombinedOutput(); err != nil {
			t.Fatalf("async %s corpus: %v\n%s", name, err, report)
		}
		deadline := time.Now().Add(time.Second)
		for {
			log, err := os.ReadFile(server.logPath)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(log), "after await") && strings.Contains(string(log), "helper panic") && strings.Contains(string(log), "helper rejected") && strings.Contains(string(log), "helper threw") {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("async %s rejection log missing: %s", name, log)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
