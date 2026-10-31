package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A checked view over an intersection of plain object types (#qqvq6n7) is described by the
// intersection's resolved properties, each with its refined type: a boolean literal refinement, a
// base field, a field narrowed from string | number to string, an added field, a nested
// intersection read field by field, and a write through it that reads back. Each agrees with Node
// natively, sanitized and on the JavaScript backend. The nested ones cast through an interface that
// extends its source with no tag of its own, the downcast .a refuses (#995z2fa), so each is .ts.
func TestIntersectionViewsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/intersection_views"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, path)
		})
	}
	if programs != 3 {
		t.Fatalf("%d programs in %s, want 3: a program added or removed needs this count changed with it", programs, directory)
	}
}

// Where the object doesn't hold what the intersection refines a field to, the read stops with the
// view's message and exit 70, after the untagged call has printed what Node prints: a literal tag
// that holds another number, a boolean refined to true that holds false, and a field narrowed to a
// string that holds a number. A write stops where the slot holds what the view says it can't.
// Node, without the checks, prints on.
func TestIntersectionViewsStopAtTheRead(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/intersection_views_stops/"
	for _, stop := range []struct{ name, node, stdout, stderr string }{
		{"nested_literal.ts", "other\nnode:fs 12\n", "other\n", "field read failed: imported.argument.literal expected 11, found number 12"},
		{"nested_boolean.ts", "other\n4 false\n", "other\n", "field read failed: declaration.clause.isTypeOnly expected true, found boolean false"},
		{"nested_token.ts", "other\nnode:fs undefined\n", "other\n", "field read failed: imported.argument.token is not a string; expected string, found number"},
		{"nested_write.ts", "other\nstamped\nnumber\n", "other\n", "field read failed: <write>.flags is not a number; expected number, found string"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(repository, directory+stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte("adamic: panic: " + stop.stderr + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}
