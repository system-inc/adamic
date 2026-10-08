# ecmascript/regexpattern package claim

Branch: lint-helpers/ecmascript-regexpattern. Base: origin/area/stage1-lint at cd56db1dd7ab7eda7a5f492804dcd1849a1494aa.
Triage: d7ab0bc4, rank 19. No retained port (0/18); build fresh. Every higher-ranked eligible package is reserved; tonight exclusions and runtime regex compilation exclusions honored.
Forecast: 0 rules alone, 1 additional with all earlier packages (127 cumulative); conditional helper readiness.
Consumers: no-control-regex and no-regex-spaces. Completing regexpattern and regexsyntax removes no-control-regex's remaining helper blockers; no-regex-spaces also needs ecmascript/literal.

Checked every claims/ file on all origin/lint-helpers/* and origin/codex/lint-helpers* branches after fetch: 30 files, no regexpattern reservation.
No runtime regex compiler is used; this package scans pattern source.
The full walker depends on ecmascript/regexsyntax. Its package is reserved by origin/lint-helpers/ecmascript-regexsyntax, currently claim-only f997f145cceca508f0c2767778fe503463c31354. Build independent regexpattern helpers first; stop dependent helpers rather than duplicate that reservation unless a finished port becomes available.
One helper per file, live consumer Go captures, source Node/emitted JavaScript/sanitized native agreement and one output mutant per helper on every backend required before the finished-unit push. One consuming rule is required when its prerequisite closure is available.

Independent port and explicit stopping boundary:

- Twelve fresh helpers, one per file under stage1/cohere/lint/helpers/ecmascript/regexpattern/. No partial regexpattern port existed to reuse. The test harness is adapted from our ecmascript/text unit.
- Go strings are represented by exact byte arrays (offsets obey Go slice/index preconditions): indexes remain Go byte indexes and malformed UTF-8 remains observable. Character records are copied before callback depth projection; WalkerState owns depth and stopped state and delegates to the caller's actual callback contract. Captures preserve callback observations and caller return values; production helpers contain no Go-answer tables or callback substitutes.
- Supported equivalents: explicit fields plus constructor assignments replace parameter properties; checked array accesses provide strict TypeScript bounds fallbacks. No native/runtime regex compiler is used.
- Stopped Walk (pattern.go:108), walker.run (walk.go:25), walker.walkClass (walk.go:243), walker.readEscape (walk.go:325), escapeValue (escape.go:27) and unicodeEscapeValue (escape.go:85) at their unported ecmascript/regexsyntax dependencies. This is a package ownership/prerequisite boundary, not a compiler language gap.
- No consuming rule descriptor is registered: both no-control-regex and no-regex-spaces require Walk; no-regex-spaces also requires ecmascript/literal. Therefore zero new rules are marked unblocked. The original consuming suites are still used to capture every call to all delivered helpers, including exact zero-call boundaries for isAsciiLetter and octalEscapeValue in those suites.
- Symbol/file/Go-line index: stage1/cohere/lint/helpers/ecmascript/regexpattern/testdata/symbols.json.
- Capture script and fixture counts: stage1/cohere/lint/helpers/ecmascript/regexpattern/testdata/capture.py and coverage.json. All 19 no-control-regex and 126 no-regex-spaces pinned Go fixture cases were asserted during capture; package tests and explicit controls are counted separately.

Final local validation:

- Independent regexpattern package: 68,162 captured Go calls, every invocation retained; 145 consuming fixtures (19 no-control-regex, 126 no-regex-spaces) plus package tests and explicit controls. Source Node, emitted JavaScript and sanitized native output is byte-identical. All 12/12 compiling output mutants complete every capture and are caught on all three runtimes.
- `go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=30m`: PASS for root helpers, comments, regexpattern and module; all package input sets and mutants included.
- `ADAMIC_TYPESCRIPT_SOURCE=/tmp/text-typescript-pinned go test ./stage1/cohere/lint -count=1 -v -parallel=4 -timeout=60m`: PASS (1238.389s), all 4,576 registered source/rule/options combinations, compiler/stage1, JSX discovery, shard, link-table, registration, witness, recovery and 83 registered mutant tests included. TypeScript input pin 050880ce59e30b356b686bd3144efe24f875ebc8.
- After removing five trailing-whitespace lines: the entire owned helper package passed again (161.292s), including all agreement rows, 12 three-runtime mutants and the cohere capture pin; all 929 compiler/stage1 source files passed again (192.850s). No source edit was made during either comparison.
- Vet and staged whitespace checks passed. No generated files, capture logs, shared source edits or incomplete rule registrations are committed.
- Final prerequisite fetch still shows regexsyntax's claim-only f997f145cceca508f0c2767778fe503463c31354. The delivery is explicitly 12/18, with the six dependent helpers and both consuming rule ports stopped at that reserved package. Zero complete rules are newly unblocked. No remaining compiler language gap was encountered by the delivered twelve.
