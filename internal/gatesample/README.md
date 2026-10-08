# Landing corpus samples

`ADAMIC_GATE_SAMPLE` is an opt-in landing switch: exactly 40 hexadecimal characters, representing the main SHA. An unset switch preserves the existing full corpus and its order. An empty or malformed set value fails the corpus test by name. `ADAMIC_GATE_CHANGED` optionally names a UTF-8 file containing changed repository paths, one per line. An unreadable list fails rather than silently dropping required files.

The shared `Select` helper sorts a copy of the physical corpus paths. It selects indices congruent to `uint32(first eight hex digits) % stride`, together with every directly named changed corpus file and every corpus file in a changed Go package's `testdata/` or `fixtures/` tree. A changed adapter inside those trees belongs to the package above the fixture tree. Explicit physical control witnesses can also be retained. There is no timing-dependent selection, random seed, environment-dependent stride, or copied selector implementation.

Every sampled corpus test logs once, using `Selection.Log`:

```text
gate-sample: <TestName> <checked>/<total> files, stride <n>, offset <k>
```

JSON generated documents, exhaustive scalar and sequence checks, explicit controls, mutants and native canaries run in full. Markdown generated layout cross-products now use indexed sampling, as described below. JSON retains its numeric-separator witness, its 34 generated inputs, the exact nine-difference upstream report, sanitized/native/backend/Node comparisons, leaks, and the unconditional release comparison. Its timing opt-in remains separate. Markdown retains full mutant inputs independently of its sampled baseline; numeric expected rows are projected from the full Go mutant oracle without weakening comparisons. The layout fixture still checks the unchanged sanitized and release builds against Node and Go, poisons every supplied width field, and performs the original leak check.

## Pins and coverage

Base: `origin/area/stage1-format` at `6a8eebf41710e9bb696da9a7872d19765b1e7476`, descended from the requested `06bb3e90`. No compiler or gate-runner changes are included. The gate producer must provide the two agreed switches; an unset switch still runs the full existing corpus.

The physical totals below are this checkout's measured totals. Markdown's existing census opt-in is unchanged: default is the unit's three physical documents; `ADAMIC_MARKDOWNBLOCKS_CENSUS` expands it to repository/submodule documents. Both modes now sample generated layout cross-products. The remote gate-log branch confirms the nine shared layouts and tokenizer events were among the slow tests; chunks shares their corpus and selector. The six other Markdown corpus tests also use the helper, with stride 1, so no corpus test in this package silently ignores the switch.

| Test | Pinned stride | Physical total (default corpus) |
| --- | ---: | ---: |
| JSON `TestPortMatchesGoCohere` | 32 | 2458 |
| JSON `TestUpstreamRepositoryCorpusParity` | 8 | 2458 |
| Markdown `TestMarkdownListLayout` | 9 | 3 |
| Markdown `TestMarkdownQuoteLayout` | 9 | 3 |
| Markdown `TestMarkdownTableLayout` | 9 | 3 |
| Markdown `TestMarkdownCodeBlockLayout` | 9 | 3 |
| Markdown `TestMarkdownHTMLBlockLayout` | 9 | 3 |
| Markdown `TestMarkdownWhitespaceLayout` | 9 | 3 |
| Markdown `TestMarkdownLeafComposition` | 9 | 3 |
| Markdown `TestMarkdownRootLayout` | 9 | 3 |
| Markdown `TestMarkdownStructureLayout` | 9 | 3 |
| Markdown `TestMicromarkInputChunks` | 9 | 3 |
| Markdown `TestTokenizerEvents` | 9 | 3 |
| Markdown `TestFrontMatterStage` | 1 | 3 |
| Markdown `TestMarkdownParserPrefixes` | 1 | 3 |
| Markdown `TestMarkdownASTPreprocessing` | 1 | 3 |
| Markdown `TestMarkdownAstPath` | 1 | 3 |
| Markdown `TestNativeMdastConstruction` | 1 | 3 |
| Markdown `TestWholeDocumentOraclePreflight` | 1 | 3 |

## Measurement arithmetic and limits

