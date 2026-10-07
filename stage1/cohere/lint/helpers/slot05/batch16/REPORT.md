Built updateExpression, parameter and firstThrowableFork as separate .a files; each serves the four rules in CONSUMERS.md.
SHAs: published claim 5fa27f1d before source; landing base 6814e163; implementation/publication SHA reported in final response.
Commands and outputs: four-way helper suite PASS 21.986s; scoped vet and uncached six-fixture input oracle PASS 0.917s; setup 40s, nproc 5.
Mutants: all fifteen compiling semantic variants caught; every mutation and Go divergence in evidence/mutants.json.
Not covered: full rule diagnostics, arbitrary external dependency implementations, full repository test gate and invalid parser/state representations.

# Ownership and readiness

All 41 prior helpers were complete and pushed on current main 39638d9e and area d65a8f93. The previous unit reran all fifteen existing packages uncached and caught all 130 prior semantic variants after the area runtime changed. Those inputs remain unchanged during this unit. All twenty helper branches were fetched and their claims inspected before reserving these three tied maximum unclaimed concrete symbols. The higher-count comments leaves remain owned by the shared comments bundle. The claim was pushed before any batch16 source was written.

Each helper removes one prerequisite from array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Twelve prerequisite occurrences removed, no final blocker removed. Cumulative owned inventory: 44 helpers, 271 prerequisite occurrences, 70 unique consumer rules, 50 helper-ready rules under the frozen inventory assumptions. These are dependency calculations, not completed rule findings.

# Observations and validation

Go cohere at the pinned commit supplied callback/state observations for actual methods through a temporary overlay; no shared harness, rule registry, production cohere or compiler file was edited. Read README.md for the exact dependency and test boundaries. Both native and emitted JavaScript are built by Adamic. Native uses sanitizers, must finish successfully and produce no stderr. Semantic mutants must compile and run normally before a wrong output counts; type errors, panics and sanitizer faults do not count.

The final helper suite ran uncached with 2,644 rows and 5,204 helper calls: 13 update operands, 71 parameters and 2,560 twice-invoked fork states. All consumer files were inspected: array-callback-return 166 literal occurrences, consistent-return 272, no-unreachable-loop 71, rules-of-hooks 555; 801 distinct strings after three controls. Consumer parser values carry stable arena IDs so nil, identity and different fields are distinguished. Fork controls cover depths zero through four, repeated calls, flags, absent/nested handlers and both reachability states.

Mutants: update skips reads, skips writes and reverses them; parameter uses the name instead of type, swaps name/initializer and reverses expr/bind; fork rejects index zero, omits handler linking, sets thrownAny after linking, ignores unreachable input, ignores already-forked state, omits each flag, omits continuation linking and omits entering. All fifteen were caught by Go byte comparisons; exact first divergences are archived. No production mutation is retained.

The initial run failed because dependency execution introduced nested instrumentation into the oracle's outer callback trace. Wrappers were corrected to preserve only direct observations while still executing the real dependency. A subsequent parameter mutation using if(false) failed TypeScript narrowing, so it was explicitly uncredited and replaced by a compiling field mutation. Both superseded logs are archived; only helpers.log is the final successful run.

# Commands

```
bash cloud/setup.sh > /tmp/lint05-batch16-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch16 -count=1 -v -timeout=20m > /tmp/lint05-batch16-verified.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch16 > /tmp/lint05-batch16-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch16-oracle.log 2>&1
go vet ./... > /tmp/lint05-batch16-final-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch16 > /tmp/lint05-batch16-format.log 2>&1
```

Setup timing lines: Go, clang, Node and submodules ready 0s; cache warm 40s; done 40s on five processors, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Input oracle: zero hits, six misses. Test output went directly to logs. No fourth helper is claimed.

Final publication checks: repository-wide vet and formatting both exit zero with empty logs. All twenty helper branch claims were scanned again; no other branch claims any of this trio. Main and area remain 39638d9e and d65a8f93, both ancestors of this branch. No earlier helper, compiler, runtime, dependency or oracle source changed during this unit, so the immediately preceding full retained gate remains applicable. Only the owned branch is published.
