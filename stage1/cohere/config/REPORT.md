Strict CohereSettings and JSONC tsconfig loaders are implemented in stage 1.
Go production oracles compare canonical settings and ordered files across three backends.
Discovery reads inherited lint ignores from the Adamic settings loader.
Three new loader mutants supplement the discovery and formatter mutation checks.
Directory-link parity is closed; advanced tsconfig gaps remain documented.

## Original discovery and formatter implementation (commit 0272907)

Read CLAUDE.md, stage1/cohere/gitignore and stage1/cohere/formatfiles whole before changes. Started from origin/main 5d4c801 on codex/stage1-config. No compiler territory or submodule files changed. New code is in stage1/cohere/config; formatfiles now applies settings-relative lint globs. Its settings resolution is still supplied by the Go oracle. No claim of Adamic-resolved settings is made.

See [GAPS.md](GAPS.md) for exact settings observations, the lowering refusal proving program, generated discovery cases, three discovery mutants and performance limitations. A checked JSON parser implemented in Adamic remains possible; JSON.parse refusal does not prove the whole unit impossible.

## Original setup and verification

Commands ran with /workspace/adamic-tools/env.sh sourced. Test stdout and stderr were redirected into logs, never piped.

- `bash cloud/setup.sh > /tmp/stage1-config-setup.log 2>&1`: Go, clang, Node and submodules ready 0 s; cache warm 78 s; done 78 s. `nproc`: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0.
- `ADAMIC_CONFIG_TIMING=1 go test -v -count=1 -timeout 30m ./stage1/cohere/config > /tmp/stage1-config-test-final.log 2>&1`: PASS, 16.035 s. 128 roots, 626 project markers, 288 glob pairs; all three discovery mutants caught on native and Node; native ASan/UBSan and LeakSanitizer passed.
- `go vet ./... > /tmp/stage1-config-vet.log 2>&1`: passed, empty log.
- `go test -v -count=1 -timeout 30m ./stage1/cohere/config ./stage1/cohere/gitignore ./stage1/cohere/formatfiles > /tmp/stage1-config-regression.log 2>&1`: PASS, package times 35.828 s, 32.041 s, 32.731 s. This preceded the expanded formatter corpus; final formatter run is below.
- `ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestInputAgreesWithNode$' > /tmp/stage1-config-oracle.log 2>&1`: PASS, 11.278 s, six uncached input fixtures; Node/native/backend and sanitizer/leak checks.
- `go vet ./stage1/cohere/config ./stage1/cohere/formatfiles > /tmp/stage1-config-vet-final.log 2>&1`: passed, empty log.

The full repository test gate was not run. Scoped stage1 packages plus the input oracle were used. The initial discovery parity run found a missing-directory error translation mismatch; the corrected rerun passes. An initial gap probe checked only typechecking, then was corrected to inspect the exact lowering refusal. These are observations, not missing checks silently accepted.

## Formatter cases and mutants

Final corpus includes all three requested real roots, all 20 literal/helper cases, 400 generated cases, and settings above the walk root with a nested repository boundary: 423 roots and 29,366 files offered. File order, errors, ignores, extension counts, directories and repository queries are all compared. Go settings are inputs, not a ported loader.

Every formatter mutant is required to exit successfully and differ in output on both native and Node. Three added mutants:

- Lint ignores never match: offered list/count differs.
- A file glob prunes a matching directory: walked 4 versus Go 5.
- Lint globs relative to walk root instead of settings directory: ancestor-settings fixture differs.

Nine inherited mutants also run: JavaScript extension lowercasing; UTF-16 declined-extension sorting; symlink-loop suppression; DEL quoting; quote quoting; dangling symlink suppression; trailing-dot extension suppression; dangling .git mistaken for repository; dangling .prettierignore accepted. Their exact caught discrepancies and final results are recorded below.

## Performance interpretation

Discovery timing is documented in GAPS.md. Formatter Go times resolution, enumeration and question/output construction inside its started process; native times process startup and the driver over already resolved settings. They are harness comparisons, not isolated throughput benchmarks. Files offered are counted with repetition across the three overlapping real roots. They are not tsconfig source files checked.

## Remaining work

Settings JSON/JSONC parser, canonical resolved settings, defaults and embedded rule sets, extends merges, errors and provenance; tsconfig include/exclude, source extensions/order, extends/defaults, solution references/ownership. CohereSettings is strict JSON in the pinned Go implementation; tsconfig is JSONC. The request for JSONC CohereSettings and byte-for-byte Go parity conflicts. Current code preserves observed Go behavior and does not implement settings loading.

