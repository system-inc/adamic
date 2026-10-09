Built a hash-keyed adapted-tree store and independent Python/corpus-file test shards.
Base: origin/main 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8; implementation commit: this commit.
Measured 178 fresh-process invocations: maximum 15.788 s; original apply test 151.065 s.
Seventeen new guard, partition, and planted-source mutants were caught; planted failures had exactly one owning shard.
Developer tools owns remote publication; historical proof measurement gaps remain listed below.

# Developer tools publish hook

Owner: developer tools publisher #x2651cf, kirkouimet assets, restricted to adamic/build-cache/. This script performs no upload and reads no cache credential. The gate runs preparation once, outside test units:

```bash
source /workspace/adamic-tools/env.sh
export STAGE3_PRODUCT_STORE=/absolute/local/stage3-products
bash stage3/apply.sh --key > /tmp/stage3-product-key.txt
bash stage3/apply.sh --build-product /absolute/new/adapted-tree > /tmp/stage3-product-build.log 2>&1
bash stage3/apply.sh --publish-hook > /tmp/stage3-product-hook.json
```

The hook JSON supplies format, key, inputs, pinned_source and payloads. Each payload has an absolute local path and an asset destination under adamic/build-cache/. Inputs are pinned stage3/source.json, apply.py, API package/lock, all adaptation entries (excluding generated __pycache__), and node --version. Names, modes, file bytes, directories, and file symlink targets participate in the hash; directory symlinks are refused. The payload is a gzip tar of the complete adapted tree including .git snapshot refs and patch-set.md, split into 16 MiB keyed .pNNN files, plus a keyed .manifest containing part and complete archive sizes and SHA256 hashes. The hook verifies each part before describing it. Developer tools publishes parts first and the manifest last, using only its own publisher and credentials.

For fetches, ADAMIC_BUILD_CACHE_URL must be the public URL corresponding to the published adamic/build-cache/ asset directory. Requests append KEY.manifest or KEY.pNNN to that base. Existing workers.dev is the fallback base for compatibility; developer tools supplies the correct asset-serving base if different. STAGE3_PRODUCT_STORE can instead select a local directory or explicit URL. Missing manifest builds locally with `stage3 product cache miss: KEY; building adapted tree locally` on stderr and saves the result in the local product store. On a remote miss, that store defaults to STAGE3_CACHE/products. There is no remote write. Corruption is fatal rather than a cache miss.

The test suite now replaces the earlier never-build-on-miss test with build-and-report coverage, plus corrupt-product refusal, local hook contract, and precise missing-manifest classification. Earlier timings and mutant records below describe the initial revision and are retained as historical evidence. Current-revision timings and mutants are recorded separately in the follow-up validation section.

# Measurements and limits

Measurements used this Codex instance: nproc=5, cgroup CPU quota=4. Each invocation started a new process with fresh test output directories after product preparation. This is process/output cold, not an OS page-cache flush. Product retrieval was from the local keyed store; remote latency was not measured. The pinned API parser was already installed from its lockfile. No checks were removed.

Setup: GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Go ready 0.080 s; Node 0.102 s; clang 0.591 s; markdown dependencies 2.051 s; submodules 31.974 s; Go build 495.576 s; build cache warm 496.027 s; done 496.130 s. Uncontended explicit product preparation took 127.870 s and is not a test unit.

Original command: python3 stage3/test_apply.py in a detached origin/main checkout: 151.065 s, exit 0. The ordinary case now takes 14.159 s. Four indexed-read whole-script measurements retain the original checks; their new per-file units use sorted exported corpus files modulo shard count. Before/after trees were exported from adaptation snapshot refs.

Run python3 stage3/test_apply.py --list-shards N to list every Python case and owner. ADAMIC_TEST_SHARD=i/N selects its cases. ADAMIC_TEST_LIST_SHARDS=1 ADAMIC_TEST_SHARD=0/N node stage3/adapt/SLUG/verify.cjs BEFORE AFTER lists every corpus file and owner. The partition tests check several counts, including more shards than cases, for exact union and unique ownership.

Product preparation: bash stage3/apply.sh --build-product OUTPUT. Ordinary bash stage3/apply.sh OUTPUT retrieves by the hash of pinned source metadata, every adaptation directory, apply.py, API locks, and Node version. STAGE3_PRODUCT_STORE selects a local store; ADAMIC_BUILD_CACHE_URL selects the remote store. As requested in the October 8 follow-up, absent manifests (local missing file or HTTP 404) now cause an explicit logged local build. Corrupt products, missing parts, authorization errors, and other retrieval failures still fail without rebuilding. A missing-cache build exceeds 30 seconds; only tests after fetching a published product satisfy the measured budget. Products contain git snapshot refs needed by proof helpers. Manifest and each part are verified; extracted paths use the data tar filter. Source tables change only with --write-table.

# Before and after table

