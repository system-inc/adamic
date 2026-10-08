# one-var wave2 landing

The landing branch starts from lint-batch/wave2-02 at 95968dd93 and merges
area/stage1-lint at 334509eea. No rebase was used. The only owned changes are
inside this rule directory. The existing one-var implementation from
codex/lint-port-one-var is carried unchanged: named VariableDeclarationList
listeners receive ParseNode, options decode to upstream OneVarSettings,
messages remain verbatim, and every automatic edit uses reportRange's edits
array. The complete TestOneVar prefix captures all sixteen upstream functions
and 340 unique source/file/options cases at the current cohere pin.

## Validation

Registry generation, gofmt -l on oracle.go and go vet ./stage1/cohere/lint/...
pass with empty formatting/vet output. Setup completes in 200.266 seconds on
5 processors. The TypeScript input is pinned to 6.0.3 commit
050880ce59e30b356b686bd3144efe24f875ebc8.

The entire lint package runs once with every optional input set, including
compiler input, throughput, profile artifacts and profile snapshots. It exits
0: 41 top-level PASS, zero FAIL, one SKIP; including subtests, 147 PASS, zero
FAIL, one SKIP. The skip is the inherited TestCheckerBridgeRefusalPending,
which awaits codex/tsgo-errors-as-values. It was neither removed nor relaxed.

TestRulesAgree passes in 312.66 seconds: 5,545 captured source/rule/options
combinations and 14,346,558 identical bytes across Go, original Node, emitted
JavaScript and sanitized native. TestOwnedWitnesses passes in 13.18 seconds:
470,625 identical bytes, including the multi-edit witness. Compiler/stage1
parity, shards and profile snapshots also pass. Compiler/stage1 compares
36,136,622 identical bytes.

All 94 registered mutant gates pass. The shared syntax gate checks Node and
emitted JavaScript and uses one sanitized native canary. The owned
one-var-combine-message mutant additionally runs through testdata/native_mutant.py,
which appends a direct sanitized-native/Go comparison using a temporary Go test
overlay. It changes no shared source and removes no assertion. The owned mutant
is caught on all three sides at case 1 line 19, with normal execution and no
sanitizer diagnostics. This additional check passes in 140.592 seconds.

The complete compiler benchmark has 77 files and 35,069 identical findings.
Best of five: Go 1.784134 seconds, native 10.726646 seconds, source Node
5.897897 seconds. Native is 6.012 times Go in this whole-package benchmark;
this is not an isolated one-var timing.

Inputs, exact source hashes, case counts, full JSON test log, native mutant log,
setup, registry, formatting, vet and independent upstream Go logs are in
validation/unpark/. No full repository gate or release of the retained checker,
IR or control-flow reservations is claimed.

## Retained wave-21 rules not registered on this branch

All four independent upstream Go packages pass. Counts below are captured Go
cases, not native certifications. The available generic checker answers do not
expose the required reference-symbol declaration identities/trees; symbol-origin
returns only the ValueDeclaration filename and declarations accepts only classes
and interfaces. Type-origin exposes declaration file metadata, not syntax
ancestry or binding identity. No private bridge or helper substitute was added.

| Rule | Upstream cases | Blocker | Exact Go call and source |
| --- | ---: | --- | --- |
| @typescript-eslint/no-mixed-enums | 81 | checker declarations and enum local symbols | `ctx.TypeChecker.GetSymbolAtLocation(name); node.LocalSymbol()` at `cohere/internal/lint/rules/typescript/no_mixed_enums.go:259` |
| nexus/correctness-no-collection-misuse | 6 | checker declaration/default-library facts | `ctx.TypeChecker.GetSymbolAtLocation(access.Name())` at `cohere/internal/lint/rules/nexus/correctness_no_collection_misuse.go:430` |
| nexus/correctness-no-discarded-outcome | 31 | checker type-symbol declaration trees and union ancestry | `checker.Type_symbol(constituent)` at `cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome.go:173` |
| nexus/correctness-no-discarded-pure-result | 22 | checker method declarations, interface owners and library facts | `ctx.TypeChecker.GetSymbolAtLocation(name)` at `cohere/internal/lint/rules/nexus/correctness_no_discarded_pure_result.go:153` |
| nexus/correctness-no-process-exit-after-output | 56 | control-flow graph | `control_flow_graph.Build(root.node, control_flow_graph.Hooks[correctnessNoProcessExitAfterOutputEvent]{...})` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:511` |
| nexus/correctness-no-uncleared-race-timeout | 21 | checker declarations and shorthand binding identity | `ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent)` at `cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:337` |
| nexus/correctness-require-blocking-standard-streams | 48 | control-flow graph and cross-file symbol facts | `control_flow_graph.Build(root, control_flow_graph.Hooks[event]{...})` at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:840` |
| no-obj-calls | 181 | checker binding/reference identities, including shorthand; reference tracker not supplied | `typeChecker.GetShorthandAssignmentValueSymbol(parent)` at `cohere/internal/lint/ecmascript/reference/value.go:128` |
| no-object-constructor | 59 | checker first symbol declaration and IsDeclarationFile | `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103` |
| no-promise-executor-return | 135 | checker first symbol declaration and IsDeclarationFile | `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103` |
| react-hooks/set-state-in-effect | 53 | React Compiler IR and graph transforms | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_effect.go:267` |
| react-hooks/set-state-in-render | 50 | React Compiler IR, SSA and capture translation | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_render.go:183` |
| react-hooks/static-components | 22 | React Compiler IR, SSA and phi data | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/static_components.go:120` |
| react/jsx-fragments | 46 | checker import/binding/variable symbol declaration trees | `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_fragments.go:290` |
| react/jsx-no-constructed-context-values | 191 | checker ordered declarations, binding identities and stability facts | `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:125` |
| react/jsx-no-undef | 45 | checker symbol presence and every declaration source identity | `ctx.TypeChecker.GetSymbolAtLocation(reference)` at `cohere/internal/lint/rules/react/jsx_no_undef.go:101` |
| react/style-prop-object | 76 | checker shorthand symbol and first ordered variable declaration | `ctx.TypeChecker.GetShorthandAssignmentValueSymbol(shorthand)` at `cohere/internal/lint/rules/react/style_prop_object.go:205` |
