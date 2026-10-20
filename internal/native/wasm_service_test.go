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
// linear memory flat and, counted, nothing live after each request. It runs as legs, each building what it needs, so
// no leg can pass on another's setup: the plain and counted wasm runs, the sanitized native command's agreement in four
// quarters, and the mutants, one of which shrinks the request count and must be refused.
const serviceRequests = 100000

// serviceQuarters are the request ranges, [start, end), the sanitized native command answers, one test each. A
// sanitized run of all 100,000 at once doesn't fit the gate's 60 s per-test budget, so every request is still checked
// under ASan, but in four instances: the trade is that no sanitized native instance holds all 100,000, which the
// counted wasm leg's single-instance live 0 still proves.
var serviceQuarters = [][2]int{{0, 25000}, {25000, 50000}, {50000, 75000}, {75000, 100000}}

// tileRequests refuses ranges that don't cover [0, total) exactly once, in order: a gap leaves requests unchecked and
// an overlap checks some twice while the claim counts them once.
func tileRequests(ranges [][2]int, total int) error {
	next := 0
	for index, span := range ranges {
		if span[0] != next {
			return fmt.Errorf("range %d starts at %d, want %d: a gap or an overlap", index, span[0], next)
		}
		if span[1] <= span[0] {
			return fmt.Errorf("range %d, [%d, %d), is empty", index, span[0], span[1])
		}
		next = span[1]
	}
	if next != total {
		return fmt.Errorf("ranges end at %d, want %d", next, total)
	}
	return nil
}

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

// nativeQuarter has the host generate all the requests and Node's responses, then builds a sanitized native command
// importing the very same handler and has it answer one quarter's exact lines of that request file.
func nativeQuarter(t *testing.T, quarter int) {
	t.Helper()
	if err := tileRequests(serviceQuarters, serviceRequests); err != nil {
		t.Fatal(err)
	}
	setup := newServiceSetup(t)
	output, err := setup.runHost(setup.host, "-", "0")
	if err != nil {
		t.Fatalf("generate: %v\n%s", err, output)
	}
	if err := serviceReport(output, false); err != nil {
		t.Fatal(err)
	}
	requests, err := os.ReadFile(filepath.Join(setup.directory, "requests.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	responses, err := os.ReadFile(filepath.Join(setup.directory, "responses.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	requestLines := strings.Split(strings.TrimSuffix(string(requests), "\n"), "\n")
	responseLines := strings.Split(strings.TrimSuffix(string(responses), "\n"), "\n")
	if len(requestLines) != serviceRequests || len(responseLines) != serviceRequests {
		t.Fatalf("generated %d requests and %d responses, want %d of each", len(requestLines), len(responseLines), serviceRequests)
	}
	span := serviceQuarters[quarter]
	quarterRequests := filepath.Join(setup.directory, "quarter.jsonl")
	writeWASIFile(t, quarterRequests, strings.Join(requestLines[span[0]:span[1]], "\n")+"\n")
	commandSource := filepath.Join(setup.directory, "command.a")
	writeWASIFile(t, commandSource, "import { handleRequest } from "+strconv.Quote(filepath.ToSlash(setup.source))+";\n"+
		"import { programArguments, readTextFile } from 'adamic';\n"+
		"const input = readTextFile(programArguments()[0] ?? '');\n"+
		"if (input.kind === 'Ok') { for (const request of input.text.split('\\n').slice(0, -1)) { console.log(handleRequest(request)); } } else { console.error(input.message); }\n")
	binary := filepath.Join(setup.directory, "service-native")
	wasiCommand(t, setup.repository, setup.compiler, "build", commandSource, "-o", binary, "--sanitize")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	native, err := exec.CommandContext(ctx, binary, quarterRequests).Output()
	if err != nil {
		if failed, ok := err.(*exec.ExitError); ok {
			t.Fatalf("native service: %v\n%s", err, failed.Stderr)
		}
		t.Fatalf("native service: %v", err)
	}
	actualLines := strings.Split(strings.TrimSuffix(string(native), "\n"), "\n")
	for index := span[0]; index < span[1]; index++ {
		if index-span[0] >= len(actualLines) {
			t.Fatalf("native command ended before response %d", index)
		}
		if actualLines[index-span[0]] != responseLines[index] {
			t.Fatalf("native response %d: got %q, want %q", index, actualLines[index-span[0]], responseLines[index])
		}
	}
	if len(actualLines) != span[1]-span[0] {
		t.Fatalf("native command emitted %d lines, want %d", len(actualLines), span[1]-span[0])
	}
	t.Logf("native sanitized command: responses %d to %d agree with Node", span[0], span[1]-1)
}

func TestWASIServiceNativeAgrees_0(t *testing.T) {
	t.Parallel()
	nativeQuarter(t, 0)
}

func TestWASIServiceNativeAgrees_1(t *testing.T) {
	t.Parallel()
	nativeQuarter(t, 1)
}

func TestWASIServiceNativeAgrees_2(t *testing.T) {
	t.Parallel()
	nativeQuarter(t, 2)
}

func TestWASIServiceNativeAgrees_3(t *testing.T) {
	t.Parallel()
	nativeQuarter(t, 3)
}

// The four tests above are one per quarter, so a fifth quarter would go unrun; and quarters with a gap or an overlap
// must be refused, not passed.
func TestWASIServiceQuartersTile(t *testing.T) {
	t.Parallel()
	if len(serviceQuarters) != 4 {
		t.Fatalf("%d quarters, and TestWASIServiceNativeAgrees_0 to _3 run four", len(serviceQuarters))
	}
	if err := tileRequests(serviceQuarters, serviceRequests); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct {
		name   string
		ranges [][2]int
	}{
		{"gap", [][2]int{{0, 25000}, {25001, 50000}, {50000, 75000}, {75000, 100000}}},
		{"overlap", [][2]int{{0, 25000}, {24999, 50000}, {50000, 75000}, {75000, 100000}}},
		{"short", [][2]int{{0, 25000}, {25000, 50000}, {50000, 75000}, {75000, 99999}}},
		{"empty", [][2]int{{0, 25000}, {25000, 25000}, {25000, 75000}, {75000, 100000}}},
	} {
		if tileRequests(mutant.ranges, serviceRequests) == nil {
			t.Errorf("%s quarters passed the tiling check", mutant.name)
		}
	}
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
