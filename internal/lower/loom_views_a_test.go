package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loomViewFixture returns the fixture file's source and its extension, the mode it lowers in.
func loomViewFixture(test *testing.T, file string) (string, string) {
	test.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "loom_views_a", file))
	if err != nil {
		test.Fatal(err)
	}
	return string(source), filepath.Ext(file)
}

// requireLoomViewTwinRefusal lowers the fixture's source as main.ts, where its cast stays a checked
// view, so the read it was written about is reached and refused as .a refused it before #q97a1rr.
func requireLoomViewTwinRefusal(test *testing.T, name, diagnostic, fix string) {
	test.Helper()
	source, _ := loomViewFixture(test, name+".a")
	_, err := lowerSourceAs(test, source, ".ts")
	if err == nil ||
		!strings.Contains(err.Error(), diagnostic) ||
		!strings.Contains(err.Error(), fix) ||
		!strings.Contains(err.Error(), ".ts:") {
		test.Fatalf("twin: want located refusal %q and fix %q, got %v", diagnostic, fix, err)
	}
}

func requireLoomViewRefusal(test *testing.T, name, output, diagnostic, fix string) {
	test.Helper()
	source, extension := loomViewFixture(test, name+".a")
	truth := runAgreementNode(test, writeAgreementSource(test, "main"+extension, source))
	if truth.code != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.code, truth.stdout, truth.stderr)
	}
	_, err := lowerSourceAs(test, source, extension)
	if err == nil ||
		!strings.Contains(err.Error(), diagnostic) ||
		!strings.Contains(err.Error(), fix) ||
		!strings.Contains(err.Error(), ".a:") {
		test.Fatalf("want located refusal %q and fix %q, got %v", diagnostic, fix, err)
	}
}

// In .a the cast of key, views_p01, p05, p06, p19, p72, native_p37 and union_key tests a tag every
// source value passes, so the cast is refused (#q97a1rr) before the read is reached; each one's .ts
// twin pins the read's own refusal.
func TestLoomViewKeyRead(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"key",
		"NaN\nfalse\n",
		"a downcast to Holder whose tag test is vacuous",
		"give Holder a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "key", "view member read by element access is not yet checked", "use view.count")
}

// In .a unsupported's cast is an untagged view whose checked reads no proof erases, so the cast is refused
// (#m8dvvmf) before the read is reached; its .ts twin pins the read's own refusal.
func TestLoomViewUnsupportedRead(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"unsupported",
		"xx / 0\n",
		"a downcast from Base to View in .a whose checks no proof erases",
		"move the file to .ts",
	)
	requireLoomViewTwinRefusal(test, "unsupported", "checked view read of field", "contract before reading this field")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewUnsupportedCallableRead(test *testing.T) {
	test.Parallel()
	source, extension := loomViewFixture(test, "callable.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

// key.a's fix lowers as .ts, where the cast stays a checked view: in .a its tag test every source
// value passes is refused (#q97a1rr).
func TestLoomViewKeyFix(test *testing.T) {
	test.Parallel()
	source, _ := loomViewFixture(test, "key.a")
	source = strings.ReplaceAll(source, "view['count']", "view.count")
	source = strings.Replace(source, "count: 'seven'", "count: 7", 1)
	lowersAndAgreesWithNodeAs(test, source, ".ts")
}

// unsupported.a without its read lowers as .ts, where the cast stays a checked view: in .a a downcast
// whose checks no proof erases is refused (#m8dvvmf).
func TestLoomViewUnsupportedUnreadControl(test *testing.T) {
	test.Parallel()
	source, _ := loomViewFixture(test, "unsupported.a")
	source = strings.Replace(source, "console.log(`${view.label} / ${view.count}`);", "console.log('created');", 1)
	lowersAndAgreesWithNodeAs(test, source, ".ts")
}

func TestLoomViewOriginalViewsP01(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"views_p01",
		"undefined\n",
		"a downcast to Holder whose tag test is vacuous",
		"give Holder a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "views_p01", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP05(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"views_p05",
		"NaN\nfalse\n",
		"a downcast to Holder whose tag test is vacuous",
		"give Holder a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "views_p05", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP06(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"views_p06",
		"2\n",
		"a downcast to Holder whose tag test is vacuous",
		"give Holder a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "views_p06", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP19(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"views_p19",
		"undefined\n",
		"a downcast to Receiver whose tag test is vacuous",
		"give Receiver a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "views_p19", "view member read by element access is not yet checked", "use view.")
}

func TestLoomViewOriginalViewsP72(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"views_p72",
		"number 2\n",
		"a downcast to View whose tag test is vacuous",
		"give View a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "views_p72", "view member read by element access is not yet checked", "use view.")
}

// Views D checks this read against its contract, so it compiles and prints what Node prints.
func TestLoomViewOriginalNativeP21(test *testing.T) {
	test.Parallel()
	source, extension := loomViewFixture(test, "native_p21.ts")
	lowersAndAgreesWithNodeAs(test, source, extension)
}

func TestLoomViewOriginalNativeP37(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"native_p37",
		"xx\n0\n",
		"a downcast to OtherView whose tag test is vacuous",
		"give OtherView a tag of its own",
	)
	requireLoomViewTwinRefusal(test, "native_p37", "checked view read of field", "contract before reading this field")
}

func TestLoomViewUnionKeyRead(test *testing.T) {
	test.Parallel()
	requireLoomViewRefusal(
		test,
		"union_key",
		"true\n",
		"a downcast to Holder whose tag test is vacuous",
		"give Holder a tag of its own",
	)
	requireLoomViewTwinRefusal(
		test,
		"union_key",
		"view member read by element access is not yet checked",
		"use view.value",
	)
}
