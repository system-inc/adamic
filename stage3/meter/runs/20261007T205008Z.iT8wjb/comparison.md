Measured main p2 with the untouched merged checkout.

Whole program: main: 2/79; area: 2/79
Own file: main: 55/79; area: 55/79

Compiler: 84d5eadf00b30a540b9aa4dd79ce8a690aae9136. Embedded census VCS revision matches this SHA, with vcs.modified=false; see compiler-build-info.log.
No compiler files outside stage3 differ from main 48c05d091. input-comparison.json records all 82 adapted files on each tree as byte-identical to the 20:16 run. The checker gains on those identical bytes measure the compiler change.

No own-file or whole-program regressions against either the 07:35 or 20261007T201631Z baseline. No source files were added or removed. Each tree gained 29 own-file passes and one whole-program pass (src/compiler/corePublic.ts). The gained-file lists are identical against both baselines.

| File | Main fail to pass | Area fail to pass |
| --- | --- | --- |
| src/compiler/binder.ts | yes | yes |
| src/compiler/builderState.ts | yes | yes |
| src/compiler/corePublic.ts | yes | yes |
| src/compiler/debug.ts | yes | yes |
| src/compiler/emitter.ts | yes | yes |
| src/compiler/executeCommandLine.ts | yes | yes |
| src/compiler/expressionToTypeNode.ts | yes | yes |
| src/compiler/factory/emitHelpers.ts | yes | yes |
| src/compiler/factory/nodeFactory.ts | no | yes |
| src/compiler/factory/parenthesizerRules.ts | yes | yes |
| src/compiler/factory/utilities.ts | yes | yes |
| src/compiler/moduleSpecifiers.ts | yes | yes |
| src/compiler/parser.ts | yes | yes |
| src/compiler/path.ts | yes | yes |
| src/compiler/scanner.ts | yes | yes |
| src/compiler/transformers/declarations/diagnostics.ts | yes | yes |
| src/compiler/transformers/destructuring.ts | yes | yes |
| src/compiler/transformers/es2018.ts | yes | yes |
| src/compiler/transformers/legacyDecorators.ts | yes | yes |
| src/compiler/transformers/module/module.ts | yes | yes |
| src/compiler/transformers/module/system.ts | yes | yes |
| src/compiler/transformers/taggedTemplate.ts | yes | yes |
| src/compiler/transformers/ts.ts | yes | yes |
| src/compiler/transformers/utilities.ts | yes | yes |
| src/compiler/tsbuild.ts | yes | yes |
| src/compiler/types.ts | yes | yes |
| src/compiler/utilities.ts | yes | yes |
| src/compiler/utilitiesPublic.ts | yes | yes |
| src/compiler/watch.ts | yes | no |
| src/compiler/watchUtilities.ts | yes | yes |

Commands run from /tmp/stage3-meter-p2-merged:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stage3-meter-p2-merged-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export PATH=/workspace/adamic-tools/bin:$PATH
node --version # v24.19.0
nproc # 5
STAGE3_METER_RUNS=/workspace/adamic/stage3/meter/runs bash stage3/meter/twice-daily.sh > /tmp/stage3-meter-p2-merged.log 2>&1; echo "exit=$?" # exit=0
```

Setup exited 1 in stage3/adapt/75-optional-widening/coverage/checker.go:49:33: a string filename is passed where the pinned checker requires tspath.RootedFilePath. setup.log retains the exact error and all printed timing lines. Targeted ordinary and latent census builds succeeded, so the meter completed without editing that unrelated coverage package.

The full repository gate and native/Node differential oracle were not run. Latent counts measure observed blockers on checker-rejected programs, not exhaustive lowering or successful compilation.

main latent totals: {"NotYet": 1470, "Refused": 4684, "SkippedDependency": 3, "error": 0, "panic": 3}

area latent totals: {"NotYet": 1468, "Refused": 5059, "SkippedDependency": 3, "error": 0, "panic": 3}
