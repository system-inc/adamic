package native

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The service oracle's claim is 100,000 generated requests through one instance, every response equal to Node's,
// linear memory flat and, counted, nothing live after each request. It runs as four legs, each building what it needs,
// so no leg can pass on another's setup: the plain and counted wasm runs, the sanitized native command's agreement
// (opt-in, too long for the gate), and the mutants, one of which shrinks the request count and must be refused.
const serviceRequests = 100000

// serviceSetup is what every leg needs: the repository, the shared stage 0 compiler product, the service source and
// host, a fresh directory, and whether the generated C ends statement regions.
type serviceSetup struct {
	repository, compiler, source, host, directory string
	contents                                      []byte
	regions                                       string
}

func newServiceSetup(t *testing.T) serviceSetup {
	t.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1, source cloud/setup.sh --wasi-sdk's environment, and use Node 24")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Fatal("WASI_SYSROOT is required once the service oracle is opted in")
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	setup := serviceSetup{
		repository: repository,
		compiler:   buildcache.Adamic(t),
		source:     filepath.Join(repository, "internal/native/wasm/service/service.a"),
		host:       filepath.Join(repository, "internal/native/wasm/service/host.mjs"),
		directory:  t.TempDir(),
		regions:    "0",
	}
	if setup.contents, err = os.ReadFile(setup.source); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(wasiCommand(t, repository, setup.compiler, "c", setup.source)), "adamic_region_end(") {
		setup.regions = "1"
	}
	return setup
}

// runHost runs host over module ("-" generates the requests and Node's responses only) and returns its output.
func (setup serviceSetup) runHost(host, module, counted string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", host, module, setup.source,
		filepath.Join(setup.directory, "requests.jsonl"), filepath.Join(setup.directory, "responses.jsonl"), counted, setup.regions)
	command.Dir = setup.repository
	return command.CombinedOutput()
}

// serviceReport reads the host's closing line; a run that served fewer than all the requests is no evidence.
func serviceReport(output []byte, counted bool) error {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var report map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &report); err != nil {
		return fmt.Errorf("service report: %v\n%s", err, output)
	}
	if requests, _ := report["requests"].(float64); requests != serviceRequests {
		return fmt.Errorf("service served %v requests, want %d\n%s", report["requests"], serviceRequests, output)
	}
	if counted && report["live"] != float64(0) {
		return fmt.Errorf("service left %v values live, want 0\n%s", report["live"], output)
	}
	return nil
}

func (setup serviceSetup) wasmLeg(t *testing.T, counted string) {
	t.Helper()
	module := filepath.Join(setup.directory, "service-"+counted+".wasm")
	arguments := []string{"build", "--target", "wasm32-wasi", setup.source, "-o", module}
	if counted == "1" {
		arguments = append(arguments, "--count")
	}
	wasiCommand(t, setup.repository, setup.compiler, arguments...)
	output, err := setup.runHost(setup.host, module, counted)
	if err != nil {
		t.Fatalf("service counted=%s: %v\n%s", counted, err, output)
	}
	if err := serviceReport(output, counted == "1"); err != nil {
		t.Fatal(err)
	}
	t.Logf("service counted=%s: %s", counted, output)
}

// Not parallel: the service oracle measures request rates, and a neighbor's load would skew them.
func TestWASIService(t *testing.T) {
	newServiceSetup(t).wasmLeg(t, "0")
}

// Not parallel: the service oracle measures request rates, and a neighbor's load would skew them.
func TestWASIServiceCounted(t *testing.T) {
	newServiceSetup(t).wasmLeg(t, "1")
}

