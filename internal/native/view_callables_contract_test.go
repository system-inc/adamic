package native

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// These exercise the lane-owned runtime helper, not compiler dispatch coverage.
func TestViewCallableShapeNative(t *testing.T) {
	t.Parallel()
	header, err := runtime.ReadFile("runtime/view_callables_contract.h")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, setup, out, found string }{
		{"valid", "", "8\n", ""},
		{"wrong-arity", "recorded.arity = 0;", "", "function with arity 0"},
		{"wrong-parameter", "recorded.parameters = boolean_parameters;", "", "function with incompatible parameter representations"},
		{"wrong-result", "recorded.result = 2;", "", "function with incompatible result representation"},
		{"unknown-result", "recorded.result = 0; expected.result = 0;", "", "function with unknown signature"},
		{"unknown-parameter", "recorded.parameters = unknown_parameters; expected.parameters = unknown_parameters;", "", "function with incompatible parameter representations"},
		{"unknown-signature", "metadata = NULL;", "", "function with unknown signature"},
		{"number", "value = &number.heap;", "", "number"},
		{"optional-absent", "value = NULL; optional = true;", "undefined\n", ""},
		{"required-absent", "value = NULL;", "", "undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := string(header) + `
static adamic_value add(adamic_closure *self, adamic_value *arguments, size_t count) {
    (void)self; (void)count;
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
    adamic_closure number = {.heap = {.references = 0, .kind = adamic_kind_number, .slab = 0}, .code = add, .receiver = false, .count = 0};
    (void)number;
    bool optional = false;
` + probe.setup + `
    const adamic_heap *checked = adamic_view_callable_shape(value, metadata, &expected, "node.run", optional);
    if (checked == NULL) { puts("undefined"); }
    else {
        adamic_value arguments[] = {{.number = 7}};
        const adamic_closure *callable = (const adamic_closure *)checked;
        printf("%.0f\n", callable->code((adamic_closure *)callable, arguments, 1).number);
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
		})
	}
}
