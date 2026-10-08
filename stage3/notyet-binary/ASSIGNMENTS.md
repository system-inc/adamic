# Assignment values

Built store-and-yield IR for plain `=` expressions through existing local, field and array-element stores. The receiver and index run before the right side and run once. The result is the assigned value, held independently of the target until the surrounding statement ends. Target storage conversions do not change the result: assigning 300.75 into Uint8Array yields 300.75 and stores 44; assigning into a Weak slot yields the strong value and stores only a weak handle. Scalar/maybe and boxed local storage preserve the checked result representation. Accessor dispatch invokes the setter once, without invoking a getter to obtain the result.

## Coverage

The saved raw CSV contains 190 unique `=` sites: 87 value/value and 103 number/number. [assignment-evidence/sites.csv](assignment-evidence/sites.csv) records them individually. This addresses their assignment-value operator family; it is not a measured net census delta or a claim that all enclosing TypeScript compiler bodies compile. One known site remains outside the supported target stores: `builder.ts:1424:16`, `root.length = root.length - 1`, needs an array-length store. Thus 189 are candidate sites for the represented store rules, with earlier checker, binding and storage stops still applicable.

Existing representation boundaries remain: union fields, boolean/undefined and mixed boxed arrays, destructuring assignment values, tuple element stores and super/private-static accessor assignment values. No object representation is invented for unknown values. Constant undefined results reuse the existing Undefined IR constant and NULL backend representation. An automatic approval review rejected a broader checker-type-based undefined interpretation; the approved implementation admits only an already-lowered constant-undefined store.

## Analyses and ownership

Freshness sees each store and the returned alias. Conditional and short-circuit expression states join both possible paths rather than allowing a nonexecuted store to erase an alias. Flow follows stores and result aliases; a local assigned within an operand is conservatively externally mutable, rather than claiming an unconditional statement-level SSA definition. Native reads snapshot those locals before a later operand can write them. Existing statement-based reuse, region, element borrowing, chain borrowing and exact receiver plans are conservatively disabled for programs with assignment values. This trades optimization for counted heap ownership; no garbage collector is introduced. A future expression-effect CFG can recover those optimizations.

## Validation

43 fixtures cover all currently represented local values; represented field and element values; globals, parameters and captured variables; runtime-built references; new-value retention across a subsequent store; undefined; all three typed-array conversions; weak slots; and accessor receiver/setter/getter counts. The focused oracle holds source Node to release native, ASan, UBSan, LeakSanitizer and backend JavaScript. The final assignment oracle passes in 4.987s. The combined assignment/logical/equality regression passes in 7.300s. Assignment cycle and existing storage-boundary tests pass in 0.117s. Counts refresh passes in 30.349s and adds only 43 fixture rows. Vet exits 0 with no output.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/assignment_value -count=1 -timeout 15m > /tmp/assignment-oracle-final.log 2>&1
go test ./internal/lower -run TestAssignmentValue -count=1 > /tmp/assignment-lower-final.log 2>&1
python3 stage3/notyet-binary/run-assignment-mutants.py > /tmp/assignment-mutants-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts > /tmp/assignment-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(assignment_value|binary_logical|binary_equality)' -count=1 -timeout 15m > /tmp/assignment-regression.log 2>&1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/ir > /tmp/assignment-vet.log 2>&1
```

The eight backend mutants duplicate the right side, duplicate the field receiver, duplicate the element receiver, duplicate its index, yield the old local, yield the old field, yield the old element, and delay a local read. Source Node/native stdout catches each. The ninth drops the assignment-result alias from freshness; `TestAssignmentValuePreservesCycleChecks` then unexpectedly admits the cycle and fails. No build-error mutant is counted. The initial alias mutant did fail compilation; it was corrected and rerun. Every mutation is restored. A tenth removes the flow tracking boundary; the independent tracked-read count fails (graph 1, IR 0).

## Replay

Both chosen original signatures reproduce before this rule, using the required replay command with this branch's assignment dispatch temporarily disabled. After enabling it, neither original signature remains:

| Original table site | Before | After next named stop |
| --- | --- | --- |
| scanner.ts:2466:16, number/number | reproduced, exit 0 | debug.ts:213:28, NotYet `a value of type unknown` |
| sys.ts:156:35, value/value | reproduced, exit 0 | sys.ts:156:17, NotYet `assigning an element of a value` |

Each after replay exits 1 at its missing-original-signature assertion. Complete JSON and logs are retained in assignment-evidence. The initial choices builder.ts:1424:16 and builderState.ts:288:34 stop earlier on `a value of type Path`; they were replaced and are not counted as successful before reproductions.

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-binary-adapted/src/tsc/tsc.ts -where /tmp/notyet-binary-adapted/src/compiler/scanner.ts:2466:16 -kind NotYet -reason 'a BinaryExpression with a number and a number'
go run ./stage3/census/latent/replay -project /tmp/notyet-binary-adapted/src/tsc/tsc.ts -where /tmp/notyet-binary-adapted/src/compiler/sys.ts:156:35 -kind NotYet -reason 'a BinaryExpression with a value and a value'
```

Toolchain setup and nproc are recorded in README.md: setup done 221.364s, nproc 5. No code was copied from cohere and no full repository gate ran. Touched-package runs under the expanded permission are recorded below.

## Touched-package verification under the expanded permission

Worker histories were fetched and checked with `git log --glob='refs/remotes/origin/codex/notyet-*' -- <file>`. No competing worker change to finishAccessors or counters was found. Binary dispatch changes only the equals-token case; no other expression rule is changed. Every file outside assignment() is named in the commit body.

```sh
go test ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh ./internal/lower -count=1 -timeout 15m > /tmp/assignment-packages-focused.log 2>&1
go test ./internal/native -run 'Test.*(Borrow|Reuse|Region|Devirtual|InheritanceMemory)|TestPassThroughsAreNotConsumers' -count=1 -timeout 15m > /tmp/assignment-native-focused.log 2>&1
go test ./internal/flow -count=1 -timeout 15m > /tmp/assignment-flow-final.log 2>&1
```

IR passes in 1.598s; JavaScript has no test files; freshness passes in 63.036s; lowering passes in 44.102s; native ownership/dispatch selection passes in 0.287s. The initial flow run correctly detected that its independent read census had not learned the new untracked-local category. Updating that separate census and proving its mutant fixes it; the entire flow package then passes in 69.460s. The initial full native-package run was interrupted in the unrelated exhaustive ASCII runtime corpus; it is not reported as a pass. No runtime C file changed.
