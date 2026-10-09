# Step 31 first native stops

Built two reduced source witnesses, exact scout evidence, owner attribution and four scratch merge attempts.
Base main 031a1259bc7973934792dc6cb1bd4074fc2204b9; train 1e81b051c52eee40d47c44f4a5f68b821e606708; scout 350e9ca6ebb8bd16c66943a44073fa9018e48ec1.
Node and stock TypeScript 6.0.3 checks pass: two outputs, two diagnostic codes and two responsible-option ablations.
Two mutants caught: empty declarations throws TypeError; omitting originalPath changes own-key output.
Native confirmation is unfinished: required submodule authentication fails, main cannot build and all four candidate merges conflict.

## Exact stops

The copied scout observations cover both its main pin and train tip, rather than a new native measurement. Both exit 1 before producing binaries. Complete stderr, command, exit and report files are under evidence. Paths below retain the scout's temporary slice coordinates; source mapping restores the upstream coordinates.

Checker source checker.ts:12034:30, slice checker.ts:10894:30:

```
/tmp/step31-train-checker-slice/src/compiler/checker.ts:10894:30: error TS18048: 'file' is possibly 'undefined'.
```

The source gets `file` from `getSourceFileOfNode(symbol.declarations![0])` and reads `file.endFlowNode`. The assertion unwraps the declarations array, not element zero. An unchecked indexed read changes overload selection to the optional-return overload. The reduction retains that call/read and overload relation; its SourceFile stores only the observed flow number. This is deliberately a reduction, not the full getFlowTypeFromCommonJSExport body. The identity helper replaces the source-file parent walk.

Emitter source moduleNameResolver.ts:2772:21, slice moduleNameResolver.ts:1549:21:

```
/tmp/step31-train-emitter-slice/src/compiler/moduleNameResolver.ts:1549:21: error TS2322: Type 'SearchResult<{ path: string; extension: string; packageId: PackageId | undefined; originalPath: string | undefined; resolvedUsingTsExtension: boolean | undefined; }>' is not assignable to type 'SearchResult<Resolved>'.
  Type '{ readonly value: { path: string; extension: string; packageId: PackageId | undefined; originalPath: string | undefined; resolvedUsingTsExtension: boolean | undefined; } | undefined; }' is not assignable to type 'SearchResult<Resolved>'.
    Types of property 'value' are incompatible.
      Type '{ path: string; extension: string; packageId: PackageId | undefined; originalPath: string | undefined; resolvedUsingTsExtension: boolean | undefined; } | undefined' is not assignable to type 'Resolved | undefined'.
        Type '{ path: string; extension: string; packageId: PackageId | undefined; originalPath: string | undefined; resolvedUsingTsExtension: boolean | undefined; }' is not assignable to type 'Resolved' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'originalPath' are incompatible.
            Type 'string | undefined' is not assignable to type 'string | true'.
              Type 'undefined' is not assignable to type 'string | true'.
```

The reduction preserves the returned object and toSearchResult statements, simplifies PackageId to string and replaces surrounding resolution with a supplied result. Explicit `originalPath: undefined` creates an own field. Omitting that assignment would change JavaScript semantics. The driver therefore prints its own keys as well as path and undefined status.

## Ownership and admission evidence

The copied candidate states CSV has CRLF normalized to LF; its row content is unchanged.

