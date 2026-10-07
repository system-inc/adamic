# Variant registry helpers

One helper per .a file: newVariantRegistry returns independent, nonnil empty maps and zero ordering state; register inserts a new name at groupOrder or lastOrder+1, advances lastOrder only outside a group, and updates an existing kind without moving its order; attachComparison stores/replaces a nonnil callback at the exact order and leaves the map untouched for nil (represented by undefined).

registry.a is the shared data model, not a fourth helper. Registration values are copied on update, matching Go map value semantics. The nullable group pointer is represented by groupPresent plus groupOrder; groupOrder is irrelevant when absent. Comparison arguments are opaque numeric handles for ParsedVariant values: attachment never inspects or invokes those values. The caller's AST adapter supplies callback closures accepting handles; complete VariantRegistry.Compare and variant parsing remain separate dependencies. This package does not pretend to port them.

The contract requires valid UTF-8-derived strings, initialized registries and safe-integer orders including each computed lastOrder+1. Go's full int64 range/overflow and arbitrary malformed adapter state are outside this bounded representation. Registry mutation is sequential, matching Go's documented lack of concurrent mutation safety.

Tests capture 157 actual runtime inputs from all six consuming rules, derive names from those sources and lexical tokens, and include empty/control/supplementary text, prototype-like names, negative/zero/framework/high orders, present/absent groups, arbitrary kind strings, duplicate registration, distinct names, nil/non-nil/replaced callbacks and constructor independence. An oracle-only Go overlay exposes private snapshots and group state while using the unmodified real constructor and public methods. The same .a source runs on Node, emitted JavaScript and sanitized native. Eight temporary compiling semantic mutants must differ from Go.

From the repository root, source /workspace/adamic-tools/env.sh and run go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m, redirecting output directly to a log. Regeneration is python3 stage1/cohere/lint/helpers/slot03/batch7/testdata/regenerate.py. The separate upstream capture gate fails its installed-Tailwind/corpus guards; capture still records every consumer before those guards. This is helper parity, not full rule finding/fix/suggestion integration.

See REPORT.md and readiness.json for exact consumers, mutant witnesses and bounded gate evidence.
