Rebased all wave-25 work onto main e8ba3d5d; no new rules claimed, and every carried patch is preserved.
Original pushed tip 16a717e8 is preserved through the rebases; final tested head a0da5a43 plus this accompanying evidence commit is pushed on codex/typeaware-wave-25.
Second wave gate PASS 394.989s; dependency PASS 534.513s; 876 control files plus both frozen corpora match Go findings, fixes and suggestions under sanitizers.
All 24 rule mutants are caught only by independent Go bytes; bridge fault injections, released-handle mutants and the uncached Node one-byte mutant pass.
Incomplete: boolean regex options, cross-file props and typed wrappers; full repository gate, options matrix, shared profile/JS rule comparison and CLI fixes. No additional claims.

## First landing attempt

The only branch pushed by this worker is codex/typeaware-wave-25. It was not on
main. Rebase onto e011f8f60899586d6373a5ccb07335ad82cfbf3c completed without
conflicts. Range-diff marks every one of the 13 carried commits equal.
The original pushed tip is 16a717e87985d6ed37411707d4c0a550e7198350; its
rebased counterpart is f0cfe2811ea3ce6043c5e117b822bb247c7d1520.
No shared harness, registration generator or protected compiler file was edited
by this landing work. Two own tests now explain why they run serially.

All five wave suites ran with the frozen 287-file repository manifest and
77-file TypeScript v6.0.3 manifest, normal and under ASan/UBSan/LSan. The 840
valid controls produce 527 findings and 117 ordered fixes, with no suggestions.
The fourth batch excludes the same three parse-invalid controls. The fifth
batch also compares nil boolean options, producing 84 findings instead of 141.
The carried ten-rule bridge dependency ran 36 controls with 53 findings,
6 fixes and 10 suggestions. Its repository corpus produces 180 findings;
its compiler corpus produces 16,589 findings, 2,222 fixes and 152 suggestions.
Every complete stream matches independent production Go cohere, preserving
ordered fixes, suggestions and duplicate findings. No source expectations are
imported into the oracle.

The complete bridge suite passes in 69.845s, checker in 0.242s, and Go vet
passes. The uncached filtered Node oracle passes in 7.913s on eight fixtures
plus TestTheOracleCatchesOneByte; cache statistics show zero hits. Released
handles refuse with panic 70. The bridge catches off-by-one input/output
lengths with ASan, removed frees and region heap allocations with LSan,
stale handles with assertions, wrong source positions with oracle bytes,
and removed link opt-in with its refusal check. Every individual wave and
dependency rule mutant and its first differing byte is in the retained logs.

Three isolated interleaved full-output samples of the fifth batch match bytes:
compiler native median 3.945301s vs Go 0.389238s (10.136x); repository native
0.561033s vs Go 0.135577s (4.138x). These measure load, queries, traversal and
serialization. Boolean naming declines with nil options on these corpora,
just as Go does; they do not measure its enabled mode.

Setup was rerun: go ready 0s; clang ready 0s; node ready 0s; submodules ready
0s; build cache warm 101s; done 101s on 5 processors, cpu.max 400000 100000,
17.6 GB. nproc is 5. Source /workspace/adamic-tools/env.sh for every build shell.
Go 1.27.1, clang 20.1.8, Node v24.19.0. All test output went directly to logs.
The exact gate commands, complete output hashes and compressed streams are in
validation-wave-25-landing, with source inputs and benchmark samples.

Current main now supports a constant new RegExp probe, which prints true.
A runtime-pattern probe still exits 1 at 3:31 with
`stage 0 can't lower RegExp with a nonconstant pattern yet`. This supersedes
the old fifth-batch observation that even the constant constructor was refused.
Boolean naming remains partial: arbitrary runtime regex options need compiler
support, while cross-file props annotations and typed wrapper component paths
remain unimplemented. Dedicated refusal tests exit 70 before findings for
those three paths. No new rules are claimed.

A final fetch found main had advanced to e8ba3d5d81de4d3773c723914fccd4c76248b965,
adding call-target routing, devirtualization and memory-analysis changes.
The completed first gate does not establish agreement with that compiler.
It is preserved here while the branch is rebased and revalidated again.
Not covered: full repository gate, nondefault option matrix, full shared-profile
integration, emitted-JavaScript comparison of these rule ports and CLI fixes.


