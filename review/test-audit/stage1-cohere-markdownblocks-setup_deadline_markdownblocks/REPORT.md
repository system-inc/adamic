u129 audited 24 discovered names as ten grouped rows at d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb.
Whole-package and full-slice cold baselines cooked at 90 seconds; every individual grouped clean row passed three times.
Nine rows are setup-checks; the deadline worker is a helper for the deadline setup check.
Nine construction mutants were caught; eight product rows passed their own empty-entry probes.
Standalone diffs, switched scratch evidence, matrices and all commands are retained; source was restored and the final slice passed.

```json
[
  {
    "test": "TestMarkdownLayoutSetupHasNoDeadline",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 0.034,
    "oracle": "self: no setup deadline, subprocess must end with its work deadline, and setup context must remain uncanceled.",
    "oracle_kind": "self",
    "kills": [
      "S1"
    ],
    "unique_kills": [],
    "last_proven_fail": "S1: setup_deadline_markdownblocks_test.go:31: setup has deadline 2026-10-09 12:53:05.067083033 +0000 UTC m=+0.004664484",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownLayoutSetupHasNoDeadline"
    ],
    "evidence": "ADAMIC_MUTANT=S1 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownLayoutSetupHasNoDeadline)$; S1: setup_deadline_markdownblocks_test.go:31: setup has deadline 2026-10-09 12:53:05.067083033 +0000 UTC m=+0.004664484",
    "members": [
      "TestMarkdownLayoutSetupHasNoDeadline"
    ]
  },
  {
    "test": "TestMarkdownLayoutDeadlineWorker",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 0.008,
    "oracle": "Subprocess entry activated by ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER=1; parent checks its termination.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "No construction matrix: subprocess helper for TestMarkdownLayoutSetupHasNoDeadline.",
    "members": [
      "TestMarkdownLayoutDeadlineWorker"
    ],
    "parent": "TestMarkdownLayoutSetupHasNoDeadline"
  },
  {
    "test": "TestProduct_MarkdownQuote_build family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 2.386,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S2"
    ],
    "unique_kills": [],
    "last_proven_fail": "S2: quote_layout_shards_test.go:230: open /tmp/u129/cache/S2/8b1c72f093ae5c041802a9117fc1aad459aca7e9ffe0894742de102602015b32/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownQuote_build family",
      "TestProduct_MarkdownQuote_native family"
    ],
    "evidence": "ADAMIC_MUTANT=S2 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownQuote_lowered|TestProduct_MarkdownQuote_lists|TestProduct_MarkdownQuote_layout|TestProduct_MarkdownQuote_native_true|TestProduct_MarkdownQuote_native_false)$; S2: quote_layout_shards_test.go:230: open /tmp/u129/cache/S2/8b1c72f093ae5c041802a9117fc1aad459aca7e9ffe0894742de102602015b32/program.c: no such file or directory",
    "members": [
      "TestProduct_MarkdownQuote_lowered",
      "TestProduct_MarkdownQuote_lists",
      "TestProduct_MarkdownQuote_layout"
    ]
  },
  {
    "test": "TestProduct_MarkdownQuote_native family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 1.92,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S2",
      "S3"
    ],
    "unique_kills": [],
    "last_proven_fail": "S3: quote_layout_shards_test.go:628: build markdown-quote-layout-native-false: native: clang failed: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 2,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownQuote_build family",
      "TestProduct_MarkdownQuote_native family"
    ],
    "evidence": "ADAMIC_MUTANT=S3 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownQuote_native_true|TestProduct_MarkdownQuote_native_false)$; S3: quote_layout_shards_test.go:628: build markdown-quote-layout-native-false: native: clang failed: exit status 1",
    "members": [
      "TestProduct_MarkdownQuote_native_true",
      "TestProduct_MarkdownQuote_native_false"
    ]
  },
  {
    "test": "TestProduct_MarkdownStructure family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 2.559,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S4"
    ],
    "unique_kills": [],
    "last_proven_fail": "S4: structure_layout_shards_test.go:186: open /tmp/u129/cache/S4/0b4a7cefba37501053bf67fc3a21055c777fa06d8cd85d8081c6ca3cd544d539/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownStructure family"
    ],
    "evidence": "ADAMIC_MUTANT=S4 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownStructure_lowered|TestProduct_MarkdownStructure_native|TestProduct_MarkdownStructure_adamic_markdown_lists|TestProduct_MarkdownStructure_adamic_markdown_doclayout)$; S4: structure_layout_shards_test.go:186: open /tmp/u129/cache/S4/0b4a7cefba37501053bf67fc3a21055c777fa06d8cd85d8081c6ca3cd544d539/program.c: no such file or directory",
    "members": [
      "TestProduct_MarkdownStructure_lowered",
      "TestProduct_MarkdownStructure_native",
      "TestProduct_MarkdownStructure_adamic_markdown_lists",
      "TestProduct_MarkdownStructure_adamic_markdown_doclayout"
    ]
  },
  {
    "test": "TestProduct_MarkdownTable family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 3.373,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S5"
    ],
    "unique_kills": [],
    "last_proven_fail": "S5: setup_deadline_markdownblocks_test.go:109: open /tmp/u129/cache/S5/cb82c2cbde51144e483a8df0cb8f632d46111c871d26b7c570a431d54034c56a/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownTable family"
    ],
    "evidence": "ADAMIC_MUTANT=S5 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S5 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownTable_lowered|TestProduct_MarkdownTable_native_true|TestProduct_MarkdownTable_native_false|TestProduct_MarkdownTable_list|TestProduct_MarkdownTable_document)$; S5: setup_deadline_markdownblocks_test.go:109: open /tmp/u129/cache/S5/cb82c2cbde51144e483a8df0cb8f632d46111c871d26b7c570a431d54034c56a/program.c: no such file or directory",
    "members": [
      "TestProduct_MarkdownTable_lowered",
      "TestProduct_MarkdownTable_native_true",
      "TestProduct_MarkdownTable_native_false",
      "TestProduct_MarkdownTable_list",
      "TestProduct_MarkdownTable_document"
    ]
  },
  {
    "test": "TestProduct_MarkdownWhitespace_build family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 2.567,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S6"
    ],
    "unique_kills": [],
    "last_proven_fail": "S6: whitespace_layout_split_test.go:270: open /tmp/u129/cache/S6/894ac7f6b20f8479fa74d3eca77db1a6645d63d36b69653a43f97ffcc5bd0ebf/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownWhitespace_build family",
      "TestProduct_MarkdownWhitespace_manifest",
      "TestProduct_MarkdownWhitespace_native family",
      "TestProduct_MarkdownWhitespace_policy_native"
    ],
    "evidence": "ADAMIC_MUTANT=S6 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S6 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownWhitespace_lowered|TestProduct_MarkdownWhitespace_lists|TestProduct_MarkdownWhitespace_layout|TestProduct_MarkdownWhitespace_policy_lowered|TestProduct_MarkdownWhitespace_native_true|TestProduct_MarkdownWhitespace_native_false|TestProduct_MarkdownWhitespace_policy_native|TestProduct_MarkdownWhitespace_manifest)$; S6: whitespace_layout_split_test.go:270: open /tmp/u129/cache/S6/894ac7f6b20f8479fa74d3eca77db1a6645d63d36b69653a43f97ffcc5bd0ebf/program.c: no such file or directory",
    "members": [
      "TestProduct_MarkdownWhitespace_lowered",
      "TestProduct_MarkdownWhitespace_lists",
      "TestProduct_MarkdownWhitespace_layout",
      "TestProduct_MarkdownWhitespace_policy_lowered"
    ]
  },
  {
    "test": "TestProduct_MarkdownWhitespace_native family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 1.959,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S6",
      "S8"
    ],
    "unique_kills": [],
    "last_proven_fail": "S8: whitespace_layout_split_test.go:888: native: linking units: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 2,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownWhitespace_build family",
      "TestProduct_MarkdownWhitespace_manifest",
      "TestProduct_MarkdownWhitespace_native family",
      "TestProduct_MarkdownWhitespace_policy_native"
    ],
    "evidence": "ADAMIC_MUTANT=S8 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S8 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownWhitespace_native_true|TestProduct_MarkdownWhitespace_native_false|TestProduct_MarkdownWhitespace_manifest)$; S8: whitespace_layout_split_test.go:888: native: linking units: exit status 1",
    "members": [
      "TestProduct_MarkdownWhitespace_native_true",
      "TestProduct_MarkdownWhitespace_native_false"
    ]
  },
  {
    "test": "TestProduct_MarkdownWhitespace_policy_native",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 1.645,
    "oracle": "self: successful product construction and artifact reads; Go oracle executables are built but not run or compared. Empty helper return is not asserted by this wrapper.",
    "oracle_kind": "self",
    "kills": [
      "S6",
      "S9"
    ],
    "unique_kills": [],
    "last_proven_fail": "S9: whitespace_layout_split_test.go:888: native: linking units: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 2,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownWhitespace_build family",
      "TestProduct_MarkdownWhitespace_manifest",
      "TestProduct_MarkdownWhitespace_native family",
      "TestProduct_MarkdownWhitespace_policy_native"
    ],
    "evidence": "ADAMIC_MUTANT=S9 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S9 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownWhitespace_policy_native|TestProduct_MarkdownWhitespace_manifest)$; S9: whitespace_layout_split_test.go:888: native: linking units: exit status 1",
    "members": [
      "TestProduct_MarkdownWhitespace_policy_native"
    ]
  },
  {
    "test": "TestProduct_MarkdownWhitespace_manifest",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go",
    "seconds": 1.129,
    "oracle": "self: deserialize the cached product manifest, read backend files, and require nonempty Go binary path; no executable or formatting output is compared.",
    "oracle_kind": "self",
    "kills": [
      "S6",
      "S7",
      "S8",
      "S9"
    ],
    "unique_kills": [],
    "last_proven_fail": "S9: whitespace_layout_split_test.go:117: build markdown-whitespace-layout-setup-v2: whitespace policy setup failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownWhitespace_build family",
      "TestProduct_MarkdownWhitespace_manifest",
      "TestProduct_MarkdownWhitespace_native family",
      "TestProduct_MarkdownWhitespace_policy_native"
    ],
    "evidence": "ADAMIC_MUTANT=S9 ADAMIC_BUILD_CACHE_DIR=/tmp/u129/cache/S9 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestProduct_MarkdownWhitespace_policy_native|TestProduct_MarkdownWhitespace_manifest)$; S9: whitespace_layout_split_test.go:117: build markdown-whitespace-layout-setup-v2: whitespace policy setup failed",
    "members": [
      "TestProduct_MarkdownWhitespace_manifest"
    ]
  }
]
```

