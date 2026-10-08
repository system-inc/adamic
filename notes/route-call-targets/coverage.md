# Route call target coverage

Base: f64641f388f32d877522140e174dd0b1dcb6d13e.
Branch: coverage/route-call-targets.

Eight new oracle programs, zero differential failures on the unmodified branch.
Every new fixture is registered explicitly in oracle_test.go, with lowers=true and
checked=false, and has a generated counts.md row. Simply placing an .a in testdata
would not register it. Runtime-built labels make freeing the wrong object observable.

## Review

Read CLAUDE.md, README.md, docs/0.1.md and docs/memory.md; reviewed
`git diff origin/main...f64641f`, the commit messages, and the changed code and tests.
The branch's commits introduce bounded closure helpers and a reader guard, restrict
const proofs to literal initializers, extend the guard to external test packages,
merge the borrowed-element work, then route memory analyses through the helpers.
The guard is a Go AST/type-checker constraint, not a runtime language feature.

## Case inventory

Existing means present at f64641f, including its five call_targets fixtures. Paths
in this table are relative to internal/oracle/testdata; omitted extensions are .a.
New entries name stronger combinations where the basic syntax already existed.

| Code condition or value/call shape | Existing oracle evidence | Added evidence |
|---|---|---|
| Direct ir.Call, exactly one target | regions, call_targets_sort | route_targets_recursive: recursive and mutually recursive calls |
| Virtual class call, all implementations including descendants | class_inheritance, class_inheritance_interface, call_targets_element | route_targets_bound_method: a stored arrow calls through Reader to Grandchild; route_targets_virtual_fresh: base, child, grandchild, sibling |
| Structural accessor dispatch (Virtual == -1), getter/setter | class_features_accessors | route_targets_accessor: getter through View removes an element, retained first element is read afterward |
| A target throws / all targets do not throw | class_inheritance_exceptions, closures_throw | route_targets_try_loop: keep, pop, throw in override; base returns; alias read in finally |
| Inline MakeClosure target | closures, borrow_element | route_targets_bound_method.literal: direct immediately invoked writer |
| Const initialized directly with an arrow, ConstantClosure != 0 | call_targets_element.boundRead and boundWrite, call_targets_closure | route_targets_bound_method: literal captures receiver; route_targets_callbacks: const callbacks |
| Mutable binding, ConstantClosure == 0, Unknown | call_targets_closure, call_targets_element | route_targets_sort_values: changed comparator |
| Parameter function value, Unknown | named_function_values.apply, closures_throw.attempt | route_targets_unknown.inspect/invoke: closure passed through two functions removes the last element |
| Property function value, Unknown | named_function_values.table, library_fnexpr_store.holder | route_targets_unknown.forms(0): property removes element; route_targets_sort_values.holder |
| Returned function value, Unknown | closures.makeGreeter/adder, named_function_values.pick | route_targets_unknown.forms(1): call returned arrow directly; route_targets_sort_values.returned |
| Conditional/join function value, Unknown | named_function_values.pick | route_targets_unknown.forms(2): conditional used directly as callee; route_targets_sort_values conditional comparator |
| Const alias of a named function remains Unknown | named_function_values.held | route_targets_unknown.forms(3): named remover alias; route_targets_sort_values.compare is also called directly |
| Captured cells and structural receiver on CallClosure remain conservative in flow/fresh | closures, method_closures, class_inheritance_interface | route_targets_bound_method captures reader; route_targets_unknown returns a closure and calls property/parameter values |
| ArrayMap callback | closures, closures_throw | route_targets_callbacks const map |
| ArrayVisit callbacks: forEach, filter, find, some, every | visits, closures_throw, weak_narrowed | route_targets_callbacks const visit/test across all five |
| ArrayReduce callback | searches, method_closures, closures_throw | route_targets_callbacks const reduce |
| ArrayFrom callback | array_from, closures_throw | route_targets_callbacks inline callback; const form cannot lower |
| MapForEach, Map and Set | collections, closures_throw | route_targets_callbacks const mapVisit/visit |
| ArraySort named comparator (Callback == nil) | call_targets_sort, sorts, closures_throw.strict | route_targets_sort_values.compare used by wrappers; original named fixture covers direct sort |
| ArraySort literal/const bounded callback | sorts, closures_throw, library_function_expressions | route_targets_sort_values: fixed const comparator directly passed to sort |
| ArraySort Unknown callback: conservatively mark every function a comparator and clear fresh direct-function set | library_fnexpr_store.holder.compare | route_targets_sort_values: parameter sort reached by literal, const, mutable, property, returned and conditional values, with spread in compare |
| Bounded harmless closure can allow element borrowing | call_targets_element.boundRead | existing case retained; counts and native plan test check optimization |
| Any bounded changing target prevents borrowing | call_targets_element.boundWrite | route_targets_bound_method: transitive virtual mutation and direct literal mutation |
| Unknown closure prevents borrowing | call_targets_element.closureRead | route_targets_unknown: parameter, property, returned, conditional, named alias removers |
| All direct/virtual targets harmless versus any target changing | call_targets_element.virtualRead, borrow_element_virtual_store | route_targets_accessor and route_targets_bound_method; route_targets_recursive propagates changes to fixed point |
| Mutable reference values (objects, arrays, maps/sets, closure cells, unions and Weak) versus numbers, booleans, strings and void | generic_values (object/array/map/set/closure and primitive arguments/results), unions.pickOne, weak_parent.depth, undefined_references | route_targets_callbacks uses string/boolean/void callbacks; route_targets_virtual_fresh and recursive use object and number results |
| Top-level fresh analysis versus summary call; non-direct targets may keep arguments; mutable versus scalar/void returns | call_targets_region, fresh_refused/call_targets_virtual, regions | route_targets_virtual_fresh returns Item objects and numbers; route_targets_try_loop returns void; route_targets_recursive returns objects/numbers and calls at top level |
| Fresh call joins each target's state from the same pre-call state | call_targets_region, call_targets_reuse | route_targets_virtual_fresh: several descendants produce fresh objects, one retains input |
| Region permitted for one direct fresh target; refused for nonfresh or virtual calls even when all targets are fresh | regions, call_targets_region | route_targets_recursive: mutually recursive fresh make/other; route_targets_virtual_fresh: all virtual producers fresh but heap ABI remains required |
| Region argument cannot escape any target; any keeping/unknown-position target prevents region use | regions, call_targets_region, class_inheritance_memory | route_targets_try_loop: keeper throws after save; route_targets_recursive: recursive drop and recursive keep |
| Region context absent or call at another expression depth | regions, regions_throw | route_targets_virtual_fresh: nested direct and virtual producers |
| Call consumes an argument only if every target consumes that position; otherwise ordinary borrowed argument path | call_targets_reuse, class_inheritance_memory, borrow_element_virtual_move | route_targets_virtual_fresh: one override spreads/replaces, others return same input; all ABI conventions joined before calls |
| Read argument moves only if consumed; non-Read arguments still handed over with a count | call_targets_reuse, move_throw | route_targets_virtual_fresh: local item handed to replace; route_targets_try_loop: fresh make result handed to virtual act |
| Global touched by an override prevents moving it away before call | call_targets_reuse.observe, class_inheritance_memory | existing case retained |
| Calls in loops, try/catch/finally, throws after retaining input | borrow_element_throw, regions_throw, closures_throw | route_targets_try_loop: three iterations each for returning base and throwing override |
| Recursive fixed points, with and without mutation/escape | regions (build/count), named_function_values.factorial, declared_later.isEven/isOdd and Walker.down/up | route_targets_recursive: mutually recursive writer, mutually recursive fresh return, recursive keep/drop |

