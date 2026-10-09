# base/consistency-no-hand-built-declared-error

Ported from the ledger's batch4:base_consistency_no_hand_built_declared_error.ts onto origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. The `.a` module owns its verbatim policy message, descriptor, independent upstream Go adapter, raw witness and mutant. It subscribes only to `NewExpression`, declares `node: true`, and uses the node handed to it without refetching it or checking its kind for relevance. The few argument/property inspections are local; no shared helper or harness file changed.

Pinned Go cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. The upstream rule has no options schema or options type: its Run argument is `any` and unused. The adapter nevertheless decodes and retains JSON on selected option-bearing rows, so it cannot silently return nil for supplied options. No artificial options schema or witness sidecar is introduced.

The descriptor's `upstreamTest: TestNoHandBuiltDeclaredError` captures both real upstream functions: `TestNoHandBuiltDeclaredErrorStaysSilent` (10 cases) and `TestNoHandBuiltDeclaredErrorFires` (4 cases). Coverage includes both exempt filenames, MyBaseError.ts, the second argument, non-object options, namespaced constructors, computed/string/spread keys, explicit identifier properties and shorthand. The owned witness fires three times and includes silent counterexamples.

## Commands and observations

Source the existing `/workspace/adamic-tools/env.sh`; `nproc` is 5. Registry, formatting and vet:

- `go run ./cmd/lint-registry`: PASS.
- `gofmt -w` and `gofmt -l` on oracle.go: the final formatting log is empty.
- `go vet ./stage1/cohere/lint`: PASS; log is empty.

Final uncached commands use `GOFLAGS=-p=1 GOMAXPROCS=2`:

- `ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave109-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=20m`: PASS, 240.777s. TestRulesAgree: 129.74s and 13,057,451 identical bytes over 2,031 upstream source/rule/options combinations plus generated cases. TestCompilerAndStage1Agree: 86.37s, 397 files and 20,436,911 identical bytes. TestOwnedWitnesses: 24.65s and 93,498 identical bytes. Every comparison covers unchanged Go, source Node, emitted JavaScript and ASan/UBSan native. The compiler checkout is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8.
- `ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave109-typescript ADAMIC_LINT_BENCH=1 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestThroughput|TestMutants)$' -count=1 -v -timeout=25m`: PASS, 799.223s. All 41 registered mutants pass their detection checks (760.34s); throughput passes (38.86s).

The owned mutant changes the constructor guard from BaseError to OtherError. It compiles and executes cleanly, but suppresses the three witness findings. Only comparison with Go catches it, independently on source Node, emitted JavaScript and sanitized native.

The first concurrent test attempts exhausted the 16 GiB runner's memory: Go compiler subprocesses were killed, and cgroup memory events recorded OOM kills. They are retained as failed infrastructure evidence. Remaining processes were stopped and all checks were rerun in one test process with one Go package compiler at a time. No oracle, sanitizer or comparison guard was weakened.

## Throughput and limits

Best of five interleaved samples over 77 compiler files and 15,027 matching findings: native 5,492.30 findings/s (2.736011s), source Node 8,345.31 (1.800651s), Go 19,356.06 (0.776346s). This measures the complete registered scanner, not an isolated base-rule rate.

The shared gate explicitly checks refusal of its known malformed parser-recovery inputs. Those are not claimed as supported recovery. This unit did not run the full repository gate. No dynamic RegExp, checker queries, Tailwind inputs or shared harness changes are required by this rule. The separate wave1-09 branch retains four parked rules and their reproducers; this branch contains only the assigned new rule on the area base.

Logs and command manifests are retained under evidence/. The successful final logs are base-declared-error-parity-serial.log and base-declared-error-mutants-serial.log; the parallel attempts are explicitly failed, not certification.
