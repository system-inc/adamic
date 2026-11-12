package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectorBareLetInitializationAgrees(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/inspector_bare_let.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	capturedUnionRuns(t, path, program, "undefined\nfalse\nReferenceError\nundefined\n")
}

func TestInspectorDoomedLetAgrees(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/inspector_doomed_let.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	capturedUnionRuns(t, path, program, "disabled\n")
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if why := disagreement(run{stdout: []byte("disabled\n")}, onWASI(t, native.C(program))); why != "" {
			t.Fatal(why)
		}
	}
}

func TestInspectorLetNonNullCheckPrecedesTrap(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/inspector_doomed/non_null_let.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	message := ""
	for _, s := range program.Strings {
		if strings.Contains(s, "non-null") {
			message = s
		}
	}
	if message == "" {
		t.Fatal("missing non-null check")
	}
	want := run{stderr: []byte("adamic: panic: " + message + "\n"), exitCode: 70}
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onSliceJavaScript(t, program)} {
		if why := disagreement(want, got); why != "" {
			t.Errorf("%s: %s: stdout=%q stderr=%q exit=%d", name, why, got.stdout, got.stderr, got.exitCode)
		}
	}
	t.Log(message)
}

func TestInspectorLetCountsRecords(t *testing.T) {
	for _, name := range []string{"inspector_bare_let.a", "inspector_doomed_let.a"} {
		key := "internal/oracle/testdata/" + name
		_, measure := counted(t, key, false, nil, false, false)
		row := measure()
		if *updateCounts {
			measuredCounts.Store(key, row)
		} else {
			regenerateCountRecord(t, "runtime", key, row)
		}
		t.Log(row)
	}
}
