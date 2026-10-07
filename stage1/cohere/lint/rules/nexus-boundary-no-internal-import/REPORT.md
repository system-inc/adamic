Built three registered Nexus import-boundary candidates in .a; previous claimed work was already pushed, and exact shared parser/witness gaps remain recorded.
Commits: prior 4fb908d9; claim 304a28ba; internal e2d22ac5; outside a24c8067; project 28b1ae22; evidence is the subsequent report commit.
Checks: 42 upstream cases, 597 compiler/stage1 rule cases, 81 corner cases and 399 path resolutions agree on Go, Node, emitted JavaScript and sanitized native; owned package, registry, vet and filtered uncached oracle pass.
Mutants: first internal owner, unbounded alias prefix, prefix whitelist and broken parent reduction all compile/run cleanly and are caught only by output comparison on every backend.
Not covered: three blocked original fixtures, unchanged shared witness gate and production profile/.a integration, missing-option skip/coverage serialization, full repository gate and complete CLI/self-lint parity.

## Selection and ownership

Existing claims and implementations were pushed through 4fb908d9 before taking this batch. Those include the earlier suggestion ports and the independently tested JSX/Tailwind candidates, whose extraction/design-system blockers remain unchanged. This continuation fetched all origin heads and found 356 refs and 54 unique Markdown claim blobs. Main remains ef3d907ecdc4c771b016f7d9c52372def057a340. All 46 helper-ready names are unavailable. The first three remaining inventory names with needs_type_information=false, in origin/codex/lint-inventory array order, are the Nexus internal/outside/project boundary rules. Claim 304a28ba was pushed before code; evidence/selection.json records the search. No additional rules were claimed.

Only this claim and the three owned rule directories are changed. No shared generator, production harness, parser, compiler or Go cohere source is edited. All new Adamic modules are .a; witnesses remain .ts.txt. No PR is opened.

## Implementation

All three rules use normal rule.json directory registration with no order. Each directory owns its concrete class/factory, exact policy messages, Go oracle adapter, witness and compiling mutant. Each independently reads ImportDeclaration and CallExpression: direct string module specifiers, dynamic import first arguments, and direct require calls with exactly one literal argument. Parenthesized callees, member methods, computed arguments, reexports and import-equals are not widened into findings. Call argument extraction uses the parser's list count, preserving generic and optional-call layout. Static imports with attributes retain the direct module specifier instead of mistaking the attributes node for it. Findings cover the quoted/backticked specifier, not the declaration or call; no edits are invented.

The internal rule distinguishes real path segments from substrings, refuses aliased/absolute internal access, resolves relative sources from the importing directory, chooses the last internal segment, and compares ownership against the slash-normalized original filename. Its owned path helper retains POSIX, UNC, drive, URL, file-volume and untitled roots, parent reduction and trailing separators. Actual TypeScript Go tspath functions decide the path results; the test-only oracle does not duplicate them.

The outside rule activates only within /libraries/nexus/ and matches @project, @structure and @base at an exact alias boundary. The project rule activates only within the configured nonempty libraryDirectory, matches @project at a boundary and admits exactly the allowed specifiers. Its typed option decoder uses the shared JSON grammar and then processes validated top-level fields in source order. This preserves Go's duplicate/null semantics: a null string field leaves its earlier value intact, while null/empty allowed arrays clear the whitelist. ASCII field-name folding, null array elements and ignored unknown fields match the Go adapter's encoding/json behavior. Malformed-option diagnostic prose and shared registry/schema acceptance are not certified here.

## Observed correctness

The published harness is f4d98cab50048692781da3599131317dc569d466, used in an isolated worktree. It provides .a loading, emitted-JavaScript comparison, suggestions and profile fixes. Production shared files remain unchanged. Go cohere stays pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db.

An owned Go overlay records actual upstream Run results before assertion helpers return. All 45 unique original source/rule/filename/options combinations were captured; the upstream Go family tests passed. Three have independently proven shared parser gaps. The remaining 42 retain their original directory layout and options in the temporary manifests and produce 10,480 identical bytes across Go/source Node/emitted JavaScript/sanitized native. This includes finding ids/descriptions/ranges, source formatting protocol and unchanged fixed sources, not only counts.

