package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Mutate the emitted semantic operation, not an expected string or the compiler's build.
// Each mutant must compile and disagree with the source running on Node.
func TestErrorClassMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name   string
		mutate func(*ir.Program) bool
	}{
		{"lose Error ancestry", func(p *ir.Program) bool {
			for i, c := range p.Classes {
				if c.Name == "MyError" {
					p.Classes[i].Base = 0
					p.Classes[i].OwnStart = 0
					return true
				}
			}
			return false
		}},
		{"lose TypeError ancestry", func(p *ir.Program) bool {
			for i, c := range p.Classes {
				if c.Name == "TypeError" {
					p.Classes[i].Base = 0
					p.Classes[i].OwnStart = 0
					return true
				}
			}
			return false
		}},
		{"RangeError has TypeError identity", func(p *ir.Program) bool {
			var identity int
			for _, c := range p.Classes {
				if c.Name == "TypeError" {
					identity = c.Definition
				}
			}
			for i, c := range p.Classes {
				if c.Name == "RangeError" {
					p.Classes[i].Definition = identity
					return true
				}
			}
			return false
		}},
		{"message lost in super", func(p *ir.Program) bool {
			return changeErrorFunction(p, "Error_initialize", func(value any) any {
				if set, ok := value.(ir.SetProperty); ok && set.Name == "message" {
					set.Value = ir.StringConstant{Index: errorString(p, "")}
					return set
				}
				return value
			})
		}},
		{"default name changed", func(p *ir.Program) bool {
			return changeErrorFunction(p, "Error_initialize", func(value any) any {
				if set, ok := value.(ir.SetProperty); ok && set.Name == "name" {
					set.Value = ir.StringConstant{Index: errorString(p, "MyError")}
					return set
				}
				return value
			})
		}},
		{"cause ignored", func(p *ir.Program) bool {
			return changeErrorFunction(p, "Error_initialize", func(value any) any {
				if set, ok := value.(ir.SetProperty); ok && set.Name == "cause" {
					set.Value = ir.Undefined{Of: ir.Union}
					return set
				}
				return value
			})
		}},
		{"empty name ignored", func(p *ir.Program) bool {
			return changeErrorFunction(p, "Error_toString", func(value any) any {
				if c, ok := value.(ir.Conditional); ok {
					if _, outer := c.WhenNot.(ir.Conditional); outer {
						c.Condition = ir.BooleanConstant{}
						return c
					}
				}
				return value
			})
		}},
		{"empty message ignored", func(p *ir.Program) bool {
			return changeErrorFunction(p, "Error_toString", func(value any) any {
				if c, ok := value.(ir.Conditional); ok {
					if _, inner := c.WhenNot.(ir.Concat); inner {
						c.Condition = ir.BooleanConstant{}
						return c
					}
				}
				return value
			})
		}},
		{"separator wrong", func(p *ir.Program) bool {
			for i, s := range p.Strings {
				if s == ": " {
					p.Strings[i] = ":: "
					return true
				}
			}
			return false
		}},
		{"instanceof folded true", func(p *ir.Program) bool {
			changed := false
			p.Main = errorRewrite(reflect.ValueOf(p.Main), func(value any) any {
				if test, ok := value.(ir.InstanceOf); ok && p.Classes[test.Class-1].Name == "TypeError" {
					if read, ok := test.Value.(ir.Read); ok && p.Locals[read.Local].Name == "made" {
						changed = true
						return ir.BooleanConstant{Value: true}
					}
				}
				return value
			}).Interface().([]ir.Statement)
			return changed
		}},
		{"toFixed guard removed", errorGuardMutant("toFixed() digits argument")},
		{"repeat guard removed", errorGuardMutant("Invalid count value")},
		{"normalize guard removed", errorGuardMutant("The normalization form")},
		{"precision guard removed", errorGuardMutant("toPrecision() argument")},
		{"radix guard removed", errorGuardMutant("toString() radix argument")},
		{"undefined read guard removed", errorFunctionGuardMutant("error_defined")},
		{"undefined write guard removed", errorFunctionGuardMutant("error_set_property")},
		{"exponential guard removed", errorGuardMutant("toExponential() argument")},
		{"default constructor loses cause", func(p *ir.Program) bool {
			return changeErrorFunction(p, "DefaultError_new_initialize", func(value any) any {
				if call, ok := value.(ir.Call); ok && p.Functions[call.Function].Name == "Error_initialize" {
					call.Arguments = append([]ir.Expression{}, call.Arguments...)
					call.Arguments[2] = ir.Undefined{Of: ir.Union}
					return call
				}
				return value
			})
		}},
		{"undefined primitive receiver default lost", errorPrototypeDefaultMutant("undefined")},
		{"undefined generic name default lost", errorPrototypeDefaultMutant("Error")},
		{"undefined generic message default lost", errorPrototypeDefaultMutant("")},
		{"prototype throws RangeError", func(p *ir.Program) bool {
			var ctor int
			for _, c := range p.Classes {
				if c.Name == "RangeError" {
					ctor = c.Constructor
				}
			}
			return changeErrorFunction(p, "Error_prototype_toString", func(value any) any {
				if thrown, ok := value.(ir.Throw); ok {
					call := thrown.Value.(ir.Call)
					call.Function = ctor
					thrown.Value = call
					return thrown
				}
				return value
			})
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/error_classes.a"))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if !mutant.mutate(p) {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, p)
			if difference := disagreement(want, got); difference == "" {
				t.Fatal("mutant survived Node comparison")
			} else {
				t.Logf("caught: %s (Node exit %d, mutant exit %d)", difference, want.exitCode, got.exitCode)
			}
		})
	}
}

