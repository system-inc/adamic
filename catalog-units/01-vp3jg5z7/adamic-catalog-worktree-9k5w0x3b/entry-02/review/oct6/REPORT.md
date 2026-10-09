# October 6 coverage

Branch `codex/coverage-oct6` starts at inheritance commit `281ee9d25b7189b15794d13daa8e217b1bd414e2`. Matcher commit `6766c9de7bb730d6b95467f461f9d8b2670be8e7` was merged in `6598037`. The task-specific inheritance base takes precedence over the general main-base instruction. The checkout fetched only main by default, so both feature refs were fetched explicitly. No compiler implementation files were changed by this coverage work.

Added four executable inheritance fixtures, registered in `internal/oracle/oracle_test.go`, with four generated counts rows. Added 30 matcher pattern/input cases and 38 stateful loops (179 sequential executions). All 72 executable scenarios agree with Node; there are zero observed runtime disagreements. One additional construction probe is preserved here because the compiler refuses it.

| Fixture | Coverage | Allocations/frees |
|---|---|---:|
| class_oct6_deep.a | Four levels, overrides inherited past Second, super calls at Second/Third/Fourth, mixed base array, instanceof at every level, fields initialized from earlier fields | 43/43 |
| class_oct6_parameters.a | Base and derived constructor parameters, transformed super arguments, chained field initializers, inherited fields in Grandchild | 43/43 |
| class_oct6_release.a | Dynamic strings, arrays and objects held at four depths, base-typed replacement with a surviving alias, mixed array pop, function and loop scope releases | 372/372 |
| class_oct6_subclass_holder.a | Parent holds Child and Grandchild instances, replacing and clearing the child while an alias survives, all released | 76/76 |

The shared inheritance oracle checks source Node, generated JavaScript, sanitized native, native release and leaks. The matcher tests compare success/failure, every UTF-16 capture span, undefined captures, named capture spans and lastIndex against Node 24. Loops reuse the same RegExp on each side; they cover failure resets, restart after failure, g/y/gy, u/v, astral input, lastIndex inside a surrogate pair, unchanged empty-match indices and explicit caller advancement. Every loop has a fixed bound, including deliberately repeated empty matches.

## Construction probe: observed outputs

`review/oct6/class_oct6_construction.a` has a base constructor calling a virtual method overridden at two deeper levels. Node runs the Leaf override before Middle and Leaf fields have been initialized.

Node stdout, exit 0, empty stderr:

```text
leaf report base1/undefined/undefined
leaf report base1/middle2/leaf3
```

Adamic build stdout is empty, exit 1, stderr:

```text
adamic: /workspace/adamic/review/oct6/class_oct6_construction.a:3:21: Adamic 0.1 refuses this escaping a base constructor before derived fields are initialized; use this only to read or write initialized base fields; call methods and publish the object after construction
```

There is no native executable or native runtime output for this probe. This is an explicit language refusal, not an observed silent miscompile. The shared harness has flags for NotYet and inserted checks, but neither represents this Refused result, so it is excluded from the fixture list and saved under review as requested.

Reproduce:

```sh
source /workspace/adamic-tools/env.sh
node --disable-warning=ExperimentalWarning oracle/node.mjs review/oct6/class_oct6_construction.a > /tmp/oct6-construction-node.log 2>&1
go run ./cmd/adamic build review/oct6/class_oct6_construction.a -o /tmp/oct6-construction > /tmp/oct6-construction-native.log 2>&1
```

The latter Go wrapper adds `exit status 1` after the diagnostic. The initial attempt omitted required `-o`, producing usage text; the corrected command above produced the actual refusal.

## Mutants run and caught

All mutants are retained as tests that mutate in-memory IR, compiled matcher bytecode or generated C strings. No compiler source mutation is required. Each semantic mutant executes normally and differs from external Node; the leak mutant builds with -Werror and preserves stdout, stderr and exit status when leak detection is disabled.

