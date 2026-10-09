Implemented immutable intrinsic aliases, proven call/apply adapters, primitive map callbacks and bound String methods.
Node holds both fixtures, both backends, release builds and ownership checks; the uncaught exit-70 convention is unchanged.
The complete test262 comparison is recorded below, including unchanged baseline failures and crashes.
44 valid mutants were caught: 20 behavior, 8 own-field loads, 2 capture ownership, 1 alias read and 13 admission checks.
Implementation commit 57bf6dc1b0e9b1b593c691ba6b4d1c141f34546c is on codex/method-values, from main 5d4c801; the final response records the report commit.

## What works

`const f = Array.prototype.indexOf; f.call(a, x)` and const alias chains retain their declaration identity and checked reads. Array receivers must have a proven dense representation. The implemented prototype adapters are indexOf, includes, lastIndexOf, slice, join, at, push, reverse, concat, fill and splice, subject to their existing argument/element restrictions. An Array result cannot be reannotated with a different element type.

String prototype adapters reuse the existing primitive String implementations with explicit argument lists. Number prototype valueOf and formatting methods, Number statics, fixed-arity Math statics, String.fromCharCode/fromCodePoint and supported Object statics reuse their proven implementations. Existing immediate String and Number prototype calls retain their established paths. Object.prototype.toString uses the prototype method's tags, including `[object Array]` and `[object Null]`, rather than the receiver's own toString result. Object.prototype.hasOwnProperty and propertyIsEnumerable preserve their descriptor restrictions.

`.call` evaluates its alias, this argument and method arguments in source order, including a discarded static-method this argument. `.apply` accepts a dense argument literal, lowering the elements in order without altering the checker's AST. A zero-argument String method can be bound to a present primitive string. An ordinary captured parameter cell snapshots that receiver, survives a factory return and is visible to ownership analysis. The bound function can be called, returned or used by map.

`[].map(String)`, primitive arrays' `.map(String)`, `.map(Number.parseInt)` and `.map(Number.parseFloat)` work, including const callback aliases. The parseInt adapter passes map's index as radix, preserving JavaScript's surprising results. Optional String elements normalize undefined into the actual string `undefined` for String conversion. An empty map still evaluates a callback alias's temporal-dead-zone read.

An intrinsic alias is an opaque token until a supported use supplies its ABI proof. Its read never becomes an own-field load. The expression guard rejects ordinary escapes; the up-front shorthand-field guard closes the path that bypasses ordinary expression lowering. The existing inherited-library-member own-field guard remains in force.

## Refused and not covered

Mutable bindings, reads from instance methods, arbitrary callbacks and alias escapes through fields, returns, identity observations or general arguments remain refused/not-yet. Wrap a call in an ordinary arrow when its receiver or callable ABI cannot be represented. Bound String functions are actual closures and can escape; opaque intrinsic aliases cannot.

Apply of a variable tuple or dynamic argument array, holes, spread, generic array-like receivers and boxed primitive receivers lack an adapter proving length, presence and representations. Bind supports only zero-argument primitive String methods, not Array/Object captures or partial/optional argument adapters. Delayed Object mutation with an erased result shape, Object.prototype.valueOf/toLocaleString, isPrototypeOf and other unimplemented families have explicit reasons. The inherited-library guard still covers them.

The existing library-failure refusal also survives specialization: a probe with `const format = Number.prototype.toFixed; try { format.call(1, -1); } catch { ... }` printed `caught` on Node, and Adamic refused it with `a try around toFixed, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md)`. No native binary was produced. This unit does not change library exception classes or their catchability, nor any exit-code convention.

This is a measured subset, not a claim that the complete library-value wall is gone. The runner's var-to-let adaptation does not turn bindings into consts. The lack of Array/String survey gains is an observation; the const-only and receiver/argument restrictions explain some of the remaining refusals, but are not a claim that every remaining case has the same cause.

## Node fixtures and mutants

`internal/oracle/testdata/library_method_values.a` covers alias chains, call/apply, optional defaults, NaN membership, prototype tags and descriptors, callbacks, array mutation, discarded-this evaluation, captured dynamic strings and contextual primitive unions. `library_method_values_dead_zone.a` compares the ReferenceError's stderr and exit 70 with Node. Finishing programs must leak nothing. Counts are 138 allocations / 138 frees / 176 retains / 322 releases / peak 42 for the first fixture, and all zero for the TDZ fixture. Existing fixture counts are unchanged.

