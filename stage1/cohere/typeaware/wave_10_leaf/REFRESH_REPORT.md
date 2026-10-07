Landing: rebased wave 10 onto area/stage1-lint b84a9d931, which contains main c7991b900.
Source commit: 2a0d6cfdc36c05125c86a8552b3edf47cee865c8; evidence commit is the commit containing this report.
Commands: setup, registry generation, fresh compiler build and the explicitly filtered gates below.
Mutants: nine rule mutants plus checker, reporting and registry checks are recorded in evidence/landing-refresh.
Limits: no new claim; all 197 ranked rules are ported or claimed across 630 origin refs.

No rule implementation or shared harness file changed. The incoming lowering proof and record-runtime changes are retained. Setup reported Go, clang, Node and submodules ready in 0s each, build-cache warm in 137s, total 137s on 5 processors; cpu.max is 400000 100000. Registry generation and fresh cmd/adamic compilation succeeded. The Go bridge sources are unchanged across the landing bases, so the existing normal and sanitized checker archives were reused with the newly built compiler.

## Commands and observations

All test output was written directly to logs. The environment points ADAMIC_TYPESCRIPT_SOURCE at /workspace/wave-10-typescript and the two ADAMIC_WAVE10_*_MANIFEST variables at the frozen compiler77 and repository287 manifests. Native executables and comparison artifacts use /workspace/wave-10-refresh-*.

- go test ./stage1/cohere/typeaware -run '^TestWave10AgreementAndMutants$' -count=1 -timeout=10m -v
- go test ./stage1/cohere/typeaware/wave_10_next -run '^TestTimeoutAgreementAndMutants$' -count=1 -timeout=10m -v
- go test ./stage1/cohere/typeaware/wave_10_next -run '^TestLandingNativeRulesAndMutants$' -count=1 -timeout=10m -v
- python3 stage1/cohere/typeaware/wave_10_leaf/validate.py with fresh stage0, checker archives and refresh-isolated artifacts
- python3 stage1/cohere/typeaware/wave_10_leaf/live_validate.py with refresh-isolated inputs, refresh-live artifacts and both frozen corpora
- python3 stage1/cohere/typeaware/wave_10_leaf/require_await/validate_reporting.py with the fresh compiler and refresh-isolated/await-oracle
- go test ./bridge/tsgo/checker -count=1; go vet ./bridge/tsgo/checker
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/internal/oracle/testdata/regexp(_exec|_unicode)?[.]a$' -count=1 -timeout=10m -v
- go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestMutants|TestRegistrationMutant)$' -count=1 -timeout=30m -v
- go test ./stage1/typescript/parser ./stage1/typescript/scanner -run '^(TestCompilerExpressionsAgree|TestWholeCompilerAgrees|TestScannerAgreesWithTypescriptGo)$' -count=1 -timeout=30m -v

The first three-rule suite passed: controls 61 findings / 18551 identical bytes, compiler 9 / 8346, repository 1 / 18863, normal and sanitized. Mutants suppress the loop report (difference byte 554), redundant constituent report (4602) and includes report (8498); all exit 0 with empty stderr and are caught by the Go byte comparison. Released member-parameter queries panic at 70; an incorrectly retained registry is caught by the required panic check.

Timeout passed: 37 controls, 15 findings / 10260 bytes; compiler and repository 5318 / 18485 bytes. Normal/sanitized comparisons agree. The lost-handle rule mutant exits 0 and differs at byte 54; released runtime-context queries panic at 70, with a retained-registry mutant caught by that check.

Leaf controls pass normally and sanitized: symbol-description 2114 bytes, react-hook-no-any-type 5902, raw require-await facts 882. Their comparison-only mutants differ at bytes 66, 70 and 628; the hook-name regex mutant differs at 819. The raw-facts candidate intentionally refuses unsupported full require-await analysis at exit 70; the complete live rule is tested separately.

