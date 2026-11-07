package native_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// These are research probes, not a prototype implementation. The C model uses two ordinary
// counted slots to isolate the ownership rule proposed for a hidden prototype link.
const prototypeOwnershipModel = `#include "adamic.h"
#include <stdio.h>
// Both outside references are explicitly released. Ignore stale stack/register bit patterns,
// which can otherwise keep this intentionally leaked cycle conservatively reachable to LSan.
const char *__lsan_default_options(void) { return "use_stacks=0:use_registers=0"; }
static const char *const names[] = {"a", "[[Prototype]]"};
static const bool references[] = {true, true};
static const adamic_field_kind kinds[] = {adamic_field_reference,adamic_field_reference};
static const adamic_shape shape = {2, names, references, NULL, kinds};
__attribute__((noinline)) static void run(void) {
    adamic_object *prototype = adamic_object_new(&shape);
    adamic_object *object = adamic_object_new(&shape);
    /* graph allocation classification */
    prototype->slots[0].reference = adamic_retain(object);
    object->slots[1].reference = adamic_retain(prototype);
    printf("%s\n", ((adamic_object *)object->slots[1].reference)->slots[0].reference == object ? "true" : "false");
    adamic_release(object);
    adamic_release(prototype);
}
int main(void) { run(); return 0; }
`

func prototypeNode(test *testing.T, source string) string {
	test.Helper()
	// .a is an outside Node input here. Native lowering currently refuses these APIs.
	command := exec.Command("node", "--input-type=module", "-e", `import {stripTypeScriptTypes} from 'node:module';
const source = stripTypeScriptTypes(process.argv[1], {mode: 'transform'});
await import('data:text/javascript,' + encodeURIComponent(source));`, source)
	command.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if runError := command.Run(); runError != nil || stderr.Len() != 0 {
		test.Fatalf("Node: %v\n%s", runError, stderr.String())
	}
	return stdout.String()
}

func prototypeFixture(test *testing.T, name string) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata", "prototypes", name+".a"))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func TestPrototypeNodeResearch(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, expected, before, after string }{
		{
			"debug_shapes",
			"4 flow 4\n8 false\nNodeArray 1,2 true\nmapper 2\nflow false\nsource 7 false\nupdated\n",
			"Object.setPrototypeOf(flowNode, flowPrototype);",
			"Object.setPrototypeOf(flowNode, flowPrototype); flowNode.flags = 9;",
		},
		{
			"errors",
			"cycle: TypeError: Cyclic __proto__ value\ncreate: TypeError: Object prototype may only be an Object or null: 1\nset: TypeError: Object prototype may only be an Object or null: 1\nundefined: TypeError: Object prototype may only be an Object or null: undefined\nnull receiver: TypeError: Object.setPrototypeOf called on null or undefined\n",
			"Object.setPrototypeOf(first, second);",
			"Object.setPrototypeOf(first, null);",
		},
		{"mixed_cycle", "true\n", "prototype.child = object;", "prototype.child = undefined;"},
	}
	for _, fixture := range cases {
		test.Run(fixture.name, func(test *testing.T) {
			source := prototypeFixture(test, fixture.name)
			if got := prototypeNode(test, source); got != fixture.expected {
				test.Fatalf("got %q, want %q", got, fixture.expected)
			}
			if strings.Count(source, fixture.before) != 1 {
				test.Fatal("mutant lost its anchor")
			}
			mutant := strings.Replace(source, fixture.before, fixture.after, 1)
			if got := prototypeNode(test, mutant); got == fixture.expected {
				test.Fatal("input mutant survived the output check")
			}
			test.Log("Node observation held; input mutant changes the required output")
		})
	}
}

