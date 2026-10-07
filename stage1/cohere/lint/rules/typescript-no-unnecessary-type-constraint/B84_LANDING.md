Rules: clean rebase of the seven retained ports onto current area; no owned implementation change.
Commits: previous rule tip b17b7952dfa15528423284b580f776ae16a07948; pre-evidence tip 703e3287458d8d7b77c3b85823d891001a9dc4b8; helper evidence pushed at eab7ae08b.
Commands: fresh supported parity PASS 630.081s, six mutants PASS 483.306s, vet zero, external oracle PASS 9.537s; exact commands and logs below.
Mutants: six rule and four helper semantic mutants freshly caught only by Go output comparison; const annotation mutant remains blocked.
Limits: full witness gate still red on two-edit const fixes, malformed computed-key recovery still refused; no new helper claim, broader seventeen correctness checks or throughput benchmark.

Fetched every origin head and rebased onto area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, containing current main c7991b900362796aefd111474e65eb5398e91953. Both remote bases were checked again before recording evidence and are unchanged. Accepted inherited proven-relation/lowering and native record changes without editing them. The full previously read dedup ledger is byte-identical, SHA256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f. The seven retained and eleven retired copies remain those listed in AREA_LANDING.md; no new batch assignment is inferred. No shared context, finding, driver, registry, oracle or comparator is edited.

Fresh commands, from this worktree with source /workspace/adamic-tools/env.sh, redirect output directly to the named logs:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-b84-supported.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-b84-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave08-b84-witness.log 2>&1
go vet ./... > /tmp/wave08-b84-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-b84-oracle.log 2>&1
```

The owned overlay adds unified_test.go.txt as a virtual test file and replaces no shared source. It calls the real shared comparison, actual Go oracle and sanitized native build. Compiler corpus input is present and checked against pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. No selected owned correctness test skips for missing input.

Upstream PASS 273.49s: 414 supported cases across six rules, all four outputs identical at 149648 bytes. Counts are computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21 and core default-param-last 128. All JSX cases are included; the two malformed computed-key sources are explicitly excluded and remain blocking cases.

Corpus PASS 310.79s: all 364 compiler/stage1 .ts/.a files and 2184 selected rule cases produce identical Go/source-Node/emitted-JavaScript/ASan-UBSan-native output at 80203185 bytes. Three const single-edit/negative controls PASS 45.72s, identical at 1011 bytes. Selected gate PASS 630.081s. These observations certify the stated subset, not full landing readiness.

Six rule mutants PASS 483.306s. Each compiles and executes successfully on all three port backends and is killed only by different Go output: core_default_suppressed suppresses the parameter diagnostic; computed_replacement_wrong corrupts the fix; foreign_read_suppressed suppresses foreign propTypes; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts its suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. const_append_wrong requires the annotation's two automatic edits and is still blocked by a baseline refusal, which is not counted as a mutant kill.

Whole-repository vet exits zero with empty log. The uncached external one-byte oracle PASS 9.537s, zero cache hits. The full TestOwnedWitnesses gate FAIL 227.800s: actual Go exits 2 with panic: unexpected fix shape at its automatic-edit guard. Direct Node probes also reproduce panic: a finding carries at most one automatic edit for let value: 'hello' = 'hello'; and parser slice expected CloseBracketToken, got CloseBraceToken at 8 for ({ ['x' });. Both exit 70. Their source probes and manifests are /tmp/wave08-b84-{const,computed}.ts.txt and /tmp/wave08-b84-blockers.manifest; raw output is recorded beside the gate logs. The parser gap also prevents the shared-harness-only parking exception. Stop before a new claim.

The helper branch rebased cleanly onto current main and is pushed at eab7ae08b. Its fresh uncached four-helper baseline and four-mutant Go/Node/emitted-JavaScript/sanitized-native comparison PASS 108.340s, vet exits zero, external one-byte oracle PASS 45.093s with no cache hits. Helper mutants are nonzero NewTheme deadKeys, the wrong ClearNamespace hyphen boundary, omitted var( refusal and reversed unresolved breakpoint comparison. Each is caught only by output comparison. Their raw logs and exact commands are in helpers/from_wave08/LANDING.md and evidence/landing_c799_*.log on that branch. The four helpers still serve the same six Tailwind consuming rules listed in its reports; none alone removes all final blockers. No new helper implementation or claim.

bash cloud/setup.sh PASS: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm 302s, total 302s. nproc is 5, CPU quota four. Tools remain Go 1.27.1, clang 20.1.8 and Node 24.19.0. No full repository gate, broader seventeen external-input correctness checks or fresh throughput measurement. No check is relaxed or deleted to obtain green. Push only owned branches, using exact expected-old-head leases; never main or area.
