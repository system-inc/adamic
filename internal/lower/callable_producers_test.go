package lower

import (
	"os"
	"path/filepath"
	"testing"
)

// callableProducerSource returns the fixture file's source and its extension, the mode it lowers in.
func callableProducerSource(t *testing.T, file string) (string, string) {
	t.Helper()
	source, err := os.ReadFile("testdata/callable_producers/" + file)
	if err != nil {
		t.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

func TestCallableProducerP21(t *testing.T) {
	t.Parallel()
	// Native shows that a named producer's code identity uses its closure thunk.
	source, extension := callableProducerSource(t, "p21.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}
func TestCallableProducerP60(t *testing.T) {
	t.Parallel()
	// Native checks the named function ABI, which JavaScript cannot expose.
	source, extension := callableProducerSource(t, "p60.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}
func TestCallableProducerP61(t *testing.T) {
	t.Parallel()
	// Native must not compare the captured class method's direct-call ABI.
	source, extension := callableProducerSource(t, "p61.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}
func TestCallableProducerCertificatesUseClosureThunks(t *testing.T) {
	t.Parallel()
	source, extension := callableProducerSource(t, "p21.ts")
	program := lowersAndAgreesWithNodeAs(t, source, extension)
	// The native defensive filter masks stale certificates at runtime; inspect
	// this lowering invariant so that guard cannot hide an invalid certificate.
	for _, contract := range program.ViewContracts {
		if !contract.ProducerCertified {
			continue
		}
		for _, index := range contract.Functions {
			function := program.Functions[index]
			if !function.Closure || function.Receiver {
				t.Fatalf("producer certificate includes direct function %s", function.Name)
			}
		}
	}
}