The corpus comprises 77 compiler sources at TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8 plus 122 stage1 .ts/.a sources from this branch. Generated files and explicit gaps are excluded. All three rules run for each source: 199 files, 597 selected cases, 37,294,174 identical output bytes. The project rule receives libraryDirectory="/" so it is active, rather than being certified through its disabled default. Fixture and stress positives separately prove that quiet corpus results are not an inert implementation.

The 81 corner cases cover three importer directories, all three selected rules and nine option configurations, including duplicate/null fields and arrays, field casing, unknown nested fields, whitespace/UTF-8 offsets, import attributes, optional/generic require, extra require arguments, parenthesized/method callees, nonliteral imports, backticked imports and nested internal ownership. They produce 80,249 identical output bytes. A fake raw-backslash filename was removed after Go's SourceFile constructor explicitly rejected it as violating its normalized-absolute-path precondition; it is not represented as Windows parser coverage.

The path model independently compares 399 combinations against Go tspath, including relative/root/drive/UNC/URL/file-volume/untitled and Unicode paths: 6,861 identical bytes on all four paths. Go exposed a C: root mismatch, which was fixed before the final run. The owned package, containing both path and blocked-fixture tests, passes in 70.599s. Registry tests pass in 0.048s, including negative descriptor and stale-extension controls. Owned and published-harness vet logs are empty. TestTheOracleCatchesOneByte passes uncached in 0.407s, with native and Node cache hits 0 and misses 1.

Some combined runs are retained as failures: an incorrect compiler directory, deprecated assert syntax, and the invalid raw-backslash filename prevented the corresponding corpus/corner subtests. Successful subtests from those logs are named individually above; no combined failure is called a green package gate. The final corrected corner test passes in 29.738s. The final upstream and per-rule mutant subtests pass in their retained log. The owned production Go package passes as a whole.

## Every semantic mutant

Compilation, sanitizer errors, panic and stderr do not count as kills. All four variants compile and exit 0 with empty stderr, then disagree with external Go:

| Mutation | Discriminating input | Catch |
| --- | --- | --- |
| Choose first rather than last internal segment | Import from ./deep/internal/Detail while already within an outer internal folder | Missing outsideInternal, all three backends |
| Match unbounded alias prefixes | @basement/Thing and @projections/Thing | Invented forbiddenOutsideImport, all three backends |
| Widen whitelist to prefixes | @project/ProjectSettingsSecret beside allowed ProjectSettings | Missing forbiddenProjectImport, all three backends |
| Push .. instead of reducing a parent | Paths containing relative parents | Changed resolved path, all three backends |

The three rule mutants are recorded in final-fixtures-mutants-and-invalid-filename.log.txt. The path mutant is in owned-final.log.txt. Original fixtures are actual external inputs, not synthetic expected strings copied from the port.

## Exact remaining shared blockers

1. Two original specifier-range fixtures end with a literal backslash-n after the statement, not a newline. Go's rule test harness accepts the recovered AST and reports outsideInternal/forbiddenProjectImport. Stage1 parsing refuses an Unknown token at positions 87 and 84 respectively.
2. The outside-rule fixture `const loaded = await import('@structure/source/Thing');` receives a Go finding, but the shared parser refuses top-level await at position 21 with "parser slice expected semicolon". Async-function import fixtures do pass.
3. The unchanged shared ownedWitnesses helper recreates every witness as a temporary basename and supplies no options. Thus the outside rule sees a file outside Nexus, and the project rule sees no libraryDirectory. TestOwnedWitnesses fails in 4.514s with `nexus/boundary-no-nexus-outside-import witness reports no findings`. The project rule has the same configuration limitation after that first failure. positive.case.json records the intended importer and options as reviewable data; it does not claim that the shared harness already understands that sidecar.
4. On this production branch setup/cache warming still fails at profile_test.go:32 because portFiles is now a function. Its generator predates .a support. The isolated published harness removes those two issues for validation but is deliberately not merged or cherry-picked into shared production files.
5. RuleContext exposes findings, not a skip/coverage destination. The project rule's missing-directory decline matches Go findings, but the Go Skip reason cannot be serialized by this shared interface. Full coverage/CLI behavior remains pending that API.

