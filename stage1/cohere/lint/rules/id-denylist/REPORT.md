# Wave 29 on shared checker facts

Branch: `lint-rules/facts-wave-29`. Base: `d845dccde413c89643293e808626344d12e3f023`. Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.

## Implemented

`id-denylist` uses the handed Identifier/PrivateIdentifier and the shared SymbolDetails decoder with `checker.ask(index, 'node-symbol-details')`. It preserves unresolved names versus present symbols with zero declarations, and the all-declaration-file exemption. Its oracle decodes both the upstream captured IdDenylistSettings object and the config's variadic name tuple. Messages are copied verbatim. No fixes or suggestions exist upstream.

`no-shadow-restricted-names` consumes shared `reference_writes.a` and `binding-declarations`, preserving the first same-file declaration anchor. A lazy file scan runs once only when a bare undefined declaration needs the carve-out; exact declaration byte spans distinguish a shadow from a write to the judged binding. Its oracle decodes NoShadowRestrictedNamesOptions through the upstream UnmarshalJSON implementation. Messages are copied verbatim. No fixes or suggestions exist upstream.

Both descriptors declare `programReads: []`, exactly the upstream rule declarations: neither core/id_denylist.go nor core/no_shadow_restricted_names.go declares ProgramReads. TypeReachShapes is not a ProgramRead declaration. These rules ask their handed/current-file nodes directly; they do not use declarationAnswer's ReadsOtherFiles wrapper, askFile, foreign source loading or the raw default-library bit. No bridge, harness or shared helper file is edited. No private checker is used.

## Remaining claims

Locations are relative to `cohere/internal/lint/` at the named pin. Shared analyses must be native helpers, never Go lint verdict questions.

