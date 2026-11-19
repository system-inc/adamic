Mutant evidence for objects fixtures 06, 08 and 12's stage-0 records in status.json.

What moved: a named write of an optional field is now judged once lowering is done, by what reaches its receiver (reachHoldsSlot, internal/lower/slot_writes.go, #w8nanzw), and its NotYet names the allocation found without the field. 06_watch_close.a:25:17 (WatchedFileWithUnchangedPolls without isClosed), 08_loop_state.a:16:17 (ConvertedLoopState without argumentsName) and 12_decorator_descriptor.a:13:21 (MemberInfo without memberDescriptorName) stop at the same place for the same reason, "writing a possibly absent optional own field", and only their words grow by the allocation. The fixtures are unchanged.

Mutant: reachHoldsSlot's slot test answers true (`holds := true || holdsSlot(literal.Fields, name)`), on top of this change; the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixtures(Objects|Taste)$' -v
Result with the mutant: exit 1, red on exactly these three records and taste's 11 and 17, the stop then named by the own-type judgment ("..., whose layout holds no isClosed") (J-ftjkfmg, at 247d1ac7). Without the mutant both tests pass in the same job (control, exit 0).

    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/objects/06_watch_close.a:25:17: stage 0 can't lower writing a possibly absent optional own field, where a value reaching it is built as { fileName: string; callback: FileWatcherCallback; unchangedPolls: number; mtime: number; }, whose layout holds no isClosed yet Stdout:}
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/objects/08_loop_state.a:16:17: stage 0 can't lower writing a possibly absent optional own field, where a value reaching it is built as { loopParameters: string[]; loopOutParameters: string[]; }, whose layout holds no argumentsName yet Stdout:}
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/objects/12_decorator_descriptor.a:13:21: stage 0 can't lower writing a possibly absent optional own field, where a value reaching it is built as { memberDecoratorsName: string; }, whose layout holds no memberDescriptorName yet Stdout:}
            --- FAIL: TestFixturesObjects/objects/06_watch_close.a/stage0 (0.00s)
            --- FAIL: TestFixturesObjects/objects/12_decorator_descriptor.a/stage0 (0.00s)
            --- FAIL: TestFixturesObjects/objects/08_loop_state.a/stage0 (0.00s)
