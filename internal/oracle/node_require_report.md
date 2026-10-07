Built both perf_hooks require spellings on library host 94df56b, including tsc's guarded Partial destructuring.
Commit: this feature branch's performance host merge (see git log); no main or area branch was pushed.
Checks: adapted performanceCore.ts has zero diagnostics; loader/lower, focused Node oracle, counts and vet results below.
Mutants: nonliteral require acceptance, omitted projection support, and absent runtime performance projection all failed their intended tests.
Limits: incoming host fixtures block full flow/freshness; newer library tip remains pending.

Both literal spellings reuse nodeProcessValue and its existing ProcessCall runtime. Plain const
namespace bindings have the same IR as namespace imports. Immediate performance destructuring
can use Partial<typeof import(...)> because the module view does not escape. The selected property
must keep the module's declared type, optionally undefined. Escaping and changed-member views
remain refused or NotYet. No checker options or upstream source were changed.

The Node source runner supplies Node's Error, process and performance globals to its CommonJS
context. Otherwise errors thrown by the required Node module belong to a different realm than
that context's Error, incorrectly making the source's instanceof Error false. The initial oracle
caught this harness mismatch; both backends agree with Node after correcting the source context.

Stage 3 verification used a detached worktree at 7ad8666 and:

```sh
bash /workspace/require-stage3/stage3/apply.sh /workspace/require-adapted > /tmp/require-perf-apply.log 2>&1
go run ./cmd/adamic types /workspace/require-adapted/src/compiler/performanceCore.ts > /tmp/require-perf-performanceCore-final.log 2>&1
```

apply.sh exited 0. The stage-0 types command exits 1 because other transitive compiler files
still have diagnostics. Filtering the complete diagnostic log by /performanceCore.ts: finds
**0 diagnostics**, including 0 TS2591 and 0 TS2375. The tree still has its literal require
and adaptation 20's truthful optional declarations. The pinned Node package was made available
through a scratch /workspace/stage3/api/node_modules symlink to this branch's installed package.

Setup: Go 0s, clang 0s, Node 0s, submodules 0s, cache 152s, total 152s. nproc: 5.
The environment file is /workspace/adamic-tools/env.sh.

Validation logs (every test run was redirected, never piped):

- /tmp/require-perf-packages.log: load passed 8.219s, lower passed 89.369s;
  flow failed 156.828s and fresh failed 86.567s on incoming node_process fixtures.
  node_process_cwd_error.a and node_process_errors.a use refused in syntax.
  node_process_environment_mutation.a gets TS2542 from the readonly base process declaration;
  node_process_directory_mutation.a gets TS2339 for argv/cwd/chdir without a node import.
  These host probes need the newer library landing before a full green gate is possible.
- /tmp/require-perf-restored.log: all require, CommonJS refusal and local-name lowering tests pass 3.652s.
- /tmp/require-perf-counts-final.log: TestRequire and TestCountsAreRecorded with -update-counts pass 20.165s.
  Both require performance fixtures run their source on Node, native under ASan/UBSan/leaks,
  and the JS backend, observing hooks true false, monotonic clock checks and the missing-mark error.
- /tmp/require-perf-vet.log: go vet ./... exits 0, no diagnostics.

The new fixture rows each record 14 allocations, 14 frees, 39 retains, 47 releases, peak 6,
regions 0, identical to the library's namespace-import performanceCore fixture. The entire ledger
was regenerated. Existing node_fs_directory_system counts also moved to 2208/2208/3753/3289/1179/0;
that fixture scans the repository. This is an observation, not a claim that require lowering causes
those changes by itself. Other added rows come from the incoming host branch.

Mutants were each applied to real lowering code and restored:

- /tmp/require-perf-mutant-nonliteral.log: accepting a nonliteral as fs fails TestCommonJSRefusals,
  including the binding that incorrectly lowers successfully under the mutant (exit 1).
- /tmp/require-perf-mutant-projection.log: omit the immediate projection allowance;
  TestRequirePerformanceAndImportHaveTheSameIR fails on the real guarded Partial binding (exit 1).
- /tmp/require-perf-mutant-runtime.log: supply undefined for the performance module projection;
  TestRequirePerformanceAgreesWithNode fails for both spellings and both backends (exit 1).
  The mutated programs compile and run, but choose the global fallback and print hooks true true
  instead of Node's hooks true false. No clang warning or unrelated refusal catches this mutant.

