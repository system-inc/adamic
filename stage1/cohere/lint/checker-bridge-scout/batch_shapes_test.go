package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"testing"
)

func TestBatchPlanAndGeneration(t *testing.T) {
	t.Parallel()
	q := batchQuestion{Operation: "intentional-error"}
	original := []plannedFile{{Path: "fixture.ts", Questions: []batchQuestion{q}}}
	frozen := freezePlan(original)
	original[0].Questions[0].Operation = "type-name"
	if frozen[0].Questions[0].Operation != q.Operation {
		t.Fatal("caller mutation changed frozen plan")
	}
	owner := batchOwner{Number: 2, Generation: 8, IDs: map[*checker.Type]uint64{}}
	id := owner.intern(&checker.Type{})
	if _, err := owner.operand(id); err != nil {
		t.Fatal(err)
	}
	stale := id
	stale.Generation--
	if _, err := owner.operand(stale); err == nil {
		t.Fatal("generation mutant survived")
	}
	foreign := id
	foreign.Owner++
	if _, err := owner.operand(foreign); err == nil {
		t.Fatal("cross-owner mutant survived")
	}
	unknown := id
	unknown.Local++
	if _, err := owner.operand(unknown); err == nil {
		t.Fatal("unknown ID mutant survived")
	}
	if a := owner.answer(nil, q); a.Error == "" {
		t.Fatal("question error was swallowed")
	}
}

func TestBatchExecutionAndReplay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	first := filepath.Join(dir, "a.ts")
	second := filepath.Join(dir, "b.ts")
	config := filepath.Join(dir, "tsconfig.json")
	must(os.WriteFile(first, []byte("export const a = true;\n"), 0600))
	must(os.WriteFile(second, []byte("export const b = false;\n"), 0600))
	must(os.WriteFile(config, []byte(`{"compilerOptions":{"noLib":true,"strict":true},"files":["a.ts","b.ts"]}`), 0600))
	canonical := ""
	for _, n := range []int{1, 2, 12} {
		p := batchProgram(config, n)
		plan := []plannedFile{}
		for _, path := range []string{first, second} {
			f := p.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
			node := probeNode(f)
			q := batchQuestion{Operation: "type-name", Selector: batchSelector{node.Pos(), node.End(), node.Kind}}
			plan = append(plan, plannedFile{path, []batchQuestion{q, {Operation: "intentional-error"}, {Operation: "id-name", FromQuestion: true, OperandQuestion: 0}, q}})
		}
		session := newBatchSession(p)
		generation := session.Generation
		answers := executePlan(session, plan, true)
		for _, file := range answers {
			if file[1].Error == "" || file[2].Error != "" || file[3].Error != "" || file[0].Value == "" || file[0].Value != file[2].Value || file[0].Value != file[3].Value {
				t.Fatal("error interrupted batch or ID route changed answer")
			}
			if file[0].ID.Generation != generation {
				t.Fatal("generation lost")
			}
		}
		replay := replayBatch(plan, answers)
		if canonical == "" {
			canonical = replay
		} else if canonical != replay {
			t.Fatal("N/completion order changed canonical replay")
		}
		if n == 2 {
			if answers[0][0].ID.Owner == answers[1][0].ID.Owner {
				t.Fatal("cross-owner fixture did not use two owners")
			}
			target := answers[1][0].ID
			routed := freezePlan(plan)
			routed[0].Questions = append(routed[0].Questions, batchQuestion{Operation: "id-name", Operand: target})
			rerun := executePlan(session, routed, false)
			idleOwner := executePlan(session, []plannedFile{{Path: first, Questions: []batchQuestion{{Operation: "id-name", Operand: target}}}}, false)
			if idleOwner[0][0].Error != "" || idleOwner[0][0].Value != answers[1][0].Value {
				t.Fatal("ID did not route to issuing owner absent from current file plan")
			}

			if last := rerun[0][len(rerun[0])-1]; last.Error != "" || last.Value != answers[1][0].Value || last.ID.Owner != target.Owner {
				t.Fatal("cross-owner question was not refused and re-asked on issuer")
			}
		}
		answers[0][0], answers[0][1] = answers[0][1], answers[0][0]
		if replayBatch(plan, answers) == canonical {
			t.Fatal("arrival-order mutant survived")
		}
	}
}
