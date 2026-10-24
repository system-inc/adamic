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

func prototypeNode(t *testing.T, source string) string {
	t.Helper()
	// .a is an outside Node input here. Native lowering currently refuses these APIs.
	command := exec.Command("node", "--input-type=module", "-e", `import {stripTypeScriptTypes} from 'node:module';
const source = stripTypeScriptTypes(process.argv[1], {mode: 'transform'});
await import('data:text/javascript,' + encodeURIComponent(source));`, source)
	command.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("Node: %v\n%s", err, stderr.String())
	}
	return stdout.String()
}

func prototypeFixture(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "prototypes", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestPrototypeNodeResearch(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, expected, before, after string }{
		{"debug_shapes", "4 flow 4\n8 false\nNodeArray 1,2 true\nmapper 2\nflow false\nsource 7 false\nupdated\n", "Object.setPrototypeOf(flowNode, flowPrototype);", "Object.setPrototypeOf(flowNode, flowPrototype); flowNode.flags = 9;"},
		{"errors", "cycle: TypeError: Cyclic __proto__ value\ncreate: TypeError: Object prototype may only be an Object or null: 1\nset: TypeError: Object prototype may only be an Object or null: 1\nundefined: TypeError: Object prototype may only be an Object or null: undefined\nnull receiver: TypeError: Object.setPrototypeOf called on null or undefined\n", "Object.setPrototypeOf(first, second);", "Object.setPrototypeOf(first, null);"},
		{"mixed_cycle", "true\n", "prototype.child = object;", "prototype.child = undefined;"},
	}
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			source := prototypeFixture(t, fixture.name)
			if got := prototypeNode(t, source); got != fixture.expected {
				t.Fatalf("got %q, want %q", got, fixture.expected)
			}
			if strings.Count(source, fixture.before) != 1 {
				t.Fatal("mutant lost its anchor")
			}
			mutant := strings.Replace(source, fixture.before, fixture.after, 1)
			if got := prototypeNode(t, mutant); got == fixture.expected {
				t.Fatal("input mutant survived the output check")
			}
			t.Log("Node observation held; input mutant changes the required output")
		})
	}
}

