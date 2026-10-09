# Area-stack parser rerun

Fresh scratch starts at origin/compiler/area-stack 2c7d9fd22bdd5f2d2d9cda0a892b7ec734c3d758.
Library tip 3398253013be45d6e91ccca48c6f1a258d0ebe0d contains e40216f2 but conflicts
in 45 paths, including lowering, loading and runtime. The merge is aborted as
instructed. No scratch compiler merge or compiler source edit is pushed.
Go build -p 1 -mod=mod succeeds with cohere replacements redirected to the
existing pinned 7945d102 SDK. nproc=5; source /workspace/adamic-tools/env.sh.
Setup/tool versions remain those recorded in front33; setup is not rerun here.

Source tree is the previous unit's current-main 89ac4a8c fresh full-lane tree,
with 76 already applied and 65 reapplied idempotently/type-only. Both adapters'
checks pass, runtime JavaScript is unchanged, and sameMap is unchanged. This
unit makes no source adaptation or public API change; full upstream oracle is
not rerun. Both prior full lanes passed, and 32 local lane tests pass here.
The shared slicer gathers the seven driver roots afresh; verify.cjs passes on
2082 original spans and 79 module import prefixes. Code: 27 declaration files,
1993 declarations and 41857 copied-span lines.

Full and slice parser/run.sh outputs remain identical to the extended reference:
36429231 bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end, dropped JSDoc tags and JSDoc diagnostic mutants are caught on both.

measure-builds.py tests split off, on and repeat with jobs=5, timeout=60. All stop
before C: debug.ts:113:19 and 114:19 TS2339, captureStackTrace missing on
ErrorConstructor. Attempt wall times are not clang times. No native parser
acceptance, binary, performance/corpus measurements or warm cache claim.

A discovery copy replaces Debug.fail's body with a throw, solely to recheck the
core enum stop. The compiler then refuses core.ts:11:52, exactly as before.
The raw placeholder and exact discovery patch are saved. That copy is never
used for Node or acceptance and is not the archived validated slice.

Twenty-one focused witnesses execute on Node, then build natively. Eight compile
and match stdout, stderr and exit; all eight byte mutants are caught. Three
namespace declaration refusals clear (this/returned method, namespace class,
function/namespace merge). Those fixtures only print loaded messages, so their
green means declaration/lowering admission, not executed receiver or constructor
semantics. Non-null is newly refused. Enum initialization, mutable namespace
export and namespace initialization remain refused. Predicate overloads and
callback result remain green; predicate callback parameter stays refused.
arguments.length and its function-value forwarding both match (forwarding 1).
No accepted build produces a mismatch. comparison.json records every transition.

Commands and reproduction: git fetch named refs, worktree add, merge library
then abort; GOWORK=off GOTMPDIR=... go build -p 1 -mod=mod ./cmd/adamic; 65/adapt.cjs
and 76/check.cjs; slice/run.sh and verify.cjs; parser/run.sh on full and slice
with /tmp/parser-adapted10; measure-builds.py; parser-front34-minimals.py; the
single error-placeholder.cjs invocation; unittest discover under stage3/lane.
Reports retain exact paths, commands, outputs and timings. All tests go to logs.
Only codex/stage3-parser-proof is pushed, once at completion.