The helpers enumerate supported call forms, not arbitrary TypeScript syntax.
Flow and fresh deliberately retain conservative CallClosure effects even for a
bounded target; this branch does not promise precise captured-cell summaries.

## Cases that cannot be expressed as successful oracle programs

- Detached class/interface methods are refused by internal/lower/refusals.go's
  unbound-method rule (line 136). Used the prescribed `(items) => reader.read(items)`
  wrapper stored and invoked later. Function-valued fields are covered separately.
- Array.from with a const, parameter or property callback cannot lower on this
  revision: "a callback that isn't an arrow function written in place". Its literal
  arm is covered; the other ClosureTargets arms cannot be reached from this source.
- Missing virtual target sets, an unsupported expression passed to ClosureTargets,
  or a direct CallTargets set with zero/multiple targets are malformed IR. Valid
  lowering cannot produce them; they need Go-level IR tests, not .a programs.
- A mismatch in consumed conventions among virtual implementations cannot survive
  planReuse's convention join. Successful source probes test the joined behavior;
  they cannot force inconsistent internal maps or missing parameter positions.
- A self-capturing arrow initializer cannot lower (locals.go's initializer-capture
  check). Recursive named functions and a pair of mutually recursive named functions
  test the supported alternative.
- Draft probes using shift() and a const callback with an unknown parameter were
  rejected before execution. Used splice(0, 1) and the supported inline Array.from
  callback instead. These were compiler limitations, not differential failures.

