package lower

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func callableContractSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/fx7_wrong_aborts/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestCallableContractP18(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "p18")
	lowersAndAgreesWithNode(t, source)
	// Native must check the stored callable before packing its argument.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestCallableContractP20(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "p20")
	lowersAndAgreesWithNode(t, source)
	// Native combines the checked union field and the checked callable read.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestCallableContractWideProducer(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "wide")
	lowersAndAgreesWithNode(t, source)
	// The implementation's argument is boxed even though the union signature takes a number.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestCallableContractReceiverBinding(t *testing.T) {
	t.Parallel()
	source := `interface Runner { factor: number; run(value: number): number; }
const object: Runner = { factor: 2, run(value: number): number { return this.factor * value; } };
console.log(String(object.run(4)));`
	lowersAndAgreesWithNode(t, source)
	// Native must pass the original receiver to ordinary receiver methods.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestCallableContractReceiverEvaluatedOnce(t *testing.T) {
	t.Parallel()
	source := strings.Replace(callableContractSource(t, "p18"), "console.log(`${view.value(4)}`);", `function receiver(): Receiver { console.log('receiver'); return view; }
function argument(): number { console.log('argument'); return 4; }
console.log(String(receiver().value(argument())));`, 1)
	lowersAndAgreesWithNode(t, source)
	// Native's checked member lookup must reuse the evaluated receiver.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestCallableContractChecksBeforeArguments(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "bad_callable")
	original := filepath.Join(t.TempDir(), "source.a")
	if err := os.WriteFile(original, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := runAgreementNode(t, original)
	if truth.code != 0 || string(truth.stdout) != "before\nargument\n99\n" {
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
		t.Logf("backend %d: exit %d stdout %q stderr %q; Node stdout %q", index, got.code, got.stdout, got.stderr, truth.stdout)
	}
	for _, got := range observations {
		if got.code != 70 || string(got.stdout) != "before\n" || !strings.Contains(string(got.stderr), "expected") {
			t.Fatalf("callable contract must reject before argument effects: exit %d stdout %q stderr %q", got.code, got.stdout, got.stderr)
		}
	}
}

func TestCallableContractSpreadRefused(t *testing.T) {
	t.Parallel()
	source := strings.Replace(callableContractSource(t, "p18"), "view.value(4)", "view.value(...[4])", 1)
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "main.a:") || !strings.Contains(err.Error(), "spread arguments through a checked callable view") {
		t.Fatalf("want located checked-callable spread refusal, got %v", err)
	}
}

func TestCallableContractPlainSpread(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/closure_convention_receiver_rest.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNode(t, string(source))
	// Native retains the ordinary receiver/rest ABI outside checked callable views.
	lowersAndAgreesWithNodeNative(t, string(source))
}
