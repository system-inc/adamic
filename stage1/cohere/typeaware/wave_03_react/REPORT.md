Built: native react-hooks/unsupported-syntax and react-hooks/use-memo; react/boolean-prop-naming supported profiles, with arbitrary configured regex explicitly blocked.
Commits: claim 570f68e1; implementation and comparison evidence 59d090ce536ff4bbc72a1146a87bcd30491a2338; report commit follows.
Commands/results: TestWave03ReactAgreementAndMutants PASS 272.943s; 266 parseable controls and all 364 frozen corpus roots match Go byte for byte; sanitizer and released-handle checks pass.
Mutants: three byte-only rule mutations, memo alias filter, regex refusal, 27 compiler-query mutations and retained released handle were caught by their intended checks.
Not covered: arbitrary configured regex, emitted-JavaScript external-checker execution, full repository gate or exhaustive syntax/Unicode fuzzing; no additional rules claimed.

The nine previously claimed rules were already completed and pushed at c2758b18. This continuation fetched all origin heads and audited ports on main and the bridge branch plus claims on every origin branch. The first three remaining entries were the three above, each with zero measured compiler/repository findings. The selection snapshot is `../validation-wave-03/continuation-3-selection.json`, recording 389 origin refs, the distinct claim blobs and the ranked candidates. The reservation was pushed before implementation. These three claims remain reserved; boolean naming is partial until general native regex is available.

All sixteen production native modules are .a. Rule decisions, component inference, return traversal, dependency validation, callback checks, prop-type selection, naming and message interpolation run in Adamic. New Go files return raw syntax graphs, trivia positions and symbol declarations. Virtual-library queries operate on already-loaded compiler sources through a real source anchor, preserving bundled URIs without changing the shared path resolver. The only prior dispatcher edit forwards its default to this owned dispatcher and removes the now-unused fmt import. Shared registration generators, harness files, facts.go and protected compiler implementation files are untouched.

Go's compilation-root inference and the separate component detector used by boolean naming are reproduced separately. The latter preserves its differing null-return and traversal rules. Go's foreign-declaration diagnostic behavior is also preserved: the range trims trivia in the file being linted rather than the declaration's source. Controls cover a physical global declaration and bundled DOM interfaces and type aliases.

The independent oracle loads the pinned cohere program and calls its unmodified production rules. It imports no bridge code. Fixture bodies are extracted from production Go test tables, supplemented by the vendored eval fixture and controls for Unicode, virtual declarations, aliases, callback assignments and naming boundaries. All 266 candidates parse independently. Every output includes rule/message ids, exact UTF-8 spans, UTF-16 escaped message bytes, fix counts and edits, and suggestion counts and edits. These three Go rules produce no fixes or suggestions; all compared counts are zero.

| Profile | Findings |
| --- | ---: |
| No boolean options | 90: 35 unsupported syntax, 55 memo |
| Boolean options object with defaults | 177 |
| Default pattern with nested props | 182 |
| is-only fixture pattern | 192 |
| Unanchored fixture pattern | 176 |
| Extended fixture pattern | 172 |
| Nested props, custom prop types and message interpolation | 185 |
| Empty prop-type list fallback and short is-only pattern | 192 |

All ten unsupported-syntax/use-memo message ids have positive controls. Boolean controls exercise identifiers, quoted/computed keys, isRequired, nested calls, TypeScript interfaces, aliases, unions/intersections, binding annotations, null/JSX returns, foreign declarations, bundled aliases and Go simple-uppercase cases. Custom templates include unknown placeholders, overlapping braces, nonbreaking whitespace and vertical tabs, preserving Go's interpolation behavior.

The frozen corpus is the same 77 TypeScript compiler roots and 287 repository roots used by the preceding waves. TypeScript is 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db and typescript-go remains 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Normal and sanitized native executions match Go with zero findings on both populations, with boolean options absent and with defaults plus nested validation enabled. Zero corpus findings are supplemented by the positive controls above.

The rule mutants change only output bytes: one extra space in the inline-class explanation, one extra space in the missing-callback explanation, and a straight apostrophe instead of U+2019 in boolean naming. They compile, exit zero, write no stderr and retain findings counts 90, 90 and 182. Comparison catches bytes 772, 24054 and 45373 respectively. The memo alias filter mutant compiles and exits zero but loses alias findings, caught at byte 36132. The unsupported-pattern guard mutant exits zero where the real implementation must panic 70. Retaining a released registry entry similarly exits zero where the released syntax query must panic 70.