// The command imports the very same handler and answers the exact generated request file. A sanitized native run
// of all 100,000 requests takes about 58 s on an idle Linux box and more than twice that on a loaded gate box, past
// the gate's 60 s per-test budget even alone, so the ordinary gate skips it: the wasm legs hold the request count and
// live 0, and this leg runs on its own with ADAMIC_WASI_SERVICE_NATIVE=1.
func TestWASIServiceNativeAgrees(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_WASI_SERVICE_NATIVE") != "1" {
		t.Skip("set ADAMIC_WASI_SERVICE_NATIVE=1 to run the sanitized native agreement over all 100,000 requests")
	}
	setup := newServiceSetup(t)
	output, err := setup.runHost(setup.host, "-", "0")
	if err != nil {
		t.Fatalf("generate: %v\n%s", err, output)
	}
	if err := serviceReport(output, false); err != nil {
		t.Fatal(err)
	}
	commandSource := filepath.Join(setup.directory, "command.a")
	writeWASIFile(t, commandSource, "import { handleRequest } from "+strconv.Quote(filepath.ToSlash(setup.source))+";\n"+
		"import { programArguments, readTextFile } from 'adamic';\n"+
		"const input = readTextFile(programArguments()[0] ?? '');\n"+
		"if (input.kind === 'Ok') { for (const request of input.text.split('\\n').slice(0, -1)) { console.log(handleRequest(request)); } } else { console.error(input.message); }\n")
	binary := filepath.Join(setup.directory, "service-native")
	wasiCommand(t, setup.repository, setup.compiler, "build", commandSource, "-o", binary, "--sanitize")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	native, err := exec.CommandContext(ctx, binary, filepath.Join(setup.directory, "requests.jsonl")).Output()
	if err != nil {
		if failed, ok := err.(*exec.ExitError); ok {
			t.Fatalf("native service: %v\n%s", err, failed.Stderr)
		}
		t.Fatalf("native service: %v", err)
	}
	expected, err := os.ReadFile(filepath.Join(setup.directory, "responses.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(native) != string(expected) {
		actualLines := strings.Split(string(native), "\n")
		expectedLines := strings.Split(string(expected), "\n")
		for index, line := range expectedLines {
			if index >= len(actualLines) {
				t.Fatalf("native command ended before response %d", index)
			}
			if actualLines[index] != line {
				t.Fatalf("native response %d: got %q, want %q", index, actualLines[index], line)
			}
		}
		t.Fatalf("native command emitted %d lines, want %d", len(actualLines), len(expectedLines))
	}
	t.Logf("native sanitized command: all %d responses agree with Node", serviceRequests)
}

func TestWASIServiceMutants(t *testing.T) {
	t.Parallel()
	setup := newServiceSetup(t)
	mutants := []struct {
		name, before, after, caught string
	}{
		{"response-byte", `service":"orders`, `service":"ordert`, "wasm semantic pin"},
		{"retained-object", "const parser = new Parser(request);", "const parser = new Parser(request); retained.push(parser);", "retained a live allocation"},
		{"unicode-escape", "String.fromCharCode(code)", "String.fromCharCode(code + 1)", "wasm semantic pin"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Count(string(setup.contents), mutant.before) != 1 {
				t.Fatal("mutant must change exactly one source location")
			}
			changed := strings.Replace(string(setup.contents), mutant.before, mutant.after, 1)
			if mutant.name == "retained-object" {
				changed += "\nconst retained: Parser[] = [];\n"
			}
			mutantSource := filepath.Join(setup.directory, mutant.name+".a")
			writeWASIFile(t, mutantSource, changed)
			module := filepath.Join(setup.directory, mutant.name+".wasm")
			wasiCommand(t, setup.repository, setup.compiler, "build", "--target", "wasm32-wasi", mutantSource, "-o", module, "--count")
			output, err := setup.runHost(setup.host, module, "1")
			if err == nil || !strings.Contains(string(output), mutant.caught) {
				t.Fatalf("mutant not caught by %q: %v\n%s", mutant.caught, err, output)
			}
			t.Logf("mutant compiled and was caught by %q\n%s", mutant.caught, output)
		})
	}
	// A leg that served fewer requests than the claim must be refused, not passed: the host told to generate a
	// thousand stands in for any split that quietly shrinks the count.
	t.Run("shrunk-count", func(t *testing.T) {
		original, err := os.ReadFile(setup.host)
		if err != nil {
			t.Fatal(err)
		}
		shrunk := strings.Replace(string(original), "Array.from({length:100000}", "Array.from({length:1000}", 1)
		if shrunk == string(original) {
			t.Fatal("shrunk-count mutant did not change the host")
		}
		host := filepath.Join(setup.directory, "host-shrunk.mjs")
		writeWASIFile(t, host, shrunk)
		output, err := setup.runHost(host, "-", "0")
		if err != nil {
			t.Fatalf("shrunk host: %v\n%s", err, output)
		}
		if serviceReport(output, false) == nil {
			t.Fatalf("a run of fewer than %d requests passed the report check\n%s", serviceRequests, output)
		}
		t.Log("shrunk-count mutant refused by the report check")
	})
}