Live leaf comparisons include full findings, fixes and suggestions: symbol controls 2114 bytes, hook 5902, require-await 4866; each matches compiler 5318 and repository 18485. require-await additionally matches 87 extracted complete upstream source programs, 59 findings / 40041 bytes. Live rule mutants differ at 66, 70 and 2266. The incorrect union flag and self-inference mutants differ at 34397 and 35762. All run successfully with empty stderr. All three reject released handles at 70 normally and sanitized. Reporting matches Go over 32 cases / 20087 bytes normally and sanitized; its range mutant exits cleanly and differs at byte 58.

The filtered uncached RegExp oracle passed native and emitted JavaScript comparisons with Node and the one-byte mutant; cache hits were zero. The bridge package tests and vet passed. The independent cohere checker reported zero findings on the 13 owned Adamic sources.

Live compiler timings (three alternating rounds, medians in seconds): symbol native 1.863938 / Go 0.407883 = 4.570x; hook 2.348498 / 0.486465 = 4.828x; require-await 2.401966 / 0.407180 = 5.899x. Repository ratios are 1.839x, 1.680x and 1.784x. These whole-process timings include loading and serialization under concurrent gate load; they are observations, not isolated rule costs. Exact rounds are preserved in timings.json.gz.

## Interruptions and scope

The filesystem filled during the first attempt: timeout failed creating its directory, leaf mutation preparation failed with ENOSPC, and the stream harness lacked a prepared artifact directory. Obsolete owned scratch builds with already committed evidence were removed, recovering 3.2 GB. Timeout and leaf checks were rerun unchanged. The stream directory was prepared with the fresh compiler and existing bridge archives and its gate rerun. Failed-attempt logs remain beside successful logs; no check was skipped or relaxed.

The three React claims remain parked and are not certified native implementations. JSX parsing is available on this base; the earlier JSX blocker is historical. Shared RuleContext checker/program/source-path attachment remains the concrete integration gap for the nine owned type-aware listeners. The shared syntax lint comparisons do not certify that wiring. The live listener runner's emitted-JavaScript comparison, the full repository gate and the complete 17 required-input checks are outside this filtered gate; absent external postcss/graphql/prettier inputs are not claimed green. No rule remains available in the global ranked inventory, so no new claim was made.

## Final gate results

All listed selected checks completed successfully. Original suite 195.613s; timeout retry 72.604s; prepared stream gate 207.768s; isolated, live and reporting scripts exit 0; bridge 0.282s; uncached RegExp oracle 9.929s; shared lint 496.165s; parser 45.658s; scanner 31.934s. No selected test skipped.

Stream controls match Go normally and sanitized: process 92 sources, 77 findings / 45901 bytes; blocking 92 sources, 30 findings / 24200 bytes. Each matches compiler 5318 and repository 18485 bytes. The process chain-suppression mutant differs at byte 114 and the blocking-state mutant at 13904, both exit 0 with empty stderr.

Shared lint witness comparisons match 13044077 bytes; the 394-file compiler/stage1 comparison matches 21138417. Existing explicit unsupported-recovery cases retain their panic assertions; their reported limits are not newly excluded. All 15 existing registry mutants are caught by Node, emitted JavaScript and native comparisons; TestRegistrationMutant passes. The exact names and mismatches are preserved in the compressed lint log.

Parser parity matches 77 compiler files: 28836875 expression-tree bytes and 44766682 whole-tree bytes. Scanner parity matches 77 compiler, 280 stage1 files and 18236 generated inputs / 30671195 bytes. Its punctuation, invalid-decimal-separator and omitted-regex-rescan mutants are caught by both Node and native.

Stream compiler medians: process-compiler native 1.792510s / Go 0.420656s = 4.261x; blocking-compiler native 2.443199s / Go 0.332944s = 7.338x; whole-process observations under concurrent gate load. Exact rounds are in stream-timings.json.