// This is the requested admission mutant: bypass the source refusal and implement both
// b.a and a.[[Prototype]] as ordinary counted slots. LeakSanitizer must reject that admission.
func TestPrototypeMixedCountedCycleProbe(t *testing.T) {
	t.Parallel()
	expected := prototypeNode(t, prototypeFixture(t, "mixed_set_cycle"))
	directory := t.TempDir()
	counted := filepath.Join(directory, "counted")
	if err := native.Build(prototypeOwnershipModel, counted, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	command := exec.Command(counted)
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stdout.String() != expected {
		t.Fatalf("counted model: %v %q %s", err, stdout.String(), stderr.String())
	}
	report := prototypeUnbalanced(prototypeExecution{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()})
	if !strings.Contains(report, "heap values leaked: 2") {
		t.Fatalf("counting did not expose the cycle: %s %s", report, stderr.String())
	}
	t.Log(strings.TrimSpace(stderr.String()))

	sanitized := filepath.Join(directory, "sanitized")
	if err := native.Build(prototypeOwnershipModel, sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command = exec.Command(sanitized)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	stdout.Reset()
	stderr.Reset()
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stdout.String() != expected || stderr.Len() != 0 {
		t.Fatalf("sanitized model: %v %q %s", err, stdout.String(), stderr.String())
	}
	if report := prototypeLeaks(t, prototypeOwnershipModel, sanitized); report == "" {
		t.Fatal("the leak check missed the mixed ownership cycle")
	} else {
		if runtime.GOOS == "linux" && !strings.Contains(report, "LeakSanitizer: detected memory leaks") {
			t.Fatalf("admission mutant failed without the required leak report: %s", report)
		}
		t.Log(report)
	}

	// Removing the real data back edge defeats the leak observation. This mutant is still legal C,
	// and no compiler warning or crash can substitute for the leak check that must distinguish it.
	mutant := strings.Replace(prototypeOwnershipModel, "prototype->slots[0].reference = adamic_retain(object);", "prototype->slots[0].reference = NULL;", 1)
	binary := filepath.Join(directory, "no-back-edge")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := prototypeLeaks(t, mutant, binary); report != "" {
		t.Fatalf("without the back edge should free: %s", report)
	}

	// Main has no graph-region runtime yet. Keep the mixed-edge admission mutant
	// and back-edge control; graph admission remains a later unit.
	t.Log("counted admission leaks; back-edge removal frees it; graph runtime is not on main")
}

// Pin each debug.ts link separately: an earlier create/descriptor refusal must not hide the setter.
// Availability tests are recorded too, so a "no edge" site is not confused with a linked object.
func TestPrototypeDebugSiteRefusals(t *testing.T) {
	t.Parallel()
	cases := []struct{ line, api, expected string }{
		{"554", "", "true\n"}, {"558", "Object.create", "created\n"},
		{"561", "Object.setPrototypeOf", "linked\n"}, {"594", "", "true\n"},
		{"598", "Object.create", "created\n"}, {"601", "Object.setPrototypeOf", "linked\n"},
		{"859", "Object.setPrototypeOf", "linked\n"}, {"944", "Object.create", "created\n"},
	}
	for _, fixture := range cases {
		t.Run("debug_"+fixture.line, func(t *testing.T) {
			name := filepath.Join("testdata", "prototypes", "sites", "debug_"+fixture.line+".a")
			source, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			if got := prototypeNode(t, string(source)); got != fixture.expected {
				t.Fatalf("Node: %q, want %q", got, fixture.expected)
			}
			checked, err := load.Load([]string{name})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), checked)
			if fixture.api == "" {
				var refusal *lower.Refused
				if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "unbound-method") {
					t.Fatalf("availability refusal: %v", err)
				}
				t.Logf("refused availability probe, no prototype edge: %v", err)
				prototypeRefusalControl(t, name, "console.log(\"true\");\n")
				return
			}
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || refusal.What != fixture.api {
				t.Fatalf("want named %s refusal for debug.ts:%s, got %v", fixture.api, fixture.line, err)
			}
			if !strings.Contains(refusal.Fix, "prototypes expose or replace fields outside the declared shape") {
				t.Fatalf("refusal lost its reason: %v", err)
			}
			t.Logf("refused debug.ts:%s: %v", fixture.line, err)
			lines := strings.Split(string(source), "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, fixture.api+"(") {
					lines[index] = ""
				}
			}
			prototypeRefusalControl(t, name, strings.Join(lines, "\n"))
		})
	}
	name := filepath.Join("testdata", "prototypes", "mixed_set_cycle.a")
	checked, err := load.Load([]string{name})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), checked)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || refusal.What != "Object.setPrototypeOf" {
		t.Fatalf("mixed-edge admission not refused: %v", err)
	}
}

func prototypeRefusalControl(t *testing.T, original, source string) {
	t.Helper()
	name := filepath.Join(t.TempDir(), filepath.Base(original))
	if err := os.WriteFile(name, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{name})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lower.Lower(context.Background(), checked); err != nil {
		t.Fatalf("removing the prototype operation must defeat the refusal observation: %v", err)
	}
	t.Log("operation-removal mutant defeats the expected refusal")
}

type prototypeExecution struct {
	Stdout, Stderr []byte
	ExitCode       int
}

var prototypeCounts = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`)

func prototypeUnbalanced(run prototypeExecution) string {
	match := prototypeCounts.FindSubmatch(run.Stderr)
	if run.ExitCode != 0 || match == nil {
		return fmt.Sprintf("unfinished counted probe: %+v", run)
	}
	var counts [3]int64
	for index, group := range []int{1, 2, 6} {
		value, err := strconv.ParseInt(string(match[group]), 10, 64)
		if err != nil {
			return fmt.Sprint(err)
		}
		counts[index] = value
	}
	if counts[0] != counts[1]+counts[2] {
		return fmt.Sprintf("heap values leaked: %d", counts[0]-counts[1]-counts[2])
	}
	return ""
}
func prototypeLeaks(t *testing.T, source, binary string) string {
	t.Helper()
	if runtime.GOOS == "darwin" {
		counted := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(source, counted, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(counted)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if err != nil {
			t.Fatal(err)
		}
		if report := prototypeUnbalanced(prototypeExecution{stdout.Bytes(), stderr.Bytes(), 0}); report != "" {
			return report
		}
		command = exec.Command("leaks", "--atExit", "--", counted)
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("leaks: %v %s", err, output)
		}
		return ""
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("leak run: %v %s", err, output)
	}
	return ""
}
