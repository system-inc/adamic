package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestGenericFunctionPolymorphicRecursionIsRefused(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `
function nest<Item>(item: Item, depth: number): number {
    return depth === 0 ? 0 : 1 + nest([item], depth - 1);
}
nest(1, 3);
`)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) || !strings.Contains(lowerSourceError.Error(), "polymorphic recursion") {
		test.Fatalf("want a polymorphic recursion refusal, got %v", lowerSourceError)
	}
}

// Hold the fixture's instantiations as well as its observable behavior: typeof
// currently tests union tags dynamically, so either specialization prints alike.
func TestGenericUnionFixtureHasSeparateInstances(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/generic_instance_key_typeof.a")
	if readError != nil {
		test.Fatal(readError)
	}
	program, lowerSourceError := lowerSource(test, string(source))
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	instances := 0
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "describe_") {
			instances++
		}
	}
	if instances != 2 {
		test.Fatalf("got %d describe instances for distinct unions, want 2", instances)
	}
}

func TestGenericJSONUnionArrayIsNotYet(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/reland_refused/generic_instance_key_json.a")
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) ||
		notYet.What != "JSON.stringify unions containing arrays (adamic/json-union-array)" {
		test.Fatalf("want the named JSON union array NotYet, got %v", lowerSourceError)
	}
}