| Unit | Before seconds | After seconds | Exit |
| --- | ---: | ---: | ---: |
| apply-ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean | 151.065 | 14.159 | 0 |
| apply-ProductTests.test_existing_output_is_refused | new | 0.150 | 0 |
| apply-ProductTests.test_input_key_covers_all_inputs | new | 0.224 | 0 |
| apply-ProductTests.test_manifest_mutant_is_rejected | new | 0.145 | 0 |
| apply-ProductTests.test_missing_product_never_builds | new | 0.135 | 0 |
| apply-ProductTests.test_payload_manifest_mutant_is_rejected | new | 0.135 | 0 |
| apply-ProductTests.test_payload_mutant_is_rejected | new | 0.135 | 0 |
| apply-ProductTests.test_restore_fetches_verified_product | new | 0.153 | 0 |
| apply-ProductTests.test_shards_partition_every_case | new | 0.130 | 0 |
| apply-ProductTests.test_unkeyed_parser_override_is_refused | new | 0.136 | 0 |
| apply-ProductTests.test_write_table_is_explicit | new | 0.146 | 0 |
| apply-ProofShardTests.test_every_indexed_read_file_has_one_shard | new | 0.223 | 0 |
| 30-indexed-reads-whole | unchanged checks | 3.551 | 0 |
| 30-indexed-reads-core.ts | unchanged checks | 0.552 | 0 |
| 30-indexed-reads-debug.ts | unchanged checks | 0.517 | 0 |
| 30-indexed-reads-factory_baseNodeFactory.ts | unchanged checks | 0.253 | 0 |
| 30-indexed-reads-factory_emitHelpers.ts | unchanged checks | 0.389 | 0 |
| 30-indexed-reads-factory_emitNode.ts | unchanged checks | 0.334 | 0 |
| 30-indexed-reads-factory_nodeChildren.ts | unchanged checks | 0.276 | 0 |
| 30-indexed-reads-factory_nodeConverters.ts | unchanged checks | 0.296 | 0 |
| 30-indexed-reads-factory_nodeFactory.ts | unchanged checks | 0.894 | 0 |
| 30-indexed-reads-factory_nodeTests.ts | unchanged checks | 0.357 | 0 |
| 30-indexed-reads-factory_parenthesizerRules.ts | unchanged checks | 0.328 | 0 |
| 30-indexed-reads-factory_utilities.ts | unchanged checks | 0.458 | 0 |
| 30-indexed-reads-factory_utilitiesPublic.ts | unchanged checks | 0.282 | 0 |
| 30-indexed-reads-parser.ts | unchanged checks | 1.439 | 0 |
| 30-indexed-reads-path.ts | unchanged checks | 0.416 | 0 |
| 30-indexed-reads-scanner.ts | unchanged checks | 0.885 | 0 |
| 30-indexed-reads-utilities.ts | unchanged checks | 1.463 | 0 |
| 30-indexed-reads-utilitiesPublic.ts | unchanged checks | 0.592 | 0 |
| 31-indexed-reads-checker-whole | unchanged checks | 3.990 | 0 |
| 31-indexed-reads-checker-checker.ts | unchanged checks | 3.828 | 0 |
| 32-indexed-reads-program-whole | unchanged checks | 3.517 | 0 |
| 32-indexed-reads-program-binder.ts | unchanged checks | 0.739 | 0 |
| 32-indexed-reads-program-builder.ts | unchanged checks | 0.586 | 0 |
| 32-indexed-reads-program-builderPublic.ts | unchanged checks | 0.336 | 0 |
| 32-indexed-reads-program-builderState.ts | unchanged checks | 0.417 | 0 |
| 32-indexed-reads-program-builderStatePublic.ts | unchanged checks | 0.253 | 0 |
| 32-indexed-reads-program-commandLineParser.ts | unchanged checks | 0.687 | 0 |
| 32-indexed-reads-program-corePublic.ts | unchanged checks | 0.268 | 0 |
| 32-indexed-reads-program-executeCommandLine.ts | unchanged checks | 0.423 | 0 |
| 32-indexed-reads-program-expressionToTypeNode.ts | unchanged checks | 0.498 | 0 |
| 32-indexed-reads-program-moduleNameResolver.ts | unchanged checks | 0.712 | 0 |
| 32-indexed-reads-program-moduleSpecifiers.ts | unchanged checks | 0.658 | 0 |
| 32-indexed-reads-program-performance.ts | unchanged checks | 0.310 | 0 |
| 32-indexed-reads-program-performanceCore.ts | unchanged checks | 0.275 | 0 |
| 32-indexed-reads-program-program.ts | unchanged checks | 0.854 | 0 |
| 32-indexed-reads-program-programDiagnostics.ts | unchanged checks | 0.369 | 0 |
| 32-indexed-reads-program-resolutionCache.ts | unchanged checks | 0.452 | 0 |
| 32-indexed-reads-program-semver.ts | unchanged checks | 0.361 | 0 |
| 32-indexed-reads-program-symbolWalker.ts | unchanged checks | 0.326 | 0 |
| 32-indexed-reads-program-sys.ts | unchanged checks | 0.475 | 0 |
| 32-indexed-reads-program-tracing.ts | unchanged checks | 0.362 | 0 |
| 32-indexed-reads-program-tsbuild.ts | unchanged checks | 0.308 | 0 |
| 32-indexed-reads-program-tsbuildPublic.ts | unchanged checks | 0.619 | 0 |
| 32-indexed-reads-program-types.ts | unchanged checks | 1.122 | 0 |
| 32-indexed-reads-program-watch.ts | unchanged checks | 0.429 | 0 |
| 32-indexed-reads-program-watchPublic.ts | unchanged checks | 0.408 | 0 |
| 32-indexed-reads-program-watchUtilities.ts | unchanged checks | 0.500 | 0 |
| 33-indexed-reads-emit-whole | unchanged checks | 2.749 | 0 |
| 33-indexed-reads-emit-emitter.ts | unchanged checks | 0.859 | 0 |
| 33-indexed-reads-emit-sourcemap.ts | unchanged checks | 0.403 | 0 |
| 33-indexed-reads-emit-transformer.ts | unchanged checks | 0.386 | 0 |
| 33-indexed-reads-emit-transformers_classFields.ts | unchanged checks | 0.552 | 0 |
| 33-indexed-reads-emit-transformers_classThis.ts | unchanged checks | 0.277 | 0 |
| 33-indexed-reads-emit-transformers_declarations.ts | unchanged checks | 0.459 | 0 |
| 33-indexed-reads-emit-transformers_declarations_diagnostics.ts | unchanged checks | 0.405 | 0 |
| 33-indexed-reads-emit-transformers_destructuring.ts | unchanged checks | 0.362 | 0 |
| 33-indexed-reads-emit-transformers_es2015.ts | unchanged checks | 0.705 | 0 |
| 33-indexed-reads-emit-transformers_es2016.ts | unchanged checks | 0.287 | 0 |
| 33-indexed-reads-emit-transformers_es2017.ts | unchanged checks | 0.393 | 0 |
| 33-indexed-reads-emit-transformers_es2018.ts | unchanged checks | 0.457 | 0 |
| 33-indexed-reads-emit-transformers_es2019.ts | unchanged checks | 0.274 | 0 |
| 33-indexed-reads-emit-transformers_es2020.ts | unchanged checks | 0.345 | 0 |
| 33-indexed-reads-emit-transformers_es2021.ts | unchanged checks | 0.277 | 0 |
| 33-indexed-reads-emit-transformers_esDecorators.ts | unchanged checks | 0.515 | 0 |
| 33-indexed-reads-emit-transformers_esnext.ts | unchanged checks | 0.379 | 0 |
| 33-indexed-reads-emit-transformers_generators.ts | unchanged checks | 0.546 | 0 |
| 33-indexed-reads-emit-transformers_jsx.ts | unchanged checks | 0.413 | 0 |
| 33-indexed-reads-emit-transformers_legacyDecorators.ts | unchanged checks | 0.377 | 0 |
| 33-indexed-reads-emit-transformers_module_esnextAnd2015.ts | unchanged checks | 0.357 | 0 |
| 33-indexed-reads-emit-transformers_module_impliedNodeFormatDependent.ts | unchanged checks | 0.270 | 0 |
| 33-indexed-reads-emit-transformers_module_module.ts | unchanged checks | 0.478 | 0 |
| 33-indexed-reads-emit-transformers_module_system.ts | unchanged checks | 0.475 | 0 |
| 33-indexed-reads-emit-transformers_namedEvaluation.ts | unchanged checks | 0.338 | 0 |
| 33-indexed-reads-emit-transformers_taggedTemplate.ts | unchanged checks | 0.298 | 0 |
| 33-indexed-reads-emit-transformers_ts.ts | unchanged checks | 0.544 | 0 |
| 33-indexed-reads-emit-transformers_typeSerializer.ts | unchanged checks | 0.363 | 0 |
| 33-indexed-reads-emit-transformers_utilities.ts | unchanged checks | 0.424 | 0 |
| 33-indexed-reads-emit-visitorPublic.ts | unchanged checks | 0.436 | 0 |
| 45-regex-captures-test | unchanged checks | 8.326 | 0 |
| 46-fix-pragma-test | unchanged checks | 2.934 | 0 |
| 45-regex-captures-evidence | unchanged checks | 0.037 | 0 |
| 41-constructors | unchanged checks | 2.684 | 0 |
| 41-debugger | unchanged checks | 0.407 | 0 |
| 41-guards | unchanged checks | 2.644 | 0 |
| 47-host-errors | unchanged checks | 3.061 | 0 |
| 48-memoize | unchanged checks | 1.385 | 0 |
| 65-node-builtins | unchanged checks | 1.419 | 0 |
| 76-truthful-casts | unchanged checks | 3.480 | 0 |

