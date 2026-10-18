package native

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// These exercise the lane-owned runtime helper, not compiler dispatch coverage.
type viewCallableShapeProbe struct{ name, setup, out, found string }

func viewCallableShapeProbeCheck(t *testing.T, probe viewCallableShapeProbe) {
	t.Helper()
	header, err := runtime.ReadFile("runtime/view_callables_contract.h")
	if err != nil {
		t.Fatal(err)
	}
	source := string(header) + `
static adamic_value add(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    return (adamic_value){.number = arguments[0].number + 1};
}
int main(void) {
    unsigned char parameters[] = {1}, boolean_parameters[] = {2}, unknown_parameters[] = {0};
    (void)boolean_parameters; (void)unknown_parameters;
    adamic_callable_signature expected = {1, parameters, 1, "function(number) -> number", NULL, 0};
    adamic_callable_signature recorded = expected;
    const adamic_callable_signature *metadata = &recorded;
    adamic_closure *closure = adamic_closure_new(add, 0);
    const adamic_heap *value = &closure->heap;
    /* Keep a code slot in the non-callable witness so an illicit call is defined
       release C and fails on behavior, rather than a sanitizer or a warning. */
    adamic_closure number = {.heap = {.references = 0, .kind = adamic_kind_number, .slab = 0}, .code = add, .count = 0};
    (void)number;
    bool optional = false;
` + probe.setup + `
    const adamic_heap *checked = adamic_view_callable_shape(value, metadata, &expected, "node.run", optional);
    if (checked == NULL) { puts("undefined"); }
    else {
        adamic_value arguments[] = {{.number = 7}};
        const adamic_closure *callable = (const adamic_closure *)checked;
        printf("%.0f\n", callable->code((adamic_closure *)callable, arguments).number);
    }
    adamic_release(closure);
    return 0;
}
`
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "callable")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(binary)
		var out, stderr bytes.Buffer
		command.Stdout, command.Stderr = &out, &stderr
		err := command.Run()
		exit := 0
		if err != nil {
			if failure, ok := err.(*exec.ExitError); ok {
				exit = failure.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		wantExit, wantError := 0, ""
		if probe.found != "" {
			wantExit = 70
			wantError = fmt.Sprintf("adamic: panic: field read failed: node.run expected function(number) -> number, found %s\n", probe.found)
		}
		if exit != wantExit || out.String() != probe.out || stderr.String() != wantError {
			t.Fatalf("exit %d stdout %q stderr %q; want %d %q %q", exit, out.String(), stderr.String(), wantExit, probe.out, wantError)
		}
	}
}

func TestViewCallableShapeNativeValid(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"valid", "", "8\n", ""})
}

func TestViewCallableShapeNativeWrongArity(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"wrong-arity", "recorded.arity = 0;", "", "function with arity 0"})
}

func TestViewCallableShapeNativeWrongParameter(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"wrong-parameter", "recorded.parameters = boolean_parameters;", "", "function with incompatible parameter representations"})
}

func TestViewCallableShapeNativeWrongResult(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"wrong-result", "recorded.result = 2;", "", "function with incompatible result representation"})
}

func TestViewCallableShapeNativeUnknownResult(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"unknown-result", "recorded.result = 0; expected.result = 0;", "", "function with unknown signature"})
}

func TestViewCallableShapeNativeUnknownParameter(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"unknown-parameter", "recorded.parameters = unknown_parameters; expected.parameters = unknown_parameters;", "", "function with incompatible parameter representations"})
}

func TestViewCallableShapeNativeUnknownSignature(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"unknown-signature", "metadata = NULL;", "", "function with unknown signature"})
}

func TestViewCallableShapeNativeNumber(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"number", "value = &number.heap;", "", "number"})
}

func TestViewCallableShapeNativeOptionalAbsent(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"optional-absent", "value = NULL; optional = true;", "undefined\n", ""})
}

func TestViewCallableShapeNativeRequiredAbsent(t *testing.T) {
	t.Parallel()
	viewCallableShapeProbeCheck(t, viewCallableShapeProbe{"required-absent", "value = NULL;", "", "undefined"})
}
