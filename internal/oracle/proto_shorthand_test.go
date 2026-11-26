package oracle

import (
	"path/filepath"
	"testing"
)

// A shorthand { __proto__ } makes an own key named __proto__ in JavaScript, never the prototype, so the
// JavaScript backend writes it computed (objectKey, internal/javascript/javascript.go) and each program
// here agrees with Node on both backends, natively under the leak check (#pyak7te, Adamic's ruling: fix
// the backend, don't refuse the program). The quoted key it wrote before set the prototype, and
// Object.keys then stopped with exit 70 on the JavaScript backend alone.
func TestShorthandProtoIsAnOwnKey(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"shorthand":    "__proto__,own\n",
		"inline_plain": "__proto__,own\n",
		"inline":       "__proto__,own\n",
		"then_clean":   "__proto__,own\na,z\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/proto_shorthand", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != stdout {
				t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, truth.exitCode, truth.stdout, truth.stderr)
			}
			backend := onJavaScriptBackend(t, program)
			if diff := disagreement(truth, backend); diff != "" {
				t.Fatalf("javascript: %s; backend stdout %q stderr %q exit %d", diff, backend.stdout, backend.stderr, backend.exitCode)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if diff := disagreement(truth, actual); diff != "" {
					t.Fatalf("native: %s; stdout %q stderr %q exit %d", diff, actual.stdout, actual.stderr, actual.exitCode)
				}
			}
		})
	}
}
