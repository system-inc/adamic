Ported @typescript-eslint/no-inferrable-types into its own unified harness descriptor directory.
Branch starts from area d3a37422c; source port is batch4-typescript 63782c536:stage1/cohere/lint/no_inferrable_types.ts.
Registry, gofmt and vet pass; TestOwnedWitnesses passes, but TestRulesAgree is blocked by multi-edit automatic fixes.
The number-inference mutant is caught on source Node, emitted JavaScript and sanitized native output.
Full upstream parity is not certified; shared automatic multi-edit serialization is required.

The descriptor uses node:true, named VariableDeclaration/PropertyDeclaration/Parameter buckets,
a concrete factory and the handed node. Only this directory changed. The old batch declaration
parts helper is rule-local; existing RuleContext.functionLike/has and scanning helpers are reused.
Messages are moved verbatim to messages.a. The adapter unmarshals field 5 into upstream
NoInferrableTypesOptions, retaining false defaults and returning a typed value. It does not
return nil for configured rows. The upstream strict decoder cannot handle all-rule bags containing
other rules' keys; typed json.Unmarshal follows the registration contract and preserves our flags.

upstreamTest is TestNoInferrableTypes and captures all five actual functions, retained in
upstream-tests.json, including MatchesUpstream (99 cases), RewritesWhatUpstreamRewrites (44 fixes),
readonly exemptions, independent option suppression and parameter-property anchors.
The witness fires three findings, with a readonly property as a clean control. The area base
has not yet merged witness-options 29c41e102, so the witness uses default options. Independent
selected-rule Go/Node probes cover all four option combinations, matching counts 3,2,2,1.
Those probes certify counts, not complete wire output for nondefault options.

Commands with /workspace/adamic-tools/env.sh sourced:

- go run ./cmd/lint-registry: exit 0.
- gofmt -w on the owned oracle.go; gofmt -l cmd internal and the owned adapter: empty output.
- go vet ./...: exit 0 on final source.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$|^TestRulesAgree$' -count=1 -v -timeout=30m:
  TestOwnedWitnesses PASS 20.22s, 91836 identical bytes across unchanged Go, source Node,
  emitted JavaScript and sanitized native, including all-rule runs. TestRulesAgree FAIL 171.41s
  at the Go wire guard panic unexpected fix shape; 2128 unique upstream combinations captured.
  Existing method-signature malformed cases remain explicit parser recovery limitations in the
  shared run. No guards, tests or shared sources changed.
- go test ./stage1/cohere/lint -run '^TestMutants$/number-inference$' -count=1 -v -timeout=15m:
  PASS 26.428s. Changing the numeric-literal inference result from number to empty removes real
  findings/fixes while compiling and running successfully; byte comparison catches the mutation
  on all three execution paths. Full TestMutants over unrelated rules was not run.

Blocking reproducer, evidence/multifix.ts.txt:

```ts
const fn = (a?: number = 5) => {};
```

Selected unchanged Go oracle --count reports 1 and exits 0. Its normal wire run reports the
correct finding then panics unexpected fix shape at oracle.go:75, exit 2: upstream emits one
removal for '?' and a separate removal for ': number'. Source Node refuses with a finding
carries at most one automatic edit, exit 70, through RuleContext.reportNode. The same limitation
applies to definite-assignment class properties. The port preserves both upstream edits;
it neither drops one nor merges the two to make the guard pass. The reproducer stays outside
owned testdata witnesses so the ordinary supported witness and real mutant remain executable.
The full upstream corpus still includes it and correctly fails certification.

Compilation filled Go's reproducible cache. Standard go clean -cache restored workspace capacity;
no source, evidence, fixture or shared code was removed. The toolchain was already installed and
its environment was sourced; setup was not repeated. No full repository gate, compiler throughput
benchmark or complete upstream byte certification is claimed. Work stops at the shared multifix
model/wire blocker, with the partial port and logs preserved for integration.
