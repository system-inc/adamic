Built: IsExportedByName, HasExportModifier and IsExported, one .a file each.
Commits: claim 28d74b6 pushed before code; based on current main e8ba3d5, after green pushed tip 8688f52.
Checks: new helper package PASS 20.107s; six uncached Node oracle probes PASS 1.525s; vet and formatting clean.
Mutants: six compiling semantic mutants caught on source Node, sanitized native and emitted JavaScript; ten separate consumer omissions caught.
Not covered: full repository gate, rule integration/findings/fixes, checker/binding/export-list semantics or nil modifier elements outside the adapter boundary.

## Behavior

Named export requires an export modifier with no default modifier. General export
requires the export modifier, including default exports. The node wrapper checks
nil before reading its modifier list and delegates to the general predicate.
These distinctions preserve Go's behavior on order, repeated modifiers, unknown
kinds, nil nodes, nil lists and present empty lists. Classification uses AST
kinds, so a string or identifier spelling export is not an export modifier.
No shared compiler, harness or registration generator was changed.

All 18 origin codex/lint-helpers* branches were fetched before reservation and
again before delivery. Every greater-than-five concrete helper was reserved;
the original comment bundle remains owned by the base worker. These three tie
the highest remaining unclaimed concrete count, five each. The final claim
check finds only this slot's reservation for all three. Main remains e8ba3d5;
previously retained code, compiler base and oracle inputs are unchanged.

## Rules and dependency accounting

The batch removes 15 prerequisite entries across ten distinct rules.
HasExportModifier and IsExported together remove the final blockers for:

- @typescript-eslint/no-useless-empty-export
- nexus/consistency-no-screaming-snake-case

Those two predicates each also remove one prerequisite from these three:

- structure/network-require-hook-options-parameter
- structure/network-require-hook-request-suffix
- structure/network-require-hook-variables-type

IsExportedByName removes one prerequisite from each of:

- @next/next/no-typos
- structure/boundary-no-project-theme-value
- structure/next-no-near-miss-route-export
- structure/react-component-no-const-assignment
- structure/react-component-require-named-export

Local readiness.json records every residual dependency. Subtracting this batch
alone from the frozen base moves helper-ready count from 46 to 48. This is
helper readiness conditional on the original common AST adapter assumption;
it does not claim two complete rule ports, refresh stage1 status, or credit
other workers' dependency removals. Prior slot reports retain their accounting.

## Observed validation

With /workspace/adamic-tools/env.sh sourced, test output went directly to logs.
Earlier setup in this session passed: Go 0s, clang 1s, Node 1s, submodules 2s,
build cache warm 165s, total 165s; nproc 5. Setup was not repeated.

```
python3 stage1/cohere/lint/helpers/slot04_wave10/testdata/capture.py
go test ./stage1/cohere/lint/helpers/slot04_wave10 -count=1 -v -timeout=15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m
go vet ./stage1/cohere/lint/helpers/slot04_wave10/...
gofmt -l stage1/cohere/lint/helpers/slot04_wave10
git diff --check
```

The real Go next, nexus, typescript and structure rule suites passed with a
temporary capture overlay. It recorded 424 unique asserted source fixtures
covering all ten consumers. The parser oracle observes every AST node in those
fixtures, yielding 8,827 rows. It does not claim every observed node was passed
to a production predicate by a rule. All actual caller-node shapes in those
source trees are included, alongside harmless additional nodes.

Controls yield 2,319 observations: all 2,186 combinations of length-zero-through-
six modifier sequences over export/default/other, crossed with list presence,
plus 133 parsed nodes from twenty declaration/module/syntax controls. The parser
controls include default and named functions/classes, anonymous and async
functions, namespaces, separate export lists, export assignment, type/interface/
enum exports, switch default, identifiers/comments/string text and duplicate
modifier recovery. Go decides their actual modifier lists and verdicts.

All rows match actual Go on source Node, sanitized native and emitted JavaScript.
The six filtered input probes bypassed cache, zero hits and six misses. Every
successful comparison process exits zero with no stderr. Native uses ASan/UBSan
and default Linux leak checking. Vet and format logs are empty. Evidence logs
are in evidence/. No superseded test failures occurred in this batch.

## Every mutant and what caught it

Each semantic mutant compiles, runs successfully on all three Adamic execution
paths, and is caught only by differing output against actual Go.

| File | Mutation | Independent witness |
| --- | --- | --- |
| is_exported_by_name.a | Replace export-and-not-default with export-or-default | Default-only or export-plus-default list wrongly counts as named |
| is_exported_by_name.a | Do not remember default | Export-plus-default list wrongly counts as named |
| has_export_modifier.a | Return true for nil list | Go nil-list predicate is false |
| has_export_modifier.a | Recognize default instead of export | Export-only/default-only lists diverge |
| is_exported.a | Return true for nil node | Go nil-node predicate is false |
| is_exported.a | Negate forwarded modifier result | Actual parsed export/nonexport node verdicts reverse |

The readiness-based coverage check also removes every fixture for each of the
ten consumers separately and catches all ten omissions. Mutants use temporary
copies; production files are never mutated. These logs prove the checks can fail.

## Limits and inference

Observed: the three predicates agree with Go on the parser-derived and bounded
synthetic domains above. Inference: their frozen helper dependencies can be
removed when consumers use this adapter. Whole-rule correctness still requires
rule-local construction, options, listeners, reports, fixes and suggestions.
The full repository gate was not run. Arbitrarily long synthetic lists, nil
modifier elements, raw invalid UTF-8, node payload corruption and checker binding
resolution are not tested. Source parsing observes fixtures as supplied; it
does not establish that malformed controls are valid TypeScript programs.
