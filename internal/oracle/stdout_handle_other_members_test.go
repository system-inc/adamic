package oracle

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

var stdoutHandleOtherMembers = []struct{ file, node, diagnostic string }{
	{"queue.a", "0\n", "2:16: stage 0 can't lower process.stdout._handle seen as { writeQueueSize?: number; }, which names writeQueueSize where the runtime's stdout handle holds only setBlocking yet"},
	{"bytes.a", "0\n", "2:16: stage 0 can't lower process.stdout._handle seen as { bytesRead?: number; }, which names bytesRead where the runtime's stdout handle holds only setBlocking yet"},
	{"any_fd.a", "1\n", "2:16: stage 0 can't lower process.stdout._handle seen as { fd?: number; writeQueueSize?: number; }, which names fd where the runtime's stdout handle holds only setBlocking yet"},
	{"typed_fd.a", "1\n", "3:16: stage 0 can't lower process.stdout._handle seen as Handle, which names fd where the runtime's stdout handle holds only setBlocking yet"},
	{"any_in.a", "true\n", "2:16: stage 0 can't lower process.stdout._handle seen as { fd?: number; writeQueueSize?: number; }, which names fd where the runtime's stdout handle holds only setBlocking yet"},
	{"any_keyed.a", "1\n", "2:16: stage 0 can't lower process.stdout._handle seen as { fd?: number; writeQueueSize?: number; }, which names fd where the runtime's stdout handle holds only setBlocking yet"},
	{"empty_in.a", "true\n", ""},
	{"alias_fd.a", "1\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { fd?: number; }: optional field fd has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"alias_keyed.a", "1\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { fd?: number; }: optional field fd has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"alias_queue.a", "0\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { writeQueueSize?: number; }: optional field writeQueueSize has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"alias_fd_ts.ts", "1\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { fd?: number; }: optional field fd has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"},
	{"alias_keyed_ts.ts", "1\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { fd?: number; }: optional field fd has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"},
	{"alias_queue_ts.ts", "0\n", "3:14: Adamic 0.1 refuses an unproven relation from {} to { writeQueueSize?: number; }: optional field writeQueueSize has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"},
}

func stdoutHandleOtherMemberPath(file string) string {
	return "internal/oracle/testdata/node_process_handle_other_members/" + file
}

func init() {
	// The refused sources stay out of executable corpora, as the detached-handle
	// witnesses do. The one runtime-stop source has an ordinary count record.
	const file = "empty_in.a"
	path := stdoutHandleOtherMemberPath(file)
	additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) {
		return counted(t, path, false, nil, false, false)
	}})
}

// The initial view and later cast stops already protect the named members.
// The empty view's in test reaches the host stop instead. Removing the initial
// view check reds /any_fd.a, /typed_fd.a, /any_in.a, /any_keyed.a, /queue.a and /bytes.a.
func TestNodeProcessStdoutHandleOtherMembers(t *testing.T) {
	t.Parallel()
	for _, each := range stdoutHandleOtherMembers {
		t.Run(each.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, stdoutHandleOtherMemberPath(each.file)))
			if err != nil {
				t.Fatal(err)
			}
			runner := filepath.Join(repository, "oracle/node.mjs")
			for _, destination := range []string{"pipe", "tty"} {
				truth := nodeProcessTerminal(t, destination, "node", "--disable-warning=ExperimentalWarning", runner, path)
				want := each.node
				// The descriptor number depends on the terminal and open descriptors;
				// the pin requires the numeric value Node supplies, never undefined.
				if destination == "tty" && each.node == "1\n" {
					fd, err := strconv.Atoi(strings.TrimSuffix(string(truth.stdout), "\n"))
					if err != nil || fd < 0 {
						t.Fatalf("terminal descriptor: %+v", truth)
					}
					want = string(truth.stdout)
				}
				if difference := disagreement(run{stdout: []byte(want)}, truth); difference != "" {
					t.Fatalf("Node %s: %s: %+v", destination, difference, truth)
				}
			}
			program, err := lowered(t, path)
			if each.diagnostic != "" {
				if err == nil || err.Error() != path+":"+each.diagnostic {
					t.Fatalf("want exact stop %q, got %v", each.diagnostic, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: this runtime value's properties aren't read by a run-time key yet (reading 'fd')\n")}
			got, binary := nativelyUncached(t, program)
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("native pipe: %s: %+v", difference, got)
			}
			script := filepath.Join(t.TempDir(), "program.mjs")
			if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
				t.Fatal(err)
			}
			for _, destination := range []string{"pipe", "tty"} {
				for _, backend := range []struct {
					name    string
					command []string
				}{
					{"native", []string{binary}}, {"javascript", []string{"node", "--disable-warning=ExperimentalWarning", runner, script}},
				} {
					got := nodeProcessTerminal(t, destination, backend.command...)
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s %s: %s: %+v", backend.name, destination, difference, got)
					}
				}
			}
		})
	}
}
