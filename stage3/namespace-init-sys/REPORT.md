4318e17b29301aea7519c2f2cc5d3a9b24a79f9d Excluded ambient host namespaces from executable initialization; real namespace reachability and readiness remain checked.
Commits: implementation f11380454dd1319682d420ce8beb59397f2f6df7, count rows f8fea8a789705eb00d52dc251cb7c59ca3ce4a8b, main merge 919e53f2, parameter-property repair be7c88fc and diagnostic contract 4318e17b; feature branch only.
Commands and outputs: full lower passes (305.967s); final scoped Node oracle passes (8.281s); vet and count gates pass; production advances to core.ts:307:5.
Mutants: reaching access misclassified, ambient exclusion erased and parameter-property recognition dropped; each is caught by its refusal/preflight pins.
Not covered: exact fs.native and process.cwd reductions reach independent host NotYet in both backends; neither exact host program runs natively.

## What changed

namespaceRuntime now returns false for inherited ambient context, an explicit
ambient modifier or a declaration file. Such declarations describe the host;
there is no program statement that initializes them. The SCC call graph still
records executable namespace reads, and unresolved reads retain the existing
runtime readiness check. No host member, loader, ordinary enum rule or production
probe refusal was bypassed.

## Reductions and independent observations

[realpath.a](realpath.a) cuts sys.ts:1492:142 to the native-property selection.
[cwd.a](cwd.a) cuts sys.ts:1502:51 to memoize(() => process.cwd()). Both print
`true` with exit 0 and empty stderr on independent Node. Both pass the isolated
namespace preflight. Full native and JavaScript lowering then return named
host NotYet: `node:fs.native` at realpath.a:3:22 and
`node:process.Process.cwd` at cwd.a:7:43. [Exact commands and bytes](reductions.json).
These are measured host limitations, not native success or language decisions.

The structural controls explicitly substitute represented object values for the
unimplemented host. realpath-control.a prints `pathpath:false`; cwd-control.a
prints `cwdcwd:cwdcwd:false`. The latter uses a callback forwarder: the exact
memoizer's captured function parameter triggers the existing cycle-capable type
rule even after substituting the host. The initial control failure is retained.
These controls prove branch/deferred-callback behavior in both backends, not
implementation of fs.realpathSync.native or process.cwd.

## Sound boundaries and mutant

TestNamespaceAmbientHostInitialization loads both exact programs against the
pinned Node declarations, requires initialization success and requires an
independent named host stop. TestNamespaceAmbientContextsDoNotExecute checks
explicit declare, inherited ambient context and declaration-file context while
requiring an ordinary executable namespace to remain executable.

The existing direct/helper/cycle fixtures in internal/lower/testdata/namespaces_notyet
retain their located premature-read NotYet. The reaching mutant deletes only
namespace read recording in namespaceCallGraph.discover. All three
TestNamespaceInitializationReachability subcases fail with
`reaching call lost its initialization refusal: <nil>`. Compilation failure is
not the catch. Five existing unknown-before fixtures still lower, run and stop
with Node's exact error/effects in both backends, exercising the runtime fallback.

## Setup and validation

Setup succeeded in 262.537s, nproc=5, cgroup CPU quota four. Timing lines:
Go ready 0.074s, Node 0.073s, submodules 0.136s, markdown 0.237s,
clang 0.539s, Go build 260.622s, warm cache 262.321s.
Source environment: /workspace/adamic-tools/env.sh.
An initial host test run failed because @types/node 25.3.3 was missing after
switching baselines; npm ci --prefix stage3/api installed the pinned package,
and the rerun passes. No compiler behavior was changed for that setup failure.

Test output is retained in logs. The scoped oracle includes the two new
controls and five namespace unknown-before cases. The full lower run uses no
skips, clearing the baseline's two fs.readFile expectation failures.

## Production rerun

The requested baseline has no stage3/stricter-indexed-all/README.md. Its REPORT.md
reproduction commands are authoritative. The input is recreated using the
3b255125 adaptation worktree and npm ci. build-production-probe.py verifies all
79 dated source hashes: zero mismatches. The probe uses the original owning
project settings and the unchanged exact exclusion manifest. Only five catch
errors are excluded from the probe; ordinary production loading remains unchanged.
The first implementation push was f11380454dd1319682d420ce8beb59397f2f6df7.
All 16 production attempts now pass namespace preflight and stop at:

`src/compiler/core.ts:307:5: Adamic 0.1 refuses a write to discriminant field 'length' that could move the object to variant 0 or number; changing variant means building a new object`

The exact statement is `array.length = outIndex` in filterMutate<T>.
This is an independent discriminant-write refusal. No indexed guard is reached.

| Original ledger group | Scheduled | Emitted | Checker errors retained in census | Read-local not reached |
| --- | ---: | ---: | ---: | ---: |
| Indexed reads | 99 | 0 | 0 | 99 |
| Optional writes | 67 | 0 | 0 | 67 |
| JSON stringify | 2 | 0 | 0 | 2 |
| Catch variables | 0 | 0 | 5 | 5 |
| Total | 168 | 0 | 5 | 173 |

[All 173 final original row outcomes](production-final/census-row-outcomes.csv) retain the
common blocker separately from unvisited read-local outcomes. The probe and
classifier exit zero because they collect evidence; no native artifact exists.
Both sys sites move from a false namespace blocker to successful preflight.
The 99 indexed rows change common blocker, with their measured states unchanged.

