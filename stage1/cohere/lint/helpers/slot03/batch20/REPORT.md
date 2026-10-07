Built tryStatement, IsDocumentFile and lastSeparator in separate .a files; sixty owned helpers total.
Commits: claim 52dbeaf0 pushed before source; final implementation publication SHA is named in the final response; current lint area d3a37422 contains current main b6b1538b.
Commands: new helper gate PASS 139.993s, 4,587 Go/source Node/emitted JavaScript/sanitized native comparisons; vet/format clean; six uncached Node probes PASS 1.624s; setup 140s, nproc 5.
Mutants: all twenty-four new compiling semantic variants caught by exact Go output comparison, plus the independent missing-consumer check; every replacement and witness is recorded below and in evidence/helpers.log.
Not covered: complete rule findings/fixes/suggestions, independent graph/recursive/splitPath backends, arbitrary malformed ASTs or AST-mutating callbacks, and the full repository gate including its seventeen external comparisons.

# Landing and ownership

Before reserving helpers, all fifty-seven previous helpers were complete, oracle-green and pushed at e9d694c3. That full current-base gate passed in 1,264.599s with 3,054,680 comparisons, all 230 owned variants, four inherited variants and the missing-consumer check, with shared helpers passing in 91.887s and six selected shared-harness checks passing in 290.864s. Main b6b1538b and lint area d3a37422 remain unchanged and are ancestors of this branch. Prior source is unchanged, so this unit uses the bounded touched-helper package and filtered external oracle gate. Only codex/lint-helpers-03 is this worker's pushed branch.

Wildcard fetch checked all twenty origin codex/lint-helpers* branches and nineteen claim files before selection. tryStatement has four consumer rules and is the highest eligible unclaimed helper; each selected Next.js helper ties the next highest eligible count at three. Comments remain reserved by their shared bundle; regexp-engine internals are excluded under the JS RegExp instruction. Original reservation 52dbeaf0 was pushed before writing implementation files. The refreshed scan still finds only this worker's reservation for each symbol. ownership.json preserves initial claim bodies; ownership-refresh.json preserves all refreshed branch tips and claim bodies. No fourth helper is reserved.

Territory: claims/03.md, slot03/batch20 and batch20_test.go. Shared harness, registry, compiler, runtime and actual cohere worktree files are unchanged. New Adamic source files are .a. Temporary Go overlays rename methods and execute their unchanged bodies, and add observation-only export/control files. No new rule dispatch or regex matching is introduced.

# Rules unblocked

| Helper | Consumers |
|---|---|
| tryStatement | array-callback-return; consistent-return; no-unreachable-loop; react-hooks/rules-of-hooks |
| IsDocumentFile | @next/next/no-head-import-in-document; @next/next/no-page-custom-font; @next/next/no-styled-jsx-in-document |
| lastSeparator | @next/next/no-head-import-in-document; @next/next/no-page-custom-font; @next/next/no-styled-jsx-in-document |

Ten dependency edges across seven rules. Subtracting only this batch from the frozen readiness inventory removes zero final helper blockers. Remaining dependencies are listed in readiness.json. These counts describe helper readiness and do not assert complete rule findings parity. In particular, splitPath remains an explicit backend dependency, not an implementation delivered by this unit.

# Actual Go observations

The pinned cohere oracle is 715ba94f3608a6500086b1076ce5cb7e51b836db. Fixture capture executes the upstream core, next and react packages and exits zero, with 2,207 deduplicated runtime sources and an asserted exact seven-consumer set. Typed fixture strings are captured before external-engine availability checks. This does not assert the upstream external engines or whole-rule findings were exercised by helper observation.

The actual Go parser and CFG builder reach the unchanged tryStatement helper over all four consumer corpora. Direct dependency wrappers execute original Go backends, recording kind, arguments, return value, full frame fields, stack identities, current block/reachability and every block's incoming flag before and after each operation. Nested dependency work is suppressed from an enclosing trace. A per-node control walk exposes nested try statements independently and supplies three valid enclosing-frame configurations, including catch position and a catch without finally. Arena records remain observable after a stack pop, so the frame's flags and implicit predecessors still decide the abrupt finally copy. Native/source/emitted ports independently produce their call sequence and frame mutations while replaying dependency results.

Observed real consumer try calls: array-callback-return 8, consistent-return 3, no-unreachable-loop 56, react-hooks/rules-of-hooks 13; total 80. Seventy-one controls yield 151 CFG comparisons. The zero Next.js CFG counts are a lack of observed try helper reach in those fixture sources, not missing filename coverage.

Both raw fixture filenames and their canonical rooted shape are evaluated by actual Go IsDocumentFile, splitPath and lastSeparator. This provides 4,414 corpus path evaluations, including all 34 no-head-import, 35 no-page-custom-font and 19 no-styled-jsx fixture sources. Twenty-two additional paths cover empty/bare/dotted document names, permissive directory prefixes, case sensitivity, separators at either end, mixed separators, multibyte characters, invalid UTF-8 bytes and NUL bytes. There are 4,436 path output rows, each comparing both owned helpers. Raw-byte input preserves the exact Go lastSeparator offset even when JSON replaces an invalid UTF-8 string byte; its ASCII document-prefix answer remains unchanged. No finding-position conversion is performed in a byte-offset helper.

