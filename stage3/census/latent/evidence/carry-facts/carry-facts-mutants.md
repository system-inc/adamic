# Carried-fact mutants, mode-complete version

The same three required mutants are retained and rerun on this branch. `TestLatentCarryFactsAuditedFields` checks only the three audited carried fields, their exact types and generated capture/install/snapshot copy. It additionally requires the shared helper's owned installation clone, because fresh first-error units have no full-mode unit snapshot. Unrelated fields retain default snapshots. `TestLatentCarryFactsReplayBodiesMatchMain` still fingerprints the four normalized main-source query bodies.

The shell ran `/workspace/scratch/carry3-run-mutants.py`; tests themselves run no compiler or Go subprocess and call `t.Parallel()`.

1. Remove `entriesRecords: programFact` from the generator table; restore in `finally`.
2. Change `entriesProgramRecordsKnown bool` to `int` in `/workspace/scratch/carry3-type-mutant-lower.go`. A scratch test-source overlay redirects the guard's AST input only.
3. Insert `_ = 0` in `entriesProgramHasRecords` in `/workspace/scratch/carry3-body-mutant-entries_records.go`. A scratch test-source overlay redirects the fingerprint test's input only.

```text
carry3-table-mutant: guard_test.go:252: carried lowering.entriesRecords is missing from the programFact table
carry3-type-mutant: guard_test.go:252: carried lowering.entriesProgramRecordsKnown type changed: got int, audited bool
carry3-body-mutant: carry_replay_test.go:77: entries_records.go: entriesProgramHasRecords body changed: carry replay was proven against the old bodies; re-prove it with the subset byte-identity run in stage3/census/latent/evidence/carry-facts/report.md (got 9dcfaeeb08a7609f4ae51c555b705b3a8f8ae966c8a4c7623809d155fa1e896f, want d08d63eabd0fb9f08c9750de9bce09b724ecbaf5ec5c3b52dfdd19117100157a)

```

Logs: [table](guard-mutant.log.txt), [type](type-mutant.log.txt), [body](body-mutant.log.txt), [restored green guards](guard-green.log.txt). The complete census tests and vet passed afterward. No production `internal/` file was edited.
