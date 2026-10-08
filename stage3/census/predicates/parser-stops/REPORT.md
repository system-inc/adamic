Built: pinned callback signature admission and executable generic/nongeneric argument-body regressions; assertion step pending.
Commits: based on a58ba402b428280d4932afe6c98c1f47a5d3c495, main merge e02efc1fd6782de2c387c211c82e04496ecc6706; callback commit follows this artifact.
Commands/results: callback Node, backend JavaScript, ASan/UBSan native pass; complete internal/lower passes in 21.574s.
Mutants: erased inferred callback proof caught by TestParserCallbackLie; source lie fails on Node with TypeError and exit 70.
Limits: mixed number/string closure argument ABI remains NotYet; explain-checks reports overload calls only; assertion step pending.

The requested base already admits the exact unused callback signature probe. It is
pinned byte for byte from 5d777de3b9506d2a303a4a69d66d21f37f608ede,
`stage3/drivers/parser/native-predicate-callback-parameter.a`, as
`internal/lower/testdata/predicates/parser_callback_parameter.a`.
The executable witnesses exercise narrowing inside the callee with named and
inferred arrow argument functions, including a generic `TOut extends TIn` callback
contract. Existing production admission independently proves each argument body.
No annotation or inferred checker claim substitutes for that proof.

The first attempted execution witnesses used number/string mixed unions. Both
stopped at the existing closure ABI refusal for differently held members. The
conservative scope assumption is to test the predicate lane using the supported
string/undefined representation and retain that independent capability gap.
The original probe remains unchanged, including its unused generic `any` default.

The callback mutant inserts unconditional `return true` into
`proveInferredPredicate`. TestParserCallbackLie then fails with
`want callback body proof refusal, got <nil>`, not a compilation or warning error.
The runner restores the production file in a finally block. The original source
lie passes checker inference via an overload and fails on Node with
`adamic: panic: TypeError: Cannot read properties of undefined (reading 'toUpperCase')`,
exit 70. The independent compiler argument-body proof refuses it before emission.

`--explain-checks` for the original callback probe and the executed generic callback
both prints exactly:

```text
adamic: predicate checks: proven 0 checked 0 unobservable 0
```

These are overload-reporting counts, not a claim that the callback's proof is absent.
The original declaration-only witness emits no callback invocation. Per-direction
reporting for ordinary predicate and callback calls is not implemented by this unit.

Setup used `GOPROXY=https://proxy.golang.org|direct` and
`source /workspace/adamic-tools/env.sh`; no setup failure occurred. An early probe
attempt before the checker checkout finished failed with missing
`cohere/TypeScript/tsc/go.mod`; it was rerun once the dependency was available.
Timing lines: Node 0.085s, Go 0.122s, clang 0.727s, markdown dependencies 1.161s,
submodules 162.088s, Go build 480.451s, deferred test binaries 480.587s,
build cache warm 480.589s, done 480.625s. nproc 5, cgroup quota 4 CPUs.
Go 1.27.1, Node 24.19.0, clang 20.1.8. Cold test compilation overlapped setup.

Commands wrote output directly to logs:

```sh
go test ./internal/lower -run 'TestParserCallback|TestPredicateOverloadRuntime/(parser_callback_parameter|callback_parameter_read|generic_callback_parameter_read)$' -count=1 -timeout 10m
python3 stage3/census/predicates/parser-stops/mutants.py callback
go test ./internal/lower -count=1 -timeout 10m
go vet ./internal/lower
```

Evidence is in `evidence/`. The full repository gate was not run.
