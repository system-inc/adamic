# Paired overload certificate census — #5fws6q8

In progress. Base: `910f21afffd892c747bd82f72229743a899586ea`.
Head: `812c342689ded7e1cc8a31c35a86f3546f7b0772`, exactly one commit on base.
No adaptation, ledger, or internal change is made by this report branch.

Both measured worktrees initialized recursive submodules before any build:
cohere `5360753e19e4417675a90e92a1823bea66b13918`, TypeScript-go
`d92d9bfee114c80be2c375d72edae966176e3a4f`.

`git diff BASE HEAD -- stage3/apply.sh stage3/apply.py stage3/adapt stage3/api
stage3/source.json` is empty. One unchanged base-adapted tsc tree is being
prepared for both. Each sha independently supplies its latent generator,
meter entry overlay, scan-all generator, compiler sources and binaries.

The first-error entry pass will match the meter: tsc.ts root,
CommonJSBundle, LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1.
Each full run selects exactly one absolute file with LATENT_ONLY_FILE and uses
LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1. The caller list is resolved
with getResolvedSignature from the entry closure and complete src tree. Files
inside the entry closure retain that same tsc.ts root; extra src callers use
their own entry closure, so they are measured rather than silently omitted.

No declaration-pass count, certified-call count, or other-overload conclusion
is established yet. The final report will retain child exits, complete streams,
position-stripped finding tuples, reached-stop distinctions, source hashes,
and probe build provenance. A blocked/incomplete run will be labeled as such.

## Preparation checkpoint

Both shas' entry, normal latent and scan-all builds completed, with SHA-256
provenance under builds/. Both apply keys are
`6c54c265b3ddf357d631b5dc1bb7f13cf328782bdf8f8e9db38e24a5aec7fae4`.
The shared adapted tree is at the upstream pin and its source fingerprints are
retained. SETUP.json and callers.json record the exact source/program scope.

The `tsc.ts` checker closure contains 81 sources and **11 mapped calls**, not 17.
The prior scout's compiler/services domain contains the requested **17** (the
other six are service calls outside the entry closure). Whole-src resolution
finds 129 calls, 27 mapped, across 47 files. All 47 exact files are scheduled,
including core.ts. The 17-call answer will be reported separately from the extra
server/harness/testRunner calls.

The completed core.ts full pair has 14 findings on each side and no changes.
Its generic declarations are deferred: this absence is not being counted as
proof that their overloads pass. A separate normal-overlay compiler-directory
core pair agrees (14/14), and remains a supplemental scope, never substituted
for the entry-root pair.

Execution note: entry and full batches run concurrently, each with two workers.
The initial coordinator's main thread was interrupted to prevent it scheduling
a duplicate full batch; its already-running entry children continue, write their
own receipts and return their own exit codes. The full batch has its own
coordinator/manifest. This is scheduling only, not a cancellation or alteration
of a compiler probe. Any coordinator exit is recorded separately from probe
exits. The final report will require all stream trailers and exact selections.

## Completed entry checkpoint

Both first-error tsc.ts entry probes exited **0**, with full relation trailers,
CommonJSBundle, identical checker-clean headers, and no relation overflow at
load or end. Findings: **4,488 -> 4,485**. Grades: **6 gone, 3 reached,
0 new, 0 changed**. There are **19 non-target overload sets / 67 findings**
in the entry ledger, all unchanged. The full caller-file survey is incomplete,
so this is not yet an overall answer for the additional contexts.

The six removed observations are arrayFrom overload 1 at four caller units and
arrayToMap overload 2 at two caller units. The three newly reached stops are
the arrayFrom/arrayToMap storage refusals in builder and a ToPrimitive/String
conversion stop in executeCommandLine. None is graded as a fresh regression
in a previously unblocked unit. Entry raw streams, timings, receipts and the
planned scheduler-interruption log are retained under entry/. The probes
completed normally; the scheduler itself exited 130 after joining them.

## Exact-file checkpoint

The compiler caller-file pairs have completed. No non-target ordinal or non-ordinal overload movement has appeared in the completed pairs. Three of the 17 compiler/services mapped calls have an observed certified-storage stop so far; this is provisional until services finish. Both revisions panic while preparing computed accessor names in six harness-file contexts; their child exit codes are 2, and no complete-stream conclusions are drawn from these failed attempts. Server/services/testRunner and scan-all runs continue.
