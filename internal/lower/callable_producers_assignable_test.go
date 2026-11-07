package lower

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallableProducerLiteralReturn(test *testing.T) {
	test.Parallel()
	// Native confirms that the narrower logical result retains the scalar ABI.
	source, extension := callableProducerSource(test, "p23.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

// In .a each of these casts tests a tag every source value has, so the cast is refused (#q97a1rr)
// before the producer is reached. Its .ts twin, where the cast stays a checked view, reaches the
// producer and stops there as .a did before: NotYet when pending, else Refused.
func callableProducerAdapterRefusal(test *testing.T, name string, pending bool) {
	test.Helper()
	path, absolutePathError := filepath.Abs("testdata/callable_producers/" + name + ".a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := runAgreementNode(test, path)
	if want.code != 0 || string(want.stdout) != "true\n" {
		test.Fatalf("Node control: %+v", want)
	}
	source, extension := callableProducerSource(test, name+".a")
	_, lowerSourceAsError := lowerSourceAs(test, source, extension)
	if lowerSourceAsError == nil ||
		!strings.Contains(lowerSourceAsError.Error(), "main.a:") ||
		!strings.Contains(lowerSourceAsError.Error(), "whose tag test is vacuous") ||
		!strings.Contains(lowerSourceAsError.Error(), "a tag of its own") {
		test.Fatalf("want located vacuous downcast refusal, got %v", lowerSourceAsError)
	}
	test.Log(lowerSourceAsError)
	_, lowerSourceAsError = lowerSourceAs(test, source, ".ts")
	var notYet *NotYet
	var refused *Refused
	if lowerSourceAsError == nil ||
		!strings.Contains(lowerSourceAsError.Error(), "main.ts:") ||
		!strings.Contains(lowerSourceAsError.Error(), "wrap the producer in an arrow") ||
		pending && !errors.As(lowerSourceAsError, &notYet) ||
		!pending && !errors.As(lowerSourceAsError, &refused) {
		test.Fatalf("twin: want located producer adapter refusal (pending %v), got %v", pending, lowerSourceAsError)
	}
	test.Log(lowerSourceAsError)
}
func TestCallableProducerFewerParametersRefused(test *testing.T) {
	test.Parallel()
	callableProducerAdapterRefusal(test, "p22", false)
}
func TestCallableProducerMethodShorthandRefused(test *testing.T) {
	test.Parallel()
	callableProducerAdapterRefusal(test, "p59", true)
}
func TestCallableProducerExtraOptionalRefused(test *testing.T) {
	test.Parallel()
	callableProducerAdapterRefusal(test, "p62", true)
}

func TestCallableProducerFewerParametersAdapted(test *testing.T) {
	test.Parallel()
	source, extension := callableProducerSource(test, "p22_adapted.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}
func TestCallableProducerMethodShorthandAdapted(test *testing.T) {
	test.Parallel()
	source, extension := callableProducerSource(test, "p59_adapted.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}
func TestCallableProducerExtraOptionalAdapted(test *testing.T) {
	test.Parallel()
	source, extension := callableProducerSource(test, "p62_adapted.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestCallableProducerDiscardedObjectResultRefused(test *testing.T) {
	test.Parallel()
	// Assignability to void alone cannot supply a registry result certificate.
	callableProducerAdapterRefusal(test, "p135_object_result", false)
}

func TestCallableProducerTaggedUnreadMethod(test *testing.T) {
	test.Parallel()
	// Native must defer the receiver-method obligation behind a tagged selector.
	source, extension := callableProducerSource(test, "p135_tagged_unread.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestCallableProducerAmbiguousRefused(test *testing.T) {
	test.Parallel()
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
	_, lowerSourceError := lowerSource(test, source)
	if lowerSourceError == nil ||
		!strings.Contains(lowerSourceError.Error(), "a downcast to View whose tag test is vacuous") ||
		!strings.Contains(lowerSourceError.Error(), "give View a tag of its own") ||
		!strings.Contains(lowerSourceError.Error(), "main.a:") {
		test.Fatalf("want located vacuous downcast refusal, got %v", lowerSourceError)
	}
	// The .ts twin keeps the cast a checked view, so the ambiguous producer is reached and refused.
	_, lowerSourceAsError := lowerSourceAs(test, source, ".ts")
	if lowerSourceAsError == nil ||
		!strings.Contains(lowerSourceAsError.Error(), "assignable to more than one union member") ||
		!strings.Contains(lowerSourceAsError.Error(), "wrap the producer in an arrow") ||
		!strings.Contains(lowerSourceAsError.Error(), "main.ts:") {
		test.Fatalf("twin: want located ambiguous-producer refusal, got %v", lowerSourceAsError)
	}
}

// A producer whose type is identical to one member is carried by that member,
// even when it is also assignable to another (wide.ts is the same with the
// identical member second).
func TestCallableProducerExactMemberCarries(test *testing.T) {
	test.Parallel()
	source, extension := callableProducerSource(test, "p135_exact_member.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestCallableProducerUniqueObjectResultRefused(test *testing.T) {
	test.Parallel()
	callableProducerAdapterRefusal(test, "p135_unique_object_result", true)
}

func TestCallableProducerAmbiguousCertificate(test *testing.T) {
	test.Parallel()
	// .ts, where the cast stays a checked view: in .a its tag test every Root passes is refused (#q97a1rr).
	program := lowersAndAgreesWithNodeAs(test, `interface Root {readonly kind:'Holder'}
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
			test.Fatal("ambiguous producer was certified")
		}
	}
	if !seen {
		test.Fatal("missing callable certificate control")
	}
}
