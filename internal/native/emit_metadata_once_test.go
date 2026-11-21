package native

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The three facts writeProperty asks for every property write answer from the program as the one walk
// read it: a record literal anywhere; an absent field or array metadata anywhere; a property read with a
// readiness or a view, by its name, or a name the lowering summarized as uninitialized. A program
// without them answers false for each.
func TestPropertyWriteFactsAnswerFromTheProgram(t *testing.T) {
	t.Parallel()
	plain := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "a", Value: ir.NumberConstant{Value: 1}}}}}}}}
	if plain.hasRecordStorage() || plain.constructionNeeded() || plain.fieldReadinessNeeded("a") {
		t.Fatalf("a plain object literal: records %t, construction %t, readiness of a %t, want none", plain.hasRecordStorage(), plain.constructionNeeded(), plain.fieldReadinessNeeded("a"))
	}
	record := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Record: true}}}}}
	if !record.hasRecordStorage() {
		t.Error("a record literal: no record storage")
	}
	absent := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "a", Absent: true, Value: ir.NumberConstant{}}}}}}}}
	if !absent.constructionNeeded() || !absent.fieldTypesNeeded() {
		t.Errorf("an absent field: construction %t, field types %t, want both", absent.constructionNeeded(), absent.fieldTypesNeeded())
	}
	metadata := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.ArrayLiteral{Metadata: []ir.Field{{Name: "pos", Absent: true, Value: ir.NumberConstant{}}}}}}}}
	if !metadata.constructionNeeded() {
		t.Error("array metadata: no construction")
	}
	read := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.Property{Object: ir.Read{Of: ir.Object}, Name: "kind", Readiness: "kind", Of: ir.Number}}}}}
	if !read.fieldReadinessNeeded("kind") || read.fieldReadinessNeeded("other") {
		t.Errorf("a read of kind with a readiness: kind %t, other %t, want true and false", read.fieldReadinessNeeded("kind"), read.fieldReadinessNeeded("other"))
	}
	viewed := &emitter{program: &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.Property{Object: ir.Read{Of: ir.Object}, Name: "kind", View: "Node", Of: ir.Number}}}}}
	if !viewed.fieldReadinessNeeded("kind") {
		t.Error("a read of kind through a view: no readiness")
	}
	summarized := &emitter{program: &ir.Program{UninitializedFields: ir.SortedMap[string, bool]{"late": true}}}
	if !summarized.fieldReadinessNeeded("late") {
		t.Error("a name the lowering summarized as uninitialized: no readiness")
	}
}

// Each fact is read once per emitter, never once per write: emission doesn't change the IR these facts
// read (prepareObjectMetadata), so an emitter that has answered keeps its answer while the program grows
// under it, and a new emitter reads the grown program. #67w3g0x's scout-link round 5 measured the
// per-write walks as writes × program. Mutants: hasRecordStorage, constructionNeeded or
// fieldReadinessNeeded walking the program again, each of which answers the grown program here.
func TestPropertyWriteFactsAreReadOncePerEmitter(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Main: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{}}}}
	asked := &emitter{program: program}
	if asked.hasRecordStorage() || asked.constructionNeeded() || asked.fieldReadinessNeeded("kind") {
		t.Fatal("an empty literal answered a fact")
	}
	program.Main = append(program.Main,
		ir.Return{Value: ir.ObjectLiteral{Record: true}},
		ir.Return{Value: ir.ArrayLiteral{Metadata: []ir.Field{{Name: "pos", Absent: true, Value: ir.NumberConstant{}}}}},
		ir.Return{Value: ir.Property{Object: ir.Read{Of: ir.Object}, Name: "kind", Readiness: "kind", Of: ir.Number}},
	)
	if asked.hasRecordStorage() {
		t.Error("hasRecordStorage walked the program again, a walk per write")
	}
	if asked.constructionNeeded() {
		t.Error("constructionNeeded walked the program again, a walk per write")
	}
	if asked.fieldReadinessNeeded("kind") {
		t.Error("fieldReadinessNeeded walked the program again, a walk per write")
	}
	fresh := &emitter{program: program}
	if !fresh.hasRecordStorage() || !fresh.constructionNeeded() || !fresh.fieldReadinessNeeded("kind") {
		t.Errorf("a new emitter of the grown program: records %t, construction %t, readiness %t, want all", fresh.hasRecordStorage(), fresh.constructionNeeded(), fresh.fieldReadinessNeeded("kind"))
	}
}
