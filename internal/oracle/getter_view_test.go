package oracle

import (
	"path/filepath"
	"testing"
)

// A getter-backed member read through a structural view's readonly field runs the getter, as Node does:
// finishAccessors rewrites every read of an accessor name into its dispatcher (class_accessors.go), so the view
// #rv75m9k's road B puts over tsc's NodeFactory reads what tsc reads. Natively and on the JavaScript backend, each
// read makes a fresh closure. Mutant: noteAccessorNames records no name, and the native row reads the getter's slot
// instead of running it.
func TestGetterReadThroughView(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/getter_view/factory.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("a getter-backed member read through a view must lower: %v", err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "12\n13\nfactory\n" {
		t.Fatalf("Node: %+v", node)
	}
	native, _ := natively(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("%s disagrees with Node: %s", backend, difference)
		}
	}
}