| Mutant | Witness and check that caught it |
|---|---|
| Fourth dispatches First.inherited rather than Second.inherited | Fourth prints `first method` instead of `second method`; native stdout oracle |
| Fourth loses its base ancestry | `false/false/false/true` instead of `true/true/true/true`; native stdout oracle |
| Generated release calls become no-ops | Output still exactly equals Node; LeakSanitizer reports 26,914 bytes in 395 allocations |
| Nested repeated captures are not cleared | `^((a(b)?)+)+$` on `aba`: capture 3 remains [1,2] instead of undefined |
| Lookbehind alternatives have reversed priority | `(?<=((a)\|(ba)))c` on `bac`: capture 1 [0,2] instead of [1,2], capture 2 undefined instead of [1,2] |
| Repeated backreferences are ignored | `^((a\|b)\2)+$` on `aaba`: mutant matches [0,4], Node fails |
| Non-ASCII backreference loses IgnoreCase | `^(é+)\1$` with i on `éÉÉé`: mutant fails, Node matches [0,4] with capture [0,2] |
| Global flag removed | First `a+` match on ` a aa` from index 1 leaves lastIndex 1 instead of 2 |
| Sticky becomes searching global | `(a)\|(b)` on `ab ab`, step 2: mutant skips gap to [3,4] with lastIndex 4; Node fails and resets to 0 |
| Empty pattern bytecode consumes a unit | Empty g match on `ab` from index 1 returns [1,2], lastIndex 2 instead of [1,1], lastIndex 1 |

The empty-pattern mutant changes the compiled pattern to consume a code unit. It proves the empty-span and lastIndex oracle can fail; it does not specifically mutate Exec's empty-match index update. The leak mutant omits every generated release, rather than targeting one destructor field. These are limits of the mutation proof, not broader compiler correctness claims. macOS uses the mutated unsanitized executable under leaks; only Linux was run here.

## Setup and validation

`bash cloud/setup.sh > /tmp/coverage-oct6-setup.log 2>&1` succeeded. The printed environment file is `/workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node v24.19.0; `nproc` printed 5, cgroup quota four CPUs.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (50s)
setup: done in 50s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Every test command wrote to a log without a pipe. Commands and observed results:

```sh
source /workspace/adamic-tools/env.sh
go test -v -count=1 -timeout 10m ./internal/regexp > /tmp/oct6-regexp-full.log 2>&1
# exit 0; ok github.com/system-inc/adamic/internal/regexp 3.614s
# Includes the existing full matcher/parser corpus and the new cases and seven mutants.

go test -count=1 -timeout 30m -parallel 4 ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/oct6-counts.log 2>&1
# exit 0; ok github.com/system-inc/adamic/internal/oracle 108.923s
# Four rows added; no pre-existing counts changed.

go test -v -count=1 -timeout 10m -parallel 4 ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/class_(oct6|inheritance)' > /tmp/oct6-inheritance-final.log 2>&1
# exit 0; ok github.com/system-inc/adamic/internal/oracle 17.535s
# Four new plus five existing inheritance fixtures pass all shared oracle checks.

go test -v -count=1 -timeout 10m ./internal/oracle ./internal/regexp -run '^(TestOct6|TestMatcherOct6)' > /tmp/oct6-mutants-final.log 2>&1
# exit 0; oracle 26.304s, regexp 0.198s; all ten mutants caught.

go vet ./... > /tmp/oct6-vet-final.log 2>&1
gofmt -l cmd internal > /tmp/oct6-gofmt-final.log
# Both commands exit 0 with empty logs.
```

The first exploratory inheritance run exited 1 solely because the construction probe was refused; the probe was retained for review and removed from registration. No mismatch was reclassified as a passing ordinary fixture.

The repository-wide full test gate was not run: the merged matcher's prior report records a 1,384.803-second oracle run. This unit instead ran the complete matcher package, the filtered inheritance oracle, every fixture's counts, vet and formatting. Other oracle behaviors, whole-repository integration, native RegExp embedding (the matcher is presently Go), catastrophic-backtracking performance, all Unicode folds and other platforms are outside this added coverage. The inherited matcher corpus and Unicode controls still run in its package tests.

A targeted cohere formatting invocation also exited 0 and left the fixture bytes unchanged:

```sh
cd cohere
go run ./command/cohere --directory /workspace/adamic --format-only /workspace/adamic/internal/oracle/testdata/class_oct6_deep.a /workspace/adamic/internal/oracle/testdata/class_oct6_parameters.a /workspace/adamic/internal/oracle/testdata/class_oct6_release.a /workspace/adamic/internal/oracle/testdata/class_oct6_subclass_holder.a > /tmp/oct6-cohere-format.log 2>&1
# exit 0; types and lint explicitly reported as not checked.
```

The formatter's lack of edits is an observation only; this is not a claim of a full cohere type/lint gate. `git diff --check` passed. All new fixtures were accepted by Adamic's checker and held to source Node through the oracle.
