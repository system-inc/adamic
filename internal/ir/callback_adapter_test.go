package ir

import "testing"

func TestCallbackAdapterTargets(t *testing.T) {
	t.Parallel()
	p := Program{Functions: []Function{{}, {Closure: true}}, Locals: []Local{{Type: Closure}}}
	nested := Effects{Body: []Statement{Declare{Local: 0, Value: MakeClosure{Function: 0}}}, Result: Effects{Result: MakeClosure{Function: 1}}}
	for _, call := range []Expression{ArrayMap{Callback: nested}, ArrayVisit{Callback: nested}, ArrayReduce{Callback: nested}, ArraySort{Callback: nested}, CallClosure{Closure: nested}} {
		got := p.ClosureTargets(call)
		if got.Unknown || len(got.Functions) != 1 || got.Functions[0] != 1 {
			t.Fatalf("%T: %#v", call, got)
		}
	}
	// Binding a closure does not prove the identity of an arbitrary read result.
	if got := p.ClosureTargets(CallClosure{Closure: Effects{Result: Read{Local: 0, Of: Closure}}}); !got.Unknown {
		t.Fatalf("unknown result became bounded: %#v", got)
	}
}
