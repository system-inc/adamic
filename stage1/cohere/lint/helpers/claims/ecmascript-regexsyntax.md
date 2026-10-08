# ecmascript/regexsyntax package claim

Branch: lint-helpers/ecmascript-regexsyntax. Base: origin/area/stage1-lint.
Triage d7ab0bc4 rank 10: earlier eligible rules/react, ecmascript/text and ecmascript/nextjs are reserved; tonight exclusions honored.
Seven retained helpers / thirteen needed, partial source origin/codex/lint-helpers-05. No complete package port.
Conditional forecast: one rule fully unblocked alone; increment one with prior packages (89 cumulative).
Consumers: @typescript-eslint/prefer-includes, nexus/correctness-no-test-on-global-regex, no-control-regex, no-empty-character-class, no-misleading-character-class, no-regex-spaces, no-useless-backreference, no-useless-escape, prefer-regex-literals. Completing this package removes all remaining helper blockers for no-useless-escape; other rules retain their other package blockers.
Claim pushed before code; every captured helper use and Node/emitted-JS/sanitized-native comparison, Node/native semantic mutants and one consuming rule gate required. Stop explicitly at unported dependencies or compiler gaps.

Partial sources reused: origin/codex/lint-helpers-05 slot05/batch24 (IsHexDigit, AllHexDigits), batch25 (ParseRegexFlags, PatternAndFlags, RegexFlags.UV), batch26 (SkipPatternEscape, ClassEnd). Files relocated into regexsyntax with local imports; callbacks remain caller-supplied UTF-8 widths. Fresh UTF-8 projection supplies actual Go-compatible decoding in the package and rule.

Local helper proof: 18,314 recorded calls (consuming upstream cases, helper seeds and explicit byte controls), 3,275 capture rows, 97,572 identical output bytes on source Node, emitted JavaScript and ASan/UBSan native. All thirteen helper mutants compile, finish successfully and disagree with Go on Node and native. Symbol/file index: ../regexsyntax/testdata/symbols.json. Results: ../regexsyntax/testdata/evidence/helpers.log and counts.json. Capture generation: ../regexsyntax/testdata/capture.py, unchanged Go bodies under an oracle-only overlay, pinned cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.

Helper command: `go test ./stage1/cohere/lint/helpers/regexsyntax -count=1 -v -timeout=30m` (fresh capture by default); final repeated run reused the independently regenerated `/tmp/regexsyntax-capture-final` with `ADAMIC_REGEXSYNTAX_CAPTURE`. Root helpers and comments suites also passed, including all 23,539 root helper cases.

Consumer port: ../../rules/no-useless-escape/, complete upstream TestNoUselessEscape family: 357 executed harness invocations (92 fires, 207 silent, 58 repairs). Full rule corpus comparison passed: 4,176 unique source/rule/options combinations and 13,836,892 identical bytes on Go, Node, emitted JavaScript and sanitized native.

Shared gate limitation: `stage1/cohere/lint/jsx_integration_test.go:53` has a fixed captured JSX inventory and requires a new `"no-useless-escape": 7` row. The first unmodified lint package run failed there; the permitted owned-file scope and docs/lint-registration.md prohibit committing that shared update. A temporary Go test overlay adds only that expected count to exercise all remaining inputs without weakening the comparison. No helper dependency or compiler-language stop.

Full lint proof with the single expected-count test overlay: exit 0, 2321.906s; all 891 compiler/stage1 files, 30,142,337 identical Go/Node/emitted-JavaScript/sanitized-native bytes; 4,351 combined rows identical at 1, 2 and 5 shards (33,695,133 bytes); owned witnesses passed; all 76 registered mutants caught, including the new no-useless-escape quoted-finding mutant on Node, emitted JavaScript and native. Full log: ../../rules/no-useless-escape/testdata/evidence/lint.log. Command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-typescript go test -overlay=/tmp/regexsyntax-lint-overlay.json ./stage1/cohere/lint -count=1 -v -failfast -timeout=60m`. TypeScript compiler pin: 050880ce59e30b356b686bd3144efe24f875ebc8. Benchmark/profile snapshot opt-ins introduce no new correctness input set and were left disabled.

Push withheld: the unmodified shared JSX inventory still fails. Permission requested to include its one expected-count row outside the user's allowed commit scope; absent that permission, this implementation remains a local reviewable commit. No shared test, compiler file or submodule pointer is changed.
