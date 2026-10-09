Pinned both construction refusal boundaries and added two-site blame for source adapter calls, call/apply and bound forwarding.
Construction commit: 0c53029a6. The separate blame commit follows it on compiler/views-v4.
All V4 oracle tests: 13.367 s; focused lower 1.909 s, native 15.926 s, JavaScript 0.603 s; reader guard and lane output are recorded alongside this report.
Read-site omission mutants were caught in native, ASan/UBSan native and JavaScript for direct argument/result, call, apply and bound calls. Exit/output and the actual call site stay unchanged; the required read site disappears.
Checked construction remains deferred. Runtime callback source-site blame is explicitly pending because its IR lacks a source position. Generics and the remaining lane 5 rewrite were outside this unit.

The construct-signature fixture reports NotYet with path and fix. The no-signature fixture is stopped earlier by TS7009 with path and fix; the checker is never bypassed. Node successfully executes both witnesses. Their test leaves took 0.22 s and 0.12 s in the focused run (construct-refusals.log).

The adapter retains the read site where the cached adapter was created. Invocation context carries the actual later source call site; native uses a thread-local borrowed static string, JavaScript uses a synchronous context restored with try/finally. Native restores the context before the caller handles a thrown Adamic exception. Bound forwarding inherits its caller's context. No additional allocation occurs on successful calls; native formats the expanded diagnostic only on the fatal panic path.

Far-escaping fixture leaves in the final V4 run: argument 1.29 s, result 1.19 s, call 1.21 s, apply 1.15 s, bound 1.79 s. Nested compatible adapter invocation followed by an outer result failure passed in 0.84 s, proving restoration of the outer call site. Existing argument/result, call/apply and bind omission mutants also passed in the V4 run. Runtime callback source-site coverage is not claimed; its named test is pending.

Commands ran from the repository root with the configured toolchain. Each Go test used -count=1 -timeout 90s and outer timeout 90; output was written directly to review logs:

- go test ./internal/oracle -run 'TestV4EscapeAdapter(ConstructSignature|WithoutConstructSignature)Refusal$' -v: 0.231 s.
- go test ./internal/load ./internal/lower -run 'View|Construct|Diagnostic': lower 2.024 s; load had no matching tests.
- go test ./internal/ir -run TestCallTargetReaders: construct-guard.log and blame-guard.log.
- go test ./internal/oracle -run TestV4 -v: 13.367 s, all active tests passed, runtime-callback blame pending.
- go test ./internal/lower -run 'View|ClosureSurface|Bound': 1.909 s.
- go test ./internal/native -run 'View|ClosureSurface|Bound': 15.926 s, split view scope stayed below 90 s.
- go test ./internal/javascript -run 'View|ClosureSurface|Bound': 0.603 s.
- Integration lane command: construct-lane.log and blame-lane.log.

Toolchain setup: Go ready 0.024 s, Node 0.026 s, markdown ready 0.082 s, submodules 0.084 s, clang ready 0.205 s, Go build ready 11.685 s, setup total 11.923 s. nproc=5, CPU quota=4. No registered oracle fixture count rows changed; witnesses are review sources or inline temporary test inputs.
