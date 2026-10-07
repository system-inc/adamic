Landing: wave 10 rebased onto area/stage1-lint d3a37422c, containing main b6b1538b0.
Source SHA: 8ccaa47efd17f0e25095b3722de668a3566f0de1; the evidence commit contains this report.
Checks: all nine native oracles re-green, supplied-input parity and uncached typeof/RegExp checks pass; setup 157s, nproc 5.
Mutants: all nine clean-exit rule mutants caught; raw-handle, reporting, decoded-options, scanner and incoming typeof mutants pass.
Limits: no unclaimed ranked rule across 651 origin refs; shared checker wiring and full required-input gate remain unverified.

The incoming typeof-null lowering, slot-presence and native runtime changes are retained unchanged. No rule implementation, shared harness, registration generator or protected compiler source is edited by wave 10. bash cloud/setup.sh reports Go, clang, Node and submodules ready in 0s each; build-cache warm 157s; total 157s on five processors, cpu.max 400000 100000. go run ./cmd/lint-registry succeeds with 40 descriptors. cmd/adamic is rebuilt and copied into the stream harness. Existing normal/sanitized Go checker archives are reused because bridge sources are unchanged.

## Repeated owned gates

Commands from LEGACY_LANDING_REPORT.md are repeated with /tmp/wave10-typeof-*.log output and /workspace/wave-10-refresh-* artifacts. The TypeScript source and compiler77/repository287 manifests are supplied. Test output goes directly to logs. All nine rules compare complete findings, fixes and suggestions with production Go cohere, normally and sanitized.

- TestWave10AgreementAndMutants: controls 61 findings / 18551 bytes; compiler 9 / 8346; repository 1 / 18863. Loop report mutant differs at byte 554, redundant constituent report mutant 4602, includes report mutant 8498. All exit 0 with empty stderr and are caught by byte comparison. Released member-parameter query rejects at 70; retained-registry mutant is caught by the required panic check.
- TestTimeoutAgreementAndMutants: 37 controls, 15 findings / 10260 bytes; compiler 5318, repository 18485. Lost-handle mutant differs at byte 54. Released runtime-context query rejects at 70, with a retained-registry mutant caught by that requirement.
- TestLandingNativeRulesAndMutants: process 92 controls, 77 findings / 45901 bytes; blocking 92 controls, 30 / 24200. Each matches compiler 5318 and repository 18485 normally and sanitized. Process chain-suppression mutant differs at 114; blocking-state mutant 13904; both exit 0 with empty stderr.
- validate.py exits 0: symbol 2114 bytes, hook 5902, raw await facts 882. Comparison-only mutants differ at 66, 70 and 628; hook regex mutant 819. The raw-facts await candidate retains its explicit refusal; the complete live rule is tested separately.
- live_validate.py exits 0: symbol controls 2114 bytes, hook 5902, complete require-await 4866. Each matches both corpora (5318 / 18485). Require-await matches 87 extracted complete upstream programs, 59 findings / 40041 bytes. Rule mutants differ at 66, 70 and 2266; incorrect union flag 34397; self-inference 35762. Released handles reject at 70 normally and sanitized for all three rules.
- validate_reporting.py exits 0: 32 reporting candidates / 20087 bytes, normally and sanitized; clean-exit range mutant differs at byte 58.
- Bridge package tests and vet pass; independent cohere lint reports zero findings on 13 owned Adamic sources.

## Supplied-input and independent checks

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestRegistrationMutant)$' -count=1 -timeout=30m -v > /tmp/wave10-typeof-lint.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript go test ./stage1/typescript/parser ./stage1/typescript/scanner -run '^(TestCompilerExpressionsAgree|TestWholeCompilerAgrees|TestScannerAgreesWithTypescriptGo)$' -count=1 -timeout=30m -v > /tmp/wave10-typeof-typescript.log 2>&1
go test ./stage1/cohere/lint -run '^TestDecodedOptionsAndMutant$' -count=1 -timeout=10m -v > /tmp/wave10-typeof-options.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/internal/oracle/testdata/(typeof_[a-z_]+|regexp(_exec|_unicode)?)[.]a$' -count=1 -timeout=10m -v > /tmp/wave10-typeof-node.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTypeOf.*Mutant$' -count=1 -timeout=10m -v > /tmp/wave10-typeof-mutants.log 2>&1
```

All selected tests pass without skips. Shared lint witnesses match Go, source Node, emitted JavaScript and native over 13051687 bytes; expanded 441-file corpus 21139204 bytes. Existing explicit unsupported-recovery cases retain their panic assertions. Owned witnesses and registration mutant pass. The ignored decoded-option mutant is caught on Node, emitted JavaScript and native; the nil-options guard remains unchanged. The 40-rule shared mutation matrix is unchanged and was proved in the previous landing; it was not repeated on this compiler-only rebase.

Parser matches 77 compiler files: 28836875 expression-tree bytes and 44766682 whole-tree bytes. Scanner matches 77 compiler, 327 stage1 files and 18236 generated inputs / 30795294 bytes. Punctuation changed to ==, invalid decimal separator accepted and regex-rescan-omitted mutants are caught on Node and native.

Uncached typeof/RegExp fixtures and the one-byte oracle pass against Node, with zero cache hits. Incoming TestTypeOfNullMutant passes over five witnesses; TestTypeOfNullSlotPresenceMutant, TestTypeOfConstructorMutant and TestTypeOfStringLiteralMutant pass. These mutants compile, exit cleanly without sanitizer errors or leaks, and only stdout comparison catches the wrong classifications.

Live compiler medians, seconds: symbol native 1.660663 / Go 0.365852 = 4.539x; hook 1.891692 / 0.434483 = 4.354x; require-await 2.180925 / 0.537735 = 4.056x. Repository ratios are 1.519x, 1.559x and 1.561x. Three alternating rounds include loading and serialization under uncontrolled concurrent gate load; these observations are not isolated rule costs. Exact rounds and stream timings are retained in evidence/typeof-landing.

## Remaining scope

Three React claims remain parked and are not certified native implementations; the historical JSX gap has closed. The shared RuleContext checker/program/source-path attachment remains missing, so shared-registry execution of the nine type-aware listeners is not certified. Live listener emitted-JavaScript comparison, full repository gate and all 17 required-input checks are not run; external postcss/graphql/prettier package inputs are absent and no passing result is claimed for them. No check is skipped, relaxed or deleted. All 197 ranked rules are accounted for by 25 base/main ports and 172 claims across origin branches; no new claim is made.

## Actual package completion lines

- original: github.com/system-inc/adamic/stage1/cohere/typeaware	232.296s
- timeout: github.com/system-inc/adamic/stage1/cohere/typeaware/wave_10_next	77.470s
- process: github.com/system-inc/adamic/stage1/cohere/typeaware/wave_10_next	205.160s
- bridge: github.com/system-inc/adamic/bridge/tsgo/checker	0.205s
- node: github.com/system-inc/adamic/internal/oracle	57.163s
- mutants: github.com/system-inc/adamic/internal/oracle	4.344s
- options: github.com/system-inc/adamic/stage1/cohere/lint	42.745s
- lint: github.com/system-inc/adamic/stage1/cohere/lint	328.957s
- typescript: github.com/system-inc/adamic/stage1/typescript/parser	57.968s; github.com/system-inc/adamic/stage1/typescript/scanner	41.503s

process-compiler median: native 1.795780s / Go 0.325397s = 5.519x.

blocking-compiler median: native 2.261603s / Go 0.315584s = 7.166x.
