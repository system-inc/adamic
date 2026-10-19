package ir

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"reflect"
	"testing"
)

// A program with entries in every map is the same bytes every time gob writes it, and reads back as it was: estree's
// and lint's lowered-program products are a function of the program (landable-13b's tree, Oct 11, refused two estree
// products for differing bytes under one key). Mutant: SortedMap without GobEncode (a plain map's random order).
func TestAProgramIsTheSameBytesEveryTimeGobWritesIt(t *testing.T) {
	t.Parallel()
	program := Program{CheckedWrites: SortedMap[string, bool]{}, ContractTypeNames: SortedMap[int, string]{},
		ViewContractTypes: SortedMap[int, ViewContractID]{}, UninitializedFields: SortedMap[string, bool]{},
		CheckedFields: SortedMap[string, bool]{}, MethodTargets: SortedMap[int, []int]{}, StructuralMethodThunks: SortedMap[int, bool]{},
		FunctionTypeTargets: SortedMap[int, []int]{}, RestArgumentSlots: SortedMap[RestArguments, int]{}}
	function := Function{OptionalParameters: SortedMap[int, bool]{}}
	for index := range 16 {
		name := string(rune('a' + index))
		program.CheckedWrites[name] = index%2 == 0
		program.ContractTypeNames[index] = "contract " + name
		program.ViewContractTypes[index] = ViewContractID(index + 1)
		program.UninitializedFields[name], program.CheckedFields[name] = true, index%2 == 0
		program.MethodTargets[index], program.FunctionTypeTargets[index] = []int{index, index + 1}, []int{index}
		program.StructuralMethodThunks[index] = true
		program.RestArgumentSlots[RestArguments{}] = index
		function.OptionalParameters[index] = index%3 == 0
	}
	program.Functions = []Function{function}
	encodings := map[[32]byte]bool{}
	var last []byte
	for range 32 {
		var encoded bytes.Buffer
		if err := gob.NewEncoder(&encoded).Encode(program); err != nil {
			t.Fatal(err)
		}
		encodings[sha256.Sum256(encoded.Bytes())] = true
		last = encoded.Bytes()
	}
	if len(encodings) != 1 {
		t.Fatalf("one program, %d encodings in 32", len(encodings))
	}
	var read Program
	if err := gob.NewDecoder(bytes.NewReader(last)).Decode(&read); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(read.CheckedWrites, program.CheckedWrites) || !reflect.DeepEqual(read.ContractTypeNames, program.ContractTypeNames) ||
		!reflect.DeepEqual(read.MethodTargets, program.MethodTargets) || !reflect.DeepEqual(read.Functions[0].OptionalParameters, function.OptionalParameters) ||
		!reflect.DeepEqual(read.RestArgumentSlots, program.RestArgumentSlots) {
		t.Fatalf("the program read back differs: %+v", read)
	}
}

// A SortedMap reads back nil or not exactly as a plain map does, whether it was nil, empty or held entries, so
// "nil keeps conservative admission" (StructuralMethodThunks) means what it meant before.
func TestASortedMapIsNilWhereAPlainMapIs(t *testing.T) {
	t.Parallel()
	type plain struct{ Field map[int]bool }
	type sorted struct{ Field SortedMap[int, bool] }
	for name, value := range map[string]map[int]bool{"nil": nil, "empty": {}, "held": {1: true}} {
		var plainBytes, sortedBytes bytes.Buffer
		if err := gob.NewEncoder(&plainBytes).Encode(plain{value}); err != nil {
			t.Fatal(err)
		}
		if err := gob.NewEncoder(&sortedBytes).Encode(sorted{value}); err != nil {
			t.Fatal(err)
		}
		var plainRead plain
		var sortedRead sorted
		gob.NewDecoder(&plainBytes).Decode(&plainRead)
		gob.NewDecoder(&sortedBytes).Decode(&sortedRead)
		if (plainRead.Field == nil) != (sortedRead.Field == nil) || len(plainRead.Field) != len(sortedRead.Field) {
			t.Fatalf("%s: a plain map reads back %#v, a SortedMap %#v", name, plainRead.Field, sortedRead.Field)
		}
	}
}

// No map a program holds is a plain one, so no field added later brings gob's random order back. Mutant: one field a
// plain map again.
func TestEveryMapAProgramHoldsIsSorted(t *testing.T) {
	t.Parallel()
	encoder := reflect.TypeFor[gob.GobEncoder]()
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type, string)
	walk = func(kind reflect.Type, at string) {
		if seen[kind] {
			return
		}
		seen[kind] = true
		switch kind.Kind() {
		case reflect.Map:
			if !kind.Implements(encoder) {
				t.Errorf("%s is a plain %s: gob writes it in random order; make it a SortedMap", at, kind)
			}
			walk(kind.Elem(), at+"[value]")
		case reflect.Pointer, reflect.Slice, reflect.Array:
			walk(kind.Elem(), at+"[]")
		case reflect.Struct:
			for index := range kind.NumField() {
				if field := kind.Field(index); field.IsExported() {
					walk(field.Type, at+"."+field.Name)
				}
			}
		}
	}
	walk(reflect.TypeFor[Program](), "Program")
}
