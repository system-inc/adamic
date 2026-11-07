package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// callableContractSource returns the fixture file's source and its extension, the mode it lowers in.
func callableContractSource(test *testing.T, file string) (string, string) {
	test.Helper()
	source, err := os.ReadFile("testdata/fx7_wrong_aborts/" + file)
	if err != nil {
		test.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

func TestCallableContractP18(test *testing.T) {
	test.Parallel()
	source, extension := callableContractSource(test, "p18.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
	// Native must check the stored callable before packing its argument.

}

func TestCallableContractP20(test *testing.T) {
	test.Parallel()
	source, extension := callableContractSource(test, "p20.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
	// Native combines the checked union field and the checked callable read.

}

func TestCallableContractWideProducer(test *testing.T) {
	test.Parallel()
	source, extension := callableContractSource(test, "wide.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
	// The implementation's argument is boxed even though the union signature takes a number.

}

func TestCallableContractReceiverBinding(test *testing.T) {
	test.Parallel()
	source := `interface Runner { factor: number; run(value: number): number; }
const object: Runner = { factor: 2, run(value: number): number { return this.factor * value; } };
console.log(String(object.run(4)));`
	lowersAndAgreesWithNode(test, source)
	// Native must pass the original receiver to ordinary receiver methods.

}

func TestCallableContractReceiverEvaluatedOnce(test *testing.T) {
	test.Parallel()
	source, extension := callableContractSource(test, "p18.ts")
	source = strings.Replace(
		source,
		"console.log(`${view.value(4)}`);",
		`function receiver(): Receiver { console.log('receiver'); return view; }
function argument(): number { console.log('argument'); return 4; }
console.log(String(receiver().value(argument())));`,
		1,
	)
	lowersAndAgreesWithNodeAs(test, source, extension)
	// Native's checked member lookup must reuse the evaluated receiver.

}

func TestCallableContractSpreadRefused(test *testing.T) {
	test.Parallel()
	source, extension := callableContractSource(test, "p18.ts")
	source = strings.Replace(source, "view.value(4)", "view.value(...[4])", 1)
	_, err := lowerSourceAs(test, source, extension)
	if err == nil ||
		!strings.Contains(err.Error(), "main.ts:") ||
		!strings.Contains(err.Error(), "spread arguments through a checked callable view") {
		test.Fatalf("want located checked-callable spread refusal, got %v", err)
	}
}

func TestCallableContractPlainSpread(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("../oracle/testdata/closure_convention_receiver_rest.a")
	if err != nil {
		test.Fatal(err)
	}
	lowersAndAgreesWithNode(test, string(source))
	// Native retains the ordinary receiver/rest ABI outside checked callable views.

}
