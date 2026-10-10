package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestStep21WritingCallRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/step21_builtin_narrow_terminal.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	want := path + ":7:21: Adamic 0.1 refuses a narrowed read of value after change() can write it; narrow again after the call"
	if err == nil || err.Error() != want {
		t.Fatalf("writing-call refusal: got %v, want %s", err, want)
	}
}

func TestStep21NonWritingCallControl(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/step21_builtin_narrow_control.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNodeNative(t, string(source))
}

func TestStep21NarrowAgainAfterWritingCall(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, "let value: unknown = new TypeError('first'); function change(): void { value = new Error('second'); } if (value instanceof TypeError) { change(); if (value instanceof TypeError) console.log(value.name); } console.log('after');")
}

func TestStep21TransitiveWritingCallRefusal(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "let value: unknown = new TypeError('first'); function write(): void { value = new Error('second'); } function call(): void { write(); } if (value instanceof TypeError) { call(); console.log(value.name); }")
	if err == nil {
		t.Fatal("admitted transitive writing call")
	}
	if refused, ok := err.(*Refused); !ok || refused.What != "a narrowed read of value after call() can write it" || refused.Fix != "narrow again after the call" {
		t.Fatalf("wrong writing-call refusal: %v", err)
	}
}
