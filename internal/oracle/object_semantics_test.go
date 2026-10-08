package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/object_semantics/cached_own.a", true, false,
	})
}

func TestObjectCachedIntrinsic(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/cached_own.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: stdout=%q stderr=%q exit=%d", want.stdout, want.stderr, want.exitCode)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(want, got); diff != "" {
		t.Errorf("native: %s", diff)
	}
	if got.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Error(report)
		}
	}
	if diff := disagreement(want, releasedUncached(t, program)); diff != "" {
		t.Errorf("release: %s", diff)
	}
	if diff := disagreement(want, onJavaScriptBackend(t, program)); diff != "" {
		t.Errorf("JavaScript: %s", diff)
	}
}

// A borrowed intrinsic must ignore the receiver's own hasOwnProperty. This
// mutant gives every query the false result of that overriding field instead.
func TestObjectCachedIntrinsicMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/cached_own.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if _, ok := value.(ir.HasOwn); ok {
			changed++
			return ir.BooleanConstant{Value: false}
		}
		return value
	}
	for index := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	if changed == 0 {
		t.Fatal("mutant did not reach a borrowed own-key query")
	}
	want := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish: %+v", got)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("native mutant escaped: %s", difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant escaped: %s", difference)
	}
	t.Logf("Node caught %d queries dispatched to the receiver override; native exited zero without leaks", changed)
}
