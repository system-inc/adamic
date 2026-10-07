# Slot 04 wave 19 report

Built three utility evaluation helpers, one per Adamic file: walk, resolveDeclaration and resolveValueFunctions. All 53 prior helpers were complete, green and pushed before this reservation; this wave raises the retained total to 56. Claim 3b65415d was pushed before code. Implementation SHA is in the final handoff.

The branch remains based on lint area d65a8f931c98655936ae04c6899f38f14862b73e, which contains current main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Final fetch confirms both are unchanged and ancestors of HEAD. No main or area branch was pushed. Incoming finding-model, emitted-JavaScript harness and allocator changes remain intact.

All twenty origin codex/lint-helpers* branches were refreshed and every claim file checked. Selected helpers were unclaimed and tied the highest available fan-out at four consumers each. The inherited five-function comment bundle remains owned even though its qualified names are absent from the textual claims. evidence/claims.json pins the inspected tips and claim paths; the own branch's pre-reservation tip is recorded there.

## Contracts and dependency removals

walk descends through CSS rule, at-rule, context and at-root containers and visits declarations in order. resolveDeclaration skips missing/empty values, parses a value, marks original declaration identity on failure, and rewrites printed text only on success. resolveValueFunctions retains five evaluation flags, marks non-ratio declarations by identity, routes value versus modifier resolution, stops at first failure and replaces successful calls without visiting substituted text. Maps preserve allocated versus absent state and existing false entries. README.md describes each callback and caller contract.

Each of the three helpers removes one listed prerequisite from each of these rules, twelve rule/helper dependency edges total:

| Rule | New prerequisites removed | Other frozen prerequisites remain |
|---|---:|---:|
| better-tailwindcss/enforce-consistent-class-order | 3 | 261 |
| better-tailwindcss/enforce-shorthand-classes | 3 | 198 |
| better-tailwindcss/no-conflicting-classes | 3 | 218 |
| better-tailwindcss/no-unknown-classes | 3 | 201 |

This table subtracts only these three from the shared frozen inventory; it does not reconcile independently delivered work on other branches or previous slot waves. No rule reaches zero blockers from this wave alone. Readiness is conditional on the independently owned ParseValue, ValueToCss, resolveValueFunction, modifierAsValue and CSS/value-node adapter dependencies. It does not mean four completed rules or native finding parity.

## Observations

The private Go overlay exports and calls the actual pinned cohere methods; no Go method is reimplemented. Go supplies parser trees and the per-function argument-resolution results used by the Adamic callbacks. Claimed traversal, flags, map updates, stopping and rewriting run in Adamic, independently compared with Go. Modes exercise a composed walk, separate declaration calls and separate value-function calls.

There are 960 deterministic controls, 112 unique asserted fixtures from all four Go consumer rule suites, and six distinct actual helper entry inputs: two walk, two resolveDeclaration, two resolveValueFunctions. Actual calls use --value(--frobnicate-*,[length]) with small and large candidates. Go captures all three helper inputs before their methods execute, including theme/value/modifier inputs. They are replayed with fresh evaluation state; synthetic controls separately cover all 32 initial flag combinations and allocated empty/false-entry maps.

Consumer fixture source text is additionally replayed as a declaration-value parser control, not as a native rule finding. Captured actual helper inputs are the direct evidence of consumer reachability. Controls include all four container kinds, ignored comment/unknown subtrees, distinct identical declarations, absent and empty values, malformed parser input, nested ordinary functions, ordered failures, defaults, literal/arbitrary/number/ratio/percentage/theme resolution and substituted text that itself contains --value. Go, source Node, emitted JavaScript and sanitized native agree on 37,440 control output lines, 963 consumer-source lines and 36 captured-call lines; newlines inside raw source text account for extra physical lines.

## Commands and outputs

All runs wrote directly to evidence logs. Environment Go 1.27.1, clang 20.1.8, Node 24.19.0. Source /workspace/adamic-tools/env.sh after cloud/setup.sh. Setup timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (267s)
setup: done in 267s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc prints 5. Setup warm builds every package and test binary but runs no repository tests; this is not a full gate.

- python3 stage1/cohere/lint/helpers/slot04_wave19/testdata/capture.py: 112 inputs from four consumers, six distinct helper inputs from six entry calls. Actual Go tailwind suite PASS, 0.580s. Tailwind 4.3.3 fixture install and Go instrumentation use temporary directories and overlays.
- python3 stage1/cohere/lint/helpers/slot04_wave19/testdata/regenerate.py: 960 controls, identical SHA256 before/after. Parser/resolution observations are generated by actual Go at test time.
- go test -count=1 -v -timeout=10m ./stage1/cohere/lint/helpers/slot04_wave19 -run '^TestUtilityTraversalGoNodeNativeJavaScript$': PASS, 14.351s. First parity run.
- go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave19: PASS, 333.506s: all parity comparisons, twelve compiling semantic mutants in all three modes, and four consumer omissions. Mutant phase 273.76s; full parity phase 59.70s during setup warm-up.
- go vet ./stage1/cohere/lint/helpers/slot04_wave19: exit 0, empty vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS, 9.816s; source Node, emitted JavaScript, release and sanitized native agree on 758 bytes, three native misses, two Node misses, zero cache hits.

Native helper runs require exit zero and empty stderr under ASan/UBSan and Linux leak checking. Temporary semantic mutants must also compile and run cleanly before comparisons count them as caught. Scoped .gitattributes preserves literal whitespace in raw logs.

## Mutants

Each semantic mutant runs on all 960 controls and is compared against actual Go in source Node, sanitized native and emitted JavaScript. No production file is mutated.

| File | Mutation | What catches it |
|---|---|---|
| walk.a | skip declarations | original declaration output/map state differs from Go |
| walk.a | skip context descent | nested declaration output/identity differs from Go |
| walk.a | skip at-rule descent | nested declaration output/identity differs from Go |
| resolve_declaration.a | ignore value presence | absent declarations resolve or drop unlike Go |
| resolve_declaration.a | mark a different identity | the failed declaration lacks Go's removal membership |
| resolve_declaration.a | retain unprinted original text | malformed parser controls differ from Go normalization |
| resolve_value_functions.a | clear usedValue | value-function state differs from Go |
| resolve_value_functions.a | clear resolvedRatio | ratio bookkeeping differs from Go |
| resolve_value_functions.a | store false for a non-ratio declaration | identity membership value differs from Go |
| resolve_value_functions.a | clear usedModifier | modifier-function state differs from Go |
| resolve_value_functions.a | substitute the function name | rewritten text differs from Go's resolved text |
| resolve_value_functions.a | continue after failed value resolution | later state/rewrites and first-failure result differ from Go |

All twelve semantic mutants compiled, ran with exit zero and no stderr, and were caught by Go output comparison in all three modes (36 successful semantic detections). Four additional structural mutants omit each consumer's fixtures in turn; the readiness-derived coverage check must report the missing rule. These are coverage checks, not compiling semantic mutants.

## Limits

No full repository gate or new rule port was run. Prior nineteen helper packages, runtime checks and shared finding harness were green on this same base in slot04_landing_area_d65a8f93; they were not repeated without a base or code change. No shared compiler, harness, registry, readiness or rule file changed. No regex is ported in these three helpers.

Callback dependencies and unrelated CSS fields are not implemented here. Tests use acyclic CSS trees without shared pointers; identity distinctions for separate equal declarations are covered, arbitrary aliasing/cycles and nil nodes are not. Value AST projection compares rewritten text, resolution side effects and declared identities, not every serialized unused AST field. Whole rule diagnostics, ranges, fixes and suggestions remain rule-worker work.
