package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func errorOwnIntrinsicPath(t *testing.T, path string) string {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	return absolute
}

func TestDetachedErrorOwnIntrinsic(t *testing.T) {
	t.Parallel()
	const fixtures = "internal/oracle/testdata/error_own_intrinsic"
	for _, probe := range []struct {
		path   string
		stdout string
	}{
		{fixtures + "/162-166-error-hidden-by-view.ts", "true\n"},
		{fixtures + "/table.ts", "false\ntrue\ntrue\nfalse\ntrue\nfalse\ntrue\n"},
		{fixtures + "/message-after-write.ts", "false\ntrue\n"},
		{fixtures + "/coded.ts", "false\ntrue\n"},
		{fixtures + "/type-error.ts", "true\n"},
		{fixtures + "/message-loop.ts", "false\ntrue\ntrue\n"},
		{fixtures + "/map.ts", "false\ntrue\n"},
		{"internal/oracle/testdata/error_own_keys/detached_has_own_runtime.a", "true\n"},
	} {
		t.Run(filepath.Base(probe.path), func(t *testing.T) {
			t.Parallel()
			path := errorOwnIntrinsicPath(t, probe.path)
			truth := onNode(t, path)
			if why := disagreement(run{stdout: []byte(probe.stdout)}, truth); why != "" {
				t.Fatalf("Node: %s", why)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			for backend, actual := range map[string]run{
				"javascript": onJavaScriptBackend(t, program),
				"native":     released(t, program),
			} {
				if why := disagreement(truth, actual); why != "" {
					t.Errorf("%s: %s, stdout %q, stderr %q", backend, why, actual.stdout, actual.stderr)
				}
			}
		})
	}
}

func TestDetachedErrorOwnIntrinsicStops(t *testing.T) {
	t.Parallel()
	const outsideTable = "an Error's own property isn't represented yet"
	for _, probe := range []struct{ file, stdout, message string }{
		{"unknown-key.ts", "false\n", outsideTable},
		{"caught-code.ts", "true\n", outsideTable},
		{"stack-value.ts", "string\n", "an Error property Node sets and the native Error lacks isn't read yet (reading 'stack')"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := errorOwnIntrinsicPath(t, filepath.Join("internal/oracle/runtime_probes/error_own_intrinsic", probe.file))
			truth := onNode(t, path)
			if why := disagreement(run{stdout: []byte(probe.stdout)}, truth); why != "" {
				t.Fatalf("Node: %s", why)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.message + "\n")}
			for backend, actual := range map[string]run{
				"javascript": onJavaScriptBackend(t, program),
				"native":     released(t, program),
			} {
				if why := disagreement(want, actual); why != "" {
					t.Errorf("%s: %s, stdout %q, stderr %q", backend, why, actual.stdout, actual.stderr)
				}
			}
		})
	}
}

func TestErrorOwnIntrinsicCompileStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, message string }{
		{"166-error-hidden-by-view.ts", "an Error may be hidden by the view"},
		{"noncallable-own-field.ts", "an Error may be hidden by the view"},
		{"callable-incompatible-return.ts", "an Error may be hidden by the view"},
		{"callable-boolean-control.ts", "an Error may be hidden by the view"},
		{"override-field-number.ts", "an Error may be hidden by the view"},
		{"override-method-number.ts", "an Error may be hidden by the view"},
		{"error-options-constructor.ts", "new Error with options"},
		{"error-own-cause.ts", "new Error with options"},
		{"override-union.ts", "an Error may be hidden by the view"},
		{"error-own-aggregate.ts", "new an Identifier"},
		{"error-own-subclass.ts", "a computed class base; name the base class directly"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := errorOwnIntrinsicPath(t, filepath.Join("internal/oracle/testdata/error_own_intrinsic", probe.file))
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.message) {
				t.Fatalf("want NotYet with %q, got %v", probe.message, err)
			}
		})
	}
}
