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
