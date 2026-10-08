# Explicit this parameter ruling needed

Preserved the existing rejection; no lowering rule was added.
Base b410340dc8f889b5799c3bc519117c63def3aa24, replay merge 1e8413ac63bc1509284a01e406d2c8bda03ec956.
Both example replays ran; checker reproduces this, utilities stops earlier on __String.
No mutants ran because no compiler rule or check changed.
Zero of the table's 46 sites are covered; no backend execution is claimed.

Assumption: the unit's explicit instruction to preserve design refusals applies
conservatively to docs/0.1.md:40, which admits this only inside methods and
constructors. docs/library_function_expressions_for_in.md:12 also explicitly
preserves rejection of dynamic receivers and explicit this parameters.
docs/parameter-properties.md:48 still records explicit dynamic this as NotYet.
No later receiver ruling was found in the language or memory documents.
This is an interpretation of the approved subset, not evidence that a typed
receiver cannot be implemented soundly.

The ruling needed is an explicit extension admitting ordinary functions with a
proven this parameter and receiver binding through call/apply, plus construction
through function values. It must settle the initialization proof for function
constructors, preserving the rejection of untyped this, unchecked receiver
casts, and escape before required fields are initialized. Merely allocating the
final typed shape with defaults would change Node's absent-property behavior.
Until that extension is ruled in, this unit keeps the existing NotYet unchanged.

## Pins and accounting

The specific branch instruction takes precedence over the generic main-start
instruction: work began from origin/area/compiler at the resolved base above.
Merged origin/codex/stage3-census-replay at
9a1f14c5d994aa855625e7cfa295677060348fec and pushed immediately.
The table ref resolves to 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7.
Reading its stage3/notyet-table/roots/raw.csv with csv.DictReader and filtering
reason == "this outside a method" gives 46 rows and 46 unique
(kind, where, reason, text) signatures: utilities.ts 45, checker.ts 1.
Coverage from this unit is 0/46, not an estimated unlock count.

## Commands and observations

All execution logs went to files. The completed adapted project was made with
bash stage3/apply.sh /tmp/notyet-this-adapted, exit 0.
The replay commands sourced /workspace/adamic-tools/env.sh and used:

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-this-adapted/src/tsc/tsc.ts -where /tmp/notyet-this-adapted/src/compiler/checker.ts:1456:5 -kind NotYet -reason 'this outside a method' > /tmp/notyet-this-checker-final.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/notyet-this-adapted/src/tsc/tsc.ts -where /tmp/notyet-this-adapted/src/compiler/utilities.ts:8473:5 -kind NotYet -reason 'this outside a method' > /tmp/notyet-this-utilities-replay.log 2>&1
```

Checker exits 0: reproduced NotYet at checker.ts:1456:5, selected NodeLinks
at 1455:1. Load/register/lower 2.466979773s; total 2.825427158s.
Utilities exits 1: signature did not reproduce. Selected Symbol at 8471:1;
its signature stops at 8471:51 with NotYet: a value of type __String.
Load/register/lower 2.926637169s; total 3.866181610s.
These are unchanged baseline outcomes, not next stops reached by new lowering.
An earlier checker replay overlapped adaptation and is superseded by the final run.

Setup used export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh.
The first invocation overlapped checkout and failed with undefined lowering
symbols (typedArrayWrite, censusFieldSlotless, parameterProperty,
classMembersWithParameters, parameterPropertyResets). Their definitions exist
on the completed merged branch. Retrying setup there exited 0.
Retry timings: Node 0.031s, Go 0.033s, markdown ready 0.088s, submodules
0.106s, clang 0.189s, Go build 44.999s, test binaries deferred 45.115s,
cache warm 45.117s, done 45.157s. nproc 5; cpu.max 400000 100000.
The requested /opt/adamic-tools/env.sh did not exist; the printed
/workspace/adamic-tools/env.sh was sourced for final replays.
Logs: /tmp/notyet-this-setup.log, /tmp/notyet-this-setup-retry.log,
/tmp/notyet-this-apply.log and the two replay logs named above.

No production compiler files, oracle fixtures, or counts changed. No tests,
mutants, whole packages or full gate ran. Fixture reduction and Node/backend/
sanitizer validation remain unbuilt pending the receiver ruling.
