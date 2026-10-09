package lower

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func requireInheritanceAnswer(t *testing.T, program *ir.Program) {
	t.Helper()
	if program == nil || len(program.Classes) == 0 || len(program.Main) == 0 {
		t.Fatal("accepted inheritance must produce classes and executable statements, not an empty answer")
	}
}

func TestInheritanceDerivedFieldOrder(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/inherited_fields_guard.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	requireInheritanceAnswer(t, program)
	for _, class := range program.Classes {
		if class.Name != "B" {
			continue
		}
		var fields []string
		for _, field := range class.Fields {
			fields = append(fields, field.Name)
		}
		if !reflect.DeepEqual(fields, []string{"x", "y"}) {
			t.Fatalf("B.Fields = %v, want inherited x before own y", fields)
		}
		return
	}
	t.Fatal("missing derived class B")
}

func TestInheritanceVoidBooleanOverrideReason(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "class A { f(): void {} }\nclass B extends A { override f(): boolean { return true; } }\nconst a: A = new B();")
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "different native result representation; keep the base method's result form") {
		t.Fatalf("want NotYet naming native result representation and repair, got %v", err)
	}
}