The owned gaps/ oracle calls the actual three Go rules on the blocked original sources and requires one finding each. The same parser probes compile to source Node, emitted JavaScript and sanitized native, and all three return exit 70 with byte-identical stderr and no stdout. See owned-final.log.txt. These are proofs of a blocker, not finding parity and not rule-mutant kills. No full-original-fixture or default-gate certification is claimed.

## Findings per second

Observed best of five interleaved count-only samples on 77 compiler files plus a file containing 1,000 positive imports. Counts agree on all three engines. Process startup, file reading, parsing and visitors are included; compilation, finding formatting and fixing are excluded. Native timing uses an unsanitized release build; all correctness/mutant comparisons use ASan/UBSan. Other bounded checks ran during some samples, so these rates describe this workload and machine, not a performance guarantee.

| Rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| nexus/boundary-no-internal-import | 544.23 | 827.19 | 3899.22 |
| nexus/boundary-no-nexus-outside-import | 552.96 | 822.13 | 4417.80 |
| nexus/boundary-no-project-import | 494.82 | 918.84 | 4270.53 |

Each sample contains 78 files and 1,000 findings. Throughput passes in 68.43s. An earlier stress-only run accidentally pointed at the wrong corpus directory; its one-file numbers are retained only as superseded evidence and are not the reported rates. The final test requires exactly 77 compiler files.

## Commands and reproduction

`bash cloud/setup.sh` was run on the production branch and failed during Go package cache warming at the known profile mismatch. Its timing lines were go ready 1s, clang ready 1s, Node ready 1s and submodules ready 1s; it printed no cache-warm/done timing because it failed. nproc prints 5. The tools themselves are available: Go 1.27.1, clang 20.1.8 and Node 24.19.0. Sourcing /workspace/adamic-tools/env.sh and using the unmodified published harness is the workaround.

Every test writes directly to a log file. The following were run in /workspace/scratch/wave12-harness with ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 and ADAMIC_STAGE1_SOURCE=/workspace/adamic/stage1:

    go test ./stage1/cohere/lint -run '^TestNexus(Upstream|Corners|Mutants)$' -count=1 -v -timeout=15m > /tmp/wave12-batch6-final-validation.log 2>&1
    go test ./stage1/cohere/lint -run '^TestNexusCorners$' -count=1 -v -timeout=10m > /tmp/wave12-batch6-corners-final.log 2>&1
    go test ./stage1/cohere/lint -run '^TestNexus(Corpus|Corners|Throughput)$' -count=1 -v -timeout=15m > /tmp/wave12-batch6-corpus-corners-rate2.log 2>&1
    go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave12-batch6-registry.log 2>&1
    go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave12-batch6-harness-vet.log 2>&1
    go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave12-batch6-default-witnesses.log 2>&1

From /workspace/adamic:

    go test ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import -count=1 -v -timeout=10m > /tmp/wave12-batch6-owned-final.log 2>&1
    go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import > /tmp/wave12-batch6-owned-vet.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave12-batch6-oracle.log 2>&1

Raw logs are copied unchanged under evidence/. validate.py --scratch <new-directory> --typescript /workspace/scratch/typescript-6.0.3 archives the exact published harness, links the unchanged pinned Go cohere, copies these three owned rules and the owned virtual tests, and writes separate logs. It does not change production shared files or run setup against the symlink checkout. Python syntax is checked; the recorded executions used the equivalent real worktree rather than this convenience wrapper.

The full repository gate is not run. No general malformed-configuration diagnostic, full CLI/schema/coverage behavior, arbitrary parser recovery, raw Windows filename parsing, or pinned CLI .a self-lint/format parity is claimed. The useful rule implementations, exact blocked inputs, compiling mutants, measured rates and reproducer are concrete and reviewable without changing shared territory.