| ID | File:line at starting origin/main | Change | Failed grouped rows | Wall seconds |
|---|---|---|---|---|
| S1 | stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go:14 | `return context.WithCancel(parent)` to `return context.WithTimeout(parent, time.Nanosecond)` | deadline | 5.822 |
| S2 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:223 | `filepath.Join(dir, "program.c")` to `filepath.Join(dir, "missing.c")` | quote build, quote native | 35.707 |
| S3 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:628 | `filepath.Join(dir, "port")` to `filepath.Join(dir, "")` | quote native | 56.792 |
| S4 | stage1/cohere/markdownblocks/structure_layout_shards_test.go:179 | `filepath.Join(dir, "program.c")` to `filepath.Join(dir, "missing.c")` | structure | 35.994 |
| S5 | stage1/cohere/markdownblocks/table_layout_shards_test.go:169 | `filepath.Join(directory, "program.c")` to `filepath.Join(directory, "missing.c")` | table | 35.819 |
| S6 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:263 | `filepath.Join(dir, "program.c")` to `filepath.Join(dir, "missing.c")` | whitespace build, whitespace native, policy native, manifest | 37.138 |
| S7 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:137 | `filepath.Join(directory, "products.json")` to `filepath.Join(directory, "missing.json")` | manifest | 40.622 |
| S8 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:900 | `filepath.Join(dir, "port")` to `filepath.Join(dir, "")` | whitespace native, manifest | 39.776 |
| S9 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:691 | `filepath.Join(dir, "port")` to `filepath.Join(dir, "")` | policy native, manifest | 40.382 |

