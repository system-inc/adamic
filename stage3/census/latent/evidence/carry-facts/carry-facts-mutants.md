# Carry fact and replay mutants

These commands are run by the shell, not inside any new Go test. All three new tests call `t.Parallel()`; their work is in-process AST parsing/formatting or generation. The unrelated-field control accepts a scratch lower.go with `futureMemo bool`.

## Removed carrier

Test: `TestLatentCarryFactsAuditedFields`. Remove only `"entriesRecords": programFact` from the generator table. The separate audited three-field contract keeps the omitted carried field red. The original generator source is restored in `finally`.

```sh
go test ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsAuditedFields$' -count=1 > /workspace/scratch/carry2-table-mutant.log 2>&1
```

## Changed carrier type in scratch

Copy main's lower.go to `/workspace/scratch/carry2-type-mutant-lower.go`, changing `entriesProgramRecordsKnown bool` to `int`. A Go test-source overlay redirects only the guard's AST input to this scratch path; no production source is edited. The guard reports the named type mismatch before generation, without panicking.

```sh
go test -overlay=/workspace/scratch/carry2-type-mutant-overlay.json ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsAuditedFields$' -count=1 > /workspace/scratch/carry2-type-mutant.log 2>&1
```

## No-op body edit in scratch

Copy entries_records.go to `/workspace/scratch/carry2-body-mutant-entries_records.go`; add `_ = 0` as the first statement of `entriesProgramHasRecords`. A scratch test-source overlay redirects the main-source fingerprint test's input to that file. Normalized AST hashing ignores comments/positions but includes the added statement, so it requires re-proof even though the edit is a no-op.

```sh
go test -overlay=/workspace/scratch/carry2-body-mutant-overlay.json ./stage3/census/latent/overlaytext -run '^TestLatentCarryFactsReplayBodiesMatchMain$' -count=1 > /workspace/scratch/carry2-body-mutant.log 2>&1
```

## Red lines and restoration

```text
carry2-table-mutant: guard_test.go:246: carried lowering.entriesRecords is missing from the programFact table
carry2-type-mutant: guard_test.go:246: carried lowering.entriesProgramRecordsKnown type changed: got int, audited bool
carry2-body-mutant: carry_replay_test.go:77: entries_records.go: entriesProgramHasRecords body changed: carry replay was proven against the old bodies; re-prove it with the subset byte-identity run in stage3/census/latent/evidence/carry-facts/report.md (got 9dcfaeeb08a7609f4ae51c555b705b3a8f8ae966c8a4c7623809d155fa1e896f, want d08d63eabd0fb9f08c9750de9bce09b724ecbaf5ec5c3b52dfdd19117100157a)
```

Logs: [removed carrier](guard-mutant.log.txt), [changed type](type-mutant.log.txt), [body edit](body-mutant.log.txt). The normal-source guards and the complete census test tree passed after restoring the table and using no test overlays. Current source and both binaries are pinned to the parent named in [report.md](report.md).
