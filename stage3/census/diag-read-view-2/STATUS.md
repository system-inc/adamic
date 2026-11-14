{
  "implementation": "d7a6aa49976828aa9627b64c137234b718c9e503",
  "base": "8e3a784aa5a601f57342c07b68f145fe56252c21",
  "upstream": "050880ce59e30b356b686bd3144efe24f875ebc8",
  "status": "PROOF_PENDING",
  "note": "v2 applies all eight reader parameter substitutions. User accepts exactly the pre-existing renamed pair; no other new or changed findings are accepted. No Loom submission.",
  "completedResults": {
    "candidate-npm-ci.result.json": {
      "label": "candidate-npm-ci",
      "args": [
        "npm",
        "ci"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/candidate-tree",
      "extra": {},
      "exit": 0,
      "seconds": 17.23017464699933,
      "limit": 6000
    },
    "mutant-program.ts-census.result.json": {
      "label": "mutant-program.ts-census",
      "args": [
        "/workspace/scratch/diag-read-view-2/latent",
        "/workspace/scratch/diag-read-view-2/mutant-tree/src/compiler",
        "/workspace/scratch/diag-read-view-2/raw/mutant-program.ts.jsonl"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {
        "LATENT_FULL": "1",
        "LATENT_ASSERT_NO_OUTPUT": "1",
        "LATENT_ONLY_FILE": "/workspace/scratch/diag-read-view-2/mutant-tree/src/compiler/program.ts",
        "CENSUS_MODULE_FORMAT": "CommonJSBundle"
      },
      "exit": 0,
      "seconds": 1039.127183566,
      "limit": 6000
    },
    "main-build-compiler.result.json": {
      "label": "main-build-compiler",
      "args": [
        "npm",
        "run",
        "build:compiler"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/main-tree",
      "extra": {},
      "exit": 0,
      "seconds": 7.19960618999994,
      "limit": 6000
    },
    "stock-tsc.result.json": {
      "label": "stock-tsc",
      "args": [
        "node",
        "/workspace/scratch/diag-read-view-2/stock.cjs"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {},
      "exit": 0,
      "seconds": 56.43669161699927,
      "limit": 6000
    },
    "main-program.ts-census.result.json": {
      "label": "main-program.ts-census",
      "args": [
        "/workspace/scratch/diag-read-view-2/latent",
        "/workspace/scratch/diag-read-view-2/main-tree/src/compiler",
        "/workspace/scratch/diag-read-view-2/raw/main-program.ts.jsonl"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {
        "LATENT_FULL": "1",
        "LATENT_ASSERT_NO_OUTPUT": "1",
        "LATENT_ONLY_FILE": "/workspace/scratch/diag-read-view-2/main-tree/src/compiler/program.ts",
        "CENSUS_MODULE_FORMAT": "CommonJSBundle"
      },
      "exit": 0,
      "seconds": 1116.9861069489998,
      "limit": 6000
    },
    "candidate-utilitiesPublic.ts-census.result.json": {
      "label": "candidate-utilitiesPublic.ts-census",
      "args": [
        "/workspace/scratch/diag-read-view-2/latent",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler",
        "/workspace/scratch/diag-read-view-2/raw/candidate-utilitiesPublic.ts.jsonl"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {
        "LATENT_FULL": "1",
        "LATENT_ASSERT_NO_OUTPUT": "1",
        "LATENT_ONLY_FILE": "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilitiesPublic.ts",
        "CENSUS_MODULE_FORMAT": "CommonJSBundle"
      },
      "exit": 0,
      "seconds": 215.36307493500044,
      "limit": 6000
    },
    "main-dts.result.json": {
      "label": "main-dts",
      "args": [
        "/workspace/scratch/diag-read-view-2/main-tree/node_modules/.bin/hereby",
        "dts"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/main-tree",
      "extra": {},
      "exit": 0,
      "seconds": 69.20705238899973,
      "limit": 6000
    },
    "candidate-build-compiler.result.json": {
      "label": "candidate-build-compiler",
      "args": [
        "npm",
        "run",
        "build:compiler"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/candidate-tree",
      "extra": {},
      "exit": 0,
      "seconds": 9.406476033000217,
      "limit": 6000
    },
    "candidate-dts.result.json": {
      "label": "candidate-dts",
      "args": [
        "/workspace/scratch/diag-read-view-2/candidate-tree/node_modules/.bin/hereby",
        "dts"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/candidate-tree",
      "extra": {},
      "exit": 0,
      "seconds": 61.5299575250001,
      "limit": 6000
    },
    "candidate-program.ts-census.result.json": {
      "label": "candidate-program.ts-census",
      "args": [
        "/workspace/scratch/diag-read-view-2/latent",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler",
        "/workspace/scratch/diag-read-view-2/raw/candidate-program.ts.jsonl"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {
        "LATENT_FULL": "1",
        "LATENT_ASSERT_NO_OUTPUT": "1",
        "LATENT_ONLY_FILE": "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/program.ts",
        "CENSUS_MODULE_FORMAT": "CommonJSBundle"
      },
      "exit": 0,
      "seconds": 1108.9316885879998,
      "limit": 6000
    },
    "candidate-apply.result.json": {
      "label": "candidate-apply",
      "args": [
        "/workspace/scratch/diag-read-view-2/repo/stage3/apply.sh",
        "/workspace/scratch/diag-read-view-2/applied-tree"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {},
      "exit": 0,
      "seconds": 837.3288123109996,
      "limit": 6000
    },
    "main-utilitiesPublic.ts-census.result.json": {
      "label": "main-utilitiesPublic.ts-census",
      "args": [
        "/workspace/scratch/diag-read-view-2/latent",
        "/workspace/scratch/diag-read-view-2/main-tree/src/compiler",
        "/workspace/scratch/diag-read-view-2/raw/main-utilitiesPublic.ts.jsonl"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {
        "LATENT_FULL": "1",
        "LATENT_ASSERT_NO_OUTPUT": "1",
        "LATENT_ONLY_FILE": "/workspace/scratch/diag-read-view-2/main-tree/src/compiler/utilitiesPublic.ts",
        "CENSUS_MODULE_FORMAT": "CommonJSBundle"
      },
      "exit": 0,
      "seconds": 214.52558809000038,
      "limit": 6000
    },
    "javascript.result.json": {
      "label": "javascript",
      "args": [
        "node",
        "/workspace/scratch/diag-read-view-2/javascript.cjs"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {},
      "exit": 0,
      "seconds": 52.28831945699949,
      "limit": 6000
    },
    "main-npm-ci.result.json": {
      "label": "main-npm-ci",
      "args": [
        "npm",
        "ci"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/main-tree",
      "extra": {},
      "exit": 0,
      "seconds": 15.988421598999594,
      "limit": 6000
    },
    "candidate-adamic-types.result.json": {
      "label": "candidate-adamic-types",
      "args": [
        "/workspace/scratch/diag-read-view-2/adamic",
        "types",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/_namespaces/ts.moduleSpecifiers.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/_namespaces/ts.performance.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/_namespaces/ts.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/binder.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/builder.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/builderPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/builderState.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/builderStatePublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/checker.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/commandLineParser.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/core.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/corePublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/debug.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/diagnosticInformationMap.generated.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/emitter.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/executeCommandLine.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/expressionToTypeNode.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/baseNodeFactory.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/emitHelpers.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/emitNode.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/nodeChildren.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/nodeConverters.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/nodeFactory.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/nodeTests.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/parenthesizerRules.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/utilities.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/factory/utilitiesPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/hostErrors.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/moduleNameResolver.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/moduleSpecifiers.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/parser.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/path.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/performance.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/performanceCore.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/program.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/programDiagnostics.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/resolutionCache.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/scanner.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/semver.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/sourcemap.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/symbolWalker.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/sys.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/tracing.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformer.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/classFields.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/classThis.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/declarations/diagnostics.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/declarations.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/destructuring.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2015.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2016.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2017.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2018.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2019.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2020.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/es2021.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/esDecorators.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/esnext.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/generators.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/jsx.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/legacyDecorators.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/module/esnextAnd2015.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/module/impliedNodeFormatDependent.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/module/module.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/module/system.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/namedEvaluation.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/taggedTemplate.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/ts.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/typeSerializer.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/transformers/utilities.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/tsbuild.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/tsbuildPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/types.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilities.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilitiesPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/visitorPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/watch.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/watchPublic.ts",
        "/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/watchUtilities.ts"
      ],
      "cwd": "/workspace/scratch/diag-read-view-2/repo",
      "extra": {},
      "exit": 0,
      "seconds": 17.330257924999387,
      "limit": 6000
    }
  }
}
