package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// V4 replaces these old review refusals with the same checked-call boundary
// used by lane 5. Keep the original .a programs, and observe Node separately:
// erased casts print undefined, whereas a number cannot enter a string producer.
func v4ReviewProducerArgument(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_v4_checked", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "undefined\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	checks := func(got run) bool {
		return got.exitCode == 70 && len(got.stdout) == 0 && strings.Contains(string(got.stderr), "callable call failed:") && strings.Contains(string(got.stderr), "argument 1 expected producer string, view number") && !strings.Contains(string(got.stderr), "Sanitizer")
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"sanitized": sanitized, "native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if !checks(got) {
			t.Fatalf("%s checked call: %#v", backend, got)
		}
		t.Logf("%s checked call: exit=%d stdout=%q stderr=%q; Node stdout=%q", backend, got.exitCode, got.stdout, got.stderr, truth.stdout)
	}
	changed := 0
	for _, function := range program.Functions {
		for _, id := range function.CallableParameters {
			if id != 0 && program.ViewContracts[id-1].Name == "string" {
				program.ViewContracts[id-1] = ir.ViewContract{Kind: ir.ViewUnknown, Of: ir.Union, Name: "unknown"}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("producer-domain mutant changed %d contracts", changed)
	}
	mutant, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"sanitized": mutant, "native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if checks(got) {
			t.Fatalf("%s producer-domain omission survived", backend)
		}
		t.Logf("%s producer-domain omission caught: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
	}
}

func TestV4ReviewMapValueCallChecksProducerArgument(t *testing.T) {
	t.Parallel()
	v4ReviewProducerArgument(t, "fxspptb_oct9_views_p75_callable_from_map_value")
}

func TestV4ReviewElementCallPreservesCompatibleProducer(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/review/refused/fxspptb_oct9_views_p17_callable_field_element_call.a")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.SplitN(string(source), "\n", 2)[1]
	before := "(value: string): number => value.length"
	if strings.Count(body, before) != 1 {
		t.Fatal("review producer changed")
	}
	body = strings.Replace(body, before, "(value: number): number => value", 1)
	v4DirectSource(t, body, "41\n", "")
}

func v4ReviewUnionMembers(t *testing.T, name string, reads int) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_v4_checked", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "true\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"sanitized": sanitized, "native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "matches no member of Target") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s union member check: %#v", backend, got)
		}
		t.Logf("%s union member check: exit=%d stderr=%q; Node stdout=%q", backend, got.exitCode, got.stderr, truth.stdout)
	}
	changed := 0
	omit := func(value ir.Expression) ir.Expression {
		if property, ok := value.(ir.Property); ok && property.Name == "value" && property.View != "" {
			property.View, property.ViewType = "", ""
			property.ViewContract, property.ViewTypeID = 0, 0
			property.ViewOrdinary = true
			changed++
			return property
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
	if changed != reads {
		t.Fatalf("union member omission changed %d reads, want %d", changed, reads)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Log("union member check omission caught: all backends instead print Node's true, sanitizer-clean")
}

func TestV4ReviewElementReadChecksUnionMembers(t *testing.T) {
	t.Parallel()
	v4ReviewUnionMembers(t, "fxspptb_oct9_views_p07_element_access_union_wrong", 2)
}

func TestV4ReviewDestructuredReadChecksUnionMembers(t *testing.T) {
	t.Parallel()
	v4ReviewUnionMembers(t, "fxspptb_oct9_views_p49_union_destructure_wrong", 1)
}

func TestV4ReviewForeignElementCallChecksProducerArgument(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/review/refused/fxspptb_oct9_views_p17_callable_field_element_call.a")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The existing lane-5 frontend adapter exercises foreign TypeScript: unlike
	// .a's proof-required escape, this path must retain its checked adapter.
	body := strings.SplitN(string(source), "\n", 2)[1]
	v4Lane5Negative(t, body, "undefined\n", "argument 1", "view['compute']")
}
