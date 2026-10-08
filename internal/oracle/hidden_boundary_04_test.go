package oracle

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/hidden_boundary_optional_array_order.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/hidden_boundary_generic_optional_array.a", true, false})
	for _, path := range []string{"internal/oracle/testdata/hidden_boundary_optional_array_mutation_refused.a", "internal/oracle/testdata/hidden_boundary_alias_cast_refused.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, false, false})
	}
}

func TestHiddenBoundary04EvaluationOrder(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_optional_array_order.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if string(truth.stdout) != "-1|1|0\n8|2|1\n-1|-1|-1\nword|none\nfalse|true\n" || truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	compiled, sanitized := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(truth, got); diff != "" {
			t.Fatalf("%s: %s: %+v", backend, diff, got)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}

func TestHiddenBoundary04Node(t *testing.T) {
	path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_generic_optional_array.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	truth := onNode(t, path)
	if string(truth.stdout) != "7|0\n" || len(truth.stderr) != 0 || truth.exitCode != 0 {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	compiled, sanitized := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(truth, got); diff != "" {
			t.Fatalf("%s: %s: %+v", backend, diff, got)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
	t.Log("Node/native/JavaScript: stdout 7|0, empty stderr, exit 0; sanitizers and leaks pass")
}

func TestHiddenBoundary04PresentArrayMutant(t *testing.T) {
	path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_generic_optional_array.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index, function := range program.Functions {
		if strings.HasPrefix(function.Name, "first_") {
			if function.Returns != ir.MaybeNumber {
				t.Fatalf("unexpected mutant ABI: %v", function.Returns)
			}
			program.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.MaybeOf{Of: function.Returns}}}
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("mutant replaced no specialized first function")
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != "0|0\n" || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s mutant: %+v", backend, got)
		}
		t.Logf("%s returning undefined for present input caught by Node stdout, clean exit 0", backend)
	}
}

func TestHiddenBoundary04EagerIndexMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_optional_array_order.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		effects, ok := value.(ir.Effects)
		if !ok {
			return value
		}
		conditional, ok := effects.Result.(ir.Conditional)
		if !ok {
			return value
		}
		indexed, ok := conditional.WhenTrue.(ir.ArrayIndex)
		if !ok {
			return value
		}
		local := len(program.Locals)
		program.Locals = append(program.Locals, ir.Local{Name: "eager_index_mutant", Type: ir.Number, Function: -1, ExpressionAssigned: true})
		effects.Body = append(effects.Body, ir.Declare{Local: local, Value: indexed.Index})
		indexed.Index = ir.Read{Local: local, Of: ir.Number}
		conditional.WhenTrue = indexed
		effects.Result = conditional
		changed++
		return effects
	})
	if changed != 2 {
		t.Fatalf("mutated %d optional receiver calls", changed)
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || !strings.HasPrefix(string(got.stdout), "-1|1|1\n8|2|2\n") || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s mutant: %+v", backend, got)
		}
		t.Logf("%s eager index caught by Node side effects, clean exit 0", backend)
	}
}

func TestHiddenBoundary04UnsafeCasesRefused(t *testing.T) {
	for _, probe := range []struct{ file, reason, node string }{
		{"hidden_boundary_optional_array_mutation_refused.a", "instantiating a generic function", "text|7\n"},
		{"hidden_boundary_alias_cast_refused.a", "a cast the runtime can't check", "3\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.file))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: %+v", truth)
			}
			_, err := lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("unsafe acceptance or wrong stop: %v", err)
			}
			t.Logf("Node stdout=%q; compiler refusal: %v", truth.stdout, err)
		})
	}
}
