Built: no-nonoctal-decimal-escape candidate in .a with exact Go rule-phase records; decimal suggestions are complete.
Commits: claim 6f28de0b was pushed before code; implementation 26254f9d; evidence commit in git history.
Commands and outputs: bounded Go parity on Node, emitted JavaScript and sanitized native; raw logs and replay saved.
Mutant: mutant.json must compile and exit cleanly before all three comparisons reject its changed behavior.
Not covered: JSX/recovery parser gaps, registered decimal suggestion rendering, ordinary harness, pinned .a self-lint or full gate.

This directory owns its implementation, message, registration, unmodified Go adapter, witness and semantic mutant. No shared parser, compiler, registration or harness source was edited. Complete record types come from the previously owned typescript-prefer-as-const/records.a on this branch; these candidates require that module during integration.

This continuation implements no-multi-str, no-nonoctal-decimal-escape and no-octal. Prior work was pushed through a3f826a4 before selection. After fetching 341 origin refs and inspecting 52 distinct recursive claims Markdown blobs, all helper-ready names were covered. These are the first remaining syntax-ready inventory entries, absent from main ef3d907ecdc4c771b016f7d9c52372def057a340 and all origin claims. Claim 6f28de0b was pushed before implementation.

The multiline rule reads raw string text and all four line terminators, with direct JSX-parent exemption. It additionally refuses TypeAssertionExpression ancestry for strings in JSX files: stage1 can misparse a JSX expression as a type assertion, which otherwise reports a false positive. The decimal rule walks escape-sized units, preserves backslash parity, reports each escape separately, and offers every digit/backslash suggestion. After a real null escape it offers three choices, including the combined out-of-diagnostic-range null rewrite. No automatic fix is offered. Octal checks raw NumericLiteral token text for zero followed by any decimal digit, including 8/9; modern prefixes, cooked values, strings and comments are not its subject.

The rule-owned driver serializes every diagnostic, suggestion ID/message and edit. Go's unmodified rules and formatter decide the expected records. UTF-16 positions are translated to UTF-8 byte offsets. These candidates expose complete records through analyze; no-multi-str and no-octal also report normally through registered visit. Decimal visit refuses positive cases because this branch's finding model cannot expose multiple suggestions and their ranges. The published shared harness work remains outside this unit's territory. This is an explicit integration blocker, not a claim that the registered decimal frontend is complete.

The oracle runs the same rule phase as cohere's own rule tests, including parser recovery. Legacy escapes and octals trigger Go parser diagnostics, and Go's automatic fixer refuses them before doing anything. Since all three rules offer no automatic fixes, replay asserts that Go returns no safe fixes and prints the unchanged source, without invoking that unrelated fixer gate. The first replay attempted it and failed with TS1488/TS1489; parity-1.log retains those failures. That run also caught a false positive in a JSX expression. The correction refuses that parser ancestry on all three backends rather than silently treating JSX content as an ordinary literal.

Setup: `GOFLAGS=-overlay=/tmp/lint-wave1-13-strings/overlay.json bash cloud/setup.sh`, PASS. Go 0s, clang 0s, Node 0s, submodules 0s, build cache 19s, total 19s; nproc=5, quota 400000/100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands write directly to logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -overlay=/tmp/lint-wave1-13-strings/overlay.json ./stage1/cohere/lint -run '^TestWave13Strings' -count=1 -v -timeout 20m > /tmp/lint-wave1-13-strings/parity-final.log 2>&1
go test -overlay=/tmp/lint-wave1-13-strings/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-13-strings/registry-final.log 2>&1
go vet -overlay=/tmp/lint-wave1-13-strings/overlay.json ./... > /tmp/lint-wave1-13-strings/vet-final.log 2>&1
```

Replay: `python3 evidence/reproduce.py --compiler /path/to/typescript-6.0.3`, pin 050880ce59e30b356b686bd3144efe24f875ebc8. Scratch overlays provide .a registration, the compatible profiling snapshot and the bounded driver comparisons. All overlay sources are saved as text and are not production edits. Every compared execution must exit zero with empty stderr, while explicit parser refusals must exit 70 and agree byte for byte across all backends. Native comparison and mutants use ASan/UBSan and leak checks.

Observed blockers: ordinary `go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1` exits 1 at profile_test.go:32:23 because portFiles is a function. Pinned cohere CLI self-lint of the seven new .a modules exits 1, nothing to check because .a is not a TypeScript/JavaScript input. The registered decimal visitor probe exits 70 with no-nonoctal-decimal-escape requires complete suggestion renderer. The exact errors are retained in default-harness.log, self-lint.log and legacy-refusal.log. No full gate or clean self-lint is claimed.

Mutants: omit U+2029 recognition; corrupt only the literal-backslash suggestion while leaving earlier suggestions intact; decline leading-zero decimal 8/9 by narrowing the second digit. Each must compile and run cleanly on Node, sanitized native and emitted JavaScript before only the output comparison is credited. The suggestion mutant specifically proves comparison of more than the first repair record.

Final suite: PASS in 230.598s. Corpus PASS in 182.46s: 810 supported inputs, 37,236,381 identical serialized bytes across Go and all three backends. Each rule covers 77 compiler files and 167 stage1 .ts/.a files, plus 21/31/31 unique upstream vectors and one owned witness. Eight attempted inputs are explicit gaps: five multiline JSX shapes (attribute, attribute braces, spread string, spread object, element braces), and octal recovery cases 01.5, 0777.5, 0755n. Their independent Go rule records execute, while Node, emitted JavaScript and sanitized native refuse with identical stderr and exit 70. These are not counted as covered or merge-ready behavior. An additional inherited return-void JSX capture warning is outside this unit.

| Rule | Covered inputs | Parser gaps | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| no-multi-str | 261 | 5 | 9 | 7.76 | 10.55 | 42.36 |
| no-nonoctal-decimal-escape | 276 | 0 | 18 | 15.34 | 18.34 | 84.31 |
| no-octal | 273 | 3 | 13 | 10.96 | 16.17 | 61.74 |

Throughput is best of three interleaved whole-process runs, including startup, I/O and parsing, on the supported manifest only. Native timing uses a separate unsanitized release build. Counts match Go and Node. The mostly clean corpus and small finding counts make these processing measurements rather than isolated algorithm benchmarks.

Mutants PASS in 35.30s: all three compile, exit zero and have empty stderr on every backend before only the Go comparison rejects them. The decimal mutant changes only the later backslash repair, proving full suggestion coverage. Extra owned-witness comparison PASS in 12.83s with 2,341 identical bytes. It uses the complete-record driver in selected-rule mode; the decimal registered frontend and all-rule integration remain blocked. Registration PASS in 0.060s; vet exits zero with an empty log. The seven final production .a files are unchanged between these runs and their implementation commits.
