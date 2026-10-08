# CFG integration stop: appendSuccessor

This is a reproduced boundary, not a completed helper package. No helpers or rule directory are exported by this work.

Claim-only remote commit: 672e42ed1b03a5ac5d7ba7066386b91b23a5a8bd. Refetch after the claim found no earlier competing package claim. Existing historical symbol claims are retained partial-port ownership, not a complete package claim.

## Reproduced on the current lint base

Go `cohere/internal/lint/ecmascript/control_flow_graph/cfg.go:329-337`, specifically line 336 `from.Successors = append(from.Successors, to)`. A real private-method self-edge call returns `1,7,true`; source Node agrees. The semantic mutant writes the target index and returns `1,8,true`, caught on source Node. Both baseline and mutant fail common lowering before JavaScript emission or native compilation:

`Adamic 0.1 refuses Block[], an array whose elements can reach back to an array like it: a cycle reference counting can't free ... (adamic/cycle-capable)`

The complete current diagnostic is reproduced by `gap_test.go`, not replaced by a golden historical log. Run `go test ./stage1/cohere/lint/helpers/control_flow_graph -count=1 -v -timeout=10m` with the setup environment sourced.

## Scope and supported alternatives

The retained direct-object probe is reused byte-for-byte from `origin/codex/lint-helpers-from-codex-lint-wave1-03`, `slot_wave1_03/gaps/`. Its overlay invokes the actual pinned Go helper; cohere sources are untouched.

This does not prove that the algorithm cannot be implemented in the language. A numeric block arena is a supported representation to investigate. It must be integrated with the 65 retained partial helpers and preserve block identity and graph ownership. The Go helper also writes the fixed two-slot `successors` backing array when possible (lines 331-334); saved slice views can share that backing storage. Replacing this with ordinary independent arrays or weak references is not a proven equivalent. The existing projected callbacks do not establish that contract. That integration is unfinished; do not advertise the direct-object refusal as a refusal of all possible arena equivalents.

Only one control call has been checked here. Full consuming-rule capture suites, all 79 helpers, emitted-JavaScript/native agreement and native semantic mutants have not passed for this package. No rule was ported; zero rules are newly unblocked. Existing full lint/helper results from the Tailwind investigation are not CFG proof. Stop at this helper with these limits explicit; retain all implementation/probes locally and push no partial code unit.