Local box: AMD EPYC 7763 reported by `lscpu`, five visible CPUs (`nproc` 5), cgroup quota four CPUs, 16 GiB memory limit. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup completed in 35.729 seconds using `GOPROXY='https://proxy.golang.org|direct'` and `bash cloud/setup.sh`. Timing output stayed in local log files.

The brief supplies the 64-core Threadripper sweep measurements: JSON port 888 seconds, upstream parity 226 seconds, and the slowest cited Markdown sweep 241 seconds. Local initial full sweeps measured 445.68, 79.43, and 83.40 seconds respectively. Calibration factors are `888/445.68 = 1.9925`, `226/79.43 = 2.8453`, and `241/83.40 = 2.8897`. These are empirical whole-sweep ratios, not a claim that serial native execution scales by `64/4`.

For JSON, amortizing the supplied seat durations over the measured 2458 physical files gives `888/2458 = 0.3613 s/file` and `226/2458 = 0.0919 s/file`. The corresponding stride estimates are `ceil(888/30) = 30` (rounded conservatively to 32) and `ceil(226/30) = 8`. Their variable-work estimates are `888/32 = 27.75s` and `226/8 = 28.25s` on the seat, before required changed files and fixed work. This is an estimate: the supplied log's corpus cardinality could not be independently checked, and sample file sizes are not uniform.

For the eleven cited Markdown sweeps, `ceil(241/30) = 9` pins the stride. `241/9 = 26.78s` is only an amortized sweep estimate. With three default physical files and 4067 unsampled generated layout documents, it is not a marginal per-file measurement or a valid prediction that the entire test will become nine times faster. Mandatory changed files can also select the whole small corpus. The six stride-1 sweeps have small measured file-processing contributions: the slowest observed three-round rates were 9127.2 texts/s for front matter, 8055.3 for prefixes, 4546.0 documents/s for AST preprocessing, 907.5 for AstPath, 5574.9 for mdast construction, and 200.0 for preflight. Even the conservative calibration factor 2.8897 gives at most `3 * 2.8897 / 200 = 0.0434s` amortized per pass of their default physical files. Compiler builds, parsing, generated cases, admission waits and full mutants are additional fixed work; those throughputs do not measure them.

**The under-90-second Threadripper landing target is not established by these local runs.** The helper and integrations are functional, but physical-file sampling cannot remove mandatory generated/scalar work or native builds. The stride estimates must be checked in the actual landing gate, especially for Markdown. No check was weakened to manufacture a deadline result.

## Validation

The fake SHA is `00000008` followed by 32 zeroes. The changed list names `CohereSettings.json` and `stage1/cohere/markdownblocks/GAPS.md`. JSON port selected 79/2458 files (stride 32, offset 8), upstream parity selected 310/2458 (stride 8, offset 0), and the eleven targeted Markdown tests selected 1/3 (stride 9, offset 8). The JSON artifact contains the explicitly named changed file and all 34 generated inputs. In Markdown, offset 8 alone selects none of the three files: the one selected file is therefore the explicitly named `GAPS.md`. Full mutant streams still include all 4070 layout documents and all 79000 numeric chunk/event cases; sampled baselines contain 4068 and 78998 respectively.

`TestSelection`, `TestUnsetIsWholeCorpus`, `TestBadChangedList` and `TestHexOffset` pass. They cover sorted deterministic selection, rotation covering every file, direct changed paths, package fixtures, fixture-adapter ownership, fixed physical controls, unset order/count preservation, malformed including empty switches, unreadable lists, repository-relative changed paths and hexadecimal rather than decimal offsets. `TestSampleRetainsFixedMarkdownInputs` proves generated inputs, numeric controls and projected oracle rows are retained/aligned.

Eight scratch source mutants were caught by behavioral assertions, without build failures or panics: dropping a changed file, dropping package fixtures, fixing the offset, accepting malformed hex, losing the unset corpus, assigning fixture adapters to the wrong package, dropping a fixed physical control, and sorting the wrong list. An accidental 42-character fake SHA also failed both JSON corpus tests by name before the corrected 40-character run. No mutated source is committed.

