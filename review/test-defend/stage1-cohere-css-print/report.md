# CSS printer defense

Base: 955e3eb92b9cd04aca420974d1006048b4615d51. Audit reference: origin/test-audit/stage1-cohere-css-print, original audit base cf79ecec3723604428ab91ebcb283400d05a1548. All 514 current top-level names matched the audit list; no added or vanished rows. Tests and oracles were never edited. Production switches were restored before the final clean check.

## Code under test and oracle, stated before mutation

- Optional boolean: the CSS TypeScript printer port, exercised under Node, Adamic's ASan/UBSan native backend, JavaScript backend and leak check. Live Go cohere supplies complete expected strings. external-run.
- Snapshots: the same port in a saved source snapshot and release, profiled, counted native builds, default and narrow options. Live Go cohere supplies complete expected strings. external-run.
- Artifacts: production native.Build, Flags and ValidateOptions, plus native compilation of the port. The row generates a corpus and files, but its verdict only requires construction to succeed. self for successful construction; Go cohere and pinned Prettier/PostCSS supply corpus data without validating the generated printer's answers here.

## Coverage and semantic differences

Go coverage cannot instrument this TypeScript port or emitted C. Per-test -coverpkg=./internal/native profiles record production builder activity. Snapshot builder coverage is zero because its binaries are prepared by the artifact row. Production V8 coverage records the Node side of the port; reached-port-functions.json lists observed functions. Optional has nine exclusive observed ranges in parsing/dependencies, but no exclusive printer ranges. Its six concrete inputs isolate namespace and boolean flags; D2, D3 and D4 target those semantics on shared printer lines.

Artifact coverage reaches the Count flag branch absent from optional coverage. Count:true has exactly one package caller, TestCSSProfileArtifacts. Snapshots alone execute that counted product and compare its stdout. The throughput checker builds without Count. D1 changes production admission of counted builds; D8 changes runtime reporting from fd 2 to fd 1, exposing counted stdout contamination. Both are menu mutations, not test/harness edits.

Throughput coverage is partial because its corrected baseline timed out. v8-snapshots-vs-throughput-PARTIAL.json is an observation, not proof of exclusive coverage. Native runtime report coverage was established through counted vs uncounted product semantics and caller search, not a C coverage profile.

## Results and limits

D1 is caught only by Artifacts in its four-row run. D8 is caught only by Snapshots across separately bounded runs; Artifacts and three other rows pass. CSSPrinterBoundaryProofs completed its Go/Node proof but skipped the external Prettier portion because its opt-in was unset in that supplemental run. counted-callers.txt explains why remaining package rows do not request or execute counted builds. These are bounded defenses. A whole-package pass is not claimed, and central replay must settle other interactions.

D2, D3 and D4 each fail Optional and Snapshots; ClosedPrinterRegexGap and CSSPrinterBoundaryProofs pass each. Three honest attempts did not defend Optional. This is evidence from three semantic mutations, not permission to delete it. Its name promises agreement for its six cases and its assertions check exact output across three executors plus leaks. No name/assertion gap was found for that row.

Artifacts is defended for counted-build admission, despite its artifact-validation gap: it does not check that expected.txt exists after generation, validate manifest contents, or execute the printer. The audit's S3 harness mutation was not repeated, because harness edits are forbidden in this defense. Snapshots is defended for exact stdout from its counted build. Throughput checks a count and aggregate UTF-16 length, and logs speeds without enforcing a time threshold; its name alone does not establish a performance gate.

## What cost time or was unclear

1. The disk rule asks for 15 GB free on an 8.8 GB /tmp filesystem. That threshold is impossible here. Authorized earlier-unit scratch/cache removal raised free space from 3.3 to 5.6 GB; /workspace remained about 11 GB. No repository or tools were deleted. Unknown bootstrap/corpus directories were retained. No ENOSPC occurred.
2. Warm tools did not include the Node packages. npm ci ran in stage3/api, and pinned Prettier 3.9.6, PostCSS 8.5.16 and postcss-scss 4.0.9 were installed in scratch.
3. I initially supplied the Prettier package directory rather than its installation root to the throughput test. That caused a dependency-resolution failure. I corrected it before relying on results. The corrected throughput run exceeded 90 seconds. The three target rows had clean individual baselines.
4. Whole-package testing exceeded 90 seconds in a memory-check test. It was stopped and narrowed. The 514 names include large generated families and built-in witnesses; no whole-package mutation matrix is claimed.
5. The prescribed Go coverage command cannot measure a TS port or compiled C. Native-builder Go profiles, production-only V8 traces and static counted callers are saved, with limits stated rather than inferred coverage claims.
6. Snapshot inputs require prebuilt binaries and source copies. D8 rebuilt all three products before the snapshot run. The selector product was built once, then its counted binary rebuilt after restoring D8. A neutral selector baseline passed before D2-D4. Every attempted standalone diff applies to the recorded base. D1 passes go vet using its standalone source overlay; D2-D4 each pass a fresh native CLI build; D8 passed the artifact row's actual clang builds with the changed runtime.
7. Three optional-row attempts were enough to show the named subsumer caught each; the original planned snapshot mutations D5-D7 were not attempted after the distinct counted-runtime defense succeeded. They are retained only as planning entries, not replay diffs or evidence of kills.
8. Two artifact builds briefly overlapped during preparation, adding CPU contention. Each completed within 90 seconds. Matrix selector runs were sequential. Individual test-binary and build times are saved, since summed durations would double-count overlapping work.

## Reproduction

Source /workspace/adamic-tools/env.sh. Install the pinned libraries and export ADAMIC_CSS_PRINTER_LIBRARY=/tmp/css-defense/library. Every test invocation redirects output to a log. For D1 apply D1.diff and run the four-row command in rows.json without ADAMIC_DEFENSE_MUTANT. For D8 apply D8.diff, run TestCSSProfileArtifacts with a fresh ADAMIC_CSS_PROFILE_DIR, then TestCSSProfileSnapshotsAgree with that directory. Runtime mutation must precede all native builds. For standalone D2-D4 apply one diff at a time, prepare fresh profile artifacts, then run the four-row regex in rows.json. Each mutant uses its own ADAMIC_BUILD_CACHE_DIR. switch.diff preserves the session's selector implementation if replaying its exact matrix instead.

## Timing

nproc: 5. Warm env worked; cloud/setup.sh skipped. Original target clean baselines: Optional 34.586 s, Artifacts 64.720 s, Snapshots 55.994 s. Whole package 90.023 s, corrected throughput 90.047 s, both cooked. D8 artifact rebuild 85.931 s; selector artifacts 72.522 s. Standalone native D2/D3/D4 rebuilds: 20.807/19.292/20.407 s. Other per-run times are in test-binary-timings.json and switched-counted-rebuild.json. Session elapsed approximately 29 minutes including disk cleanup, dependency installation, reads, coverage and publication; npm/setup probe wall timings were not independently recorded. No other packages, C instrumentation, throughput performance defense or repo-wide uniqueness replay was attempted.