## Second landing run

Main e8ba3d5d81de4d3773c723914fccd4c76248b965 carries the call-target and
devirtualization compiler changes. The second rebase completed without
conflicts. Its range-diff preserves all 14 carried commits exactly, including
the first evidence commit. Original port tip 16a717e8 becomes 5b8bec33;
the tested second head is a0da5a43357af8dc3246d5a836c9dddac5e4d08b. The final
accompanying commit changes reports, evidence packaging and claim status only.
No native rule or compiler-question implementation changed during landing.

Every suite was rebuilt against this compiler. Complete independent production
Go streams match, including fixes, suggestions and duplicates.

| Batch | Valid controls | Findings | Fixes | Suggestions | Identical control bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| original | 66 | 24 | 0 | 0 | 8192 |
| next | 60 | 57 | 0 | 0 | 29320 |
| third | 108 | 96 | 0 | 0 | 64321 |
| fourth | 364 | 209 | 117 | 0 | 85488 |
| fifth | 242 | 141 | 0 | 0 | 61122 |
| dependency | 36 | 53 | 6 | 10 | 24346 |

All wave batches compare the same frozen repository 287 roots and compiler 77
roots, normally and sanitized: zero findings, 18,485 repository bytes and
5,010 compiler bytes in each suite. The dependency compares 180 repository
findings in 85,151 bytes and 16,589 compiler findings in 7,120,613 bytes, with
2,222 fixes and 152 suggestions in the compiler. Every normal and sanitized
stream matches. The fifth nil-options control also matches 84 findings.

The second bridge suite passes in 67.071s, checker in 0.246s; the uncached
filtered Node oracle passes in 1.410s on eight fixtures and the one-byte
mutant, with zero cache hits. Go vet and diff checks pass. Released handles
refuse exactly with panic 70; retaining them in the registry makes the
mutants exit 0 and fails the required-panic expectation.

## Every rule mutant in the second run

Each builds, exits 0, emits empty stderr, and differs from its independent
Go truth stream. No compiler refusal or clang warning kills these mutants.

| Batch | Mutation | First differing byte | Catcher |
| --- | --- | ---: | --- |
| wave-25 | eval-library | 70 | Go findings/fix bytes |
| wave-25 | memo-inline | 19756 | Go findings/fix bytes |
| wave-25 | boolean-capital | 44875 | Go findings/fix bytes |
| wave-25 | throw-conditional | 1330 | Go findings/fix bytes |
| wave-25 | regex-forward | 10035 | Go findings/fix bytes |
| wave-25 | arrow-fix | 58859 | Go findings/fix bytes |
| wave-25 | collection-zero | 1474 | Go findings/fix bytes |
| wave-25 | outcome-incomplete | 17407 | Go findings/fix bytes |
| wave-25 | pure-callback | 26439 | Go findings/fix bytes |
| wave-25 | graphql-mask | 992 | Go findings/fix bytes |
| wave-25 | inject-strip | 5727 | Go findings/fix bytes |
| wave-25 | exit-state | 117 | Go findings/fix bytes |
| wave-25 | race-drop | 19461 | Go findings/fix bytes |
| wave-25 | blocking-count | 26456 | Go findings/fix bytes |
| dependency | before | 65 | Go findings/fix/suggestion bytes |
| dependency | cast | 1592 | Go findings/fix/suggestion bytes |
| dependency | methods | 2266 | Go findings/fix/suggestion bytes |
| dependency | coercion | 1189 | Go findings/fix/suggestion bytes |
| dependency | caller | 10374 | Go findings/fix/suggestion bytes |
| dependency | parameter | 377 | Go findings/fix/suggestion bytes |
| dependency | invariant | 12930 | Go findings/fix/suggestion bytes |
| dependency | optional | 15729 | Go findings/fix/suggestion bytes |
| dependency | alias | 16223 | Go findings/fix/suggestion bytes |
| dependency | unused | 17067 | Go findings/fix/suggestion bytes |

Bridge foundation mutants independently prove input/output bounds via ASan,
missing buffer frees and heap region allocation via LSan, stale handles via
assertions, wrong source position via oracle bytes, and removed link opt-in
via refusal. TestTheOracleCatchesOneByte proves external Node disagreement
is detectable. Full logs retain each observation.

## Commands and final measurements

