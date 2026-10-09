# Debugger statement

Implementation commit: 0fe2588fed6d52af313915e2899c0662162b5547.

The unit branch is codex/debugger-statement, started from current origin/main
ce0750f28ef3943057f1f852b3ae5d93e6c5d644. No changes from the separate
method-presence-test branch are included.

Removed the syntax refusal for debugger. Lowering keeps an ir.Debugger statement so
the JavaScript backend can emit debugger;. Native statement emission produces no C,
and flow building adds no instruction, reads, writes or control-flow change.
The JavaScript trace also skips marking this no-op. No native debug flag was added.
The language document now records the October 7 ruling and removes debugger from
the permanent refusal list. The fixture suppresses cohere's taste rule explicitly.

## Observations

The Debug.fail-shaped fixture executes debugger, then throws an Error carrying
Debug Failure. Invalid version. A caller catches the error and prints its message.
Node prints that exact line and exits 0. The normal oracle compares its source on
Node, the JavaScript backend, sanitized native and release native, then checks leaks.

TestDebuggerNativeEmitsNothing loads the fixture under both .a and .ts extensions.
Both forms must be accepted and keep exactly one debugger statement in JavaScript.
The C must contain no builtin trap, raise, breakpoint, DebugBreak or __debugbreak call.
The stronger pin requires byte-identical C for the same source without debugger;.
Native C generation is shared by release and sanitized builds; no debug flag selects
an alternate debugger implementation. The extension witnesses are generated in a
temporary directory; the only new repository source is .a.

TestDebuggerHasNoFlowInstruction holds a debugger followed by a value evaluation to
exactly one flow instruction, so the new statement cannot silently stop a path or
become an analysis instruction.

## Commands and output

All test commands write to logs before those logs are read. Source
/workspace/adamic-tools/env.sh in each shell.

- export GOPROXY='https://proxy.golang.org|direct'; ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
  passed in 29.995s. Cumulative timing lines: Node .023s, Go .026s, submodules .074s,
  markdown .079s, clang .180s, Go build 29.843s, cache 29.969s. nproc=5;
  cgroup cpu.max=400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8.
- node --input-type=module-typescript < internal/oracle/testdata/debugger_fail.a
  prints Debug Failure. Invalid version, exit 0; /tmp/debugger-node.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m ./internal/flow ./internal/oracle -run 'TestDebugger|TestNativeAgreesWithNode/internal/oracle/testdata/debugger_fail'
  passed: flow .010s, oracle .442s; /tmp/debugger-focus.log.
- python3 internal/lower/testdata/debugger/run-mutants.py
  all six caught; /tmp/debugger-mutants.log and /tmp/adamic-debugger-mutants/*.log.
  Raw failure logs are also preserved in this directory under mutants/.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
  passed in 21.727s; /tmp/debugger-counts.log. The new row is
  allocations 2, frees 2, retains 4, releases 4, peak 2, regions 0.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./internal/flow ./internal/oracle
  passed: lower 44.094s, IR 13.079s, native 288.807s, flow 168.835s,
  oracle 279.462s; JavaScript compiled with no standalone test files.
  /tmp/debugger-packages.log and verification.log.
- go vet ./..., gofmt -l cmd internal, git diff --check: clean.

## Mutants

The runner changes one real compiler behavior at a time and restores each source in
finally. It requires the named check to fail and rejects compiler-warning failures.

| Mutant | Catcher | Observed exit |
|---|---|---|
| Emit __builtin_trap for debugger | .a and .ts emitted-C pin identifies the trap | 1 |
| Emit __builtin_trap for debugger | Debug.fail Node oracle reports differing exit codes | 1 |
| Drop JavaScript debugger emission | JavaScript preservation pin reports zero statements | 1 |
| Erase debugger during lowering | JavaScript preservation pin reports zero statements | 1 |
| Restore the syntax refusal | .a and .ts acceptance pin reports refuses debugger | 1 |
| Add a flow instruction | Flow pin reports two instructions instead of one | 1 |

## Coverage limits

No attached-debugger interaction or optional native debug-build flag is implemented
or tested. The fixture isolates the requested debugger-then-Error-throw shape; it
does not compile TypeScript's whole Debug namespace or host-specific stack capture.
The whole-repository test gate was not run; all six touched packages were run,
while setup built and vet checked the rest. No cohere source was copied or changed,
and none of the prohibited compiler files was edited.
