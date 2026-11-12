# Reproduction

Use main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205` in `/workspace/orm-main` and the unmodified adapted tree `/workspace/orm-adapted` from `overload-results-main`. The report branch parent is `2378688b0c61c8e54fdc5ef89da86b503b52901d`. Original upstream source, submodule, overlay and build provenance are in the preceding main census. No main advance, merge or adaptation landing is part of this report.

Set `scripts=/workspace/op2-report/stage3/census/create-builder-program` and source `/workspace/adamic-tools/env.sh`. All potentially long commands below are independent children of `run.py`, which reports the child's own exit, terminates at 1,740 seconds with at most ten seconds to kill, and propagates its status. Shell wrappers report/propagate that status too. Do not edit a running shell wrapper. Only this directory is committed.

## Prepare and inspect

Run `python3 -B "$scripts/prepare.py"`. It verifies all 747 original input hashes and copies only src and package.json into two scratch trees, `/workspace/cbp-checked` and `/workspace/cbp-typed-old`. It retains exact source patches and changed-file hashes. The original tree remains untouched. The checked shape changes two files; typed-old changes only builder.ts. All other source/input hashes match the original manifest.

For each of `base`, `checked`, and `typed-old`, run this as a separate bounded child:

```sh
python3 "$scripts/run.py" "$scripts/$label-inspect.log" "$scripts/$label-inspect.time.json" node "$scripts/inspect.cjs" "$label"
rc=$?
echo "$label inspection child_exit=$rc"
exit "$rc"
```

The block is one standalone command/session. `inspect.cjs` checks all 79 recursive .ts compiler roots with stock TypeScript 6.0.3 and main's bundled Node 25.3.3 index. It records the options, complete diagnostics, every relevant identifier/call, selected overloads, every return in the whole owner, and actual whole-program emit hashes for 79 JavaScript files. Base additionally executes the entire extracted adapted createBuilderProgram in a VM against real stock builder state, and stock-checks cast-free counterexamples via a virtual source overlay. No probe is written to the original tree.

Both final shapes have zero diagnostics. An initial typed-old token replacement accidentally produced THost; its inspection is retained under typed-old-initial-* and is superseded by the corrected script/source/hash. It is not a third shape or a final diagnostic result.

## Nine-unit census

The first-error latent driver ignores LATENT_ONLY_FILE, which is supported by full mode. Two initial unsplit full-project pilots were therefore deliberately terminated; their child -15 / shell 143 outcomes and partial streams are retained under pilot-*. They supply no final count.

Build the targeted selector:

```sh
python3 -B "$scripts/select.py"
# From /workspace/orm-main:
python3 "$scripts/run.py" "$scripts/selector-build.log" "$scripts/selector-build.time.json" go build -buildvcs=false -overlay=/workspace/cbp-tool/overlay.json -o /workspace/cbp-tool/entry-census ./stage3/census/latent/tool
rc=$?
echo "selector build child_exit=$rc"
exit "$rc"
```

The selector layers over the validated scan-selector overlay from overload-results-main. Only latentRun's outer measurement loops change: honor the file selector and retain top-level declarations named createBuilderProgram, createSemanticDiagnosticsBuilderProgram or createEmitAndSemanticDiagnosticsBuilderProgram. All dependency files, registration, program facts, eligibility and normal first-error lowering remain available. There is no filter on findings within a selected unit. This is exactly nine units. Selected Go evidence is retained as `.go.txt`; compiler binaries and scratch `.go` inputs are outside the report.

Run each label with `bash "$scripts/measure.sh" "$label"`, at most two children concurrently. It clears inherited LATENT_/CENSUS_ settings, sets LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, CENSUS_SKIP_SCAN=1 and the two-file selector, and roots at the selected tree's src/tsc/tsc.ts. Repeated project-wide refusal scans are omitted; no scan delta is claimed. Both no-output guards remain enforced by the unchanged entry driver. Each run retains full header, selected file/unit/finding records, relation report and all nine per-unit timings.

The final selected baseline's nine full finding records are exactly equal to the createBuilderProgram observations in overload-results-main/overload-findings.json. Both candidates have zero checker diagnostics and exactly nine attempted unit records. Corresponding units are keyed by source path, name and ordinal within the overload set, because the source changes. No same-input source-hash pair is claimed.

## API and pins

Set `NODE_PATH=/workspace/orm-main/stage3/api/node_modules`, then run api-pins.cjs through run.py, retaining api-pins.log and api-pins.time.json. It checks literal-name presence in stock typescript.d.ts and sanctioned-api.json, compares all public builder interfaces and factory overload texts, and uses adaptation 99x's own engine to recompute its exact nested owner hashes on every tree. All nine checks match the after pins.

The retained grep commands and their actual exit codes are in grep-results.json. They search the entire original checkout for createBuilderProgram, all stage3/adapt for createBuilderProgram or builder.ts, and all stage3/adapt for the additional affected type/helper/public-owner names. Evidence line previews are limited to 500 columns to avoid copying huge minified proof records; file and line remain available for full-text inspection.

## Report and verification

```sh
python3 "$scripts/run.py" "$scripts/report.log" "$scripts/report.time.json" python3 -B "$scripts/report.py"
rc=$?
echo "report validation child_exit=$rc"
exit "$rc"
```

The report generator checks the pristine pinned main, all 747 original hashes, every candidate changed-file hash, all 79-root diagnostics, exact nine-unit baseline reproduction, unit correspondence and timings, no checker rejection, complete entry graph, no global finding, output hashes, API/owner checks and all final child statuses/bounds. It produces comparisons.json, emitted-comparison.json, validation.json and REPORT.md, including every exact new reason and all emitted hunks. Changed JavaScript before/after bytes are gzip-retained in emitted/. Full 79-file emit manifests are embedded in each inspection JSON; all emitted files can be regenerated with inspect.cjs. Large or historical evidence is clearly named and never contributes a partial result to a final table.

The local .gitattributes preserves original CRLF source evidence and unified-diff context prefixes as data. REPORT.md uses LF. Whitespace checking remains active for scripts and report prose.
