Preserved snapshot references for representations 14 through 17, satisfying runtime clearance for item 143.
One follow-up commit on compiler/fx7-snapshot; SHA is in the delivery report.
Focused oracle/native snapshot tests and TestCallTargetReaders pass; counts and lane results accompany delivery.
Dropping the new assignment fails the sanitized reference probe while the typed-array view diagnostic stays pinned.
No new mixed-union kinds or typed-array admission; no full gate run.

Choice: copy `payload.reference = slot->reference` for `storage >= adamic_rep_record`, keeping kind unknown. The mixed-union enum does not name record or typed-array kinds. Preserving the borrowed reference avoids changing reader admission or diagnostics.

One new top-level test, TestViewSnapshotTypedArrayReference, creates an Adamic source in its temporary directory. A Uint8Array field is erased through Root, then read through a string | number View. Emitted C must call adamic_object_view_union_snapshot. Node's control prints `object`. Native release and sanitized native keep exit 70 and the exact diagnostic:

```
adamic: panic: field read failed: viewed.value matches no member of string | number; expected string | number, found unsupported representation
```

The same test builds a sanitized fallback probe that checks kind unknown and pointer identity under all four counted-reference storage tags. It uses a real typed-array allocation; no unsupported heap kind is dereferenced by the snapshot. It prints `references preserved` and finishes without a sanitizer/leak report. This payload assertion is necessary because the existing diagnostic observes the tag alone.

Mutation: remove only the new reference assignment, run the new test with `timeout 90 go test ./internal/oracle -run '^TestViewSnapshotTypedArrayReference$' -v -count=1 -timeout 60s`, restore in finally. The view's pinned release and sanitized diagnostics still pass; the sanitized fallback probe exits 3 rather than 0, killing the mutant by comparison. No sanitizer memory error is needed to detect the lost pointer.

Commands and receipts:

- `timeout 120 go test ./internal/oracle -run '^TestViewSnapshotTypedArrayReference$' -v -count=1 -timeout 90s`: reference-test.log, pass; new leaf 15.04s including cold runtime build.
- Mutation above: reference-mutant.log, expected test exit 1, leaf 14.79s.
- `timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s`: reference-call-targets.log, pass 14.175s.
- `timeout 120 go test ./internal/oracle ./internal/native -run '^TestViewSnapshot' -v -count=1 -timeout 90s`: reference-final.log, pass; new leaf 0.62s, oracle package 0.633s, native package 0.240s.
- `timeout 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 9m -args -update-counts`: reference-counts.log. No persistent fixture was added; the new source is contained in the single test.

Evidence is observed output. Record semantics and each typed-array class's separate mixed-union admission remain outside this change; all four storage tags' pointer preservation is tested.

Counts refresh passed in 98.041s; counts.md is unchanged. Integration lane checks are run on the single committed follow-up and reported with its pushed SHA.