Permission/race and arbitrary errno parity, non-UTF-8 names, other platforms and ignore read failures beyond the disk adapter remain uncovered. No PR was opened.

## Final formatter run

`ADAMIC_CONFIG_TIMING=1 go test -v -count=1 -timeout 30m ./stage1/cohere/formatfiles > /tmp/stage1-config-format-final.log 2>&1`: PASS, 261.717 s. Go 1.825032 s, 16,090 files offered/s. Native rounds 19.716329, 19.908479, 22.206697 s: 1,489, 1,475, 1,322 files offered/s. ASan/UBSan, backend, Node and native leak checks passed. Full formatter run cost more than four minutes; the scoped gate above was used.

Exact mutant discrepancies from the final log (temporary prefixes retained):

```text
    formatfiles_test.go:136: natively, caught: line 401: "walked 4", Go cohere "walked 5"
    formatfiles_test.go:136: on Node, caught: line 401: "walked 4", Go cohere "walked 5"
    formatfiles_test.go:136: natively, caught: line 33006: "declined \"(none)\" 1", Go cohere "declined \".\" 1"
    formatfiles_test.go:136: on Node, caught: line 33006: "declined \"(none)\" 1", Go cohere "declined \".\" 1"
    formatfiles_test.go:136: natively, caught: line 31804: "file \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-1/del\x7f.ts\"", Go cohere "file \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-1/del\\u007f.ts\""
    formatfiles_test.go:136: on Node, caught: line 31804: "file \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-1/del\x7f.ts\"", Go cohere "file \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-1/del\\u007f.ts\""
    formatfiles_test.go:136: natively, caught: line 31837: "declined \".i̇\" 1", Go cohere "declined \".i\" 1"
    formatfiles_test.go:136: on Node, caught: line 31837: "declined \".i̇\" 1", Go cohere "declined \".i\" 1"
    formatfiles_test.go:136: natively, caught: line 32274: "walked 9", Go cohere "walked 10"
    formatfiles_test.go:136: on Node, caught: line 32274: "walked 9", Go cohere "walked 10"
    formatfiles_test.go:136: natively, caught: line 32281: "declined \".😀\" 1", Go cohere "declined \".\ue000\" 1"
    formatfiles_test.go:136: on Node, caught: line 32281: "declined \".😀\" 1", Go cohere "declined \".\ue000\" 1"
    formatfiles_test.go:136: natively, caught: line 32124: "walked 12", Go cohere "enumerate error /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-7/.prettierignore: Prettier config remains where cohere no longer reads it; delete it, since the format block's \"ignore\" in the Nexus tier (NexusCohereSettings.json) and the ignorePatterns of CohereSettings.json say what the walk skips"
    formatfiles_test.go:136: on Node, caught: line 32124: "walked 12", Go cohere "enumerate error /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-7/.prettierignore: Prettier config remains where cohere no longer reads it; delete it, since the format block's \"ignore\" in the Nexus tier (NexusCohereSettings.json) and the ignorePatterns of CohereSettings.json say what the walk skips"
    formatfiles_test.go:136: natively, caught: line 31920: "walked 1", Go cohere "enumerate error walking /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-3: gitignore: /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-3/.cache is a repository of its own"
    formatfiles_test.go:136: on Node, caught: line 31920: "walked 1", Go cohere "enumerate error walking /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-3: gitignore: /tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-3/.cache is a repository of its own"
    formatfiles_test.go:136: natively, caught: line 795: "walked 4", Go cohere "walked 3"
    formatfiles_test.go:136: on Node, caught: line 795: "walked 4", Go cohere "walked 3"
    formatfiles_test.go:136: natively, caught: line 32141: "has \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-7/node_modules/quote\".ts\" 0", Go cohere "has \"/tmp/adamic-gate/TestThePortParsesAsGoCohereDoes22722496/001/trees/generated-7/node_modules/quote\\\".ts\" 0"
    formatfiles_test.go:136: on Node, caught: line 836: "walked 1018", Go cohere "walked 1017"
    formatfiles_test.go:136: natively, caught: line 143: "walked 7", Go cohere "walked 6"
    formatfiles_test.go:136: on Node, caught: line 143: "walked 7", Go cohere "walked 6"
    formatfiles_test.go:136: natively, caught: line 836: "walked 1016", Go cohere "walked 1017"
    formatfiles_test.go:136: on Node, caught: line 841: "symbolic-links 2", Go cohere "symbolic-links 3"
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_file_glob_pruning_a_matching_directory (86.83s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_trailing_dot_no_extension (86.94s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_DEL_passed_by_the_quoting's_fast_path (87.99s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_an_extension_lowercased_as_JavaScript_does (89.16s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_link_round_a_loop_taken_for_nothing (85.43s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_the_declined_extensions_sorted_by_UTF-16_unit (89.16s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_.prettierignore_link_to_nothing_let_through (88.68s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_.git_link_to_nothing_taken_for_a_repository (89.50s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_lint_globs_relative_to_the_walk_instead_of_settings (83.40s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_quote_passed_by_the_quoting's_fast_path (81.18s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_lint_ignores_never_match (82.13s)
    --- PASS: TestThePortParsesAsGoCohereDoes/catches_a_link_to_nothing_taken_for_nothing (80.81s)
```

