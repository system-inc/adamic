Built area/compiler and compiler/stage3-front-3 against the same adapted tsc entry tree.
Compiler SHAs: b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861 and a36d1c0472649ef2ee549cc2505b93a84e893b59.
Both native split modes exit 1 at builder.ts:1246:69, with unchanged TS2345.
All fifteen saved stops remain per compiler; zero disappear, including two placeholder artifacts.
Thirty Node witness runs, two filtered oracles and nine evidence mutants pass; no native tsc binary.

| Requested compiler | Compiler SHA | Stops disappear | Stops remain | Changed messages | Native split 0 / 1 exits | First stop |
| --- | --- | --- | --- | --- | --- | --- |
| origin/area/compiler | `b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861` | 0 | 15 | 0 | 1 / 1 | `src/compiler/builder.ts:1246:69`, TS2345 |
| origin/compiler/stage3-front-3 | `a36d1c0472649ef2ee549cc2505b93a84e893b59` | 0 | 15 | 0 | 1 / 1 | `src/compiler/builder.ts:1246:69`, TS2345 |

The unchanged first diagnostic for both compilers is:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Its minimal program is [01-indexed-path.a](probes/01-indexed-path.a):

```typescript
type Path = string & { readonly pathBrand: unknown };
function accept(path: string): void { console.log(String(path)); }
const paths: readonly Path[] = [];
accept(paths[0]);
```

Node 24.19.0 prints `undefined` and exits 0. Both compiler tips report the same
TS2345 and exit 1 on this witness. There is no new first location: neither tip
passes the previous entry's first checker stop.

The adapted tree was reused at `/tmp/tsc-entry-refresh/adapted`, not regenerated
from either compiler branch. All 81 reached source hashes match the previous
main-efe9f404 measurement before and after each run. The baseline evidence was
pushed in `3caf28a8`; [REFRESH.md](REFRESH.md) records its construction and the
full fifteen diagnostics. Detached compiler worktrees have their own pinned
submodules: cohere `7945d102a6c18dd36adf9114a758ce646e8b2359` and TypeScript
`d92d9bfee114c80be2c375d72edae966176e3a4f`. Binary hashes, embedded compiler
revisions and `vcs.modified=false` are retained with the evidence.

Every reached file outside `src/compiler`, separately:

- `src/tsc/tsc.ts`
- `src/tsc/_namespaces/ts.ts`

Neither outside file is a stopping site. `executeCommandLine.ts` is inside
`src/compiler` in this source tree.

For each old stop, the driver replayed the exact preceding body replacements
in a temporary source copy, preserving UTF-16 offsets and original CRLF line
endings. It ran `adamic types` on that saved snapshot and compared the exact
message at the recorded scratch location. This checks later stops even though
an earlier checker diagnostic still blocks native compilation. Pristine entry
build diagnostics were also compared at the original adapted-source locations.

Both tips still report stops 1 to 8, 10 to 13, and 15 in the pristine tree.
Stops 9 and 14 remain only in their respective placeholder snapshots; they are
artifacts of changed inference and must not be counted as unmodified-tree
failures. Each minimal program was independently checked and built in split 0
with each compiler, and run through the scanner-style stock TypeScript loader
on Node. All fifteen minimal checker failures remain. Thirteen messages match
exactly; witnesses 3 and 12 retain the smaller structural object types documented
in the preceding unit, reproducing the same TS2375/TS2379 failure mechanism.

The following coordinates are in the original adapted tree. Each "Remain"
means an exact message match in the corresponding saved snapshot.

| Old stop | File within src/compiler | Original line:column | Diagnostic | b68b2fe1 | a36d1c04 | Minimal program |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | builder.ts | 1246:69 | TS2345 | Remain | Remain | [01-indexed-path.a](probes/01-indexed-path.a) |
| 2 | builder.ts | 1258:65 | TS2488 | Remain | Remain | [02-tuple-parameter.a](probes/02-tuple-parameter.a) |
| 3 | builder.ts | 2273:9 | TS2375 | Remain | Remain | [03-present-undefined-field.a](probes/03-present-undefined-field.a) |
| 4 | builder.ts | 395:118 | TS2345 | Remain | Remain | [04-optional-path-argument.a](probes/04-optional-path-argument.a) |
| 5 | builder.ts | 991:17 | TS2345 | Remain | Remain | [05-optional-path-argument.a](probes/05-optional-path-argument.a) |
| 6 | checker.ts | 10022:63 | TS2345 | Remain | Remain | [06-optional-symbol-argument.a](probes/06-optional-symbol-argument.a) |
| 7 | checker.ts | 10026:97 | TS2345 | Remain | Remain | [07-optional-symbol-argument.a](probes/07-optional-symbol-argument.a) |
| 8 | checker.ts | 10033:67 | TS18048 | Remain | Remain | [08-optional-member-read.a](probes/08-optional-member-read.a) |
| 9 | checker.ts | 10034:53 | TS2345 (placeholder artifact) | Remain | Remain | [09-never-argument.a](probes/09-never-argument.a) |
| 10 | checker.ts | 10319:83 | TS2345 | Remain | Remain | [10-optional-symbol-argument.a](probes/10-optional-symbol-argument.a) |
| 11 | checker.ts | 10320:51 | TS2345 | Remain | Remain | [11-optional-symbol-array.a](probes/11-optional-symbol-array.a) |
| 12 | checker.ts | 10941:33 | TS2379 | Remain | Remain | [12-present-undefined-argument.a](probes/12-present-undefined-argument.a) |
| 13 | checker.ts | 12034:30 | TS18048 | Remain | Remain | [13-optional-file-read.a](probes/13-optional-file-read.a) |
| 14 | checker.ts | 12859:13 | TS2322 (placeholder artifact) | Remain | Remain | [14-void-result.a](probes/14-void-result.a) |
| 15 | checker.ts | 13986:47 | TS2345 | Remain | Remain | [15-optional-declaration-argument.a](probes/15-optional-declaration-argument.a) |

