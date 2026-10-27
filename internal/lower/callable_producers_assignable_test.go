package lower

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCallableProducerLiteralReturn(t *testing.T) {
	t.Parallel()
	// Native confirms that the narrower logical result retains the scalar ABI.
	source, extension := callableProducerSource(t, "p23.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

// In .a each of these casts tests a tag every source value has, so the cast is refused (#q97a1rr)
// before the producer is reached.
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
	source, extension := callableProducerSource(t, name+".a")
	_, err = lowerSourceAs(t, source, extension)
	if err == nil || !strings.Contains(err.Error(), "main.a:") || !strings.Contains(err.Error(), "whose tag test is vacuous") || !strings.Contains(err.Error(), "a tag of its own") {
		t.Fatalf("want located vacuous downcast refusal, got %v", err)
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
	source, extension := callableProducerSource(t, "p22_adapted.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}
func TestCallableProducerMethodShorthandAdapted(t *testing.T) {
	t.Parallel()
	source, extension := callableProducerSource(t, "p59_adapted.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}
func TestCallableProducerExtraOptionalAdapted(t *testing.T) {
	t.Parallel()
	source, extension := callableProducerSource(t, "p62_adapted.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestCallableProducerDiscardedObjectResultRefused(t *testing.T) {
	t.Parallel()
	// Assignability to void alone cannot supply a registry result certificate.
	callableProducerAdapterRefusal(t, "p135_object_result")
}

func TestCallableProducerTaggedUnreadMethod(t *testing.T) {
	t.Parallel()
	// Native must defer the receiver-method obligation behind a tagged selector.
	source, extension := callableProducerSource(t, "p135_tagged_unread.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
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
	// In .a the cast's tag test, kind === "Holder", is one every Root passes, so the cast is refused
	// (#q97a1rr) before the producer is reached.
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "a downcast to View whose tag test is vacuous") || !strings.Contains(err.Error(), "give View a tag of its own") || !strings.Contains(err.Error(), "main.a:") {
		t.Fatalf("want located vacuous downcast refusal, got %v", err)
	}
}

// A producer whose type is identical to one member is carried by that member,
// even when it is also assignable to another (wide.ts is the same with the
// identical member second).
func TestCallableProducerExactMemberCarries(t *testing.T) {
	t.Parallel()
	source, extension := callableProducerSource(t, "p135_exact_member.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

func TestCallableProducerUniqueObjectResultRefused(t *testing.T) {
	t.Parallel()
	callableProducerAdapterRefusal(t, "p135_unique_object_result")
}

func TestCallableProducerAmbiguousCertificate(t *testing.T) {
	t.Parallel()
	// .ts, where the cast stays a checked view: in .a its tag test every Root passes is refused (#q97a1rr).
	program := lowersAndAgreesWithNodeAs(t, `interface Root {readonly kind:'Holder'}
 interface View extends Root {readonly value:((v:1)=>number)|((v:2)=>number)}
 const raw={kind:'Holder' as const,value:(v:number):number=>1};
 const base:Root=raw;
 const view=base as View;
 console.log(view.kind);`, ".ts")
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
