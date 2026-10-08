Landing meter comparison against 20261007T205008Z.iT8wjb

Whole program: main: 2/79; area: 2/79
Own file: main: 55/79; area: 55/79

No pass-to-fail or fail-to-pass moves on either tree for whole program, own file, or ordinary lowering.

| Tree | Source and compiler SHA | NotYet before | NotYet now | Delta | Refused before | Refused now | Delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| main | f4efdd2369311d1420aa53fdf5c1a55bdda811d4 | 1470 | 1470 | +0 | 4684 | 4798 | +114 |
| area | b61e70642652a7f23af113278ce2fc357d7b4646 | 1468 | 1468 | +0 | 5059 | 5167 | +108 |

Baseline source refs: main 48c05d091f0a43c31cbe051b1d6578d99eeedf19; area 03ccf222dfeedef4bfe5cc3e219e4586d14a5c13. Both baseline trees used compiler 84d5eadf00b30a540b9aa4dd79ce8a690aae9136.

Main adapted inputs are byte-identical to the baseline (82 files). Area has ten changed files; see adapted-source-comparison.json.

Both trees retain SkippedDependency=3, error=0 and panic=3.

The unowned reason "an object refinement using an open numeric enum as a literal tag" remains main 966 and area 1182. This run uses landed compilers, unlike the earlier unmerged enum-fix evidence branch.

Counts are unique sites measured on a checker-rejected program. They do not establish native output or exhaustive blocker coverage. No full compiler/oracle gate was run.
