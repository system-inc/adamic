# Deterministic runtime statics race fixtures

Normalization and exit-flush checks instrument temporary runtime snapshots. Guarded
and mutant snapshots use identical fixtures and relaxed atomic gates. Shipping
runtime synchronization is unchanged by these hooks.

`normalization_race.c` selects two callers. The first performs a real cache write
and pauses before returning. The second waits for that write, accesses the same
cache slot, then releases the first. Hooks cover both writes and cache-hit reads.
Both callers use identical input so a wrong-result abort cannot preempt the race
report. NFC of decomposed `é` exercises classes and pairs; NFD of composed `é`
exercises mappings. Controls retain thread-local caches; each mutant removes only
its target cache's thread-local declaration.

`exit_race.c` pauses one real worker after `buffer` increments `output_used`,
while it owns `output_lock` and before it appends the newline. The caller waits
for that writer and exits. The guarded finish releases the writer before normal
pool shutdown and the locked flush. The mutant removes shutdown and locking,
flushes while the writer is held, and only then releases it.

The gates use lock-free relaxed atomics, forcing the runtime accesses without
creating a happens-before edge between them. Each normalization unit and each
exit-flush variant runs three fresh processes at each of four and sixteen
workers. Mutants require unsuccessful exit and an explicit TSan data-race report
naming the intended function. A plain abort, timeout, compilation failure, or
unrelated report cannot satisfy the check. Controls require successful exit,
expected output, and no sanitizer report.

The shape registry fixture separately gives each of four workers fresh metadata
in each of 64 barrier rounds. Distinct entries keep all workers on the push path;
sharing an entry could let a late worker acquire its registered flag and skip
the registry access. Dynamic property reads still check for lost pushes and
self-linked entries.

Run `TestRuntimeStaticsProtectionMutantsUnit00` through `Unit10`,
`TestRuntimeStaticsSignalAndExit`, and `TestRuntimeStaticsAreListed` in
`internal/native`. The normalization units are `Unit08` through `Unit10`;
`TestRuntimeStaticsParallelUnit11` is the clean shape registry control.
