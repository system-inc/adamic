# React require-optimization validation

Based on `lint-helpers/react`, including the JSX inventory-discovery fix
`e7c196a9`. The [ownership audit](testdata/ownership.json) checked 1,400 origin
refs for a rule directory, manifest registration and ownership declarations.
No competing port or actual claim existed. Five historical selection inventory
mentions are recorded separately; their explicit claims lists are empty.

The oracle adapter calls unmodified Go cohere `RequireOptimization`, at cohere
pin `7945d102a6c18dd36adf9114a758ce646e8b2359`, with its typed options struct.
The shared React helpers supply the component-class recognition.

[Selected test](testdata/selected_test.go) captures all 84 unique upstream
source/rule/options combinations, checks the three owned witnesses have Go
findings, and compares upstream, witness and inherited corpus rows byte-for-byte
on source Node, emitted JavaScript and ASan/UBSan native. Witnesses cover class
fields, parenthesized bases and receivers, spreads, generic calls, export ranges
and allowed decorators, including an [options fixture](testdata/decorators.options.json).
The reproducible selector is [run-selected.sh](testdata/run-selected.sh).

Go behavior deliberately retained: any function-valued object property exempts
an ES5 component (`cohere/internal/lint/rules/react/require_optimization.go`,
`requireOptimizationIsFunctionProperty`); class fields named
`shouldComponentUpdate` do not exempt a class
(`requireOptimizationDeclaresShouldComponentUpdate`). Parenthesized heritage
expressions are declined by the strict shared class helper, while a
parenthesized React receiver is accepted as a component but fails the raw
PureComponent exemption. Allowed decorators must be bare identifiers.

Reports are emitted at finish so source-file rule diagnostics retain Go's
stable precedence at equal offsets. The all-rule classes witness caught this
ordering difference with `max-classes-per-file`; the final comparison includes
it rather than filtering the witness out.

The [mutant](mutant.json) adds class fields to the shouldComponentUpdate method
exemption. It must compile, run and disagree with Go on Node, emitted JavaScript
and native; the classes witness catches it.

The environment sets WASI SDK 27, the clean TypeScript corpus at
`050880ce59e30b356b686bd3144efe24f875ebc8`, ADAMIC_TEST_WASI and ADAMIC_LINT_BENCH,
and one fresh directory for both profile variables. The full merged-helper
lint gate and machine summary are recorded in the React helpers' REPORT.md.
There is no stopped helper or language gap.

[Final selected log](testdata/selected.log): all 175 manifest rows agree,
534,618 bytes identical. The selected comparison passes in 131.23 seconds.
The mutant is caught on Node, emitted JavaScript and native; the complete
selected package run passes.
