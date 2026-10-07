Ported @typescript-eslint/no-explicit-any from batch5 onto the unified harness.
Own directory only: .a modules, AnyKeyword subscription, node:true, exact Go message, typed options adapter and local file-extension helper.
Registry generation, gofmt and vet PASS; TestOwnedWitnesses PASS (94616 identical bytes); owned TestMutants PASS on all three backends.
TestRulesAgree FAIL: parser gap on upstream interface constructor with a body; independent reproducer exits 70 at CloseBraceToken offset 54.
Stopped at the known parser blocker; no harness, parser, guard or corpus changes and no claim of complete option/upstream certification.

The batch5 and batch4 copies were inspected against unchanged Go NoExplicitAny.
fixToUnknown emits an automatic unknown replacement; ignoreRestArgs walks
all ancestors and recognizes direct rest tokens, including nested any types.
The adapter decodes manifest field 5 into NoExplicitAnyOptions and returns
that typed value even with defaults; the options-dropping guard is intact.
upstreamTest TestNoExplicitAny captures all 14 real upstream test names.
No checker or regexp operation is needed. The helper is private to this rule.
The existing area snapshot has not merged witness-options; the owned witness
uses defaults, while upstream capture retains each test's complete options.

Commands source /workspace/adamic-tools/env.sh:

    go run ./cmd/lint-registry
    gofmt -l stage1/cohere/lint/rules/typescript-no-explicit-any/oracle.go
    go vet ./stage1/cohere/lint/...
    go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$|^TestRulesAgree$|^TestMutants$/typescript-no-explicit-any-message$' -count=1 -timeout 30m -v

The full shared upstream comparison captured 2224 unique source/rule/options
combinations and failed on this newly selected upstream source:

    interface Greeter { constructor(param: Array<any>) {} }

Copy evidence/reproducer.ts.txt to /tmp/no-explicit-any-reproducer/Thing.ts.
Write its absolute path followed by tab @typescript-eslint/no-explicit-any,
three empty legacy fields, an empty JSON-options field, and recovery as field 6
to manifest.txt. Regenerate the registry and run:

    node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/lint/main.ts --manifest /tmp/no-explicit-any-reproducer/manifest.txt

Exact result: exit 70, adamic: panic: parser slice unsupported primary
CloseBraceToken at 54. Unmodified Go's TestNoExplicitAnyFires/upstream_fail_15
passes and requires one unexpectedAny finding on this same malformed source.
See go-reproducer.log and reproducer.stderr. No recovered case was removed
or downgraded to unsupported-recovery to obtain a green comparison.

The compiled mutant appends ! to the exact diagnostic message and is caught
by Go comparison on source Node, emitted JavaScript and sanitized native.
The initial parent-array access failed type checking, then was corrected to
RuleContext.parent; that build failure is not counted as a mutant catch.
TestMutants ran only this owned semantic mutant; no full mutant sweep or
full repository gate is claimed. Full options and fix agreement are not
certified because TestRulesAgree stops at the parser refusal.
