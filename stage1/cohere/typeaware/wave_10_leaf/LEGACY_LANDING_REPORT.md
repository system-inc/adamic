Landing: wave 10 rebased onto area/stage1-lint b46914832, containing current main c7991b900.
Source SHA: ec27e7a0a0d4b0ab503848cc718f34174549be86; the evidence commit contains this report.
Checks: all nine native rule suites re-green; setup 55s, nproc 5, cpu.max four cores.
Mutants: all nine clean-exit rule mutants caught by the full Go byte comparison; additional checks listed below.
Scope: no new rule claim; 638 origin refs contain claims covering every ranked unported rule.

The incoming change moves legacy syntax rules into the shared registry. cmd, internal, bridge and stage1/typescript sources are unchanged relative to the previous validated area base b84a9d931. All incoming shared context, harness and rule-directory changes are retained; no shared source file was edited by wave 10. The registry now lists 40 rules. bash cloud/setup.sh reports Go, clang, Node and submodules ready in 0s each, cache warm 55s, total 55s on five processors (cpu.max 400000 100000). go run ./cmd/lint-registry and fresh cmd/adamic build succeed. Existing normal/sanitized checker archives are reused because their Go sources did not change.

## Owned gates

The commands from REFRESH_REPORT.md were repeated with /tmp/wave10-legacy-*.log output and /workspace/wave-10-refresh-* artifacts. Both corpus manifests and ADAMIC_TYPESCRIPT_SOURCE were supplied explicitly. Every comparison includes findings, fixes and suggestions; controls and both corpora run normally and sanitized. Tests write directly to logs, without output pipelines.

- TestWave10AgreementAndMutants PASS 143.107s: controls 61 findings / 18551 bytes; compiler 9 / 8346; repository 1 / 18863. Loop report mutant differs at byte 554, redundant constituent mutant 4602, includes mutant 8498; all exit 0 with empty stderr. Released member-parameter query rejects at 70. A retained-registry mutant is caught by the required panic check.
- TestTimeoutAgreementAndMutants PASS 72.174s: 37 controls, 15 findings / 10260 bytes; compiler 5318 and repository 18485 bytes. Lost-handle mutant differs at byte 54. Released runtime-context query rejects at 70; retained-registry mutant is caught.
- TestLandingNativeRulesAndMutants PASS 191.757s: process 92 controls, 77 findings / 45901 bytes; blocking 92 controls, 30 / 24200. Each corpus matches 5318 / 18485 bytes, normally and sanitized. Process chain-suppression mutant differs at byte 114; blocking-state mutant 13904. Both exit 0 with empty stderr and are caught only by byte comparison.
- validate.py exits 0: symbol controls 2114 bytes, hook 5902, raw await facts 882. Mutants differ at 66, 70, 628 and hook regex mutation 819; all run successfully with empty stderr. The raw await candidate retains its explicit refusal; complete require-await is held by the live gate.
- live_validate.py exits 0: symbol controls 2114 bytes, hook 5902, complete require-await 4866. Each matches compiler 5318 and repository 18485. Require-await matches 87 extracted upstream complete programs, 59 findings / 40041 bytes. Rule mutants differ at 66, 70 and 2266; wrong union flag 34397; self-inference 35762. All three released-handle checks reject at 70 normally and sanitized.
- validate_reporting.py exits 0: 32 candidates / 20087 bytes, normally and sanitized. Range mutant exits cleanly and differs at byte 58.
- go test ./bridge/tsgo/checker -count=1 and go vet ./bridge/tsgo/checker pass. The independent production cohere checker reports zero findings on the 13 owned Adamic sources.

Live compiler whole-process medians (seconds, three alternating rounds): symbol native 2.007495 / Go 0.402413 = 4.989x; hook 1.583629 / 0.467654 = 3.386x; require-await 2.050288 / 0.401380 = 5.108x. Repository ratios: 1.655x, 1.443x, 1.938x. Concurrent gate load was uncontrolled; these are measured loading-and-serialization costs, not isolated rule costs. Detailed live timing rounds and stream timings are retained with the evidence.

## Limits

The three React claims remain parked and are not certified native implementations; JSX parsing is now available, so the historical JSX blocker is not asserted as current. The shared RuleContext still has no checker/program/source-path attachment, which blocks routing the nine owned type-aware listeners through the common registry. This is left to its shared-harness owner. The live listener runner's emitted-JavaScript oracle and full repository gate are not run. Filtered supplied-input lint/parser/scanner checks are not the full 17 required-input gate; external postcss/graphql/prettier inputs are absent and no green result for those packages is claimed. No check is skipped, relaxed or deleted. Obsolete wave-10 scratch artifacts with committed evidence were removed to retain 3 GB of free workspace space before the gates completed. Global inventory has 197 ranked rules: 25 base/main ports and 172 claimed; nothing available, so stop after landing.

## Expanded shared gate results

The supplied-input commands are unchanged filters:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestMutants|TestRegistrationMutant)$' -count=1 -timeout=30m -v > /tmp/wave10-legacy-lint.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript go test ./stage1/typescript/parser ./stage1/typescript/scanner -run '^(TestCompilerExpressionsAgree|TestWholeCompilerAgrees|TestScannerAgreesWithTypescriptGo)$' -count=1 -timeout=30m -v > /tmp/wave10-legacy-typescript.log 2>&1
```

All selected checks pass with no skips. Shared lint PASS 924.750s: witnesses 13053452 identical bytes, expanded 441-file corpus 21139204 identical bytes. Every result compares Go, source Node, emitted JavaScript and native; pre-existing explicit unsupported-recovery cases retain their panic assertions. All 40 registry mutants are caught on all three ports (704.56s), owned witnesses pass (19.73s), and registration mutant passes (17.66s). All names and exact mismatches are preserved in the log and shared-mutants.json. No shared check was changed.

Parser PASS 46.330s: 77 compiler files, expression trees 28836875 bytes, whole trees 44766682. Scanner PASS 33.488s: 77 compiler files, 327 stage1 files and 18236 generated inputs, 30795294 identical bytes. Punctuation != changed to ==, invalid decimal separator accepted and omitted regex rescan mutants are each caught by Node and native comparisons. No full-17-check or foreign-package result is implied.

Stream compiler medians: process-compiler native 1.766146s / Go 0.355250s = 4.972x; blocking-compiler native 2.170196s / Go 0.365773s = 5.933x; whole-process observations under concurrent load. Evidence is in evidence/legacy-landing.
