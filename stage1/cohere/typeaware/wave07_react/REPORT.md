Built: isolated upstream JSX dependency probe; the three claimed React rules remain unported.
Commits: claims and original blocker evidence are pushed through 99d90694; this update records the newly available dependency.
Commands and outputs: fresh fetch found 417 origin refs; dependency_probe.py passes three JSX controls and one ordinary control, all exit 0.
Mutants: no React rule mutants or comparison gates ran; this probe checks parser capability only.
Not covered: React implementations, byte agreement, fixes/suggestions, sanitizer/released-handle gates, or native/Go rule timing.

# JSX dependency recheck

The original parser-blocker report below is historical. A subsequent all-head
fetch found origin/codex/stage1-jsx-lint at
`a8a62d62ca49db7415e14c3887dd305022b17309`, containing a published JSX parser.
Its JSX_REPORT.md was read whole before testing. The implementation commit is
`e715ef4a2f898230af63c40195dea6586a557899`.

[dependency_probe.py](dependency_probe.py) extracts that revision's unchanged
stage1/typescript source files into an isolated scratch dependency directory,
then compiles the existing .a parser probe against those sources. It changes no
shared files on codex/typeaware-wave-07. All three previously failing JSX
controls now exit 0 and print `jsx 1`; the ordinary control exits 0 and prints
`jsx 0`. Exact dependency source hashes, fixture strings, commands, exits and
stdout/stderr are saved in [jsx-dependency.json.gz](evidence/jsx-dependency.json.gz).
This demonstrates that the upstream dependency can remove the parser failure;
it does not demonstrate any React rule behavior.

The branch's shared parser remains unchanged and still has the previously
measured JSX refusal. Integrating the published parser requires shared changes
to parser.ts, lookahead.ts, scanner.ts and the new jsx.ts module. Ahra's explicit
instruction is "Keep your changes inside your own rule directories" and "If
anything else blocks you, say exactly what it is and stop, rather than editing
shared files." This unit therefore stops at pending shared parser integration.
No claim is made that the published parser is unavailable or defective. Once it
is integrated, the React ports still need their native value-flow implementations
and all required verification; the parser probe is not a substitute for them.

Reproduction:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave07_react/dependency_probe.py > /workspace/wave-07-artifacts/react-jsx-dependency-probe.log 2>&1
```

The established toolchain is reused (original setup 81 seconds; nproc 5).
No additional rules were claimed and none of the prior nine rule gates were
rerun. The original claim remains held for these three incomplete React rules.

# Historical blocker report

Built: third-trio claims and reproducible JSX blocker probes; none of these three React rules is ported.
Commits: claim d6724bb17f529da289d3d4e21afb559de1dbf148 was pushed before probes; blocker evidence follows in this report commit.
Commands and outputs: fetched 389 origin refs; production Go reports three positive findings, while native parser exits 70 on all three JSX controls; ordinary TypeScript exits 0.
Mutants: no React rule mutants were run because there are no React implementations; prior nine ports retain their recorded passing mutants.
Not covered: React rule ports, complete byte agreement, fixes/suggestions, sanitizer and released-handle gates, or native/Go rule timing.

# Third continuation blocked by shared JSX parser support

Previous claimed ports and evidence are pushed through 87ebb321. Selection fetched
all origin heads and inspected 389 refs and 33 distinct Markdown claim blobs.
Before this new claim those blobs named 142 ranked rules. The first eligible
names in VOLUME_REPORT.md's combined 197-row ranking were:

1. react-hooks/set-state-in-effect.
2. react-hooks/set-state-in-render.
3. react-hooks/static-components.

All have zero compiler and repository counts. Existing base ports are excluded
using the production oracle registries, including abbreviated native filenames
and naming aliases; ports on origin/main and claims on all origin branches are
also excluded. [selection.py](selection.py) reconstructs the decision from the
fetched objects while excluding this worker's third claim, and its complete
ref/claim/ranking audit is in [evidence](evidence/selection.json.gz).
The claim was pushed before any implementation or probe. No subsequent rules
were claimed. These three reservations remain incomplete.

## Observed blocker

The native Parser used by all preceding rule ports has no JSX mode or JSX node
construction. Its unary() branch in stage1/typescript/parser/parser.ts around
line 1426 treats every LessThanToken as a TypeAssertionExpression: it parses a
type and immediately requires GreaterThanToken. A self-closing JSX tag reaches
SlashToken instead. Passing a .tsx path does not change this branch.

The independent production Go oracle reports exactly one finding per fixture:

| Control | Production Go | Native parser |
| --- | --- | --- |
| Widget creates Inner from make(), then returns Inner JSX | staticComponents, span 45..50 | exit 70, SlashToken at 50 |
| Widget calls Dispatch-typed setState during render, then returns div JSX | setStateInRender, span 53..61 | exit 70, SlashToken at 76 |
| Widget calls Dispatch-typed setState in useEffect, then returns div JSX | setStateInEffect, span 68..76 | exit 70, SlashToken at 94 |

Each native stderr starts `adamic: panic: parser slice expected GreaterThanToken,
got SlashToken`. Native stdout is empty. An ordinary TypeScript function passed
to the same binary exits 0 and prints `jsx 0`, so this is a JSX grammar failure
rather than an inability to compile or run the probe. The .a probe source compiles
and its source lint passes: 276 rules, one checked, zero findings.

The native probe is deliberately a parser capability check, not a rule port.
The Go oracle independently loads the three .tsx controls with jsx:preserve,
retains a declaration root providing Dispatch/useState/useEffect, and invokes
the pinned production Run methods unchanged. It imports no bridge implementation.
Input fixture strings, full canonical production findings, native stderr, hashes,
exact commands and exit codes are preserved in evidence.

## Why work stops here

Ahra's correction says: "Keep your changes inside your own rule directories" and
"If anything else blocks you, say exactly what it is and stop, rather than editing
shared files." This is a shared native-parser gap, separate from .a module
loading, profile compilation or suggestion serialization in the shared harness.
Changing an implementation's extension to .ts would not supply JSX grammar.

Under the existing 26-port architecture, completing these rules on React inputs
requires extending the shared native parser's syntax substrate. That file is
outside this unit's territory. No shared parser, dispatcher, harness, generator,
protected compiler file or submodule pin was edited. Work stops at the measured
parser blocker rather than treating empty findings on the zero-volume corpora as
successful ports. The rules also need native value-flow machinery: production
uses SSA phis, capture translation, manual-memoization preparation and control/
post-dominator analysis. Those requirements were observed in the source; their
native coverage was not measured because the parser already prevents the controls
from reaching rule analysis. The JSX probe is the concrete runtime blocker.

## Reproduction and limits

All three production rule source files were read whole before probes: 1040 lines
for set-state-in-effect, 536 for set-state-in-render and 449 for static-components.
The reproducible command is:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave07_react/probe.py > /workspace/wave-07-artifacts/third-probe.log 2>&1
```

All test/probe output goes to files. The probe asserts each positive Go rule name,
all three exit-70 JSX refusals and the exit-0 ordinary-TypeScript control. The
existing stage-0 binary and environment from the previous gates are reused;
there was no new setup or toolchain change. Original setup was 81 seconds and
nproc reports 5. Production oracle pins and the TypeScript corpus are unchanged.

No React implementation, mutant, byte-comparison success, sanitizer result,
released-handle result or performance claim is made. There is no native React
rule time to compare with Go yet. Prior wave-07 reports preserve the nine completed
ports and their measured agreement/timings; their gates were not rerun here.