All origin codex/* and compiler/* branch heads were fetched and enumerated in evidence/origin-branches.txt. Reports and loader scheduling code were inspected. Task IDs were searched in repository history and files; no task records were found. The IDs below are the dispatcher's ownership labels, not independently retrieved task content.

| Stop | Task and candidate | Evidence and limit |
| --- | --- | --- |
| Checker D037 | Stricter options #k881crd; codex/stricter-indexed-all aee98c837f787319262cf643021cb8b7c97efae1 | whole-program/states.csv schedules D037 as indexed-presence. Slice a explicitly disputes attribution and retains an ordinary optional-return error in its older reduction. Our independent ablation removes TS18048 by disabling noUncheckedIndexedAccess. Scheduling this row is not proof that this exact slice lowers. |
| Emitter D117 | Stricter options #k881crd; codex/stricter-optional-writes e2f994418282045eb65590b10bd5d93ae80b2bac | optional-writes/REPORT.md names moduleNameResolver.ts:2772 in its final recursive-object contract group; indexed-all states schedules D117 optional-presence. Own presence, including undefined-valued fields, must survive. Production reports zero emitted original-program checks. |

Checked casts #b5w3ycg candidates codex/checked-downcasts 982e2118f968788464256005e840b80e9b193254 and codex/interface-downcasts ab4d6f902ea0614baaa4f9b4e9b93823ef895741 are downstream cast support, not established admission for these errors. Placeholders #9wc5q5j candidates codex/placeholder-nonnull c8f858b979e7461762b1742d0982dd73567b9a74 and compiler/rehearsal-placeholders 678d94f95c07e0572e61cf2193fd5014bcbe1d1e address staged placeholder use. The checker statement asserts the array but does not assert the selected element. Neither task is demonstrated to clear TS18048 or the emitter's exact-optional relation.

These options branches retain strict .a policy while admitting supported project-.ts relations through scheduled guard contracts. A clean reduced .a build would not alone prove the original project admission path. Conversely, a strict .a rejection does not falsify a project-.ts contract.

## Scratch attempts and infrastructure

Detached scratch worktrees attempted each candidate merge on both pinned bases. All four git merges exit 1 with conflicts, including loader and compiler files. evidence/scratch-results.json records every command, SHA and conflicted path; raw logs are retained. No bulk conflict resolution or diagnostic waiver was made. Candidate merged builds were not run because no merged tree exists. stop_cleared is null, state pending, never pass.

Before setup, GOPROXY was exported as `https://proxy.golang.org|direct`. Current-main setup exits 1 fetching cohere/TypeScript commit d92d9bfee114c80be2c375d72edae966176e3a4f. The old remote and the URL recorded in cohere/.gitmodules both fail with `fatal: could not read Username for 'https://github.com': No such device or address`. Updating local remote configuration did not solve it. The existing nested checkout remains 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, not the required pin. No authentication bypass was attempted.

Setup timing lines: Go ready 0.021s, Node ready 0.022s, clang ready 0.160s, markdown dependency installation 1.728s and ready 1.797s. There is no successful submodule/build/cache completion timing. nproc reports 5; CPU quota is 4. Environment sourced: /workspace/adamic-tools/env.sh; Go 1.27.1, Node 24.19.0, clang 20.1.8.

A separate `go build -o /tmp/step31-current-adamic ./cmd/adamic` also fails. Its complete log begins with undefined tspath.RootedFilePath in the shim and records other missing upstream symbols. This diagnoses the mismatched nested dependency, not either requested compiler stop. No native output comparison or silent-miscompile conclusion is possible.

## Reproduce the completed checks

```
source /workspace/adamic-tools/env.sh
npm install --prefix /workspace/scratch/step31-node --ignore-scripts --no-audit --no-fund typescript@6.0.3 > node-install.log 2>&1
STEP31_TYPESCRIPT=/workspace/scratch/step31-node/node_modules/typescript/lib/typescript.js python3 stage3/scouts/step31/first-stops/verify.py > verify.log 2>&1
```

Observed verifier time about 7 seconds, exit 0: `2 Node witnesses match; 2 exact stock strict diagnostics match; 2 mutants caught; native pending`. The runner uses the repository's Node source runner, never compiler-produced JavaScript. Each responsible-option ablation has zero diagnostics. The expected a-check headers name independently reproduced stock strict codes; in-place stage-0 measurement remains blocked and the headers must be reconfirmed when the required dependency is available.

These scout witnesses are not registered internal/oracle fixtures, so no oracle allocation-count row was added and counts.md was not changed. No whole-package tests or full gate were run. Native confirmation, conflict integration and a newly measured complete checker/emitter slice remain unfinished. Existing scout evidence is clearly separated from new Node observations.
