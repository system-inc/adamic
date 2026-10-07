package fuzz

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// stringLiteralTypeof is typeof of a string written out, which main emits C for that clang refuses.
var stringLiteralTypeof = regexp.MustCompile(`typeof \('[^']*'\)`)

// operatorsOnMain are the operators each kind takes on main, without an opt-in, written down apart
// from the scene's tables so a table that drops one is caught.
var operatorsOnMain = map[string][]string{
	"number":             {"typeof", "-", "+", "~", "1 /", "1 / -", "String", "template", "===", "!==", "<", "<=", ">", ">="},
	"string":             {"typeof", "+", "String", "template", "===", "!==", "<", "<=", ">", ">="},
	"boolean":            {"typeof", "!", "+", "String", "template", "===", "!=="},
	"number|undefined":   {"typeof", "String", "template", "=== undefined", "!== undefined", "===", "!=="},
	"string|undefined":   {"typeof", "String", "template", "=== undefined", "!== undefined", "=== null", "!== null", "===", "!=="},
	"OpsPoint|undefined": {"typeof", "=== undefined", "!== undefined", "=== null", "!== null", "===", "!=="},
	"OpsPoint":           {"typeof", "=== null", "!== null", "===", "!=="},
	"number[]":           {"typeof", "=== null", "!== null", "===", "!=="},
	"function":           {"typeof", "=== null", "!== null", "===", "!=="},
	"OpsThing":           {"typeof", "=== null", "!== null", "===", "!=="},
	"Map":                {"typeof", "=== null", "!== null", "===", "!=="},
	"Set":                {"typeof", "=== null", "!== null", "===", "!=="},
	"null":               {"typeof", "String", "===", "!=="},
	"undefined":          {"typeof", "String", "===", "!=="},
}

// What each opt-in puts in, as the scene notes it (operatorSeen): none of them appears without it.
var operatorsOptedIn = map[string][]string{
	nullableSlots:        {"held null string|null", "held null boolean|null", "held boolean boolean|null", "held loneSurrogate string|null", "applied typeof string|null", "applied === null boolean|null"},
	objectStrings:        {"applied String OpsPoint", "applied template OpsPoint|undefined", "applied template null", "applied template undefined", "applied String Map", "applied template OpsThing"},
	unaryCoercion:        {"applied - string", "applied ~ string", "applied - boolean", "applied ~ boolean"},
	nullScalarComparison: {"applied === null number|undefined", "compared null opsNumbers", "compared null opsBooleans", "compared null opsMaybeNumbers"},
}

// Within forty seeds, asked for with -with operators, the scene applies every operator to every kind
// that takes it, carries every representation by every route in every slot main lowers, and every
// program it makes checks and lowers. What main doesn't lower yet stays out until its opt-in asks.
func TestOperatorsShapes(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	directory := t.TempDir()
	for seed := uint64(1); seed <= 40; seed++ {
		generator := newGenerator(seed, nil, []string{operatorsScene})
		source := generator.program().Source()
		for what := range generator.operators {
			seen[what] = true
		}
		if stringLiteralTypeof.MatchString(source) {
			t.Errorf("seed %d took typeof of a string written out without %s", seed, typeofStringLiteral)
		}
		path := filepath.Join(directory, "program.a")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Errorf("seed %d: the checker refused it: %v", seed, err)
			continue
		}
		if _, err := lower.Lower(context.Background(), program); err != nil {
			t.Errorf("seed %d: stage 0 didn't lower it: %v", seed, err)
		}
	}
	for kind, operators := range operatorsOnMain {
		for _, operator := range operators {
			if !seen["applied "+operator+" "+kind] {
				t.Errorf("40 seeds never applied %s to a %s", operator, kind)
			}
		}
	}
	slots := map[string]bool{}
	for _, representation := range operatorRepresentations {
		var held bool
		for _, slot := range append([]*operatorKind{representation.kind}, representation.slots...) {
			if slot.optIn != "" {
				continue
			}
			held = held || seen["held "+representation.name+" "+slot.name]
			slots[slot.name] = slots[slot.name] || seen["held "+representation.name+" "+slot.name]
		}
		if !held {
			t.Errorf("40 seeds never probed %s", representation.name)
		}
	}
	for slot, held := range slots {
		if !held {
			t.Errorf("40 seeds never held a value in a %s", slot)
		}
	}
	for _, route := range append([]string{"direct", "loop"}, operatorRoutes...) {
		if !seen["route "+route] {
			t.Errorf("40 seeds never carried a value by %s", route)
		}
	}
	for _, kind := range []*operatorKind{operatorNull, operatorUndefined} {
		for _, peers := range kind.peers {
			if peers.optIn == "" && !seen["compared "+kind.name+" "+peers.name] {
				t.Errorf("40 seeds never compared %s to %s", kind.name, peers.name)
			}
		}
	}
	for optIn, shapes := range operatorsOptedIn {
		for _, shape := range shapes {
			if seen[shape] {
				t.Errorf("without %s, the scene wrote %s", optIn, shape)
			}
		}
	}

	// Each opt-in puts its shapes in.
	for optIn, shapes := range operatorsOptedIn {
		opted := map[string]bool{}
		for seed := uint64(1); seed <= 40; seed++ {
			generator := newGenerator(seed, nil, []string{operatorsScene, optIn})
			generator.program()
			for what := range generator.operators {
				opted[what] = true
			}
		}
		for _, shape := range shapes {
			if !opted[shape] {
				t.Errorf("with %s, 40 seeds never wrote %s", optIn, shape)
			}
		}
	}
	var literal bool
	for seed := uint64(1); seed <= 40 && !literal; seed++ {
		literal = stringLiteralTypeof.MatchString(GenerateFeatures(seed, nil, []string{operatorsScene, typeofStringLiteral}).Source())
	}
	if !literal {
		t.Errorf("with %s, 40 seeds never took typeof of a string written out", typeofStringLiteral)
	}
	if strings.Contains(Generate(1).Source(), "OpsBox") {
		t.Error("without -with operators, the scene was written")
	}
	if !strings.Contains(GenerateFeatures(1, nil, []string{operatorsScene}).Source(), "OpsBox") {
		t.Error("with -with operators, the scene wasn't written")
	}
}
