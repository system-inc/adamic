package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestGenericFunctionPolymorphicRecursionIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `
function nest<Item>(item: Item, depth: number): number {
    return depth === 0 ? 0 : 1 + nest([item], depth - 1);
}
nest(1, 3);
`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "polymorphic recursion") {
		t.Fatalf("want a polymorphic recursion refusal, got %v", err)
	}
}

// Hold the fixture's instantiations as well as its observable behavior: typeof
// currently tests union tags dynamically, so either specialization prints alike.
func TestGenericUnionFixtureHasSeparateInstances(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/generic_instance_key_typeof.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	instances := 0
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "describe_") {
			instances++
		}
	}
	if instances != 2 {
		t.Fatalf("got %d describe instances for distinct unions, want 2", instances)
	}
}

func TestGenericJSONUnionArrayIsNotYet(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/reland_refused/generic_instance_key_json.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "JSON.stringify unions containing arrays (adamic/json-union-array)" {
		t.Fatalf("want the named JSON union array NotYet, got %v", err)
	}
}
