package lower

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Each source calls a closure before binding and invokes the unary view with two arguments.
// Node must observe the second argument of the function returned or stored by the accessor.
func TestBoundNameAccessorsStayStopped(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name         string
		line, column int
		stop         string
	}{
		{"literal_getter", 8, 15, "stage 0 can't lower bind of a getter-backed method value yet"},
		{"class_getter", 8, 28, "Adamic 0.1 refuses a value of type Host seen as { f: Unary; }, whose readonly field f becomes writable: a readonly field may hold something narrower than Unary, which a write of Unary would replace; keep f readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"},
		{"class_getter_readonly", 9, 15, "stage 0 can't lower bind of a getter-backed method value yet"},
		{"setter", 9, 15, "stage 0 can't lower bind of a function that declares 2 parameters where the type it's bound as names 1 yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "bind_name_accessors", probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			node := runAgreementNode(t, path)
			if node.code != 0 || string(node.stdout) != "1\n9\n" || len(node.stderr) != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", node.code, node.stdout, node.stderr)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			want := fmt.Sprintf("%s:%d:%d: %s", path, probe.line, probe.column, probe.stop)
			if err == nil || err.Error() != want {
				t.Fatalf("want stop %q, got %v; Node stdout %q", want, err, node.stdout)
			}
		})
	}
}
