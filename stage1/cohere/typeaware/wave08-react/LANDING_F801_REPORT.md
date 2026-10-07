Built: existing wave 08 work rebased without conflicts onto current main f8013f0ba, with fresh compiler and checker archives.
Commits: tested rebased tip e0989ba45; this report accompanies the fresh landing evidence commit.
Checks: six complete rules, supported Globals subset, two React kernels, nine manifests/declarations, ownership and filtered Node oracle green.
Mutants: all completed-rule, Globals, kernel, numeric subscription, manifest, raw-fact/state and bridge ownership mutants caught again.
Uncovered: full React parity and numeric handed-node migration remain blocked on shared JSX/HIR/SSA/parser/driver APIs; no new claims or full repository gate.

The prior published unit tip was 3a53f4824. Main advanced from e8ba3d5d to
`f8013f0baac41ddc340d76f83bddde38536a8f07`, including lowering and runtime fixes.
The branch was rebased before any new claim. The stage0 compiler and both normal
and ASan/UBSan checker archives were rebuilt from the rebased source using the
owned scratch bridge overlay. No shared production source was edited. The final
remote check still showed f8013f0ba as main and 3a53f4824 as the unit branch.
The rebase push uses that explicit remote tip as its lease and targets only
codex/typeaware-wave-08. Rebasing was explicitly requested by the user.

## Fresh gates

The original `go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v
-timeout 30m` passed in 242.645 s with compiler/repository corpus environment
variables set. Normal and sanitized controls have 31 findings; compiler 77 has
28 findings / 20,884 identical bytes; repository 287 has two / 19,558 identical
bytes. Loop, require and cycle mutants compile, exit zero with empty stderr,
and are caught only by production byte comparison at 51, 7383 and 12195.
The retained released-registry mutant exits zero instead of the expected panic
70. Symbol ambient flags, type-node classification and resolved-module target
mutants compile and fail direct checker comparisons.

`validate_process.py` matches unmodified Go on all 100 upstream programs
(129 process-exit and 21 blocking findings), compiler 77 and repository 287,
normal and ASan/UBSan. The callee mutant is caught in six programs, first byte
164; blocking order in nineteen, first byte 1690. Both compile and exit zero
with empty stderr. All new checker questions reject released handles before
output, exit 70. `validate.py` passes race controls 22 / 13 findings, both
corpora and sanitizers; its compiling lost-timeout mutant is caught by byte
comparison, and its raw questions reject released handles.

Globals `validate.py` matches Go on 35 controls / 34 findings and both corpora
normally and sanitized. Its compiling ancestry mutant is caught at byte 9162.
The JSX witness still receives a Go finding while the native parser refuses
before output. `validate_cores.py` matches 52 sanitized native results against
unmodified Go kernels, catches compiling join and dependency-count mutants by
kernel bytes, and proves both unsupported production entry points refuse with
exit 70. Those two kernel mutants are not complete production rule mutants.

`validate_manifests.py` passes all nine rule.json declarations against the
sanitized native numeric exports and independent Go AST enum. It catches a
valid NewExpression-for-CallExpression manifest substitution by comparison.
The nested numeric validator also proves the 214-to-215 native declaration
mutant compiles, exits zero with empty stderr, and fails only numeric-byte
comparison. `validate_pending.py` passes returned-flags, resolved-target and
ancestor-name raw mutations against checker witnesses; sanitized native event
expectations match Go, and compiling write/blocking-state mutants fail those
expectations.

`GOFLAGS=-overlay=/workspace/wave08-f801/bridge/overlay.json go test
./bridge/tsgo ./bridge/tsgo/checker -count=1 -v -timeout 30m` passes: bridge
156.365 s, checker 0.418 s. It proves 100 ABI queries and retained outputs,
stale/zero handle refusal, 162-position Go/native oracle, ASan/UBSan/LSan, and
all existing foundation mutants. Input/output length mutants trigger ASan;
stale handle assertion catches a retained handle; wrong-position type differs
at byte 6; removed link guard fails refusal; omitted output free and incorrect
region ownership trigger LeakSanitizer. The owned raw Go package and vet pass.
Filtered `TestTheOracleCatchesOneByte` passes in 4.599 s with native and Node
cache misses. All test output was captured to files.

## Timing and limits

After concurrent validation finished, three alternating fresh-process rounds
measured the process profile: compiler native 4.200928 s versus Go 0.436773 s
(9.618x); repository native 0.573113 s versus Go 0.307122 s (1.866x). These
are observations; native is still slower in these end-to-end runs. Earlier
setup timing remains 132 s, nproc 5 with four-core quota; setup was not rerun.

Shared parser nodes still expose string kind, and the type-aware driver still
calls whole-file run methods. Numeric exports and rule.json subscriptions are
prepared, but legacy relevance comparisons/refetches cannot be removed safely
without the shared handed-node/numeric parser contract. JSX parsing and native
source-to-React-HIR/SSA remain unavailable; recent HIR files on other origin
branches are prepared-graph models, not source adapters. No batch 8 Diagnostic
integration SHA was supplied. No further claims were taken.

No full repository Go gate was run. The touched packages and filtered independent
oracle were run; corpus paths remain the branch's frozen 77 compiler and 287
repository manifests. Implementation and scope are unchanged by this rebase.
Earlier reports remain historical evidence for their named bases. Fresh logs,
full compressed base64 output streams, result metadata and the rebase log are
in landing-f801-validation/. Validator invocation forms are unchanged from
LANDING_REPORT.md and the listener reports, using /workspace/wave08-f801/adamic,
the rebuilt /workspace/wave08-f801/bridge archives, and the complete
/workspace/wave08-process-full fixture capture.
