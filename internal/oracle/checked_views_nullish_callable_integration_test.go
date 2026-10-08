package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// The function-kind guard alone cannot prove this optional callable's arity.
// A root-only read makes the certificate observable without invoking the value.
func TestCheckedViewNullishCallableArity(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/fixtures/callable-undefined-arity")
	truth := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("present\n")}, truth); difference != "" {
		t.Fatal(difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.value expected () => string, found function with arity 1\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := viewReadDisagreement(want, got, program); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
	changed := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "node.value" }, func(read ir.Property) ir.Property { read.ViewContract = 0; return read })
	if changed != 1 {
		t.Fatalf("want one certificate omission, got %d", changed)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("certificate omission must execute valid code: %s; got %#v", difference, got)
		}
		t.Logf("optional callable certificate omission caught by refusal pin: exit=%d stdout=%q", got.exitCode, got.stdout)
	}
}
