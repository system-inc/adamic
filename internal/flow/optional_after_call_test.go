package flow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOptionalAfterCallFlowEdges(test *testing.T) {
	test.Parallel()
	program := lowered(test, "../oracle/testdata/optional_after_call_required.a")
	checked := 0
	for function := -1; function < len(program.Functions); function++ {
		graph := Build(program, function)
		for _, instruction := range graph.Instructions {
			if instruction == nil || instruction.Expression == nil {
				continue
			}
			encoded, err := json.Marshal(instruction.Expression)
			if err != nil {
				test.Fatal(err)
			}
			// Independently select the inserted required-read error from the real fixture.
			if !strings.Contains(string(encoded), "TypeError: Cannot read properties of undefined") {
				continue
			}
			checked++
			if !CanThrow(program, instruction) {
				test.Fatal("catchable read lost its flow exception edge")
			}
		}
	}
	if checked == 0 {
		test.Fatal("fixture contains no inserted required-read check")
	}
}
