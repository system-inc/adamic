package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestRecordOptionalProperty(t *testing.T) {
	t.Parallel()
	const wrongType = "field read failed: value?.name is not a string; expected string, found number, which a computed-key or index-signature write may have stored"
	for _, test := range []struct {
		file, stdout, prototype string
		notYet, at              string
	}{
		{file: "read.a", stdout: "missing\nset\nmissing\n"},
		{file: "null_receiver.ts", stdout: "missing\nset\nmissing\n"},
		// The continuing form stops in recordAccess ("optional record indexing and optional record property reads outside the one-step path stay stopped") until #15kcz6y joins it to main's guard.
		{file: "chain.ts", stdout: "undefined\n3\n", notYet: "optional record property access", at: ":7:20"},
		{file: "missing_required.ts", stdout: "missing\ncaught\n", notYet: "optional record property access", at: ":14:28"},
		{file: "receiver_once.ts", stdout: "set\n1\nmissing\n2\n"},
		{file: "wrong_type.ts", stdout: "3\n"},
		{file: "stop_prototype.ts", stdout: "function\n", prototype: "toString"},
		{file: "stop_constructor.ts", stdout: "function\n", prototype: "constructor"},
		{file: "stop_proto.ts", stdout: "object\n", prototype: "__proto__"},
		{file: "stop_index.ts", stdout: "missing\n"},
	} {
		t.Run(test.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/record_optional_property", test.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.stdout || len(truth.stderr) != 0 {
				t.Fatalf("Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			switch {
			case test.prototype != "":
				var refusal *lower.Refused
				if !errors.As(err, &refusal) || refusal.What != "record key "+test.prototype+" names an Object.prototype member" || refusal.Fix != "use own keys" {
					t.Fatalf("want the own-key refusal, got %v", err)
				}
				return
			case test.notYet != "":
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != test.notYet || notYet.Where != path+test.at {
					t.Fatalf("want NotYet %q at %s, got %v", test.notYet, path+test.at, err)
				}
				return
			case test.file == "stop_index.ts":
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "optional record indexing") {
					t.Fatalf("want optional record indexing NotYet, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if test.file == "wrong_type.ts" {
					if backend.got.exitCode != 70 || len(backend.got.stdout) != 0 || !strings.Contains(string(backend.got.stderr), wrongType) {
						t.Fatalf("%s: want row 294 stop, got exit %d stdout %q stderr %q", backend.name, backend.got.exitCode, backend.got.stdout, backend.got.stderr)
					}
					continue
				}
				if diff := disagreement(truth, backend.got); diff != "" {
					t.Fatalf("%s: %s, exit %d stdout %q stderr %q", backend.name, diff, backend.got.exitCode, backend.got.stdout, backend.got.stderr)
				}
			}
		})
	}
}
