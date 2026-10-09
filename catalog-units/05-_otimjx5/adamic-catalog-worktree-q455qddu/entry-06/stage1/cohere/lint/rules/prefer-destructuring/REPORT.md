Ported prefer-destructuring from batch 3 to the unified node listener harness.
Baseline: area/stage1-lint bb2ece564; own implementation commit is on this branch.
Checks: registry, gofmt, repository vet, owned witnesses and rule agreement PASS.
Mutant: extra receiver parentheses caught by Go byte comparison on all three backends.
Limits: the full repository 17-check gate was not run; inherited malformed recovery is explicit.

Source: origin/codex/stage1-lint-batch3 fa9781c2c0911c52312002ba97ffd4b560d178ae,
stage1/cohere/lint/rules/prefer_destructuring.ts, with its declaration_initializer.ts
and shared.ts helpers read before adaptation. The descriptor uses node: true and
subscribes only to VariableDeclaration and BinaryExpression. It omits order.
The listener receives ParseNode and index directly. Rule-local initializer,
comment and nested-options helpers live in helpers.a; existing context ancestry,
raw text, parenthesis and literal indexes are reused. No shared files changed.
All new Adamic modules are .a; raw witness text stays outside the module graph.

Messages are moved verbatim to messages.a. The independent Go oracle returns
cohere's PreferDestructuring and decodes field 5 into PreferDestructuringOptions,
including its two per-kind switch pairs and EnforceForRenamedProperties. Captured
JSON is the shipped decoder's typed result, not its raw schema tuple. No nil
adapter or options guard bypass is used. Default options enable both pairs.

upstreamTest is TestPreferDestructuring, which covers all nine actual test
functions listed in evidence/upstream-tests.json. These include both schema
elements, renamed properties, integer-by-value indexes, optional chaining,
const/let regression, all clean/reporting vectors, and fixes and fix declines.
The upstream test run passes. Go rule bodies and the submodule are unchanged.

Four owned witnesses fire: ordinary object access (automatic fix), numeric array
access (no fix), a binding-pattern array access (no fix), and a renamed property
that fires only with renamed.options.json. The last uses the witness-options
support now integrated on the area. The unconfigured all-rule row remains
unconfigured. Its selected finding proves the typed options cross the adapter.

The mutant changes `{foo} = object` into `{foo} = (object)`. Both repairs are
valid and the mutant builds and executes successfully; only the byte comparison
catches the wrong fix. The separate initial targeted gate catches it on Node,
emitted JavaScript and ASan/UBSan native. No sanitizer failure substitutes for
this counterexample. Full mutant run results are recorded below.

Current area checks:
- go run ./cmd/lint-registry: PASS.
- gofmt -l cmd internal plus owned oracle.go: empty output.
- go vet ./...: exit 0, empty output.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/lint
  -run '^(TestOwnedWitnesses|TestRulesAgree)$' -v: PASS 67.321s.
  Rule agreement: 2,130 captured source/rule/options combinations,
  13,074,930 identical bytes across Go, Node, emitted JavaScript and native;
  test duration 47.05s. Owned witnesses: 94,612 identical bytes, 20.25s.
- The initial complete 41-mutant run on d3a37422 passed in 733.445s; the
  current witness-options baseline is rerun separately.

The inherited harness explicitly checks refusal for malformed parser-recovery
cases. Those are not certified as recovered findings. No skip or guard was
relaxed, and this rule hit no parser/type-checker/RegExp/Tailwind blocker.
Repository-wide external TypeScript/postcss/graphql correctness was not run.
The earlier private wave-16 ports are excluded from this clean area-based
branch; their parking note is on codex/typeaware-wave-16. No unified parity
is claimed for those private adapters.

Logs, source hashes and upstream test-name evidence accompany this report.

Final current-area mutant gate:

```
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/lint \
  -run '^TestMutants$' -v > prefer-area-mutants.log 2>&1
```

PASS 691.948s. All 41 descriptor mutants are caught on Node, emitted
JavaScript and ASan/UBSan native, 123 caught comparisons. The owned fix mutant
passes in 16.29s and differs at case 147 line 2753 on each backend:
`{foo} = (object)` versus Go's `{foo} = object`. The new options-only witness
participates in this complete current-baseline run. Mutant names are recorded
in evidence/mutants.json. The lint-area tip was confirmed unchanged at the end.