Survivors

None of the nine construction mutants survived the selected rows. No production port mutants were planted. P2 through P9 are passing empty-answer probes, never mutants or semantic survivors. P1 fails by nil-context panic, and its sole selected row was rerun alone under the standalone diff and failed again.

Code under test and oracle

These tests check the suite's construction: cancelable setup contexts, process-group child execution, cached lowering and product building, native build setup, Go overlay executables and the whitespace product manifest. Their oracle is self: handwritten context assertions and successful construction/error handling. Building a Go cohere oracle executable does not make a row external-run: none of these product wrappers executes it or compares formatting output. The production .ts port and the Go oracle were left untouched. Construction helper edits are the brief's explicit setup-check exception to the ban on harness edits.

Shared checkers define the eight product families, whose complete members are in rows.json. Quote build and quote native are separate because native wrappers also call quoteLayoutNativeProduct; whitespace build and native similarly differ. Structure's four wrappers and table's five wrappers each call their respective one checker with a different recipe. Whitespace build includes policy_lowered because only its probe argument differs. policy_native and manifest call different functions and remain independent rows. The default deadline worker returns immediately; the parent alone activates its one-hour sleep, and bounds it with a 25ms work context.

All 24 requested names were present, at the expected file. None moved or vanished. There were no observed skips in the requested rows. The whole-package timeout occurred before default optional width-oracle skipping could be observed. Width dependencies were subsequently installed under /tmp/u129/width-deps, but TestMarkdownUnicodeWidths is outside this slice and was not run. No claim that its oracle was exercised is made. The selected product rows do not load any node_modules directory; npm ci stage3/api ran before the baseline anyway.

