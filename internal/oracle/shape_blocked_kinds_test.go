package oracle

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestShapeRecordedBlockedKinds(t *testing.T) {
	program, path := interfaceFixture(t, "lane3/rechecked-blocked-kinds")
	reference := onNode(t, path)
	if reference.exitCode != 0 || string(reference.stdout) != "1\nfalse\n3,1\n3,1\ntrue\n" {
		t.Fatalf("Node did not exercise every recorded kind: %+v", reference)
	}
	sanitized, binary := nativelyUncached(t, program)
	for _, actual := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if why := disagreement(reference, actual); why != "" {
			t.Fatalf("recorded kind: %s: %+v", why, actual)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	mutations := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "f" {
			continue
		}
		for i, statement := range function.Body {
			returned, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			returned.Value = ir.NumberConstant{Value: 2}
			function.Body[i] = returned
			mutations++
		}
	}
	if mutations != 1 {
		t.Fatalf("wrong-overload-result requires one numeric return, got %d", mutations)
	}
	for _, actual := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if actual.exitCode != 0 || !strings.HasPrefix(string(actual.stdout), "2\n") || disagreement(reference, actual) == "" {
			t.Fatalf("valid wrong-overload-result mutant escaped: %+v", actual)
		}
		t.Logf("wrong-overload-result caught by Node stdout comparison: exit=%d stdout=%q", actual.exitCode, actual.stdout)
	}
}

func TestShapeRecordedOptionalKinds(t *testing.T) {
	program, path := interfaceFixture(t, "lane3/rechecked-optional-kinds")
	reference := onNode(t, path)
	if reference.exitCode != 0 || string(reference.stdout) != "1\n1\n" {
		t.Fatalf("Node did not exercise both recorded kinds: %+v", reference)
	}
	sanitized, binary := nativelyUncached(t, program)
	for _, actual := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if why := disagreement(reference, actual); why != "" {
			t.Fatalf("standalone optional kind: %s: %+v", why, actual)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	mutations := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "optionalRun" {
			continue
		}
		for i, statement := range function.Body {
			returned, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			returned.Value = ir.NumberConstant{Value: 2}
			function.Body[i] = returned
			mutations++
		}
	}
	if mutations != 1 {
		t.Fatalf("wrong-optional-result requires one numeric return, got %d", mutations)
	}
	for _, actual := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if actual.exitCode != 0 || string(actual.stdout) != "2\n1\n" || disagreement(reference, actual) == "" {
			t.Fatalf("valid wrong-optional-result mutant escaped: %+v", actual)
		}
		t.Logf("wrong-optional-result caught by Node stdout comparison: exit=%d stdout=%q", actual.exitCode, actual.stdout)
	}
}
