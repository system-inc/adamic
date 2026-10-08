Built: seven retained claim rules on wave2 95968dd93, with node listeners and the formerly blocked top-level-await delay witness restored.
Branch: lint-rules/wave1-13-unparked, based directly on origin/lint-batch/wave2-02. Only the seven owned rule directories differ from that base.
Checks: registry, gofmt -l and vet pass; the complete lint package passes in 1763.425s with every available input enabled.
Mutants: all seven owned mutants compile, run cleanly and are caught against unchanged Go cohere on Node, emitted JavaScript and ASan/UBSan native.
Not covered: return-void requires the runtime regex compiler; the shared bridge refusal control skips until TSGoError lands. No complete repository gate is claimed.

| Rule | Unique upstream cases |
|---|---:|
| default-case-last | 38 |
| for-direction | 46 |
| guard-for-in | 20 |
| nexus/consistency-no-hand-rolled-delay | 48 |
| no-constructor-return | 53 |
| no-delete-var | 25 |
| no-eq-null | 14 |

All 244 upstream cases and every owned witness match Go byte for byte on all three modes. None of these upstream cases is excluded for recovery. The restored input is testdata/top-level-await.ts.txt. Descriptors declare kinds by name and node: true. Visits use the supplied ParseNode and let the registry determine relevance. The delay parameter-name helper rejects rest parameters as upstream does.

The shared TestMutants uses one native canary. evidence/native_mutants_test.go.txt adds the seven individual sanitized native comparisons through a Go overlay, without changing any shared harness file. It also compares each rule's complete upstream corpus and witnesses on all three modes. evidence/verify.sh reproduces the full run. Set ADAMIC_TYPESCRIPT_SOURCE to the TypeScript 6.0.3 checkout and ADAMIC_LINT_WORK_DIR to a roomy scratch directory after sourcing the setup environment.

The actual full command was go test -overlay /tmp/wave13-unpark-overlay.json ./stage1/cohere/lint -count=1 -json -timeout=60m. Inputs were ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3, ADAMIC_LINT_BENCH=1, ADAMIC_LINT_PROFILE_DIR=/tmp/wave13-unpark-profile and ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/wave13-unpark-profile. The source pin is 050880ce59e30b356b686bd3144efe24f875ebc8. The package records 42 top-level passes, zero failures and one skip; including subtests, 161 passes, zero failures and one skip. TestCheckerBridgeRefusalPending is the skip, at stage1/cohere/lint/checker_pending_test.go:51; it awaits codex/tsgo-errors-as-values because TSGoError is absent. No input-dependent check skips. TestCompilerAndStage1Agree covers 1080 files and 31,471,801 matching bytes; profile snapshots match 44,221,713 bytes. TestRulesAgree captures 5449 unique upstream source/rule/options combinations. Complete output is evidence/lint-package.jsonl.gz, with evidence/results.json recording the owned counts and caught mutants.

nexus/consistency-no-return-void remains unregistered here under the current runtime-regex rule. The exact blocker is cohere/internal/lint/rules/nexus/consistency_no_return_void.go:116: regexp.MustCompile(`^return\s+void\s+` + regexp.QuoteMeta(operandText) + `\s*;$`). It constructs a pattern from operand source text. Its unchanged Go tests pass and capture 23 unique cases, recorded in evidence/return-void-go.log; no current Adamic parity is claimed for it. Historical withdrawn claims remain with the DEDUP_LEDGER winners. The already merged multiline-arrow-function port is inherited unchanged from wave2.

Setup reported Go ready at 0s, clang, Node and submodules ready at 1s; nproc was 5. Setup's cache warmer failed on the old parser-cases base's missing delay descriptor. The first preflight then failed before comparisons because the root filesystem was full: its Go build cache occupied 27 GB. Clearing that disposable cache recovered 26 GB. The full run used task-local caches; 3,002,213,146 bytes of completed, content-addressed Go cache entries were relocated to /workspace/scratch with their paths preserved by symlinks when the smaller temporary filesystem filled. No source, oracle or check was changed for either infrastructure failure. The first failure and setup output are retained beside the final empty gofmt and vet logs.
