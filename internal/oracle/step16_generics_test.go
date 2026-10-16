package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

const step16FixtureDirectory = "stage3/fixtures/generics/"

var step16Outcomes = []struct{ name, kind, reason string }{
	{"01_identity", "Lowered", ""},
	{"02_optional_return", "Lowered", ""},
	{"03_callback_return", "Lowered", ""},
	{"04_function_value", "NotYet", "a generic function as a value"},
	{"05_nested_class", "Lowered", ""},
	{"06_polymorphic_recursion", "Refused", "polymorphic recursion"},
	{"07_generic_cast", "Refused", "a cast the runtime can't check"},
	{"08_readonly_view", "Refused", "whose readonly field pos becomes writable"},
	{"09_recursive_optional", "Lowered", ""},
	{"10_identifier_multimap", "Lowered", ""},
	{"11_indexed_result", "Refused", "generic body return into slot T[\"value\"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }"},
	{"12_mixed_indexed_result", "Refused", "generic body return into slot U | T[\"value\"] | undefined from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained)"},
	{"13_constrained_local", "Refused", "generic body initializer into slot T[\"value\"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained)"},
	{"14_optional_literal_union", "Lowered", ""},
	{"15_array_callback", "Lowered", ""},
}

var step16BodyRefusals = map[string]string{
	"11_indexed_result":       "stage3/fixtures/generics/11_indexed_result.a:1:68: Adamic 0.1 refuses generic body return into slot T[\"value\"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations)",
	"12_mixed_indexed_result": "stage3/fixtures/generics/12_mixed_indexed_result.a:1:87: Adamic 0.1 refuses generic body return into slot U | T[\"value\"] | undefined from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained); read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations)",
	"13_constrained_local":    "stage3/fixtures/generics/13_constrained_local.a:2:11: Adamic 0.1 refuses generic body initializer into slot T[\"value\"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained); read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations)",
}

func init() {
	for _, fixture := range step16Outcomes {
		if fixture.kind == "Refused" {
			continue
		}
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			step16FixtureDirectory + fixture.name + ".a", fixture.kind == "Lowered", false,
		})
	}
}

func TestStep16GenericOutcomes(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range step16Outcomes {
		t.Run(fixture.name, func(t *testing.T) {
			path := filepath.Join(root, step16FixtureDirectory, fixture.name+".a")
			_, err := lowered(t, path)
			switch fixture.kind {
			case "Lowered":
				if err != nil {
					t.Fatal(err)
				}
			case "NotYet":
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != fixture.reason {
					t.Fatalf("want NotYet %q, got %v", fixture.reason, err)
				}
			case "Refused":
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, fixture.reason) {
					t.Fatalf("want Refused containing %q, got %v", fixture.reason, err)
				}
			}
			if want, pinned := step16BodyRefusals[fixture.name]; pinned {
				got := strings.TrimPrefix(err.Error(), root+string(filepath.Separator))
				if got != want {
					t.Fatalf("want exact ruled refusal %q, got %q", want, got)
				}
				node := onNode(t, path)
				if node.exitCode != 0 || string(node.stdout) != "2\n" || len(node.stderr) != 0 {
					t.Fatalf("source Node control changed: %+v", node)
				}
			}
			t.Logf("%s: %v", fixture.kind, err)
		})
	}
}

// A valid numeric result mutation must disagree with the independent source oracle.
func TestStep16IdentityMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, step16FixtureDirectory, "01_identity.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := range program.Functions {
		function := &program.Functions[i]
		if strings.HasPrefix(function.Name, "identity_") && function.Returns == ir.Number {
			function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 17}}}
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("want one numeric identity, changed %d", changed)
	}
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside output: %+v", actual)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant failed outside output: %+v", name, got)
		}
		if disagreement(expected, got) != "stdout differs" {
			t.Fatalf("%s result mutant escaped Node: %s", name, disagreement(expected, got))
		}
	}
	t.Log("numeric identity replaced with 17 caught by Node stdout in both backends; valid C, no sanitizer findings or leaks")
}

// Turning a present generic number into undefined keeps a valid optional ABI.
// Source Node must catch its changed fallback in both backends.
func TestStep16OptionalResultMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, step16FixtureDirectory, "02_optional_return.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if strings.HasPrefix(function.Name, "firstDefined_") && function.Returns == ir.MaybeNumber {
			function.Body = []ir.Statement{ir.Return{Value: ir.MaybeOf{Of: ir.MaybeNumber}}}
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("no optional numeric specialization to mutate")
	}
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside output: %+v", actual)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant failed outside output: %+v", name, got)
		}
		if disagreement(expected, got) != "stdout differs" {
			t.Fatalf("%s optional result mutant escaped Node: %s", name, disagreement(expected, got))
		}
	}
	t.Log("present generic number changed to undefined caught by Node stdout in both backends; valid optional ABI, no sanitizer findings or leaks")
}
