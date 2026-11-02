package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A string key starting with # never reaches a private on Node, so "#message" in error and "#x" in an
// instance are false there. Natively an Error made without a message keeps its inherited message in a
// #message slot, which adamic_has_property finds on a classless object, and the JavaScript backend keeps a
// class's privates as # string keys. in with such a key is NotYet at the key (unknown.go's inProperty), in .a
// and .ts. The fixtures sit in private_key_in, since the flow and fresh corpora lower every program at the
// top of testdata. Mutant: the # check dropped from inProperty; with it applied, these lower again.
func TestPrivateKeyInNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, stdout string }{
		{"message_hash.a", "message_hash.a:4:48", "false\n"},
		{"message_hash.ts", "message_hash.ts:4:48", "false\n"},
		{"private_hash.a", "private_hash.a:10:30", "1 false\n"},
		{"private_hash.ts", "private_hash.ts:10:30", "1 false\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/private_key_in", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), "in with a key starting with #") {
				t.Fatalf("want in with a # key NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}
