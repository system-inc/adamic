package lower

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallableProducerLiteralReturn(t *testing.T) {
	t.Parallel()
	// Native confirms that the narrower logical result retains the scalar ABI.
	lowersAndAgreesWithNodeNative(t, callableProducerSource(t, "p23"))
}

func callableProducerAdapterRefusal(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs("testdata/callable_producers/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	want := runAgreementNode(t, path)
	if want.code != 0 || string(want.stdout) != "true\n" {
		t.Fatalf("Node control: %+v", want)
	}
	_, err = lowerSource(t, callableProducerSource(t, name))
	var refusal *NotYet
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "main.a:") || !strings.Contains(refusal.What, "wrap the producer in an arrow") {
		t.Fatalf("want located producer adapter refusal, got %v", err)
	}
	t.Log(err)
}
func TestCallableProducerFewerParametersRefused(t *testing.T) {
	t.Parallel()
	callableProducerAdapterRefusal(t, "p22")
}
func TestCallableProducerMethodShorthandRefused(t *testing.T) {
	t.Parallel()
	callableProducerAdapterRefusal(t, "p59")
}
func TestCallableProducerExtraOptionalRefused(t *testing.T) {
	t.Parallel()
	callableProducerAdapterRefusal(t, "p62")
}

func TestCallableProducerFewerParametersAdapted(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, callableProducerSource(t, "p22_adapted"))
}
func TestCallableProducerMethodShorthandAdapted(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, callableProducerSource(t, "p59_adapted"))
}
func TestCallableProducerExtraOptionalAdapted(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, callableProducerSource(t, "p62_adapted"))
}

func TestCallableProducerDiscardedObjectResultRefused(t *testing.T) {
	t.Parallel()
	// Assignability to void alone cannot supply a registry result certificate.
	callableProducerAdapterRefusal(t, "p135_object_result")
}

func TestCallableProducerTaggedUnreadMethod(t *testing.T) {
	t.Parallel()
	// Native must defer the receiver-method obligation behind a tagged selector.
	lowersAndAgreesWithNodeNative(t, callableProducerSource(t, "p135_tagged_unread"))
}
