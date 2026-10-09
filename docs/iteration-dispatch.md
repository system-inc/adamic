Built step 20 general dispatch for proven object iterables under ruling #85genag.
Implementation: 8925709412da5ba42879dd4a86fe3c47e7cde47c on 1104ae2e; merged main e511df280248e408109103396292b5958da94c68.
Validation: fresh Node in both backends, native ASan/UBSan/leaks, the required sweep and Linux counts.
Mutants: all six caught by stdout disagreement in both backends, with exit 0 and clean stderr.
Pending: malformed .ts protocols need compiler/per-backend-stops 5f3b2e36; built-in storage views need adapters.

## Top finding: fixed wrong output with exit 0

The new value-getter witness exposed a JavaScript backend miscompile. Node printed
`value getter\n`, exit 0; generated JavaScript printed `undefined\nmust not close\n`, exit 0.
Native matched Node. The generated JavaScript read the class getter as an ordinary field,
then entered the body and incorrectly closed the iterator. The done-getter witness exposed
the same fault. Protocol fields and method getters now use the existing accessor metadata;
getter calls enter effect, exception, liveness and ownership analysis. Both witnesses now
match fresh Node in both backends, with clean native sanitizers and leak checks.

Reproducer: [value_throw.a](../stage3/fixtures/iteration-dispatch/value_throw.a).

```typescript
interface Step {value:string;done:boolean;}
class Result {
 get done():boolean {return false;}
 get value():string {throw new Error('value getter');}
}
const source={ [Symbol.iterator](){return {
 next():Result{return new Result();},
 return():Step{console.log('must not close');return {value:'complete',done:true};}
};}};
try{for(const value of source){console.log(value);break;}}
catch(error){console.log(error instanceof Error?error.message:'unexpected');}
```

## What landed toward step 20

Runtime method dispatch accepts iterable classes, generator-shaped objects and structural
interfaces without losing the receiver. GetIterator caches `next` once; each step invokes
that function with the iterator as this. Results can have optional boolean `done` and a
discriminated completion variant with a different value representation. Completion values
are never evaluated by the body. Ordinary reference-counted closures own cached bindings;
there is no new runtime representation or collector.

IteratorClose obtains the current optional `return` on break, return, outer labeled continue,
and body throw. It does not run on exhaustion, next throw, done getter throw or value getter
throw. Original body throws survive a throwing close. Close errors propagate on break/return.
Normal same-loop continue does not close. Built-in arrays, tuples, strings, maps, sets and
native iterator objects preserve their established lowering.

14 witnesses in `stage3/fixtures/iteration-dispatch` pass fresh Node comparison in both
backends: class, object, break, return, labeled_continue, body_throw, next_throw,
close_throw_body, close_throw_break, cached_next, completion, done_throw, value_throw,
method_getters. Native builds use address/undefined sanitizers and leak detection.
The getter witness also checks next is obtained once and return only at close.

Previously refused `iterators_override_this.a` and `iterators_hidden_return.a` now pass
fresh Node comparison in both backends. Existing `user_iterators.a` and its terminal TDZ
witness also retain their Node outcomes. The initial sweep caught tuple/collection paths
being intercepted by general dispatch; preserving the built-in paths fixed those regressions.

## Six mutants

Every mutant modifies lowered IR, compiles and runs separately in each backend, exits 0,
and has empty stderr. Only a stdout disagreement with fresh Node counts as a catch.
Compiler failures and sanitizer findings cannot satisfy the check.

| Mutant | Node witness | JavaScript | Native ASan/UBSan/leaks |
|---|---|---|---|
| Reread next each step | cached_next.a | caught | caught |
| Completion value enters body | completion.a | caught | caught |
| Skip close on break | break.a | caught | caught |
| Close on next throw | next_throw.a | caught | caught |
| Close exception replaces original throw | close_throw_body.a | caught | caught |
| Skip close on outer continue | labeled_continue.a | caught | caught |

## Pending and conservative stops

`compiler/per-backend-stops` **5f3b2e36fd4594b9e93e1cf15c5d03f8f632261b** is not
an ancestor of this base. No dependency branch was merged. As the ruling permits,
non_object_iterator, non_callable_next, non_object_step and non_object_close stay pending
for the .ts terminal-exit-70/backend-divergence proof. Their fresh Node TypeErrors are pinned;
the unchecked casts in their .a versions are refused before code generation. No catchable
TypeError or malformed .ts backend outcome is claimed.

Object views of array, tuple or built-in iterator storage need a protocol adapter. Actual protocol methods
with parameters, overloads, hidden non-object/undefined returns, or absent declared runtime fields stop
explicitly. Zero required arguments alone is insufficient for this native ABI: optional or
default parameters also need padding before admission. Generic classes retain invariant
native arguments. Generators still need owned suspended frames; a generator-shaped ordinary
object is admitted, a function* is not. No new language ruling is requested.

## tsc admission and next stop

The reduction `stage3/fixtures/iteration/user_forwarding.a`, from **src/compiler/core.ts:1736**,
now passes in both backends against Node (`4\n5\n`). A structurally typed iterable object
can forward its iterator method to the existing SetIterator representation.

Fresh no-output replay on the pinned adapted tsc source still stops at
**src/compiler/core.ts:82:17**, firstDefinedIterator: `a function returning U | undefined`.
The function signature blocks measurement of its loop. arrayFrom separately stops at
**src/compiler/core.ts:1337:1**: overload result U[] cannot be served by implementation
result (T | U)[]. Both replay results are explicitly measurements on a checker-rejected
entry-root program, not an executable proof. No claim that all tsc for...of sites compile.

