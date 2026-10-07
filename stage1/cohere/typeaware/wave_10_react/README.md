The first six wave-10 claims are native-validated and pushed after rebasing onto current main.
Landing commit: 95d3ced8; next-three claim commit: b2336a9b; this commit records the parser blocker.
Go accepts the minimal TSX source and reports one iframe finding at bytes 14..24; the unchanged native parser exits 70.
No React rule implementation or rule mutant is claimed; the preceding six native decision mutants and release guards pass.
The three new React claims remain reserved and unported; shared JSX parsing is the exact blocker, with no shared-source edits.

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

## Current-main check and numeric listener boundary

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

## Fresh oracle observations

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
