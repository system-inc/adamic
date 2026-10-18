package lower

import (
	"os"
	"testing"
)

func callableProducerSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/callable_producers/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestCallableProducerP21(t *testing.T) {
	t.Parallel()
	// Native shows that a named producer's code identity uses its closure thunk.
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p21"))
}
func TestCallableProducerP60(t *testing.T) {
	t.Parallel()
	// Native checks the named function ABI, which JavaScript cannot expose.
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p60"))
}
func TestCallableProducerP61(t *testing.T) {
	t.Parallel()
	// Native must not compare the captured class method's direct-call ABI.
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p61"))
}
func TestCallableProducerCertificatesUseClosureThunks(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNode(t, callableProducerSource(t, "p21"))
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
