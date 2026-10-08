# Unit 7 stopped on shared schema

Base: stage1-hir/wip 9addb0e8a. Oracle: cohere
7945d102a6c18dd36adf9114a758ce646e8b2359. No rebase or shared-file edit.
The lane owns this directory only. No pass certificate is claimed.

## Confirmed first blocker

Go `terminal.go:324` defines Scope with Scope, Block, Fallthrough and Order.
`dependencies.go:813` discovers subjects exclusively from these terminals;
`hoistable.go:848` uses the same terminal to key hoistable paths. In this base,
`core.ts:90` has no Scope variant and `replay/decode.ts:134` panics on it.
The Go printer already exports it, so opaque framing is not the missing piece.

The shortest record is in `testdata/scope-terminal.txt`: Scope 1, body block 2,
fallthrough block 3, order 0. `oracle_test.go` exports it through Go's real
TerminalOrder and oraclePayload, without recorded analysis answers. The
`testdata/scope-replay.a` probe imports the owner's replay API and uses
privately minted block handles. `TestScopeReplayGap` checks the exact Go bytes
before attempting each runtime. It is a schema-gap selector, not a semantic
pass mutant. It must be replaced by full comparisons when Scope replay lands.

## Requests to hir-01

1. Add Scope to the shared TerminalType, its ScopeIndex field, checked identity
   lookup in a function-owned scope arena, graph successor/evaluation-order
   handling, and decode/dump/encode. Fresh graph decoding must preserve Go scope
   IDs and translate them to the same arena as the terminal references.
   Preserve Block as the real successor and Fallthrough as structured continuation,
   not an extra CFG edge (`terminal.go:320`). Owner should choose the field names
   once for unit 6 and unit 7. Existing ScopeIndex is available; no duplicate needed.
2. Add concrete privately minted, checked DependencyPathIndex and DependencyTreeIndex
   classes to the single arena home and replay exports. `hoistable.go:89-168`
   interns paths by root and separate plain/optional property maps. The reducer's
   dependency trie is a separate structure (`dependencies.go:414-455`), not that
   registry. A hoistable tree may use a separate arena of the tree handle class;
   checked reads must reject foreign arenas. Unit 7 will own all records/codecs.
3. Extend syntax/checker input facts with value-type presence and numeric type
   flags at each identifier's source handle. Alias/type names alone cannot answer
   `always_invalidating.go:50-91`; it tests TypeFlagsObject without matching names.
   Preserve nil checker, nil node and nil value type separately. These must be
   checker inputs, never precomputed IsAlwaysInvalidatingType results.
4. Relay/import the actual stage1/cohere/mutation_aliasing module and its real
   MutableRanges API, owned elsewhere. It is absent here. Go signatures at
   `hoistable.go:826` and `dependencies.go:1138` require it. This is a module
   import request, not a request for preceding Adamic pass results. Unit 7 will
   import SSA and mutation_aliasing by reference and supply its own Go-produced
   prepass scopes, ranges, invoked-function facts, identity and optional sidecars.

## Coverage and stopping point

Native pass certificate: 0/1465. Node pass certificate: 0/1465.
Emitted JavaScript pass certificate: 0/1465. Flow pass certificate: 0/23.
No census record was sampled, rekeyed or claimed covered. No dependency pass,
hoistable pass or invalidating predicate is implemented or certified yet. No
semantic pass mutant was run; there is no runnable pass to mutate. Earlier
construction/replay certificates belong to the owner and are not counted here.
No language gap is inferred: the demonstrated refusal is a missing shared
schema. Do not work around it by analysing construction's `scopes -`, inventing
numeric handles, stringifying object addresses or copying the imported analyses.

## Validation

`GOMAXPROCS=2 go test -v -count=1 -timeout 20m
./stage1/cohere/high_level_intermediate_representation/passes/unit-7`:
1 pass, 0 fail, 0 skip, 28.206 seconds. Tagged Go input exporter: 1 pass.
Node refusal probes: 1/1. Emitted JavaScript refusal probes: 1/1.
Sanitized native refusal probes: 1/1, empty stdout and exactly the same stderr
and exit 70. These are refusal counts, not pass-state certificates. Full output
is in validation.log. gofmt -l and go vet on this owned package both pass.
No full HIR package/census run was performed because the prepass graph cannot
be decoded. The 23 Flow inputs are not dropped or claimed matched.

The initial fixture expectation included Order inside the JSON and used Go
struct-field order; the Go exporter comparison caught that mistake. The fixture
now uses the real Go sorted payload, with Order only in the terminal prefix.
Emitted JavaScript initially lacked the adamic runtime resolver; the final test
uses the existing oracle/node.mjs for both source and emitted JavaScript.
Neither observation is a shared-code blocker or a semantic pass mutant.

`bash cloud/setup.sh` completed: submodules ready 299.032s, go build ready
513.495s, build cache warm 513.665s, done 513.695s. nproc 5, CPU quota 4;
setup load before 0.00/0.00/0.00, after 8.88/5.52/2.30. Environment sourced
from /workspace/adamic-tools/env.sh. Setup log: /workspace/hir-unit-7-setup.log.
An early test invocation before submodule initialization failed to load
cohere/TypeScript/tsc/go.mod; no check was skipped and it was rerun after setup.