Every behavior mutant finishes successfully, agrees with the sanitizer/leak checks and is killed only by stdout comparison with the original source on Node:

| Mutant | Change |
|---|---|
| search value | Search for 42 instead of the written value |
| apply bound | Replace the written fromIndex with zero |
| last default | Use zero instead of the omitted lastIndexOf bound |
| NaN includes | Search for zero instead of NaN |
| trim | Use trimStart, retaining trailing whitespace |
| substring bounds | Use min instead of max for the upper bound |
| charAt missing | Produce `undefined` instead of the empty string |
| object tag | Return Object for an Array receiver |
| own field | Return false for hasOwnProperty |
| map String | Negate a numeric element inside the actual String adapter |
| parseInt radix | Use radix 10 instead of map's index |
| parseFloat | Use parseInt |
| fixed digits | Use zero digits |
| numeric radix | Use radix 10 instead of 16 |
| Math receiver | Replace abs's written numeric argument with zero |
| character code | Replace code 65 with 90 |
| bound receiver | Replace the factory argument with a different string |
| null tag | Return Object for null |
| ignored this evaluation | Drop the static call's receiver evaluation after its alias read |
| map optional String | Remove undefined-to-string normalization |

Eight valid IR mutants restore an own-field load in place of the aliases named search, trim, own, parse, abs, fixed, character and convert. All compile, then panic with `compiler bug: a field the checker proved is there is missing` and exit 70; the original source finishes on Node, so the oracle kills each mutant.

Removing the receiver-cell retain produces an ASan heap-use-after-free. Retaining that cell twice otherwise agrees with Node but is caught only by LeakSanitizer. Strings built at runtime and a bound function returned from its creating function keep this from being an immortal-literal test. Removing the alias read's Checked flag lets the TDZ mutant finish with exit 0 while the source throws on Node with exit 70; that mutant also finishes without leaks.

`python3 cloud/method-values-mutants.py /tmp/adamic-method-values-source-mutants-complete` runs reversible source mutations. Each is killed by `TestLibraryMethodValueSafety` reporting `unsafe method value was admitted`, not by a diagnostic-wording expectation or a build error. These prove admission boundaries before C generation; they are not counted as native behavior mutants.

| Source mutant | Guard removed or weakened |
|---|---|
| mutable alias | Allow let as well as const, admitting a reassigned parseInt alias |
| opaque field | Allow an intrinsic token to escape through an explicit field |
| opaque return | Allow a token to escape through a function return |
| opaque identity | Allow equality to observe a token |
| opaque shorthand | Remove the distinct shorthand-field guard |
| absent array | Allow a possibly undefined receiver |
| array element erasure | Remove exact element-type/keeping equality |
| apply spread | Truncate the literal at its spread instead of refusing it |
| call spread | Ignore a spread radix argument instead of refusing it |
| string receiver | Accept a numeric receiver for a String method |
| number receiver | Accept a String receiver without a number internal slot |
| bind arity | Admit a method needing a substring argument adapter |
| bind absent receiver | Capture a possibly undefined string |

All 13 printed `caught by admission check`, then `13 source mutants caught; original sources restored`. The script restores every hunk in finally and verifies byte equality with the pre-mutation source.

Uncounted development attempts: the first String-map mutation targeted an unrelated conversion and survived, then was scoped to the actual adapter; the first bound-receiver rewrite tripped an emitter ownership assertion, then was replaced with a valid factory-argument mutation. The first element-erasure source mutation left unused Go variables and failed to build, then retained the nil checks and was killed by unsafe admission. Early discarded-this attempts changed no matching IR because the sequence has an alias operand before its receiver; the final mutant targets the correct operand. None of these invalid/surviving attempts is included in 44.

## Toolchain and commands

Ran `bash cloud/setup.sh` and sourced `/workspace/adamic-tools/env.sh` in each subsequent toolchain shell. `nproc` printed 5. Setup printed:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (78s)
setup: done in 78s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup succeeded. Test output was written to log files.

