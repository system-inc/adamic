package lower

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestParseIntMapUsesIndexRadix(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "console.log(['10','10','10'].map(Number.parseInt).join(','));")
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name != "library_map_parseInt" {
			continue
		}
		if len(function.Parameters) != 3 || len(function.Body) != 1 {
			t.Fatalf("unexpected callback: %#v", function)
		}
		ret, ok := function.Body[0].(ir.Return)
		if !ok {
			t.Fatalf("callback body: %T", function.Body[0])
		}
		call, ok := ret.Value.(ir.NumberCall)
		if !ok || call.Function != "parseInt" || len(call.Arguments) != 2 {
			t.Fatalf("parseInt call: %#v", ret.Value)
		}
		want := []ir.Expression{ir.Read{Local: function.Parameters[0], Of: ir.String}, ir.Read{Local: function.Parameters[1], Of: ir.Number}}
		if !reflect.DeepEqual(call.Arguments, want) {
			t.Fatalf("parseInt must receive element then index radix: got %#v, want %#v", call.Arguments, want)
		}
		return
	}
	t.Fatal("missing parseInt map callback")
}

func guardedFSCall(t *testing.T, source string) ir.NodeFSFile {
	t.Helper()
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	var calls []ir.NodeFSFile
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.CanInterface() {
			if call, ok := value.Interface().(ir.NodeFSFile); ok {
				calls = append(calls, call)
				return
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				visit(value.Elem())
			}
		case reflect.Struct:
			for i := 0; i < value.NumField(); i++ {
				visit(value.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < value.Len(); i++ {
				visit(value.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(program.Main))
	if len(calls) != 1 {
		t.Fatalf("want one fs call, got %d: %#v", len(calls), program.Main)
	}
	return calls[0]
}

func TestFSOpenStringFlagsLower(t *testing.T) {
	t.Parallel()
	call := guardedFSCall(t, "import {openSync} from 'node:fs'; openSync('x','r');")
	if call.Operation != "open" || len(call.Arguments) != 3 {
		t.Fatalf("want open with mode: %#v", call)
	}
}

func TestFSRemoveDefaultRetryDelayLowers(t *testing.T) {
	t.Parallel()
	call := guardedFSCall(t, "import {rmSync} from 'node:fs'; rmSync('x',{retryDelay:100});")
	if call.Operation != "rm" || len(call.Arguments) != 3 {
		t.Fatalf("want rm with driver defaults: %#v", call)
	}
}

func TestFSExistsOperation(t *testing.T) {
	t.Parallel()
	call := guardedFSCall(t, "import {existsSync} from 'node:fs'; existsSync('x');")
	if call.Operation != "exists" || call.Of != ir.Boolean {
		t.Fatalf("want boolean exists operation: %#v", call)
	}
}

func TestFSStatThrowsByDefault(t *testing.T) {
	t.Parallel()
	call := guardedFSCall(t, "import {statSync} from 'node:fs'; statSync('x');")
	if call.Operation != "stat" || len(call.Arguments) != 2 {
		t.Fatalf("want stat with throw argument: %#v", call)
	}
	if throws, ok := call.Arguments[1].(ir.BooleanConstant); !ok || !throws.Value {
		t.Fatalf("stat must throw by default: %#v", call.Arguments[1])
	}
}
