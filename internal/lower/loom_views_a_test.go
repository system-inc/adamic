package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loomViewFixture(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "loom_views_a", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func requireLoomViewRefusal(t *testing.T, name, output, diagnostic, fix string) {
	t.Helper()
	source := loomViewFixture(t, name)
	truth := runAgreementNode(t, writeAgreementSource(t, source))
	if truth.code != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.code, truth.stdout, truth.stderr)
	}
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), diagnostic) || !strings.Contains(err.Error(), fix) || !strings.Contains(err.Error(), ".a:") {
		t.Fatalf("want located refusal %q and fix %q, got %v", diagnostic, fix, err)
	}
}

func TestLoomViewKeyRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "key", "NaN\nfalse\n", "view member read by element access is not yet checked", "use view.count")
}

func TestLoomViewUnsupportedRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "unsupported", "xx / 0\n", "checked view read of field", "contract before reading this field")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewUnsupportedCallableRead(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, loomViewFixture(t, "callable"))
}

func TestLoomViewKeyFix(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(loomViewFixture(t, "key"), "view['count']", "view.count")
	source = strings.Replace(source, "count: 'seven'", "count: 7", 1)
	lowersAndAgreesWithNode(t, source)
}

func TestLoomViewUnsupportedUnreadControl(t *testing.T) {
	t.Parallel()
	source := loomViewFixture(t, "unsupported")
	source = strings.Replace(source, "console.log(`${view.label} / ${view.count}`);", "console.log('created');", 1)
	lowersAndAgreesWithNode(t, source)
}

func TestLoomViewOriginalViewsP01(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p01", "undefined\n", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP05(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p05", "NaN\nfalse\n", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP06(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p06", "2\n", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP19(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p19", "undefined\n", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP72(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p72", "number 2\n", "view member read by element access is not yet checked", "use view.")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewOriginalNativeP21(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, loomViewFixture(t, "native_p21"))
}

func TestLoomViewOriginalNativeP37(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "native_p37", "xx\n0\n", "checked view read of field", "contract before reading this field")
}

func TestLoomViewUnionKeyRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "union_key", "true\n", "view member read by element access is not yet checked", "use view.value")
}
