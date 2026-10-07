Nine completed wave-10 ports are rebased and native-green on origin/area/stage1-lint d65a8f931.
Rebased source tip 9683d8a05; previous pushed tip 3001f75f9; branch codex/typeaware-wave-10.
All completed-rule byte-oracles, sanitizers, mutants and release checks pass; see ../wave_10_leaf/README.md.
The old JSX parser reproduction now succeeds; its historical failure below is superseded.
Three React claims remain PARKED under the user instruction; no full React rule implementation or mutant is certified here.

Claimed rules, in the by-volume order (each combined volume zero):

- react/forbid-elements
- react/forbid-prop-types
- react/iframe-missing-sandbox

The claim was pushed before this owned oracle file was written. The global
selection and any-origin checks are recorded in ../wave_10_next/evidence/landing.
No implementation of these three was found; the only any-origin source matches
were configuration inventory references. The six earlier claims are complete,
not held by this new blocker. ../wave_10_next/LANDING_REPORT.md holds their native
findings/fixes/suggestions, sanitizer, mutant, timing and rebase evidence.

Exact boundary: stage1/typescript/parser/main.ts is the unchanged shared parser
CLI used by the native lint runners. Its parser has no JSX support, consistent
with stage1/typescript/parser/GAPS.md. This standard TSX input:

```tsx
const frame = <iframe />;
```

is accepted by the independent production Go loader (exit 0). The unmodified
production react/iframe-missing-sandbox rule produces attributeMissing at bytes
14..24, with no fixes or suggestions. The native parser compiles successfully,
but running it over the same file exits 70 with empty stdout and exactly:

```text
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 22 in /workspace/wave-10-react-parser-input.tsx
```

This blocks JSX listener coverage before rule dispatch. It is a shared parser
gap, not a .a module-loading failure; current main compiles the earlier .a rules.
The new React rules cannot be counted as complete by treating their JSX branches
as silence. The oracle file imports no bridge code and calls the unchanged
production registry rule; it is verification scaffolding, not a native port.

Ahra's instruction is explicit: "If anything else blocks you, say exactly what
it is and stop, rather than editing shared files." This blocker is outside the
shared-harness exception, so work stops here. No shared parser, test harness,
registration generator or protected compiler source was edited. Non-JSX arms
of the new rules, their complete native oracles, rule mutants, sanitizer checks
and native timings have not been implemented or verified. No further rules are
claimed. All new native modules from the completed work are .a; this reproduction
adds only a Go oracle, documentation and compressed observations.

Reproduction (source the setup environment first; output is logged):

```sh
/workspace/wave-10-landing-process/adamic build \
  stage1/typescript/parser/main.ts -o /workspace/wave-10-react-native-parser \
  > /tmp/wave-10-react-parser-build.log 2>&1
/workspace/wave-10-react-native-parser /workspace/wave-10-react-parser-input.tsx \
  --whole > /tmp/wave-10-react-native-parser.stdout \
  2> /tmp/wave-10-react-native-parser.stderr
```

Build testdata/oracle_iframe.go as an overlay replacing the virtual main file
cohere/adamic_wave10_react_iframe_oracle.go, then run its binary with the existing
controls tsconfig and a manifest containing the same input file. The exact fixture,
exit statuses, expected production finding and compressed raw observations are
under evidence. The helper's loading/serialization is the same independent
production oracle used by the completed wave.

## Previous-main check and numeric listener boundary

On 2026-10-07, fetching all origin heads again leaves origin/main at
`e8ba3d5d81de4d3773c723914fccd4c76248b965`. It is an ancestor of the pushed
wave-10 branch. The JSX reproduction still gives production Go exit 0 with
one finding and native exit 70 with the exact error above.

The new numeric-listener requirement has a separate shared API prerequisite.
`stage1/typescript/parser/nodes.ts` declares `ParseNode.kind: string`; its
constructor takes `kind: string`. Searching the entire shared parser directory
for `SyntaxKind`, `kindId`, `kindCode` and `kindNumber` finds no numeric kind API.
The shared rule adapter also accepts node indexes and fetches nodes inside
`ask` and `add`. A rule-local numbering table would not be the parser's numeric
SyntaxKind and would silently invent a contract with the incoming driver.
Therefore no fabricated numeric listener declaration is added. The existing six
ports retain their previous string-kind implementation and do not satisfy the
new speed contract. Their byte-oracle results are separate from that limitation.

