package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, path := range []string{
		"stage3/namespace-parser-stops/native-namespace-object-receiver.a",
		"internal/oracle/testdata/namespace_method_receiver.a",
		"stage3/namespace-parser-stops/native-namespace-class.a",
		"internal/oracle/testdata/namespace_class_registration.a",
		"stage3/namespace-parser-stops/native-callable-namespace.a",
		"internal/oracle/testdata/namespace_callable_properties.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestParserNamespaceReceiver(t *testing.T) {
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-namespace-object-receiver.a", "internal/oracle/testdata/namespace_method_receiver.a"})
}

func TestParserNamespaceClass(t *testing.T) {
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-namespace-class.a", "internal/oracle/testdata/namespace_class_registration.a"})
}

func TestParserCallableNamespace(t *testing.T) {
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-callable-namespace.a", "internal/oracle/testdata/namespace_callable_properties.a"})
}

func TestParserCallableNamespaceMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespace_callable_properties.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	root, attached := -1, -1
	for i, function := range program.Functions {
		if function.Name == "log" {
			root = i
		}
		if function.Name == "error" {
			attached = i
		}
	}
	if root < 0 || attached < 0 {
		t.Fatal("mutant found no callable root or attached function")
	}
	redirect := func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.Call); ok && call.Function == attached {
			call.Function = root
			changed = true
			return call
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), redirect)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), redirect)
	if !changed {
		t.Fatal("mutant changed no actual attached function call")
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s attached function mutant survived: %+v", backend, got)
		}
		t.Logf("%s attached function confused with callable root caught by Node stdout", backend)
	}
}

func parserNamespaceMatchesNode(t *testing.T, sources []string) {
	t.Helper()
	for _, source := range sources {
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

func TestParserNamespaceClassRegistrationMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespace_class_registration.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	body := program.Main[:0]
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok && program.Locals[declaration.Local].Name == "DebugTypeMapper" {
			changed = true
			continue
		}
		body = append(body, statement)
	}
	program.Main = body
	if !changed {
		t.Fatal("mutant omitted no constructor registration")
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "ReferenceError: Cannot access 'DebugTypeMapper' before initialization") || disagreement(truth, got) == "" {
			t.Fatalf("%s missing registration mutant survived: %+v", backend, got)
		}
		t.Logf("%s missing constructor registration stopped: %+v", backend, got)
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
