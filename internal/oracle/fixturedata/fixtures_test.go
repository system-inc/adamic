package fixturedata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sourceAndOptions(t *testing.T, root, name, options string) string {
	t.Helper()
	source := filepath.Join(root, "internal/oracle/testdata", name)
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("console.log('fixture');\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := source + ".oracle.json"
	if err := os.WriteFile(path, []byte(options), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDiscoveryCarriesFixtureOptions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourceAndOptions(t, root, "nested/check.a", `{"lowers":true,"checked":true,"input":false,"uncounted":true}`)
	sourceAndOptions(t, root, "not_yet.a", `{"lowers":false,"checked":false,"input":false}`)
	sourceAndOptions(t, root, "input.a", `{"lowers":true,"checked":false,"input":true,"argumentsHex":["","61ff62"],"unreadable":true,"writes":true}`)
	// A helper module has no sidecar and must never become another entry point.
	if err := os.WriteFile(filepath.Join(root, "internal/oracle/testdata/helper.a"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fixtures, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 3 {
		t.Fatalf("want 3 entry points, got %d", len(fixtures))
	}
	input, check, notYet := fixtures[0], fixtures[1], fixtures[2]
	if !check.Lowers || !check.Checked || !check.Uncounted || check.Input {
		t.Fatalf("check options lost: %+v", check)
	}
	if notYet.Lowers || notYet.Checked || notYet.Input {
		t.Fatalf("not-yet options lost: %+v", notYet)
	}
	if !input.Input || !input.Unreadable || !input.Writes || len(input.Arguments) != 2 || input.Arguments[0] != "" || input.Arguments[1] != "a\xffb" {
		t.Fatalf("input options lost: %+v", input)
	}
}

func TestDiscoveryRejectsBrokenOptions(t *testing.T) {
	t.Parallel()
	for _, given := range []string{
		`{"lowers":true,"checked":false}`, // Missing input must not silently become false.
		`{"lowers":true,"checked":false,"input":false,"cheked":true}`,
		`{"lowers":true,"checked":false,"input":false} {}`,
		`{"lowers":true,"checked":false,"input":false,"path":"../escape.a"}`,
		`{"lowers":true,"checked":false,"input":false,"writes":true}`,
		`{"lowers":true,"checked":true,"input":true}`,
		`{"lowers":true,"checked":false,"input":true,"argumentsHex":["gg"]}`,
	} {
		root := t.TempDir()
		sourceAndOptions(t, root, "bad.a", given)
		if _, err := Discover(root); err == nil {
			t.Fatalf("accepted broken options: %s", given)
		}
	}
}

func TestCountUpdatesAreIndependent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourceAndOptions(t, root, "one.a", `{"lowers":true,"checked":false,"input":false}`)
	sourceAndOptions(t, root, "two.a", `{"lowers":true,"checked":false,"input":false}`)
	fixtures, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	row := func(f Fixture, n string) string { return "| " + f.Path + " | " + n + " | 0 | 0 | 0 | 0 | 0 |" }
	for _, f := range fixtures {
		if err := CheckCounts(f.CountsPath, f.Path, row(f, "1"), true); err != nil {
			t.Fatal(err)
		}
		// A fixed timestamp detects a rewrite even if its contents stayed identical.
		if err := os.Chtimes(f.CountsPath, time.Unix(100, 0), time.Unix(100, 0)); err != nil {
			t.Fatal(err)
		}
	}
	one, two := fixtures[0], fixtures[1]
	if err := CheckCounts(one.CountsPath, one.Path, row(one, "2"), false); err == nil {
		t.Fatal("changed count was accepted")
	}
	if err := CheckCounts(one.CountsPath, one.Path, row(one, "2"), true); err != nil {
		t.Fatal(err)
	}
	if err := CheckCounts(two.CountsPath, two.Path, row(two, "1"), true); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(two.CountsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(time.Unix(100, 0)) {
		t.Fatal("unchanged counts file was rewritten")
	}
	table, err := RenderCounts(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(table, row(one, "2")+"\n"+row(two, "1")+"\n") {
		t.Fatal("renderer lost a row or a value")
	}
	if err := os.Remove(two.CountsPath); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderCounts(fixtures); err == nil {
		t.Fatal("renderer silently omitted a missing record")
	}
}
