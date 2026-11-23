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
// read stops at compile time when the program writes a record of numbers or booleans and the receiver may be
// a record: object, unknown, a type parameter, or a record. Each pin names its program, Node's answer, and the
// read it stops at. If a stop goes missing, the program is run natively under the sanitizers and its run is
// part of the failure.
func TestInRecordReadStops(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ file, node, at string }{
		{"record_number.ts", "number\n", "o.k"},
		// The key is absent, so the read never runs and main agrees, but its program holds a number record
		// and the stop can't tell which records reach o.
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
			const words = "a read of 'k' through a value that may be a record of numbers or booleans"
			if notYet.Where != want || !strings.Contains(err.Error(), words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), notYet.Where, words, want)
			}
		})
	}
}

// The neighbors still lower and agree: a record of strings, whose slots are references, read the same way;
// a run-time key read, which boxes a record's value by its kind; and a receiver declared as a program object
// type, which no record reaches, alone and beside a number record in the same program.
func TestInRecordReadNeighborsAgree(t *testing.T) {
	t.Parallel()
	for file, node := range map[string]string{
		"record_string.ts":           "string\n",
		"keyed_read.ts":              "number\n",
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
