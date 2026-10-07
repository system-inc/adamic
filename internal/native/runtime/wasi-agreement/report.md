Built: three runtime fixes and the approved checked-in TestWASI assertions make all 35 fixtures agree with Node.
Commits: runtime 7f6d7d2, approved test patch d2946c8, main merge 24f45a7 (origin/main e8ba3d5).
Results: TestWASI 35/35 without overlay; five fixtures natively and as Wasm pass uncached; native package passes.
Mutants: old sort panic, old frame check and old stream open are independently caught again after merging main.
Not covered: full repository gate, full WASI oracle, other engines, Workers and regular-file aliases without preopens.

The files changed are adamic.h, sort.c and input.c. The initial runtime implementation edits are under
internal/native/runtime; the approved landing follow-up also changes
internal/native/wasm_test.go. The independent verification tool and compressed logs
are in this directory; observations.json preserves exact stdout, stderr and exit
status for Node, the fixed Wasm artifact and each isolated mutant.

The base branch did not actually contain the runtime commit, so merge 6844656
combines origin/codex/wasm32-driver at 446b300 with
origin/codex/wasm32-runtime at 52959fc before the fixes.

The original runtime-only verification below is historical. Its overlay-only
TestWASI status and test ownership restriction were superseded by the approved
landing follow-up at the end of this report.

Observed behavior before the fixes:

| Fixture | Source Node | Wasm under node:wasi | Cause |
| --- | --- | --- | --- |
| stack_forever.a | stdout `start\n`, stderr `adamic: panic: RangeError: Maximum call stack size exceeded\n`, exit 70 | empty stdout, engine RangeError stack trace on stderr, exit 1 | Pure Wasm locals required no linear frame. The check saw a fixed shadow-stack position until the engine stack trapped. |
| stack_tail_call.a | stdout `0\n`, same panic stderr, exit 70 | empty stdout, engine RangeError stack trace on stderr, exit 1 | Same as stack_forever; sibling-call prevention was already enabled. |
| write_stdout_order.a | stdout `first\nsecond\nthird\n`, empty stderr, exit 0 | stdout `first\nthird, after cannot write /dev/stdout: no such directory\n`, empty stderr, exit 0 | Reopening the host pipe through /dev/stdout failed, not a missing flush. |
| write_stderr_order.a | stdout `first\nthird\n`, stderr `second\n`, exit 0 | stdout `first\nthird\n`, empty stderr, exit 0 | Same path-resolution failure, with its Error result discarded by the fixture. |
| closures_throw.a | Complete output below, empty stderr, exit 0 | Only the prefix below, then stderr `adamic: panic: wasm32: throwing sort comparators are not supported\n`, exit 70 | WASI's explicit unsupported branch replaced the native longjmp out of sort. Other callback throws already propagated correctly. |

For both write fixtures with stdout and stderr on one pipe, source Node and the
fixed Wasm print `first\nsecond\nthird\n`, exit 0. The independent tool checks
these combined observations in addition to separate descriptors.

Complete source Node stdout for closures_throw.a (also the fixed Wasm stdout):

```
map threw: map stopped at 2
filter threw: filter stopped at 3
find threw: find stopped at b2
some threw: some stopped at d4
forEach threw: forEach stopped after a1+b2+c3
reduce threw: reduce stopped with start0a1b2
Array.from threw: from stopped at 3
Map forEach threw: Map forEach stopped at bob2 with age 42 after ann1=age 31
Set forEach threw: Set forEach stopped at cy3
3 3
sort threw: sort stopped after 6 comparisons
q9,d4,m1,a7,z2,k5
sort by name threw: the comparator met a7
q9,d4,m1,a7,z2,k5
sort by name, in a function threw: the comparator met a7
q9,d4,m1,a7,z2,k5
sort, mid-merge threw: sort stopped in a merge
n0,n37,n74,n11,n48
sort with undefined threw: the comparator met 1
3,,1,2
mapped in place stopped
x1,y2,z3
remapping w0
remapping w1
remapped: w0w0,w1w1
remapping w0
remapping w1
remapping w2
remapped threw: remapped stopped at w2
call threw: called1
through a function threw: called1
the closure's finally ran
finally inside threw: called1
```

Original Wasm stdout for closures_throw.a:

```
map threw: map stopped at 2
filter threw: filter stopped at 3
find threw: find stopped at b2
some threw: some stopped at d4
forEach threw: forEach stopped after a1+b2+c3
reduce threw: reduce stopped with start0a1b2
Array.from threw: from stopped at 3
Map forEach threw: Map forEach stopped at bob2 with age 42 after ann1=age 31
Set forEach threw: Set forEach stopped at cy3
3 3
```

Changes and limits:

