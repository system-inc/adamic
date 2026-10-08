Main 74fb6490 latent full evidence

Measurement-only overlay. Production compiler files are unchanged. No backend was invoked.

| Scope | Legacy NotYet | Full NotYet | Legacy Refused | Full Refused | Recovery boundaries |
| --- | ---: | ---: | ---: | ---: | ---: |
| Compiler files (79) | 1468 | 15284 | 5162 | 12272 | 21433 |
| tsc entry reach (81) | checker blocked | 15274 | checker blocked | 12243 | 21391 |

Legacy skips 51 top-level bodies: 4,376,342 / 7,101,403 bytes (61.63%).

| Giant body | Bytes | Full NotYet | Full Refused | Actual-lowering NotYet | Actual-lowering Refused | Boundaries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| createTypeChecker (src/compiler/checker.ts:1486:1) | 3,097,534 | 3849 | 5346 | 3849 | 1145 | 5784 |
| createNodeFactory (src/compiler/factory/nodeFactory.ts:492:1) | 309,525 | 1776 | 460 | 1776 | 51 | 2159 |
| createPrinter (src/compiler/emitter.ts:1211:1) | 223,291 | 694 | 513 | 694 | 293 | 1273 |
| transformES2015 (src/compiler/transformers/es2015.ts:488:1) | 215,187 | 479 | 244 | 479 | 83 | 711 |
| createProgram (src/compiler/program.ts:1515:1) | 184,892 | 264 | 140 | 263 | 32 | 346 |
| createBinder (src/compiler/binder.ts:509:1) | 171,392 | 331 | 426 | 331 | 108 | 481 |
| transformClassFields (src/compiler/transformers/classFields.ts:354:1) | 144,201 | 313 | 142 | 313 | 38 | 462 |
| createScanner (src/compiler/scanner.ts:1023:1) | 140,104 | 263 | 80 | 263 | 36 | 370 |

compiler: measured on a checker-rejected program; 324 checker diagnostics.
8912 function declarations; unit statuses: {'attempted': 10544, 'split_checker_body': 147}.
Nested checker exclusions: 110; 437,999 gross body bytes; 414,509 union bytes.
Actual-lowering-only totals: {'NotYet': 15274, 'Refused': 2774, 'SkippedDependency': 3, 'error': 2, 'panic': 23, 'Boundary': 21433}.
Context-sensitive reading-X sites: 4592; retained in the totals, flagged in reasons.csv. These can reflect incomplete isolated bindings after an earlier failure and need individual interpretation.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| a cast the runtime can't check | 01a11410 | 0 | 3645 | 3645 |
| a function inside a function (a closure) | OWNER BLANK | 3110 | 0 | 3110 |
| an object refinement using an open numeric enum as a literal tag | OWNER BLANK | 0 | 2776 | 2776 |
| the non-null assertion ! | 01a1130a | 0 | 1754 | 1754 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 1589 | 0 | 1589 |
| reading node | OWNER BLANK | 1043 | 0 | 1043 |
| a number as a condition | OWNER BLANK | 0 | 647 | 647 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 630 | 0 | 630 |
| a NonNullExpression | 01a1130a | 553 | 0 | 553 |
| reading Debug | 01a113e7-bfed | 509 | 0 | 509 |

Complete reasons, with every unowned row flagged, are in compiler/reasons.csv; excluded names and diagnostics are in compiler/excluded-nested.csv; all recovery boundaries are in compiler/boundaries.csv.

tsc: measured on a checker-rejected entry-root program; 324 checker diagnostics.
8912 function declarations; unit statuses: {'attempted': 10551, 'split_checker_body': 147}.
Nested checker exclusions: 110; 437,999 gross body bytes; 414,509 union bytes.
Actual-lowering-only totals: {'NotYet': 15264, 'Refused': 2784, 'SkippedDependency': 3, 'error': 2, 'panic': 25, 'Boundary': 21391}.
Context-sensitive reading-X sites: 4593; retained in the totals, flagged in reasons.csv. These can reflect incomplete isolated bindings after an earlier failure and need individual interpretation.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: | ---: |
| a cast the runtime can't check | 01a11410 | 0 | 3648 | 3648 |
| a function inside a function (a closure) | OWNER BLANK | 3109 | 0 | 3109 |
| an object refinement using an open numeric enum as a literal tag | OWNER BLANK | 0 | 2776 | 2776 |
| the non-null assertion ! | 01a1130a | 0 | 1754 | 1754 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 1596 | 0 | 1596 |
| reading node | OWNER BLANK | 1046 | 0 | 1046 |
| a number as a condition | OWNER BLANK | 0 | 643 | 643 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 627 | 0 | 627 |
| a NonNullExpression | 01a1130a | 555 | 0 | 555 |
| reading Debug | 01a113e7-bfed | 511 | 0 | 511 |

Complete reasons, with every unowned row flagged, are in tsc/reasons.csv; excluded names and diagnostics are in tsc/excluded-nested.csv; all recovery boundaries are in tsc/boundaries.csv.

Limits: a failed compound statement can prevent examination of its children. Signature/prologue failures can prevent examination of the remaining body. Nested declarations are still independently attempted. Boundary spans identify affected constructs and may overlap; they are not an exact count of unexamined bytes. Gross excluded-body bytes may overlap; union bytes avoid double counting. Isolated ancestor bindings and sibling signatures are observational context, not proof of closure support. Giant counts include signature sites and nested bodies within the giant declaration. Findings combine refusal scanning and actual lowering unless explicitly marked otherwise. Final ownership, module order, and backends were not tested.
