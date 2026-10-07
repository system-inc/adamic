Rechecked released DesignSystemForProgram on main f8013f0b; no production cache helper is delivered.
Claim 27799a336 was pushed before the support probe; completed parseFlags is pushed in e9665116f, rules parked in 1f7f2dc4.
validate_blocker.py passes: the native build exits one with TS2305 for missing Mutex and emits no binary; four focused Go cache tests pass in 0.122s.
This prerequisite probe is not a semantic mutant or parity proof; no cache mutant can be claimed without an implementation.
Six Tailwind consumer dependencies remain blocked; reservation is released and work stops at the missing primitive, not at queue exhaustion.

Observed Go contract, from complete reads of cohere/internal/lint/rules/tailwind/design_system.go and rule/program.go:

- Nil program returns a distinct error and never populates the shared cache.
- A package-level sync.Mutex protects both lookup and the complete load.
- The key is program.Identity(), not a path or a per-file object.
- The successful or failed DesignSystemResult is reused unchanged for that identity; changing identity reloads.
- Loading uses Program.DesignSystemFS(), records the read set, and returns the loaded system, table, error and entry point.

Focused Go command, with source /workspace/adamic-tools/env.sh, in cohere:

go test ./internal/lint/rules/tailwind -count=1 -v -run '^(TestDesignSystemIsBuiltOncePerProgram|TestDesignSystemCacheKeyIsTheProgram|TestDesignSystemFailedBuildIsCachedToo|TestDesignSystemIsSafeUnderTheParallelWalk)$' -timeout=10m > /tmp/wave15-cache-go-contract.log 2>&1

All four tests pass. They are upstream contract evidence, not an Adamic cache comparison.

Probe command:

python3 stage1/cohere/lint/helpers/wave15/design_system_cache/validate_blocker.py > /tmp/wave15-cache-validation.log 2>&1

The script runs go run ./cmd/adamic build on gaps/mutex.a, checks nonzero exit and exact TS2305 missing-Mutex diagnostic, and checks that no native artifact exists. This is a recorded feature boundary; when support arrives the script deliberately fails so the blocker must be reassessed. The actual compiler diagnostic is retained in evidence/mutex-build.txt. No shared harness or compiler file is edited.

Search observation: current stage1 .a/.ts sources and internal/load/prelude.d.ts contain no Mutex, ProgramIdentity or RecordingFS public adapter, and no Program or LoadedDesignSystem class/interface matching the Go API. The mutex compile refusal alone blocks this contract. Inference: a single-threaded callback cache could cover a narrower serial API but would leave Go's parallel build-once property unproven. It is not delivered as this helper.

Consumers remaining blocked: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Zero dependency edges are marked ready for this attempt. The completed parseFlags supplies four independent edges; the two prior CSS helpers remain separately validated and pushed.

Ranking correction: slot 05 explicitly withdrew this six-consumer claim. The earlier presence-only scan incorrectly treated that mention as active, selecting the four-consumer parser first. After correction and a full origin refresh, the six-consumer claim was pushed before any support probe. No other active owner was found. The reservation is now released again for a worker with the required runtime and program adapters. Other helpers remain unclaimed; this report does not claim exhaustion.
