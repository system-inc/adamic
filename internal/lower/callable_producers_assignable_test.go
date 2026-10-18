package lower

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCallableProducerLiteralReturn(t *testing.T) {
	t.Parallel()
	// Native confirms that the narrower logical result retains the scalar ABI.
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p23"))
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
	if err == nil || !strings.Contains(err.Error(), "main.a:") || !strings.Contains(err.Error(), "wrap the producer in an arrow") {
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
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p22_adapted"))
}
func TestCallableProducerMethodShorthandAdapted(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p59_adapted"))
}
func TestCallableProducerExtraOptionalAdapted(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p62_adapted"))
}

func TestCallableProducerDiscardedObjectResultRefused(t *testing.T) {
	t.Parallel()
	// Assignability to void alone cannot supply a registry result certificate.
	callableProducerAdapterRefusal(t, "p135_object_result")
}

func TestCallableProducerTaggedUnreadMethod(t *testing.T) {
	t.Parallel()
	// Native must defer the receiver-method obligation behind a tagged selector.
	lowersAndAgreesWithNode(t, callableProducerSource(t, "p135_tagged_unread"))
}

func TestCallableProducerAmbiguousRefused(t *testing.T) {
	t.Parallel()
	source := `interface Root { readonly kind: "Holder" }
 interface View extends Root { readonly value: ((v: number) => number) | ((v: string) => number) }
 const value = (v: number | string): number => 1;
 const raw = { kind: "Holder" as const, value };
 const base: Root = raw;
 const view = base as View;
 console.log("" + (view.value !== undefined));
 `
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "assignable to more than one union member") || !strings.Contains(err.Error(), "wrap the producer in an arrow") || !strings.Contains(err.Error(), "main.a:") {
		t.Fatalf("want located ambiguous-producer refusal, got %v", err)
	}
}

func TestCallableProducerUniqueObjectResultRefused(t *testing.T) {
	t.Parallel()
	callableProducerAdapterRefusal(t, "p135_unique_object_result")
}

func TestCallableProducerAmbiguousCertificate(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNode(t, `interface Root {readonly kind:'Holder'}
 interface View extends Root {readonly value:((v:1)=>number)|((v:2)=>number)}
 const raw={kind:'Holder' as const,value:(v:number):number=>1};
 const base:Root=raw;
 const view=base as View;
 console.log(view.kind);`)
	seen := false
	for _, contract := range program.ViewContracts {
		if !contract.ProducerCertified {
			continue
		}
		seen = true
		if len(contract.Functions) != 0 {
			t.Fatal("ambiguous producer was certified")
		}
	}
	if !seen {
		t.Fatal("missing callable certificate control")
	}
}
