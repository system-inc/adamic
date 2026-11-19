package oracle

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// #1trc7d3 across modules: an exported alias (stdout_alias_modules/out.a) resolves through the import to its
// declaration, so an importing module's write and isTTY, directly or through a re-export, agree with Node on
// the release and sanitized builds and the JavaScript backend, and its use as a value stops by name as a local
// alias's does. A read off a namespace import stops at an older NotYet, a method of a class stage 0 doesn't
// have, unrelated to the alias; the day it lifts, that row goes red and wants a pin of its own.
func TestStdoutAliasModules(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]struct {
		printed    []byte
		notYet, at string
	}{
		"use_write.a":     {printed: []byte("x\n")},
		"use_tty.a":       {printed: []byte("not a terminal\n")},
		"use_reexport.a":  {printed: []byte("z\n")},
		"use_value.a":     {notYet: "node:process.Process.stdout, bound to out, used other than as the receiver of write, isTTY or columns", at: ":3:15"},
		"use_namespace.a": {notYet: "a method of a class stage 0 doesn't have", at: ":3:1"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/stdout_alias_modules", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if want.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != want.notYet || !strings.HasSuffix(notYet.Where, want.at) {
					t.Fatalf("want NotYet %q at %s, got %v", want.notYet, want.at, err)
				}
				return
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || !bytes.Equal(truth.stdout, want.printed) || len(truth.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q; want stdout %q", truth.exitCode, truth.stdout, truth.stderr, want.printed)
			}
			if err != nil {
				t.Fatalf("an imported alias of process.stdout must lower: %v", err)
			}
			sanitized, binary := natively(t, program)
			for backend, got := range map[string]run{"release": released(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %q; %s exit %d, stdout %q, stderr %q", backend, difference, truth.stdout, backend, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