## Strict settings and JSONC loader continuation (commit 0918aa4, before realpath)

The user confirmed the upstream asymmetry: CohereSettings.json remains strict JSON
with Go's exact errors; tsconfig.json remains JSONC. GAPS.md records it as an upstream
observation, not a request to broaden accepted settings syntax.

New Adamic modules implement checked JSON, strict settings schema/extends/merging,
embedded house sets, rule aliases, defaults, provenance, plugins, overrides, per-file
resolution, JSONC project configurations, TypeScript globs, include/exclude and source
extension priority/order. No Go resolver bridge or host JSON parser supplies these
results. Go runs only as the external production oracle. Real-project corpora cover
Adamic, cohere and its TypeScript submodule separately. Ordinary file links work;
source-contributing directory aliases are explicitly declined pending realpath.

`gaps/directory_link.ts` proves Go emits the first alias once and deduplicates its
target; native, Node and backend must emit the exact recorded gap. JSON.parse remains
a deliberate compiler refusal, but the typed JSON implementation bypasses that
library boundary. Missing/looping settings and tsconfig extends compare exact errors.
The checked parser also exercises Unicode quoting, malformed escapes, duplicate
struct-map fields, strict comments/trailing commas, unknown keys and typed schema
errors. The 80 seeded settings and 80 seeded project trees are deterministic.

Current validation logs and final results follow below. Tests redirect stdout and
stderr to files. Repository compiler files and protected worker territory remain
untouched. The runtime realpath proposal is developed in an isolated scratch copy
and is not applied without approval to expand this unit's territory.

Additional checks completed during the continuation:

- `go vet ./... > /tmp/stage1-config-vet-optimized-final.log 2>&1`: PASS, empty log.
- `go test -v -count=1 -timeout 30m ./stage1/cohere/gitignore > /tmp/stage1-config-gitignore-continuation.log 2>&1`: PASS, 22.170 s; the complete imported matcher package and its mutants.
- `ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./stage1/cohere/gitignore ./internal/oracle -run 'TestInputAgreesWithNode|TestThePortParsesAsGoCohereDoes' > /tmp/stage1-config-scoped-continuation.log 2>&1`: input oracle PASS, 1.537 s, six uncached fixtures. That filter selects no gitignore tests, so the full package was run separately above.
- `go run stage1/cohere/config/testdata/quote_table.go > /tmp/stage1-config-quote-regenerated.ts` and `cmp stage1/cohere/config/quote_table.ts /tmp/stage1-config-quote-regenerated.ts`: PASS, exact reproducibility.
- `git diff --check` and `gofmt -l stage1/cohere/config/*.go stage1/cohere/config/testdata/*.go`: PASS, no output.
- Isolated realpath proposal test: PASS 10.767 s, native ASan/UBSan, Node and backend; `/tmp/stage1-config-realpath-proposal.log`. See `testdata/REALPATH.md` and the unapplied `testdata/realpath.patch` (11 files, 58 insertions, 21 deletions; nine supporting compiler/runtime files, two unit files). This targeted test does not establish full runtime regressions or errno coverage.

Preparation failures: the first scratch copy tried to copy nonexistent go.sum;
its module workspace was then supplied through go.work. A malformed `-timeout30m`
flag invoked the wrong Go test scope (`no Go files in /workspace/adamic`); corrected
commands use `-timeout 30m`. The initial isolated build lacked the TypeScript workspace
module and reported a missing go.sum entry; copying go.work corrected it. None of
these preparation failures counts as mutation evidence.

