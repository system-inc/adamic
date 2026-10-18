package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func TestV4IteratorPackedReceiver(t *testing.T) {
	t.Parallel()
	source := `const iterable={limit:2,[Symbol.iterator](){let index=0;const count=this.limit;return {next(){const value=index;index=index+1;return {value,done:value>=count};}};}};for(const value of iterable){console.log(""+value);}`
	p, truth := v4IdentityProgram(t, source, "0\n1\n")
	v4EscapeAdmitted(t, p, truth)
	changed := 0
	mutate := func(e ir.Expression) ir.Expression {
		if call, ok := e.(ir.CallClosure); ok && call.ReceiverPacked {
			call.ReceiverPacked = false
			changed++
			return call
		}
		return e
	}
	mutateStringExpressions(reflect.ValueOf(&p.Main).Elem(), mutate)
	for i := range p.Functions {
		mutateStringExpressions(reflect.ValueOf(&p.Functions[i].Body).Elem(), mutate)
	}
	if changed == 0 {
		t.Fatal("no packed receiver sites")
	}
	got, _ := nativelyUncached(t, p)
	if disagreement(truth, got) == "" || got.exitCode == 0 {
		t.Fatalf("double receiver mutant survived: %#v", got)
	}
	t.Log("double receiver mutant caught by native sanitizer; JavaScript's existing packed ABI stays unchanged")
}
