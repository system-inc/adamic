Built three small guarded WASI hooks; the combined build remains blocked outside W7's territory.
SHAs: claim 2f8c5f3; concurrency merge fba6fd2; runtime-area merge b9052ec; hooks/evidence commit ad25e62; final log receipt SHA is reported on push.
Checks: 50/50 native objects identical; native oracle 514 pass; WASI before/after each 751 fail, 1 pass.
Mutants: native TLS global caught by TSan exit 66; pthread-create stub returns ENXIO; reverse WASI mutant unrun.
Not covered: green WASI execution, completed pool mutant, sibling async/moves/scaling, whole-repo gate or speed.

This is an incomplete worker result, not a landing. The latest user update lifted
the runtime-file push hold. The earlier pushed wasm/integrate claim history was
kept and origin/area/runtime bb48dc9 was merged into it, as instructed.
Only codex/wasm-threads is pushed; no PR and no area/main push.

The concurrency candidate was ced5530. No fetched branch contains all five;
the compiler is included in concurrency, while scaling, async and moves are
sibling extensions. Details and hook locations are in [plan.md](plan.md).
The exact three-file hook diff is [hooks.diff](hooks.diff).

| Check | Without W7 hooks | With W7 hooks |
| --- | ---: | ---: |
| Full WASI oracle, all leaf tests | 1 pass, 751 fail | 1 pass, 751 fail |
| WASI Node agreement fixtures | 0 pass, 375 fail | 0 pass, 375 fail |
| WASI emitted-C fixtures | 0 pass, 375 fail | 0 pass, 375 fail |
| Concurrency WASI leaves, included above | 0 pass, 26 fail | 0 pass, 26 fail |
| Strict native-package TestWASI | fails at unavailable pause | fails at share.c pointer hash |
| Native oracle | not rerun separately | 514 pass, 0 fail, 0 skip |
| Concurrency native variant leaves, included above | not rerun separately | 104 pass |
| Native release runtime object comparison | baseline | 50/50 byte-identical |

The remaining failing WASI leaf is TestWASIOracleCatchesMutants, also blocked by
the driver flag conflict. TestWASIRunnerCatchesMutants passes. Native semantic
agreement is checked with hooks; the all-object comparison independently proves
native code generation unchanged by those hooks. Parent test events are excluded
from these leaf counts. Every failure is named in [counts.json](counts.json).
Full raw outputs are the adjacent *.json.gz files; no tests were piped or stopped.

Both full WASI runs reject -pthread together with -mno-atomics, before compiling
the runtime. Concurrency's native.go flag addition caused this integration
conflict. Strict TestWASI bypasses that driver and independently exposes pause
before hooks and share.c's uintptr_t shifts by 33 after hooks. These are compile
failures, not oracle disagreements observed while executing wasm.

Permission was requested for two specific out-of-territory changes, with no
answer received at report time: target-guard -pthread in internal/native/native.go,
and WASI-only uint64_t widening at the top of seen_slot in runtime/share.c.
Neither file was edited. They are necessary to finish the requested WASI gate
and runtime mutants. No green landing or semantic mutant kill is claimed from
these failures.

The native TLS mutant did compile: existing field_cache drops _Thread_local
from the native field cache, and TSan catches a race in adamic_object_data_field
with exit 66. The real wasi-libc pthread_create stub probe prints
`pthread_create=6 entered=0`; ENXIO is returned and no worker runs. That probe
establishes stub behavior but does not complete the requested mutated runtime
startup check. Reverse WASI visitation is not executed because the application
build remains blocked. These two missing runtime mutant proofs remain mandatory.

[commands.md](commands.md) names exact commands, release flags, setup timing,
observations and limitations. [object-identity.json](object-identity.json)
contains SHA-256 pairs for every runtime object. Probe and actual runtime
disassemblies show plain memory operations for atomics and TLS on unshared wasm.
Vet passed with an empty log. The Threads section makes no speed claim.

Before landing, authorize the two narrow extra hooks, finish strict TestWASI,
re-run the full WASI and native oracles, execute reverse and pool-start mutants,
recompare every native object, merge current origin/main (no rebase), re-green,
and push only codex/wasm-threads for runtime-area integration.