// This is the requested admission mutant: bypass the source refusal and implement both
// b.a and a.[[Prototype]] as ordinary counted slots. LeakSanitizer must reject that admission.
func TestPrototypeMixedCountedCycleProbe(test *testing.T) {
	test.Parallel()
	expected := prototypeNode(test, prototypeFixture(test, "mixed_set_cycle"))
	directory := test.TempDir()
	counted := filepath.Join(directory, "counted")
	if buildError := native.Build(prototypeOwnershipModel, counted, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	var stdout, stderr bytes.Buffer
	command := exec.Command(counted)
	command.Stdout, command.Stderr = &stdout, &stderr
	if runError := command.Run(); runError != nil || stdout.String() != expected {
		test.Fatalf("counted model: %v %q %s", runError, stdout.String(), stderr.String())
	}
	report := prototypeUnbalanced(prototypeExecution{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()})
	if !strings.Contains(report, "heap values leaked: 2") {
		test.Fatalf("counting did not expose the cycle: %s %s", report, stderr.String())
	}
	test.Log(strings.TrimSpace(stderr.String()))

	sanitized := filepath.Join(directory, "sanitized")
	if buildError := native.Build(
		prototypeOwnershipModel,
		sanitized,
		native.Options{Sanitize: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	command = exec.Command(sanitized)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	stdout.Reset()
	stderr.Reset()
	command.Stdout, command.Stderr = &stdout, &stderr
	if runError := command.Run(); runError != nil || stdout.String() != expected || stderr.Len() != 0 {
		test.Fatalf("sanitized model: %v %q %s", runError, stdout.String(), stderr.String())
	}
	if report := prototypeLeaks(test, prototypeOwnershipModel, sanitized); report == "" {
		test.Fatal("the leak check missed the mixed ownership cycle")
	} else {
		if runtime.GOOS == "linux" && !strings.Contains(report, "LeakSanitizer: detected memory leaks") {
			test.Fatalf("admission mutant failed without the required leak report: %s", report)
		}
		test.Log(report)
	}

	// Removing the real data back edge defeats the leak observation. This mutant is still legal C,
	// and no compiler warning or crash can substitute for the leak check that must distinguish it.
	mutant := strings.Replace(
		prototypeOwnershipModel,
		"prototype->slots[0].reference = adamic_retain(object);",
		"prototype->slots[0].reference = NULL;",
		1,
	)
	binary := filepath.Join(directory, "no-back-edge")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := prototypeLeaks(test, mutant, binary); report != "" {
		test.Fatalf("without the back edge should free: %s", report)
	}

	// Main has no graph-region runtime yet. Keep the mixed-edge admission mutant
	// and back-edge control; graph admission remains a later unit.
	test.Log("counted admission leaks; back-edge removal frees it; graph runtime is not on main")
}

// Pin each debug.ts link separately: an earlier create/descriptor refusal must not hide the setter.
// Availability tests are recorded too, so a "no edge" site is not confused with a linked object: they
// lower, and print what Node does on both backends.
func TestPrototypeDebugSiteRefusals(test *testing.T) {
	test.Parallel()
	cases := []struct{ line, api, expected string }{
		{"554", "", "true\n"}, {"558", "Object.create", "created\n"},
		{"561", "Object.setPrototypeOf", "linked\n"}, {"594", "", "true\n"},
		{"598", "Object.create", "created\n"}, {"601", "Object.setPrototypeOf", "linked\n"},
		{"859", "Object.setPrototypeOf", "linked\n"}, {"944", "Object.create", "created\n"},
	}
	for _, fixture := range cases {
		test.Run("debug_"+fixture.line, func(test *testing.T) {
			name := filepath.Join("testdata", "prototypes", "sites", "debug_"+fixture.line+".a")
			source, readError := os.ReadFile(name)
			if readError != nil {
				test.Fatal(readError)
			}
			if got := prototypeNode(test, string(source)); got != fixture.expected {
				test.Fatalf("Node: %q, want %q", got, fixture.expected)
			}
			checked, loadError := load.Load([]string{name})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerError := lower.Lower(context.Background(), checked)
			if fixture.api == "" {
				// typeof Object.setPrototypeOf only tests for the method, an observation, not a detach
				// (lower/method_presence.go, #4q1b2er): it lowers, and each backend prints Node's answer.
				if lowerError != nil {
					test.Fatalf("availability probe: %v", lowerError)
				}
				prototypeBackendsPrint(test, program, fixture.expected)
				test.Log("lowered availability probe, no prototype edge")
				return
			}
			var refusal *lower.Refused
			if !errors.As(lowerError, &refusal) || refusal.What != fixture.api {
				test.Fatalf("want named %s refusal for debug.ts:%s, got %v", fixture.api, fixture.line, lowerError)
			}
			if !strings.Contains(refusal.Fix, "prototypes expose or replace fields outside the declared shape") {
				test.Fatalf("refusal lost its reason: %v", lowerError)
			}
			test.Logf("refused debug.ts:%s: %v", fixture.line, lowerError)
			lines := strings.Split(string(source), "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, fixture.api+"(") {
					lines[index] = ""
				}
			}
			prototypeRefusalControl(test, name, strings.Join(lines, "\n"))
		})
	}
	name := filepath.Join("testdata", "prototypes", "mixed_set_cycle.a")
	checked, loadError := load.Load([]string{name})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), checked)
	var refusal *lower.Refused
	if !errors.As(lowerError, &refusal) || refusal.What != "Object.setPrototypeOf" {
		test.Fatalf("mixed-edge admission not refused: %v", lowerError)
	}
}

// prototypeBackendsPrint runs program on the JavaScript backend and natively, sanitized, and fails
// unless each prints expected and exits cleanly.
func prototypeBackendsPrint(test *testing.T, program *ir.Program, expected string) {
	test.Helper()
	directory := test.TempDir()
	script := filepath.Join(directory, "main.mjs")
	if writeError := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o600); writeError != nil {
		test.Fatal(writeError)
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for _, backend := range []struct {
		name    string
		command *exec.Cmd
	}{
		{"javascript", exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", script)},
		{"native", exec.Command(binary)},
	} {
		var stdout, stderr bytes.Buffer
		backend.command.Stdout, backend.command.Stderr = &stdout, &stderr
		if runError := backend.command.Run(); runError != nil || stdout.String() != expected {
			test.Fatalf(
				"%s: %v, stdout %q, want %q\n%s",
				backend.name,
				runError,
				stdout.String(),
				expected,
				stderr.String(),
			)
		}
	}
}

func prototypeRefusalControl(test *testing.T, original, source string) {
	test.Helper()
	name := filepath.Join(test.TempDir(), filepath.Base(original))
	if writeError := os.WriteFile(name, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	checked, loadError := load.Load([]string{name})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if _, lowerError := lower.Lower(context.Background(), checked); lowerError != nil {
		test.Fatalf("removing the prototype operation must defeat the refusal observation: %v", lowerError)
	}
	test.Log("operation-removal mutant defeats the expected refusal")
}

type prototypeExecution struct {
	Stdout, Stderr []byte
	ExitCode       int
}

var prototypeCounts = regexp.MustCompile(
	`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`,
)

func prototypeUnbalanced(run prototypeExecution) string {
	match := prototypeCounts.FindSubmatch(run.Stderr)
	if run.ExitCode != 0 || match == nil {
		return fmt.Sprintf("unfinished counted probe: %+v", run)
	}
	var counts [3]int64
	for index, group := range []int{1, 2, 6} {
		value, parseError := strconv.ParseInt(string(match[group]), 10, 64)
		if parseError != nil {
			return fmt.Sprint(parseError)
		}
		counts[index] = value
	}
	if counts[0] != counts[1]+counts[2] {
		return fmt.Sprintf("heap values leaked: %d", counts[0]-counts[1]-counts[2])
	}
	return ""
}
func prototypeLeaks(test *testing.T, source, binary string) string {
	test.Helper()
	if runtime.GOOS == "darwin" {
		counted := filepath.Join(test.TempDir(), "counted")
		if buildError := native.Build(source, counted, native.Options{Count: true}); buildError != nil {
			test.Fatal(buildError)
		}
		command := exec.Command(counted)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		runError := command.Run()
		if runError != nil {
			test.Fatal(runError)
		}
		if report := prototypeUnbalanced(prototypeExecution{stdout.Bytes(), stderr.Bytes(), 0}); report != "" {
			return report
		}
		command = exec.Command("leaks", "--atExit", "--", counted)
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil {
			return fmt.Sprintf("leaks: %v %s", combinedOutputError, output)
		}
		return ""
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		return fmt.Sprintf("leak run: %v %s", combinedOutputError, output)
	}
	return ""
}
