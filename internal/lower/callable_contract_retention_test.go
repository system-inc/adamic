package lower

import (
	"os"
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

}

func TestCallableContractP20(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "p20")
	lowersAndAgreesWithNode(t, source)
	// Native combines the checked union field and the checked callable read.

}

func TestCallableContractWideProducer(t *testing.T) {
	t.Parallel()
	source := callableContractSource(t, "wide")
	lowersAndAgreesWithNode(t, source)
	// The implementation's argument is boxed even though the union signature takes a number.

}

func TestCallableContractReceiverBinding(t *testing.T) {
	t.Parallel()
	source := `interface Runner { factor: number; run(value: number): number; }
const object: Runner = { factor: 2, run(value: number): number { return this.factor * value; } };
console.log(String(object.run(4)));`
	lowersAndAgreesWithNode(t, source)
	// Native must pass the original receiver to ordinary receiver methods.

}

func TestCallableContractReceiverEvaluatedOnce(t *testing.T) {
	t.Parallel()
	source := strings.Replace(callableContractSource(t, "p18"), "console.log(`${view.value(4)}`);", `function receiver(): Receiver { console.log('receiver'); return view; }
function argument(): number { console.log('argument'); return 4; }
console.log(String(receiver().value(argument())));`, 1)
	lowersAndAgreesWithNode(t, source)
	// Native's checked member lookup must reuse the evaluated receiver.

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

}
