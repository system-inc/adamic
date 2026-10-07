Rebased the existing worker branch onto current origin/main and revalidated ten owned rule scopes; no new helper claim.
Rebase implementation tip: 43b264b7b450cb360525f18325cc8f498d772313; main base: e011f8f60899586d6373a5ccb07335ad82cfbf3c; final evidence commit follows.
Setup PASS in 97s, nproc 5; eight profile validators and the two-rule overlay suite pass; owned vet and filtered uncached mismatch oracle pass.
Ten semantic mutants compiled and ran, then Go byte comparison caught each on source Node, emitted JavaScript and ASan/UBSan native within its stated scope.
Not fully landing-ready: shared JSX witness fails; Google whole-rule parity, custom default-case options, consistent-return and constructor-super remain unfinished.

# Rebase and ownership

The only branch pushed by this unit is codex/lint-wave1-07. Its old tip 94af70f0 was not on main. git rebase --rebase-merges origin/main completed, preserving foundation merges. The old registration merge conflicted in six shared lint files. git diff efeb3f66^1 origin/main -- stage1/cohere/lint was empty, so its exact already-tested resolution was reused. Immediately after rebase, git diff 94af70f0 HEAD -- stage1/cohere/lint was empty. The rebase retains newer main compiler fixes and does not introduce a new shared registry, harness, parser or compiler change.

The latest user explicitly asked for a rebase and push of already-published work. The replacement push uses an explicit lease against this branch's fetched old tip. No other branch is rewritten. This follows that latest request rather than the repository's earlier blanket prohibition on rewriting history.

# Fresh comparisons

Go cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db, and TypeScript compiler sources at 050880ce59e30b356b686bd3144efe24f875ebc8. Eight independent current owned profile validators compare actual Go rule observations with source Node, emitted JavaScript and ASan/UBSan native, and check their compiling mutants. Raw logs are beside this report. Raw output files remain under /tmp/wave07-landing; output-digests.json records their lengths and hashes.

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

The initial batch invoked this-alias's historical validate.py with the newer script's --compiler argument. It exited 2 before any test ran. The documented validate-owned.py was then run and passed; corrected-alias-status.json supersedes that CLI error in initial-batch-results.json. No oracle failure is hidden.

The directive/Tailwind historical owned runner was refreshed to detect the integrated .a harness and skip its obsolete compatibility patch. Its recovery-guard source matcher was updated. Its owned test filters this unit's pair rather than treating later workers' captured cases as covered. A scoped mutant hook uses the two owned witnesses and the existing shared comparison primitives. Only owned files change. Scratch Go overlays preserve original filenames and Go rule bodies, and permit the upstream recovery source corpus; they do not certify malformed-syntax rejection.

The two-rule suite passed in 212.881s: 160 supported own-pair captured cases, 48,906 identical bytes; 708 compiler/stage1 file/rule rows, 26,636,700 identical bytes; eight directive option controls, 8,689 identical bytes; both compiling mutants on all three Adamic runtimes; and a positive Go JSX control paired with three explicit stage1 parser refusals. Two JSX sources remain excluded from this pair's findings parity. Capture observed 1,870 cases across all registered rules; later workers' cases are filtered from this owned suite rather than claimed covered.

# Commands and outcomes

- bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 97s, total 97s, nproc 5.
- python3 <owned rule>/validate.py --scratch /tmp/wave07-landing/<slug> --compiler /tmp/wave07-typescript for the original three, both assertions and default-case; validate-owned.py for this-alias; validate-selector.py for Google Font Display. All eight current validators exit 0.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py --scratch /tmp/wave07-landing/pair-current --run '^(TestWave07Supported|TestWave07Corpus|TestWave07Options|TestWave07Throughput|TestWave07JsxGap|TestWave07OwnedMutants)$': PASS, 212.881s.
- go vet ./stage1/cohere/lint/rules/...: PASS, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 5m: PASS, 0.318s, one native and one Node cache miss.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m: FAIL, 3.256s, existing Google-font witness copied as .ts and rejected by Go with '>' expected. The red shared check remains recorded.

Test output was sent directly to logs. No test run was piped. The full repository gate was not run.

# Semantic mutants

- no-unsafe-negation: ordering-relations-ignored. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- no-unsafe-optional-chaining: and-left-undefined-ignored. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- no-underscore-dangle: constructor-property-name-ignored. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- typescript-no-non-null-asserted-optional-chain: parenthesized optional assertion ignored. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- typescript-no-non-null-assertion: index position incorrectly suggested. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- typescript-no-this-alias: parenthesized alias lost. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- default-case: default-comment-anchors-removed. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- next-google-font-display: block-display-accepted. Compiled and ran; Go byte comparison caught it on Node, emitted JavaScript and sanitized native.
- structure/tailwind-no-physical-direction: direction-aware exemption removed. Compiled and ran; Go byte comparison caught it on all three Adamic runtimes.
- @eslint-community/eslint-comments/require-description: empty reason accepted. Compiled and ran; Go byte comparison caught it on all three Adamic runtimes.

These mutants finish normally; compiler or sanitizer failures do not kill them. Google's mutant holds its selector decision on external attribute data only. Default-case's holds default-comment anchoring only. Expected JSX and dynamic-pattern refusals are not credited as semantic mutant kills.

# Findings per second

Best of five process launches, including startup, using release native. Concurrent compilation and comparisons ran during measurement, so these are observations under this run's load, not a controlled cross-branch performance comparison. Google is selector-only: Go parses JSX and Adamic receives externally supplied attributes. Its rates do not establish a whole-rule performance advantage.

| Rule or bounded scope | Native | Node | Go |
|---|---:|---:|---:|
| no-unsafe-negation | 89917.74 | 9224.41 | 191858.23 |
| no-unsafe-optional-chaining | 21253.54 | 8277.89 | 115753.74 |
| no-underscore-dangle | 116212.36 | 8617.83 | 205065.36 |
| typescript-no-non-null-asserted-optional-chain | 4474.76 | 5338.09 | 77639.00 |
| typescript-no-non-null-assertion | 4200.99 | 4148.93 | 127008.72 |
| typescript-no-this-alias | 153035.23 | 10448.11 | 172092.10 |
| default-case | 73793.88 | 6449.11 | 220069.41 |
| next-google-font-display | 112130.28 | 11194.05 | 75571.71 |
| structure/tailwind-no-physical-direction | 138124.73 | 13993.54 | 145617.29 |
| @eslint-community/eslint-comments/require-description | 127477.27 | 13490.77 | 145774.91 |

# Remaining blockers and landing cap

Google's positive JSX fixture reports googleFontDisplayMissing in Go. Source Node, emitted JavaScript and native each exit 70 at GreaterThanToken / Identifier at 32. The shared witness additionally fails before Adamic execution because that witness is transported as TypeScript. Neither parser nor harness is edited here.

The current compiler still refuses the owned dynamic-comment-pattern.a with "stage 0 can't lower RegExp with a nonconstant pattern yet". A custom-pattern control makes Go exit 0 clean, while all three freshly built Adamic profiles produce identical explicit NotYet output and exit 70. A runtime RE2-compatible compiler and invalid-pattern fallback remain needed. This is not full default-case option parity.

consistent-return still lacks compatible Judge / EndReachable integration. constructor-super remains reserved and unstarted. Neither is represented by a silent stub. Full shared all-options/all-rules parity is not green and is not claimed. The latest landing cap stays in force: no new helper branch or claim, no helper built, and no assertion that the unclaimed helper queue is exhausted.
