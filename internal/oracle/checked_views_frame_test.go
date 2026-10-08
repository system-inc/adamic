package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func checkV1SourceNode(t *testing.T, path, name string) {
	t.Helper()
	want := run{}
	switch name {
	case "objects-good", "interfaces-good":
		want.stdout = []byte("true:okok\ntrue\n")
	case "objects-untagged-good":
		want.stdout = []byte("true:okok\ntrue:1\n")
	case "objects-wrong-nested", "objects-untagged-wrong", "interfaces-wrong-inherited":
		want.stdout = []byte("0\n")
	case "objects-missing-nested", "objects-uninitialized-nested", "interfaces-missing-inherited":
		want.stdout = []byte("undefined\n")
	case "interfaces-uninitialized-object":
		want.exitCode = 70
		want.stderr = []byte("adamic: panic: TypeError: Cannot read properties of undefined (reading 'ready')\n")
	default:
		t.Fatalf("missing Node pin for %s", name)
	}
	if diff := disagreement(want, onNode(t, path)); diff != "" {
		t.Fatal(diff)
	}
}

func TestCheckedViewFrameReadinessMutant(t *testing.T) {
	program, path := interfaceFixture(t, "lane1/objects-uninitialized-nested")
	checkV1SourceNode(t, path, "objects-uninitialized-nested")
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.child.ready is not initialized; expected boolean, found uninitialized\n")}
	good, _ := nativelyUncached(t, program)
	if diff := disagreement(want, good); diff != "" {
		t.Fatal(diff)
	}
	changed := 0
	rewrite := func(node any) any {
		switch v := node.(type) {
		case ir.Field:
			if v.Uninitialized {
				v.Uninitialized = false
				changed++
			}
			return v
		case ir.SetProperty:
			if v.Uninitialized {
				v.Uninitialized = false
				changed++
			}
			return v
		}
		return node
	}
	program.Main = v1Rewrite(reflect.ValueOf(program.Main), rewrite).Interface().([]ir.Statement)
	program.Functions = v1Rewrite(reflect.ValueOf(program.Functions), rewrite).Interface().([]ir.Function)
	program.Classes = v1Rewrite(reflect.ValueOf(program.Classes), rewrite).Interface().([]ir.Class)
	if changed == 0 {
		t.Fatal("readiness mutant changed no slot")
	}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "false\n" || len(got.stderr) != 0 {
			t.Fatalf("readiness mutant must run valid code with false: %#v", got)
		}
		if disagreement(want, got) == "" {
			t.Fatal("readiness mutant survived")
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("drop readiness caught by pinned exit 70 and uninitialized diagnostic; mutant emits false with exit 0")
}

func v1Rewrite(value reflect.Value, rewrite func(any) any) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return value
		}
		mapped := v1Rewrite(value.Elem(), rewrite)
		out := reflect.New(value.Type()).Elem()
		out.Set(mapped)
		return out
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		out.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if out.Field(i).CanSet() && value.Field(i).CanInterface() {
				out.Field(i).Set(v1Rewrite(value.Field(i), rewrite))
			}
		}
		return reflect.ValueOf(rewrite(out.Interface()))
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			out.Index(i).Set(v1Rewrite(value.Index(i), rewrite))
		}
		return out
	default:
		return value
	}
}

func TestCheckedViewFrameTagAndOperandMutants(t *testing.T) {
	for _, name := range []string{"tag", "operand"} {
		t.Run(name, func(t *testing.T) {
			fixture := "visitor"
			if name == "tag" {
				fixture = "wrong-kind"
			}
			program, path := interfaceFixture(t, fixture)
			truth := onNode(t, path)
			want := run{stdout: []byte("idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n")}
			mutated := run{stdout: []byte("idid\n42\ntexttext\n5\ntrue\nonce3\ncalls 3\n")}
			if name == "tag" {
				want = run{exitCode: 70, stdout: []byte("casting\n"), stderr: []byte("adamic: panic: cast failed: this Node is not a Identifier\n")}
				mutated = run{stdout: []byte("casting\notherother\n")}
				if diff := disagreement(mutated, truth); diff != "" {
					t.Fatal(diff)
				}
			} else if diff := disagreement(want, truth); diff != "" {
				t.Fatal(diff)
			}
			changed := false
			rewrite := func(node any) any {
				cast, ok := node.(ir.CheckedCast)
				if !ok || changed {
					return node
				}
				if name == "operand" {
					if _, call := cast.Value.(ir.Call); !call {
						return node
					}
					changed = true
					return ir.Comma{Left: cast.Value, Right: cast}
				}
				changed = true
				return cast.Value
			}
			program.Main = v1Rewrite(reflect.ValueOf(program.Main), rewrite).Interface().([]ir.Statement)
			program.Functions = v1Rewrite(reflect.ValueOf(program.Functions), rewrite).Interface().([]ir.Function)
			if !changed {
				t.Fatal("mutant changed no cast")
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(mutated, got); diff != "" {
					t.Fatalf("valid mutant result: %s", diff)
				}
				if disagreement(want, got) == "" {
					t.Fatal("mutant survived pinned result")
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("%s mutant caught in both backends by pinned exit/stdout; valid C and no leaks", name)
		})
	}
}