```sh
bash /workspace/wave-25-validation/landing-second-gate.sh
bash /workspace/wave-25-validation/landing-second-dependency-gate.sh
python3 /workspace/wave-25-validation/landing-second-benchmark.py
```

These scripts record the exact Go test filters and environments; every test
run redirects output directly to its log. Setup and toolchain are the same
as the first run above. No build or test competes with the timing samples.
Three interleaved Go/native full-output samples match byte for byte. These
measure the fifth suite with nil boolean options, not enabled boolean naming.

| Population | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 3.963835s | 0.362370s | 10.939 |
| repository | 0.557283s | 0.131633s | 4.234 |

Native is slower in both measured populations. Samples and counters are
observations; no causal performance improvement is claimed.

The second full evidence package is validation-wave-25-landing-second.
The first remains validation-wave-25-landing, with metadata pinned to its
actual tested base and head rather than the main tip observed afterward.

Boolean naming remains partial. Runtime regex patterns require compiler
support; cross-file props and typed wrapper detection still need port work.
Its three explicit refusal controls continue to exit 70 before findings.
This landing unit does not complete those paths and does not claim more
rules. Its claim remains reserved. Nondefault options, shared-profile
integration, emitted-JavaScript comparison of the rule ports, CLI fixes and
the complete repository test gate remain outside this validation.

## Numeric listener requirement, October 7

An explicit fetch of all origin heads confirmed that current origin/main is
still e8ba3d5d81de4d3773c723914fccd4c76248b965. The pushed wave-25 tip
b2c7c75fe93a0526bb8ebcd0e150fa6b0a99d91a already contains that base;
the oracle evidence above applies without a source change.

The new speed requirement is not satisfied by these existing ports. They use
string syntax kinds and scan their own projection. On this base,
stage1/typescript/parser/nodes.ts declares ParseNode.kind as string and its
constructor takes a string. The fetched origin/codex/lint-harness-dot-a version
has the same declarations. A search of stage1/typescript,
stage1/cohere/typeaware and bridge/tsgo found no SyntaxKind, syntaxKinds,
kindIndex or kindId contract. Scanner tokens are also string maps.

Consequently there is no parser numeric SyntaxKind contract against which to
write the requested listener declarations or validate dispatch. Inventing a
private numeric enumeration would not establish agreement with the coming
shared driver. Shared parser and harness changes are outside this unit's
territory. Work stops at that dependency, as instructed, with no additional
claims. This report-only update does not change the previously tested source.
Boolean-prop-naming remains partial for arbitrary runtime regular expressions,
cross-file props annotations and typed memo/forwardRef wrappers; the latter two
remain unfinished local implementation, not shared-harness blockers.

## Rule metadata continuation, October 7

The subsequent user instruction specifies rule.json kinds rather than numeric
parser values. The fetched lint-harness-dot-a branch supplies concrete examples
whose kinds arrays contain syntax-kind names. This resolves the declaration
format ambiguity recorded above without requiring a numeric parser enum.

Added declarations under typeaware/rules for the current three React claims:
unsupported-syntax listens to Identifier, WithStatement and ClassDeclaration;
use-memo listens to CallExpression; boolean-prop-naming listens to SourceFile.
These match the corresponding upstream Go listener maps. SourceFile is the
upstream boolean-prop-naming entry point, which analyzes component descendants.
The declarations contain only name and kinds; they do not advertise a visit
method or factory that these classes do not yet implement.

The current runner does not consume these declarations. Existing run methods
still scan the projection and compare string kinds, so speed compliance remains
unfinished local work. No claim of handed-node or kind-indexed dispatch support
is made. No shared files were changed and no new rules were claimed. Source
implementations are unchanged; prior oracle results apply, and no native tests
were rerun for this metadata-only change. JSON shape and upstream listener names
were checked independently before committing.

## Third landing, October 7, current main f8013f0b

Rebased wave-25 cleanly onto f8013f0baac41ddc340d76f83bddde38536a8f07.
Tested source tip: 343b9f2a2dbb0a927c738eed953bfd6a7f6713fb.
No protected compiler file differs from main. No new rule was claimed.
The final all-heads fetch confirmed the tested main base remained current.
Compressed logs, exact scripts, source hashes and timings are preserved in
validation-wave-25-landing-third. nproc is still 5; the existing toolchain
was reused, with no setup rerun.

