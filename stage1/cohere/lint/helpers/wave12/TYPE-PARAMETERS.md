Delivered control_flow_type_parameters.a: ordered constraint/default expression delegation for supplied valid type-parameter records.
Claim 9481ccd58 pushed before code; current main b8fb957aa and owned rule branch c485ad2cb remain landing-ready under the documented harness parking exception.
Focused new helper suite PASS 6.677s, vet clean; 1,292 observations / 2,655 identical actual-Go/source Node/emitted JS/ASan+UBSan native bytes.
Drop-constraint mutant caught at row 1280 and swap-order mutant at row 1281; both compile/run cleanly on every backend and fail only Go comparison.
Full expression/CFG semantics, malformed internal Go AST lists, nil nodes and whole rule findings are outside this contract; zero final blockers removed.

`typeParameters<E>(node: TypeParameterizedNode<E>, visitExpression: (expression: E) => void)` walks supplied valid parameter records in list order, delegating constraint and default type in that order. Empty expression slots are carried by the caller's E representation; the observation adapter uses numeric pointer identity zero for nil and preserves both calls. The public API contains no parser lookup, string kind dispatch or silent malformed-node acceptance.

The original Go method remains untouched in an oracle-only overlay. A wrapper records calls made directly to Builder.expr and invokes the original walker, excluding recursive calls by depth. AST facts contain only expression identities. Go's TypeParameterList admits function-like nodes and class, interface, alias and JSDoc-template kinds. Every captured consumer AST is scanned; comparisons call the helper only on that domain. Go's method panics on unrelated AST kinds. A malformed-list probe with an Identifier confirmed that AsTypeParameterDeclaration panics before the apparently permissive nil guard. The early implementation proposal to skip wrong-kind entries was therefore removed; it was never delivered as a matching helper. The preserved negative log evidence/type-parameters-malformed.log records the actual cast refusal, not a green mutant or production helper.

All 2,119 captured input fixtures are covered. Admitted original nodes: array-callback-return 269, consistent-return 92, no-unreachable-loop 639, react-hooks/rules-of-hooks 279, 1,279 total. They delegate zero type-parameter expressions. Twelve explicit controls contribute 13 admitted nodes and 32 calls, covering absent constraints/defaults, both slots together, multiple parameters, class/interface/type-alias/function/method/arrow declarations, Unicode, typeof reads and malformed source accepted by the parser. Original fixture empty-list coverage and explicit positive controls are kept separate; no arbitrary internal-AST parity claim is made.

The four consumers each lose a seventh CFG dependency supplied by this unit. Many other builder methods and graph analysis dependencies remain; zero rules lose their last helper blocker. This helper does not implement the separately owned expression visitor.

Commands (source /workspace/adamic-tools/env.sh):
```
go test ./stage1/cohere/lint/helpers/wave12 -run '^TestTypeParametersMatchesGo$' -count=1 -v -timeout=10m > /tmp/wave12-type-parameters.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-type-parameters-vet.log 2>&1
```
The earlier full owned suite for the other eight delivered helpers passed 123.469s, with all 22 controls. The new two mutants bring the owned suite to 24 controls, with the earlier refusal-domain distinction preserved. New focused comparison and final package vet pass. Tests are logged directly; no full-repository gate or new throughput measurement is claimed. Setup timings remain Go/clang/Node/submodules 0s, cache warm and total 122s, nproc 5 (four-core quota). No shared harness, compiler or registry file is edited. The exact nextBuildCount support probe remains blocked by bigint-return lowering.
