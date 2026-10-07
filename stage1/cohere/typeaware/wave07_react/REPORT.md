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
