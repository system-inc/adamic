Rebased nine completed rules and fifteen declaration-only modules onto current main c01907a7; six React claims remain parked or blocked.
Validated source SHA: c19a92a8e3185c7d525bef3a7e53bca2182a4242; following commit stores this evidence.
Checks: all three complete rule gates PASS; checker 0.881s, Node 5.699s, vet empty; all listener/manifests and Go JSX controls PASS.
Mutants: nine semantic, six raw-fact, five registry, fifteen compiling declaration and fifteen JSON mutations caught.
Not covered: three parked HIR/SSA ports, three blocked JSX ports, numeric handed-node conversion, full repository gate, own-rule emitted JavaScript comparison.

Main advanced from f8013f0b to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06.
Rebase completed without conflicts. No shared diff was reverted; the explicitly
anticipated leak-helper changes were not present in this main delta. This
worker changes no shared test harness or parser. Only this own branch is pushed
with an exact lease against 8a1e99f2358f7c718e8ae15dcd3a3e6b39423bf3.
A final remote check still found c01907a7 as main. Never main or area/.

First batch: 48 controls, 12,885 identical finding bytes normal/sanitized;
full gate PASS 219.202s. Nexus: 121 projects, 160 findings. Core: 682 projects,
502 findings, 365 suggestions. All complete message, position, automatic fix
and suggestion edit bytes agree with unchanged production Go. Both frozen
corpora, 287 repository roots and 77 compiler roots, match normal/sanitized:
18,485 and 5,241 bytes, zero findings. Positive controls make this meaningful.
The previously documented invalid upstream TSX exclusion remains unchanged.

Each semantic mutant compiles, exits normally with empty stderr, and is caught
only by Go finding bytes, including the suggestion-span mutation. The raw-fact
mutants also change bytes. Released checker questions panic 70; five registry
mutations exit 0 and are rejected by that required-panic expectation. Fifteen
listener mutations compile and exit 0 and are caught by independent Go
production-registration bytes. Fifteen JSON first-kind increments are caught
separately; metadata checks do not validate the absent React source analyses.
Normal and ASan/UBSan/LSan metadata probes agree with empty stderr.

Commands source /workspace/adamic-tools/env.sh. Substitute PREFIX=c019 in
the exact prior LANDING_F8013F0B_REPORT.md commands' artifact and log paths:
continue-first/second/third/listeners/manifests/checker/node/vet become
c019-first/second/third/listeners/manifests/checker/node/vet. The checker command
used the correct ./bridge/tsgo/checker package. The actual commands are:

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/c019-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave20-validation/c019-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c019-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /workspace/wave20-validation/c019-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c019-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /workspace/wave20-validation/c019-third.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c019-listeners --compiler /workspace/wave20-validation/c019-third/adamic --checker /workspace/wave20-validation/c019-third/checker.a > /workspace/wave20-validation/c019-listeners.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/manifest_verify.py --oracle /workspace/wave20-validation/c019-listeners/oracle --artifacts /workspace/wave20-validation/c019-manifests > /workspace/wave20-validation/c019-manifests.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c019-jsx-listeners --compiler /workspace/wave20-validation/c019-third/adamic --checker /workspace/wave20-validation/c019-third/checker.a > /workspace/wave20-validation/c019-jsx-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /workspace/wave20-validation/c019-checker.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(strings|closures|generic_functions)[.]a$' -count=1 -v -timeout 5m > /workspace/wave20-validation/c019-node.log 2>&1
go vet ./... > /workspace/wave20-validation/c019-vet.log 2>&1
(cd cohere && go test ./internal/lint/rules/react -run '^(TestJsxFragments|TestJsxNoConstructedContextValues|TestJsxNoUndef)' -count=1 -v > /workspace/wave20-validation/c019-go-jsx.log 2>&1)
```

Go JSX controls pass 0.491s. Fresh parser probe with the newly built Go/core
oracle and native runner confirms Go accepts valid JSX, exit 0, while native
panics 70 before dispatch: expected GreaterThanToken, got SlashToken at 94.
This probe tests parsing, not findings for the three new rules. Main and
area/stage1-lint still lack JSX and numeric ParseNode.kind. The HIR-dependent
three remain PARKED on #dnv6f2c; the JSX-dependent three remain BLOCKED rather
than falsely marked complete. No new claims are taken.

The shared regex table now exists at origin/codex/lint-regex 071fb0128.
Its README and REPORT were read. None of the nine completed rules or three
pending JSX rules imports Go regexp or has a regexp.Compile/MustCompile site,
and the table has no corresponding row. No new matcher or fallback was added.
Prefer-regex-literals parses pattern syntax through the existing corresponding
Go regexpattern walk; that is not a Go regexp compile site to translate.
No arbitrary option-pattern dialect equivalence is inferred from table presence.

Whole-process observed native / Go seconds: first repository .661158 / .250285,
compiler 6.140083 / .891366; Nexus repository .365714 / .217847, compiler
2.486448 / .465607; core repository .368133 / .215263, compiler 2.373663 /
.515852. Gates overlapped, so these are observations rather than isolated
performance comparisons. Native is slower; no dispatch improvement claimed.

Toolchain reused from the preceding setup (121s): Go 1.27.1, clang 20.1.8,
Node 24.19.0; nproc rechecked 5, quota four cores. No setup rerun or full
repository gate. Every process output was saved to files; archives omit binaries.