| Rule | Stopped on exact call | Missing question/helper |
| --- | --- | --- |
| id-match | `regexp.Compile(patternText)`, `rules/core/id_match.go:108` | Required runtime `new RegExp(pattern, 'u')` is refused at `internal/lower/regexp.go:39`; no handwritten matcher |
| nexus/concurrency-no-check-then-write | `checker.SkipAlias(symbol, analysis.ctx.TypeChecker)`, `rules/nexus/concurrency_no_check_then_write.go:362` | Followed alias identity/declarations, e.g. alias-declarations; raw node identity is insufficient |
| no-restricted-globals | `reference.ReadSymbol(ctx.TypeChecker, node)`, `rules/core/no_restricted_globals.go:260`; `typeChecker.GetExportSpecifierLocalTargetSymbol(parent)`, `ecmascript/reference/value.go:130` | Actual export-specifier local target; read-symbol/reference-symbol-origins plus native IsValueReference helper |
| no-setter-return | `descriptor.IsFunctionUnder(node, "set", isGlobal)`, `rules/core/no_setter_return.go:93,109` | Native shared descriptor-position analysis; not a checker verdict |
| react/jsx-fragments | `jsx.ElementParts(opening)`, `rules/react/jsx_fragments.go:207` | Native shared JSX tag/attribute decomposition; not supplied by the facts slice |
| react/jsx-no-constructed-context-values | `walk.ctx.TypeChecker.GetResolvedSignature(call)`, `rules/react/jsx_no_constructed_context_values_stability.go:488` | Resolved signature declaration metadata for stability analysis; native memo/escape/capture helpers also remain |
| react-hooks/set-state-in-effect | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)`, `rules/react/set_state_in_effect.go:267` | Shared source-to-HIR/SSA/capture lowering with memo erasure |
| react-hooks/set-state-in-render | `high_level_intermediate_representation.ForFunction(ctx, functionNode)`, `rules/react/set_state_in_render.go:183` | Shared source-to-HIR/SSA/capture lowering |
| react-hooks/static-components | `high_level_intermediate_representation.ForFunction(ctx, functionNode)`, `rules/react/static_components.go:120` | Shared source-to-HIR/SSA/capture lowering |
| react/jsx-no-undef | Already ported by wave 1, per Ahra | Not duplicated or certified again by this branch |

Dependency reproducers are retained in the previously pushed audit `lint-checker/audit-wave-29`, `stage1/cohere/lint/checker-audit/wave-29.md` at `aaf52ab7ff55f88c06ff2d0e4e4caa3dffd7dddd`.

## Capture limitation

NoShadowRestrictedNamesOptions.AllowGlobalThis has `json:"-"` at `rules/core/no_shadow_restricted_names.go:23`. The shared capture calls `json.Marshal(result.capture.options)` at `testing/docs_capture.go:62`, so AllowGlobalThis:true is serialized as `{}` and the replayed Go adapter restores the default false. `TestNoShadowRestrictedNamesRespectsAllowGlobalThis` passes the true field at `rules/core/no_shadow_restricted_names_test.go:169`. The shared harness currently cannot preserve those original option-bearing controls; a green comparison of the captured `{}` is not evidence of original option parity. The additional configured witness explicitly supplies `{"reportGlobalThis":false}` and checks that globalThis bindings remain silent while NaN still reports. This limitation is not hidden or repaired with a source-based option guess. No shared capture file is edited.

## Validation

Setup passed in 42.022s, nproc 5, four-core quota. [Setup log](evidence/wave29-facts-setup.log).
Registry generation and lint-package vet passed. [Registry](evidence/wave29-facts-registry.log), [vet](evidence/wave29-facts-vet.log).
The first witness build exposed the compiler's constructor early-return emission: clang rejected a bare return in Rule_new. Option decoding was moved to a void method in this rule; no compiler file changed. [Initial log](evidence/wave29-facts-witnesses.log).
All owned witnesses subsequently passed Go, source Node, emitted JavaScript and ASan/UBSan native, in 68.604s. [Witness log](evidence/wave29-facts-witnesses-retry.log).

The full package is run once with `-timeout=3h`, GOMAXPROCS=4, GOFLAGS=-buildvcs=false, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript at clean 050880ce59e30b356b686bd3144efe24f875ebc8, and both profile variables set to /workspace/wave29-facts-full-profiles. The runner saves output directly to a file and samples load. [Runner](evidence/run_gate.py), [package JSON events](evidence/lint-package.jsonl). Full package exit 0, wall 2111.719s (35m 11.719s), nproc 5. Test events including subtests: 120 pass, 0 fail, 1 skip; top-level tests: 34 pass, 0 fail, 1 skip. One-minute load min/median/max: 0.268/1.494/4.895. [Gate result](evidence/gate-result.json), [load samples](evidence/load.json).

The sole skip is TestCheckerBridgeRefusalPending: awaits codex/tsgo-errors-as-values, because tsgoInspect must return TSGoError from the C error buffer. All supplied input-driven comparisons, throughput and profile checks ran. This is not a zero-skip certification; the shared bridge test is not changed.

TestRulesAgree passed in 248.90s. It replayed 158 unique id-denylist cases and 69 no-shadow-restricted-names cases byte-identically against Go on Node, emitted JavaScript and sanitized native. The latter comprises 66 faithfully captured controls plus three original AllowGlobalThis controls whose options are lost by capture, as described above. They match their captured default configuration; original option parity remains uncertified. [Captured records](evidence/upstream-capture.jsonl).

TestOwnedWitnesses and TestMutants passed in the same full package. Both compiling mutants disagree with Go on native, Node and emitted JavaScript: written undefined carved out and denied names ignored. [Six caught verdicts, with full-log line references](evidence/mutants.log). The initial constructor build failure is not counted as a caught mutant.

Aggregated upstream runtime measurements, each including program creation: id-denylist Go 8.775s versus native lint and recording 11.606s (1.323x); no-shadow-restricted-names Go 3.685s versus native lint and recording 4.867s (1.321x). Counts and ordering are the typed records in TestRulesAgree; recording overhead is included on native. [Runtime measurements](evidence/runtime-times.json).

TestCompilerAndStage1Agree passed on 883 files from the pinned TypeScript compiler and stage1 repository corpus. TestThroughput, TestShardsAgree, profile artifacts, profile snapshots and all 82 registered mutants passed. Shared parser recovery refusals are explicit dependency limits in the package log, not successful recovered-findings parity. The shared checker refusal gate remains pending as named above.
