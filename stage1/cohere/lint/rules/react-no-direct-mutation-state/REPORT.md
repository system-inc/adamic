# react/no-direct-mutation-state

The rule uses unchanged Go cohere NoDirectMutationState through a no-options adapter.
Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.
Branch base: bd5490b6ff36cfc6013d549f78ffa4f9f9a6e075 (origin/area/stage1-lint).
The remote ownership search found no existing descriptor before porting.

The port shares the certified ES6 component, strict ES5 factory, React projection,
component-base and parentheses helpers. Rule-local code walks the state member
chain and ancestor latches. Assignments report the unparenthesized target; updates
report the entire expression. No fixes, suggestions or options are introduced.

Upstream observations preserved deliberately: computed roots, parentheses inside
the chain, delete, destructuring, mutating calls and aliases are silent. Any call
ancestor inside the constructor defeats its exemption, including a synchronous
argument write. The first class ends the ancestor walk after component detection.
createClass and React.createClass are silent; createReactClass spellings report.
There is no JSX filename gate. These are Go behavior, not proposed improvements.

Selected test: 56 unique captured upstream source/file/options cases, guarded by
an exact count; five witnesses, each required to report in Go; selected and all-rule
witness rows; inherited generated rows. Total 154 rows, 585,071 output bytes matched
Go across source Node, emitted JavaScript and ASan/UBSan native.
Witnesses cover constructors and calls, dotted and computed roots, parentheses,
updates, non-write controls, factory spellings, class expressions, nested plain
classes, Unicode positions and JSX.

The compiling mutant drops the call-expression exception in constructors. The
constructor witness kills it by byte mismatch against Go on source Node and emitted
JavaScript. The harness checks its baseline native build separately; this mutant
is not claimed to have been executed natively.

Reproduce selected tests after sourcing the cloud setup environment:

```sh
bash stage1/cohere/lint/rules/react-no-direct-mutation-state/testdata/run-selected.sh > /tmp/direct-state-selected.log 2>&1
```

The full gate uses the clean TypeScript checkout at
050880ce59e30b356b686bd3144efe24f875ebc8, WASI_SYSROOT from cloud/setup.sh
--wasi-sdk, ADAMIC_LINT_BENCH=1, and the same fresh directory for
ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS.
ADAMIC_LINT_RULES is unset. The overlay adds only the owned selected test;
all existing package tests run without a filter. Test output is written to files.

No shared file, compiler, parser, submodule or registration list is edited.
No helper or language blocker was encountered. Certification is for captured
cases, witnesses and the inherited corpus, not arbitrary malformed input.

Full gate result: 160 pass events, 0 fail events,
1 skip events (named tests and subtests; package events excluded).
The skip is TestCheckerBridgeRefusalPending, awaiting TSGoError in the prelude.
All requested optional inputs were set; upstream, inherited corpus, witnesses,
profiling and mutants passed. Wall time: 1163.055s; nproc: 5.
Load before: 1.50 2.46 1.30 1/138 4116; load after: 5.06 5.64 4.16 1/143 160282.
The inherited corpus comparison matched 31,884,557 output bytes.
The full gate includes the final selected-test count guard and matches 56 upstream
cases again. Its temporary filenames produce 584,017 output bytes in this run.

Commands and outputs are preserved in evidence/setup.log, evidence/wasi-setup.log,
evidence/registry.log, evidence/selected.log, evidence/gate.jsonl and
evidence/gate-summary.json. The full command was:

```sh
go test -overlay=/tmp/s13-react-no-direct-mutation-state-overlay.json ./stage1/cohere/lint -count=1 -json -timeout=30m > /tmp/s13-react-no-direct-mutation-state-gate.jsonl 2>&1
```

The complete repository gate was not run. Only this rule directory is committed.
