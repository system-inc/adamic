Built: finished the six prior claims, then ported no-global-assign, no-implicit-globals and no-implied-eval; nine rules total on this branch.
Commits: six completed in 63298efb and 590f4fa6; next-three claim 89222643; final implementation is the commit containing this report.
Commands and outputs: final agreement PASS 125.335s; bridge PASS 154.802s; checker PASS 0.602s; vet and formatting PASS.
Mutants: three new native rule mutants caught only by full byte comparison; both checker guard mutants, retained registry and seven bridge mutants caught.
Not covered: full repository gate, nondefault options, CLI lint for .a, and JavaScript execution of these FFI rules; no requested rule remains blocked.

## Selection and completion order

The premature continuation's constructor and metadata-test blockers are resolved and reported in WAVE_22_CONTINUATION_REPORT.md. The six existing claims were completed and pushed in 590f4fa6 before selecting another rule.

After fetching every origin head again, the combined descending-volume ranking (lexical ties) was checked against base/main ports and all origin claim documents. The first three available names were no-global-assign, no-implicit-globals and no-implied-eval, each with zero recorded compiler and repository findings. Claim 89222643 was pushed before implementation. validation-wave-22-third/selection.json records fetched ref SHAs, existing ports, claims and the remaining ordering. None of these three was ported on base/main or claimed on a fetched origin head. No PR was opened and no further rules were claimed.

Each rule is in its own .a file. The two new questions, global-binding and global-source, each have dedicated Go and Adamic files and direct compiler tests. Their only shared edits are one-line dispatch registrations. The shared registration generator, existing test harness, compiler emitter and lowerer are untouched. The dedicated wave_22_third_suite.a and wave_22_third_test.go compare the new rules without requiring shared harness changes. wave_22_global_questions.md documents the protocols.

Go returns symbol declaration origins, local merged declarations, shorthand value origins, module classification, JavaScript classification and resolved AlwaysStrict. Adamic decides every write, global/shadow predicate, string argument, directive, ancestor and report. Unrelated call names are filtered before provenance queries; compiler-proven strict sources skip impossible leaks. These optimizations are held by the final byte comparisons.

The Go authority intentionally differs between global policies: no-global-assign checks the first declaration-file origin, while no-implied-eval requires a symbol and declines any source declaration, including a longer local merged declaration list. Its unresolved names are declined and its globalThis symbol with zero declarations is accepted. Native carries those exact decisions. For no-implicit-globals, top-level nonlexical declarations report even in a strict script, but leaks require sloppy code. Native retains the production default lexicalBindings false.

## Exact agreement

Cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and TypeScript source at 050880ce59e30b356b686bd3144efe24f875ebc8. The independent oracle calls unchanged production registry rules, importing no bridge. Both implementations use default options and the same roots/configuration.

| Population | Roots or executions | Findings | Identical bytes, Go/native/sanitizers |
| --- | ---: | ---: | ---: |
| Extracted and targeted paired controls | 620 | 183 | 92,857 |
| Isolated positive/strictness profiles | 8 | 29 | 9,394 aggregate |
| Frozen repository | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,857 |

The 310 source texts are extracted from production test strings plus targeted cases, each run as JavaScript and Adamic input. No source is rejected by the independent parser. The test obtains the production implied-eval ambient-global declaration fixture independently; it is data, not an Adamic implementation. Paired sources distinguish sloppy JavaScript leaks from TypeScript AlwaysStrict while preserving script/module detection.

Each rule has positive controls. Additional one-file programs prevent merged declarations in other script fixtures from hiding a global: destructuring global writes, for-await targets, execScript and nested global receivers, local timer shadows, strict function/class bodies and using declarations all compare separately. The using case confirms that its node flags are not treated as let/const by this production default rule. Exact spans and Unicode/CRLF text are compared.

Full streams include file headers, rule/message IDs, messages, byte spans, fix/suggestion counts and all repairs. These three production Go rules offer no fixes or suggestions; their zero repair fields are still compared rather than omitted. Normal and ASan/UBSan/LSan runs have empty native stderr. Zero corpus findings alone are not the evidence: positive extracted and isolated controls are required by the runner.

## Mutants and checker guards