## Fixture records

No accepted/Compiles fixture becomes Refused or NotYet. Every recorded Node stdout stays
byte for byte; a comparison against the parent records checks this. Only two stage0 fields
change in `stage3/fixtures/iteration/outcomes.json`:

| Fixture | Before | After | Reason |
|---|---|---|---|
| user_forwarding.a | NotYet, erased concrete method origin | accepted | Runtime receiver dispatch |
| iterable_view.a | NotYet, for...of over an object | NotYet, object view of built-in storage needing a protocol adapter | Stronger explicit stop; array storage is not an object protocol |

The dispatch commit changes no other stage3 status record. Current main status records are preserved by the merge.

## Merge with current main

The branch merged current main **e511df280248e408109103396292b5958da94c68**, without
rebasing. There were no textual conflicts. Automatic merges preserve both meanings:
JavaScript keeps views v2 checked reads plus iterator accessor dispatch; native reuse keeps
union ownership rules plus getter effects; lowering keeps the earlier array-view iteration
and computed iterator slots plus current view boxing/contracts. Counts retain main's new
proofs and are regenerated on the merged Linux tip. The required build, vet, stage1, stage3,
a-check, Node witnesses and six mutants were rerun on this merge.

The delivered branch also carries the already approved iteration rebuild 1104ae2e. Its six
positive fixture count rows are new relative to main, separately identified below.

## Commands and observed results

Test stdout/stderr are log files in the evidence archive. No full oracle gate was run.

| Command | Result |
|---|---|
| go build ./... | pass |
| go vet ./internal/... | pass |
| a-check on .a changed against origin/main, including untracked sources | 35: 25 checked, 10 expected refused; pass |
| go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 15m | pass |
| ADAMIC_GATE_UNCACHED=1 go test ./stage3/fixtures -count=1 -timeout 15m | pass |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts | pass |
| focused lower/flow/native: Iterator, Iteration, LiteralMethod, ClassWrongOutput, LibraryMapSet, Reuse, Getter, Accessor, Effects | pass |
| fresh Node dispatch witnesses, pending source observations and six mutants | pass |
| fresh Node old iteration witnesses, Step20 guards and TestClassWrongOutput107 | pass |

Exact oracle selectors are recorded in the evidence command list. The focused flow
run passed graph paths, liveness and mutation ranges after getter effects were added.

Toolchain: GOPROXY=https://proxy.golang.org|direct; setup succeeded; source
/workspace/adamic-tools/env.sh. nproc=5, Go1.27.1, Node24.19.0, clang20.1.8.
Setup timing lines: Go ready 0.027s; Node ready 0.028s; markdown step 0.008s,
ready 0.075s; submodules ready 0.089s; clang ready 0.216s; build ready 201.112s;
test binaries deferred 201.196s; build cache warm 201.197s; done 201.231s.

## Linux counts rows

Columns below are allocations, frees, retains, releases, peak live, values freed in regions.
New positive proofs add rows; existing custom iterator rows move because cached method
bindings allocate ordinary closure/cell ownership. The TDZ row terminates deliberately,
so its counts stop at the observed panic. Built-in collection rows do not move.

| Fixture | Previous six counts on main | Regenerated six counts | Explanation |
|---|---|---|---|
| stage3/fixtures/iteration-dispatch/class.a | new | 22, 22, 11, 22, 7, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/object.a | new | 26, 26, 14, 28, 11, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/break.a | new | 20, 20, 11, 18, 10, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/return.a | new | 21, 21, 11, 18, 13, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/labeled_continue.a | new | 44, 44, 22, 38, 12, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/body_throw.a | new | 22, 22, 17, 24, 12, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/next_throw.a | new | 12, 12, 13, 16, 7, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/close_throw_body.a | new | 22, 22, 20, 26, 12, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/close_throw_break.a | new | 20, 20, 16, 21, 10, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/cached_next.a | new | 29, 29, 17, 34, 12, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/completion.a | new | 20, 20, 10, 20, 11, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/done_throw.a | new | 13, 13, 14, 16, 10, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/value_throw.a | new | 13, 13, 14, 16, 10, 0 | New dispatch proof |
| stage3/fixtures/iteration-dispatch/method_getters.a | new | 26, 26, 16, 24, 15, 0 | New dispatch proof |
| internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a | new | 57, 57, 31, 49, 9, 0 | New dispatch proof |
| internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a | new | 28, 28, 16, 20, 8, 0 | New dispatch proof |
| internal/oracle/testdata/user_iterators.a | 663, 663, 455, 904, 64, 0 | 1050, 1050, 677, 994, 97, 0 | Runtime method binding and its owned receiver/cached closure |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5, 0, 5, 3, 5, 0 | 11, 3, 7, 3, 8, 0 | Runtime method binding and its owned receiver/cached closure |
| stage3/fixtures/iteration/collections.a | new | 35, 35, 24, 47, 11, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/strings.a | new | 20, 20, 14, 27, 9, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/object_iteration.a | new | 5, 5, 1, 6, 3, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/array_view_stress.a | new | 34, 34, 50, 72, 16, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/test262_array_views.a | new | 11, 11, 25, 38, 11, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/array_view_weak.a | new | 5, 5, 8, 13, 4, 0 | Previously approved rebuild 1104ae2e, carried on this branch |
| stage3/fixtures/iteration/user_forwarding.a | new | 22, 22, 7, 17, 14, 0 | New dispatch proof |

Evidence: [logs.tar.gz](iteration-dispatch-evidence/logs.tar.gz).
