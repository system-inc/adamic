# Regex table on the scout merge

The base is area/stage1-lint 9156bf5c5, merged with area/library, cohere-pin/f5d1934a-land and codex/lint-regex, without rebasing. Cohere is f5d1934a2d7bebe706210cb1cfd01aebff4f8ca7.

`all-patterns.json` is the engine-aware combined table: 123 source sites, plain 80, dynamic 39, (?i) 2, (?s) 1, (?m) 1. It preserves the source file, line, expression, engine, flags and translation or unchanged runtime-source contract. No fixed Go site uses a Unicode property, \z or named capture.

`sites.json` and `table.json` retain the standard Go-regexp census: 89 sites, 88 MustCompile and one Compile, 82 fixed and seven constructed. Their shapes are plain 78, dynamic 7, (?i) 2, (?s) 1, (?m) 1. `patterns.a` translates the fixed sources once at port time. Every row has an executable string-pattern shape fixture; native is now mandatory, with no pending-refusal exemption.

`option-sites.json` retains all 34 esregexp sites separately: two fixed JavaScript patterns and 32 dynamic sources. Their flags are recorded as source expressions, including runtime flags. Option sources are JavaScript by contract and never translated through RE2. The three production migrations use the runtime builders in id_length.a, no_inline_comments.a and no_warning_comments.a. Only the common positions.a reporting boundary converts a finding's Go byte range to UTF-16; comment geometry uses that same coordinate adapter.

The fixed-pattern differential retains the explicit alternate 100-file corpus in testdata/corpus-files.json, 2,104 identifier/comment/control strings. The fleet's named quiet-hundred manifest is absent from the fetched base and workspace; this corpus is not claimed to be that manifest.

The trace-only upstream capture does not change rule matching semantics. It records the current esregexp source via Source(), then compares all 204 id-length, 58 no-inline-comments and 88 no-warning-comments fixtures against the actual registered rule directories. Matcher observations and each planted mutant also run on source Node, emitted JavaScript and sanitized native.

See gaps.md for the current Go-side Unicode-property and shared option-capture gaps. No matcher fallback or shared harness/compiler edit is added.