| Command | Output / log |
|---|---|
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*library_method_values\|TestLibraryMethod' -count=1 -v` | PASS, 8.139s; `/tmp/adamic-method-values-oracle-terminal.log`; zero cached observations, native misses 50, Node misses 30 |
| `python3 cloud/method-values-mutants.py /tmp/adamic-method-values-source-mutants-complete` | All 13 caught and sources restored; `/tmp/adamic-method-values-source-mutants-complete.log` and per-mutant logs in that directory |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | PASS, 13.894s; `/tmp/adamic-method-values-counts-terminal.log`; exactly two rows added |
| `go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./cmd/adamic-test262` | PASS; lower 17.522s, native 89.697s, runner 25.136s; IR/JavaScript have no package tests. `/tmp/adamic-method-values-packages-final.log`; subsequent full gate covers final revisions |
| `gofmt -l cmd internal` | Empty; `/tmp/adamic-method-values-gofmt-terminal.log` |
| `go vet ./...` | Empty, success; `/tmp/adamic-method-values-vet-terminal.log` |
| `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` | PASS, exit 0; all 30 package rows succeeded or had no tests; lower 34.127s, native 300.048s, oracle 239.471s, unicodeproperties 770.497s, stage-1 JSON 494.133s and parser 250.274s; `/tmp/adamic-method-values-gate.log` |

Cohere was run from `/tmp/adamic-method-values-cohere` with exact `.ts` copies, the repository's compiler options and prelude, and `cohere:typescript`. The pinned binary ignores `.a` files, so checking those directly would have checked no fixture. Its formatter's edits were applied back to the `.a` files. `/tmp/adamic-errors-cohere --no-fix --no-cache library_method_values.ts library_method_values_dead_zone.ts` printed `0.03s (276 rules • 2 checked) • 50% Adamic-ready (1 of 2)`, with no findings. Readiness is not claimed to be 100%: documented inline exceptions allow the newly proven intrinsic aliases, the exact returned-array element proof and the intentionally failing TDZ input. No repository-wide lint setting changed.

During development the all-fixture counts command caught immediate `Number.prototype.toString.call(Number.prototype)` being routed through a new delayed receiver path. Preserving the established immediate String/Number paths fixed it; all old counts stayed identical. The external toString probe initially found null receiving the Object tag; adding the exact Null tag fixed the new disagreement. The final measurements, not those interim runs, follow.

## test262 before and after

Both runs use commit `3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd` and the same adapter and complete filters, with no limit:

```
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/scratch/test262 -work /tmp/adamic-method-values-before built-ins/Array built-ins/String built-ins/Object > /tmp/adamic-method-values-before.json 2> /tmp/adamic-method-values-before.log
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/scratch/test262 -work /tmp/adamic-method-values-terminal built-ins/Array built-ins/String built-ins/Object > /tmp/adamic-method-values-terminal.json 2> /tmp/adamic-method-values-terminal.log
```

| Filter | Pass before → after | Fail before → after | Refused before → after | Crashed before → after | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Array | 125 → 125 | 0 → 0 | 2339 → 2339 | 8 → 8 | 610 | 3082 |
| built-ins/String | 205 → 205 | 1 → 1 | 700 → 700 | 1 → 1 | 316 | 1223 |
| built-ins/Object | 17 → 18 | 1 → 1 | 2317 → 2316 | 0 → 0 | 1076 | 3411 |
| Total | 347 → 348 | 2 → 2 | 5356 → 5355 | 9 → 9 | 2002 | 7716 |

The sole new pass is `built-ins/Object/prototype/toString/Object.prototype.toString.call-null.js`. No pass was lost. Both pre-existing failures remain `native failed where node passed (node 0, native 70)`, in String/fromCodePoint and Object/is. The eight Array compiler crashes and one String compiler crash are also unchanged. These are observations from the main baseline, not newly introduced regressions and not claimed fixed by this unit.

The complete machine-readable artifacts are [before](method-values-before.json) and [after](method-values-after.json), including per-directory counts, refusal/skip/crash/failure reasons and every passing path. The survey totals include skipped tests; neither command limited the attempted cases.

Output-file creation-to-final-write intervals were approximately 392 seconds before and 881 seconds after, including startup. The after run overlapped the full uncached repository gate. These artifact timestamps measure the runs' elapsed time under different load, not a compiler speed comparison.

A separate final Node/native probe checked optional numeric String conversion and missing at results: both printed `1|undefined|NaN` followed by `undefined 1`. Its logs are `/tmp/adamic-method-values-optional-{node,build,native}.log`. This additional probe is not counted as a fixture or a mutant.

Staging initially failed with `No space left on device`: Go's regenerable compilation cache had grown to 29 GB. After all checks had completed, `go clean -cache` freed it, leaving 29 GB available. No source, oracle observation, survey or test log was removed.