Function inventory and matrix bounds

Read and inventoried construction functions: markdownLayoutSetupContext, markdownLayoutProductRoot, tableLayoutCommand, quoteLayoutBuildProduct, quoteLayoutNativeProduct, quoteLayoutSetupCommand, quoteLayoutContextCommand, structureLayoutBuildProduct, structureLayoutCommand, tableLayoutBuildProduct, tableLayoutBuildFiles, whitespaceLayoutBuildProduct, whitespaceLayoutBuildProbe, whitespaceLayoutBuildFiles, whitespaceLayoutSetupInputs, whitespaceLayoutSetup, whitespaceLayoutNativeSetup, whitespaceLayoutNativeProduct, whitespaceLayoutNativeBuild, whitespaceLayoutPolicySetup, whitespaceLayoutReadyProducts, whitespaceLayoutPrepare, whitespaceLayoutReadProducts, whitespaceLayoutCommand and loweredResult. The conservative original-source lexical call graph, with original file:line locations and candidate callers throughout the package, is in static-function-inventory.json. That supplemental graph was produced after choosing mutations; it is not dynamic coverage and cannot resolve callbacks or external library internals exhaustively. The pre-mutation inventory was the directly read helper set, not a complete compiler call graph.

After the whole package and the full requested slice cooked, clean baselines and three-run timings ran each grouped row independently. Every one passed. Mutation matrices then selected the requested rows that reach each changed helper. matrix.json records exact selected top-level members, grouped results and unobserved cells. A row not selected is unknown, not a survivor. This is why bounded=true and unique_kills=[]: no package uniqueness claim rests on these selectively run construction faults. Setup-check verdicts rest on their own observed construction failures, not subsumption or production semantic coverage.

Every mutant came from the declared menu before results: change a context option or an artifact/output-path constant. No assertion or expected value was edited. The scratch switch and empty-return guards are instrumentation, not standalone mutations; S1 through S9 diffs contain just their individual construction changes. Per-mutant ADAMIC_BUILD_CACHE_DIR forces constructors to execute instead of accepting an old cached product. Every pure diff applies to the starting source and passed go vet ./stage1/cohere/markdownblocks/. The switch also passed vet. No .a or .ts source, compiler, Go cohere oracle, dispatch or corpus was changed. No native or compiler mutant rebuild cap is invoked because all nine are Go construction mutations, not stage1 port mutations.