- The WASI expansion of ADAMIC_CHECK_STACK reserves a volatile 64-byte linear
  frame, including functions with only Wasm locals. Its endpoints survive
  optimization. The existing 16 KiB margin and panic path then detect recursion,
  flush the earlier stdout and exit 70 before the engine trap. Depth itself is
  intentionally different from Node, as the existing language contract allows.
  This guard is validated with the driver's 128 KiB linker stack; a substantially
  larger linker stack could again let the engine stack fail first. Keep the
  128 KiB reservation and -fno-optimize-sibling-calls. For arbitrary linker stack
  sizes the compiler must instead emit a balanced depth counter at every emitted
  function entry and every normal/exceptional return; overflow must call
  adamic_stack_overflow before entering another recursive call. No emission edit
  or new runtime hook was needed for the current driver.
- WASI TimSort tests the pending exception after each comparison and each
  potentially throwing sort helper before consuming its result. A throw returns
  false from sorted_or_stopped; the existing array-sort snapshot cleanup drops
  held references and never writes the partial sort back. Native still uses its
  existing setjmp/longjmp path. This is ordinary Preview 1 C, with no experimental
  Wasm exception-handling flag or host trap translation.
- writeTextFile keeps its existing flush before file writes. On WASI only, exact
  /dev/stdout and /dev/stderr stream aliases borrow descriptors 1 and 2 when
  fstat identifies a pipe, character device or stream socket. Node's WASI host
  identifies captured pipes as stream sockets. Borrowed descriptors are never
  closed; text still uses the existing WTF-8 to UTF-8 conversion and write loop.
  Regular files retain the original open/truncate/close path. This does not add a
  general /dev filesystem or emulate reopening regular files outside preopens.

No new exception is needed for these five fixtures under the tested Node host.
The three old exceptions in docs/wasm.md are resolved for this driver configuration.
That document and internal/native/wasm_test.go are outside the runtime-only boundary
and were left unchanged. The compiler/integration owner must remove the two special
branches in TestWASI that require old stack traps and unsupported sort panics, let
all fixtures reach the normal byte/exit equality comparison, replace the exception
counters with a single passed counter, and add both write_order fixtures. Update
its documentation from 30/33 plus three exceptions to 35/35 for that test list.
A temporary Go overlay makes exactly these assertion/list changes for verification;
it is not a committed edit to compiler-owned tests.

Each fix was independently reverted in a scratch runtime, compiled successfully,
and caught only on execution against source Node:

| Mutant | Witness and failure |
| --- | --- |
| Restore original sort.c unsupported branch | closures_throw.a: prefix-only stdout, panic stderr, exit 70 rather than 0. |
| Restore original adamic.h frame-address check | Both stack fixtures: empty stdout, engine RangeError, exit 1 rather than 70. |
| Restore original input.c open path | stdout fixture loses second and prints Error message; stderr fixture loses second from stderr. |

The mutations never changed the working-tree sources. Run the reproducible tool
with the repository's Go/Node environment and WASI_SYSROOT set:

```sh
python3 internal/native/runtime/wasi-agreement/check.py > /tmp/wasi-isolated-check.log 2>&1
```

Native object comparison: input.c is byte-identical with clang 20 -O2 before/after.
sort.c object bytes differ after introducing comparison temporaries, so byte-identical
native artifacts are not claimed. Native semantic agreement is checked by the
ordinary oracle, including release, ASan/UBSan, leaks and recorded counts.