Full messages, pristine presence, snapshot diagnostics, probe outputs and exit
codes are in the per-compiler JSON tables:

- [area/compiler comparison](evidence/compiler-tips/area-b68b2fe1/comparison.json)
- [stage3-front-3 comparison](evidence/compiler-tips/front3-a36d1c04/comparison.json)

Observed compiler behavior is consistent with the loading boundary. At both
pins, `internal/load/load.go:131` returns `CheckError` when the program has
checker diagnostics; semantic diagnostics are collected at line 217. That code
is identical to the measured main. The inserted-check implementation does not
make these particular strictness diagnostics admissible at either measured tip.
Native lowering was never reached, so this experiment does not assess the
runtime correctness of checked non-null lowering.

Reproduction, from the existing branch, with the environment sourced in each
build shell and the previous adapted tree present:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/tsc-entry-tips-setup.log 2>&1
source /workspace/adamic-tools/env.sh
git fetch origin area/compiler:refs/remotes/origin/area/compiler compiler/stage3-front-3:refs/remotes/origin/compiler/stage3-front-3 > /tmp/tsc-entry-tips-fetch.log 2>&1
git worktree add --detach /tmp/tsc-entry-compiler-area b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861
git worktree add --detach /tmp/tsc-entry-compiler-front3 a36d1c0472649ef2ee549cc2505b93a84e893b59
git -C /tmp/tsc-entry-compiler-area submodule update --init --recursive > /tmp/tsc-entry-tips-area-submodules.log 2>&1
git -C /tmp/tsc-entry-compiler-front3 submodule update --init --recursive > /tmp/tsc-entry-tips-front3-submodules.log 2>&1
(cd /tmp/tsc-entry-compiler-area && go build -o /tmp/tsc-entry-area-adamic ./cmd/adamic) > /tmp/tsc-entry-tips-area-build.log 2>&1
(cd /tmp/tsc-entry-compiler-front3 && go build -o /tmp/tsc-entry-front3-adamic ./cmd/adamic) > /tmp/tsc-entry-tips-front3-build.log 2>&1
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
python3 stage3/drivers/tsc-entry/compare-compilers.py /tmp/tsc-entry-compiler-area /tmp/tsc-entry-area-adamic /tmp/tsc-entry-refresh/adapted stage3/drivers/tsc-entry/evidence/compiler-tips/area-b68b2fe1 > /tmp/tsc-entry-tips-area-run.log 2>&1
python3 stage3/drivers/tsc-entry/compare-compilers.py /tmp/tsc-entry-compiler-front3 /tmp/tsc-entry-front3-adamic /tmp/tsc-entry-refresh/adapted stage3/drivers/tsc-entry/evidence/compiler-tips/front3-a36d1c04 > /tmp/tsc-entry-tips-front3-run.log 2>&1
python3 stage3/drivers/tsc-entry/verify-compilers.py > /tmp/tsc-entry-tips-verification.log 2>&1
python3 stage3/drivers/tsc-entry/compiler-mutants.py > /tmp/tsc-entry-tips-mutants.log 2>&1
```

The output directories must be new. The comparison runner invokes
`ADAMIC_NATIVE_SPLIT=0` and `1`, with `ADAMIC_NATIVE_JOBS=5`, on
`src/tsc/tsc.ts`. Both stdout and full stderr agree byte for byte per compiler.
It refuses a tree with source hashes differing from the saved baseline or a
binary whose embedded revision differs from the clean compiler checkout.

Setup passed: Go ready 0.037s, Node ready 0.037s, markdown ready 0.101s,
submodules ready 0.183s, clang ready 0.450s, Go build ready 49.470s, done
49.735s. `nproc=5`, cgroup quota 4 CPUs; Node 24.19.0, Go 1.27.1, clang 20.1.8.
The environment file is `/workspace/adamic-tools/env.sh`.

Verification reports `Two compiler pins verified; 30 saved stops classified;
all Node witnesses agree`. Nine independent evidence mutants were run:

| Mutant | Catcher | Exit |
| --- | --- | --- |
| Wrong compiler SHA | Expected compiler pin | 1 |
| Dirty binary provenance | Embedded clean binary revision check | 1 |
| Changed adapted-source hash | Baseline source hash comparison | 1 |
| Missing stop 15 | Required stop population | 1 |
| Changed split 1 stderr | Full split stream comparison | 1 |
| Changed Node stdout | Expected Node observation | 1 |
| Change stop 1 from remain to disappear | Recomputed exact diagnostic classification | 1 |
| Invent a different first stop | Comparison with retained native stderr | 1 |
| Change the minimal checker diagnostic | Exact baseline probe diagnostic comparison | 1 |

Each compiler also passed this existing filtered oracle from its own worktree:

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/undefined_references.a$' -count=1 -timeout 30m > ORACLE_LOG 2>&1
```

Area/compiler passed in 27.423s; stage3-front-3 passed in 11.678s. Driver Python
syntax checks and `git diff --check` pass. Complete outputs and mutant catchers
are retained under `evidence/compiler-tips/`.

No compiler, library or adaptation sources were edited. Native tsc compilation
never passes checking, so there is no binary for the conditional `--tiny`
harness. The full repository gate was not run; this unit's Node fixtures,
comparison checks, mutants and the two filtered oracles are the local validation.