Initial observed wall times:

| Run | Unset | Sampled |
| --- | ---: | ---: |
| JSON port agreement | 445.68s | 29.70s |
| JSON upstream parity | 79.43s | 29.60s |
| Entire JSON package | 455.054s | 37.732s |
| Entire Markdown package | 301.313s | 199.725s |

The initial Markdown full/sample pair may include cache-warming effects; it is not evidence that deleting two physical inputs caused the whole reduction. The final warm pair measured **188.116s unset and 205.056s sampled** on the final wiring. All 17 corpus tests logged exactly once: eleven at 1/3 files, stride 9, offset 8; six at 3/3 files, stride 1, offset 0. Both modes passed the same 82 subtests. This warm comparison shows no Markdown speedup from physical-file sampling in the default three-file corpus. The entire JSON package full run measured 455.054s, with port agreement 448.37s and upstream parity 78.71s; all 2492 original inputs and the exact nine differences remained.

Commands, with the setup environment sourced, use `go test -v -count=1 -timeout 30m`. Full mode unsets both switches. Sample mode sets the fake SHA and changed-list path. JSON runs with the installed pinned Prettier 3.9.6 via `ADAMIC_JSON_PRETTIER`; Markdown runs with setup's width dependency directory. The full package commands are `go test ./stage1/cohere/json` and `go test ./stage1/cohere/markdownblocks`; initial focused JSON validation selected both sweeps and `TestThreePortMutantsAreCaught`. All original checks and mutants run in both package modes. No whole-repository gate was run.

The existing tokenizer and mdast smoke modes also passed with truthful `0/0 files` sample diagnostics and all their fixed controls and mutants. Final package vet, gofmt and `git diff --check` were clean. Scratch logs remain local; no full gate or Threadripper landing was executed.

## Generated Markdown follow-up

`Selection.Generated` enumerates inputs in their fixed construction order and selects `index % stride == offset`. Keys are indices, independent of input text and duplicate labels. The physical selector retains changed-path inclusion. Explicit generated controls are forced in, including all audit inputs and named layout edges. All 65,536 scalar checks, 9,330 special-code sequences, column checks, mutant inputs and the native canary remain full. Numeric rows are selected by their corpus positions, not by potentially duplicate labels.

The nine layout tests and `TestMicromarkInputChunks`/`TestTokenizerEvents` retain stride 9. The actual gate log gives a maximum shared-layout time of 240.79s and tokenizer events 233.04s: `ceil(240.79/30)=9`, estimated variable work `240.79/9=26.75s`; events `233.04/9=25.89s`. This arithmetic estimates corpus work, not mandatory build/control/mutant time. Default totals are 4,070 documents including physical and generated inputs. At offset 8 with the changed GAPS.md, 557 are selected; every explicit generated control remains. Scalar and sequence inputs lie outside this document count and always run. Other corpus tests retain stride 1 and all inputs.

The gate log also identifies unsampled fixed checks as slow: source decoding 190.57s, text splitting 219.56s and Unicode widths 232.22s. Those are scalar/sequence checks the brief explicitly preserves. Consequently a 30s whole-package deadline cannot be established by changing document stride alone. No fixed check is dropped to meet that estimate.

`TestGeneratedSelection` proves indexed determinism, offset rotation, controls and unset retention. Three additional scratch semantic mutants (drop generated control, fixed generated offset, drop unset generated corpus) were caught by behavioral assertions. Package vet and diff checks are clean. The preceding unset whole-package measurement on this same 4-CPU EPYC box was 188.116s; generated metadata is ignored by the protocols and full mode selects every input, preserving counts.

Final indexed sample package run passed in **222.930s**, with all 17 diagnostics and full mutants. This overlapped the preceding 227.342s validation run for part of its duration, so it is a contended measurement, not an isolated speed comparison against the prior 188.116s unset run. The 30s whole-package target remains unproved. Both runs retained 557/4070 documents at stride 9; the final run uses positional numeric selection even for duplicate labels. `TestGeneratedLayoutSelection` separately proves those duplicate labels cannot select the wrong row.
