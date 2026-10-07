# Function expressions, for...in, and library identities

Branch: `codex/function-expressions-for-in`, based on `origin/main` at `fe3b9f2`.
Merged runner `bb766d42ba415def680c8223bc7825e228141a9a` and adaptation
`b3cb87c1af9a8ebb2fe2bb7dba0289721073f927` before measuring or implementing.
The test262 checkout was pinned at `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`.

Ordinary and named function expressions now use the existing counted closure calling convention.
Fixtures cover mutable captures, independent environments, named recursion, self identity, loop
captures, callbacks, Array.from, sorting, throws, and an ordinary function inside a class method.
Named self bindings retain the current closure instead of capturing the variable that owns it.
Dynamic receivers and explicit this parameters are refused; arrows inside ordinary functions cannot
read that ordinary function's this. Generators, async functions, generic expressions, construction
through function values, and function prototype/name/length reflection remain unsupported.

for...in enumerates a proven fixed plain object's actual own enumerable string keys. Canonical
array-index keys sort numerically, followed by ordinary string keys in insertion order. It sees
fields hidden by a narrower static type, preserves spread overwrite order, includes own undefined
values, creates per-iteration let/const bindings, and supports an existing string variable, break,
and continue. The runtime obtains keys from the actual shape, not the static interface.
Arrays and tuples are refused explicitly because holes and enumerable array properties are not
represented. Literal origins, their aliases, and all direct assignments must be proved plain.
Parameters, call results, classes, prototype-setting __proto__ literals, optional structural fields,
possibly absent spread fields, and destructuring writes are conservatively refused. No symbols,
prototype enumeration, property deletion/addition, or getters are admitted by this slice.

Object, Array, String, Number, JSON, Map, and Set can be read directly for strict identity equality
and typeof. Their identity tokens are immortal and have the appropriate runtime kind. Aliasing,
passing, returning, casting, calling through these values, and static-member reads through aliases
remain explicitly NotYet: constructors' overloaded call/new signatures and library statics cannot
soundly be represented as an ordinary closure or fixed object. This is a deliberately partial
implementation of constructor values, not a claim that passed constructor values work.

A small cycle-analysis dispatch hook recognizes the new nodes while preserving the previous
conservative opaque-call reachability model. It records no fictitious writes for these nodes.

A small property-dispatch guard refuses inherited Object.prototype members instead of reading
them as missing own fields. The before survey observed two Node/native exit disagreements under
Object/prototype/hasOwnProperty; the final survey measures this guard as well.

## Setup

`bash cloud/setup.sh` with the tools directory at `/workspace/adamic-tools` printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (12s)
setup: done in 12s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5. Go 1.27.1, clang 20.1.8, Node v24.19.0.
Every build/test command sourced `/workspace/adamic-tools/env.sh`.

## Measurements

Both complete surveys used:

```sh
go run ./cmd/adamic-test262 --adapt --json --test262 /workspace/test262 \
  --work /tmp/adamic-language-BUILD built-ins/Object built-ins/JSON \
  built-ins/Map built-ins/Set built-ins/Array built-ins/RegExp \
  > /tmp/adamic-language-BUILD.json 2> /tmp/adamic-language-BUILD.log
```

BUILD was `before` for the merged runner baseline and `final` for the final implementation.
Intermediate surveys were stopped and superseded when optional-field proof and IR integration were completed;
its incomplete results are not used below. The runner checks each passing native program against
Node's adapted test output and exit status. Skips are runner feature/harness exclusions, not passes.

Before:

| Directory | Pass | Disagree | Refused | Crash | Skip | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Object | 0 | 2 | 2333 | 0 | 1076 | 3411 |
| built-ins/JSON | 0 | 0 | 102 | 0 | 63 | 165 |
| built-ins/Map | 5 | 0 | 78 | 0 | 121 | 204 |
| built-ins/Set | 5 | 0 | 219 | 0 | 159 | 383 |
| built-ins/Array | 62 | 0 | 2409 | 1 | 610 | 3082 |
| built-ins/RegExp | 0 | 0 | 550 | 0 | 1329 | 1879 |
| Total | 72 | 2 | 5691 | 1 | 3358 | 9124 |

After:

| Directory | Pass | Disagree | Refused | Crash | Skip | Total |
|---|---:|---:|---:|---:|---:|---:|
| built-ins/Object | 0 | 0 | 2335 | 0 | 1076 | 3411 |
| built-ins/JSON | 0 | 0 | 102 | 0 | 63 | 165 |
| built-ins/Map | 6 | 0 | 77 | 0 | 121 | 204 |
| built-ins/Set | 6 | 0 | 218 | 0 | 159 | 383 |
| built-ins/Array | 63 | 0 | 2408 | 1 | 610 | 3082 |
| built-ins/RegExp | 0 | 0 | 550 | 0 | 1329 | 1879 |
| Total | 75 | 0 | 5690 | 1 | 3358 | 9124 |