# Shard mutant runs

| Unit | Seconds | Exit |
| --- | ---: | ---: |
| 30-indexed-reads-planted-0 | 0.475 | 1 |
| 30-indexed-reads-planted-1 | 0.497 | 0 |
| 30-indexed-reads-planted-2 | 0.281 | 0 |
| 30-indexed-reads-planted-3 | 0.379 | 0 |
| 30-indexed-reads-planted-4 | 0.382 | 0 |
| 30-indexed-reads-planted-5 | 0.286 | 0 |
| 30-indexed-reads-planted-6 | 0.317 | 0 |
| 30-indexed-reads-planted-7 | 0.881 | 0 |
| 30-indexed-reads-planted-8 | 0.368 | 0 |
| 30-indexed-reads-planted-9 | 0.335 | 0 |
| 30-indexed-reads-planted-10 | 0.501 | 0 |
| 30-indexed-reads-planted-11 | 0.263 | 0 |
| 30-indexed-reads-planted-12 | 1.438 | 0 |
| 30-indexed-reads-planted-13 | 0.413 | 0 |
| 30-indexed-reads-planted-14 | 0.857 | 0 |
| 30-indexed-reads-planted-15 | 1.432 | 0 |
| 30-indexed-reads-planted-16 | 0.537 | 0 |
| 31-indexed-reads-checker-planted-0 | 3.331 | 1 |
| 32-indexed-reads-program-planted-0 | 0.568 | 1 |
| 32-indexed-reads-program-planted-1 | 0.611 | 0 |
| 32-indexed-reads-program-planted-2 | 0.298 | 0 |
| 32-indexed-reads-program-planted-3 | 0.372 | 0 |
| 32-indexed-reads-program-planted-4 | 0.244 | 0 |
| 32-indexed-reads-program-planted-5 | 0.694 | 0 |
| 32-indexed-reads-program-planted-6 | 0.272 | 0 |
| 32-indexed-reads-program-planted-7 | 0.478 | 0 |
| 32-indexed-reads-program-planted-8 | 0.523 | 0 |
| 32-indexed-reads-program-planted-9 | 0.792 | 0 |
| 32-indexed-reads-program-planted-10 | 0.676 | 0 |
| 32-indexed-reads-program-planted-11 | 0.307 | 0 |
| 32-indexed-reads-program-planted-12 | 0.299 | 0 |
| 32-indexed-reads-program-planted-13 | 0.887 | 0 |
| 32-indexed-reads-program-planted-14 | 0.368 | 0 |
| 32-indexed-reads-program-planted-15 | 0.500 | 0 |
| 32-indexed-reads-program-planted-16 | 0.375 | 0 |
| 32-indexed-reads-program-planted-17 | 0.328 | 0 |
| 32-indexed-reads-program-planted-18 | 0.495 | 0 |
| 32-indexed-reads-program-planted-19 | 0.380 | 0 |
| 32-indexed-reads-program-planted-20 | 0.291 | 0 |
| 32-indexed-reads-program-planted-21 | 0.614 | 0 |
| 32-indexed-reads-program-planted-22 | 1.161 | 0 |
| 32-indexed-reads-program-planted-23 | 0.422 | 0 |
| 32-indexed-reads-program-planted-24 | 0.460 | 0 |
| 32-indexed-reads-program-planted-25 | 0.472 | 0 |
| 33-indexed-reads-emit-planted-0 | 0.719 | 1 |
| 33-indexed-reads-emit-planted-1 | 0.411 | 0 |
| 33-indexed-reads-emit-planted-2 | 0.362 | 0 |
| 33-indexed-reads-emit-planted-3 | 0.587 | 0 |
| 33-indexed-reads-emit-planted-4 | 0.306 | 0 |
| 33-indexed-reads-emit-planted-5 | 0.549 | 0 |
| 33-indexed-reads-emit-planted-6 | 0.397 | 0 |
| 33-indexed-reads-emit-planted-7 | 0.388 | 0 |
| 33-indexed-reads-emit-planted-8 | 0.721 | 0 |
| 33-indexed-reads-emit-planted-9 | 0.289 | 0 |
| 33-indexed-reads-emit-planted-10 | 0.397 | 0 |
| 33-indexed-reads-emit-planted-11 | 0.426 | 0 |
| 33-indexed-reads-emit-planted-12 | 0.268 | 0 |
| 33-indexed-reads-emit-planted-13 | 0.346 | 0 |
| 33-indexed-reads-emit-planted-14 | 0.291 | 0 |
| 33-indexed-reads-emit-planted-15 | 0.537 | 0 |
| 33-indexed-reads-emit-planted-16 | 0.401 | 0 |
| 33-indexed-reads-emit-planted-17 | 0.543 | 0 |
| 33-indexed-reads-emit-planted-18 | 0.413 | 0 |
| 33-indexed-reads-emit-planted-19 | 0.376 | 0 |
| 33-indexed-reads-emit-planted-20 | 0.364 | 0 |
| 33-indexed-reads-emit-planted-21 | 0.293 | 0 |
| 33-indexed-reads-emit-planted-22 | 0.536 | 0 |
| 33-indexed-reads-emit-planted-23 | 0.477 | 0 |
| 33-indexed-reads-emit-planted-24 | 0.356 | 0 |
| 33-indexed-reads-emit-planted-25 | 0.306 | 0 |
| 33-indexed-reads-emit-planted-26 | 0.474 | 0 |
| 33-indexed-reads-emit-planted-27 | 0.334 | 0 |
| 33-indexed-reads-emit-planted-28 | 0.372 | 0 |
| 33-indexed-reads-emit-planted-29 | 0.489 | 0 |
| apply-shard-0-mutant-False | 15.788 | 0 |
| apply-shard-0-mutant-True | 12.845 | 1 |
| apply-shard-1-mutant-False | 0.244 | 0 |
| apply-shard-1-mutant-True | 0.209 | 0 |

