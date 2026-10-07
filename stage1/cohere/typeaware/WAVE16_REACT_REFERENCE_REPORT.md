Built: an independent production Go reference runner and two controlled JSX fixtures for the already claimed React rules.
Base: current main e8ba3d5d; fifteen completed native ports retain their tested source commit 69db2dc8 and published landing evidence.
Commands/output: 128 successful Go executions; two-creator messages vary 59/5, one-creator output is identical in all 64 runs; vet/gofmt clean.
Control mutation: replacing the component tag with a host tag leaves Go exit zero with zero findings; the positive-finding assertion alone exits one and catches it.
Uncovered: native source-to-HIR/SSA integration, complete React source ports, their qualifying native mutants/sanitizers/handle checks and end-to-end timing; no new claims.

# React source-port reference limitation

Landing cap remains satisfied for the previously completed implementations:
all origin heads were fetched, main is still e8ba3d5d, and this branch descends
from it. Comparing implementation sources with the tested 69db2dc8 finds no
changes to the fifteen rule ports. Their five-suite PASS (421.255s), bridge
PASS (69.592s), fifteen rule mutants and seven foundation mutants remain
applicable. No unchanged native suite was rerun for this reference experiment.
Only this unit's branch is pushed; neither main nor an area branch is targeted.

## Updated prerequisite inventory

Wave 21's e9ec024c now contains prepared-HIR cores for the same three names:
set-state-in-effect, set-state-in-render and static-components. Its core code
is bc6750a6. Its HIR model, native post-dominance and three validators exist.
Their report explicitly does not implement native source lowering, SSA,
compilation-unit selection, memo-shadowing annotations or production harness
integration. Their prepared-input comparisons do not establish source-port
agreement. The same names also appear in wave 21's reservations; no duplicate
core code or new reservations were added by this unit.

Wave 04's 5df6d34b adds refs environment kernels and explicitly still lacks
source lowering. These contributions improve the earlier prerequisite picture:
a native graph model and validator kernels now exist. The remaining blocker is
the source adapter and its exact gates/transforms, rather than the absence of
every native graph primitive. Existing Nexus shared dispatch integration is
still separately pending as documented in WAVE16_LANDING_REPORT.md.

## Observed Go byte ambiguity

The independent runner selects the three unchanged production registry rules.
It imports no bridge implementation and performs no native lint analysis.
The fixture defines createA/createB, assigns their results to C on opposite
branches, and returns a JSX C tag from a component-like function. All 64 runs
use identical input/config bytes and the same file paths. Every run exits zero
and reports exactly one finding, with zero fixes and zero suggestions.

Two distinct canonical stdout streams are observed, both 634 bytes:

| Creation-site message | Runs |
| --- | ---: |
| at createA | 59 |
| at createB | 5 |

The outputs differ only at byte 203, the A/B in the creation-site message.
Finding ID, rule, primary range 206..207, count and every other byte are
identical. Native output was not involved in this experiment. All 64 stdout
streams are retained compressed, with stderr and source/stream hashes.

The control changes only the else-arm call from createB to createA. Its two
incoming branches still create dynamic values, but their creation text is the
same. All 64 control executions have one identical complete stdout stream.
The host-tag control mutation replaces C's JSX tag with div. Go still parses
and exits zero, with zero findings; reproduce.py exits one with
`two-creator round 0: missing positive finding`. The Go output and mutation
source are preserved. This proves that the reference-positive check can fail;
it is not presented as a qualifying native rule mutant.

Production static_components.go ranges over phi.Operands, a Go map, and keeps
the first dynamic creator. That source mechanism explains the observed
message variation. A separately executed deterministic native analysis cannot
guarantee matching that independently chosen message on this input. Neither
output was normalized, suppressed or silently accepted as a passing comparison.
This is an additional limitation for the required byte-exact source rule bar,
beyond the missing native source adapter. No production Go source was edited.

## Reproduction

The owned Go source is testdata/oracle_wave16_react.go. Its build overlay maps
an otherwise nonexistent file inside cohere to this runner, allowing access to
production internal packages without modifying the submodule. The JSX sources
are .a files with .tsx symlinks so Go selects JSX parsing.

```
source /workspace/adamic-tools/env.sh
go -C cohere build \
 -overlay /workspace/adamic/stage1/cohere/typeaware/validation-wave16-react-reference/build-overlay.json \
 -o /workspace/wave16-artifacts/react-reference/oracle \
 /workspace/adamic/cohere/adamic_wave16_react_reference.go
python3 stage1/cohere/typeaware/validation-wave16-react-reference/reproduce.py \
 /workspace/wave16-artifacts/react-reference/oracle \
 > stage1/cohere/typeaware/validation-wave16-react-reference/run.log 2>&1
```

With go -C, pass the overlay's absolute repository path; the retained actual
build used /workspace/wave16-artifacts/react-reference/overlay.json. The saved
build-overlay.json has the same contents. build.log is empty. Go reference
stderr carries load/rule/run timing and is preserved; only canonical stdout
is compared for the observation. Input setup initially used the submodule
working directory for repository-relative evidence paths; those path errors
were corrected before the recorded experiment. The final build and all 128
reference runs succeeded. Vet and gofmt logs are empty.

Setup timing and nproc remain those recorded in the landing report. No native
performance comparison, full gate or new React native sanitizer/handle result
is claimed. All three React source claims remain incomplete. This unit stops
under Ahra's instruction to report prerequisites outside its owned rule work
rather than edit shared files, without claiming another set.
