package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// admitted counts every corpus, so each admitted program is named on stderr with how it was judged, and listed in
// admitted_paths, even one outside the diff: J-1x881sm read admitted = 4 where its caller printed two (#1zscjp1).
// corpus/input.a is outside the diff here, since base and head are one commit, and only the head accepts it.
// Mutants: the stderr line per admitted program dropped, and each row's name is missing; admitted_paths left
// empty, and each row's list is wrong.
func TestAdmittedProgramsAreNamed(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, corpus, native, budget string
		exit                         int
		verdict, judgment            string
	}{
		{"agrees", "witnesses", "42", "0", 0, "pass", "agrees with Node"},
		{"disagrees", "witnesses", "41", "0", 1, "fail", "disagrees with Node"},
		{"omitted", "fixtures", "42", "1", 0, "sampled", "not run on Node, omitted by the budget"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir, sha, base, head := commandFixture(t)
			writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\nc) echo C;;\njs) echo \"console.log('42');\";;\nbuild) printf '#!/bin/sh\\nprintf \""+test.native+"\\\\n\"\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
			blob, err := git(dir, "rev-parse", "HEAD:corpus/input.a")
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(manifest{SHA: sha, Generator: "generator.py", Corpora: []corpus{{Name: test.corpus, Programs: []program{{Path: "corpus/input.a", Blob: blob}}}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json", "--timeout", "2s", "--budget", test.budget)
			var r report
			if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
				t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
			}
			if o.Exit != test.exit || r.Verdict != test.verdict || r.Admitted != 1 || r.DiffCount != 0 {
				t.Fatalf("exit %d, verdict %q, admitted %d, diff %d; want exit %d, %q, 1 admitted outside an empty diff: %s", o.Exit, r.Verdict, r.Admitted, r.DiffCount, test.exit, test.verdict, o.Stderr)
			}
			if !reflect.DeepEqual(r.AdmittedPaths, []string{"corpus/input.a"}) {
				t.Fatalf("admitted_paths %q, want corpus/input.a", r.AdmittedPaths)
			}
			want := "\nadmitted corpus/input.a (" + test.corpus + "): " + test.judgment + "\n"
			if !strings.Contains("\n"+o.Stderr, "\nadmitted: 1\n") || !strings.Contains("\n"+o.Stderr, want) {
				t.Fatalf("stderr doesn't name the admitted program as %q: %s", strings.TrimSpace(want), o.Stderr)
			}
		})
	}
}
