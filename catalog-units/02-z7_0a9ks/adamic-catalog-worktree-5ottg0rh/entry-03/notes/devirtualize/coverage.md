# Devirtualization coverage

Base: origin/main e8ba3d5. Inspected git diff ceb29f7 fe522c5 and both commit messages, all changed compiler and test files, docs/devirtualize.md and bench/devirtualize.a. The branch changes native emission only.

| Condition or value | Existing oracle evidence | Added evidence |
| --- | --- | --- |
| Direct nonvirtual call has one target | functions.a, class_inheritance.a (super) | modules/main.a (super) |
| Virtual slot has one implementation, including inherited slots | devirtualize.a (stable) | devirtualize_abstract.a (stable and visit) |
| Virtual slot has multiple implementations, late declarations | devirtualize.a | devirtualize_receivers.a, cross-module subclass |
| Exact temporary allocation selects overridden slot | devirtualize.a | devirtualize_generic.a |
| Exact uncaptured local initialized with allocation | devirtualize.a | devirtualize_interfaces.a |
| Allocation call must be nonvirtual, actual nonstatic constructor | devirtualize.a | free factory in devirtualize_receivers.a; virtual factory result and static factory in devirtualize_factory.a |
| Static classes do not establish exact instance identity | class_features_static.a, class_features_static_private.a | existing evidence sufficient |
| Captured local is not exact | devirtualize.a (global mutable), method_closures.a | local captured and reassigned in devirtualize_receivers.a |
| Assigned local is not exact, even assignments inside nested bodies | devirtualize.a | if, for, while, do, switch, try, catch, finally in devirtualize_receivers.a |
| One declaration with allocation required; read aliases not followed | class_as_interface.a (fields, parameters, returns) | alias, holder, factory and conditional in devirtualize_receivers.a; interface alias and return |
| Parameters, joins, property reads have no exact fact | class_as_interface.a, class_inheritance_interface.a | devirtualize_receivers.a and devirtualize_interfaces.a |
| Interface exact named method uses adapter | devirtualize.a, class_as_interface.a | both implementations and inherited method in devirtualize_interfaces.a |
| Interface unknown receiver retains method/closure lookup | class_as_interface.a, devirtualize.a | parameter, literal and argument-time receiver replacement |
| Interface ABI number, boolean, string, object input and void result | class_as_interface.a | string adapter in devirtualize_interfaces.a |
| Named method absent from class method table: own function field needs closure lookup | regexp_cycle_closures.a has class function fields but not this interface case | devirtualize_function_field.a: exact interface, inherited field and saved function |
| Method not dispatchable cannot use exact adapter | union and boolean-or-undefined function-value calls are NotYet | no accepted source can reach these ABI cases |
| Generic instances have separate targets | devirtualize.a, class_inheritance_generic.a | base/override number and string in devirtualize_generic.a |
| Abstract slot implementations | class_inheritance.a | devirtualize_function_field.a | /tmp/devirtualize-function-field |
| devirtualize_factory.a | /tmp/devirtualize-factory-final |
| devirtualize_abstract.a |
| Method calls in closures and loops | method_closures.a, class_as_interface.a | saved arrow wrapper, captured local, abstract this, nested loop reassignment |
| Super calls remain lexical direct calls | class_inheritance.a, class_inheritance_interface.a | cross-module super and generic super |
| Getters and setters, multiple implementations (separate accessorCall path, not optimized by callCode) | class_features_accessors.a | devirtualize_accessors.a |
| Throws and exceptional/result cleanup | devirtualize.a, class_inheritance_exceptions.a | existing evidence sufficient; assignment in catch/finally added |

All added names above are under internal/oracle/testdata; modules means devirtualize_modules. Eight entry programs, ten .a fixture files including the two module dependencies.

Cannot express duplicate declarations of the same IR local in accepted source: distinct lexical declarations have distinct local IDs. Missing target sets are malformed IR, not source programs. No target and multiple declaration guards are compiler invariants. A detached instance method is forbidden by the language; saved-call probes use an arrow wrapper or a function-valued field, whose closure already holds its capture. Calling an overridable method from a base constructor is refused before lowering; constructor_refused.a records the attempted probe. Node prints ready/base, ready/child, base/child, but native compilation is refused, so this is not an output disagreement.

Setup: go ready 0s; clang ready 0s; node ready 0s; submodules ready 0s; build cache warm 223s; done 223s. Environment sourced /workspace/adamic-tools/env.sh. nproc: 5; CPU quota: 400000 100000.

## Mutation proof

Temporarily changed internal/native/class_inheritance.go:48 from `len(targets) == 1` to `len(targets) >= 1`, ran the uncached cross-module oracle, and restored the file in a Python finally block. Both Node and native exited zero, without sanitizer or clang failures. Node printed:

```text
base-stable/base-changing
base-stable/child-changing
base-changing
```

Native under the mutant printed:

```text
base-stable/base-changing
base-stable/base-changing
base-changing
```

The oracle failed with stdout differs and exit 1. Restored focused oracle: pass (1.638s). This is an intentionally introduced failure, not a disagreement in the submitted branch.

