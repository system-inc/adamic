# Commands and exits

Setup, initial failures, branch creation and checkpoint commands are also listed in SETUP.md. Each command below ran in the cwd recorded in commands.jsonl.gz, with exact arguments and selected environment retained there. Log names are the command name plus .log.gz.

- Exit **0**, overlay (4.552 seconds):

  ```sh
  python3 stage3/census/latent/make_overlay.py /workspace/scratch/single-watcher-forwarder /workspace/scratch/single-watcher-overlay
  ```

- Exit **0**, overlay-gofmt (0.043 seconds):

  ```sh
  bash -c 'gofmt -w /workspace/scratch/single-watcher-overlay/*.go'
  ```

- Exit **0**, generate-api (12.393 seconds):

  ```sh
  python3 stage3/lane/generate-api.py /workspace/scratch/single-watcher-typescript
  ```

- Exit **0**, build-latent (260.606 seconds):

  ```sh
  go build -buildvcs=false -overlay=/workspace/scratch/single-watcher-overlay/overlay.json -o /workspace/scratch/single-watcher-evidence/latent ./stage3/census/latent/tool
  ```

- Exit **2**, generate-plan (0.045 seconds):

  ```sh
  python3 stage3/lane/shard_plan.py
  ```

- Exit **0**, generate-shards-check (0.170 seconds):

  ```sh
  python3 stage3/lane/generate_shards.py --check
  ```

- Exit **0**, latent-audit (0.380 seconds):

  ```sh
  python3 stage3/census/latent/audit.py /workspace/scratch/single-watcher-evidence/latent
  ```

- Exit **0**, build-adamic-receipt (0.655 seconds):

  ```sh
  go build -o /workspace/scratch/single-watcher-evidence/adamic ./cmd/adamic
  ```

