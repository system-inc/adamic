# Scoped-tail mutant evidence

The unchanged positive guard is
`TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg`. It loads a small source
fixture, lowers an admitted discriminated-union assertion and its
guarded union-field read, and calls the real census unit-end hook with
`LATENT_TAIL=1`. The guard runs in an isolated child of the current test binary;
it does not build a compiler. Its fixture execution takes less than one second.

| Mutant | Changed line | Case made red | Observed result |
| --- | --- | --- | --- |
| Copy finalized state back | After `probe := l.latentSnapshot()` in the scoped-tail implementation, add `defer func() { *l = *probe }()`. | `TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg` | Child exits 1: `tail copied mutations into unit/base state`. |
| Bypass the enabled tail | In `latentFinishUnit`, replace the `LATENT_TAIL` condition with an unconditional return. | `TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg` | Child exits 1: `missing tail finding: []`. |

Both mutations were applied to separate scratch copies of the actual overlay
source; neither is retained in the implementation. The restored guard exits 0.
`ScopedTailReportTestsLoom03dg.test_tail_stop_does_not_change_body_totals_loom03dg`
holds the separate meter line and verifies a tail NotYet contributes zero to the
body-lowering NotYet count and does not make the unit whole-program measured.

`TestLatentScopedTailEntryFindingLoom03dg` also runs the fixture through the
real first-error census and checks that the body remains attempted while the
finding names the tail pass. It runs automatically in the overlay package suite.

`RootOrderTests.test_artifact_tail_stop_preserves_body_count_loom03dg` holds the
same separation for root-order artifact body counts and retains its tail stop
as separate entry evidence.