## Commands run

All Go/Node/build commands sourced `/workspace/adamic-tools/env.sh` in their shell. Every test wrote to a log file, which was read afterward.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
git fetch origin
git diff ceb29f7 fe522c5
git show -s --format=full 6965ac2 fe522c5
git switch -c coverage/devirtualize origin/main
```

Focused oracle command, run for the initial six programs, expanded receivers/interfaces, attempted constructor probe (failed with Refused), final six programs, after restoring the mutant, and for the final seven-program set:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/devirtualize_' -count=1 -timeout 30m
```

Counts command (run for both six-program validations, then for the final seven-program set):

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

Each of the following source files was built with `go run ./cmd/adamic build <file> -o <out>`, its output executed, and source run with `node --disable-warning=ExperimentalWarning oracle/node.mjs <file>`:

| File under internal/oracle/testdata | Output |
| --- | --- |
| devirtualize_function_field.a | /tmp/devirtualize-function-field |
| devirtualize_factory.a | /tmp/devirtualize-factory-final |
| devirtualize_abstract.a | /tmp/devirtualize-devirtualize_abstract |
| devirtualize_accessors.a | /tmp/devirtualize-devirtualize_accessors |
| devirtualize_generic.a | /tmp/devirtualize-devirtualize_generic |
| devirtualize_interfaces.a | /tmp/devirtualize-devirtualize_interfaces |
| devirtualize_receivers.a | /tmp/devirtualize-devirtualize_receivers |
| devirtualize_modules/main.a | /tmp/devirtualize-main |
| devirtualize_modules/base.a | /tmp/devirtualize-module-base |
| devirtualize_modules/child.a | /tmp/devirtualize-module-child |

The constructor probe was also attempted with `go run ./cmd/adamic build internal/oracle/testdata/devirtualize_constructor.a -o /tmp/devirtualize-constructor`. Its first draft had standalone new expressions, rejected as NotYet; changing them to const declarations exposed the base-constructor Refused diagnostic. Source Node ran both drafts. The final draft is notes/devirtualize/constructor_refused.a; no native executable was produced.

```sh
# While the one-line mutant was active:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/devirtualize_modules/main.a' -count=1 -timeout 30m
# After restoration:
gofmt -w internal/oracle/devirtualize_coverage_test.go
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
git diff --check
```

The final seven-program focused oracle passed in 10.677s. The seventh program was added while the first full gate was running; the final complete oracle package is rerun with all seven fixtures registered. No other package is changed by the seventh registration.

The initial whole-repository gate used a test binary registered before the seventh fixture was added. Its counts check compared a six-fixture list against the final seven-row addition and failed (no existing row moved). This was a validation scheduling mistake, not a compiler output disagreement. All earlier fixture comparisons passed. The complete final oracle package, including counts without updating them, is rerun after registration and counts are stable:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m
```

Formatting and go vet are also repeated against the final registration.

The complete uncached oracle with the stable seven-program registration passed (329.244s), including the counts assertion without updating it. The last audit found the virtual-factory-result receiver guard was not exercised by an existing oracle program; the standalone factory probe agreed on Node/native and was held outside registration until that full oracle run finished. It was then moved to devirtualize_factory.a and registered as the eighth program. Its initial CLI build and Node run used notes/devirtualize/factory_probe.a before the move. Final counts are regenerated and the complete oracle rerun with the final eight-program set.

Unsupported ABI probes, outside testdata:

- maybe_boolean_refused.a: Node prints undefined; build fails at line 6 with NotYet: passing boolean | undefined to a function value.
- union_refused.a: Node prints union; build fails at line 4 with NotYet: a function value returning union of differently held members.

Both were tested with `go run ./cmd/adamic build notes/devirtualize/<name>.a -o /tmp/devirtualize-<name>` and `node --disable-warning=ExperimentalWarning oracle/node.mjs notes/devirtualize/<name>.a`. There is no native executable or output for either, so these are unsupported cases, not output disagreements.

## Final observations

Eight oracle entry programs and two module dependencies were kept; none disagreed. Every fixture .a file was built and executed separately and compared with source Node. The final eight-program uncached oracle passed in 1.789s, including source Node, native sanitizer and release builds, JavaScript backend, and leak checks. Final counts update passed in 51.452s; the subsequent counts assertion without -update-counts passed in 21.062s. Formatting, go vet, and git diff --check passed. Compiler implementation files have no diff; the singleton mutant was restored.

The complete seven-program oracle passed in 329.244s before the last factory registration. The final complete eight-program oracle attempt and the remaining whole-repository gate processes were terminated by the environment refresh, with their files preserved but tool process handles lost. The whole gate had already reported the stale registration/counts mismatch described above; its bridge, flow, fresh, fuzz, IR, load, lower, native and regexp packages had passed. Remaining packages have no completed result, so the whole gate is not claimed as passed. After refresh, all eight new programs and the strict counts assertion were rerun successfully with stable inputs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/devirtualize_' -count=1 -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m
gofmt -l cmd internal
go vet ./...
git diff --check
```
