# Wave 25 parked from unified landing

No wave-25 rule is certified on the unified harness. The earlier green evidence
is for the standalone checker-bridge suites and must not be presented as
TestOwnedWitnesses, TestMutants or TestRulesAgree certification.

All 14 owned rules are parked for unified-harness type-checker access:

- base/correctness-require-graphql-nullable-parity
- base/correctness-require-matching-inject-type
- nexus/correctness-no-collection-misuse
- nexus/correctness-no-discarded-outcome
- nexus/correctness-no-discarded-pure-result
- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams
- no-throw-literal
- no-useless-backreference
- prefer-arrow-callback
- react-hooks/unsupported-syntax
- react-hooks/use-memo
- react/boolean-prop-naming

Reproducer and required facts: the owned standalone runners instantiate
SyntaxProjection(program, path), and its syntax-projection checker query feeds
every rule. The complete fixtures and reproducing commands are preserved in
validation-wave-25-typeof/typeof-gate.sh. Unified RuleContext exposes source,
parser, scanner, settings and ancestry, but no checker program or question API.
Full semantics additionally require raw-shape/type-symbol (GraphQL and inject),
symbol declarations and references (process, regex constructors and React),
and expression/type facts (collection, outcome and callbacks). Running only
a syntax-only subset would not reproduce the certified upstream rule.

Boolean also has a native compiler blocker, due for runtime support October 9:
validation-wave-25-typeof/options-regexp.a. Reproduce with:

    go run ./cmd/adamic build stage1/cohere/typeaware/validation-wave-25-typeof/options-regexp.a

Observed build exit 1, at 3:28:

    stage 0 can't lower RegExp with a nonconstant pattern yet

The original source and standalone evidence remain on this parked branch,
starting at 5120286ca72a046bdae5b754b8d1be0c22f28572. No landing branch is
created: every owned rule needs the unified type checker before full parity
can be held there. Existing minimal typeaware/rules metadata is not a unified
lint/rules registration. No main or area ref is pushed.