- Exit **0**, prepare (0.505 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/prepare.cjs
  ```

- Exit **0**, contract (0.679 seconds):

  ```sh
  node stage3/adapt/99-correctness-single-watcher-forwarder/contract.cjs /workspace/scratch/single-watcher-before
  ```

- Exit **0**, stock-base (13.883 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/stock.cjs /workspace/scratch/single-watcher-before /workspace/scratch/single-watcher-evidence/base
  ```

- Exit **0**, stock-adapted (13.969 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/stock.cjs /workspace/scratch/single-watcher-after /workspace/scratch/single-watcher-evidence/adapted
  ```

- Exit **1**, runtime (0.594 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/runtime.cjs /workspace/scratch/single-watcher-evidence/base /workspace/scratch/single-watcher-evidence/adapted /workspace/scratch/single-watcher-evidence/runtime-traces.json
  ```

- Exit **0**, scope (0.647 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/scope.cjs /workspace/scratch/single-watcher-before /workspace/scratch/single-watcher-after /workspace/scratch/single-watcher-evidence/base /workspace/scratch/single-watcher-evidence/adapted
  ```

- Exit **0**, types-base (6.395 seconds):

  ```sh
  /workspace/scratch/single-watcher-evidence/adamic types /workspace/scratch/single-watcher-before/src/compiler/_namespaces/ts.moduleSpecifiers.ts /workspace/scratch/single-watcher-before/src/compiler/_namespaces/ts.performance.ts /workspace/scratch/single-watcher-before/src/compiler/_namespaces/ts.ts /workspace/scratch/single-watcher-before/src/compiler/binder.ts /workspace/scratch/single-watcher-before/src/compiler/builder.ts /workspace/scratch/single-watcher-before/src/compiler/builderPublic.ts /workspace/scratch/single-watcher-before/src/compiler/builderState.ts /workspace/scratch/single-watcher-before/src/compiler/builderStatePublic.ts /workspace/scratch/single-watcher-before/src/compiler/checker.ts /workspace/scratch/single-watcher-before/src/compiler/commandLineParser.ts /workspace/scratch/single-watcher-before/src/compiler/core.ts /workspace/scratch/single-watcher-before/src/compiler/corePublic.ts /workspace/scratch/single-watcher-before/src/compiler/debug.ts /workspace/scratch/single-watcher-before/src/compiler/diagnosticInformationMap.generated.ts /workspace/scratch/single-watcher-before/src/compiler/emitter.ts /workspace/scratch/single-watcher-before/src/compiler/executeCommandLine.ts /workspace/scratch/single-watcher-before/src/compiler/expressionToTypeNode.ts /workspace/scratch/single-watcher-before/src/compiler/factory/baseNodeFactory.ts /workspace/scratch/single-watcher-before/src/compiler/factory/emitHelpers.ts /workspace/scratch/single-watcher-before/src/compiler/factory/emitNode.ts /workspace/scratch/single-watcher-before/src/compiler/factory/nodeChildren.ts /workspace/scratch/single-watcher-before/src/compiler/factory/nodeConverters.ts /workspace/scratch/single-watcher-before/src/compiler/factory/nodeFactory.ts /workspace/scratch/single-watcher-before/src/compiler/factory/nodeTests.ts /workspace/scratch/single-watcher-before/src/compiler/factory/parenthesizerRules.ts /workspace/scratch/single-watcher-before/src/compiler/factory/utilities.ts /workspace/scratch/single-watcher-before/src/compiler/factory/utilitiesPublic.ts /workspace/scratch/single-watcher-before/src/compiler/hostErrors.ts /workspace/scratch/single-watcher-before/src/compiler/moduleNameResolver.ts /workspace/scratch/single-watcher-before/src/compiler/moduleSpecifiers.ts /workspace/scratch/single-watcher-before/src/compiler/parser.ts /workspace/scratch/single-watcher-before/src/compiler/path.ts /workspace/scratch/single-watcher-before/src/compiler/performance.ts /workspace/scratch/single-watcher-before/src/compiler/performanceCore.ts /workspace/scratch/single-watcher-before/src/compiler/program.ts /workspace/scratch/single-watcher-before/src/compiler/programDiagnostics.ts /workspace/scratch/single-watcher-before/src/compiler/resolutionCache.ts /workspace/scratch/single-watcher-before/src/compiler/scanner.ts /workspace/scratch/single-watcher-before/src/compiler/semver.ts /workspace/scratch/single-watcher-before/src/compiler/sourcemap.ts /workspace/scratch/single-watcher-before/src/compiler/symbolWalker.ts /workspace/scratch/single-watcher-before/src/compiler/sys.ts /workspace/scratch/single-watcher-before/src/compiler/tracing.ts /workspace/scratch/single-watcher-before/src/compiler/transformer.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/classFields.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/classThis.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/declarations.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/declarations/diagnostics.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/destructuring.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2015.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2016.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2017.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2018.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2019.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2020.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/es2021.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/esDecorators.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/esnext.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/generators.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/jsx.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/legacyDecorators.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/module/esnextAnd2015.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/module/impliedNodeFormatDependent.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/module/module.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/module/system.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/namedEvaluation.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/taggedTemplate.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/ts.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/typeSerializer.ts /workspace/scratch/single-watcher-before/src/compiler/transformers/utilities.ts /workspace/scratch/single-watcher-before/src/compiler/tsbuild.ts /workspace/scratch/single-watcher-before/src/compiler/tsbuildPublic.ts /workspace/scratch/single-watcher-before/src/compiler/types.ts /workspace/scratch/single-watcher-before/src/compiler/utilities.ts /workspace/scratch/single-watcher-before/src/compiler/utilitiesPublic.ts /workspace/scratch/single-watcher-before/src/compiler/visitorPublic.ts /workspace/scratch/single-watcher-before/src/compiler/watch.ts /workspace/scratch/single-watcher-before/src/compiler/watchPublic.ts /workspace/scratch/single-watcher-before/src/compiler/watchUtilities.ts
  ```

- Exit **0**, types-adapted (4.706 seconds):

  ```sh
  /workspace/scratch/single-watcher-evidence/adamic types /workspace/scratch/single-watcher-after/src/compiler/_namespaces/ts.moduleSpecifiers.ts /workspace/scratch/single-watcher-after/src/compiler/_namespaces/ts.performance.ts /workspace/scratch/single-watcher-after/src/compiler/_namespaces/ts.ts /workspace/scratch/single-watcher-after/src/compiler/binder.ts /workspace/scratch/single-watcher-after/src/compiler/builder.ts /workspace/scratch/single-watcher-after/src/compiler/builderPublic.ts /workspace/scratch/single-watcher-after/src/compiler/builderState.ts /workspace/scratch/single-watcher-after/src/compiler/builderStatePublic.ts /workspace/scratch/single-watcher-after/src/compiler/checker.ts /workspace/scratch/single-watcher-after/src/compiler/commandLineParser.ts /workspace/scratch/single-watcher-after/src/compiler/core.ts /workspace/scratch/single-watcher-after/src/compiler/corePublic.ts /workspace/scratch/single-watcher-after/src/compiler/debug.ts /workspace/scratch/single-watcher-after/src/compiler/diagnosticInformationMap.generated.ts /workspace/scratch/single-watcher-after/src/compiler/emitter.ts /workspace/scratch/single-watcher-after/src/compiler/executeCommandLine.ts /workspace/scratch/single-watcher-after/src/compiler/expressionToTypeNode.ts /workspace/scratch/single-watcher-after/src/compiler/factory/baseNodeFactory.ts /workspace/scratch/single-watcher-after/src/compiler/factory/emitHelpers.ts /workspace/scratch/single-watcher-after/src/compiler/factory/emitNode.ts /workspace/scratch/single-watcher-after/src/compiler/factory/nodeChildren.ts /workspace/scratch/single-watcher-after/src/compiler/factory/nodeConverters.ts /workspace/scratch/single-watcher-after/src/compiler/factory/nodeFactory.ts /workspace/scratch/single-watcher-after/src/compiler/factory/nodeTests.ts /workspace/scratch/single-watcher-after/src/compiler/factory/parenthesizerRules.ts /workspace/scratch/single-watcher-after/src/compiler/factory/utilities.ts /workspace/scratch/single-watcher-after/src/compiler/factory/utilitiesPublic.ts /workspace/scratch/single-watcher-after/src/compiler/hostErrors.ts /workspace/scratch/single-watcher-after/src/compiler/moduleNameResolver.ts /workspace/scratch/single-watcher-after/src/compiler/moduleSpecifiers.ts /workspace/scratch/single-watcher-after/src/compiler/parser.ts /workspace/scratch/single-watcher-after/src/compiler/path.ts /workspace/scratch/single-watcher-after/src/compiler/performance.ts /workspace/scratch/single-watcher-after/src/compiler/performanceCore.ts /workspace/scratch/single-watcher-after/src/compiler/program.ts /workspace/scratch/single-watcher-after/src/compiler/programDiagnostics.ts /workspace/scratch/single-watcher-after/src/compiler/resolutionCache.ts /workspace/scratch/single-watcher-after/src/compiler/scanner.ts /workspace/scratch/single-watcher-after/src/compiler/semver.ts /workspace/scratch/single-watcher-after/src/compiler/sourcemap.ts /workspace/scratch/single-watcher-after/src/compiler/symbolWalker.ts /workspace/scratch/single-watcher-after/src/compiler/sys.ts /workspace/scratch/single-watcher-after/src/compiler/tracing.ts /workspace/scratch/single-watcher-after/src/compiler/transformer.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/classFields.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/classThis.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/declarations.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/declarations/diagnostics.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/destructuring.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2015.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2016.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2017.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2018.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2019.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2020.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/es2021.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/esDecorators.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/esnext.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/generators.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/jsx.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/legacyDecorators.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/module/esnextAnd2015.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/module/impliedNodeFormatDependent.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/module/module.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/module/system.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/namedEvaluation.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/taggedTemplate.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/ts.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/typeSerializer.ts /workspace/scratch/single-watcher-after/src/compiler/transformers/utilities.ts /workspace/scratch/single-watcher-after/src/compiler/tsbuild.ts /workspace/scratch/single-watcher-after/src/compiler/tsbuildPublic.ts /workspace/scratch/single-watcher-after/src/compiler/types.ts /workspace/scratch/single-watcher-after/src/compiler/utilities.ts /workspace/scratch/single-watcher-after/src/compiler/utilitiesPublic.ts /workspace/scratch/single-watcher-after/src/compiler/visitorPublic.ts /workspace/scratch/single-watcher-after/src/compiler/watch.ts /workspace/scratch/single-watcher-after/src/compiler/watchPublic.ts /workspace/scratch/single-watcher-after/src/compiler/watchUtilities.ts
  ```

- Exit **0**, runtime-fixed (0.358 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/runtime.cjs /workspace/scratch/single-watcher-evidence/base /workspace/scratch/single-watcher-evidence/adapted /workspace/scratch/single-watcher-evidence/runtime-traces.json
  ```

- Exit **0**, count (4.020 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/count.cjs /workspace/scratch/single-watcher-after /workspace/scratch/single-watcher-evidence/callers.json
  ```

- Exit **0**, census-base (99.125 seconds):

  ```sh
  timeout --kill-after=10s 1800 /workspace/scratch/single-watcher-evidence/latent /workspace/scratch/single-watcher-before/src/compiler /workspace/scratch/single-watcher-evidence/base/latent.jsonl
  ```

- Exit **0**, stock-mutant (19.386 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/stock.cjs /workspace/scratch/single-watcher-mutant /workspace/scratch/single-watcher-evidence/mutant
  ```

- Exit **0**, npm-ci-scratch (1.674 seconds):

  ```sh
  npm ci --prefix stage3/api
  ```

- Exit **0**, generated-freshness (0.287 seconds):

  ```sh
  python3 stage3/lane/generate_shards.py --check
  ```

- Exit **0**, census-adapted (99.139 seconds):

  ```sh
  timeout --kill-after=10s 1800 /workspace/scratch/single-watcher-evidence/latent /workspace/scratch/single-watcher-after/src/compiler /workspace/scratch/single-watcher-evidence/adapted/latent.jsonl
  ```

- Exit **0**, runtime-after-event (0.518 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/runtime.cjs /workspace/scratch/single-watcher-evidence/base /workspace/scratch/single-watcher-evidence/adapted /workspace/scratch/single-watcher-evidence/runtime-traces.json
  ```

- Exit **0**, census-mutant (103.158 seconds):

  ```sh
  timeout --kill-after=10s 1800 /workspace/scratch/single-watcher-evidence/latent /workspace/scratch/single-watcher-mutant/src/compiler /workspace/scratch/single-watcher-evidence/mutant/latent.jsonl
  ```

- Exit **1**, census-compare (0.538 seconds):

  ```sh
  python3 /workspace/scratch/single-watcher-evidence/compare.py
  ```

- Exit **0**, census-compare-fixed (0.123 seconds):

  ```sh
  python3 /workspace/scratch/single-watcher-evidence/compare.py
  ```

- Exit **0**, verify (0.049 seconds):

  ```sh
  node /workspace/scratch/single-watcher-evidence/verify.cjs
  ```

- Exit **0**, product-key (7.665 seconds):

  ```sh
  stage3/apply.sh --key
  ```
