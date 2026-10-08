package oracle

import (
	"os"
	"slices"
	"testing"
)

func TestCheckedViewTupleOriginalModuleSpecifiers(t *testing.T) {
	root, manifest := tupleOriginalInputs(t)
	if manifest.PrivateModuleWorker.Declaration == "" {
		t.Fatal("prepare_optional_tuple.cjs must bind the full original private worker declaration")
	}
	_, p := tupleOriginalProgram(t, root, "module-specifiers-present")
	complete := false
	for _, c := range p.ViewContracts {
		if c.Name != "OriginalModuleSpecifierTuple" || !c.FixedTuple || !c.TupleVariable || c.TupleMinimum != 0 || len(c.Tuple) != 5 || len(c.Fields) != 5 {
			continue
		}
		optional := true
		for _, f := range c.Fields {
			optional = optional && f.Optional
		}
		file := p.ViewContracts[c.Tuple[2]-1]
		fields := []string{}
		for _, f := range file.Fields {
			fields = append(fields, f.Name)
		}
		slices.Sort(fields)
		complete = complete || optional && file.Name == "SourceFile | undefined" && slices.Equal(fields, manifest.Fields["SourceFile"]) && len(fields) > 50
	}
	if !complete {
		t.Fatal("original five-position tuple and full SourceFile projection were reduced")
	}
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"module-specifiers-absent", "absent\n", "", ""},
		{"module-specifiers-undefined", "absent\n", "", ""},
		{"module-specifiers-present", "specifier9\n", "", ""},
		{"module-specifiers-wrong-kind", "boolean\n", "", "field read failed: entry[1] is not a readonly string[] | undefined; expected readonly string[] | undefined, found boolean"},
		{"module-specifiers-wrong-element", "false\n", "", "element read failed: entry[1][element] expected string, found boolean"},
		{"module-specifiers-wrong-arity", "absent\n", "", "field read failed: viewed.value is not a OriginalModuleSpecifierTuple; expected OriginalModuleSpecifierTuple, found array"},
	})
}

func TestCheckedViewTupleOriginalModuleArityMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_MODULE_MUTANT") != "arity" {
		t.Skip("opt-in original optional tuple arity mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "module-specifiers-wrong-arity")
	changed := 0
	for i, c := range p.ViewContracts {
		if c.Name == "OriginalModuleSpecifierTuple" && c.FixedTuple {
			p.ViewContracts[i].Tuple = append(c.Tuple, c.Tuple[0])
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("changed %d tuple arity plans", changed)
	}
	actual := onJavaScriptBackend(t, p)
	if actual.exitCode != 70 {
		t.Fatalf("original optional tuple arity mutant caught: expected refusal; %#v", actual)
	}
	// An escaped mutant must leave this runner green, exposing the missing kill.
	return
}

func TestCheckedViewTupleOriginalModuleKind(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"module-kind-absent", "undefined\n", "", ""},
		{"module-kind-undefined", "undefined\n", "", ""},
		{"module-kind-present", "relative\n", "", ""},
		{"module-kind-wrong-literal", "wrong\n", "", "field read failed: entry[0] expected \"ambient\" | \"node_modules\" | \"paths\" | \"redirect\" | \"relative\" | undefined, found string wrong"},
		{"module-kind-wrong-kind", "false\n", "", "field read failed: entry[0] is not a \"ambient\" | \"node_modules\" | \"paths\" | \"redirect\" | \"relative\" | undefined; expected \"ambient\" | \"node_modules\" | \"paths\" | \"redirect\" | \"relative\" | undefined, found boolean"},
	})
}
