Built and certified mixed-array consumers, alias-preserving array-to-tuple admission, original fileNames admission and watcher callback arrays.
Commits: mixed b8c9c155, tuple e24b37f3, fileNames 70522aa1, callbacks 4712e290 and 908cafeb; all pushed to codex/views-arrays-callables-parser.
Commands: Node, sanitized native, release native, JavaScript and finishing leak checks pass in each group's focused logs; latest callback suite 34.465s, counts 5.521s, vet passes; next boundary test passes in 0.525s.
Mutants: nine mixed, five tuple, four fileNames and eight callable mutants are caught by their stopping or Node oracles; each is described in its group report.
Not covered: next three-read pairs, production consumer/intrinsic totals, optional/rest callback source-call expansion and nonvoid result boxing; SourceFile.amdDependencies reaches an intersection refusal.

The three requested joints are finished. Cross-lane changes have their own commits: b2aa9c9c changes lane 4 union adapters, 0acf7c6c changes lane 7 controls, 998de79f changes lane 4b controls, 78c73bab changes lane 4 union receiver reads, and c451c9e8 changes lane 4c tuple runtime and frontier controls. Shared tuple hooks are f4563b82; shared common-array cast hooks are cddc09c1. Detailed declarations, original spans, pinned negatives, mutants and measured counts are in joints/MIXED-ARRAYS-REPORT.md, joints/ARRAY-TUPLE-REPORT.md, joints/FILE-NAMES-REPORT.md and RANKED23-ARRAYS-REPORT.md.

Lane 2 totals are 184 certified pairs / 2903 original static candidate reads, remaining 150 / 286 of 334 / 3189. This batch adds five array-field pairs / eleven reads. The separate tuple frontier adds one tuple pair / four reads without duplicate lane 2 credit. Two mixed consumer pairs are also certified, with their array overlap counted once. Own fields remain 3 / 26 of 30 / 72. Production execution counts are not inferred from static spans.

The next ranked SourceFile.amdDependencies pair has three original candidate reads. A complete original SourceFile cast followed by amdDependencies[0]!.path prints a on Node. Lowering refuses exactly: `Adamic 0.1 refuses checked view read of field path with unsupported intersection contract; prove or implement the intersection contract before reading this field`. TestCheckedViewRanked24AmdDependencyFrontier pins that unchanged boundary. This is an intersection handoff to investigate with lane 7; its cause has not been inferred from the diagnostic alone. No backend, leak or mutant certification is claimed for this refused probe. The inline regression adds no fixture row. TypeMapper.sources, TypeMapper.targets and FlowReduceLabelData.antecedents have three reads each and compile-only discovery observations; they remain uncertified.

The final fetch succeeds; origin/codex/views-integration remains 432d4913d31daaa49d8ca4eb46f30f21f90da5ee, already an ancestor. No merge is needed. No code conflicts occurred. The next boundary command was:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ARRAY23_ORIGINAL_DECLS=/workspace/lane2-original-declarations23 go test ./internal/oracle -run '^TestCheckedViewRanked24AmdDependencyFrontier$' -count=1 -v > /tmp/lane2-ranked24-frontier.log 2>&1
```

The required global count refresh failed in pre-existing fixtures for each group; scoped refreshes preserve prior rows and record 54 owned fixtures across the four groups. Those failures and exact commands are retained in each report. No whole package test or full gate ran. Setup completed in 43.041s; timing lines were Node/Go 0.023s, markdown 0.069s, submodules 0.079s, clang 0.221s, build 42.907s, deferred 43.012s, cache 43.014s. nproc is 5, with quota 4.
