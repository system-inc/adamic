package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func unsupportedViewSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("testdata/unsupported_view/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func assertUnsupportedView(t *testing.T, source, member, typeName, where string) {
	t.Helper()
	program, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want creation-site refusal, got program %v, error %v", program != nil, err)
	}
	if refused.What != "view type has an unsupported member: "+member || !strings.HasSuffix(refused.Where, where) || !strings.Contains(refused.Fix, typeName) || program != nil {
		t.Fatalf("wrong refusal: %#v; program present: %v", refused, program != nil)
	}
}

func TestUnsupportedViewMapMember(t *testing.T) {
	t.Parallel()
	assertUnsupportedView(t, unsupportedViewSource(t, "p37"), "extra", "Map<string, number>", "main.a:8:14")
}

func TestUnsupportedViewCallableMember(t *testing.T) {
	t.Parallel()
	assertUnsupportedView(t, unsupportedViewSource(t, "p21"), "value", "Target", "main.a:11:14")
}

func TestUnsupportedViewUnreadMember(t *testing.T) {
	t.Parallel()
	source := unsupportedViewSource(t, "p37")
	assertUnsupportedView(t, strings.Split(source, "console.log")[0], "extra", "Map<string, number>", "main.a:8:14")
}

func TestSupportedViewMembersAgree(t *testing.T) {
	t.Parallel()
	source := unsupportedViewSource(t, "control")
	lowersAndAgreesWithNode(t, source)
	// Native slot representation is the failure that the original probes expose.
	lowersAndAgreesWithNodeNative(t, source)
}