func errorString(p *ir.Program, text string) int {
	for i, s := range p.Strings {
		if s == text {
			return i
		}
	}
	p.Strings = append(p.Strings, text)
	return len(p.Strings) - 1
}

func changeErrorFunction(p *ir.Program, name string, change func(any) any) bool {
	changed := false
	for i, f := range p.Functions {
		if f.Name != name {
			continue
		}
		p.Functions[i].Body = errorRewrite(reflect.ValueOf(f.Body), func(value any) any {
			next := change(value)
			if !reflect.DeepEqual(value, next) {
				changed = true
			}
			return next
		}).Interface().([]ir.Statement)
	}
	return changed
}

func errorGuardMutant(message string) func(*ir.Program) bool {
	return func(p *ir.Program) bool {
		changed := false
		for i, f := range p.Functions {
			if !f.LibraryGuarded {
				continue
			}
			found := false
			errorRewrite(reflect.ValueOf(f.Body), func(value any) any {
				if c, ok := value.(ir.StringConstant); ok && strings.Contains(p.Strings[c.Index], message) {
					found = true
				}
				return value
			})
			if found {
				branch := p.Functions[i].Body[0].(ir.If)
				branch.Condition = ir.BooleanConstant{}
				p.Functions[i].Body[0] = branch
				changed = true
			}
		}
		return changed
	}
}

func errorRewrite(value reflect.Value, change func(any) any) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return value
		}
		out := reflect.New(value.Type()).Elem()
		out.Set(errorRewrite(value.Elem(), change))
		return out
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			out.Index(i).Set(errorRewrite(value.Index(i), change))
		}
		return out
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			out.Field(i).Set(errorRewrite(value.Field(i), change))
		}
		return reflect.ValueOf(change(out.Interface()))
	}
	return value
}

// Reporting mutants compile successfully and alter the runtime behavior after unwinding.
func TestErrorReportingMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, replacement string }{
		{"restored panic exit", `adamic_panic("restored panic exit", 19);`},
		{"lost buffered stdout", `_Exit(1);`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/error_classes_uncaught.a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			code := native.C(p)
			if !strings.Contains(code, "adamic_uncaught();") {
				t.Fatal("mutant changed nothing")
			}
			code = "#include <stdlib.h>\n" + strings.ReplaceAll(code, "adamic_uncaught();", mutant.replacement)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			if difference := disagreement(want, got); difference == "" {
				t.Fatal("mutant survived Node comparison")
			} else {
				t.Logf("caught: %s (Node exit %d, mutant exit %d)", difference, want.exitCode, got.exitCode)
			}
		})
	}
}

// Stock Node's uncaught handler, without the oracle's diagnostic normalization, exits 1.
func TestStockNodeErrorExit(t *testing.T) {
	t.Parallel()
	got := execute(t, "node", "--eval", `throw new Error("failed 🦉")`)
	if got.exitCode != 1 || !strings.Contains(string(got.stderr), "Error: failed 🦉\n") {
		t.Fatalf("stock Node: exit %d, stderr %q", got.exitCode, got.stderr)
	}
}

func errorFunctionGuardMutant(name string) func(*ir.Program) bool {
	return func(p *ir.Program) bool {
		return changeErrorFunction(p, name, func(value any) any {
			if branch, ok := value.(ir.If); ok {
				branch.Condition = ir.BooleanConstant{}
				return branch
			}
			return value
		})
	}
}

func errorPrototypeDefaultMutant(fallback string) func(*ir.Program) bool {
	return func(p *ir.Program) bool {
		return changeErrorFunction(p, "Error_prototype_toString", func(value any) any {
			if coalesce, ok := value.(ir.Coalesce); ok {
				if text, ok := coalesce.Fallback.(ir.StringConstant); ok && p.Strings[text.Index] == fallback {
					coalesce.Fallback = ir.StringConstant{Index: errorString(p, "lost default")}
					return coalesce
				}
			}
			return value
		})
	}
}
