package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Scratch probe for #fgqj0tm's design, never for main: which roads Library's stop covers, what the
// by-name road does on a plain copy natively, and the copy's IR.
func TestFgqDesignProbe(t *testing.T) {
	directory := filepath.Join(repository, "internal/oracle/testdata/fgq_design")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			t.Logf("PROBE %s node exit %d stdout %q stderr %.300q", name, truth.exitCode, truth.stdout, truth.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s lowering: %v", name, err)
				return
			}
			t.Logf("PROBE %s ir: %s", name, fgqDesignSummary(program))
			native, _ := natively(t, program)
			t.Logf("PROBE %s sanitized exit %d stdout %q stderr %.600q", name, native.exitCode, native.stdout, native.stderr)
			release := released(t, program)
			t.Logf("PROBE %s release exit %d stdout %q stderr %.300q", name, release.exitCode, release.stdout, release.stderr)
			backend := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit %d stdout %q stderr %.300q", name, backend.exitCode, backend.stdout, backend.stderr)
		})
	}
}

// fgqDesignSummary names each spread literal, each call through a function value and each direct call
// with a virtual slot, and each read of weigh or weight, in the program.
func fgqDesignSummary(program *ir.Program) string {
	found := []string{}
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}
			if value.Kind() == reflect.Interface {
				switch node := value.Elem().Interface().(type) {
				case ir.ObjectLiteral:
					if node.Spread != nil {
						fields := []string{}
						for _, field := range node.Fields {
							fields = append(fields, field.Name)
						}
						found = append(found, fmt.Sprintf("ObjectLiteral{Spread %T, Class %d, Fields %v, Methods %d, SpreadSite %d, NoReuse %t, SpreadGetters %v}", node.Spread, node.Class, fields, len(node.Methods), node.SpreadSite, node.NoReuse, node.SpreadGetters))
					}
				case ir.CallClosure:
					property, check, method := ir.MethodCallee(node.Closure)
					found = append(found, fmt.Sprintf("CallClosure{Closure %T, method %t, name %q, check %T %+v, Optional %t}", node.Closure, method, property.Name, check, check, node.Optional))
				case ir.Call:
					if node.Virtual != 0 || node.Accessor != "" {
						found = append(found, fmt.Sprintf("Call{Function %d %q, Virtual %d, Accessor %q}", node.Function, program.Functions[node.Function].Name, node.Virtual, node.Accessor))
					}
				case ir.Property:
					if node.Name == "weigh" || node.Name == "weight" {
						found = append(found, fmt.Sprintf("Property{Name %q, Of %d, Method %t, MethodTruthiness %t, Absent %t, Class %d, View %q}", node.Name, node.Of, node.Method, node.MethodTruthiness, node.Absent, node.Class, node.View))
					}
				}
			}
			visit(value.Elem())
			return
		}
		switch value.Kind() {
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				if value.Type().Field(index).IsExported() {
					visit(value.Field(index))
				}
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
		}
	}
	visit(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	classes := []string{}
	for _, class := range program.Classes {
		fields := []string{}
		for _, field := range class.Fields {
			fields = append(fields, fmt.Sprintf("%s(private %t)", field.Name, field.Private))
		}
		classes = append(classes, fmt.Sprintf("%s{Literal %t, Fields %v, Methods %v}", class.Name, class.Literal, fields, class.Methods))
	}
	return strings.Join(found, "; ") + " | classes: " + strings.Join(classes, "; ")
}
