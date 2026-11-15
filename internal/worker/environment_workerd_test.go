package worker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func workerdEnvironmentProduct(t *testing.T, build bool) string {
	t.Helper()
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Name = "workerd string environment and independent twin"
	inputs.Files = append(inputs.Files, "internal/worker/environment_workerd_test.go", "workers/environment/handler.a", "workers/environment/sync-handler.a", "workers/environment/twin/worker.ts")
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
		for _, variant := range []struct{ entry, out string }{{"handler.a", "worker"}, {"sync-handler.a", "sync"}} {
			command := exec.CommandContext(t.Context(), compiler, "worker", "workers/environment/"+variant.entry, "--out", filepath.Join(directory, variant.out))
			command.Dir = root
			if report, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("environment Worker build: %w\n%s", err, report)
			}
		}
		twin := filepath.Join(directory, "twin")
		if err := os.Mkdir(twin, 0755); err != nil {
			return err
		}
		command := exec.CommandContext(t.Context(), "node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; process.stdout.write(stripTypeScriptTypes(readFileSync(process.argv[1],'utf8')));`, filepath.Join(root, "workers/environment/twin/worker.ts"))
		source, err := command.Output()
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(twin, "worker.mjs"), source, 0644)
	})
}
func TestProduct_WorkerdEnvironment(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	workerdEnvironmentProduct(t, true)
}
func TestWorkerdEnvironmentHandler(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	product := workerdEnvironmentProduct(t, false)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	const bindings = `(name="GREETING", text="hello"),(name="EXTRA",text="SECRET_EXTRA_1fa9"),(name="ZONE",text="world")`
	for _, kind := range []string{"worker", "sync", "twin"} {
		server, err := startWorkerdWithBindings(t, filepath.Join(product, kind), "127.0.0.1:0", bindings)
		if err != nil {
			t.Fatal(err)
		}
		command := exec.CommandContext(t.Context(), "node", filepath.Join(root, "workers/replay.mjs"), "--url", "http://"+server.address, filepath.Join(root, "workers/environment/corpus/requests.jsonl"), "--compare", filepath.Join(root, "workers/environment/corpus/responses.jsonl"))
		if report, err := command.CombinedOutput(); err != nil {
			t.Fatalf("environment %s corpus: %v\n%s", kind, err, report)
		}
	}
}
func TestWorkerdEnvironmentPrivacy(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	product := workerdEnvironmentProduct(t, false)
	const secret = "SECRET_BINDING_8ad3f1"
	const message = `environment binding "GREETING" is missing or not a string`
	for _, kind := range []string{"worker", "sync", "twin"} {
		for _, bad := range []string{"", `,(name="GREETING",json="827451")`} {
			bindings := `(name="ZONE",text="` + secret + `"),(name="EXTRA",text="` + secret + `_other")` + bad
			server, err := startWorkerdWithBindings(t, filepath.Join(product, kind), "127.0.0.1:0", bindings)
			if err != nil {
				t.Fatal(err)
			}
			response, err := http.Get("http://" + server.address + "/private")
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != 500 || string(body) != "internal error" {
				t.Fatalf("environment privacy response: status %d body %q", response.StatusCode, body)
			}
			var log []byte
			deadline := time.Now().Add(time.Second)
			for {
				log, err = os.ReadFile(server.logPath)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(log), message) || time.Now().After(deadline) {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if strings.TrimSpace(string(log)) != message || strings.Contains(string(log), secret) || strings.Contains(string(body), secret) || strings.Contains(string(log), "827451") || strings.Contains(string(body), "827451") {
				t.Fatalf("environment privacy log/response: %q / %q", log, body)
			}
		}
	}
}
