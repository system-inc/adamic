# The carry guard's walk sha, moved by #ny0c3gb

Mutant: carryReplayBodies' entriesRecordLiteralWalk back to its old sha,
ee77f34596203a624dbeded6e2576020d94eb6c0208b3df2a879508d350ab68a.

Measured red at 32edac07 (this change's lowering, before the sha moved), Codex 01a1c15e,
`go test ./stage3/census/latent/... -count=1`:

    TestLatentCarryFactsReplayBodiesMatchMain
    carry_replay_test.go:77: entries_records.go: entriesRecordLiteralWalk body changed: carry replay was proven against the old bodies; re-prove it with the subset byte-identity run in stage3/census/latent/evidence/carry-facts/report.md (got 06ec77fe884d39399a712f143b37922d7c67eb7ec18ac46bbf1b3c715c6aa478, want ee77f34596203a624dbeded6e2576020d94eb6c0208b3df2a879508d350ab68a)

The other three bodies' shas are unchanged (Codex 01a1c1f3 computed all four at 32edac07).

What was shown before the sha moved (Codex 01a1c24d, 32edac07, the adapted tree at the pinned upstream 050880ce):
- Over the whole program, 1,258 carried entriesRecords answers; exactly 2 literals pass the new satisfies parent
  test (transformers/declarations/diagnostics.ts:616:48 and :631:36 adapted, 607 and 622 upstream), and both
  answers are false, so no carried answer moves in tsc.
- A census of diagnostics.ts alone: carried, uncarried and a satisfies mutant byte-identical in first-error
  mode (8 units, 58 findings, 7867d06a...), and carried and uncarried identical in full mode (35 units, 150
  findings, 1763a252...). A mutant setting all 11 true answers false was identical too, so that file's census
  can't see a wrong carry; it shows only that this change moves nothing there.

Owed: uncarried, carried and that mutant over the tsc.ts entry census on the landed sha, first-error mode, on
#3cvj5sk's box gap (Correctness), where the mutant has to differ.