## Mutation evidence

Temporarily changed internal/native/element_borrow.go:167, inside the CallClosure
arm, from `if changing[target] {` to `if false && changing[target] {`.
Ran the new route_targets_bound_method.a oracle alone uncached. It compiled and
failed at runtime (test exit 1), with ASan heap-use-after-free in inspect. This was
not a clang or -Werror failure. The release binary exited 0 with incorrect stdout.

Node and the restored native program:

```text
1/method0/1
1/method1/1
0/method2/0
literal3/0
```

Mutated native release program:

```text
1/method0/1
1/method1/1
0/0/0
0/0
```

Sanitized mutated native: exit 1, stdout empty, ASan heap-use-after-free.
The exact original line was restored; git diff for element_borrow.go is empty.
The final gate and standalone builds run with restored code.

## Commands and observations

All commands below ran from /workspace/adamic. Every Go/build command used
`source /workspace/adamic-tools/env.sh`. The first setup attempt raced the branch
checkout during cache warming and failed from mixed IR revisions; the second
completed on the fixed branch. No toolchain or compiler fixes were committed.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
git fetch origin main codex/route-call-targets
git log --format='%h %s%n%b' origin/main..f64641f
git diff origin/main...f64641f
git switch -c coverage/route-call-targets f64641f
```

Successful setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (90s)
setup: done in 90s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc: 5. Go 1.27.1, clang 20.1.8, Node v24.19.0.

Oracle commands, including exploratory rejected shapes:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/route_targets_' -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/route_targets_(callbacks|virtual_fresh)' -count=1 -timeout 10m
# With the one-line mutant, then restore:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/route_targets_bound_method.a' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
```

Test output was redirected to /tmp/adamic-gate/route-*.log and read afterward.
The final eight-program oracle run passed (1.202s); counts update passed (12.735s).
Formatting and vet passed. The complete uncached gate exited 1 solely because
TestMarkdownUnicodeWidths could not find its external emoji-regex dependency.
All other packages passed, including internal/oracle (215.552s), internal/native,
internal/flow, internal/fresh, internal/ir and internal/lower.
The existing markdownblocks/REPORT.txt specifies the missing scratch packages.
Installed those exact pins, then reran the sole failed test uncached:

```sh
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -count=1 -timeout 30m
```

Install: added 3 packages in 733ms. Retry: PASS, 128.831s, exit 0.
No repository changes were required for this repair. The full gate was not
repeated after the repair; all its other checks had already passed.
The original full output is in gate.log; the repaired check's result is in
width-retry.log.

Standalone builds (after restoration), each followed by execution and a Node diff:

```sh
mkdir -p /tmp/adamic-gate/route-call-targets
for file in internal/oracle/testdata/route_targets_*.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/route-call-targets/$name" || exit 1
 "/tmp/adamic-gate/route-call-targets/$name" > "/tmp/adamic-gate/route-call-targets/$name.native" 2> "/tmp/adamic-gate/route-call-targets/$name.stderr" || exit 1
 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" > "/tmp/adamic-gate/route-call-targets/$name.node" || exit 1
 diff -u "/tmp/adamic-gate/route-call-targets/$name.node" "/tmp/adamic-gate/route-call-targets/$name.native" || exit 1
 echo "$name agrees"
done
```

All eight standalone binaries exited 0, stderr was empty, and stdout matched Node.
The oracle also checked the JavaScript backend, ASan/UBSan and leaks.

## Commit and push

```sh
git add internal/oracle/oracle_test.go internal/oracle/counts.md internal/oracle/testdata/route_targets_*.a notes/route-call-targets
git commit -m "Cover routed call targets with oracle programs"
git add -f notes/route-call-targets/gate.log notes/route-call-targets/width-retry.log
git add notes/route-call-targets/coverage.md
git commit -m "Record gate output and the repaired width check"
git push -u origin coverage/route-call-targets
```
