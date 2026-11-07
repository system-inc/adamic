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
func TestInterfaceTypeMethodGap(test *testing.T) {
	main, absolutePathError := filepath.Abs("gaps/interfaceTypeMethod.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if got := string(onNode(test, main)); got != "1\n" {
		test.Fatalf("Node: %q", got)
	}
	loaded, loadError := load.Load([]string{main})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), loaded)
	isGap := func(lowerError error) bool {
		var diagnostic *lower.NotYet
		return errors.As(lowerError, &diagnostic) &&
			diagnostic.Where == main+":10:12" &&
			diagnostic.What == "a class method through a view that erases its prototype origin"
	}
	if !isGap(lowerError) {
		test.Fatalf("recorded lowering gap changed: %v", lowerError)
	}
	test.Logf("Node prints 1; lowering refuses before native emission: %s", lowerError)
	control := estreeInterfaceControlPath(test)
	loaded, loadError = load.Load([]string{control})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if _, lowerError = lower.Lower(context.Background(), loaded); lowerError != nil || isGap(lowerError) {
		test.Fatalf("default-free control lowering: %v", lowerError)
	}
	controlBinary, controlScript := build(test, control, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, control),
		"emitted": onNode(test, controlScript),
		"native":  execute(test, "", controlBinary),
	} {
		if string(got) != "1\n" {
			test.Fatalf("%s default-free control: %s", name, got)
		}
	}
	test.Log(
		"renamed default-free method through an explicit callback yields 1 in all builds; the lowering-gap check rejects that control",
	)
}
