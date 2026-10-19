from pathlib import Path
import subprocess,json,sys
root=Path('/workspace/adamic')
e=Path('/tmp/adamic-ambient-evidence')
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
base=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=root,text=True).strip()
refusals=json.loads((e/'sweep-new-refusals.json').read_text())
rows=[json.loads(s) for s in (e/'sweep.jsonl').read_text().splitlines()]
preexisting=[r for r in refusals if not r['path'].startswith('internal/oracle/testdata/ambient_refused/')]
report=f'''# Ambient declarations refusal, task #3bddhs8

Code: `{sha}` on `compiler/lower-ambient-declarations`.
Base: `{base}`, newest origin/main when the work was based and rechecked before commit.
Evidence is separate, on `evidence/compiler/lower-ambient-declarations`.
Code pushed green at {(e/"code-pushed-at.txt").read_text().strip()}.

Program modules must define runtime values rather than have the checker believe an ambient declaration. Lower now refuses ambient values, functions, classes, enums, value-bearing namespaces, augmentations, and parsed triple-slash references. Type-only declarations remain legal. The existing program-module walk keeps the prelude, bundled library, and generated platform bindings trusted.

## Files and lines

- `internal/lower/ambient.go:12`: parsed reference directives and source locations; `:40`: ambient flags; `:46`: nested namespace value detection; `:61`: declaration and augmentation refusal.
- `internal/lower/refusals.go:37`: reference refusal beside the directive checks; `:168`: declaration refusal immediately after namespaceRefusal.
- `internal/lower/enum_initialization.go:12`, `internal/lower/enums.go:393`, `internal/lower/namespaces.go:140`: let ambient declarations reach their refusal instead of an earlier NotYet.
- `internal/lower/ambient_test.go:15`: defined/type-only control; `:34`: identifier-module AST; `:46`: defined imported module control.
- `internal/lower/enums_test.go:126`, `internal/lower/namespaces_test.go:10`: update former NotYet expectations.
- `internal/oracle/ambient_refused_test.go:13`: directory test, source Node behavior, complete diagnostics, and a-check headers.
- `internal/oracle/testdata/ambient_refused/`: 13 programs, each with a pinned `.refused`, plus the reference-only `.d.ts` helper.

## Reproduction and controls

On the same `declare let limit: number` arithmetic program, base main's native binary prints `1` and exits 0. Source Node through the oracle reports `ReferenceError: limit is not defined` and exits 70. The new compiler exits 1 with the pinned ambient-value refusal. See `main-let-*`, `source-let.*`, and `new-let.*`.

All 13 refused fixtures throw ReferenceError on source Node. The lib-reference and module-augmentation witnesses include a subsequent ambient value read so Node also exercises the missing-value hole while the directive or augmentation remains the first compiler refusal.

The lower controls accept `interface`, `type`, a type-only ambient namespace, and a relative imported module whose function and values are defined. No accepted oracle fixture was added.

The mutant prepends `return nil` in both `ambientRefusal` and `referenceRefusal`. All 13 fixture subtests fail their refusal or whole-diagnostic assertion. `declare var` reaches the old var refusal, which the whole diagnostic pin correctly rejects. No build failure or clang warning killed the mutant. The runner restores `ambient.go` in `finally`; the final unmutated oracle gate ran afterward, and the required package gates also passed unmutated. See `mutant.py`, `mutant-summary.log`, and `mutant.jsonl.gz`.

## Validation

Toolchain: Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test command wrote its complete output to a file, without head/tail pipelines.

```
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal
  exit 0, no output (gofmt.log)
go vet ./internal/lower ./internal/oracle
  exit {(e/'vet.exit').read_text().strip()}, no output (vet.log)
go test -count=1 -timeout 30m ./internal/lower ./internal/load ./internal/fresh ./internal/flow
{(e/'packages.log').read_text().rstrip()}
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
{(e/'oracle.log').read_text().rstrip()}
```

Counts changed: none. The complete oracle checked recorded counts. Counts were not regenerated because the accepted controls are lower tests, and there is no new accepted oracle row.

## Sweep

Both compilers were built from the recorded revisions: base main and the code commit. `adamic c` ran on all {len(rows):,} git-tracked `.a` and `.ts` programs, excluding `cohere/`, every `node_modules/`, and `.d.ts`. This includes review inputs, but no file under review was edited. Per-command timeout: 120 seconds. Timeouts: {sum(r[k]['exit']=='timeout' for r in rows for k in ('main','new'))}. Full command results, exit codes and stderr: `sweep.jsonl.gz`. Manifest: `sweep-paths.txt`. Reproduction: `sweep.py main`, `sweep.py new`. The five strengthened witness inputs were refreshed on base main with `refresh-main-witnesses.py` before the branch sweep; their final observations replace the earlier rows.

{len(refusals)} programs acquire this refusal as their first stop: {len(preexisting)} pre-existing programs and 13 new fixtures. The complete list follows.

```
'''
report+='\n'.join(r['path'] for r in refusals)+'\n```\n\n'
report+='Old and new diagnostics for every changed program are in `sweep-new-refusals.json`.\n\n'
report+='''
The sweep adds no new acceptances. Nine review probes panic natively on both revisions with the same first diagnostic and exit 2; their differing stack addresses and checkout paths account for the other nine raw-stderr differences. They were left outside this refusal task. Their full paths are in `sweep-summary.json`, and both stack traces are retained in `sweep.jsonl.gz`.

The changed pre-existing consumers are probes, corpus programs, or fixtures, not Adamic's own tooling. In particular, `stage3/fixtures/host/22_createHash_fallback.a:23` is now refused for its ambient createSHA256Hash function. It was not edited. Runtime Library must re-measure host status.json after landing.

## Pinned fixture diagnostics

'''
for p in sorted((root/'internal/oracle/testdata/ambient_refused').glob('*.refused')):
 report+='```text\n'+p.read_text().rstrip()+'\n```\n\n'
report+='''## Scope choices and limitations

The pinned checker already rejects identifier `declare module Name` with TS1540, before Lower can run. A lower test supplies that AST keyword on a checked namespace to verify the same outermost ambient refusal; no loader or checker exception was added. The string-module augmentation fixture uses the existing 'adamic' module because unresolved literal 'x' is rejected first with TS2664. These independent checker stops are recorded in `identifier-module.stderr` and `unresolved-module.stderr`.

No accepted oracle fixture or counts regeneration was needed. The whole-repository ./... gate was not run; the requested packages, vet targets, formatting check, and full uncached oracle were run. No host fixture or host status.json was changed.
'''
for log in e.glob('*.log'):
 report=report.replace(log.name, log.name+'.gz')
(e/'REPORT.md').write_text(report)
print('report written',len(report),'characters')
