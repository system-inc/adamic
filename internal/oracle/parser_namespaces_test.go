package oracle

import (
	"os"
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
	t.Parallel()
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-namespace-object-receiver.a", "internal/oracle/testdata/namespace_method_receiver.a"})
}

func TestParserNamespaceClass(t *testing.T) {
	t.Parallel()
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-namespace-class.a", "internal/oracle/testdata/namespace_class_registration.a"})
}

func TestParserCallableNamespace(t *testing.T) {
	t.Parallel()
	parserNamespaceMatchesNode(t, []string{"stage3/namespace-parser-stops/native-callable-namespace.a", "internal/oracle/testdata/namespace_callable_properties.a"})
}

func TestParserCallableNamespaceMutant(t *testing.T) {
	t.Parallel()
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
		t.Run(source, func(t *testing.T) {
			t.Parallel()
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
		})
	}
}

func TestParserNamespaceClassRegistrationMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/namespace_class_registration.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Keep registration at module scope and catch only the omitted registration's readiness error.
	source = []byte(strings.Replace(string(source), "const first =", "try {\nconst first =", 1) + `
} catch (error) {
    if (error instanceof ReferenceError) console.log(error.message);
    else throw error;
}
`)
	path = filepath.Join(t.TempDir(), "registration-catch.a")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "0:1:false\nconstruct;construct;\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node registration: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control, controlBinary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": control, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s registration control: %s: %+v", backend, difference, got)
		}
	}
	if report := leaksUncached(t, program, controlBinary); report != "" {
		t.Fatal(report)
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
	compiled, mutantBinary := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		want := run{stdout: []byte("Cannot access 'DebugTypeMapper' before initialization\n")}
		if disagreement(want, got) != "" || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s missing registration mutant survived: %+v", backend, got)
		}
		t.Logf("%s missing constructor registration stopped: %+v", backend, got)
	}
	if report := leaks(t, program, mutantBinary); report != "" {
		t.Fatal(report)
	}
}

func TestParserNamespaceReceiverMutant(t *testing.T) {
	t.Parallel()
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
