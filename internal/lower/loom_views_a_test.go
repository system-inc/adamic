package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loomViewFixture returns the fixture file's source and its extension, the mode it lowers in.
func loomViewFixture(t *testing.T, file string) (string, string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "loom_views_a", file))
	if err != nil {
		t.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

// requireLoomViewTwinRefusal lowers the fixture's source as main.ts, where its cast stays a checked
// view, so the read it was written about is reached and refused as .a refused it before #q97a1rr.
func requireLoomViewTwinRefusal(t *testing.T, name, diagnostic, fix string) {
	t.Helper()
	source, _ := loomViewFixture(t, name+".a")
	_, err := lowerSourceAs(t, source, ".ts")
	if err == nil || !strings.Contains(err.Error(), diagnostic) || !strings.Contains(err.Error(), fix) || !strings.Contains(err.Error(), ".ts:") {
		t.Fatalf("twin: want located refusal %q and fix %q, got %v", diagnostic, fix, err)
	}
}

// requireLoomViewUnbuiltTwins lowers the fixture's source as main.ts, where an unbuilt field is the
// field's alone (#rgrtt32): as written it agrees with Node; with its honest count: 0 swapped for a
// value the view's type doesn't hold it stops at the read on both backends, printed what Node printed
// before it; and with its read line swapped for read, a read of the unbuilt field, it's NotYet at
// at, in the unbuilt field's words.
func requireLoomViewUnbuiltTwins(t *testing.T, name, line, printed, read, field, at string) {
	t.Helper()
	source, _ := loomViewFixture(t, name+".a")
	lowersAndAgreesWithNodeAs(t, source, ".ts")
	overloadCheckStops(t, strings.Replace(source, "count: 0", "count: true", 1), printed, "view.count")
	_, err := lowerSourceAs(t, strings.Replace(source, line, read, 1), ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, "main.ts:"+at) || !strings.Contains(notYet.What, "field "+field+" whose contract isn't built") {
		t.Fatalf("reading %s: want a NotYet at main.ts:%s naming the field, got %v", field, at, err)
	}
}

func requireLoomViewRefusal(t *testing.T, name, output, diagnostic, fix string) {
	t.Helper()
	source, extension := loomViewFixture(t, name+".a")
	truth := runAgreementNode(t, writeAgreementSource(t, "main"+extension, source))
	if truth.code != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.code, truth.stdout, truth.stderr)
	}
	_, err := lowerSourceAs(t, source, extension)
	if err == nil || !strings.Contains(err.Error(), diagnostic) || !strings.Contains(err.Error(), fix) || !strings.Contains(err.Error(), ".a:") {
		t.Fatalf("want located refusal %q and fix %q, got %v", diagnostic, fix, err)
	}
}

// In .a the cast of key, views_p01, p05, p06, p19, p72, native_p37 and union_key tests a tag every
// source value passes, so the cast is refused (#q97a1rr) before the read is reached; each one's .ts
// twin pins the read's own refusal.
func TestLoomViewKeyRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "key", "NaN\nfalse\n", "a downcast to Holder whose tag test is vacuous", "give Holder a tag of its own")
	requireLoomViewTwinRefusal(t, "key", "view member read by element access is not yet checked", "use view.count")
}

// In .a unsupported's cast is an untagged view whose checked reads no proof erases, so the cast is refused
// (#m8dvvmf) before the read is reached. Its .ts twin lowers since #rgrtt32: extra's unbuilt contract is
// the field's alone, and the read of its sibling count is checked, so a lie in count stops at the read and
// a read of extra is NotYet.
func TestLoomViewUnsupportedRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "unsupported", "xx / 0\n", "a downcast from Base to View in .a whose checks no proof erases", "move the file to .ts")
	requireLoomViewUnbuiltTwins(t, "unsupported", "console.log(`${view.label} / ${view.count}`);", "", "console.log(String(view.extra));", "extra", "10:20")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewUnsupportedCallableRead(t *testing.T) {
	t.Parallel()
	source, extension := loomViewFixture(t, "callable.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

// key.a's fix lowers as .ts, where the cast stays a checked view: in .a its tag test every source
// value passes is refused (#q97a1rr).
func TestLoomViewKeyFix(t *testing.T) {
	t.Parallel()
	source, _ := loomViewFixture(t, "key.a")
	source = strings.ReplaceAll(source, "view['count']", "view.count")
	source = strings.Replace(source, "count: 'seven'", "count: 7", 1)
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// unsupported.a without its read lowers as .ts, where the cast stays a checked view: in .a a downcast
// whose checks no proof erases is refused (#m8dvvmf).
func TestLoomViewUnsupportedUnreadControl(t *testing.T) {
	t.Parallel()
	source, _ := loomViewFixture(t, "unsupported.a")
	source = strings.Replace(source, "console.log(`${view.label} / ${view.count}`);", "console.log('created');", 1)
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

func TestLoomViewOriginalViewsP01(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p01", "undefined\n", "a downcast to Holder whose tag test is vacuous", "give Holder a tag of its own")
	requireLoomViewTwinRefusal(t, "views_p01", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP05(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p05", "NaN\nfalse\n", "a downcast to Holder whose tag test is vacuous", "give Holder a tag of its own")
	requireLoomViewTwinRefusal(t, "views_p05", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP06(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p06", "2\n", "a downcast to Holder whose tag test is vacuous", "give Holder a tag of its own")
	requireLoomViewTwinRefusal(t, "views_p06", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP19(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p19", "undefined\n", "a downcast to Receiver whose tag test is vacuous", "give Receiver a tag of its own")
	requireLoomViewTwinRefusal(t, "views_p19", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP72(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "views_p72", "number 2\n", "a downcast to View whose tag test is vacuous", "give View a tag of its own")
	requireLoomViewTwinRefusal(t, "views_p72", "view member read by element access is not yet checked", "use view.")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewOriginalNativeP21(t *testing.T) {
	t.Parallel()
	source, extension := loomViewFixture(t, "native_p21.ts")
	lowersAndAgreesWithNodeAs(t, source, extension)
}

// native_p37's .ts twin lowers since #rgrtt32 as unsupported's does: value's unbuilt contract is the
// field's alone, and count's read is checked.
func TestLoomViewOriginalNativeP37(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "native_p37", "xx\n0\n", "a downcast to OtherView whose tag test is vacuous", "give OtherView a tag of its own")
	requireLoomViewUnbuiltTwins(t, "native_p37", "console.log(`${view.count}`);", "xx\n", "console.log(String(view.value));", "value", "12:20")
}

func TestLoomViewUnionKeyRead(t *testing.T) {
	t.Parallel()
	requireLoomViewRefusal(t, "union_key", "true\n", "a downcast to Holder whose tag test is vacuous", "give Holder a tag of its own")
	requireLoomViewTwinRefusal(t, "union_key", "view member read by element access is not yet checked", "use view.value")
}
