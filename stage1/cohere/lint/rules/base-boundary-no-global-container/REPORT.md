Built: base/boundary-no-global-container registered on the unified harness, with a handed-node Identifier listener and exact Go message.
Commits: based on area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; source batch4 d486b03a15f3202acc317dc81c40cc21818e21f7; delivery commit follows this report.
Commands and outputs: lint-registry PASS; gofmt output empty; vet PASS; TestRulesAgree PASS 51.05s, TestMutants PASS 724.07s, TestOwnedWitnesses PASS 19.97s; combined PASS 795.105s.
Mutant: reverse the identifier name judgement; compiled and completed on Node, emitted JavaScript and ASan/UBSan native, then caught only by actual Go comparison on all three.
Not covered: full repository gate and independent parser recovery; inherited explicit recovery limits remain visible in unified.log. No shared files edited.

The ledger source is batch4:base_boundary_no_global_container.ts. The original
module and its named helper were inspected. Kind relevance now comes solely
from rule.json kinds Identifier; visit consumes the handed ParseNode and reads
its text. No per-rule kind comparison, helper, type checker or regex is needed.
Messages are copied verbatim from cohere/policy/messages/boundary-no-global-container.json:
Usage of 'getGlobalContainer' is not allowed.

The real upstream functions are TestNoGlobalContainer and
TestNoGlobalContainerRendersTheSourceRulesMessage. The prefix TestNoGlobalContainer
captures both, including all thirteen source-shape cases and the exact-message
assertion. It must not be narrowed to the batch file stem.

Upstream declares schema [] and defines no options type or decoder; its Run
ignores options. Our adapter still parses every supplied field-5 JSON value to
json.RawMessage and returns that non-nil value, preserving the oracle guard and
invalid-JSON rejection. Empty/null uses upstream defaults. No option fields are
invented. Witnesses need no options sidecars and fire on the existing area
harness without depending on the pending witness-options integration.

Witnesses cover the call, aliased import, property key/access and declaration;
a string holding the same spelling is silent. TestOwnedWitnesses checks every
registered witness in selected and all-rule modes and matched 92,250 bytes.
TestRulesAgree matched 13,056,596 bytes. These are unified registry counts, not
counts attributed entirely to this rule. All targets use actual Go cohere as
the oracle; native comparison uses ASan/UBSan. The unchanged shared harness's
unsupported-recovery rows remain explicit, not newly relaxed or suppressed.

Commands, with source /workspace/adamic-tools/env.sh:

- go run ./cmd/lint-registry
- gofmt -w rules/base-boundary-no-global-container/oracle.go (from lint root), then gofmt -l: empty
- go vet ./stage1/cohere/lint
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=30m

Pinned TypeScript corpus 050880ce59e30b356b686bd3144efe24f875ebc8; pinned
cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. No shared harness, registry
generator, compiler or comparison changes. New Adamic source files are .a.
Raw script inputs retain .ts.txt. Full logs are in evidence/.