The apply table-writing mutant failed only Python shard 0/2. The emitted throw mutants failed only core.ts 0/17, checker.ts 0/1, binder.ts 0/26, and emitter.ts 0/30, through the existing emitted-JavaScript byte comparisons. Every other planted-run shard passed.

# Guard mutants

| Disabled check | Catcher |
| --- | --- |
| part-hash | test_payload_mutant_is_rejected: one assertion failure, zero errors |
| payload-hash | test_payload_manifest_mutant_is_rejected: one assertion failure, zero errors |
| manifest-key | test_manifest_mutant_is_rejected: one assertion failure, zero errors |
| lock-input | test_input_key_covers_all_inputs: one assertion failure, zero errors |
| node-input | test_input_key_covers_all_inputs: one assertion failure, zero errors |
| write-table | test_write_table_is_explicit: one assertion failure, zero errors |
| existing-output | test_existing_output_is_refused: one assertion failure, zero errors |
| parser-override | test_unkeyed_parser_override_is_refused: one assertion failure, zero errors |
| directory-symlink | test_input_key_covers_all_inputs: one assertion failure, zero errors |
| cache-miss-build | test_missing_product_never_builds: one assertion failure, zero errors |
| missing corpus case | actual-file shard union assertion |
| duplicate corpus case | actual-file unique ownership assertion |