Vacuity and oracle strength

P2-P9 return empty product structs or empty path strings at each row's own construction entry. Every product row still passes: these wrappers discard the returned product and rely entirely on errors raised inside construction. The probes do not remove or change the wrappers' assertions. For the manifest row, its assertion lives inside whitespaceLayoutReadyProducts, so returning at that construction entry bypasses it and the wrapper notices nothing. These are construction vacuity findings, not proof that production differential rows are vacuous. P1 returns a nil setup context; the explicit context assertion panics, so the deadline row's vacuous value is false. The helper was not probed and has null.

Artifact-name mutations prove that missing generated C and missing manifests can fail construction. Native output-path mutations fail the linker because its output names an existing directory. These are appropriate setup failures and are not credited as production lint or formatting semantic kills. Sanitizer executables being built does not prove sanitizer cleanliness. Most cached rows only read or construct products; their timing median measures those warm conditions, not cold compilation.

Evidence locations

Failure locations in rows.json are mapped exactly from switched source lines to the original commit using the retained switch.diff. Raw logs preserve observed scratch line numbers; the standalone diffs contain original locations. matrix commands are retained in matrix-timings.json and rows.json. The fixed mutant plan is plan.json; measure.py, mutate.py, validate.py and report.py retain the session's orchestration. P1-alone.log records the panic rerun. Restored-slice-baseline.log is the final clean 24-name run.

Brief ambiguities and costs

The brief calls these 24 functions rows, but its family rule groups the wrappers into eight product rows, plus the deadline check and helper. Treating them as 24 independent rows would contradict that rule. Group costs were measured with one Go binary invocation selecting all family members, repeated three times; individual parent PASS lines under parallel scheduling would not measure the whole family.

The setup-check exception is essential here. A formatting mutant would not test construction correctness; these rows prepare artifacts and leave behavior comparisons to other rows. Nine construction faults are fewer than the suggested approximately three per meaningful row, but cover every non-helper row and remain below the 20-mutant cap. No deletion or redundancy recommendation follows from that small construction set.

The prescribed whole-package baseline spends 90 seconds in earlier preparation before the requested rows can all complete. The full 24-name slice also cooks cold with simultaneous compilation. Individual families passed once preparation products became cached. This required two cooked baselines and subsequent narrower runs. Neither timeout was treated as red, and neither establishes a green whole-package baseline. No real test failure preceded either timeout.

The phrase entry of code under test needs adaptation for these construction rows. Each probe targets the specific construction helper called by its row, not Lower or the port's runtime main. Preparation beneath those helpers, native.Build internals and behavior checks outside this slice were not separately probed. The eight vacuity findings are therefore explicitly construction-specific.

The inventory requirement can expand to the compiler, buildcache, operating system and Go build internals. A complete dynamic call inventory was not collected. The original-source lexical inventory is conservative supporting evidence, not a claim of exhaustive runtime reach. Its original-location mapping corrects the initial scratch inventory, which included the instrumentation helper functions.

Timing and coverage limits

Warm tools worked and setup was skipped, 0 seconds; nproc=5. npm ci took 0.397 seconds and discovery 5.789 seconds. Whole-package baseline wall time was 91.888 seconds (binary 90.050); the full requested cold slice also hit its 90-second binary budget. Group timing medians are in the JSON rows. Each construction matrix's wall time, including necessary lowering and native/Go construction, is listed in the mutant table. There is no separately measured native.Build callback stopwatch, so no fabricated per-native-product rebuild time is reported. build-cache observations remain in the raw logs, and command totals are in timing-summary.json. Detailed probe, vet and restoration timings are retained.

All source is restored. No tests in other packages, repository-wide uniqueness replay, production semantic mutants, complete gate or PR were attempted. Evidence stayed outside the checkout during tests so the corpus scanner could not ingest the new audit artifacts.
