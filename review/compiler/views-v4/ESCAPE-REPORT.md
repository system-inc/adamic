Built the compiler half of V4 point (b): escaping callable demand and complete invocation contracts, with .a relation refusals.
Commit: this compiler-half commit on compiler/views-v4 above 2466284b; no runtime precision is claimed complete.
Tests: oracle 6.842s, lower 0.806s, native 2.982s, JavaScript 0.843s; 24 passed leaves and four explicitly pending adapter leaves.
Mutants: six compiler-boundary omissions caught; each admitted clean Node-equivalent native release, ASan/UBSan and JavaScript execution before failing the boundary assertion.
Not covered: escaping adapter execution, adapter check omission, weak caching, identity/key equality, no stacking and bounded allocation/retain counts.

The user explicitly allowed stopping at the compiler half when the required weak cache primitive is absent. The missing runtime API belongs in internal/native/runtime/closure.c and its declarations/lifecycle wiring in adamic.h and heap.c. Existing weak.c offers weak handles and death invalidation; it has no adapter interning table. Existing adamic_closure_canonical caches by activation/environment and code, not by underlying callable and view type. There is no adapter unwrapping/identity primitive. Neither existing mechanism establishes the required function/view-type cache, its non-owning lifetime, or no stacking. This is an implementation dependency, not a claim that a weak cache is impossible.

The compiler now distinguishes an immediate callee (including parentheses/type-only wrappers) from an escaping read. Each escape retains its independently built finite parameter/result domains, read site, existing view-type identity, and source-language policy. Demand is decided by the existing allocation graph after all bodies, so an unrelated ordinary read is not poisoned by a checked field name. A proven .a escape preserves the raw ABI and requires every known producer to establish the parameter and result relation. Missing/opaque producers and unsupported domains do not count as proof. Conservative finite-domain relations keep mutable object and array payloads invariant. Unproven .a escapes refuse with read path and fix. Demanded .ts escapes report explicit NotYet naming the missing cache, instead of reaching the old read-time signature rejection or an uncached wrapper.

The two proven-read controls exposed pre-existing emission omissions: native reads lacked the view_callables.h declaration include, and program JavaScript emission omitted the existing shape-check runtime. Both are fixed. No callable check was deleted and no producer metadata was fabricated. No runtime cache or adapter was implemented.

Commands, all test runs logged without piping:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 90 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
timeout 90 go test ./internal/oracle -run '^TestV4(Escape|Direct)' -count=1 -v -timeout 90s
timeout 90 go test ./internal/lower ./internal/native ./internal/javascript -run '^Test(ViewCallable.*|ViewCallables.*|PrepareViewCallableRead|CallableNamespace.*)$' -count=1 -v -timeout 90s
timeout 90 python3 review/compiler/views-v4/run-escape-mutants.py
```

Setup completed in 36.664s; clang ready 0.158s, go build ready 36.512s, cache warm 36.637s. nproc=5, cgroup CPU quota=4. The new active leaves ran in: ArgumentRefusal 0.53s, ResultRefusal 0.47s, PassedRefusal 0.54s, ReturnedRefusal 0.53s, MixedProducerRefusal 0.49s, Compatible 1.60s, TypeScriptBoundary 0.61s, OrdinaryRead 1.60s. All are top-level parallel tests. No persistent fixture inventory rows were added; counts.md remains unchanged.

Guard mutants cover unproven argument/result relations, argument-position escape, returned escape, mixed producer flow, and bypass of the .ts NotYet boundary. Patch sources and output logs are adjacent. The original compiler source was restored byte-for-byte after each mutant. These prove compiler guards, not the pending adapter runtime checks. The existing direct-call argument/result omission mutants also ran in the final oracle command.

Pending tests use these exact reasons:

- TestV4EscapeAdapterArgument: awaits compiler/views-v4: weak adapter cache per underlying function and view type, then escaped argument checks and their omission mutant
- TestV4EscapeAdapterResult: awaits compiler/views-v4: weak adapter cache per underlying function and view type, then escaped result checks and their omission mutant
- TestV4EscapeAdapterIdentity: awaits compiler/views-v4: weak adapter cache and underlying identity for ===, Object.is, Map and Set keys, with unequal-read mutant
- TestV4EscapeAdapterNoStacking: awaits compiler/views-v4: adapter unwrapping or composition and bounded allocation and retain counts, with unconditional-wrap mutant

Native/JavaScript proven controls match Node and pass finishing leak checks. Refusal probes run successfully in Node but deliberately refuse in .a; the .ts probe runs successfully in Node but remains NotYet. No new runtime divergence is claimed. Integration lane checks and separate vet run after committing; their local logs are escape-lane-checks.log and escape-vet.log. This unit does not implement later V4 points or lane 5's pending adapter rewrite.