## Feature-only main merge

Fetched origin/main bfa0bfec9ab6f49c81e22984d3fca4ee21a472e8 and merged it into
this feature branch as 919e53f2. No main or area branch was modified or pushed.
The class conflict uses main's transitive method/getter initializer analysis
and retains the existing initialized-this fallback diagnostic rule tag. The host status
conflict is remeasured after exact independent Node agreement: 19 NotYet,
6 Checker, zero native host builds. The new a-check headers shift diagnostic
lines; old missing-Node-declaration checker observations are superseded by
measured named host stops. Node outputs are unchanged.

The earlier vet process reached the merge while conflict markers were present
and failed to parse class_inheritance.go. The subsequent resolved-tree vet is
reported separately; the parse failure is neither a mutant catch nor a green gate.


## Commands

All commands run with /workspace/adamic-tools/env.sh sourced. Output goes to logs.

```sh
go test ./internal/lower -run 'TestNamespaceAmbient|TestNamespaceInitializationReachability|TestNamespaceCallGraph|TestNodeLibraryNamesUnimplementedMembers' -count=1 -timeout 10m -v
go test ./internal/lower -count=1 -timeout 10m
go test ./internal/oracle -run 'TestNativeAgreesWithNode/stage3/namespace-init-sys|TestNativeAgreesWithNode/internal/oracle/testdata/namespaces_unknown' -count=1 -timeout 10m -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m
go vet ./internal/lower ./internal/oracle
python3 stage3/stricter-indexed-all/build-production-probe.py --tree /tmp/namespace-init-sys-typescript --scratch /tmp/namespace-init-sys-production
/tmp/namespace-init-sys-production/production-probe /tmp/namespace-init-sys-typescript stage3/stricter-indexed-all/evidence/ledger-options.json /tmp/namespace-init-sys-production/exclusions.json /tmp/namespace-init-sys-production/native
python3 stage3/stricter-indexed-all/classify-census.py --result /tmp/namespace-init-sys-production/result.json --tree /tmp/namespace-init-sys-typescript --output stage3/namespace-init-sys/production
```

The post-main probe repeats those production commands with scratch
/tmp/namespace-init-sys-production-merged and output production-merged.
The merged Node gate runs TestNativeAgreesWithNode, TestNamespace and
TestClassWrongOutput, count=1, timeout=20m. Whole lower and vet are rerun.
The count gate is updated and verified after that merge as well.

A second mutant removes only the ambient exclusion using a scratch Go overlay.
Both exact host preflight pins fail at their namespace accesses, and the ambient
context pin fails. The production source file remains unchanged for this mutant.

## Merge safety repair

The first complete merged lower run fails exactly three existing
TestParameterPropertiesSoundness pins: inherited initializer, early constructor
default and field initializer. Main's new initializerReads follows ordinary
property declarations but omitted parameter-property declarations. The repair
uses the existing parameterProperty predicate alongside PropertyDeclaration;
its availability map and transitive method/getter traversal remain intact.
The focused parameter-property, incoming class and ambient namespace tests pass
(19.662s). Dropping only that predicate via an isolated Go overlay returns
success for those three unsafe programs and fails all three refusal pins.
The final full lower gate is rerun after this repair.

The superseded production-merged run was terminated to give CPU to the final
repaired-tree production run. Its incomplete JSON is not counted as evidence.
The first implementation production run remains complete, with all 16 entries;
the final production run is independently rebuilt and classified.

## Final measured results

The final Node filter passes in 8.281s: 92 native cache hits, 85 Node hits.
It covers both new controls, namespace runtime and static semantics, existing
parameter-property programs, namespace mutants and all four incoming class
proofs. The whole lower package passes in 305.967s. The final vet exits zero.
The merged count update passes in 267.766s and verification in 273.789s.
Only two count rows were added; every previous row remains unchanged.

The attempted broad oracle was interrupted after 631.746s and is not a claimed
full gate. The first final scoped run failed three exact incoming class repair
text assertions: the merge had added a rule tag to main's new path diagnostic.
Commit 4318e17b restores that path diagnostic's exact text. The final scoped run
passes. This was a diagnostic contract mismatch, not a native/Node disagreement.
The full lower and production run contain the parameter-property safety repair;
4318e17b's subsequent compiler change affects only that class diagnostic text.
The final scoped gate and vet cover its exact final source.

[Final production summary](production-final/census-summary.json),
[every row](production-final/census-row-outcomes.csv) and
[raw result](production-final/result.json) independently reproduce the same
core.ts:307:5 blocker on all 16 entries after the safety repair. All 79 source
hashes still match. Scheduled/emitted/not-reached counts do not change after
main integration. The only original program rows changed are their common
blocker text: sys.ts ambient namespace initialization to core.ts discriminant
write. Neither production run emits a native artifact or any indexed guard.

[Reaching programs and exact Node/backend diagnostics](reaching.json) preserve
the direct, helper and cycle pins. Independent Node stops at the undefined
namespace read; both backend commands retain the located known-premature-read
NotYet. The compiler executable used for these pins is the initial namespace
fix; the namespace implementation is byte-identical after the main merge.

No full repository gate is claimed. Exact host programs are still NotYet;
container escape, callable namespace host properties and process cwd lowering
remain separate work. No main or area branch was pushed, and no PR was opened.
