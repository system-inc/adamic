Ported max-nested-callbacks from batch 4 to the unified registry.
Registry, gofmt, vet, TestOwnedWitnesses, TestMutants and TestRulesAgree passed; logs are in evidence/.
The default eleven-deep witness fires without an options sidecar.
The mutant changes <= to < and was caught on Node, emitted JavaScript and native.
No shared harness or registration-generator files were edited.

Source: origin/codex/stage1-lint-batch4 at d486b03a15f3202acc317dc81c40cc21818e21f7, stage1/cohere/lint/max_nested_callbacks.ts. The diagnostic message moved verbatim to messages.a.

The node-aware descriptor subscribes to FunctionExpression and ArrowFunction. The supplied node is used directly. Callback depth is inherited from cached ancestor functions, with direct callees excluded and constructor callbacks controlled by the upstream option. Findings preserve the function head or arrow-token range. This rule has no fixes or suggestions upstream.

The actual upstream test functions are TestMaxNestedCallbacksStaysSilent, TestMaxNestedCallbacksFires, TestDecodeMaxNestedCallbacksOptions, TestMaxNestedCallbacksHandlesNilOptions, TestMaxNestedCallbacksDirectCalleeIsNotACallback, and TestMaxNestedCallbacksReportsTheExceedId. The valid identifier prefix Test includes all six, including the separately named decoder test. Captured rows are filtered by discovered rule names by the existing harness. A direct upstream run also verifies these six functions.

The oracle adapter restores captured Go options to core.MaxNestedCallbacksOptions, and uses the production decoder for configured JSON, including bare integers and maximum/max precedence. Shared all-rule rows carry a union of unrelated options; only this rule's keys are retained before decoding those rows. No option-bearing row returns nil.

Commands:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/max-nested-callbacks/oracle.go
go vet ./...
go test ./stage1/cohere/lint -run '^Test(OwnedWitnesses|Mutants|RulesAgree)$' -count=1 -timeout=30m -v
# From cohere:
go test ./internal/lint/rules/core -run '^Test(DecodeMaxNestedCallbacksOptions|MaxNestedCallbacks)' -count=1 -v
```

Setup: Go, clang, Node and submodules each took 0s; cache warming and total setup took 96s. nproc is 5; cgroup quota is four CPUs.

Scope: the required unified tests and direct upstream tests, rather than the repository-wide gate. Existing malformed method-signature cases remain explicit parser-recovery refusals in the shared comparison; they are unrelated to this rule. No type checker, dynamic RegExp, parser or Tailwind dependency blocks this port.

Mutation result: the eleven-deep witness gains an incorrect depth-ten finding, at column 159 instead of Go's first finding at column 176. TestMutants caught the compiled and executing mutant on Node, emitted JavaScript and sanitized native code.

Final shared run: PASS in 801.037s. TestRulesAgree compared 13,075,327 output bytes; TestOwnedWitnesses compared 90,746 bytes; all 41 registered mutants were caught. Native builds use the shared sanitizer configuration. The direct upstream run passed in 0.008s.
