package lower

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tupleRecognitionSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/fx7_tuple_recognition/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestTupleRecognitionP68(t *testing.T) {
	t.Parallel()
	source := tupleRecognitionSource(t, "p68")
	lowersAndAgreesWithNode(t, source)
	// Native's object-backed tuple is the representation the JavaScript marker follows.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestTupleRecognitionP69(t *testing.T) {
	t.Parallel()
	source := tupleRecognitionSource(t, "p69")
	lowersAndAgreesWithNode(t, source)
	// The selected tuple member must still support subsequent indexed field reads.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestTupleRecognitionOrdinaryArray(t *testing.T) {
	t.Parallel()
	source := `const values: number[] = [1,2]; console.log(String(values[0]) + ':' + String(values.length));`
	lowersAndAgreesWithNode(t, source)
	// Array literals retain their array layout and are not branded as tuple objects.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestTupleRecognitionRejectsArrayObjectView(t *testing.T) {
	t.Parallel()
	source := `interface Named { readonly text: string; }
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number] | Named; }
const pair: (string | number)[] = ['first',1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
const value = viewed.value;
console.log(typeof value);`
	original := filepath.Join(t.TempDir(), "source.a")
	if err := os.WriteFile(original, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := runAgreementNode(t, original)
	if truth.code != 0 || string(truth.stdout) != "object\n" {
		t.Fatalf("unexpected Node control: %+v", truth)
	}
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	observations := []nodeObservation{runAgreementNode(t, generated), runAgreementNative(t, native.C(program))}
	for index, got := range observations {
		t.Logf("backend %d exit %d stdout %q stderr %q", index, got.code, got.stdout, got.stderr)
	}
	for _, got := range observations {
		if got.code != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "field read failed") {
			t.Fatalf("ordinary arrays must fail object-backed tuple membership: %+v", got)
		}
	}
}