Existing scripts also exercised their own mutants: regex capture non-null removal and alternate/optional capture behavior; empty or regex-injecting pragma arguments; diagnostic/baseline/oracle evidence filtering; missing constructor fields; missing debugger pause; interface anchor deletion/drift; wrong host-error message field; changed memoize stdout byte; wrong truthful-cast keys and the old sameMap signature. These remain part of the measured scripts.

# Exact commands and logs

Every measured command and elapsed result follows in JSON. ADAMIC_TEST_SHARD is i/corpus-count for indexed-read file runs and planted runs, and i/2 for Python shard runs; ordinary runs have it unset. NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules was used for proof runs. STAGE3_APPLY_MUTANT=1 was set only for Python mutant runs. Outputs were redirected to individual files, never piped.

```json
[
  {
    "unit": "apply-ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean",
    "seconds": 14.159344985999951,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean"
    ]
  },
  {
    "unit": "apply-ProductTests.test_existing_output_is_refused",
    "seconds": 0.14983828499998708,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_existing_output_is_refused"
    ]
  },
  {
    "unit": "apply-ProductTests.test_input_key_covers_all_inputs",
    "seconds": 0.22381982999991124,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_input_key_covers_all_inputs"
    ]
  },
  {
    "unit": "apply-ProductTests.test_manifest_mutant_is_rejected",
    "seconds": 0.14451953699995101,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_manifest_mutant_is_rejected"
    ]
  },
  {
    "unit": "apply-ProductTests.test_missing_product_never_builds",
    "seconds": 0.1349269209999875,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_missing_product_never_builds"
    ]
  },
  {
    "unit": "apply-ProductTests.test_payload_manifest_mutant_is_rejected",
    "seconds": 0.13525874800006932,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_payload_manifest_mutant_is_rejected"
    ]
  },
  {
    "unit": "apply-ProductTests.test_payload_mutant_is_rejected",
    "seconds": 0.13458579799998915,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_payload_mutant_is_rejected"
    ]
  },
  {
    "unit": "apply-ProductTests.test_restore_fetches_verified_product",
    "seconds": 0.153377644000102,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_restore_fetches_verified_product"
    ]
  },
  {
    "unit": "apply-ProductTests.test_shards_partition_every_case",
    "seconds": 0.13036705300009999,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_shards_partition_every_case"
    ]
  },
  {
    "unit": "apply-ProductTests.test_unkeyed_parser_override_is_refused",
    "seconds": 0.13617711699998836,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_unkeyed_parser_override_is_refused"
    ]
  },
  {
    "unit": "apply-ProductTests.test_write_table_is_explicit",
    "seconds": 0.14585805299998356,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProductTests.test_write_table_is_explicit"
    ]
  },
  {
    "unit": "apply-ProofShardTests.test_every_indexed_read_file_has_one_shard",
    "seconds": 0.22328986999991685,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py",
      "ProofShardTests.test_every_indexed_read_file_has_one_shard"
    ]
  },
  {
    "unit": "30-indexed-reads-whole",
    "seconds": 3.550577224000108,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-core.ts",
    "seconds": 0.5516375929998958,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-debug.ts",
    "seconds": 0.5171772050000527,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_baseNodeFactory.ts",
    "seconds": 0.25334619499994915,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_emitHelpers.ts",
    "seconds": 0.38914407499987647,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_emitNode.ts",
    "seconds": 0.33427058200004467,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_nodeChildren.ts",
    "seconds": 0.2759691789999579,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_nodeConverters.ts",
    "seconds": 0.2963375689998884,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_nodeFactory.ts",
    "seconds": 0.8943904919999568,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_nodeTests.ts",
    "seconds": 0.35655870899995534,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_parenthesizerRules.ts",
    "seconds": 0.3276677159999508,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_utilities.ts",
    "seconds": 0.45811926800001856,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-factory_utilitiesPublic.ts",
    "seconds": 0.2820836360001522,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-parser.ts",
    "seconds": 1.4389851470000394,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-path.ts",
    "seconds": 0.41553380799996376,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-scanner.ts",
    "seconds": 0.8850118729999394,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-utilities.ts",
    "seconds": 1.462577913999894,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-utilitiesPublic.ts",
    "seconds": 0.5924555310000414,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-0",
    "seconds": 0.4752443879999646,
    "exit": 1,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-1",
    "seconds": 0.49749347400006627,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-2",
    "seconds": 0.28144558400003916,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-3",
    "seconds": 0.37930582200010576,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-4",
    "seconds": 0.3824655159999111,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-5",
    "seconds": 0.2862973610001518,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-6",
    "seconds": 0.3174572249999983,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-7",
    "seconds": 0.8805650120000337,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-8",
    "seconds": 0.36779063900007714,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-9",
    "seconds": 0.33456506800007446,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-10",
    "seconds": 0.5009592189999239,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-11",
    "seconds": 0.2628261089998887,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-12",
    "seconds": 1.438344661000201,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-13",
    "seconds": 0.4130298090001361,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-14",
    "seconds": 0.8567201840000962,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-15",
    "seconds": 1.4316471079998792,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "30-indexed-reads-planted-16",
    "seconds": 0.5366423489999761,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/30-indexed-reads/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-before",
      "/tmp/stage3-apply-proof-inputs/30-indexed-reads-after"
    ]
  },
  {
    "unit": "31-indexed-reads-checker-whole",
    "seconds": 3.9900167549999423,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/31-indexed-reads-checker/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-before",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-after"
    ]
  },
  {
    "unit": "31-indexed-reads-checker-checker.ts",
    "seconds": 3.827818331000117,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/31-indexed-reads-checker/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-before",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-after"
    ]
  },
  {
    "unit": "31-indexed-reads-checker-planted-0",
    "seconds": 3.330778918000078,
    "exit": 1,
    "command": [
      "node",
      "stage3/adapt/31-indexed-reads-checker/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-before",
      "/tmp/stage3-apply-proof-inputs/31-indexed-reads-checker-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-whole",
    "seconds": 3.517332398999997,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-binder.ts",
    "seconds": 0.739222805000054,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-builder.ts",
    "seconds": 0.5858245750000606,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-builderPublic.ts",
    "seconds": 0.33645383700013554,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-builderState.ts",
    "seconds": 0.4169612660000439,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-builderStatePublic.ts",
    "seconds": 0.2531431289999091,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-commandLineParser.ts",
    "seconds": 0.6865034750001087,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-corePublic.ts",
    "seconds": 0.2678935609999371,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-executeCommandLine.ts",
    "seconds": 0.4231205910000426,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-expressionToTypeNode.ts",
    "seconds": 0.49754725200000394,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-moduleNameResolver.ts",
    "seconds": 0.7121189320000667,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-moduleSpecifiers.ts",
    "seconds": 0.6583031350000965,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-performance.ts",
    "seconds": 0.3101358410001467,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-performanceCore.ts",
    "seconds": 0.2752695950000543,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-program.ts",
    "seconds": 0.8538308559998313,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-programDiagnostics.ts",
    "seconds": 0.3691857839999102,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-resolutionCache.ts",
    "seconds": 0.45230074799997055,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-semver.ts",
    "seconds": 0.3610596709997935,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-symbolWalker.ts",
    "seconds": 0.3257725699997991,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-sys.ts",
    "seconds": 0.4747992019999856,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-tracing.ts",
    "seconds": 0.3624515280000651,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-tsbuild.ts",
    "seconds": 0.3077203479999753,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-tsbuildPublic.ts",
    "seconds": 0.6193127980000099,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-types.ts",
    "seconds": 1.1215819149999788,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-watch.ts",
    "seconds": 0.4289856979999058,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-watchPublic.ts",
    "seconds": 0.4076426409999385,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-watchUtilities.ts",
    "seconds": 0.5004776040000252,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-0",
    "seconds": 0.5682275870001376,
    "exit": 1,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-1",
    "seconds": 0.6112961270000596,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-2",
    "seconds": 0.2983825759999945,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-3",
    "seconds": 0.3716968440000983,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-4",
    "seconds": 0.24423163000005843,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-5",
    "seconds": 0.693727514000102,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-6",
    "seconds": 0.2722290830001839,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-7",
    "seconds": 0.477666175999957,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-8",
    "seconds": 0.5229315499998393,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-9",
    "seconds": 0.7916298469999674,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-10",
    "seconds": 0.6759294160001446,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-11",
    "seconds": 0.30663176900020517,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-12",
    "seconds": 0.29947111100000257,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-13",
    "seconds": 0.8874798020001435,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-14",
    "seconds": 0.36760966800011374,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-15",
    "seconds": 0.4998534549999931,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-16",
    "seconds": 0.3747730219999994,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-17",
    "seconds": 0.327518788000134,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-18",
    "seconds": 0.4948161329998584,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-19",
    "seconds": 0.38024407400007476,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-20",
    "seconds": 0.29128068099998927,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-21",
    "seconds": 0.613991679000037,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-22",
    "seconds": 1.1610984370001916,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-23",
    "seconds": 0.42211550399997577,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-24",
    "seconds": 0.4597397640000054,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "32-indexed-reads-program-planted-25",
    "seconds": 0.4716258549999566,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/32-indexed-reads-program/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-before",
      "/tmp/stage3-apply-proof-inputs/32-indexed-reads-program-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-whole",
    "seconds": 2.749351191000187,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-emitter.ts",
    "seconds": 0.8592183769999338,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-sourcemap.ts",
    "seconds": 0.40269289100001515,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformer.ts",
    "seconds": 0.3859975000000304,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_classFields.ts",
    "seconds": 0.5521687210000437,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_classThis.ts",
    "seconds": 0.2771701479998683,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_declarations.ts",
    "seconds": 0.45881337900004837,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_declarations_diagnostics.ts",
    "seconds": 0.4052305459999843,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_destructuring.ts",
    "seconds": 0.3622398840000187,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2015.ts",
    "seconds": 0.7054621009999664,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2016.ts",
    "seconds": 0.28715746199986825,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2017.ts",
    "seconds": 0.3925606669999979,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2018.ts",
    "seconds": 0.4566059060000498,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2019.ts",
    "seconds": 0.2741249739999603,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2020.ts",
    "seconds": 0.34475500899998224,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_es2021.ts",
    "seconds": 0.27743176900003164,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_esDecorators.ts",
    "seconds": 0.5151095270000496,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_esnext.ts",
    "seconds": 0.3792594670001108,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_generators.ts",
    "seconds": 0.5457125719999567,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_jsx.ts",
    "seconds": 0.41342434299986053,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_legacyDecorators.ts",
    "seconds": 0.37734850199990433,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_module_esnextAnd2015.ts",
    "seconds": 0.35651075600003423,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_module_impliedNodeFormatDependent.ts",
    "seconds": 0.2704819430000498,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_module_module.ts",
    "seconds": 0.4784729499999685,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_module_system.ts",
    "seconds": 0.4750611449999269,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_namedEvaluation.ts",
    "seconds": 0.3378590989998429,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_taggedTemplate.ts",
    "seconds": 0.29797923499995704,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_ts.ts",
    "seconds": 0.5442535120000684,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_typeSerializer.ts",
    "seconds": 0.3633461909998914,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-transformers_utilities.ts",
    "seconds": 0.42374172700010604,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-visitorPublic.ts",
    "seconds": 0.436148459000151,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-0",
    "seconds": 0.7193182539999725,
    "exit": 1,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-1",
    "seconds": 0.4110429649999787,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-2",
    "seconds": 0.3620537530000547,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-3",
    "seconds": 0.5870613009999488,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-4",
    "seconds": 0.3064105770001788,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-5",
    "seconds": 0.5485499549999986,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-6",
    "seconds": 0.3965788140001223,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-7",
    "seconds": 0.3876418720001311,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-8",
    "seconds": 0.7214813760001562,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-9",
    "seconds": 0.28898573499986924,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-10",
    "seconds": 0.3972391330000846,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-11",
    "seconds": 0.42643248400008815,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-12",
    "seconds": 0.26826730499988116,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-13",
    "seconds": 0.34623847900002147,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-14",
    "seconds": 0.29100992400003634,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-15",
    "seconds": 0.537245435000159,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-16",
    "seconds": 0.400522920999947,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-17",
    "seconds": 0.543262451000146,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-18",
    "seconds": 0.41327099300019654,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-19",
    "seconds": 0.3755316149999999,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-20",
    "seconds": 0.36407014900009926,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-21",
    "seconds": 0.293444602000136,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-22",
    "seconds": 0.5358944479999082,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-23",
    "seconds": 0.476807565999934,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-24",
    "seconds": 0.3558212769999045,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-25",
    "seconds": 0.30609201399988706,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-26",
    "seconds": 0.47435874200004946,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-27",
    "seconds": 0.33438744499994755,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-28",
    "seconds": 0.3719734219998827,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "33-indexed-reads-emit-planted-29",
    "seconds": 0.48888081999984934,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/33-indexed-reads-emit/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-before",
      "/tmp/stage3-apply-proof-inputs/33-indexed-reads-emit-after"
    ]
  },
  {
    "unit": "45-regex-captures-test",
    "seconds": 8.326310945999921,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/45-regex-captures/test.cjs",
      "/tmp/stage3-apply-pristine"
    ]
  },
  {
    "unit": "46-fix-pragma-test",
    "seconds": 2.933828665999954,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/46-fix-pragma-empty-argument/test.cjs",
      "/tmp/stage3-apply-pristine"
    ]
  },
  {
    "unit": "45-regex-captures-evidence",
    "seconds": 0.03665250400013065,
    "exit": 0,
    "command": [
      "python3",
      "stage3/adapt/45-regex-captures/verify.py",
      "stage3/adapt/45-regex-captures/evidence/census-before.jsonl.gz",
      "stage3/adapt/45-regex-captures/evidence/census-after.jsonl.gz",
      "stage3/adapt/46-fix-pragma-empty-argument/evidence"
    ]
  },
  {
    "unit": "41-constructors",
    "seconds": 2.684397196999953,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/41-explicit-any-remaining/constructor-proof.cjs",
      "/tmp/stage3-apply-proof-inputs/41-explicit-any-remaining-before",
      "/tmp/stage3-apply-proof-inputs/41-explicit-any-remaining-after",
      "/tmp/stage3-apply-cold/constructors.json"
    ]
  },
  {
    "unit": "41-debugger",
    "seconds": 0.4070139600000857,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/41-explicit-any-remaining/debugger-proof.cjs",
      "/tmp/stage3-apply-proof-inputs/41-explicit-any-remaining-after"
    ]
  },
  {
    "unit": "41-guards",
    "seconds": 2.644323937000081,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/41-explicit-any-remaining/guard-proof.cjs",
      "/tmp/stage3-apply-proof-inputs/41-explicit-any-remaining-before",
      "/tmp/stage3-apply-proof-inputs/41-explicit-any-remaining-after",
      "/tmp/stage3-apply-cold/guards.json"
    ]
  },
  {
    "unit": "47-host-errors",
    "seconds": 3.060852463000174,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/47-host-errors/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/47-host-errors-before",
      "/tmp/stage3-apply-proof-inputs/47-host-errors-after",
      "/tmp/stage3-apply-cold/host.json"
    ]
  },
  {
    "unit": "48-memoize",
    "seconds": 1.3852298169999813,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/48-memoize/verify.cjs",
      "/tmp/stage3-apply-proof-inputs/48-memoize-after",
      "/tmp/stage3-apply-cold/memoize"
    ]
  },
  {
    "unit": "65-node-builtins",
    "seconds": 1.4193057680001857,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/65-temporary-node-builtins/check.cjs",
      "/tmp/stage3-apply-proof-inputs/65-temporary-node-builtins-after"
    ]
  },
  {
    "unit": "76-truthful-casts",
    "seconds": 3.4799602230000346,
    "exit": 0,
    "command": [
      "node",
      "stage3/adapt/76-truthful-casts/check.cjs",
      "/tmp/stage3-apply-proof-inputs/76-truthful-casts-after"
    ]
  },
  {
    "unit": "apply-shard-0-mutant-False",
    "seconds": 15.787772762000031,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py"
    ]
  },
  {
    "unit": "apply-shard-0-mutant-True",
    "seconds": 12.845225196000001,
    "exit": 1,
    "command": [
      "python3",
      "stage3/test_apply.py"
    ]
  },
  {
    "unit": "apply-shard-1-mutant-False",
    "seconds": 0.24422361300003104,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py"
    ]
  },
  {
    "unit": "apply-shard-1-mutant-True",
    "seconds": 0.20854928199992173,
    "exit": 0,
    "command": [
      "python3",
      "stage3/test_apply.py"
    ]
  }
]
```

