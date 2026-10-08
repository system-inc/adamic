package lower

import (
	"errors"
	"testing"
)

func TestEvolvingArrayRejectsChangingStorage(t *testing.T) {
	_, err := lowerSource(t, `function collect(): (string|number)[] { const result = []; result.push(1); console.log((result[0] ?? 0).toString()); result.push("owned" + "!"); return result; } console.log(collect().length.toString());`)
	var unsupported *NotYet
	if !errors.As(err, &unsupported) || unsupported.What != "an array of any" {
		t.Fatalf("want changing read storage unsupported, got %v", err)
	}
}
