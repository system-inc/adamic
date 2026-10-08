package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewTupleOriginalWatchEvent(t *testing.T) {
	root, _ := tupleOriginalInputs(t)
	_, program := tupleOriginalProgram(t, root, "watch-event-long")
	complete := false
	for _, contract := range program.ViewContracts {
		if contract.Name != "Args" || !contract.TupleUnion || len(contract.Members) != 2 {
			continue
		}
		short, long := false, false
		for _, id := range contract.Members {
			member := program.ViewContracts[id-1]
			if member.FixedTuple && !member.TupleVariable && len(member.Tuple) == 1 {
				short = true
			}
			if member.FixedTuple && member.TupleVariable && member.TupleMinimum == 2 && len(member.Tuple) == 3 && len(member.Fields) == 3 && member.Fields[2].Optional {
				child := program.ViewContracts[member.Tuple[2]-1]
				hasGetTime := false
				for _, field := range child.Fields {
					hasGetTime = hasGetTime || field.Name == "getTime"
				}
				long = child.Name == "Date | undefined" && hasGetTime
			}
		}
		complete = complete || short && long
	}
	if !complete {
		t.Fatal("original callback tuple alternatives or Date projection were reduced")
	}

	verifyOriginalTupleCases(t, []originalTupleCase{
		{"watch-event-short", "undefined\n", "", ""},
		{"watch-event-conditional-short", "absent\n", "", ""},
		{"watch-event-conditional-long", "1\n", "", ""},
		{"watch-event-narrowed-missing", "NaN\n", "", "undefined where the checker narrowed it away: a call since the narrowing put it back"},
		{"watch-event-long", "number\n", "", ""},
		{"watch-event-undefined", "number\n", "", ""},
		{"watch-event-wrong-event", "boolean\n", "", "field read failed: args[1] is not a FileWatcherEventKind | undefined; expected FileWatcherEventKind | undefined, found boolean"},
		{"watch-event-wrong-arity", "undefined\n", "", "field read failed: viewed.args matches no member of Args; expected Args, found object"},
	})
}

func TestCheckedViewTupleOriginalWatchArityMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_WATCH_MUTANT") != "arity" {
		t.Skip("opt-in watch tuple arity mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "watch-event-wrong-arity")
	changed := 0
	for i, c := range p.ViewContracts {
		if c.FixedTuple && !c.TupleVariable && len(c.Tuple) == 1 {
			p.ViewContracts[i].Tuple = nil
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("changed %d alternatives", changed)
	}
	actual := onJavaScriptBackend(t, p)
	if actual.exitCode != 70 {
		t.Fatalf("watch arity mutant caught: expected named refusal; %#v", actual)
	}
	// An escaped mutant must leave this runner green, exposing the missing kill.
	return
}

func TestCheckedViewTupleWatchNarrowingMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_WATCH_MUTANT") != "narrow" {
		t.Skip("opt-in lost narrowing guard")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "watch-event-narrowed-missing")
	code := javascript.JavaScript(p)
	before := "const adamicDefined = (value, message) => value === undefined ? panic(message) : value;"
	if strings.Count(code, before) != 1 {
		t.Fatal("guard missing")
	}
	code = strings.Replace(code, before, "const adamicDefined = (value, message) => value;", 1)
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	actual := onNode(t, file)
	if actual.exitCode != 70 {
		t.Fatalf("lost narrowing guard mutant caught: expected refusal; %#v", actual)
	}
	// An escaped mutant must leave this runner green, exposing the missing kill.
	return
}

func TestCheckedViewTupleOriginalWatchFilename(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"watch-filename-short", "file9\n", "", ""},
		{"watch-filename-long", "file9\n", "", ""},
		{"watch-filename-wrong", "7\n", "", "field read failed: args[0] is not a string; expected string, found number"},
	})
}