## Verification and mutants

All test output was written to files. Commands and results:

* `go test ./internal/lower -count=1`: passed, 1.886s before the final optional-field boundary addition.
* `go test ./internal/oracle -count=1 -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_' -timeout 10m`: passed, 5.739s, including Node source, JavaScript backend, sanitized native, release native, and leak checks.
* `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts`: passed, 122.577s. The new function, enumeration, and global fixtures allocate/free 56/56, 74/74, and 5/5 respectively; zero region-held values.
* `gofmt -l cmd internal`: empty output.
* `go vet ./...`: empty output, exit 0.
* `go test -count=1 -timeout 30m ./...`: the initial full invocation exited 1 solely because TestEveryWriteIsRecordedAndKnown did not recognize the three new IR nodes. All other packages passed, including the full oracle (759.146s), native (263.674s), flow (158.248s), and all six stage1/cohere packages. After adding the slice integration hook, fresh/lower and the filtered feature oracles were rerun and passed. The full 13-minute suite was not repeated after that hook; this report does not claim a clean single full-gate invocation on the final tree. The original full-gate log is /tmp/adamic-language-gate.log; corrected integration and oracle logs are /tmp/adamic-language-integration.log and /tmp/adamic-language-final-oracle.log.
* `go test ./internal/fresh ./internal/lower -count=1 -timeout 10m`: passed after the missing IR integration hook was added, fresh 41.082s, lower 11.138s.
* The filtered feature oracle command above was repeated after that hook: passed, 22.684s.

One temporary source mutant was run for each feature and restored immediately afterward. Each
compiled, sanitized and release executions exited 0 with empty stderr, and the JavaScript backend
still agreed with Node. Only the Node/native stdout comparison detected the semantic error:

| Feature | Mutant | Node | Native | Detector |
|---|---|---|---|---|
| Named function expressions | Replace ClosureSelf's retained current closure with a freshly made closure | self() === self prints true | prints false | stdout differs |
| for...in | Reverse integer-key insertion-sort comparison from previous < index to previous > index | starts 0, 2, 10, 4294967294 | starts 4294967294, 10, 2, 0 | stdout differs |
| Global values | Alternate each identity read between its stable token and the next token | seven self comparisons print true | all seven print false | stdout differs |

For each mutant the command was:

```sh
go test ./internal/oracle -count=1 \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/FIXTURE.a' \
  -timeout 10m > /tmp/adamic-mutant-FEATURE.log 2>&1
```

FIXTURE/FEATURE pairs: `library_function_expressions`/`function`, `library_for_in`/`for-in`,
`library_globals`/`globals`. Each command exited 1 because the comparison failed. Log runtimes:
5.765s, 6.935s, and 5.406s. Mutants are absent from the final source.

Boundary tests in library_language_test.go cover direct/hidden arrays, parameter origins,
prototype-bearing literals, optional fields/spreads, destructuring writes, dynamic receivers,
constructor aliases/passing, and inherited Object.prototype reads. These refusal tests check
compiler boundaries; the three semantic mutants above prove the Node comparison separately.

The completed surveys observe 3 newly passing tests, 0 lost passes, and 0 final disagreements
(the baseline had 2). The three new passes are:

* `built-ins/Map/constructor.js`
* `built-ins/Set/constructor.js`
* `built-ins/Array/constructor.js`

These constructor tests exercise global identity. No additional complete library tests were
unlocked by function expressions or for...in in these six directories, despite their passing
Node oracle fixtures. Subsequent unsupported operations and strict typing still block them.
The largest remaining refusal bucket is TS7006 (implicit parameter types), 987 tests. Other large
buckets: TS2345 545, TS2339 522, Object as a value outside equality/typeof 453, TS18048 365,
and TS7009 302. These are observations from the final runner report, not estimated unlock counts.

One inherited runner crash remains: Array/prototype/sort/stability-2048-elements.js. A separate
adapted run reproduced it; the captured program.c was exactly 262,144 bytes, while running
`/tmp/adamic-language-crash/adamic c /tmp/adamic-language-crash/program.ts` directly produced
882,561 bytes and exited 0. The runner's 256 KiB limitedBuffer truncates compiler stdout as well
as program stdout. Clang then reports no newline at end of file. This slice does not change that
runner infrastructure or claim that test agrees with Node. The standalone diagnostic report is
/tmp/adamic-language-crash.json and byte counts are /tmp/adamic-language-crash-size.log.

All six directories' after counts above are from the single final completed run, with --adapt.
The baseline and final JSON/log files are /tmp/adamic-language-before.{json,log} and
/tmp/adamic-language-final.{json,log}. Final survey exit: 0; the inherited crash is counted
in the report rather than treated as a disagreement.

