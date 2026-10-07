Ported numeric JSX value-attribute extraction for the reserved constructed-context rule.
Commits: builds on pushed aae00cd92; this report and source are committed together on codex/typeaware-wave-04.
Checks: references 64 nodes/302 records/3801 bytes and reporting 56 records/4941 bytes match Go across five backends.
Mutants: eight reference mutants and six reporting/refusal mutants caught; all byte mutants compile/run cleanly under sanitizers.
Not covered: full JSX source rules, corpus findings/fixes/suggestions, source-rule throughput or the full repository gate.

## Attribute extraction

`JsxNoConstructedContextValues.valueExpression` consumes the supplied numeric attributes node. Snapshot arguments_ records ordered properties, name records the name child, and receiver records an initializer/expression child. It matches Go's production `jsxNoConstructedContextValuesValueExpression`: skip spreads and other names; accept only an identifier named `value` whose initializer is a JSX expression; return that expression's child, including -1 for an empty container. The first matching attribute decides, including an invalid boolean/string initializer; it does not search later duplicate attributes. No checker verdict, regex or source position conversion is introduced.

The Go overlay creates actual factory ASTs, invokes the production helper and translates returned pointer identity to a node index. Twelve attribute-list controls include empty properties, valid/other names, boolean shorthand, string literals, empty expression, spreads and duplicate-order cases. Existing structural/tag predicates run over the expanded graph too.

## Commands and evidence

Required setup: `bash cloud/setup.sh > /workspace/typeaware-wave-04-setup-current.log 2>&1`, then `source /workspace/adamic-tools/env.sh`; tool readiness lines each 0s, build cache warm 111s, total 111s. `nproc`: 5; cpu.max 400000 100000, memory 17.6 GB. Setup succeeded.

`python3 stage1/cohere/typeaware/wave_04_jsx/validate_references.py /workspace/typeaware-wave-04-attributes-final --adamic /workspace/typeaware-wave-04-harness-ab70/adamic > /workspace/typeaware-wave-04-attributes-final.log 2>&1`: PASS 64 numeric AST nodes, 302 records, 3801 bytes; actual Go production helpers/native/ASan-UBSan-LeakSanitizer/source Node/emitted JS. Eight compiled sanitized mutants exit zero with empty stderr and differ only in comparison: React receiver renamed Other, react module renamed preact, template kind 14 changed to 8, lowercase member root declined, receiver wrapper kind 218 changed to 8, value name changed to Value, expression kind 295 changed to 8, and invalid first value changed from return to continue.

`python3 stage1/cohere/typeaware/wave_04_jsx/validate_partial.py /workspace/typeaware-wave-04-attributes-reporting --adamic /workspace/typeaware-wave-04-harness-ab70/adamic > /workspace/typeaware-wave-04-attributes-reporting.log 2>&1`: PASS 56 records, 4941 bytes across the same five backends. Three clean byte mutants (fragment end, component lowercase bound, object label) and three removed-source-refusal mutants caught.

Exact streams, controls, source hashes and logs are retained in evidence/attributes. The first reference run had seven mutants; final evidence reruns the final validator with all eight. No full-rule timing claim follows from helper measurements. The previous six completed-rule corpus/sanitizer/mutant evidence remains at HARNESS_AB70.md: main c01907a7 is unchanged and an ancestor, so no new landing rebase is necessary.

## Remaining work and stop point

Shared numeric registry compatibility remains blocked: actual registry.Descriptor expects []string, whereas required rule.json kinds are numeric. No shared files were edited. Attribute extraction now supersedes one unfinished constructed-context helper; factories/adapters, provider symbol resolution, component ancestry, recursive construction, memo stability/escape analysis and final spans remain unfinished in this rule. Fragment aliases/options/attributes and undef declaration-file/binding decisions remain unfinished too. All source entry points still refuse explicitly. HIR-dependent React claims remain parked. No new claims were taken, and no complete JSX rule parity is claimed.
