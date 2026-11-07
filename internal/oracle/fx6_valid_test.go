package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p38_number.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p35.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p36.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p37.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p38.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p51.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p53.ts", true, false})

}

func TestOracleViewsViewedProducerMutation(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p35.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedMutableField(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p36.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedUnionField(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p37.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedOptionalStringField(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p38.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsDestructuredScalarUnion(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p51.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsDestructuredCallableUnion(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p53.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{
		"native":     releasedUncached(test, program),
		"sanitized":  func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript": onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

// In .a the cast before the destructuring is refused first: its tag test, kind === 'Holder', is one every source
// value passes (#q97a1rr). Its .ts twin keeps the cast a checked view and is refused at the destructuring.
func TestOracleViewsDestructuredInvalidObjectUnionRefuses(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../lower/testdata/fx6_refused/p49.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source := onNode(test, path)
	if source.exitCode != 0 || string(source.stdout) != "true\n" {
		test.Fatalf("Node control: %+v", source)
	}
	_, loweredError := lowered(test, path)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		refused.What != `a downcast to Holder whose tag test is vacuous: every { readonly kind: "Holder"; } has kind "Holder"` ||
		refused.Fix != "give Holder a tag of its own in a discriminated union, or build the value as a Holder so the compiler can prove it" ||
		!strings.Contains(refused.Where, "p49.a:9:14") {
		test.Fatalf("want the located vacuous downcast refusal with fix at 9:14, got %v", loweredError)
	}
	_, loweredError = lowered(test, typeScriptTwin(test, path))
	if !errors.As(loweredError, &refused) ||
		refused.What != "a destructured object-union view with an incompatible literal producer" ||
		refused.Fix != "build a member of Target before destructuring field value" ||
		!strings.Contains(refused.Where, "p49.ts:") {
		test.Fatalf("twin: want located incompatible-producer refusal with fix, got %v", loweredError)
	}
}

// The stored function takes a string and no member of Target does, so the destructured read is refused where it
// stands, as p49's incompatible literal producer is, rather than compiled to stop at run time where Node prints 5
// (Compiler's ruling (b) on #aecx10a). In .a the cast before it is refused first: its tag test, kind === 'Receiver',
// is one every source value passes (#q97a1rr); its .ts twin keeps the cast a checked view and is refused at the read.
func TestOracleViewsDestructuredInvalidCallableRefuses(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../internal/lower/testdata/fx6/p51_wrong.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if source := onNode(test, path); source.exitCode != 0 {
		test.Fatalf("Node control failed: %#v", source)
	}
	_, loweredError := lowered(test, path)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		refused.What != `a downcast to Receiver whose tag test is vacuous: every { readonly kind: "Receiver"; } has kind "Receiver"` ||
		refused.Fix != "give Receiver a tag of its own in a discriminated union, or build the value as a Receiver so the compiler can prove it" ||
		!strings.Contains(refused.Where, "p51_wrong.a:6:14") {
		test.Fatalf("want the located vacuous downcast refusal at 6:14, got %v", loweredError)
	}
	_, loweredError = lowered(test, typeScriptTwin(test, path))
	if !errors.As(loweredError, &refused) ||
		refused.What != "checked view read of field value with unsupported representation conversion contract" ||
		refused.Fix != "prove or implement the representation conversion contract before reading this field" ||
		!strings.Contains(refused.Where, "p51_wrong.ts:7:9") {
		test.Fatalf("twin: want the located representation conversion refusal at 7:9, got %v", loweredError)
	}
}
