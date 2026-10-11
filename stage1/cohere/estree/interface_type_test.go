package estree

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestInterfaceTypeMethodGap(t *testing.T) {
	main, err := filepath.Abs("gaps/interfaceTypeMethod.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(onNode(t, main)); got != "1\n" {
		t.Fatalf("Node: %q", got)
	}
	loaded, err := load.Load([]string{main})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	isGap := func(err error) bool {
		var diagnostic *lower.NotYet
		return errors.As(err, &diagnostic) && diagnostic.Where == main+":10:12" && diagnostic.What == "a class method through a view that erases its prototype origin"
	}
	if !isGap(err) {
		t.Fatalf("recorded lowering gap changed: %v", err)
	}
	t.Logf("Node prints 1; lowering refuses before native emission: %s", err)
	control := estreeInterfaceControlPath(t)
	loaded, err = load.Load([]string{control})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lower.Lower(context.Background(), loaded); err != nil || isGap(err) {
		t.Fatalf("default-free control lowering: %v", err)
	}
	controlBinary, controlScript := build(t, control, true)
	for name, got := range map[string][]byte{"Node": onNode(t, control), "emitted": onNode(t, controlScript), "native": execute(t, "", controlBinary)} {
		if string(got) != "1\n" {
			t.Fatalf("%s default-free control: %s", name, got)
		}
	}
	t.Log("renamed default-free method through an explicit callback yields 1 in all builds; the lowering-gap check rejects that control")
}