The full gate was not rerun after its flow/freshness package blockers were established.
The newer library tip will be merged when supplied, followed by renewed validation.

Previous fs/path report follows for its historical evidence.

Built: literal fs and path require use the existing library hosts and the pinned Node declarations on codex/require-builtins-2.
Commits: d46ab6a dependency base; ff093cd cherry-pick of c880985; 36abd5f merge of the existing directory/path host.
Checks: focused loader/lower, required Node oracles, complete counts regeneration and vet pass; final load/lower/flow/fresh, focused Node oracle, counts and vet pass; the full gate passed every compiler package before its unrelated long-running checks were stopped.
Mutants: nonliteral acceptance fails refusal tests; structural fs type fails checker identity; local-name activation fails the Node-global isolation test.
Uncovered: perf_hooks lowering awaits the library host SHA; the actual submodule performanceCore has one separate exact-optional-property diagnostic.

## Current implementation

The branch was created from the rebased fs-file landing commit, without merging
that dependency into the old require branch. The old branch was left untouched.
The library directory/path branch at 3a090c7 was merged into the new feature
branch to reuse its path runtime rather than add another implementation.
Conflicts retained both fs-file and directory IR handling and runtime field
layouts. No main or area/ branch was pushed to or merged into.

Unbound require/module identifiers activate the same pinned @types/node 25.3.3 loader as
node:* imports. The declaration refinement augments Node's actual NodeJS.Require
interface; it declares no fs/path/performance host members. Individual literal
overloads preserve the exact imported module type and member declaration
identity. A union overload alone loses to Node's broad any signature; the
checker identity test detected that during integration.

Both require('path') and require('node:path') now lower to the existing join
and dirname host calls. Their IR matches namespace imports exactly. Both
fixtures print a/b then a from their original source on Node, and both emitted
backends agree, including ASan/UBSan and leak checks. Each new require path
fixture counts 2 allocations, 2 frees, 0 retains, 2 releases, peak 1, no regions.
The existing fs require row and all fs-file rows are unchanged.

The shared member guard now registers the implemented directory and path
members and names realpathSync.native correctly. The directory value handler
leaves sibling Stats fields to the file host. Complete counts regeneration
observes changes in four inherited directory rows after integration:
entries 225/225/140/238/29/0; realpath 118/118/58/148/7/0;
system 2032/2032/3400/3025/1047/0; permissions 82/82/70/105/32/0.
These are observed allocations/frees/retains/releases/peak/regions, not a claim
that require routing alone caused the changes.

## Perf-hooks typing and pending runtime

Both perf_hooks literal spellings already preserve the pinned imported module
type. TestRequirePerformanceCoreShapeChecks checks tsc's Node-like guard,
try/catch, destructured Partial module view, and all six requested members:
now, timeOrigin, mark, measure, clearMarks and clearMeasures. It passes through
the checker. Runtime lowering remains NotYet pending the library's host SHA.

The actual cohere submodule source is read in place, never copied:
cohere/TypeScript/tsc/testdata/fixtures/compiler/performanceCore.ts.
Running adamic types over it and its actual dependencies yields no missing
require diagnostic in that file. Its remaining diagnostic is TS2375 at line 66:
PerformanceHooks has optional fields without explicit undefined in their types,
but initializes them to undefined under exactOptionalPropertyTypes. The file
is not checker-clean. The stage-3 adapted source path has been requested; no
compiler option was weakened and no source suppression was added.

## Reproduction and evidence

Install the loader's exact dependency (no Node declarations were vendored):
npm install --prefix stage3/api --no-audit --no-fund --save-exact @types/node@25.3.3.
The local npm metadata and node_modules are workspace artifacts, not unit code.

Setup initially failed because it ran during unresolved cherry-pick/merge
conflicts. After resolution, bash cloud/setup.sh exits 0: Go ready 0s, clang
ready 0s, Node ready 0s, submodules ready 1s, build cache warm 181s, total 181s.
nproc is 5; cgroup cpu.max is 400000 100000. Every Go shell sources
/workspace/adamic-tools/env.sh.