Required shared changes before these React ports can proceed: JSX parse nodes,
a public numeric SyntaxKind mapping, and an adapter that accepts the supplied
node without refetching it. These are shared-parser/driver changes, outside this
unit's permitted directories. No shared files were changed and no further claim
was made. Current-check logs are stored under evidence/current-main.

## Previous-main oracle observations

The unchanged six implementations were rechecked after the fetch. Commands
source `/workspace/adamic-tools/env.sh` and run these exact Go tests with
`-count=1 -timeout=10m -v`, redirecting all output to the compressed logs above:

- `./stage1/cohere/typeaware -run '^TestWave10AgreementAndMutants$'`: PASS, 92.045s.
- `./stage1/cohere/typeaware/wave_10_next -run '^TestTimeoutAgreementAndMutants$'`: PASS, 64.276s.
- `./stage1/cohere/typeaware/wave_10_next -run '^TestLandingNativeRulesAndMutants$'`: PASS, 167.210s.

Environment: `ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest`,
`ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest`,
`ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript`; original artifacts
`ADAMIC_WAVE10_ARTIFACTS=/workspace/wave-10-landing-original`; timeout enabled
with `ADAMIC_WAVE10_NEXT_VALIDATE=1` and artifacts
`ADAMIC_WAVE10_NEXT_ARTIFACTS=/workspace/wave-10-landing-timeout`; landing
`ADAMIC_WAVE10_LANDING_ARTIFACTS=/workspace/wave-10-landing-process` and
`ADAMIC_WAVE10_LANDING_FIXTURES=/workspace/wave-10-process-validation`.
Landing reuses the already-built stage-0 compiler and bridge archives from the
same unchanged source commit; no fresh process bootstrap or full gate is claimed.

All normal and sanitized findings/fixes/suggestions agree over controls,
compiler77 and repository287. Original controls: 61 findings; timeout: 15;
process: 77; blocking: 30. All six native decision mutants exit 0 with empty
stderr and are caught only by Go byte comparison: loop byte 554, redundant
constituents 4602, includes 8498, timeout 54, process 114, blocking 13904.
Original and timeout released-handle checks exit 70; the retaining-registry
mutants exit 0 and fail the required panic check. No new React mutant is claimed.

Setup: Go 0s, clang 0s, Node 0s, submodules 0s, cache 28s, total 28s;
`nproc` is 5. Native timing against Go below uses three alternating runs,
complete-process medians including loading and serialization. Every timed output
is compared; worker load is uncontrolled. Ratios denote native slowdown.

| Rule and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| process-controls | 3.751363 | 0.081383 | 46.095x |
| process-compiler | 1.755326 | 0.302095 | 5.811x |
| process-repository | 0.272872 | 0.134415 | 2.030x |
| blocking-controls | 5.548422 | 0.077356 | 71.726x |
| blocking-compiler | 2.137497 | 0.299683 | 7.133x |
| blocking-repository | 0.377423 | 0.131031 | 2.880x |

## Landing on main f8013f0b

This section supersedes the earlier landing observations. Fetching origin moves
main to f8013f0baac41ddc340d76f83bddde38536a8f07. Rebase of the eleven owned
commits succeeds without conflicts, yielding source tip bbb91e53. Main is fetched
again after validation and remains f8013f0b. All artifacts below are freshly
built with the rebased compiler; no previous-main compiler or bridge archive is
reused. Logs and raw extra timing outputs are in evidence/landing-f801.

Setup prints Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
cache warm 100s and total 100s. nproc is 5, cpu.max is 400000 100000,
with 17.6 GB. Shells source /workspace/adamic-tools/env.sh.

Exact gates, each logged, with -count=1 -v:

- go test ./stage1/cohere/typeaware -run '^TestWave10AgreementAndMutants$' -timeout=10m: PASS 92.810s.
- go test ./stage1/cohere/typeaware/wave_10_next -run '^TestTimeoutAgreementAndMutants$' -timeout=10m: PASS 53.530s.
- go test ./stage1/cohere/typeaware/wave_10_next -run '^TestProcessNativeAgreement$' -timeout=3m: PASS 47.298s.
- go test ./stage1/cohere/typeaware/wave_10_next -run '^TestLandingNativeRulesAndMutants$' -timeout=10m: PASS 179.812s.
- Selected checker-question tests in ./bridge/tsgo/checker and ./stage1/cohere/typeaware/wave_10_next: PASS.
- go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware ./stage1/cohere/typeaware/wave_10_next: exit 0, empty output.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -timeout=5m: PASS 10.537s, 19 Node and 28 native cache misses.

