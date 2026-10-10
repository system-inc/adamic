Built step (3): callable adapter identity and bounded re-viewing in both backends.
Commits already pushed: cache revision 19df538d4, adapter execution ed82a09a2; this report's commit completes step (3).
Tests: oracle 5.887s; focused lower 1.016s, native 6.549s, JavaScript 1.125s.
Mutants: unequal escaped reads and unconditional wrapping caught in both backends; actual runtime root-normalization omission caught by source Node stdout in 14.548s.
Pending: function surface, actual later-call-site blame, generic relations beyond existing refusals, remaining lane 5 negatives, runtime conditions 3/4.

Native closure ===/!== and boxed closure equality normalize through adamic_view_adapter_underlying. Object.is shares boxed equality. Map/Set hash and compare callable keys by root; stored keys remain their checking adapters. Direct calls through re-viewed fields unwrap before selecting the producer signature. Ordinary object, scalar, string and collection identity behavior remains held by the relevant controls.

JavaScript equality/Object.is use the same underlying identity. Callable/boxed-key Map and Set constructors select specialized builtin subclasses whose lookups normalize identity but whose keys/iteration retain the first stored checking adapter. Updates preserve that key as Node does; deletion uses the root; iterators and forEach preserve live collection behavior. Set metadata is private, with no enumerable implementation field. Other key representations continue to use ordinary builtins. The WeakMap/WeakRef interning boundary already unwrapped adapters; repeated views have one adapter layer and one live adapter for the shared function/view type.

Enabled TestV4EscapeAdapterIdentity (1.15s) and NoStacking (1.97s). Identity pins source Node equality for repeated same-view reads, different view types and the producer, Object.is, Map updates/lookup/delete, Set duplicate keys/lookup/delete, and calling iterated keys. Additional controls: ConstructorIdentity 0.60s (Array equality under ASan; underlying identity is also directly tested in the runtime C fixture), CollectionKeyContract 0.74s (iteration retains argument checks), ScalarIdentityControls 1.60s (existing Object.is/NaN/signed-zero and Map/Set key fixtures). Array as a value in Object.is still has its existing lowerer NotYet, so that unlanded surface was not broadened. All new leaves below 60s. The fixture-path and constructor-control errors were corrected before the final passing run.

Native counted/sanitized re-view rows (allocations frees retains releases peak regions):
0: 3 3 7 9 3 0
1000: 1003 1003 5007 4009 4 0
2000: 2003 2003 10007 8009 4 0
Only the fresh holder object allocates each iteration; adapter allocation stays at one, peak stays four, and each iteration adds five retain/four release operations. Unconditional wrapping produces 2003 allocations at 1000 iterations instead of 1003, while matching source Node stdout and finishing ASan/LeakSanitizer clean. JavaScript factory observation requires 1; unconditional wrapping yields 1001, with ordinary stdout still matching Node.

The native comparison omission in generated code makes the two distinct-view escaped values unequal without touching checks or calls. The JavaScript root-identity omission makes equality, Object.is and collection keys disagree with source Node. In addition, run-identity-root-mutant.py temporarily removes normalization from the actual closure.c accessor, requires the oracle to fail only by stdout, and restores closure.c byte-for-byte; its patch and logs are adjacent. The actual runtime omission independently exercises native boxed/Object.is and Map/Set identity. Positive runtime tests and all seven cache mutation categories were rerun in the focused native command, including the constructor overread, dying adapter, removal order and weak-leak-root cases.

Commands:
- ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^Test(V4Escape|V4Direct|ViewAdapterBaselineCounts)' -count=1 -v -timeout 90s
- timeout 90 go test ./internal/lower ./internal/native ./internal/javascript -run '^Test(ViewAdapter.*|ViewCallable.*|ViewCallables.*|PrepareViewCallableRead|CallableNamespace.*|RuntimeReleasePaths|ClosureConventionRuntimeFeaturesIgnoreLiterals)$' -count=1 -v -timeout 90s
- timeout 90 python3 review/compiler/views-v4/run-identity-root-mutant.py (expected inner oracle failure; wrapper passes only on semantic kill and source restoration)

Recorded counts rows for closures.a, method_closures.a, closure_convention_receiver_rest.a remain unchanged. No persistent oracle fixture was added, so counts.md has no new row. Runtime conditions 3 and 4 retain their exact pending tests/comments. No area/runtime dependency imported and no aggregate, array-write, callback-widening or checked-wider-write scope added. Step (2)'s separate vet run passed after lane checks skipped vet at their budget. Step (3) lane checks and separate vet run after this commit, logged alongside this report.
