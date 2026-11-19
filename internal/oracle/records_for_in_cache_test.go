package oracle

import (
	"path/filepath"
	"testing"
)

// #1hp4ww2: a for...in over a record reads the key list the record's table keeps (runtime record.c's
// adamic_record_for_in_keys), made on the first loop and let go of when a write adds, deletes or hides a
// key (map.c's adamic_map_forget_key_list). order.a loops before and after a key added, one deleted,
// index keys added and a deleted key added back, during a loop that deletes one key and adds another,
// nested over one record (one nest adding a key each outer pass), and around an Object.keys whose array
// is reversed, which must change no loop; a value written to a key between loops needs no new list.
// hidden.a hides a key with Object.defineProperty, which deletes nothing and adds nothing. Each agrees with
// Node on both backends with nothing leaked. Mutants (records_for_in_cache.mutants.md): the insert's
// let-go dropped turns order.a red at its third line, define_hidden's turns hidden.a red at its third,
// and the shared bypass dropped is TestParallelMutant_for_in_keys_shared's TSan race. The delete's
// let-go has no witness in output: a stale list still naming a deleted key is what the loop's hasOwn
// guard already skips, so it frees the list sooner and changes no line.
func TestRecordForInCachedKeys(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"order.a", "hidden.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_for_in_cache", name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || len(node.stdout) == 0 || len(node.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %q; %s %q, stderr %q", backend, difference, node.stdout, backend, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}