Total 4,587 output rows agree across real Go, source Node, generated JavaScript and sanitized native. This is composition evidence. It does not implement or independently prove recursive statements/binding, graph storage/reachability, fork snapshot/restore, handler lookup/targets or splitPath.

# Commands and findings

Source /workspace/adamic-tools/env.sh first. Every test writes stdout/stderr directly to a log file. Native baselines and variants run with ASan/UBSan. A variant must compile, exit zero and emit empty stderr before a stdout mismatch counts. A compilation refusal, panic or sanitizer failure is not counted.

```
bash cloud/setup.sh > /tmp/slot03-batch20-setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch20/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch20/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch20' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch20/evidence/helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch20/evidence/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch20/evidence/input-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestConsumerCoverageRejectsMutant$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch20/evidence/coverage-mutant.log 2>&1
```

Setup timing lines: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s, build cache warm 140s, done 140s on five processors; cgroup cpu.max 400000 100000, 17.6 GB. nproc is 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Repository vet and Go format logs are empty. All six uncached external Node probes pass in 1.624s, zero probe hits and six misses. The coverage check catches removal of better-tailwindcss/no-unknown-classes from an expected consumer set.

The initial 2,335-row baseline passes, but two outward-frame flag variants survive; that failed gate is preserved in initial-mutant-survivors.log (90.581s). The observer only reached outermost tries and fresh per-node builders had no enclosing frames. Actual-Go enclosing-frame controls are added, and both variants are retained and caught.

The expanded 4,587-row baseline passes, but one existing variant replacing implicit predecessors with normal ends produces an out-of-order dependency sequence. The first replay borrowed an unrelated callback return value and panicked with invalid throw stack (exit 70); expanded-mutant-replay-failure.log preserves that failed gate (149.199s). It is not a semantic catch. The replay now supplies the missing-value sentinel for an out-of-order callback; its independently generated kind/arguments/pre/post trace still fails exact comparison. This change is only in the owned driver, and neither accepts a mismatch nor changes Go helper logic. Every variant is retained, rerun and required to finish cleanly.

Final new gate passes 139.993s with all twenty-four compiling semantic variants caught and no skipped selected tests. Witnesses show the first changed output line and byte, with a bounded context window so a late frame mutation is visible. Source Node and generated JavaScript are baseline comparisons; semantic variants execute on sanitized native.

# Every mutant

| Variant | Exact replacement |
|---|---|
| try_statement.a | `if (view.hasCatch) {` -> `if (false) {` |
| try_statement.a#01 | `if (view.bindingPresent)` -> `if (false)` |
| try_statement.a#02 | `frame.position = 1` -> `frame.position = 0` |
| try_statement.a#03 | `frame.thrownForked = false` -> `frame.thrownForked = true` |
| try_statement.a#04 | `state.stack.pop();` -> `(omit)` |
| try_statement.a#05 | `if (!view.hasFinally)` -> `if (false)` |
| try_statement.a#06 | `for (const source of frame.implicit)` -> `for (const source of normalEnds)` |
| try_statement.a#07 | `dependencies.restoreForks(snapshot);` -> `(omit)` |
| try_statement.a#08 | `if (state.incoming.includes(frame.finallyEntry))` -> `if (false)` |
| try_statement.a#09 | `if (state.reachable)` -> `if (true)` |
| try_statement.a#10 | `if (frame.returnedAny)` -> `if (false)` |
| try_statement.a#11 | `if (frame.thrownAny)` -> `if (false)` |
| try_statement.a#12 | `outer.returnedAny = true;` -> `(omit)` |
| try_statement.a#13 | `outer.thrownAny = true;` -> `(omit)` |
| try_statement.a#14 | `dependencies.markFinal(state.cur);` -> `(omit)` |
| try_statement.a#15 | `dependencies.markThrown(state.cur);` -> `(omit)` |
| try_statement.a#16 | `dependencies.enter(after); / }` -> ` / }` |
| is_document_file.a | `startsWith('_document.')` -> `startsWith('_document')` |
| is_document_file.a#01 | `parts.baseName.startsWith('index') && parts.parentName.startsWith('_document')` -> `parts.baseName === 'index' && parts.parentName === '_document'` |
| is_document_file.a#02 | `&& parts.parentName.startsWith('_document')` -> `(omit)` |
| last_separator.a | `if (byte === 92)` -> `if (false)` |
| last_separator.a#01 | `if (byte === 47)` -> `if (false)` |
| last_separator.a#02 | `if (windows > posix)` -> `if (windows < posix)` |
| last_separator.a#03 | `posix = i;` -> `posix = 0;` |

# Limits

Callbacks must preserve immutable parsed metadata and valid backend state. Wrong-kind or malformed AST assertions, AST mutation by callbacks and arbitrary byte values outside a Go string byte sequence are outside the successful-input contract. There are no complete rule findings/fixes/suggestions comparisons or integration into a complete native CFG builder. Backends and splitPath remain dependencies as documented in README.md. No rule or adapter options are dropped by this unit.

The full repository gate, including the seventeen required external stage 1 comparisons, was not run. No selected correctness check was skipped, relaxed or deleted. The earlier fifty-seven-helper gate is not represented as rerun by this new-helper-only gate; it remains green on the unchanged integration bases. Final publication is to the own branch only, with no push to main or an area branch and no PR.
