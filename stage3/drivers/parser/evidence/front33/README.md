# Land-area-next native rerun

Compiler scratch is exactly origin/cloud/land-area-next
28285421bf59ea46269545144c19123156c00c91. The requested records-maplike-library
2648ad3362812b0f184a8cb922bc322d18729c66 merge conflicted in 44 paths, including
lowering, loading and native runtime; paths and complete merge output are saved.
The merge was aborted, not resolved, as instructed. No compiler merge is pushed.
The scratch go.mod redirects the pinned cohere 7945d102 SDK to its existing
checkout; no compiler sources change. Go build -p 1 -mod=mod succeeds.

The adapted source pipeline is the completed proof branch at 4dca4e4c, with
65 and permanent 76; sameMap is unchanged. Fresh apply exits 0; 65 is idempotent.
76's idempotence, writer, enumerator and old sameMap signature mutants pass.
The prior complete main-versus-adapted lanes remain the baseline evidence; this
unit does not rerun the full 106366-test suite because no adaptation changed.
32 local lane tests pass. No public API sanctions are added.

Both native split modes stop in the checker at debug.ts:113:19 and 114:19:
TS2339, captureStackTrace is missing on ErrorConstructor. This is a regression
relative to the prior library scratch, not a native green. No C, clang timing,
parser binary, native output or native acceptance corpus exists. Attempt times
0.5183s unsplit and 0.4692s split are pre-clang wall times. Repeated split also
fails before clang and is not a warm-cache measurement.

The ordered source list has ten rows. Row 1 is the real build failure; all later
rows are behind throwing discovery placeholders. Rows 2-8 retain the seven enum
initialization sites. Rows 9-10 retain early namespace calls. Those last calls
are observed at helper placeholders, and the unchanged Map-before-namespace
minimal independently reproduces the same namespace preflight. Original column
29 for debug.ts:333 is distinguished from observed helper column 109. Both
failed discovery attempts and exact patches are saved separately.

Discovery stopped before the fifteen-row limit: removing call-shaped throwing
initializers requires a typed uninitialized binding plus a guaranteed throwing
arithmetic guard. That creates TS2454 on enumMemberCache; stubbing its consumer
then creates an inferred void-result error. Neither is a source/compiler stop,
so neither enters the ordered list. The first attempt also duplicated an
initializer helper; those duplicate-implementation errors are artifacts too.
No edited discovery tree is used for Node or native acceptance.

Twenty focused probes: six build and match Node stdout, stderr and exit; each
one-byte native-output mutant is caught. Green: non-null read, predicate
overload, predicate callback result, some overload declarations, arguments.length,
and arguments.length through a function value (prints 1, not the old wrong 9).
The generic predicate callback parameter remains refused. All three namespace
forms remain refused, as do enum initialization, MapLike and the two casts.
The Map-before-namespace minimal is an additional refusal witness. All exact
Node, compiler and comparison outputs are saved. No silent miscompile observed.

Final source audit: 27 declaration files, 1993 code declarations, 41857 copied
span lines; 2082 spans and 79 ordered module import lists pass byte audit.
Full and slice Node output are identical, 36429231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end, dropped-tags and JSDoc-diagnostic mutants are caught on both trees.
The prior 10406-case manifest is unchanged and not rerun in this unit.

Commands: git fetch of the named refs, worktree add, merge 2648ad33 then abort;
GOWORK=off GOTMPDIR=... go build -p 1 -mod=mod ./cmd/adamic; stage3/apply.sh;
adapters 65 and 76/check.cjs; stage3/slice/run.sh with seven driver roots and
verify.cjs; measure-builds.py with jobs=5 and timeout=60; the retained minimal
and discovery scripts; parser/run.sh on full and slice with /tmp/parser-adapted10;
python3 -m unittest discover -s stage3/lane -p test*.py. Exact commands appear
in scripts and JSON reports. Only the proof branch is pushed.

Setup exits 1 during go list -deps -export, with empty list.log and packages.json;
no further failure text or completed warming time is emitted. The available
Go/Node/clang and the separately built compiler work. Timing lines: Go 0.017s,
Node 0.021s, markdown 0.077s, clang 0.170s, submodules 0.504s; nproc=5.
The setup script restored the pinned root submodules. Its ASan overflow report
is the setup's expected sanitizer capability probe, not the parser failure.

Run `python3 stage3/drivers/parser/evidence/front33/check.py` to verify the
recorded outcomes. Its planted artifact-count mutant is rejected; comparisons
and six native byte mutants pass. No binary or generated C is committed.
