# ecmascript/regexpattern package claim

Branch: lint-helpers/ecmascript-regexpattern. Claim commit: df0e695283113b123754908ca456330a16d20b41. Base: origin/area/stage1-lint at cd56db1dd7ab7eda7a5f492804dcd1849a1494aa.
Triage: origin/lint-helpers/triage d7ab0bc4, rank 19. No retained regexpattern port existed (0/18); all eighteen bodies are fresh. The test harness was adapted from our ecmascript/text unit.
Forecast: 0 rules alone, 1 additional with all earlier packages (127 cumulative); conditional helper readiness. Consumers: no-control-regex and no-regex-spaces. No runtime regular expression compiler is used.

The initial exhaustive helper claim scan found 30 claim files and no competing regexpattern reservation. The first delivery bfea2f868230f1675235a4a31f16fe89abe46d7e stopped six helpers at the then claim-only regexsyntax prerequisite. That boundary is superseded: completed origin/lint-helpers/ecmascript-regexsyntax bca49a5ac1694099bcae388642126bf5281f1702 is merged by bf179d90c, and its real shared scanner/parser helpers are reused directly.

Consuming rule reservation: no-control-regex. After fetching every origin branch, no branch had stage1/cohere/lint/rules/no-control-regex/ and no rule claim reserved it. Old selection inventories explicitly showed claims: []; helper consumer lists describe prerequisites rather than rule ownership. no-regex-spaces is left to the ecmascript/literal worker whose claim names that intended proof rule. Completing regexpattern and regexsyntax removes no-control-regex's helper blockers; no-regex-spaces also requires ecmascript/literal.

One helper per file under stage1/cohere/lint/helpers/ecmascript/regexpattern/; exact symbol/file/upstream-line mapping in testdata/symbols.json. Go strings remain UTF-8 byte arrays, including malformed bytes, so all offsets retain Go's byte semantics. Character records are copied before callback depth projection. Captures retain actual caller decisions and all callback observations, stopped flags, and depth transitions. Production helpers use the completed regexsyntax scanner and character-class parser; they contain no captured-answer tables.

Supported equivalents: explicit fields and constructor assignments replace parameter properties; checked switch operands provide concrete number values; explicit depth restoration before each return replaces Go's defer. No compiler language gap remains. Walk, run, walkClass, readEscape, escapeValue, and unicodeEscapeValue are all complete.

Actual Go overlay capture: 73,584 invocations across all 18 helpers. Both consuming suites are captured in full: 19 no-control-regex cases and 126 no-regex-spaces cases, plus the regexpattern package suite and explicit controls. Coverage and per-input-set counts are in testdata/coverage.json; regeneration is testdata/capture.py. Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.

Consuming rule directory: stage1/cohere/lint/rules/no-control-regex/. It owns its registration, exact message, Go oracle, witness and control-range mutant. Findings batch spelled control codes once per literal and preserve UTF-8 source ranges through the shared UTF-16 projection.

The JSX worker can consume UnescapeStringLiteralText from lint-helpers/ecmascript-text at 74047eb86456284a8d36d4d8dbf9cb5022966e79.

Completion validation is recorded below after all local packages finish. No further unit is started; one final completion push is authorized.

Final completion validation:

- `go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=40m`: PASS for every package and input set, including the merged regexsyntax package. Owned regexpattern: PASS in 481.458s; 73,584 actual Go calls agree byte-for-byte on source Node, emitted JavaScript and ASan/UBSan native. All 18 compiling output mutants finish every input and are caught on all three backends.
- `ADAMIC_TYPESCRIPT_SOURCE=/tmp/text-typescript-pinned go test ./stage1/cohere/lint -count=1 -v -parallel=4 -timeout=60m`: PASS in 1179.459s. All 4,892 unique upstream source/rule/options combinations, owned witnesses, compiler/stage1 sources, JSX discovery, registration, shards, link-table, recovery checks and all 85 registered rule mutants pass. The no-control-regex control-range mutant is caught on source Node, emitted JavaScript and native; ordinary rule comparisons use sanitized native. TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8.
- `go vet ./stage1/cohere/lint/helpers/ecmascript/regexpattern`: PASS. All eighteen symbol/file entries and mutant records reviewed; all owned text files pass whitespace review. Cohere worktree remains unchanged. No production source edit was made during these full comparisons.
- Newly complete helper closure: no-control-regex (ported, all 19 upstream cases green). no-regex-spaces gains the complete regexpattern prerequisite and becomes helper-ready together with ecmascript/literal; its rule is left to that worker. No helper is stopped and no remaining language gap is recorded.
- Only the owned helpers, their capture/tests, the existing claim, and the no-control-regex directory are included in the completion commit. Generated dispatch, scratch logs and unrelated shared source changes are excluded.
