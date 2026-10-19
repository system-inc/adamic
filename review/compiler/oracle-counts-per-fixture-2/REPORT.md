Re-cut the committed per-fixture counts migration, preserving all current main rows.
Base: dd5009aacdec4c871095e8d8068de2da8ac4992d. One commit retains the original message and Ahra trailer.
Validation: Linux clang 20.1.8, GOMAXPROCS=4, parallel=4; whole oracle and every requested check passed.
Mutants: missing record and orphan caught by census; changed count caught by recorded-counts shard; CRLF writer caught by determinism.
Not covered: actual macOS execution; no compiler or runtime production code changed.

Only conflict: counts.md modify/delete. Main’s entire table was migrated, including seven new uncaught-error rows, then the aggregate was deleted. Main’s header is verbatim in the counts README. No values changed relative to main.

```
runtime: 1120/1120 keys and every value identical; row bytes identical
predicates: 15/15 keys and every value identical; row bytes identical
TOTAL: 1135 rows; 0 missing, 0 extra, 0 changed values
```

The Linux -update-counts run left all 1,135 files byte-identical. Determinism and census passed. The existing mirrored runtime/predicate layout and registry exclusions are retained. No additional implementation assumption was needed.

Commands and wall seconds (exact arguments in commands.json):

- update-counts: PASS, 108.528 s.
- equality-proof: PASS, 0.064 s.
- determinism-census: PASS, 17.217 s.
- oracle: PASS, 461.14 s.
- mutants: PASS, 26.278 s.
- ir-guard: PASS, 42.132 s.
- parallel-guard: PASS, 0.622 s.
- vet: PASS, 0.776 s.
- gofmt: PASS, 0.008 s.
- equality-final: PASS, 0.114 s.

The oracle package itself reports 459.595 s. TestCallTargetReaders, TestEveryMapAProgramHoldsIsSorted and TestEveryTestIsParallelOrSaysWhy passed. Audit: 5310 test identities, 0 unresolved failures, 0 leaves at or above 60 s; maximum leaf 24.66 s.

Setup: the initial shared-cache helper stalled and was terminated. Retrying with ADAMIC_GOCACHE_OFF=1 passed in 37.020 s; nproc=5, cgroup quota=4 CPUs. Timing lines are in setup-output.txt. Ordinary oracle observation caches were permitted, as in the original unit.