All five owned rule suites passed against Go findings, fixes and suggestions
on controls, repository sources and TypeScript compiler sources, including
ASan/UBSan and released handles. The aggregate first command exited 1 because
the separate refusal test reused an existing scratch symlink, not because of
a source disagreement. Its isolated fresh-directory retry passed in 23.165s.
No source or shared harness change was needed. The initial aggregate log,
including that failure and the five passing suites, is retained intact.

The inherited ten-rule dependency passed in 515.365s: controls 53 findings,
repository 180 findings and compiler 16,589 findings with 7,120,613 identical
serialized bytes under both normal and sanitizer runs. Bridge packages passed
in 66.992s and 0.242s. The filtered uncached Node oracle passed in 10.690s,
with zero cache hits, 25 native misses and 17 Node misses. go vet ./... exited
0 with empty output. These are scoped checks, not the complete repository gate.

All 24 rule mutants compiled, exited 0 and were caught only by independent Go
byte comparisons. Owned mutants: graphql-mask, inject-strip, collection-zero,
outcome-incomplete, pure-callback, exit-state, race-drop, blocking-count,
throw-conditional, regex-forward, arrow-fix, eval-library, memo-inline and
boolean-capital. Dependency mutants: before, cast, methods, coercion, caller,
parameter, invariant, optional, alias and unused. Their exact differing byte
offsets are recorded in the preserved logs. Released-registry mutants were
caught by the required invalid-handle panic, separately from rule mutants.
Bridge buffer bounds and leak mutants and the one-byte oracle mutant passed
their intended negative checks.

Three interleaved whole-process Go/native measurements used the rebuilt fifth
suite with nil boolean options. Each output pair matched byte for byte.
Compiler medians: native 4.078352513s, Go 0.388058155s, ratio 10.510.
Repository medians: native 0.584706747s, Go 0.142006629s, ratio 4.117.
These are total adapter times, not equivalent per-rule listener times.

This landing preserves the existing limitations: boolean-prop-naming is
partial for arbitrary runtime regex, imported props and typed wrappers;
handed-node dispatch and string-kind removal remain unfinished. Its three
React kinds declarations are present but not consumed by the current runner.
No full gate or new emitted-JavaScript rule comparison was run.

## RegExp literal correction, October 7

The new regex instruction supersedes the previous hand-written boolean prop
matcher. Replaced it with one module-level JS RegExp literal,
/^(is|has)[A-Z]([A-Za-z0-9]?)+/u, matching the upstream default expression.
No codex/lint-regex remote branch exists in the all-heads fetch, so no shared
translation row was available. The literal is initialized once per module.
The formerly specialized alternate configured pattern now refuses explicitly;
no option pattern is implemented by a hand-written matcher. Arbitrary options
remain incomplete until native dynamic RegExp construction can lower.

The affected fifth agreement/mutant suite and refusal suite passed together
in 107.628s on main f8013f0b. Controls with defaults: 141 findings and 61,122
identical serialized bytes; nil options: 84 findings and 53,794 bytes.
Repository and compiler findings, fixes and suggestions remain byte-identical
under normal and sanitizer builds. Released syntax handles still panic 70.
The boolean-capital mutant now changes the RegExp literal's [A-Z] to [a-z];
it compiles, exits 0 with empty stderr, and only the Go byte comparison catches
byte 44,875. eval-library and memo-inline also passed their negative checks at
bytes 70 and 19,756. Unaffected suites retain the preceding landing evidence.

Observed affected-suite whole-process timings: repository native
0.534577645s versus Go 0.131878477s; compiler native 3.800501925s versus Go
0.387902896s. These are single samples from the suite, not medians.

A separate well-typed .a probe of new RegExp(pattern, 'u'), where pattern comes
from programArguments, exits 1 at line 3, column 28 with:
stage 0 can't lower RegExp with a nonconstant pattern yet.
The initial probe had a console output type error; its corrected source and
the actual lowering failure are preserved in validation-wave-25-regex.
No compiler or shared harness file was edited to bypass this dependency.

The remaining React blockers do not qualify for the user's IR, single-
assignment or capture-analysis parking exception. Imported props and typed
wrappers remain unfinished local implementation. Handed-node dispatch and
string-kind removal are also unfinished. The claim remains partial and
reserved, with no new claims. Final fetch confirms main remains f8013f0b.
