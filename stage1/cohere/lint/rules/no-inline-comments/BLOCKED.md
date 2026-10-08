# no-inline-comments: blocked on Go RE2 option matching

Stopped under task instruction 4 before creating a partial registration or a private shared helper.
Base: origin/area/stage1-lint, 9b7547976bea1b00f8b04508b9102d03511f65b3.
Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.

Required symbols in cohere/internal/lint/rules/core/no_inline_comments.go:

- regexp.Compile(resolved.IgnorePattern), line 103: arbitrary user patterns, invalid patterns ignored.
- (*regexp.Regexp).MatchString(body), line 171: Go RE2 semantics against trimmed inner comment text.

The shared stage1/cohere/lint/regex/options.a optionPattern constructor creates a JavaScript RegExp with flag u. It is not a Go RE2 compiler/matcher. No shared Adamic helper implements this option contract. In particular, Go accepts (?i)todo, \p{Greek}, a\z and (?P<word>a), which that constructor rejects; Go \s excludes U+00A0, which JavaScript \s includes. Substituting JavaScript matching, constraining user patterns, enumerating the fixture patterns, or implementing a local translator would violate the requested oracle contract or shared-helper restriction.

Evidence: evidence/option-dialect.log, produced with:

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/regex -run '^TestOptionDialectGap$' -count=1 -v -timeout=5m

Read whole: docs/lint-registration.md, stage1/cohere/lint/helpers/README.md, both no-warning-comments and unicode-bom ports, and all three requested upstream files. Also inspected the shared comment and regex documentation and code. Historical comment-helper documentation claims a JSX parser gap; that claim is stale on this base: stage1/typescript/parser/jsx.ts implements JsxExpression. JSX is not the blocker reported here.

No port parity, rule mutant, registry validation or whole lint-package run is claimed. Those steps depend on the missing shared Go-compatible runtime pattern helper. Upstream has 49 corpus rows plus five directive asymmetry rows; zero were certified against a port. No Go cohere source was edited.
