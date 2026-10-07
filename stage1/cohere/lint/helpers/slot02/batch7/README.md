# Slot 02 seventh helper batch

| File | Contract | Remaining consumers |
|---|---|---:|
| call_expression_source.a | Go imports.CallExpressionSource: dynamic import/import.defer or single-argument require, literal-like first argument, decoded source and separate presence. | 5 |
| match_ignoring_case.a | Go jsx.MatchIgnoringCase: rune-by-rune Unicode simple folding without multi-character expansion. | 4 |
| has_attribute_named.a | Go jsx.HasAttributeNamed: identifier-named attributes only, ordered matching, first-hit termination and presence independent of initializer. | 4 |

RULES.md lists all thirteen consuming rules. readiness.json removes thirteen dependency entries. CallExpressionSource removes the final listed helper blocker for nexus/import-no-forbidden-source; the other twelve rules retain blockers. These are helper ports; no whole rule is implemented by this batch.

CallExpressionSource reads a dense, read-only projection of actual Go parser nodes. Arena index -1 means nil. Preserve exact node kinds, decoded identifier/literal text, callee edge, MetaProperty keyword/name and argument list presence/order. import.defer is accepted, import.source and ordinary member calls are declined. Import may have additional arguments; require must have exactly one. A no-substitution template is string-literal-like; an interpolated template is declined. Empty sources remain present. Nil nodes decline. Invalid arena edges refuse through panic instead of producing a clean result. Arbitrary malformed manually constructed require calls with nil argument lists panic in actual Go and are outside this parser-node projection.

HasAttributeNamed reuses this slot's AttributeName helper. Preserve attributes node kind, properties-list presence, ordered property edges and identifier name text. Nil property and name edges decline, as do spreads and namespaced names. Empty identifier names remain named. The supplied match callback gets candidate then wanted without normalization, only for named attributes. Callback errors/side effects follow ordinary invocation; no extra match is invoked after a true result. Attribute initializer values do not affect presence.

MatchIgnoringCase uses a sorted generated table of noncanonical Unicode simple-fold mappings, with binary search and explicit UTF-16 surrogate-pair decoding. Go Unicode 17.0.0 supplies all 1,512 mappings. Kelvin K, long s, sigma cycles and supplementary Deseret letters agree with strings.EqualFold. Sharp s versus SS, ligature expansion and canonically equivalent but differently encoded strings remain unequal. Table/version drift is checked against Go over all Unicode scalar values. Text supplied by the AST/JSON adapter must preserve valid decoded Unicode; arbitrary invalid UTF-8 byte strings are outside this projection. JSON normalization of unpaired surrogates is not claimed as an independent parser port.

The Go oracle directly calls the actual exported helpers against pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. No replacement helper body is used. The capture overlay adds observations at testing entry points in temporary files; it never edits the cohere worktree or the shared Adamic harness. Actual Go Next, Nexus and React rule packages must pass to regenerate fixtures. Every consumer must be observed. All outputs go directly to logs.

The corpus contains 578 captured rule/file/source inputs, 522 distinct parsed sources including controls, 10,571 nodes, 13,880 case-fold/name pairs and 623 attribute targets with twenty-one requested names and four matcher modes. Ordinary and factory controls cover nil nodes/lists/names/properties, empty identifiers, duplicates, optional calls, callee parentheses, extra arguments, import.defer, template holes, Unicode and decoded NUL/newline specifiers. Matching traces cover exact, case-insensitive, always-false and stateful second-call matching. Expected Want is removed before Adamic reads the corpus. Source strings compare as decimal UTF-16 units to preserve decoded text, including supplementary characters, through all backends.

The baseline and each mutant must compile, exit successfully without stderr and agree across source Node, emitted JavaScript and ASan/UBSan native. Only a subsequent mismatch against real Go counts as a semantic mutant caught. Consumer capture and folding artifacts reproduce byte for byte.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch7/testdata/regenerate.py > /tmp/slot02-batch7-capture.log 2>&1
python3 stage1/cohere/lint/helpers/slot02/batch7/testdata/regenerate_folds.py > /tmp/slot02-batch7-folds.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/slot02-batch7-helpers.log 2>&1
```

Bounded helper parity does not establish integrated findings, fixes, suggestions or arbitrary AST-adapter correctness. No full repository gate is claimed. See REPORT.md and evidence/ for observed commands, mutants, ownership and limitations.
