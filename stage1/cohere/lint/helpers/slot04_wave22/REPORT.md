# Slot 04 wave 22 report

Built three .a helpers: UtilityEvaluator.Compile, UtilityEvaluator.compile and react.DecodeCompilerRuleOptions. This slot retains 65 completed helpers through this wave. Claim 0e4e3f70d0a395190f3e9c689e589d2f17538540 was pushed before implementation. No new rule, shared harness, registration generator or compiler file was edited.

The branch already contained origin/main c7991b900362796aefd111474e65eb5398e91953 and origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da at landing commit 8d222fb5. All 22 retained helper packages passed that landing check. Both protected bases stayed unchanged through this wave. Incoming harness, finding model, compiler and allocator changes are retained. Only codex/lint-helpers-04 is pushed.

## Consumer prerequisites

All 20 origin codex/lint-helpers* branches and every claims file were inspected. The original shared branch already delivered the comment bundle. The three selected symbols tied at four consumers, the largest remaining unclaimed fan-out. The snapshot is evidence/claims.json.

Each utility compiler helper removes one prerequisite from these four rules:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

The React decoder removes one prerequisite from these four rules:

- react-hooks/config
- react-hooks/error-boundaries
- react-hooks/incompatible-library
- react-hooks/void-use-memo

Thus this wave removes 12 listed dependency edges across eight rules. Zero rules lose their final listed helper blocker from these three alone. readiness.json subtracts these symbols only, without crediting unrelated worker implementations. This is conditional helper readiness, not eight native rule ports or findings parity.

## Observations and commands

Every test wrote directly to a log. Source environment: /workspace/adamic-tools/env.sh. Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. nproc printed 5. Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 223s; total 223s. Raw setup output is evidence/setup.log.

- python3 stage1/cohere/lint/helpers/slot04_wave22/testdata/regenerate.py: 237 controls. Regeneration reproduced witnesses.json byte for byte against coverage.json's SHA-256.
- python3 stage1/cohere/lint/helpers/slot04_wave22/testdata/capture.py: 297 unique asserted fixture inputs from all eight consumers. Selected actual-Go Tailwind and React suites PASS. Two distinct real UtilityEvaluator.Compile calls, for synthetic-fn-small and synthetic-fn-large, were captured with their candidate, normalized definition and theme state. No live decoder call occurred in those selected suites. All four React consumers' captured decoded option objects were {}.
- go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave22: PASS, 39.949s, evidence/helpers-final.log. Actual Go versus source Node, emitted JavaScript and sanitized native: 1186 control output lines, 10 live-call lines and 1485 consumer-option lines per mode. Every process exits successfully with empty stderr. The native build enables ASan/UBSan and Linux leak checking. Captured consumer-option records invoke the real Go decoder on re-encoded options; Tailwind source texts are retained as fixture provenance, not run as native rule programs.
- go vet ./stage1/cohere/lint/helpers/slot04_wave22: PASS, empty evidence/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS, 1.536s, evidence/oracle.log. Node, emitted JavaScript, release native and sanitized native agree on 758 bytes. Native misses 3, Node misses 2, cache hits 0.

Initial and intermediate logs are preserved separately. No selected check skipped. The full repository gate and the 17 broader TypeScript/postcss/graphql/parser comparisons were not selected or claimed; their required inputs were not bypassed, their checks were not relaxed, and their files were not edited.

## Mutants

All 13 mutants compile, exit 0 and have empty stderr in source Node, emitted JavaScript and sanitized native. Ordinary byte-for-byte comparison with actual Go catches each. A compile failure, panic or sanitizer error is not counted.

1. Accept rejected nil/non-functional candidates.
2. Accept a missing utility definition.
3. Reject empty raw compiler options.
4. Public Compile uses mode 1, retaining dropped declarations.
5. Omit the used/resolved-value postcondition.
6. Omit the reached-but-unresolved modifier postcondition.
7. Omit the ratio/modifier conflict postcondition.
8. Omit the unconsumed written-modifier postcondition.
9. Omit dropped-declaration removal.
10. Omit ratio-splice removal.
11. Accept JSON null as an options object.
12. Accept non-empty compiler configuration objects.
13. Preserve reverse input key order instead of Go sorted order.

Eight additional coverage mutants remove all fixtures for one consumer each. The readiness-derived coverage check rejects every omission. Production source is never mutated; temporary copies are compiled for these checks.

## Limits and inferences

The utility implementation owns orchestration, postconditions and removal-set selection. Lookup, deep clone, value/modifier-aware walk and CSS node removal are explicit dependencies. The driver supplies actual-Go observations for them and compares canonical serialized Go node structure. It does not provide a native walker/remover or whole findings/fix/suggestion execution. Generic rejection returns an empty node projection with ok false; callers must preserve Go nil semantics at their adapter boundary. Theme and candidate payloads are captured by the walk callback.

The decoder owns acceptance gates and exact refusal prose, with fresh empty options. encoding/json-compatible decoding and Go byte-key sorting are explicit dependencies. Valid UTF-8 configurations, duplicate keys, replacement of an escaped lone surrogate, malformed JSON, non-objects, whitespace, newline keys and non-BMP versus BMP ordering are covered. Arbitrary invalid raw byte input is outside the string API and requires a byte adapter before integration. Original configuration file bytes, all configurations, raw Go decoded-struct construction and the full option parser are not claimed.

237 controls and two live utility calls are bounded evidence. The 297 Go consumer fixtures are asserted upstream outcomes and option provenance; they are not 297 live utility compilations. No regex helper or matcher is ported, and no finding position or byte-to-UTF-16 conversion is built here. The dependency removals are an inventory inference conditional on the stated callback contracts and common AST adapter.
