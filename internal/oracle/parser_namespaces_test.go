package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, path := range []string{
		"stage3/namespace-parser-stops/native-namespace-object-receiver.a",
		"internal/oracle/testdata/namespace_method_receiver.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestParserNamespaceReceiver(t *testing.T) {
	for _, source := range []string{"stage3/namespace-parser-stops/native-namespace-object-receiver.a", "internal/oracle/testdata/namespace_method_receiver.a"} {
		path, err := filepath.Abs(filepath.Join(repository, source))
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		compiled, sanitized := natively(t, program)
		for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
			if diff := disagreement(truth, got); diff != "" {
				t.Fatalf("%s %s: %s; Node %+v; backend %+v", source, backend, diff, truth, got)
			}
		}
		if report := leaks(t, program, sanitized); report != "" {
			t.Fatal(report)
		}
		t.Logf("%s: both backends match Node stdout=%q exit=%d", source, truth.stdout, truth.exitCode)
	}
}

func TestParserNamespaceReceiverMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespace_method_receiver.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
		if property, ok := value.(ir.Property); ok && property.Name == "flags" {
			changed = true
			return ir.NumberConstant{Value: 99}
		}
		return value
	})
	if !changed {
		t.Fatal("mutant changed no actual receiver read")
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s wrong receiver mutant survived: %+v", backend, got)
		}
		t.Logf("%s namespace receiver substituted for method receiver: Node stdout catches clean exit 0", backend)
	}
}
