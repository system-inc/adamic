package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A statement's checks share one way out, its ladder (internal/native/exceptions.go), where that's
// shorter than each check written in place. Each probe is one long statement whose every call can
// throw, run once for each call that does, held to Node and run under ASan with the leak check, and
// its C has to use a ladder, or the probe pins nothing.
//
// kept: an object literal takes the concatenation it's made with (kept), and calls after it throw.
// A check before the move lets go of the concatenation; a check after it lets go of the object,
// which holds it, and never of the concatenation itself. Mutant: a check's list taken as every name
// the statement ever owned, deletes ignored, and ASan reports the double free.
//
// aside: a ?: whose branch throws, while the statement owns the string it's building. The check in
// the branch lets go of the branch's own temporaries and jumps into the statement's ladder for the
// rest. Mutant: a check in an aside forgetting the statement's list, and the leak check reports the
// built string.
func TestThrowLadderLetsGoOfWhatEachCheckOwns(t *testing.T) {
	t.Parallel()
	const part = "function part(n: number, fail: number): string {\n\tif (n === fail) {\n\t\tthrow new Error('part ' + String(n));\n\t}\n\treturn 'p' + String(n);\n}\n"
	const loop = "for (const fail of [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]) {\n\ttry {\n\t\tconsole.log(build(fail));\n\t} catch (error) {\n\t\tif (error instanceof Error) {\n\t\t\tconsole.log(error.message);\n\t\t}\n\t}\n}\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"kept", part +
			"function joined(box: { head: string; tail: string }, fail: number): string {\n\treturn box.head + '/' + box.tail + part(9, fail);\n}\n" +
			"function build(fail: number): string {\n\treturn part(1, fail) + part(2, fail) + joined({ head: part(3, fail) + part(4, fail), tail: part(5, fail) }, fail) + part(6, fail) + part(7, fail) + part(8, fail);\n}\n" + loop,
			"p1p2p3p4/p5p9p6p7p8\npart 1\npart 2\npart 3\npart 4\npart 5\npart 6\npart 7\npart 8\npart 9\n"},
		{"aside", part +
			"function build(fail: number): string {\n\treturn 'built ' + String(fail) + part(1, fail) + part(2, fail) + (fail % 2 === 0 ? part(3, fail) : part(4, fail) + part(5, fail)) + part(6, fail) + part(7, fail) + part(8, fail);\n}\n" + loop,
			"built 0p1p2p3p6p7p8\npart 1\npart 2\nbuilt 3p1p2p4p5p6p7p8\nbuilt 4p1p2p3p6p7p8\npart 5\npart 6\npart 7\npart 8\nbuilt 9p1p2p4p5p6p7p8\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != probe.stdout {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(native.C(program), "_rung_") {
				t.Fatal("no statement of the probe's uses a ladder")
			}
			sanitized, binary := nativelyUncached(t, program)
			if difference := disagreement(want, sanitized); difference != "" {
				t.Fatalf("native sanitized: %s: exit %d, stdout %q, stderr %q", difference, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Errorf("native leaks: %s", report)
			}
		})
	}
}
