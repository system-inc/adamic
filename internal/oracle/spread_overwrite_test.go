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

// A field written after a leading spread, { ...source, k: undefined }, lands in the slot the spread brings
// for it, since the literal takes or copies the source object. Family 65 found two writes native can't keep
// where Node and the JavaScript backend agree (#xy74m5x, main ac44317c): a number slot written as a reference
// released the number's bits as a pointer (exit 139), and a field optional in the source's type had no slot in
// an actual {} (exit 70 for a string, 139 for a number). Each stops at the overwrite until a spread writes its
// own layout. Seats' four programs are here as they ran, beside a source that is never undefined, to show the
// slot's representation is the cause and not the source's nullability, and a field the source's type doesn't
// declare, which by width the actual source may still hold. If a stop goes missing, the program is
// run natively under the sanitizers and its run is part of the failure.
func TestSpreadOverwriteStops(t *testing.T) {
	t.Parallel()
	const held = "after a spread whose source holds k as"
	const optional = "its type makes k optional"
	const width = "a spread that adds a field the source doesn't have"
	for _, pin := range []struct{ file, node, at, words string }{
		{"number_source.ts", "undefined\nundefined\n", "k: undefined", held},
		{"number_present.ts", "undefined\n", "k: undefined", held},
		{"string_source.ts", "undefined\nundefined\n", "k: undefined", held},
		{"optional_string.ts", "true\nundefined\ntrue\nundefined\n", "k: undefined", optional},
		{"optional_number.ts", "true\nundefined\ntrue\nundefined\n", "k: undefined", optional},
		// By width a { readonly a: number } may hold k in a slot laid out for a string. A property stops at
		// object.go's own check, a method in spreadOverwrite, in the same words.
		{"width.ts", "1 5\n", "k: 5", width},
		{"width_method.ts", "1 5\n", "k() {", width},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_overwrite", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(pin.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			want, err := spreadOverwriteLocation(path, pin.at)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				t.Fatalf("lowered; want a stop at %s saying %q. Sanitized native: exit %d, stdout %q, stderr %q", want, pin.words, native.exitCode, native.stdout, native.stderr)
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet, got %v", err)
			}
			if notYet.Where != want || !strings.Contains(err.Error(), pin.words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), notYet.Where, pin.words, want)
			}
		})
	}
}

// The neighbors still lower and agree with Node: a spread with no overwrite, an overwrite whose value the
// source's slot holds alike over a source that may be undefined (its fallback object is laid out by the
// literal), and one over a present source beside a field the spread carries through.
func TestSpreadOverwriteNeighborsAgree(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_overwrite", "neighbors.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("1\nundefined\n2\n2\none 2\n")}, truth) != "" {
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
}

// spreadOverwriteLocation is the path:line:column where text first occurs in the file, as Program.Where prints it.
func spreadOverwriteLocation(path, text string) (string, error) {
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
