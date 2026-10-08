Merged checked non-null c41c0e062e99da37820f822968d4df1b48cdaee7 and current main 749a69adbfae2a7bf22c1f0436d9ef73345c3070.
Code groups are b50442f2 (Uint16Array) and 974c2bc2 (class constructor values); integration merge is c47cddda9bc10ec8c50d27de800f94c90045aa0d.
Rebuilt guarded census replay and replayed all 22 original roots; counts refresh passed 92.633s and changed no row.
All ten earlier independent mutants were killed by source Node comparisons; no changes to their implementation paths in this merge.
The original coverage remains 1/13 Identifier, 0/8 ParenthesizedExpression, 0/1 PropertyAccessExpression; the remaining roots retain explicit stops.

The only merge conflict was counts.md; both sets of rows were retained. The subsequent update-counts run made no change. No existing test was changed by this worker to accommodate the merge.

The former NonNullExpression at utilities.ts:10514:46 no longer appears. The next newly exposed stop is utilities.ts:10518:17, a BinaryExpression with a number and a boolean; the exact next-stop assertion exits 0 (evidence/nonnull-next-stop.log.gz). The same selected unit also reports the earlier direct-case declaration, binary-expression statement and a number/string binary at 10523:23. None is attributed to constructor lowering. At sys.ts:1964:12 the non-null stop is gone, leaving the earlier closure at 1469:5; the inspector allocation is still masked. Asserted old non-null signatures exit 1 because they disappeared, not because the whole unit lowers.

All 22 replay commands and findings are recorded in evidence/nonnull-roots.json. Original constructor signatures still reproduce for the six weak collections, SymbolLinks and all eight parenthesized sites. Ordinary allocator sites may be masked by __String, checker signatures or dependencies; no masked site is counted as cleared. The NodeLinks function body still reproduces this outside a method. The class-value oracle demonstrates correct general syntax dispatch but the table's ordinary function constructors require a separate receiver/initialization language ruling. SymbolLinks is a class-expression binding requiring its owning expression-dispatch worker. Weak collections require weak-key/ephemeron ownership design; inspector requires a real host adapter. No stronger-map or empty-inspector substitution was made.

Commands sourced /workspace/adamic-tools/env.sh and wrote output to logs:
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/new-expression-nonnull-overlay
 go build -buildvcs=false -overlay=/tmp/new-expression-nonnull-overlay/overlay.json -o /tmp/new-expression-nonnull-replay ./stage3/census/latent/replay/worker
 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m
 go test ./internal/ir ./internal/native ./internal/javascript ./internal/flow ./cmd/adamic -count=1 -timeout 30m
 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts

Runtime owner review remains limited to the two new independent files internal/native/runtime/uint16_array.c and uint16_array.h, introduced by b50442f2. Existing runtime C files remain unchanged. Initial setup timing was 231.369s; nproc 5 (CPU quota 4). No full repository gate or PR.

Post-merge lower and oracle package tests passed in 34.276s and 348.762s. The counts refresh passed in 92.633s with no diff. Full source Node oracle includes the new constructor and Uint16 fixtures and the imported checked non-null fixtures.
Other post-merge packages passed: ir 46.868s, native 402.877s, flow 185.935s, cmd/adamic 14.655s; JavaScript compiled with no test files. Evidence includes package outputs and individual mutant failure logs. All 22 roots replayed; 18 old constructor assertions reproduced, three remained masked, and Uint16Array reached the next named binary stop.
