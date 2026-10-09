package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func scout22Slice2Fixture(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	result, binary := natively(t, program)
	if d := disagreement(expected, result); d != "" {
		t.Fatalf("native: %s\nNode %+v\nnative %+v", d, expected, result)
	}
	if d := disagreement(expected, onJavaScriptBackend(t, program)); d != "" {
		t.Fatalf("JavaScript: %s", d)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if d := disagreement(expected, onWASI(t, native.C(program))); d != "" {
			t.Fatalf("WASI: %s", d)
		}
	}
}

// Mutations remain well typed and terminate without diagnostics or leaks. The
// source Node execution is the independent witness of the semantic difference.
func scout22Slice2Mutant(t *testing.T, family string) {
	t.Helper()
	fixture := family
	switch family {
	case "padStart", "normalize", "toFixed", "toExponential", "toPrecision", "toString":
		fixture = "range_errors"
	case "error_identity":
		fixture = "conversion_errors"
	}
	path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_"+fixture+".a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	stringHint := -1
	for i, function := range program.Functions {
		if strings.Contains(function.Name, "ordinary_string") {
			stringHint = i
			break
		}
	}
	wrong := len(program.Strings)
	program.Strings = append(program.Strings, "Wrong primitive conversion")
	mutate := func(value ir.Expression) ir.Expression {
		if changed != 0 {
			return value
		}
		switch family {
		case "conversion_hints":
			if call, ok := value.(ir.Call); ok && strings.Contains(program.Functions[call.Function].Name, "ordinary_default") && stringHint >= 0 {
				call.Function = stringHint
				changed++
				return call
			}
		case "conversion_errors":
			if call, ok := value.(ir.MakeError); ok {
				call.Message = ir.StringConstant{Index: wrong}
				changed++
				return call
			}
		case "property_errors":
			if call, ok := value.(ir.ObjectCall); ok && call.Method == "freeze" {
				call.Method = "preventExtensions"
				changed++
				return call
			}
		case "error_identity":
			if call, ok := value.(ir.ObjectCall); ok && call.Method == "errorIsType" {
				changed++
				return ir.BooleanConstant{Value: false}
			}
		case "padStart", "normalize":
			if call, ok := value.(ir.StringCall); ok && call.Method == family {
				if family == "normalize" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "NFC")
					call.Arguments[0] = ir.StringConstant{Index: index}
				} else {
					call.Arguments[0] = ir.NumberConstant{Value: 0}
				}
				changed++
				return call
			}
		case "toFixed":
			if call, ok := value.(ir.ToFixed); ok {
				call.Digits = ir.NumberConstant{Value: 0}
				changed++
				return call
			}
		case "toExponential", "toPrecision", "toString":
			if call, ok := value.(ir.NumberFormat); ok && call.Method == family {
				argument := 1.0
				if family == "toString" {
					argument = 10
				}
				call.Argument = ir.NumberConstant{Value: argument}
				changed++
				return call
			}
		case "range_errors":
			if call, ok := value.(ir.StringCall); ok && call.Method == "repeat" && len(call.Arguments) > 0 {
				// Skipping the negative-count error exercises the catch continuation.
				if count, constant := call.Arguments[0].(ir.NumberConstant); !constant || count.Value < 0 {
					call.Arguments[0] = ir.NumberConstant{Value: 0}
					changed++
					return call
				}
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
	if changed != 1 {
		t.Fatalf("want one mutation, got %d", changed)
	}
	scout22MutantMatchesOnlyNode(t, path, program)
}

func TestScout22ConversionHints(t *testing.T) {
	t.Parallel()
	scout22Slice2Fixture(t, "conversion_hints")
}

func TestScout22ConversionErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: TypeError and RangeError identity needs runtime error tags")
	scout22Slice2Fixture(t, "conversion_errors")
}

func TestScout22PropertyErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: frozen writes and Object.assign need catchable runtime TypeErrors")
	scout22Slice2Fixture(t, "property_errors")
}

func TestScout22RangeErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: repeat, padding, normalization and number formatting need catchable runtime RangeErrors")
	scout22Slice2Fixture(t, "range_errors")
}

func TestScout22TscToString(t *testing.T) {
	t.Parallel()
	scout22Slice2Fixture(t, "tsc_to_string")
}

func TestScout22MutantConversionHints(t *testing.T) {
	t.Parallel()
	scout22Slice2Mutant(t, "conversion_hints")
}

func TestScout22MutantConversionErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: conversion_errors mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "conversion_errors")
}

func TestScout22MutantPropertyErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: property_errors mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "property_errors")
}

func TestScout22MutantRangeErrors(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: range_errors mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "range_errors")
}

func TestScout22MutantPadStart(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: padStart mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "padStart")
}

func TestScout22MutantNormalize(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: normalize mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "normalize")
}

func TestScout22MutantToFixed(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: toFixed mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "toFixed")
}

func TestScout22MutantToExponential(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: toExponential mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "toExponential")
}

func TestScout22MutantToPrecision(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: toPrecision mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "toPrecision")
}

func TestScout22MutantToString(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: toString mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "toString")
}

func TestScout22MutantErrorIdentity(t *testing.T) {
	t.Parallel()
	t.Skip("awaits runtime/step22-to-primitive: error_identity mutant needs its catchable error fixture")
	scout22Slice2Mutant(t, "error_identity")
}

func init() {
	for _, path := range []string{"internal/oracle/testdata/scout22_conversion_hints.a", "internal/oracle/testdata/scout22_tsc_to_string.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestScout22PendingSourcesOnNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"conversion_errors", "property_errors", "range_errors", "prototype_create", "prototype_links", "prototype_lifetimes"} {
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_"+name+".a"))
		if err != nil {
			t.Fatal(err)
		}
		result := onNode(t, path)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s: Node %+v", name, result)
		}
	}
}

func TestScout22CachedIntrinsicProvenance(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/scout/22-slice2/gaps/compiler-cached-intrinsic.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	result, binary := natively(t, program)
	if d := disagreement(expected, result); d != "" {
		t.Fatalf("native: %s", d)
	}
	if d := disagreement(expected, onJavaScriptBackend(t, program)); d != "" {
		t.Fatalf("JavaScript: %s", d)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
