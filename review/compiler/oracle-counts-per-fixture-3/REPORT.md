Refreshed the same committed per-fixture counts split on current main.
Base: 23ff023b9ed986834d20404d43e2c363667edf9b. One commit retains the original message and Ahra trailer.
Validation: whole internal/oracle, regeneration, equality, census, determinism, guards, vet and gofmt passed.
Mutants: missing and orphan records caught by census; changed count caught by recorded counts; CRLF writer caught by determinism.
Not covered: actual macOS execution. No compiler or runtime production code changed.

Only conflict was counts.md modify/delete. Main’s entire current table was carried into per-fixture files before removing the aggregate. All 30 rows added or changed since dd5009aa are listed in main-rows-carried.txt; no values differ from the new base. The README retains main’s header verbatim.

```
runtime: 1147/1147 keys and every value identical; row bytes identical
predicates: 15/15 keys and every value identical; row bytes identical
TOTAL: 1162 rows; 0 missing, 0 extra, 0 changed values
```

Linux clang 20.1.8 -update-counts left all 1,162 files byte-identical. The existing mirrored layout and documented registry exclusions are retained. No new assumption was required.

Commands (full arguments in commands.json):

- update-counts: PASS, 123.425 wall seconds.
- equality-proof: PASS, 0.115 wall seconds.
- determinism-census: PASS, 21.428 wall seconds.
- oracle: PASS, 528.062 wall seconds.
- mutants: PASS, 29.638 wall seconds.
- ir-guard: PASS, 60.827 wall seconds.
- parallel-guard: PASS, 0.619 wall seconds.
- vet: PASS, 0.822 wall seconds.
- gofmt: PASS, 0.016 wall seconds.
- equality-final: PASS, 0.064 wall seconds.

Oracle package: 526.209 seconds. Audit: 5452 test identities, 0 unresolved failures, 0 test leaves at or above 60 seconds; maximum leaf 24.49 seconds.

Setup passed using the local Go cache in 47.542 seconds. nproc=5, cgroup quota=4 CPUs; GOMAXPROCS=4, parallel=4. Exact timing lines are in setup-output.txt. Ordinary oracle observation caches were permitted.
