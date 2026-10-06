Project discovery, lint globs and formatter ignorePatterns are implemented.
Settings loading and tsconfig source enumeration remain incomplete.
Go production oracles compare Node, native and JavaScript backend bytes.
Three discovery and twelve formatter mutants exercise the comparisons.
This branch is a partial port with an explicit JSON parsing gap.

## Implementation and scope

Read CLAUDE.md, stage1/cohere/gitignore and stage1/cohere/formatfiles whole before changes. Started from origin/main 5d4c801 on codex/stage1-config. No compiler territory or submodule files changed. New code is in stage1/cohere/config; formatfiles now applies settings-relative lint globs. Its settings resolution is still supplied by the Go oracle. No claim of Adamic-resolved settings is made.

See [GAPS.md](GAPS.md) for exact settings observations, the lowering refusal proving program, generated discovery cases, three discovery mutants and performance limitations. A checked JSON parser implemented in Adamic remains possible; JSON.parse refusal does not prove the whole unit impossible.

## Setup and verification

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
