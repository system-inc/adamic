Built exact class-value constructor dispatch and lazy global/lexical caches; kept unsupported constructor lessons explicit.
Commits: 974c2bc2 class values; 6b0d705e Uint16 withdrawal; 968172dc global ||; final lexical-cache SHA accompanies the push.
Node/backend/sanitizer fixtures and touched-package tests are recorded in evidence; counts refreshed with one new row.
Six dispatch/order mutants, eager ||, Uint16 acceptance, and three lexical-cache mutants were killed by their fixtures.
Original root coverage is 0/22; all 22 were replayed, with remaining dependencies and design rulings below.

The latest standing rule governs this final push: once after the unit's own fixtures, tests and mutants. No unlanded worker branch was merged in this continuation. Current compiler-area base b68b2fe1 is present through merge 1133c74b. Checked .ts non-null c41c0e06 is present; views ffe428ab is not an ancestor. The attempted views merge was aborted earlier, so no revert was needed. The October 8 ruling keeps ! refused in .a.

## Implemented behavior

Registered non-generic source class values dispatch through the actual constructor object, with exact identity checks. Constructor evaluation precedes arguments, and both occur once. Unknown allocators trap instead of silently choosing a static allocator. Both ?? and || cache only Object-represented constructors; objects are truthy, so their nullish fallback is exact. Globals use a direct helper. Lexical caches use the existing closure representation with exactly one captured cache cell, preserving independent factories and cache lifetime. Only a global constructor read or a closed zero-argument function call is accepted as a lexical initializer. More general captures remain NotYet.

The new lexical fixture exercises escaped closures, independent factory cells, direct local caches, global reads and closed calls. The negative fixture is held to Node output 9 and pins the additional-capture NotYet before emission. No runtime C helper was added. The former uint16_array.c/.h pair and its lowering were removed in 6b0d705e; Node's 70000 -> 4464 conversion fixture pins the restored Uint16Array NotYet.

## Final original-kind disposition

| Kind | Roots | Disposition | Validated original roots |
| --- | ---: | --- | ---: |
| new an Identifier | 13 | Blocked: weak lifetime support (6), runtime-owned Uint16 integration (1), ordinary-function constructor receiver semantics (5), class-expression registration (1) | 0 |
| new a ParenthesizedExpression | 8 | Class caches implemented, but actual roots use ordinary-function constructors; NodeLinks additionally uses any, needing a language ruling | 0 |
| new a PropertyAccessExpression | 1 | Blocked: inspector host adapter; enclosing closure still stops replay | 0 |

These are named dependencies, not territory skips. Arbitrary function constructor bodies require allocation, receiver binding and JavaScript return replacement; ordinary calls cannot substitute. Doctrine currently confines this to class constructors/methods and refuses any. Weak collections need an owner-approved lifetime model; strong collections cannot substitute. SymbolLinks needs the class-expression registration owner's change. Runtime Uint16 must arrive via the compiler area; inspector.Session needs a real native host adapter. No unlanded dependency is merged, and no root is cancelled as an echo.

Fresh guarded census replay used the current production source with scratch measurement overlay and the same adapted 81-file corpus. Eighteen exact original stops reproduce. Uint16 reaches the restored named element-type NotYet. Binder/checker Symbol remain masked by __String; inspector remains masked by its enclosing closure. All commands, outputs and exits are in evidence/cache-local-roots.json. Replay exit zero means the stop reproduced, not that a site lowered. None of these checker-rejected census observations certifies a native program.

## Checks and mutants

New commands (all output redirected to logs):

- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_class_cache_local -count=1 -timeout 10m: passed 0.540s, Node against JavaScript/native/release and sanitizers.
- go test ./internal/oracle -run TestNewExpressionCacheCaptureRemainsNotYet -count=1 -timeout 10m: passed 0.185s.
- python3 stage3/notyet-new-expression/run-cache-local-mutants.py: eager initialization killed by Node stdout; missing cache store killed by allocator trap and Node exit comparison; removed capture proof killed by exact NotYet test (got nil before emission). Scratch overlays leave production unchanged; no compiler failure counted as a mutant kill.
- go test ./internal/lower ./internal/oracle -count=1 -timeout 30m: passed lower 43.395s and oracle 220.076s; output in evidence/cache-local-packages.log.txt.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts: passed 57.910s. Only new row: allocations/frees 49/49, retains/releases 36/78, peak 10, regions 0.
- python3 stage3/census/latent/make_overlay.py followed by go build -buildvcs=false -overlay=... ./stage3/census/latent/replay/worker; worker's exact 22 commands are recorded in cache-local-roots.json.

Earlier retained checks/mutants are documented in CLASS-VALUES.md, UINT16-WITHDRAWN.md and AREA-CONTINUATION.md: eager ??, no cache store, forced static allocator, constructor reread after arguments, repeated constructor and repeated argument all fail source Node output/exit comparisons; eager || fails stdout; accepting Int32-backed Uint16 fails its exact restored NotYet assertion. Historical Uint16 conversion mutants are superseded by withdrawal and make no claim about supported Uint16 semantics. No full repository gate ran. Setup took 231.369s with nproc 5 (quota 4); original timing lines remain in REPORT.md and evidence/setup.log.txt.

A proposed generic capture traversal was rejected by automatic approval review because it could infer an incorrect closure environment or lifetime. The implemented alternative requires a narrowly proven closed initializer and captures one known cell; its positive sanitizer fixtures and negative refusal mutant pass. This rejection leaves no pending approval request.