The unoptimized final package run completed the baseline (143 settings, 130 configs,
1,232 source-file records), all three discovery mutants and the directory-link gap.
It was stopped during loader mutants to remove an unnecessary alias search when no
registered names are supplied. The optimized complete rerun is the final evidence
below. The earlier complete loader run `/tmp/stage1-config-loaders-final.log` passed
122.997 s and all three loader mutants, before the added rule-alias cases; its timing
was 69,058 Go files/s versus 14,884 to 17,216 native files/s over 1,222 records.
The initial inherited-options mutant failed typechecking and was discarded, then
corrected to change only the inherited options value. No compiler error was counted
as a killed executable mutant.

The full repository gate was not run: scoped stage1 tests and a filtered input oracle
were used, as permitted by the unit instructions. The full expanded formatter run
and its twelve executable mutants remain the unchanged results recorded above.
Setup remains the original successful 78 s cache warm/setup; nproc 5. Original
commit: 0272907727b063d3fd573179ea8acae95af2ad3b. Continuation commit is reported in
the final message and branch history; no PR is opened.

Final command:

```
ADAMIC_CONFIG_TIMING=1 go test -v -count=1 -timeout 30m ./stage1/cohere/config > /tmp/stage1-config-optimized-final.log 2>&1
PASS
ok github.com/system-inc/adamic/stage1/cohere/config 124.264s
132 discovery roots; 637 project markers; 288 lint glob queries
143 settings results; 130 tsconfigs; 1,232 ordered source-file records
Native ASan/UBSan, Node source, JavaScript backend, LeakSanitizer: PASS
```

All three final loader mutants compile and exit successfully, then differ on native
and Node. Exact first catches (temporary path normalized here only for readability):

```
strict JSON comments, line 2822:
  mutant: settings-root "ROOT"
  Go: error parsing lint config ROOT/comments.json: invalid character '/' looking for beginning of object key string
inherited rule options, line 2681:
  mutant: rule "@typescript-eslint/no-shadow" error []
  Go: rule "@typescript-eslint/no-shadow" error ["option"]
tsconfig excludes, line 26931:
  mutant: file "ROOT/src/b.ts"
  Go: file "ROOT/src/x.cts"
```

Final discovery mutants also compile and exit successfully on native and Node:
pruning gives counts 2 0 4 versus 2 3 4; requiring a double-star segment gives
glob 0 versus 1; entering nested repositories gives the cohere own-repository
error versus counts 1 0 21. Together with the twelve formatter mutants recorded
above, eighteen executable unit mutants are caught (gitignore's independent
mutation suite also passes). The updated loader catch exercises alias option
inheritance before the older eqeqeq counterexample.

Tsconfig-only throughput over the same 1,232 records:
Go 0.019389 s, 63,542 files/s; native rounds 0.075788/0.073673/0.077974 s,
16,256/16,722/15,800 files/s. Every timed native output still matches full Go bytes.
The count includes repetitions from overlapping project roots. Native timing
includes process startup/output capture; Go times production work/output within
its already-started test. These are harness measurements, not isolated equivalent
microbenchmarks and not compilation/checking throughput. Discovery native timings
were 0.268921/0.309941/0.358644 s for 637 markers; source timing is separate.

Remaining coverage: source-contributing directory symlinks require the tested
unapplied realpath proposal and permission to expand territory; arbitrary malformed
JSONC diagnostics/compiler-option errors, advanced package exports, published
version admissibility, relative entry paths, command-level ownership/solution
expansion, unusual errno, non-UTF-8 filenames and non-Linux platforms are not fully
covered. No full arbitrary configuration parity claim is made beyond the measured
real-project and generated corpora. See GAPS.md for exact proving programs and
upstream observations. No requested code is sent to an external resolver.

## Authorized realpath continuation

The user authorized applying the proposal as its own commit. It adds realPath to the
typed prelude, IR, input lowering, freshness analysis, native expression emission,
C filesystem runtime and JavaScript backend/adapter. No protected compiler file was
changed. Canonical visited-directory keys close the directory-link enumeration gap;
file output retains the first alias spelling as Go does. The source/target alias,
backlink cycle and dangling-link fixture emits only a-alias/a.ts on Go, native, Node
source and backend. Existing project discovery still skips directory links, as Go's
separate discovery walk does; tsconfig enumeration follows them.

