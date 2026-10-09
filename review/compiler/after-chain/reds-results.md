Composed lowering-chain-reds fcc9df32 with the held self-compare static-inline fix ca082675 on compiler/after-chain.
Merge: 0071c42e; ca082675 was already an ancestor through b0ef560b, so the subsequent merge reported Already up to date.
Requested checks pass: full lowering, call-target readers, split/self-compare, backend stops, search shrink, inheritance witness and feature-set link guards.
Search skip-check and old-stop mutants, constant self-equality mutant, and deliberate runtime feature mismatches were caught.
Not covered: full gate, opt-in WASI execution, backend-stop external mutant runner or checker-archive execution; three archive-dependent split tests skipped.

All commands source /workspace/adamic-tools/env.sh. GOPROXY=https://proxy.golang.org|direct. Existing setup retained; nproc 5, quota 4. Every test writes its complete output to review/compiler/after-chain/reds-*.log.

Commands and observed results:

- timeout 90 go test ./internal/lower -count=1 -timeout 90s: PASS 70.067s during overlapping compilation; repeat timeout 60 with same Go flags: PASS 40.101s.
- timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s: PASS 48.344s, leaf 48.33s.
- timeout 120 go test ./internal/native ./internal/oracle -run '(?i)split|^TestSelfCompare|^TestRuntimeFeatureMismatch|^TestRuledBackendOutcomes$|^TestUnlistedBackendAgreement$|^TestTerminalStackStop$|^TestNonOverflowRangeErrorCatch$|^TestSearchShrink|^TestNativeAgreesWithNode$/internal/oracle/testdata/inherited_fields_guard.a$' -count=1 -v -timeout 90s: PASS native 1.202s, oracle 5.196s. Fourteen search fixture leaves pass; skip-check 0.38s and old-stop 0.47s caught. Self-equality mutant 1.49s caught NaN output mismatch. Named native stack pin preserved. Inheritance witness 0.50s passes. Inheritance lower guards are included in the full package run.
- timeout 90 go test ./internal/fuzz ./cmd/adamic-test262 -run '^Test(Fuzzer|Runner)FeatureMismatchStopsAtLink$' -count=1 -v -timeout 90s: PASS 8.082s and 8.746s. Each deliberately mismatched feature set fails at link with undefined adamic_runtime_features_closure_convention.
- Narrowed TestCountsAreRecorded/fixtures for node_buffer_digest_twice.a and pow_fractional.a with -count=1 -timeout 90s -args -update-counts: PASS 32.603s including additional registered counters. Both selected rows match incoming values exactly. Apply only those rows back to the full table.

The readiness conflict uses the entire reds branch test: source Node, emitted JavaScript and sanitized native are held to stdout and ruled exit 1, with no unexpected native stderr. The incoming fix supersedes the held agreeEntry resolution.

Counts merged without conflicts. Relative to pushed f130c846 the union adds node_buffer_digest_twice.a (5/5/7/10/3/0) and pow_fractional.a (6/6/0/6/3/0). Full counts verification then reported a table mismatch with no fixture numeric differences; canonical-order regeneration is recorded separately.

Integration lane checks: PASS 9.5s, gofmt and tools on 270 Go files, t.Parallel on 22 test packages, vet 22 packages. No production compiler edits were made by this integration unit; the only conflict resolution is the authorized readiness test.

Full canonical regeneration: timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts passed in 60.993s. Fixture-keyed comparison against the merged table found zero changed, added or removed rows. Only digest_twice and finalized swap order: node_buffer_test.go registers digest_twice in its initial loop and finalized afterward. Thus finalized moves position solely because the incoming new row must precede it; its 5/5/7/10/3/0 values remain unchanged. No existing numerical row moved. reds-counts-order.diff and reds-counts-comparison.txt record the audit.
