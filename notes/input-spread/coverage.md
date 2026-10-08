# Input spread coverage

Base: origin/compiler/input-spread, 85328b26aa5c0c614ea21c6e0455ece8185e6afe.
After refreshing origin/main, the requested three-dot diff contains only input.go and input_test.go.
The commit adds a guard that refuses call spreads into seven imported prelude functions.
No runtime output disagreements were found. Two runnable oracle programs and 47 diagnostic .a
programs were added. The diagnostics live in internal/lower/testdata/input-spread and are checked
by TestInputSpreadCoverage. Intentionally checker-invalid array probes have a narrow cohere ignore.

## Cases and previous coverage

| Function and accepted arguments | Existing oracle programs with ordinary calls | New diagnostic files |
|---|---|---|
| readTextFile(string) | read_files.a, read_arguments.a, utf8_sweep.a | readTextFile_{tuple,nested,alias,parentheses,array}.a |
| writeTextFile(string, string) | write_files.a, write_stdout_order.a, write_stderr_order.a | writeTextFile_{tuple,nested,alias,parentheses,array,later,multiple}.a |
| utf8Length(string) | utf8_view.a, shared_slices.a | utf8Length_{tuple,nested,alias,parentheses,array,empty}.a |
| utf8At(string, number) | utf8_view.a, utf8_view_fails.a | utf8At_{tuple,nested,alias,parentheses,array,later,multiple}.a |
| readDirectory(string) | walk.a | readDirectory_{tuple,nested,alias,parentheses,array}.a |
| fileStatus(string) | walk.a, status_of_stdout.a | fileStatus_{tuple,nested,alias,parentheses,array}.a |
| programArguments() | arguments.a, read_arguments.a, read_files.a | programArguments_{tuple,nested,alias,parentheses,array}.a |

No existing oracle program exercises a spread into these prelude calls. The branch's existing
lower test already checks a single tuple per function, with strings, a numeric index, and an empty
tuple for programArguments. New files make those programs independently buildable with the CLI.

Every function has direct tuple, nested tuple literal spread, renamed import, parenthesized callee,
and array probes. Tuple/nested/alias/parentheses cases produce NotYet naming the original function,
not the renamed import. Two-argument functions also have a normal argument before the spread and
multiple spreads. utf8Length_empty proves that even an empty spread before a normal argument is
refused syntactically. programArguments covers an empty spread and the no-argument loop path is
already covered by arguments.a. All named diagnostics assert the call's source line.

The guard's unmatched-callee path is checked by input_spread_local.a: ordinary local functions
with all seven names receive strings built from tuple elements, arrays and nested array literals.
The no-spread argument path is checked by input_spread_ordinary.a: aliased/parenthesized UTF-8 calls,
a runtime-built string with BMP, supplementary and lone-surrogate units, a numeric index, and an
array spread inside an ordinary string argument. Array-literal spreads are distinct from call spreads.
Seven additional _local.a probes call same-named local functions with tuple spreads. They assert a
general NotYet without incorrectly naming a prelude function; general call spreads do not lower yet.
Both runnable programs match source Node, JavaScript backend, sanitized native, release native,
and the leak check. Their count rows are the only changes to counts.md.

## Limits

All seven ordinary-array call spreads fail Load with the checker's message:
"A spread argument must either have a tuple type or be passed to a rest parameter."
The prelude signatures have fixed arity, including programArguments's zero arguments. A sound
program cannot pass such arrays through Load to the new guard. We assert this earlier CheckError
rather than bypassing the checker or claiming that it is the branch's NotYet.

Wrong arity and wrong argument types likewise cannot reach input.go's internal-error fallbacks in
checker-accepted source. These unchanged defensive branches are not runtime features.
General rest-parameter functions are also not lowered ("a parameter that isn't a plain name").
Tuple join is not lowered either (tuples are held as objects). The local control uses supported
ordinary parameters and indexed tuple reads, so neither unrelated limitation is in oracle testdata.
Rejected programs cannot produce native binaries, so they have no native output to compare or run.

## Mutants

At input.go's new argument-kind test, change == to !=. The new input_spread_ordinary.a oracle
case fails at line 4 with "stage 0 can't lower a spread argument to utf8Length yet". This proves a
new runnable oracle program observes the branch's guard. No clang warning is involved.

At the same line, add false && before the condition. The writeTextFile_tuple.a lower case fails:
"writeTextFile takes a path and a text, and the checker let another count through", instead of NotYet.
Both one-line mutants were restored in a finally block, and lower and oracle coverage passed again.

## Toolchain and commands

bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc

Setup printed: go ready (0s); clang ready (1s); node ready (1s); submodules ready (1s);
build cache warm (75s); done in 75s on 5 processors (cpu.max 400000 100000), 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc: 5.

Verification commands (test output redirected to /tmp/adamic-gate/input-spread-*.log, then read):

```sh
go test ./internal/lower -run 'TestInputSpread' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/input_spread_' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
for file in internal/oracle/testdata/input_spread_*.a; do
  go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/$(basename "$file" .a)"
  "/tmp/adamic-gate/$(basename "$file" .a)"
done
```

A Python subprocess loop also ran `go run ./cmd/adamic build <file> -o
/tmp/adamic-gate/input-spread-rejected` on each of the 47 diagnostic files, requiring failure
and the expected message. Its full observed diagnostics are in /tmp/adamic-gate/input-spread-cli-diagnostics.log.
Test logs stay outside the commit, as .gitignore directs.

Mutant commands:

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/input_spread_ordinary.a' -count=1 -v
go test ./internal/lower -run 'TestInputSpreadCoverage/writeTextFile_tuple.a' -count=1 -v
```

After restoring the code:

```sh
go test ./internal/lower -run 'TestInputSpread' -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/input_spread_' -count=1
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
```

During fixture development the same targeted lower/oracle/counts commands were also run on initial
versions: the lower location assertion incorrectly assumed line 3 for two line-4 calls; local rest
parameters then tuple.join were refused. Those fixture/test issues were corrected before the passing
runs above. These were compile-time limitations, not observed differences in runtime output.
Git inspection: git log --oneline origin/main..origin/compiler/input-spread;
git diff origin/main...origin/compiler/input-spread; git show 85328b2; git diff --check.
The first fetch populated FETCH_HEAD only; fetching explicitly into refs/remotes/origin/compiler/input-spread
and refreshing origin/main made the requested diff refer to the current main. The branch was created
with git switch -c coverage/input-spread origin/compiler/input-spread.

Final validation: formatting and go vet were clean. The complete uncached gate exited 0;
internal/native passed in 230.400s, internal/oracle in 176.631s, and the longest suite,
internal/unicodeproperties, in 674.483s. All remaining packages passed or had no test files.
The final 47-file targeted lower coverage passed in 3.700s, and all 49 new .a files were tried
with the CLI: two built and ran, 33 gave named NotYet, seven gave general NotYet, and seven
were rejected by the checker. The compiler source is restored without a diff.