`internal/oracle/testdata/realpath.a` is registered as an input fixture. The dedicated
TestRealPathAgreesWithNode compares its three execution paths to direct Node
fs.realpathSync, independently of the Adamic JavaScript adapter. Cases include a
linked directory, dangling link, empty path, dangling/.. normalization and a file
through the directory link. Node prints a canonical inner-directory path, the exact
`Error cannot resolve path walking/dangling: no such file`, cwd for the empty path,
walking for dangling/.. and walking/B.txt for the file case. Observed plain Node
fs.realpathSync normalizes empty/dot-segment inputs before filesystem traversal;
its native variant instead returns ENOENT for empty and dangling/.. inputs. Native
Adamic now performs lexical input normalization before POSIX realpath.

Two mutants alter only successful canonical paths or only dangling-link errors by
appending /wrong. Both compile under -Werror, exit 0 with empty stderr and have no
ASan/UBSan or LeakSanitizer finding. Each is caught by stdout comparison with direct
Node. The first targeted run passed 4.829 s (two basic cases); expanded final direct
fixture passed 5.108 s before the combined uncached run. No compilation failure
counts as a killed mutant. Full configuration regression passed 125.135 s, including
143 settings, 130 tsconfigs, 1,232 source records and six loader/discovery mutants.
The enhanced link/cycle/dangling fixture is also run separately below.

Historical proposal: testdata/realpath.patch records the original reviewed diff.
It has been applied and superseded by the tested lexical normalization and new
fixtures; do not apply it again. testdata/REALPATH.md records this transition.
Final commands and outputs follow after the running checks complete.

Final runtime checks (test output redirected to logs, never piped):

- `ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^(TestInputAgreesWithNode|TestCountsAreRecorded|TestRealPathAgreesWithNode)$' -args -update-counts > /tmp/stage1-realpath-oracle-final.log 2>&1`: PASS 50.631 s. Seven uncached input fixtures, full recorded-count corpus and both realpath mutants passed. The new row records allocations/frees 16/16, retains/releases 16/22, peak live values 4 and values freed in regions 0; all older rows are unchanged.
- `go test -v -count=1 -timeout 30m ./stage1/cohere/config -run '^TestDirectoryLinksAgreeWithGoCohere$' > /tmp/stage1-realpath-links-final.log 2>&1`: PASS 17.299 s. Go emits a-alias/a.ts once despite the backlink cycle and dangling link; native ASan/UBSan, Node and backend match.
- `go test -v -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/flow > /tmp/stage1-realpath-compiler.log 2>&1`: PASS. Package times: load 1.399 s, lower 29.290 s, native 121.184 s, fresh 42.288 s, flow 62.548 s. JavaScript package has no tests; its backend is exercised by the oracle comparisons.
- `go vet ./... > /tmp/stage1-realpath-vet.log 2>&1`, `gofmt -l cmd internal stage1/cohere/config > /tmp/stage1-realpath-gofmt.log`, and `git diff --check`: PASS, empty outputs.

The full repository gate was not run. Affected compiler packages, complete allocation
count corpus, uncached input oracle, direct Node mutants and configuration tests are
the scoped evidence. Advanced tsconfig diagnostic/option validation, package exports,
command-level ownership, unusual filesystem errors, deleted cwd, symlink targets
containing unusual dot segments and non-Linux platforms remain outside the measured
coverage. The directory-link refusal is removed, not merely hidden in the report.
The prior branch commits 0272907 and 0918aa4 remain intact; this authorized runtime
continuation is a separate commit. Push is to codex/stage1-config, with no PR.

Final complete configuration command against the finished runtime:

```
ADAMIC_CONFIG_TIMING=1 go test -v -count=1 -timeout 30m ./stage1/cohere/config > /tmp/stage1-realpath-config-final.log 2>&1
PASS
ok github.com/system-inc/adamic/stage1/cohere/config 120.172s
132 discovery roots, 637 markers, 288 lint glob queries
143 settings results, 130 tsconfigs, 1,232 source-file records
Directory alias/cycle/dangling parity: PASS
Three loader and three discovery executable mutants: caught on native and Node
Native ASan/UBSan, Node source, JavaScript backend, LeakSanitizer: PASS
```

Source-file harness timing with realpath applied: Go 0.016949 s, 72,688 files/s;
native 0.074162/0.074167/0.080792 s, 16,612/16,611/15,249 files/s. Every timed
output matches Go bytes. The earlier caveats still apply: overlapping project
records, native process startup/output capture and Go's already-started test process.
The final loader mutants retain the exact comments, no-shadow inherited-options
and excluded src/b.ts catches recorded above. Two additional Node realpath mutants
bring the executable unit mutation evidence to twenty, excluding the imported
matcher package's independent suite. git diff --check passes; protected files remain
unchanged. The authorized runtime commit is reported in the final summary and git
history, separately from 0918aa4.
