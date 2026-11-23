package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A named dynamic read of a record returned its slot as a reference whatever the record's value storage
// (union.c:332 to :337), so `'k' in o` then `typeof o.k`, with o a Record<string, number> passed as object,
// printed number on Node and the JavaScript backend and faulted natively under ASan (#859ksab, Language's
// probes J-t5mxr1v and J-dwhn493 on main 0c9b1201). Until Library's #3tjjcjs boxes the slot by its kind, the
// read stops at compile time when the receiver may be a record: object, unknown, a type parameter, or a
// record. A record of strings reads right today but stops too, since which records a program makes can't be
// read from its types (from_entries.ts makes one with no type node, and stops as Object.fromEntries does). Each pin names its program, Node's answer, and the
// read it stops at. If a stop goes missing, the program is run natively under the sanitizers and its run is
// part of the failure.
func TestInRecordReadStops(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ file, node, at string }{
		{"record_number.ts", "number\n", "o.k"},
		{"record_string.ts", "string\n", "o.k"},
		// The key is absent, so the read never runs and main agrees, but the stop can't tell which records
		// reach o.
		{"record_absent.ts", "absent\n", "o.k"},
		{"record_string_of.ts", "41\n", "o.k"},
		{"unknown_param.ts", "number\n", "o.k"},
		{"generic_param.ts", "number\n", "o.k"},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/in_record_read", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(pin.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			want, err := inRecordReadLocation(path, pin.at)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				t.Fatalf("lowered; want a stop at %s. Sanitized native: exit %d, stdout %q, stderr %q", want, native.exitCode, native.stdout, native.stderr)
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet, got %v", err)
			}
			const words = "a read of 'k' through a value that may be a record, whose numbers and booleans a named dynamic read doesn't box yet"
			if notYet.Where != want || !strings.Contains(err.Error(), words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), notYet.Where, words, want)
			}
		})
	}
}

// Two roads stop at compile time before any read can misread a record, each at a stop of its own: a record
// made with no type node naming it, Object.fromEntries' { [k: string]: number } (Compiler's must on
// #859ksab: the stop can't count on type nodes), and a literal key read o['k'] after the in test, which
// lowering doesn't take yet. Each is pinned, so a change that lets either through meets this file.
func TestInRecordReadEarlierStops(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"from_entries.ts", "keyed_read.ts"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/in_record_read", file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte("number\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				t.Fatalf("lowered; want a compile-time stop. Sanitized native: exit %d, stdout %q, stderr %q", native.exitCode, native.stdout, native.stderr)
			}
			var refused *lower.Refused
			var notYet *lower.NotYet
			if !errors.As(err, &refused) && !errors.As(err, &notYet) {
				t.Fatalf("want a compile-time stop, got %v", err)
			}
		})
	}
}

// The neighbors still lower and agree: a receiver declared as a program object type, which no record
// reaches, alone and beside a number record.
func TestInRecordReadNeighborsAgree(t *testing.T) {
	t.Parallel()
	for file, node := range map[string]string{
		"fixed_control.ts":           "number absent\n",
		"fixed_in_record_program.ts": "number absent 1\n",
	} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/in_record_read", file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// inRecordReadLocation is the path:line:column where text first occurs in the file, as Program.Where prints it.
func inRecordReadLocation(path, text string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for number, line := range strings.Split(string(source), "\n") {
		if column := strings.Index(line, text); column >= 0 {
			return fmt.Sprintf("%s:%d:%d", filepath.ToSlash(path), number+1, column+1), nil
		}
	}
	return "", fmt.Errorf("%q isn't in %s", text, path)
}
