Built three scoped rule candidates with descriptors, oracle adapters, witnesses and mutants; temporary .ts modules follow Ahra's correction.
Commits: claim 6b3c1bdc; rules 602f008d, 281bea59, 7647ff47; transport f8e13adc; temporary extensions 51259134.
Checks: 179 fixture contracts and 214 source files per rule matched on Node, emitted JavaScript and ASan/UBSan native; registry passes.
Mutants: remove rtl/ltr exemption, remove description exemption, permit display=block; all compiled and ran, then failed only comparison on all three backends.
Uncovered: full independent parser parity and full gate; shared source harness rejects captured case-103.ts, and independent JSX parsing is unsupported. Stopped here, no further claims.

# Rules and ownership

- structure/tailwind-no-physical-direction
- @eslint-community/eslint-comments/require-description
- @next/next/google-font-display

Claims were pushed before implementation. Rule changes stay in the three owned directories. The latest correction explicitly permits .ts until integration codemods them to .a; no shared generator or test harness was edited. The earlier .a compatibility proposal remains historical evidence only; Ahra assigned that work to codex/lint-harness-dot-a. No PR was opened.

# Verified behavior

Captured 179 distinct original Go rule-test cases, retaining original filenames and options: tailwind 54, description 109, font 16. Every finding offset, ID, resolved message, empty repair and unchanged fixed source matched unmodified Go rules on all three backends using projected Go ASTs. The rules provide no Go fixes or suggestions. 161 fixtures also matched using independent Adamic parsing; 18 require JSX handling. The explicit JSX refusal control failed on all three backends. All three semantic mutants compiled and ran successfully with empty stderr and were caught solely by output comparison.

The TypeScript corpus is v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Compiler src/compiler plus stage1 sources total 214 files. Complete projected-tree comparisons passed on all three backends for each rule: tailwind 12,279,607 output bytes, description 12,328,594, font 12,277,467. An environment restart occurred after tailwind completed; final-before-restart.log and corpus-resumed.log show the separate runs. After changing the eight owned Adamic files to temporary .ts, every changed source was rechecked across all three rules and backends (29,673 / 27,479 / 27,295 output bytes). Unchanged source observations are retained from the complete corpus runs.

Final temporary-.ts verification: TestRuleContracts passed in 57.70s; changed-source TestSourceCorpora in 20.30s; package 78.006s. Ordinary registry tests passed in 0.020s, including deterministic generation and descriptor mutants. The pre-rename fixture tests also passed. Initial single-process corpus loading exceeded Node's default heap; bounded batches of ten files fixed that. Source transport uses UTF-8-safe 4096-byte chunks and output joins. Interrupted runs are preserved, not represented as whole-test passes.

# Exact remaining blockers

The unchanged shared TestRulesAgree captured 395 combinations but failed in Go before backend comparison: panic: invalid corpus .../case-103.ts: [Unterminated regular expression literal.]. lint_test.go upstream() writes every captured case as case-NNN.ts, discarding original TSX filenames. Its Go oracle also rejects parser diagnostics, whereas actual cohere rule tests include intentionally malformed source. This differs from the owned harness, which preserves filenames and Go-recovered fixture trees.

Independent Adamic JSX parsing remains unsupported, so projected trees certify rule logic, not the full frontend pipeline. The shared harness currently compares Node/native and has no emitted-JavaScript side; the owned harness proves that side separately. Full independent compiler/stage1 parity and the full repository gate are not certified. Ahra said to report other blockers and stop instead of editing shared files; this is the stopping point.

# Commands and setup

Setup bash cloud/setup.sh passed: Go ready 0s, clang ready 1s, Node ready 1s, submodules 1s, cache 17s, done 18s. nproc=5. Environment source /workspace/adamic-tools/env.sh.

    go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run TestRuleContracts -count=1 -v > contracts.log 2>&1
    go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run TestSourceCorpora -timeout 30m -count=1 -v > corpus.log 2>&1
    ADAMIC_WAVE1_RULES='@eslint-community/eslint-comments/require-description,@next/next/google-font-display' go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run TestSourceCorpora -timeout 30m -count=1 -v > corpus-resumed.log 2>&1
    ADAMIC_WAVE1_CORPUS=owned go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run 'TestRuleContracts|TestSourceCorpora' -timeout 30m -count=1 -v > ts-final.log 2>&1
    go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run TestFixtureRates -timeout 5m -count=1 -v > rates.log 2>&1
    go test ./stage1/cohere/lint/registry -count=1 -v > registry-ts.log 2>&1
    go test ./stage1/cohere/lint -run '^TestRulesAgree$' -timeout 10m -count=1 -v > shared-source-gate.log 2>&1

Full corpus reproduction requires cloning pinned TypeScript into /tmp/lint-wave1-typescript. The final command fails at the shared blocker above. No edits to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or internal/oracle/oracle_test.go.

# Observed findings per second

One process per rule over its original upstream fixtures, includes startup/input loading; sanitized native. Go parses source, Node/native use projected trees. These observations are not equivalent end-to-end parser benchmarks. Measured before the temporary filename change, concurrently with corpus verification. Counts match on all sides.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| tailwind-no-physical-direction | 33 | 447.526 | 298.695 | 3802.580 |
| require-description | 55 | 596.884 | 382.022 | 6149.189 |
| google-font-display | 7 | 177.444 | 68.835 | 867.454 |

TestFixtureRates passed in 15.830s. Original tailwind corpus count pass: zero findings, so zero findings/s; Go elapsed 0.474804s, Node 9.764146s, native 117.493010s. This is not a speed ranking.

Historical registry.log records the earlier .a-only generator failure. registry-overlay.log proves the proposed temporary overlay passed, without modifying shared files. registry-ts.log is the current ordinary passing registry test. compatibility.patch was never applied to the repository.
