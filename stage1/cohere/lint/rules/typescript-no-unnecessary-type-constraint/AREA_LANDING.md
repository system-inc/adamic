Rules: retained seven winning/unique ports, retired eleven losing copies, connected four ports to unified repairs.
Commits: rebased from rule tip 492c18b04 onto area/stage1-lint 7481e0324; evidence commit follows this report.
Validation: fresh unified upstream comparison, selected corpus, single-edit controls, six semantic mutants, registry, vet and external one-byte oracle; raw logs beside this report.
Mutants: six retained-rule comparison-only mutants freshly proved; prefer-as-const annotation mutant remains blocked by the shared two-edit refusal.
Limits: prefer-as-const multi-edit findings and two malformed computed-key parser cases block full parity; no new claims, batch ports or helper work.

# Area harness integration

Fetched all origin heads before working. Read the whole DEDUP_LEDGER.md at the shared harness and area base; the ledger content is identical (SHA256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f). Rebased only the owned rule commits onto origin/area/stage1-lint 7481e0324e34a2537aafa9db7eeacda50405611b. That base contains current origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad and harness 41eb6eab2. Pre-evidence implementation tip is 4219e9d2cc7573a81aec6e5a2776a1d7faacc909. No shared-file conflict resolution or shared harness, registry, oracle, compiler or lowerer edit was made.

Retained rules: no-useless-computed-key, react-hooks/gating, react/forbid-foreign-prop-types, @typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const, @typescript-eslint/prefer-enum-initializers and default-param-last. The three TypeScript rules are ledger winners, default-param-last has this unique registry copy, and the remaining three are uncontested. All inherited area descriptors remain intact.

Retired losing directories and ledger winners:

| Dropped copy | Winner |
| --- | --- |
| @next/next/no-assign-module-variable | wave1-15 |
| @typescript-eslint/default-param-last | wave1-15 |
| structure/tailwind-no-physical-direction | wave1-05 |
| default-case-last, for-direction, no-constructor-return, no-delete-var, no-eq-null | wave1-13 |
| no-multi-str, no-nonoctal-decimal-escape, no-octal | wave1-15 |

Prior parking evidence was moved from the retired no-multi-str directory to landing_history/ here. These are historical logs and reports, not current parity claims. Its historical reproduction script is retained as parking.py.txt, rather than an executable with stale paths. The ledger lists 35 batch-only source copies but gives no wave1-08 assignment; none is taken while the landing cap remains closed.

# Unified repairs

The old owned comparison sidecar was not read by the integrated driver. repairs.a now builds the shared SuggestionEdit and Suggestion objects and calls context.reportNode. The constraint, enum, const and computed-key ports call that adapter. The shared seven-argument reporting API and driver are unchanged. Single-edit repairs and multi-edit suggestions now reach the real comparison wire. Both independent annotation fixes in prefer-as-const are preserved; the existing shared single-edit guard explicitly refuses them. Combining them into a different automatic edit would violate the Go byte comparison.

Descriptors already declare kind names validated against ast.Kind. No new rule or dispatch loop is added. Existing Go manual-scanner logic stays as ported; the regex translation table has no Go-regexp entry for these eighteen original rule names, and existing JS literals are unchanged. All Adamic implementation files remain .a.

# Fresh commands

Run from this rule worktree, with source /workspace/adamic-tools/env.sh:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=20m > /tmp/wave08-unified-witness.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-unified-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-unified-supported.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave08-unified-registry.log 2>&1
go vet ./... > /tmp/wave08-unified-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-unified-oracle.log 2>&1
```

The overlay adds only an owned test file to the package: its source is unified_test.go.txt in this directory, mapped to the absolute virtual lint/wave08_unified_test.go. It never replaces the comparator, oracle, driver or shared files. Regenerate the overlay mapping using the worktree's absolute paths; the TypeScript corpus must match the shared compilerCommit pin. The test uses the unmodified shared upstream, manifest, buildPort and compare functions. All three port backends are checked against the actual Go oracle, including emitted JavaScript and ASan/UBSan native.

Six semantic mutants PASS in 186.067s, each compiles and executes normally on Node, emitted JavaScript and sanitized native and is killed only by differing Go output. The mutants suppress default-param-last, corrupt computed-key replacement, suppress forbidden foreign propTypes, suppress invalid gating, corrupt constraint suggestion text and corrupt the second enum suggestion. prefer-as-const const_append_wrong was comparison-killed on all three backends in historical rich-transport evidence; it cannot be freshly proved through this harness because its witness requires two automatic edits. A baseline refusal is not counted as a mutant kill.

Supported upstream comparison PASS in 129.69s: 414 cases (computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21 and core default-param-last 128), all four outputs identical at 149648 bytes. Compiler/stage1 comparison PASS in 207.78s: all 364 .ts/.a files selected for the six rules, 2184 rule cases, all four outputs identical at 80203185 bytes. Three prefer-as-const single-edit/negative controls PASS in 33.18s, all four outputs identical at 1009 bytes. Combined selected gate PASS in 370.674s. The two excluded malformed sources and multi-edit const coverage are explicit limits, not complete rule certification.

Registry PASS in 0.068s. Whole-repository vet exits zero with empty log. Uncached external one-byte oracle PASS in 0.391s, zero native/Node cache hits. No full repository gate or new throughput measurement; original performance measurements remain historical in the claims and rule reports.

# Exact blockers

The full TestOwnedWitnesses gate FAILS: actual Go panics with unexpected fix shape on the prefer-as-const annotation witness before comparison. A direct Node probe of that same witness prints a finding carries at most one automatic edit and exits 70. These are the ledger's predicted shared multi-edit limitations. The rule retains both edits; implementing a false single repair or editing the shared comparator is outside scope.

The computed-key upstream sources ({ ['x' }); and ({ ['x': 0 }); still require parser recovery; only these two cases are excluded from the supported upstream test and named in its raw log. All other computed-key cases, including JSX, are selected. These parser limits prevent claiming full fixture parity or the harness-only parking exception. Stop before claiming a new helper.

The previously completed helper branch is rebased onto current main and freshly green: four actual-Go/Node/emitted-JavaScript/sanitized-native baselines and four comparison-only mutants PASS 26.323s, whole-repository vet exits zero, external one-byte oracle PASS 0.677s. Details are helpers/from_wave08/LANDING.md. Its four helpers serve the same six Tailwind consuming rules listed in REPORT.md and BREAKPOINTS_REPORT.md; none alone removes all blockers. No new helper claim, main/area push or pull request.

Tool setup most recently ran in the preceding landing refresh: Go/clang/Node/submodules ready 0s each, cache warm 212s, total 212s; nproc is 5 and CPU quota four. Tools are Go 1.27.1, clang 20.1.8 and Node 24.19.0. This integration reuses that toolchain and reruns the actual validation commands above.
