Mutant evidence for taste fixtures 11 and 17's stage-0 records in status.json.

What moved: a named write of an optional field is now judged once lowering is done, by what reaches its receiver (reachHoldsSlot, internal/lower/slot_writes.go, #w8nanzw), and its NotYet names the allocation found without the field. 11_build_info_pending.a:7:12 (BuilderProgramStateWithDefinedProgram without buildInfoEmitPending) and 17_binder_flow.a:23:13 (ConditionalExpression without flowNodeWhenTrue, the binder's write the construction-window slot rule of rulings.md:78 is for) stop at the same place for the same reason, "writing a possibly absent optional own field", and only their words grow by the allocation. The fixtures are unchanged.

Mutant: reachHoldsSlot's slot test answers true (`holds := true || holdsSlot(literal.Fields, name)`), on top of this change; the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixtures(Objects|Taste)$' -v
Result with the mutant: exit 1, red on exactly these two records and objects' 06, 08 and 12, the stop then named by the own-type judgment (J-ftjkfmg, at 247d1ac7). Without the mutant both tests pass in the same job (control, exit 0).

    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/taste/11_build_info_pending.a:7:12: stage 0 can't lower writing a possibly absent optional own field, where a value reaching it is built as { hasErrorsFromOldState: true; hasErrors: false; }, whose layout holds no buildInfoEmitPending yet Stdout:}
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/taste/17_binder_flow.a:23:13: stage 0 can't lower writing a possibly absent optional own field, where a value reaching it is built as { condition: true; questionToken: number; whenTrue: number; colonToken: number; whenFalse: number; }, whose layout holds no flowNodeWhenTrue yet Stdout:}
            --- FAIL: TestFixturesTaste/taste/11_build_info_pending.a/stage0 (0.00s)
            --- FAIL: TestFixturesTaste/taste/17_binder_flow.a/stage0 (0.00s)