Commands (all test output goes directly to logs):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' -v -count=1 -timeout 15m > /tmp/wasi-five-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' -v -count=1 -timeout 15m > /tmp/wasi-native-five.log 2>&1
go test ./internal/native/... -count=1 -timeout 15m > /tmp/wasi-native-package.log 2>&1
PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASI$' -count=1 -v -timeout 15m > /tmp/wasi-integration.log 2>&1
PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 go test -overlay /tmp/wasi-test-overlay.json ./internal/native -run '^TestWASI$' -count=1 -v -timeout 15m > /tmp/wasi-integration-overlay.log 2>&1
go vet ./internal/native/... > /tmp/wasi-vet.log 2>&1
gofmt -l cmd internal > /tmp/wasi-format.log
```

The production five-fixture WASI oracle passes in 8.718 s. The uncached ordinary
native oracle passes all five in 18.237 s. The isolated tool passes five separate
stream comparisons, two combined-stream comparisons and three independent mutants
with five witnesses. Checked-in TestWASI fails only its three obsolete exception
assertions; all remaining 30 comparisons and its request benchmark pass, in 33.783 s.
Its diagnostics print identical Node/Wasm bytes and exits for all three failures.

Setup was run as required with bash cloud/setup.sh, then additionally --wasi-sdk.
The first setup completed with go ready 0s, clang/node/submodules ready 1s, build
cache warm 335s, done 335s. The two setup runs overlapped their shared
/tmp/adamic-gate/setup-warm.log: the SDK setup reported native [build failed], but
the log was interleaved/overwritten and retained no underlying diagnostic.
Tool installation succeeded (SDK ready 9s). A sequential clean setup rerun and
its timings are recorded in the final results below. nproc = 5, cgroup quota four
CPUs, reported memory 17.6 GB. Env file: /workspace/adamic-tools/env.sh.

Final verification:

- Native package: PASS, 239.725 s.
- TestWASI temporary overlay: PASS, 35/35 equivalent, 163.125 s. All 48 runtime
  units compiled under strict C11. The linear-stack canary passed. The request
  benchmark checked 100,000 replies, zero live values between requests, flat
  memory at 393,216 bytes, and 6,300,000 region objects ended.
- Native vet, repository formatting and whitespace checks: no output.
- Sequential clean setup: PASS. Go, native clang, Node, SDK and submodules ready
  at 0 s; build cache warm at 168 s; done in 168 s on 5 processors.
- Final explicit fetch and merge of origin/codex/wasm32-driver: its remote tip
  remained 446b3009174192dbf2a478c203352fb1f0e09be9. Merge reported Already up to
  date. No newer compiler tip was published at this check, so no newer compiler
  commit could be integrated. The runtime branch is pushed; no pull request opened.

[test-expectations.patch](test-expectations.patch) is the precise proposed test
change verified by the overlay. It is evidence for the compiler/integration owner,
not a modification to internal/native/wasm_test.go. The original TestWASI failure
log remains beside the overlay success log so these results cannot be confused.

The full repository gate, full WASI oracle, other Node/Wasm engines, Wasm sanitizers,
Wasm counted lifetime checks specifically for throwing sort callbacks, and deployed
Workers are not covered. No native byte-identical claim is made for sort.c.

Approved landing follow-up, October 7, 2026:

The compiler owner approved applying test-expectations.patch to TestWASI on this
branch. Commit d2946c8 applies that patch exactly. Commit 24f45a7 merges current
origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965 into codex/wasi-agreement,
without conflicts. A remote check after testing still reported that main tip.
No main or area/ branch is pushed or merged into; only codex/wasi-agreement is
pushed for the user's area/runtime integration.

All requested gates now run against the merged tree with the real checked-in
TestWASI assertions. There is no Go overlay or alternate runtime in these gates:

```sh
bash cloud/setup.sh --wasi-sdk > /tmp/wasi-landing-setup.log 2>&1
source /workspace/adamic-tools/env.sh
PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASI$' -v -count=1 -timeout 15m > /tmp/wasi-landing-integration.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' -v -count=1 -timeout 15m > /tmp/wasi-landing-wasm-five.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' -v -count=1 -timeout 15m > /tmp/wasi-landing-native-five.log 2>&1
go test ./internal/native/... -count=1 -timeout 15m > /tmp/wasi-landing-native-package.log 2>&1
python3 internal/native/runtime/wasi-agreement/check.py > /tmp/wasi-landing-mutants.log 2>&1
go vet ./internal/native/... > /tmp/wasi-landing-vet.log 2>&1
gofmt -l cmd internal > /tmp/wasi-landing-format.log
git diff --check > /tmp/wasi-landing-diff.log
```

Observed results:

| Gate | Result |
| --- | --- |
| TestWASI, no overlay | PASS, 35/35 equivalent, 124.646 s. |
| Five named fixtures as Wasm | PASS, 5/5, uncached, 14.354 s. |
| Five named fixtures in the ordinary native oracle | PASS, 5/5, uncached, 10.036 s; includes release, sanitizers, leaks and recorded counts. |
| Native package | PASS, 200.761 s. |
| Independent runtime mutants | PASS: all three reverted mechanisms caught again; five separate-stream and two combined-stream comparisons also pass. |
| Native vet, formatting, whitespace | PASS, empty logs. |

TestWASI strictly compiles all 48 runtime translation units, passes the linear
stack canary, and checks 100,000 request responses with zero live values between
requests, memory fixed at 393,216 bytes and 6,300,000 region objects ended.
landing-observations.json records exact observations for the merged compiler.
The wasi-landing-*.log.gz files contain this follow-up's complete gate logs.

Setup timings: Go, native clang, Node and WASI SDK ready at 0 s; submodules ready
at 1 s; build cache warm at 181 s; done in 181 s. nproc = 5, cgroup quota four CPUs,
reported memory 17.6 GB. Setup and all requested gates completed successfully.
The full repository gate and full WASI oracle were not run in this follow-up.