The 27 Go overlay mutations compile and fail their intended raw-query assertions. They cover spans, parent/child links, name/type/initializer/body identities, flags, conditional/binary fields, rest/argument/parameter lists, text/operator fields, exact request guards, canonical offsets, trivia positions, virtual source selection and virtual symbol selection. Their changes and logs are in validation/bridge-mutants.json and the compressed query-mutant evidence. A first empty-argument-list mutation was rejected by Go as an unused variable; it was replaced by a compiled empty slice and is not counted as a killed check. A missing toolchain environment also failed before testing and was corrected.

Final native timing was measured sequentially after the builds and other tests finished, with exact stdout comparisons. These are observations from one round. Native is slower; this syntax-transfer approach is not evidence of a speed win.

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository, 287 roots | 2.611470305 | 0.428424306 | 6.10 |
| Compiler, 77 roots | 7.273282214 | 1.684246362 | 4.32 |

The memo import-name filter reduced ordinary symbol queries without changing alias behavior. Final corpus queries are one raw syntax tree per root: 287 and 77. Repository native load/query/run: 315.846/844.324/1910.284 ms; Go load/rule/run: 202.770/6.008/109.666 ms. Compiler native: 916.963/2872.568/6146.751 ms; Go: 965.649/239.592/637.918 ms. These phase boundaries differ, so they are not isolated rule-time comparisons. Timing streams are preserved in validation/quiet-timings.json and runs/.

Commands began with `source /workspace/adamic-tools/env.sh`; all test subprocess output was redirected to logs. The original successful setup remains applicable: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, warm cache 174s, total 174s. nproc remains 5, with four cgroup CPUs. Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- `ADAMIC_WAVE03_REACT_ARTIFACTS=/workspace/wave-03/react-final-evidence ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03ReactAgreementAndMutants$' -count=1 -v`: PASS 272.943s, controls, seven option profiles, byte-only mutants, alias/refusal mutants, all corpus comparisons, sanitizer and released handles.
- `go test ./bridge/tsgo/checker -run '^TestWave03(ForeignSyntaxAndSymbol|SourceSyntaxTree|SourceTokenStart)$' -count=1 -v`: PASS 0.913s; 78 exact raw nodes, 48 trivia positions including out-of-file positions, and bundled AST/symbol controls. The final full touched-package run passed checker in 1.215s; code_path_graph has no direct tests and remains covered by the prior Nexus comparison.
- `go vet ./bridge/tsgo/checker ./bridge/tsgo/code_path_graph ./stage1/cohere/typeaware`: exit zero, empty log. Go files are gofmt-clean and git diff --check is clean.
- `ADAMIC_WAVE03_QUICK=1 ADAMIC_WAVE03_NEXT_QUICK=1 ADAMIC_WAVE03_MORE_QUICK=1 go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants)$' -count=1 -v`: PASS 130.679s, all previous nine rules' positive comparisons including ordering, imports, fixes and option variants.
- `TMPDIR=/tmp/adamic-gate go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(functions|maps_and_text|lone_surrogates|sorting|closures|string_index|method_closures|generic_functions)[.]a$' -count=1 -v`: eight existing Node/emitted-JavaScript/native/released/sanitizer fixture checks PASS 56.529s.
- The production source and formatting gate, through the preceding wave's preserved scratch .a loader adapters, passes all sixteen modules: 276 rules, 100% Adamic-ready, no findings, exit zero. The prelude is loaded for typing but is not in the checked file selection. No harness or production compiler change implements these scratch adapters.

Validation preserves 225 compressed streams, source hashes, the 266 fixture bodies and config, frozen manifests, matched-output hashes and all mutation overlays. Independent expected results come from production Go; native rule code generates no Go verdict fixtures.

General native regex is a concrete remaining block. The minimal gaps/general_regex.a program type-checks, then `adamic build` exits one with `stage 0 can't lower new an Identifier yet` at new RegExp. The naming implementation therefore accepts the five nonempty configured patterns used by Go's entire fixture table, plus the empty pattern, and refuses others before running with an explicit panic 70. It does not claim arbitrary RE2 support. A native regex implementation or shared compiler support is required to finish that option surface; the shared compiler was not edited.

Emitted-JavaScript checker execution remains another concrete gap. `adamic js main.a` exits one at foreign_node_symbol_context.a:4 with `Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>`. The fetched lint-harness-dot-a branch provides .a and four-way support for the generic stage1/cohere/lint framework, not this standalone external-C-checker driver. No fake checker or replayed verdict replaces that execution. Generic registration and JavaScript checker integration remain for that shared work.

No full repository gate or exhaustive syntax/Unicode fuzzing was run. The previous shared unknown-question mutation anchor in TestInspectRequestRefusals remains untouched; owned malformed-request controls are independently proved instead. Native regex support remains partial, so this unit stops with the existing claims and takes no more rules.