Common corpus environment remains the exact manifest/config paths in the
previous section. Artifact paths change to /workspace/wave-10-f801-original,
/workspace/wave-10-f801-timeout and /workspace/wave-10-f801-process using the same
named environment variables. Timeout uses ADAMIC_WAVE10_NEXT_VALIDATE=1;
process uses ADAMIC_WAVE10_PROCESS_VALIDATE=1 and
ADAMIC_WAVE10_PROCESS_ARTIFACTS=/workspace/wave-10-f801-process. Landing uses
ADAMIC_WAVE10_LANDING_ARTIFACTS=/workspace/wave-10-f801-process and the existing
92 controls at ADAMIC_WAVE10_LANDING_FIXTURES=/workspace/wave-10-process-validation.
The selected checker test regex is
^Test(RuntimeContextFactsAndLiveDispatch|SyntaxMembershipAndDestructuringAncestry|CoverageCheckerQuestions|FactEncoding|ShapeAndNameFacts|ExactIndexMatchesCompilerNodes|MemberParameterSourceText|AdamicRootKeepsConfigDeclarations)$.

Findings, fixes and suggestions agree byte for byte, normal and sanitized, on
compiler77 and repository287 and all controls. Original controls have 61
findings; compiler 9, repository 1. Timeout controls have 15 findings; both
corpora zero. The process bootstrap covers 64 controls with 47 findings; final
process covers 92 controls with 77 findings and blocking covers 92 with 30.
Process and blocking have zero findings on both corpora. Native sanitizer
stderr is empty.

Clean-exit native rule mutants are caught only by byte comparison: loop 548,
redundant constituents 4569, includes 8447, timeout 51, process 114,
blocking 13904. The first four positions change with the artifact path embedded
in serialized file names. Original member-parameters and timeout runtime-context
released queries exit 70 with the required invalid/released-handle panic;
the retaining-registry mutants exit 0 and are caught by that check. The Node
one-byte mutant is also caught.

The JSX CLI is rebuilt using /workspace/wave-10-f801-original/adamic and still
exits 70 on the same TSX input, with exactly the error recorded above. A freshly
built production Go iframe oracle exits 0 with one finding at bytes 14..24.
The shared parser, kind API and Diagnostic files do not change in this landing.
No numeric kind mapping or rule.json schema is present in current main; no
fabricated declarations or silent JSX omissions are added. The six legacy ports
still use their existing string-kind implementation. React implementations,
non-JSX arms, their native mutants and numeric listener conversion are not
complete. No full repository gate, full upstream fixture matrix or shared
emitted-JavaScript lint comparison is claimed. No further rules are claimed.

Fresh timing medians, three alternating complete-process runs for every row,
including load/parse/checker/serialization; all stdout bytes are compared.
The original three are measured as their combined suite, not individually.
Native stderr must be empty; the Go oracle's existing timing-counter line is
accepted only when it exactly matches its numeric load_ns/rule_ns/run_ns format.
The first two extra-timing attempts rejected this expected instrumentation;
their failure logs are retained. They did not find a rule disagreement.
Worker load is uncontrolled and other validation jobs overlap some runs.
Ratios are native slowdowns, not speedups.

| Rule or suite and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| process-controls | 4.022138 | 0.081926 | 49.095x |
| process-compiler | 1.929741 | 0.351042 | 5.497x |
| process-repository | 0.291002 | 0.134879 | 2.158x |
| blocking-controls | 5.468090 | 0.080602 | 67.840x |
| blocking-compiler | 2.148160 | 0.304330 | 7.059x |
| blocking-repository | 0.376349 | 0.148935 | 2.527x |
| original-three-suite-controls | 0.025833 | 0.033506 | 0.771x |
| original-three-suite-compiler | 5.412461 | 0.680368 | 7.955x |
| original-three-suite-repository | 0.524448 | 0.169518 | 3.094x |
| timeout-controls | 0.181971 | 0.040229 | 4.523x |
| timeout-compiler | 1.704854 | 0.296843 | 5.743x |
| timeout-repository | 0.281035 | 0.135875 | 2.068x |


Latest parser observation on the lint integration base: /workspace/wave-10-area-parser accepts the exact earlier TSX witness with exit 0 and empty stderr, emitting JsxSelfClosingElement. Canonical named kinds replace the historical numeric requirement. These old blockers are closed; this landing pass does not claim a React port. Evidence is in ../wave_10_leaf/evidence/landing-area/jsx.stdout.gz and jsx.stderr.gz.