All test output is saved under /tmp/require-builtins-2-*.log. Commands and
results so far:
- Focused load/lower require, CommonJS and Node library checks pass.
- go test -count=1 ./internal/load -run 'TestRequire|TestNodeBuiltin': exit 0,
  3.184s, including both perf-hooks checker shapes.
- go test -count=1 -timeout=30m ./internal/oracle -run
  'TestRequire|TestNodeFSFileAgreesWithNode|TestCountsAreRecorded'
  -args -update-counts: exit 0, 18.899s, after host integration corrections.
- go test -count=1 ./internal/oracle -run
  'TestRequire|TestInputAgreesWithNode/internal/oracle/testdata/node_path'
  -timeout=30m: exit 0, 23.961s.
- The first complete load/lower/flow/fresh package run passed load and lower
  and fresh (143.906s), but flow caught the directory dispatch/registration
  conflicts. Those conflicts are fixed; final validation is the full gate.
- Both mutants were rerun on this branch, each exits 1 through its intended
  test assertions. Both source files were restored before final validation.
- go vet ./... exits 0; gofmt and git diff --check have no output.
- go test -count=1 -timeout=30m ./... at e3f1efd passed load (7.767s),
  lower (72.253s), flow (146.560s), fresh (76.536s), native (411.372s),
  complete oracle (436.670s) and bridge (422.187s). It was stopped after more
  than twelve minutes while Unicode and stage-1 checks remained. Exit 1 from
  interruption; this is a partial gate, not a complete green repository gate.
- After the final local-binding detection correction, go test -count=1
  -timeout=30m ./internal/load ./internal/lower ./internal/flow ./internal/fresh:
  exit 0; load 5.816s, lower 45.568s, flow 83.150s, fresh 52.396s.
- Final focused Node oracle and count verification (same filter as above,
  without -update-counts): exit 0, 10.045s. Final vet and formatting pass.
- The actual sys.ts fs/path bindings use explicit typeof-import annotations.
  The added lowering checks preserve those annotations and compare the path
  IR to namespace imports: exit 0, 1.564s.
- Local require/module symbols now keep the ordinary prelude and do not load
  Node globals. The scope mutant ignores binding identity; its regression
  test exits 1 with 'local require or module activated Node globals'. The
  mutant was restored before all final package and oracle checks.

## Historical first attempt

The report below records the old branch and its old dependency. It is retained
as history and does not describe the current declaration or path implementation.

# Literal builtin require

Partial implementation on `codex/require-builtins`, started from main
`ef3d907ecdc4c771b016f7d9c52372def057a340`.

Current main had no Node builtin declarations or lowering. Inspected the pushed
`codex/host-fs-file`, `codex/host-fs-directory`, `codex/host-process`, and
`codex/host-buffer-crypto` branches. Fast-forwarded the existing fs file host
commit `080789ff46f07af44fafc8fb64c2b0ed1e24e129` as the dependency.

The directory tip `f5702191f4aff7028464df3ab2b334389b7ca496` includes the path
runtime, but its corrected report explicitly says the shared @types/node
25.3.3 loader hook is pending. Its lowering recognizes declarations from
node/fs.d.ts and node/path.d.ts, which the current fs-file tip does not provide.
Refreshing both branches confirmed these tips. The directory branch was not
merged because its corrected fixtures are intentionally blocked on that shared
hook. No second path runtime or declaration loader was implemented here.

The fs host serves selected @types/node 25.3.3 declarations in the embedded
prelude and recognizes calls by their declaration symbols. Literal `fs` and
`node:fs` require bindings use `typeof import('node:fs')`, so the checker gives
exactly the namespace import type and the same member symbols. Plain const
namespace bindings need no runtime local, just like namespace imports. Calls
reuse the existing fs host intrinsics. Source is never rewritten, and a local
function called require or a local module object remains ordinary user code.

The NodeRequire any result is refined to a module type for supported literals.
The fallback is unknown and is never admitted to lowering. The full @types/node
package is not loaded; the existing host's selected declarations remain the
source of its declared module type.

Non-literal calls (including a const whose inferred type is the literal fs),
non-builtin packages, relative paths, invented node: names, require.resolve,
require.cache, require read as a value, and module.exports are refused with an
import fix. Indexed module access is refused too. A checker error can precede
this refusal when a rejected require result is used with an invalid type; this
unit does not reorder loader diagnostics.

