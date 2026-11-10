package oracle

import (
	"path/filepath"
	"testing"
)

// A program namespace read as a value is one object, made at the start of its block (#67jk7wq). tsc's
// tracing shape and the object's identity, within a module and across an import, hold to Node natively,
// sanitized with leaks checked, and on the JavaScript backend.
func TestNamespaceValueAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, program := range []string{"tracing/main.a", "identity/main.a"} {
		t.Run(program, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, "internal/oracle/testdata/namespace_value/"+program)
		})
	}
}

// The name read as a value before its block ran, through an edge the initialization walk can't follow,
// is stopped by the namespace's readiness where Node reads undefined and goes on: native stops there,
// exit 70, and the JavaScript backend, carrying the same check, stops with it.
func TestNamespaceValueReadBeforeItsBlockIsChecked(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespace_value/premature.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	native, _ := natively(t, program)
	backend := onJavaScriptBackend(t, program)
	if native.exitCode != 70 || node.exitCode == 70 {
		t.Errorf("want the readiness check to stop natively (exit 70) where Node runs on: native %d, stdout %q, stderr %q; Node %d", native.exitCode, native.stdout, native.stderr, node.exitCode)
	}
	if difference := disagreement(backend, native); difference != "" {
		t.Errorf("%s\nbackend: exit %d, stdout %q, stderr %q\nnative:  exit %d, stdout %q, stderr %q", difference, backend.exitCode, backend.stdout, backend.stderr, native.exitCode, native.stdout, native.stderr)
	}
}