Every native rule mutant builds, exits 0 and has empty stderr. Only diagnostic byte comparison catches it:

| Mutant | First differing byte |
| --- | ---: |
| Reverse first-declaration-file provenance | 5,478 |
| Report a global declarator on its name only | 74,522 |
| Treat minus rather than plus as string concatenation | 56,758 |

Scratch Go overlays remove the binding mode guard and the source suffix guard. TestGlobalBinding fails by accepting malformed global-binding, and TestGlobalSource fails by accepting an extra suffix in all four source classifications. Neither is a compilation failure. Baseline tests compare all origin fields with direct compiler symbols, including shorthand value symbols and unresolved names. Source tests cover scripts, modules, explicit strict false and JavaScript classification.

Both new question modes queried after release panic with exit 70 and the exact invalid-or-released-checker-handle message. A registry mutant keeping the released handle live exits 0 through global-binding, caught by the required refusal. The full bridge package tree additionally checks 100 C ownership queries, retained output after program release, zero/stale handles, and 1,600 compiler positions across four files, with 54,982 identical bytes under sanitizers. All seven existing input/output length, registry, wrong-position, link and leak mutants are caught.

The filtered Node gate ran earlier in this same continuation and passed in 63.365s, including eight fixtures through native and JavaScript and the one-byte oracle mutant. No compiler code changed afterward. That is compiler/ownership evidence; it is not a claim that these FFI rule programs execute through JavaScript.

## Native time against Go

Three quiet alternating full-output runs after validation, with byte comparison on every timed run. Timings include startup, load, native parse, lint traversal, serialization and teardown; builds and sanitizers are excluded.

| Population | Native median | Go median | Native / Go | Native queries |
| --- | ---: | ---: | ---: | ---: |
| Repository | 0.521294s | 0.215915s | 2.414 | 1,424 |
| Compiler | 3.619700s | 0.659208s | 5.491 | 6,198 |

Native is slower on both workloads. Exact runs, counters and output hashes are in validation-wave-22-third/timings.json. The previous continuation's medians, 1.134065/0.354046s for repository and 13.328035/2.448317s for compiler, are separately reported in WAVE_22_CONTINUATION_REPORT.md. These are observed timings, not speed guarantees.

## Commands and evidence

Source /workspace/adamic-tools/env.sh first. Test output went directly to logs, never through a pipe.

```sh
ADAMIC_WAVE22_THIRD_ARTIFACTS=/workspace/wave-22-third-complete ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22ThirdAgreement$' -count=1 -timeout=30m -v > /workspace/wave-22-third-complete.log 2>&1
TMPDIR=/workspace/wave-22-scratch ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-22-third-bridge.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-22-third-checker-final.log 2>&1
go test -overlay /workspace/wave-22-third-binding-mode-overlay.json ./bridge/tsgo/checker -run '^TestGlobalBinding$' -count=1 -v > /workspace/wave-22-third-binding-mode-mutant.log 2>&1
go test -overlay /workspace/wave-22-third-source-suffix-overlay.json ./bridge/tsgo/checker -run '^TestGlobalSource$' -count=1 -v > /workspace/wave-22-third-source-suffix-mutant.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /workspace/wave-22-third-vet-final.log 2>&1
python3 stage1/cohere/typeaware/validation-wave-22-third/benchmark.py /workspace/wave-22-third-complete /workspace/adamic /workspace/wave-22-typescript-pinned /workspace/wave-22-third-benchmark > /workspace/wave-22-third-benchmark.log 2>&1
```

Native rule and released-registry mutants reproduce in TestWave22ThirdAgreement. Guard overlays are generated by replacing only the relevant question-validation expression in a scratch copy; their exact failing assertions are in the committed logs. Evidence includes full gzip streams and hashes, accepted source texts, isolated profiles, portable manifests, ambient fixture text as JSON, control configuration, timing counters and logs.

Initial setup reported go/clang/node/submodules ready 0s each, cache warm 118s, total 118s and nproc 5 with four-core quota. It was not repeated. The full repository gate and every upstream option/configuration permutation were not run. The base CLI's .a lint limitation and FFI JavaScript execution limit remain explicit; no shared-file workaround was introduced. All nine claimed rules are now ported and validated on the requested frozen corpora under their production defaults.