Node 24.19.0 observes `true true` and `true` from bare, prefixed and function-local
fs.existsSync calls. Native and JavaScript match stdout, stderr and exit status;
ASan, UBSan and the leak check pass. The fs require fixture adds only its count
row: 2 allocations, 2 frees, 0 retains, 2 releases, peak 1, no regions. Existing
rows did not move. An exact IR comparison also passes against namespace imports.

Pending the directory host's shared declaration hook, both path fixtures are
run from source on Node: join and dirname print `a/b`
and `a`. Lowering reports the missing node:path host as NotYet. These are gap
fixtures, not claims that path is implemented.

Mutants actually run and restored:

- Accept a non-literal argument as node:fs. TestCommonJSRefusals fails, including
  a const namespace binding which the mutant really accepts (`got <nil>`).
  The test exits 1; the compiler builds, so no compiler error or clang warning
  catches this mutant.
- Replace the fs module result with a structurally similar existsSync function
  object. TestRequireBindingHasTheImportedModuleType fails for both bindings:
  the checker type differs and the member declaration identity is lost. Exit 1.

Setup: Go ready at 0s; clang ready at 1s; Node ready at 1s; submodules ready at
1s; build cache warm at 95s; done at 95s. nproc is 5, with a four-core cgroup
quota. Tools were sourced from /workspace/adamic-tools/env.sh. Setup succeeded.

Workspace logs are /tmp/require-builtins-{setup,packages,focused,oracle,counts,
mutant,type-mutant,final-tests,format,vet,gate}.log. Logs are not committed,
as .gitignore requires. Commands actually run:

- bash cloud/setup.sh: exit 0, timing lines above.
- source /workspace/adamic-tools/env.sh: used in every test shell.
- go test -count=1 ./internal/load ./internal/lower: both packages print ok.
- go test -count=1 -v ./internal/oracle -run TestRequire: prints PASS; Node
  source observations and the named path gaps are logged.
- go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded
  -args -update-counts: prints ok, 19.427s, only the new row changes.
- go test -count=1 -v ./internal/lower -run TestCommonJSRefusals, with the
  non-literal mutant: exit 1, want CommonJS refusal but got <nil>.
- go test -count=1 -v ./internal/load -run
  TestRequireBindingHasTheImportedModuleType, with the structural type mutant:
  exit 1, module types differ and member declaration identity is lost.
- gofmt -l cmd internal: no output.
- go vet ./...: no output, exit 0.

Final checks after the prefix-only builtin correction and gap-fixture relocation:

- go test -count=1 -timeout 30m ./internal/load ./internal/lower: exit 0;
  load 1.752s, lower 33.898s.
- go test -count=1 -timeout 30m ./internal/flow: exit 0, 125.520s. The path gap
  fixtures live under testdata/require_gaps so the compilable-program glob does
  not treat them as supported programs.
- go test -count=1 -v -timeout 30m ./internal/oracle -run
  'TestRequire|TestNodeFSFileAgreesWithNode|TestCountsAreRecorded': exit 0,
  52.678s. Both path source observations, fs and its dependency's twelve host
  fixtures, sanitizers, leaks and counts pass.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...: exit 1. The initial
  run caught the misplaced path gaps; that is fixed and the full flow package
  subsequently passed. It also caught an inherited fs-host integration gap:
  internal/fresh reports ir.NodeFSFile as unknown, including the dependency's
  pre-existing import fixtures. The fs-file dependency adds the IR node but
  does not update internal/fresh. That directory is outside this unit's
  territory, so the check was not weakened or bypassed.

The full gate was interrupted after more than twelve minutes while unrelated
stage-1 port checks and a four-billion-comparison Unicode probe were still
running. It is not a complete green gate. Before interruption, load, lower,
native and the complete oracle package passed; the external bridge passed in
546.097s. The gate log retains the failures and interrupt. Final format, vet
and git diff --check have no output.

Remaining work: path and other host modules; mutable, destructured, returned or
inline namespace require values; runtime namespace identity comparisons; and
integration against tsc's complete sys.ts with full @types/node. These shapes
are rejected or report NotYet instead of silently receiving a fabricated
runtime namespace. All currently supported Node hosts are ambient declarations,
so they have no source module body to schedule; general graph evaluation for a
future builtin implemented in Adamic is not established by these fixtures.