Local evidence: /tmp/stage3-apply-cold/*.log and timings.json; /tmp/stage3-apply-baseline-cold.log; /tmp/stage3-apply-baseline-time.json; /tmp/stage3-apply-check-mutants/*.log and results.json; /tmp/stage3-apply-check-mutants/selectors.json. Initial Python shard import failures are retained as initial-failed-* and excluded from the valid measurements after correcting the loader and rerunning. Earlier preparation concurrent with setup was excluded. No whole Go package or gate was run. No new .a fixture was added.

# Outstanding work

Remote publishing belongs to developer tools under the explicit follow-up instruction. This worker neither uses cache credentials nor uploads products. Only local product retrieval and tiny file-URL retrieval have been tested; no remote hash product has been published by this worker.

Not measured or converted here: 20 baseline proof, 40 historical compiled class/filesystem/error-host proofs, 41 full artifact/oracle proof, 70/71/75 proof families, and ancillary census/API scripts. In particular, 41/oracle-proof.cjs invokes stage3/oracle/run.sh whose runner unconditionally installs/builds outside this territory; it has no supplied-product option. Their cold budget is not established by these measurements. This report does not certify the entire stage 3 tier. The follow-up requests pushing the fetch side and documented hook with these coverage gaps retained.

# Follow-up validation

Current fetch/hook/fallback tests, run individually in fresh processes:

| Unit | Seconds | Exit |
| --- | ---: | ---: |
| ProductTests.test_corrupt_product_never_builds | 0.202 | 0 |
| ProductTests.test_existing_output_is_refused | 0.186 | 0 |
| ProductTests.test_input_key_covers_all_inputs | 0.299 | 0 |
| ProductTests.test_manifest_mutant_is_rejected | 0.247 | 0 |
| ProductTests.test_missing_product_builds_and_reports | 0.158 | 0 |
| ProductTests.test_only_manifest_absence_is_a_cache_miss | 0.227 | 0 |
| ProductTests.test_payload_manifest_mutant_is_rejected | 0.193 | 0 |
| ProductTests.test_payload_mutant_is_rejected | 0.276 | 0 |
| ProductTests.test_publish_hook_describes_local_payloads | 0.395 | 0 |
| ProductTests.test_restore_fetches_verified_product | 0.173 | 0 |
| ProductTests.test_shards_partition_every_case | 0.138 | 0 |
| ProductTests.test_unkeyed_parser_override_is_refused | 0.159 | 0 |
| ProductTests.test_write_table_is_explicit | 0.151 | 0 |
| ProofShardTests.test_every_indexed_read_file_has_one_shard | 0.281 | 0 |

Commands: `python3 stage3/test_apply.py CASE`, output to `/tmp/stage3-followup-CASE.log`; results `/tmp/stage3-followup-tests.json`. The hook rejection assertion was subsequently strengthened to require the exact error class/message and its focused test rerun passed. Each isolated check mutant below produced one assertion failure and zero errors:

| Mutant | Catcher |
| --- | --- |
| part-hash | test_payload_mutant_is_rejected |
| payload-hash | test_payload_manifest_mutant_is_rejected |
| manifest-key | test_manifest_mutant_is_rejected |
| lock-input | test_input_key_covers_all_inputs |
| node-input | test_input_key_covers_all_inputs |
| write-table | test_write_table_is_explicit |
| existing-output | test_existing_output_is_refused |
| parser-override | test_unkeyed_parser_override_is_refused |
| directory-symlink | test_input_key_covers_all_inputs |
| cache-miss-fallback | test_missing_product_builds_and_reports |
| cache-miss-message | test_missing_product_builds_and_reports |
| corrupt-fallback | test_corrupt_product_never_builds |
| http-miss | test_only_manifest_absence_is_a_cache_miss |
| hook-key | test_publish_hook_describes_local_payloads |
| hook-empty | test_publish_hook_describes_local_payloads |
| hook-suffix | test_publish_hook_describes_local_payloads |
| hook-assets | test_publish_hook_describes_local_payloads |
| hook-part-hash | test_publish_hook_describes_local_payloads |

Mutant commands: `python3 /tmp/stage3-followup-mutants.py > /tmp/stage3-followup-mutants.log 2>&1`; per-mutant logs and results.json in `/tmp/stage3-followup-mutants/`. The script mutates isolated apply modules, not repository files. The first hook-key probe accidentally mutated the fetch guard and survived; the corrected probe targets only the hook guard and was killed.

Final product preparation exercises the ordinary local-store cache-miss path against the final input hash; its output and elapsed result are in `/tmp/stage3-followup-build-final.log` and `/tmp/stage3-followup-build-final.json`. This is preparation outside test units and is expected to exceed 30 seconds. Final ordinary apply and both clean/planted Python shards run after this product exists, with outputs and timings in `/tmp/stage3-followup-final-*.log` and `/tmp/stage3-followup-final.json`. The CLI hook output is `/tmp/stage3-followup-publish-hook.json`; developer tools can reproduce it with `bash stage3/apply.sh --publish-hook`. No product upload or credential use occurs.
