# Shared checker landing blockers

Current area merge is c4bdc23fa. Active unified descriptors use only RuleContext.checker.
The former standalone drivers and their reports remain historical proof artifacts; they
are not a second checker wired into the unified harness. No new claims were made.

## Existing ports awaiting shared dependencies

| Rule | Missing dependency | Upstream consumer |
| --- | --- | --- |
| nexus/correctness-no-process-exit-after-output | Production routing for wave07-control-flow and wave07-symbol-context | cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:314 GetResolvedSignature and :511 control_flow_graph.Build |
| nexus/correctness-no-uncleared-race-timeout | Production routing for wave07-symbol-context, including library origin and reference/read facts | cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:173 type_checking.IsSymbolFromDefaultLibrary and :335 GetSymbolAtLocation |
| nexus/correctness-require-blocking-standard-streams | Production routing for wave07-program-modules and wave07-control-flow | cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:283 GetSourceFileForResolvedModule and :840 control_flow_graph.Build |
| prefer-regex-literals | Production routing for wave07-regex-structure; regex grammar facts must replace the historical literal matcher | cohere/internal/lint/rules/core/prefer_regex_literals.go:566 and :835 regexpattern.Walk |
| react/jsx-fragments | External declaration AST/initializer view through the shared checker | cohere/internal/lint/rules/react/jsx_fragments.go:303 jsxFragmentsDeclarationIsFragment |
| react/no-adjacent-inline-elements | External pragma declaration/initializer view through the shared checker | cohere/internal/lint/rules/react/checked_requires_onchange_or_readonly.go:351 bindsToPragmaImport, :365 declarationComesFromPragma, :401 initializerComesFromPragma |

The private question implementations remain in their owned bridge files, unchanged
apart from compatibility with the incoming SourceFile/path APIs. They are not
registered in production. A test calling an implementation method directly does
not certify its production route. Shared routing was not changed by this unit.

For external declaration views, node-symbol-details supplies declaration anchors,
not the initializer AST of an arbitrary other file. The historical standalone
pragma helper fetched that other file through its private program view. It cannot
be transplanted as a shared checker substitute. Example projects needing this
view have a global `const F = React.Fragment;` in other.ts and `<F />;` in main.tsx;
or a global `const createElement = React.createElement;` in other.ts and a
createElement call containing adjacent inline children in main.ts. These examples
explain the dependency, not a new claim of measured parity.

## Parked hook analyses

| Rule | Upstream consumer not available to the Adamic rule |
| --- | --- |
| react-hooks/set-state-in-effect | cohere/internal/lint/rules/react/set_state_in_effect.go:267 high_level_intermediate_representation.ForFunctionWithoutManualMemoization |
| react-hooks/set-state-in-render | cohere/internal/lint/rules/react/set_state_in_render.go:183 high_level_intermediate_representation.ForFunction |
| react-hooks/static-components | cohere/internal/lint/rules/react/static_components.go:120 high_level_intermediate_representation.ForFunction |

These need native high-level IR, SSA and capture facts. Their existing parked
claims and implementation/probe history were preserved.

## Mandatory skip retained

TestCheckerBridgeRefusalPending awaits TSGoError in internal/load/prelude.d.ts
and the bridge error-as-value transport. Its exact skip reason and all package
results are archived beside CHECKER_LANDING_REPORT.md. No input-dependent check
was deliberately disabled, and no skip or failed assertion was removed.

## Newly measured unified-harness blockers

- @typescript-eslint/consistent-indexed-object-style: two captured clean recovery
  inputs, `interface Foo {\n  [];\n}\n` and `type Foo = { [] };\n`,
  reach stage1/typescript/parser/parser.ts:176 primary() and exit 70 on
  CloseBracketToken. Cohere's TestConsistentIndexedObjectStyleStaysSilentOnUpstreamPassCases
  includes these at cohere/internal/lint/rules/typescript/consistent_indexed_object_style_test.go:114
  and :112. The other 119 captured cases match all four runtimes. Their sources,
  manifests, unchanged Go outputs and native stderr are in checker-landing evidence.
- prefer-promise-reject-errors: the captured `input.tsx` containing
  `Promise.reject(<string>'x');` is an intentionally malformed JSX recovery
  case at cohere/internal/lint/rules/core/prefer_promise_reject_errors_test.go:298,
  TestPreferPromiseRejectErrorsShapesUpstreamCannotWrite. The shared oracle
  collect() at stage1/cohere/lint/testdata/oracle.go:186 rejects the captured row
  without a recovery marker and exits 2 before any port comparison. The other
  123 captured cases match all four runtimes. No row was silently renamed to .ts
  or omitted.
- The new descriptors add 45 JSX cases for react/jsx-no-undef and one for
  prefer-promise-reject-errors. The fixed expected map in shared jsxSources(),
  stage1/cohere/lint/jsx_integration_test.go:61, omits both rules and fails
  TestJsxLintTrees and TestJsxLintReleaseAndThroughput. This unit preserves that
  check; shared test-map upkeep is an integration change.
