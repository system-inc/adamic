Rebased only this unit's worker branch onto newly advanced main and reran its ten existing rule scopes; no new claims.
Rebased implementation/evidence tip: dec08ff01d39b24b37ac06f5f1cc3619c8487844; main base: e8ba3d5d81de4d3773c723914fccd4c76248b965; final report commit follows.
Setup PASS in 130s, nproc 5; eight profile validators pass, pair suite PASS 235.721s, vet PASS, filtered uncached oracle PASS 0.271s.
All ten compiling semantic mutants were caught by Go byte comparisons on source Node, emitted JavaScript and ASan/UBSan native within their stated scopes.
Still not fully landing-ready: shared JSX witness is red; Google whole-rule parity, custom default-case patterns and the two unimplemented reserved rules remain uncovered.

# Landing-first continuation

Fetch observed main move from e011f8f6 to e8ba3d5d. The previous pushed worker tip was 02a0efb3. git rebase --rebase-merges origin/main completed. Main's lint files had not changed, so the exact earlier registration merge resolution was reused for its six conflicts. git diff 02a0efb3 HEAD -- stage1/cohere/lint was empty immediately after rebase, and main is an ancestor of the new tip. This retains main's new compiler work without a new shared-source edit.

The only push destination is refs/heads/codex/lint-wave1-07, protected by an explicit lease against its fetched old tip. No main or area/ branch is pushed. The latest user request authorizes rebasing this published branch; the leased replacement follows that request rather than the older repository blanket history rule. No other branch was pushed by this unit.

# Fresh commands and observations

- bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh: Go, clang, Node and submodules each 0s; cache warm and total 130s; nproc 5.
- The same eight owned profile commands documented in wave1-07-landing-report.md were rerun with scratch /tmp/wave07-cap/<slug>. this-alias used validate-owned.py; Google used validate-selector.py. All eight exit 0, with their compiling mutants checked on three runtimes.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py --scratch /tmp/wave07-cap/pair --run '^(TestWave07Supported|TestWave07Corpus|TestWave07Options|TestWave07Throughput|TestWave07JsxGap|TestWave07OwnedMutants)$': PASS, 235.721s. This owns 160 supported upstream pair cases, 708 compiler/stage1 file/rule rows, eight directive options, two mutants and a JSX refusal control. Two JSX cases remain excluded from findings parity. Its scratch overlay preserves the original recovery-corpus boundary, not malformed-syntax rejection.
- go vet ./stage1/cohere/lint/rules/...: PASS, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 5m: PASS, 0.271s, one native and one Node cache miss.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m: FAIL, 2.690s. The existing Google-font JSX witness is copied as .ts; Go rejects it with '>' expected before Adamic runs.
- adamic build of the owned default-case/gaps/dynamic-comment-pattern.a remains explicitly refused: RegExp with a nonconstant pattern. The diagnostic is retained; runtime Go-compatible compilation/fallback is not implemented here.

Test output went directly to logs, never through a pipe. Go cohere and TypeScript source pins remain those in the previous landing report. Retained logs and output hashes are in wave1-07-cap-evidence; raw outputs remain under /tmp/wave07-cap.

# Compared scopes

- no-unsafe-negation: PASS upstream: 62 file rows, 21752 identical bytes
- no-unsafe-negation: PASS compiler-stage1: 354 file rows, 13286435 identical bytes
- no-unsafe-optional-chaining: PASS upstream: 103 file rows, 45814 identical bytes
- no-unsafe-optional-chaining: PASS compiler-stage1: 354 file rows, 13286435 identical bytes
- no-underscore-dangle: PASS upstream: 93 file rows, 266000 identical bytes
- no-underscore-dangle: PASS compiler-stage1: 354 file rows, 13401408 identical bytes
- typescript-no-non-null-asserted-optional-chain: PASS upstream: 26 file rows, 8017 identical bytes
- typescript-no-non-null-asserted-optional-chain: PASS compiler-stage1: 354 file rows, 13290307 identical bytes
- typescript-no-non-null-assertion: PASS upstream: 39 file rows, 39811 identical bytes
- typescript-no-non-null-assertion: PASS compiler-stage1: 354 file rows, 14796786 identical bytes
- typescript-no-this-alias: PASS upstream: 34 file rows, 8832 identical bytes
- typescript-no-this-alias: PASS compiler-stage1: 354 file rows, 13286435 identical bytes
- default-case: PASS upstream: 31 file rows, 9777 identical bytes
- default-case: PASS compiler-stage1: 354 file rows, 13471857 identical bytes
- next-google-font-display: PASS selector only: 27 rows including all 16 upstream sources; 4228 identical bytes

Google covers selector decisions over externally supplied attributes, not independently parsed JSX or whole-rule spans. Default-case covers default configuration, not nonempty commentPattern configurations. These bounded successes do not turn the red shared gate green.

# Mutants

- no-unsafe-negation: ordering-relations-ignored; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- no-unsafe-optional-chaining: and-left-undefined-ignored; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- no-underscore-dangle: constructor-property-name-ignored; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- typescript-no-non-null-asserted-optional-chain: parenthesized optional assertion ignored; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- typescript-no-non-null-assertion: index position incorrectly suggested; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- typescript-no-this-alias: parenthesized alias lost; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- default-case: default-comment-anchors-removed; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- next-google-font-display: block-display-accepted; compiled, ran and was caught only by Go byte comparison on all three Adamic runtimes.
- structure/tailwind-no-physical-direction: direction-aware exemption removed; compiled, ran and was caught by Go byte comparison on all three Adamic runtimes.
- @eslint-community/eslint-comments/require-description: empty reason accepted; compiled, ran and was caught by Go byte comparison on all three Adamic runtimes.

Expected JSX and dynamic-pattern refusals are not credited as semantic mutant kills. Google's mutant is selector-only and default-case's holds default-comment anchoring.

# Findings per second

Best of five launches including startup; release native. Compilation and comparisons overlapped, so rates describe this run's load rather than a controlled performance comparison. Google is selector-only and supplies Adamic attributes while Go parses JSX.

| Rule or bounded scope | Native | Node | Go |
|---|---:|---:|---:|
| no-unsafe-negation | 97214.13 | 9392.71 | 190880.49 |
| no-unsafe-optional-chaining | 68758.05 | 8890.00 | 207020.31 |
| no-underscore-dangle | 112330.46 | 9178.22 | 195021.80 |
| typescript-no-non-null-asserted-optional-chain | 4106.89 | 5478.89 | 145574.84 |
| typescript-no-non-null-assertion | 4568.85 | 4458.47 | 158913.06 |
| typescript-no-this-alias | 140335.81 | 8395.15 | 153580.60 |
| default-case | 77010.14 | 7987.51 | 234919.90 |
| next-google-font-display | 115647.21 | 12941.11 | 77690.91 |
| structure/tailwind-no-physical-direction | 144832.01 | 13615.41 | 155220.70 |
| @eslint-community/eslint-comments/require-description | 131139.41 | 12209.70 | 140366.10 |

# Remaining cap

The positive Google JSX control still reports in Go while all three Adamic runtimes explicitly refuse its parser shape. The shared JSX witness gate is also red. Nonempty default-case commentPattern needs runtime Go-compatible regex compilation with invalid-pattern fallback. consistent-return needs compatible Judge / EndReachable integration. constructor-super remains reserved and unstarted. No silent stub, shared-source bypass, full-options parity claim or full repository gate claim is introduced.

Because the full branch remains not landing-ready, no new helper branch or helper claim is taken. No helper was built in this continuation and no queue-exhaustion claim is made. The latest landing-first cap remains in force.
