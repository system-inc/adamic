Evidence only: compiler includes unmerged enum fix 06e85073. Do not merge this branch to main.

Whole program: main: 2/79; area: 2/79
Own file: main: 55/79; area: 55/79

Compiler: e1198a5772ba4b6d7d322043292696ea11989044 (single mode).

| Tree | Baseline compiler | Current source SHA | NotYet before | NotYet now | Delta | Refused before | Refused now | Delta |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| main | 84d5eadf00b30a540b9aa4dd79ce8a690aae9136 | ce0750f28ef3943057f1f852b3ae5d93e6c5d644 | 1470 | 1560 | +90 | 4684 | 3735 | -949 |
| area | 84d5eadf00b30a540b9aa4dd79ce8a690aae9136 | 234ab1aa5f728a5221fb6075c35b94f88a2c6437 | 1468 | 1558 | +90 | 5059 | 3894 | -1165 |

Target reason: an object refinement using an open numeric enum as a literal tag. Main 966 -> 0; area 1182 -> 0.

No pass-to-fail or fail-to-pass moves on either tree for whole program, own file, or ordinary lowering.

Main adapted inputs are byte-identical to the baseline (82 files). Area has ten changed files; see adapted-source-comparison.json.

Both trees: SkippedDependency=3, error=0, panic=2 (baseline panic=3).

Enum-tag-related newly exposed NotYet rows:

main:

| Reason | NotYet | Refused | Owner |
| --- | ---: | ---: | --- |
| a checked numeric enum object view with an incompatible structured or optional payload field parent | 24 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _propertyAccessExpressionLikeQualifiedNameBrand | 20 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _expressionBrand | 16 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field left | 6 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field statements | 4 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _literalExpressionBrand | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field escapedText | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field expression | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field name | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _jsDocTypeAssertionBrand | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field tag | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field textSourceNode | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field elements | 1 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field multiLine | 1 | 0 | OWNER BLANK |

area:

| Reason | NotYet | Refused | Owner |
| --- | ---: | ---: | --- |
| a checked numeric enum object view with an incompatible structured or optional payload field parent | 24 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _propertyAccessExpressionLikeQualifiedNameBrand | 20 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _expressionBrand | 16 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field left | 6 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field statements | 4 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _literalExpressionBrand | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field escapedText | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field expression | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field name | 3 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field _jsDocTypeAssertionBrand | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field tag | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field textSourceNode | 2 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field elements | 1 | 0 | OWNER BLANK |
| a checked numeric enum object view with an incompatible structured or optional payload field multiLine | 1 | 0 | OWNER BLANK |

Disappearance of the old Refused reason does not prove 101 sites lower successfully. Both trees gain 90 NotYet sites and 17 other Refused sites. The census does not independently validate the compiler author's 865/101 classification or runtime semantics.

Meter exit=0. Setup exit=0. All 23 meter tests, including two real planted mutants, pass. No full compiler/oracle gate was run.
