package lower

import (
	"os"
	"path/filepath"
	"testing"
)

// callableProducerSource returns the fixture file's source and its extension, the mode it lowers in.
func callableProducerSource(test *testing.T, file string) (string, string) {
	test.Helper()
	source, readError := os.ReadFile("testdata/callable_producers/" + file)
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source), filepath.Ext(file)
}

func TestCallableProducerP21(test *testing.T) {
	test.Parallel()
	// Native shows that a named producer's code identity uses its closure thunk.
	source, extension := callableProducerSource(test, "p21.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}
func TestCallableProducerP60(test *testing.T) {
	test.Parallel()
	// Native checks the named function ABI, which JavaScript cannot expose.
	source, extension := callableProducerSource(test, "p60.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}
func TestCallableProducerP61(test *testing.T) {
	test.Parallel()
	// Native must not compare the captured class method's direct-call ABI.
	source, extension := callableProducerSource(test, "p61.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}
func TestCallableProducerCertificatesUseClosureThunks(test *testing.T) {
	test.Parallel()
	source, extension := callableProducerSource(test, "p21.ts")
	program := lowersAndAgreesWithNodeAs(test, source, extension)
	// The native defensive filter masks stale certificates at runtime; inspect
	// this lowering invariant so that guard cannot hide an invalid certificate.
	for _, contract := range program.ViewContracts {
		if !contract.ProducerCertified {
			continue
		}
		for _, index := range contract.Functions {
			function := program.Functions[index]
			if !function.Closure || function.Receiver {
				test.Fatalf("producer certificate includes direct function %s", function.Name)
			}
		}
	}
}
