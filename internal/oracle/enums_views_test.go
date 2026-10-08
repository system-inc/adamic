package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"structured-good", "structured-wrong", "nullable-structured-wrong", "undefined-good", "undefined-wrong", "undefined-null", "undefined-false", "optional-undefined-wrong", "optional-good", "optional-wrong", "optional-child-wrong", "unread"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/fixtures/enum-views/" + name + ".a", true, name != "structured-good" && name != "undefined-good" && name != "optional-good" && name != "unread"})
	}
}

func TestEnumTagSharedViews(t *testing.T) {
	for _, probe := range []struct{ name, stdout, node, message string }{
		{"unread", "narrowed\nafter\n", "narrowed\nafter\n", ""},
		{"structured-good", "narrowed\ntrue\nafter\n", "narrowed\ntrue\nafter\n", ""},
		{"structured-wrong", "narrowed\n", "narrowed\n0\nafter\n", "field read failed: child.ready is not a boolean; expected boolean, found number"},
		{"nullable-structured-wrong", "narrowed\n", "narrowed\n0\nafter\n", "field read failed: child.ready is not a boolean; expected boolean, found number"},
		{"undefined-null", "narrowed\n", "narrowed\nwrong\nundefined\nafter\n", "field read failed: v.brand is not a undefined; expected undefined, found null"},
		{"undefined-false", "narrowed\n", "narrowed\nwrong\nundefined\nafter\n", "field read failed: v.brand is not a undefined; expected undefined, found boolean"},
		{"undefined-wrong", "narrowed\n", "narrowed\nwrong\nundefined\nafter\n", "field read failed: v.brand is not a undefined; expected undefined, found number"},
		{"optional-undefined-wrong", "narrowed\nundefined\n", "narrowed\nundefined\nwrong\nafter\n", "field read failed: v.marker is not a undefined; expected undefined, found number"},
		{"optional-good", "narrowed\ntrue\ntrue\nnarrowed\nfalse\nafter\n", "narrowed\ntrue\ntrue\nnarrowed\nfalse\nafter\n", ""},
		{"optional-wrong", "narrowed\n", "narrowed\n0\n0\nafter\n", "field read failed: v.multiLine is not a boolean | undefined; expected boolean | undefined, found number"},
		{"optional-child-wrong", "narrowed\nfalse\n", "narrowed\nfalse\n0\nafter\n", "field read failed: alias.ready is not a boolean; expected boolean, found number"},
		{"undefined-good", "narrowed\nundefined\nundefined\nnarrowed\nundefined\nundefined\nafter\n", "narrowed\nundefined\nundefined\nnarrowed\nundefined\nundefined\nafter\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enum-views", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: %+v", truth)
			}
			want := run{stdout: []byte(probe.stdout)}
			if probe.message != "" {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.message + "\n")
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %+v", difference, got)
				}
			}
			if probe.name == "structured-wrong" {
				dropNestedView(program)
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(want, got) == "" {
						t.Fatalf("transitive-read mutant must finish cleanly and fail the pinned stop: %+v", got)
					}
				}
				t.Log("transitive-read erasure caught by pinned stop in both backends")
			}
			if probe.message == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
