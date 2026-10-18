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

## Host runtime storage audit

`adamic_start` calls `adamic_host_start` once, before user code can start
parallelMap's lazy worker pool. Thread creation and the pool scheduler publish
those startup writes. The startup API is not a worker operation; callers must
not restart the runtime while workers run.

| Static in host_runtime.c | Writers and timing | Readers and protection |
|---|---|---|
| saved_count | adamic_host_start, before the pool starts | adamic_host_argv; immutable during worker execution |
| saved_values | adamic_host_start, before the pool starts | adamic_host_argv; borrowed argv and its bytes live unchanged until process exit |
| identity_bytes | capture_executable_identity during startup; release_identity at exit | adamic_host_exec_path and its argv/executing-file callers; cleanup explicitly shuts down and joins the pool before freeing and clearing the pointer |
| identity_length | capture_executable_identity during startup | adamic_host_exec_path; immutable while workers run |
| identity_error | static initialization, then adamic_host_start on capture failure | adamic_host_exec_path; immutable while workers run |
| identity_saved | adamic_host_start, before the pool starts | adamic_host_start only; the once-only startup contract excludes concurrent calls |
| exit_lock | static mutex initialization; pthread mutex operations thereafter | setter and getter serialize both exit fields, including worker calls and natural-exit status reads |
| exit_present | adamic_host_set_exit_code under exit_lock | adamic_host_exit_code under the same lock; adamic_host_exit_status delegates to that getter |
| exit_value | adamic_host_set_exit_code under exit_lock | adamic_host_exit_code under the same lock; presence and value are one coherent snapshot |
| monotonic_origin | capture_clock_origin constructor before main and the pool | adamic_host_performance_now; immutable during worker execution |
| epoch_origin | capture_clock_origin constructor before main and the pool | adamic_host_time_origin; immutable during worker execution |
| backing_bytes | adamic_host_backing_add/drop, including allocations and final releases on workers | adamic_host_memory_usage; all accesses are relaxed atomic operations, providing a race-free ledger without publishing payload storage |
| eol | static ADAMIC_STRING initialization only | adamic_host_eol and string consumers; a one-byte ASCII immortal literal, so retain/release and indexing never mutate it |

The new raw writer in adamic.c also accesses existing inventoried output,
output_used and broken storage. It now holds output_lock across both flushes,
write_text, and the broken read, like main's line writer. The exit-code mutex
and raw-output lock repair unordered writes that could occur while workers run.
No new storage was introduced by 29a in adamic.c, adamic.h, heap.c or typed_array.c;
the new host_runtime.h contains declarations only.

TestRuntimeStaticsHostWorkers exercises startup observations, clocks, the atomic
ledger, exit status and raw output on the real pool. Independent lock-removal
mutants must compile and produce an explicit ThreadSanitizer data-race report;
crashes, timeouts and mutex-misuse reports do not count as race evidence.
