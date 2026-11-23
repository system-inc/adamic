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

// A private field is held under privateName's spelling, "#secret@1", as an ordinary slot on native and the
// JavaScript backend alike, so a string key spelling it found it: derived.hasOwnProperty('#secret@1') printed
// true on both backends where Node prints false (#a0xw8mg, from CLI's agreement daily). Until a private slot
// is stored where no string can reach it, an own-key or presence test whose key may spell one stops at compile
// time: a literal with a leading # and an @, or any key the lowering can't read on a receiver that can hold a
// private slot. Object.hasOwn already refuses any # key. Each pin names the program, the text its stop sits
// at, and the words it must say; the fixtures sit in private_key_own, since the flow and fresh corpora lower
// every program at the top of testdata.
func TestPrivateKeyOwnStops(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ file, at, words string }{
		{"literal.a", "'#secret@1'", "an own-property test when the key #secret@1 spells a private slot's native name"},
		{"computed.a", "'#secret@' + suffix", "an own-property test when a key that isn't a string literal may spell a private slot's native name"},
		{"detached.a", "'#secret@' + suffix", "an own-property test when a key that isn't a string literal may spell a private slot's native name"},
		{"in_literal.a", "'#secret@1'", "in when the key #secret@1 spells a private slot's native name"},
		{"hasown.a", "Object.hasOwn(", "hasOwn requires a string literal naming a declared public field or method"},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/private_key_own", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte("false\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			want, err := privateKeyLocation(path, pin.at)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			where := ""
			switch {
			case err == nil:
				t.Fatalf("lowered; want a stop at %s saying %q", want, pin.words)
			case errors.As(err, &notYet):
				where = notYet.Where
			case errors.As(err, &refused):
				where = refused.Where
			default:
				t.Fatalf("want a compile-time stop, got %v", err)
			}
			if where != want || !strings.Contains(err.Error(), pin.words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), where, pin.words, want)
			}
		})
	}
}

// The neighbors still lower and agree: a # key that can't spell a private slot's name ('#secret', in and
// hasOwnProperty), a declared field, and a key the lowering can't read on a plain object no class with a
// private field can be seen as.
func TestPrivateKeyOwnNeighborsAgree(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/private_key_own", "control.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("1 false true true false true\n")}, truth) != "" {
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

// privateKeyLocation is the path:line:column where text first occurs in the file, as Program.Where prints it.
func privateKeyLocation(path, text string) (string, error) {
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
