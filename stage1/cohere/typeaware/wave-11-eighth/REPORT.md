Built native default-mode no-danger-with-children, no-multi-comp and no-namespace with numeric handed-node listeners.
Claim 667e6a47 was pushed before code; rebased implementation 47e4fbc8 is based on main b8fb957a.
Go/native suite PASS: 250 parse-clean controls, 193 findings, both frozen corpora, sanitizers, metadata and ownership; earlier JSX regression PASS.
Three rule mutants and two raw-fact mutants run normally and fail Go-byte comparison; released-registry mutant fails required panic 70.
Limits: ignoreStateless=true and emitted-JavaScript/shared-harness lint execution uncovered; four React analysis claims parked; full raw snapshot remains slow.

## Selection and landing

Fetched origin explicitly with +refs/heads/*:refs/remotes/origin/* because the
configured default fetch covers only main. Audited 569 origin refs, 554 distinct
trees and 34 reservation documents. JSON inventories of available rules are not
reservations. The combined-volume ranking, lexical ties, excluded prior ports
and every named reservation. These three were the first remaining candidates;
main and bridge contain count/inventory references, not their native ports.
Selection snapshot and reconstruction are in selection/. Claim 667e6a47 was
pushed before any rule or bridge implementation was written.

The existing eighteen native production-default ports were already oracle-green
on main c01907a7 and pushed through 242f4172 before claiming this batch. Main advanced to b8fb957a after the initial push. The branch was rebased
cleanly and every owned suite was run again; results are in validation-b8/. The
new shared model ab70f38d4 exists on origin/lint-rules/harness and is not on this
main base. No shared harness, generator, parser, allocator helper or protected
compiler files were edited. No main/area branch is pushed and no PR is opened.

## Implementation and comparison

Each rule has its own .a file and numeric rule.json kinds. The owned driver
indexes callbacks by numeric SyntaxKind and passes its decoded node directly.
There is one raw snapshot query per source file and no per-rule node refetch.
The SourceFile multi-component listener owns the same source-order collection
and return traversal as Go. Namespaced/bare wrappers, known-component exclusion,
return traversal's intentionally narrow statement set, factories, class heritage,
component anchors, Unicode capitalization and declaration ordering follow Go.
Dangerous-HTML detection preserves the direct-object versus variable-spread
asymmetry, first-child-only whitespace behavior, first declaration selection,
computed literal property keys and cycle termination. Namespace detection keeps
the narrower pragma/import callee test and reports paired tags on their opening
node rather than the whole element.

The separate react_syntax_details.go and react_syntax_details.a implement the
new raw question. The shared dispatcher edit is one registration line. Go
returns AST roles, parser trivia flags, declaration syntax and Unicode simple
uppercase text, never React judgments or diagnostic decisions. The simple
uppercase primitive preserves Go's pinned mapping rather than silently adopting
JS full-uppercase expansion for sharp-s or comparing half a supplementary rune.
Native code performs the capitalization test. None of these production rules
has a Go regexp; no hand-rolled regex matcher was introduced.

The independent Go oracle imports unchanged production rules through a legal
build overlay inside the pinned cohere module. Its loader and AST walk import
no bridge implementation. Controls consist of 18 focused inputs plus 232 literal
source inputs extracted from Go fixture tables. Expectations and non-default
options are not copied. All 250 sources parse cleanly; findings are dangerous-HTML
28, multi-component 134 and namespace 31, total 193. Every byte range, full rule
name, message id/text, fix and suggestion field matches. The production defaults
emit zero fixes and suggestions, which are serialized and compared explicitly.

Both frozen corpora match too: 287 repository roots and 77 compiler roots, zero
findings. These populations do not supply positive JSX evidence; controls do.
Normal and ASAN/UBSAN/LSAN runs match all controls and both populations with empty
native stderr. Go heaps are not instrumented by C sanitizers. Released handles
panic 70 with the exact invalid-or-released-handle message.

## Every new mutant

| Mutant | Change | Catch |
| --- | --- | --- |
| danger-property | Append ! to the wanted own-property name | Exit 0, empty stderr; Go output differs at byte 1095 |
| namespace-colon | Search call argument for / instead of : | Exit 0, empty stderr; Go output differs at byte 2624 |
| component-first | Begin reporting at component 2 instead of 1 | Exit 0, empty stderr; Go output differs at byte 4951 |
| raw whitespace | Clear the parser's JSX trivia flag | Exit 0, empty stderr; Go output differs at byte 1043 |
| raw uppercase | Return identity text instead of simple uppercase | Exit 0, empty stderr; Go output differs at byte 4951 |
| released registry | Retain released program entries | Probe exits 0; required panic-70 ownership check fails |

An initial danger-property mutant matched two source sites and was refused by
the test's unique-target assertion. Its target was narrowed; only the compiled,
exit-0 byte mismatch in the table counts as the semantic mutant. An initial
manual build used the wrong CLI flag order and printed usage; the correct build
and the independent test builds all passed. Neither setup error is a mutant kill.

## Earlier JSX-text crash

The earlier raw numeric-syntax-bindings question mistakenly called Node.Text on
JsxText, which Go does not support. Running the previously pushed native binary
on <React.Fragment>text</React.Fragment> exited 70 with:

```
adamic: panic: Unhandled case in Node.Text: *ast.JsxText
```

Removed JsxText from that owned question's text accessor list. Its older rules
consume the numeric kind and tree, never a JSX-text payload, so empty payload is
the correct raw non-answer. A focused text/space regression was added to the
older worker test. Full earlier three-rule comparison, sanitizer, listener,
rule/raw-fact mutants and released-handle suite now passes: 230 parse-clean
controls, 307 findings, both corpora, 97.767s. Existing earlier mutant witnesses
are retained in eighth-seventh-regression.log. This actual pre-fix crash is
regression evidence, not counted as one of the comparison-only semantic mutants.

## Commands and observations

All commands source /workspace/adamic-tools/env.sh. Test output is redirected
to log files, never piped. Setup passed: Go/clang/Node/submodules ready 0s,
cache warm 94s, total 94s. nproc 5; cgroup quota 4 CPUs, memory 17.6 GB.

```
ADAMIC_WAVE_11_EIGHTH_ARTIFACTS=/tmp/wave-11-eighth-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test ./stage1/cohere/typeaware -run '^TestWave11EighthAgreementAndMutants$' -count=1 -timeout 15m -v
ADAMIC_WAVE_11_SEVENTH_ARTIFACTS=/tmp/wave-11-eighth-seventh-regression ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test ./stage1/cohere/typeaware -run '^TestWave11SeventhAgreementAndMutants$' -count=1 -timeout 15m -v
go test ./bridge/tsgo/... -count=1 -timeout 10m
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(sorting|string_index|functions|closures|call_targets_(closure|element|region|reuse|sort)|library_map_set_iterator_.*|047cb0d_n_.*|codepoint.*)[.]a$' -count=1 -timeout 10m -v
git diff --check
```

New suite PASS 111.157s. Earlier JSX regression suite PASS 97.767s.
Final bridge PASS 64.027s, checker 0.118s; package vet is clean. Filtered compiler
oracle PASS 3.931s, actually 53 fixtures plus its one-byte mutant; original Node,
native and emitted JavaScript are compared. This is compiler coverage and does
not claim emitted-JavaScript lint-rule execution. The full repository gate was
not run. Only identified, reproducible worker scratch archives were removed
when workspace capacity fell below 500 MB; logs and source evidence were kept.

## Quiet native versus Go timing on the original c01907a7 base

Three alternating rounds after all builds/tests completed, complete output to
files, independently checked identical in every timed run. Whole-process median:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Repository, 287 roots | 0.9227s | 0.1380s | 6.69x |
| Compiler, 77 roots | 7.0781s | 0.3421s | 20.69x |

Native query counts are 287/77, one per source file. All raw timing lines and
rounds are in validation/benchmark/results.json and adjacent streams. These
measurements do not claim a speedup: whole AST/binding serialization and native
frame/object decoding remain expensive despite numeric callback dispatch. Timed
intervals printed inside the concurrent verification suite are not used as the
quiet performance comparison.

## Limits

Production defaults, measured controls and the two pinned populations are
covered. No ignoreStateless=true option surface is exposed by this standalone
entrypoint. Fixture sources built from nonliteral Go expressions and direct
[]string tables are not extracted; the literal tables and focused controls are
the positive population. No full option matrix, suppression/edit engine or
shared-harness/emitted-JavaScript lint-rule execution is claimed. Shared harness
work and allocator-helper replacement are left to their owners. Four older
HIR/SSA/capture/return-escape analysis claims remain parked under #dnv6f2c and
are not native ports. Prior fifteen legacy rules still have their documented
string-kind execution limitation pending the shared handed-node API. This batch
adds three native ports, bringing wave 11 to 21 production-default ports; it
makes no further reservations.

## Landing validation on b8fb957a

Rebased implementation: 47e4fbc80e823d4c366296da4d7cc052b22057d1.
Original claim 667e6a47 was pushed before implementation; its rebased commit is
cfbf0291c7258c376e198c0a32e3dfe454213507. All eight owned comparison and
metadata suites passed, plus full bridge tests, package vet and filtered
Node/native/emitted-JavaScript compiler agreement including inherited static
field reads and the one-byte oracle mutant. Exact commands, completion codes
and timings are in validation-b8/b8-results.json; complete logs are adjacent.
No protected compiler or shared harness files were edited.

All 21 production-default rules were revalidated over their controls and both
frozen corpora, including sanitizers, released handles and 46 worker mutant
observations. Eighth batch passed in 135.017s; its 250 controls retain 193
findings. Seventh batch includes the repaired JSX-text regression and retains
230 controls and 307 findings. New semantic mutants again exit normally with
empty stderr and fail only full Go-byte comparison: danger-property byte 1071,
namespace-colon 2576, component-first 4879, raw whitespace 1027, raw uppercase
4879. The released-registry mutant exits 0 and fails required panic 70. Byte
indices changed because scratch artifact paths are shorter.

The original quiet timings above remain the standalone performance measurement.
Final-base concurrent verification measured repository native 1.9163s versus
Go 0.1764s and compiler native 8.4314s versus Go 0.3673s; these concurrent
observations are not substituted for the quiet benchmark. Full repository gate
and shared emitted-JavaScript lint execution remain uncovered.
