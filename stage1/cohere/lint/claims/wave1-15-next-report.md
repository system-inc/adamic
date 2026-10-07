Built three .a candidates: Tailwind physical direction, Next module variable assignment, TypeScript default parameter ordering; default integration and one JSX case remain blocked.
Commits: claim 375564c; Next 0a06e58; TypeScript e166b17; Tailwind 4bcc710; this report and evidence are in the final continuation commit.
Commands: overlay parity PASS 103.434s (401 captured cases, 214 corpus files); selected mutants PASS 36.20s; throughput PASS 37.106s; registry, vet and filtered oracle passed.
Mutants: module_name_ignored, optional_parameter_ignored, direction_exemption_removed and literal-message interpolation were caught by output comparison on source Node, emitted JavaScript and native.
Not covered: complete JSX parity, default .a registration, full gate, stage1 CLI self-lint; no shared compiler or registration files were edited.

## Selection and claim

Fetched 297 origin refs and audited main production ports and claims on every fetched origin branch. The frozen audit is [wave1-15-next-selection.json](wave1-15-next-selection.json). Complete rule-name mentions in claim documents were conservatively treated as reserved, including skipped dispositions. The original helper-ready list has 46 entries; its first available entry was structure/tailwind-no-physical-direction. The next two available entries in inventory order were @next/next/no-assign-module-variable and @typescript-eslint/default-param-last, from the inventory's `syntax ready for AST/API adaptation` wave. Claim commit 375564c was pushed before implementation. Continued on codex/lint-wave1-15; no PR.

## Implementation and limitations

Each rule has its own directory, rule.json registration metadata, .a implementation/messages, unchanged Go oracle adapter, witness and mutant. Findings, spans, messages and unchanged fixed source compare byte for byte on supported inputs. These three upstream rules propose no fixes. Next visits variable statements and identifies a plain module binding. TypeScript inspects parameter order, optional/default/rest markers and executable bodies. Tailwind preserves Go class syntax, Unicode whitespace, twenty physical families, negative values, direction exemptions, static template pieces and filename gating. Literal message substitution uses split/join to preserve dollar signs and Go's sequential placeholder semantics.

Observed foundation failures: default registry generation asks for rule.ts and fails; default lint package compilation fails at profile_test.go:32 because portFiles is now a function. The reviewable compatibility.patch and driver in the Tailwind directory apply changes only to scratch copies through a Go overlay, extending the proposal published by origin/codex/lint-wave1-12. This proves candidate behavior, not successful default integration.

The original Tailwind JSX fixture `<div className="flex ml-4" />` yields a Go finding, while the shared parser explicitly refuses the source on Node and sanitized native (`expected GreaterThanToken, got Identifier at 22`). Exactly this one pinned case is excluded from the parity suite, and an independent Go TSX probe plus parser-refusal checks preserve evidence of the gap. Inference: this candidate is not a complete Tailwind JSX port.

## Validation

Pinned TypeScript v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8. Pinned cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Toolchain: Go 1.27.1, Node 24.19.0, clang 20.1.8; nproc=5.

`bash cloud/setup.sh` reported Go, clang, Node and submodules ready in 0s each, then failed test-cache warmup at the inherited profile_test.go error; it emitted no final warmup/done timing. Sourced /workspace/adamic-tools/env.sh and used the scratch overlay. Broad dependency warmup later failed on an unrelated klauspost/compress archive redirected to storage.googleapis.com (Forbidden). Downloading only regexp2 and regexp2/v2 succeeded; the driver now uses that targeted command. A prepare-only driver check passed. Logs contain the setup failure and sanitized broad-download error.

All test output was written directly to log files. Final commands used `-overlay=/tmp/lint-wave1-15-reviewed/overlay.json`:

- `go test ./stage1/cohere/lint -count=1 -v -timeout=20m -run '^(TestOwnedWitnesses|TestRulesAgree|TestCompilerAndStage1Agree|TestWave15Shapes|TestWave15JsxGap|TestWave15MessageMutant)$'`: PASS 103.434s. Captured cases: Next 13, TypeScript 117, Tailwind 53, inherited rules 218, total 401. Parity payload 185,532 bytes; corpus 77 compiler plus 137 stage1 files, payload 12,628,931 bytes. Witnesses 8,430 bytes; additional shape probes 16,341 bytes.
- `go test ./stage1/cohere/lint -count=1 -v -timeout=10m -run '^TestMutants$/(module_name_ignored|optional_parameter_ignored|direction_exemption_removed)$'`: PASS 36.20s. Mutants compiled and exited successfully with no sanitizer diagnostics; each differed only in compared output on all three executions. Module and optional mutants miss findings; direction mutant adds an rtl finding.
- The separate literal-message mutant restores JavaScript replacement-string semantics; it also compiles and exits successfully and is caught on all three executions by the ml-$& probe.
- `go test ./stage1/cohere/lint -count=1 -v -timeout=10m -run '^TestWave15Throughput$'`: PASS 37.106s.
- Registry package with overlay: PASS 0.013s. `go vet -overlay=... ./...`: exit 0.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m`: PASS 0.264s; a compiling one-byte sentinel proved the comparator fails on mismatched output.
- `git diff --check`: exit 0. Full gate was not run; default package failures are explicitly recorded.

Source Node, emitted JavaScript and sanitized native compare to unchanged Go oracle output. Native correctness uses sanitizer-enabled execution. Reproduce with the Tailwind README's validate.py command and a checkout of the pinned TypeScript source. The dependency-command-only driver update was checked separately after the final semantic suite.

## Findings per second

Best of five interleaved end-to-end count-only runs on 77 compiler files plus one synthetic positive file per rule; every backend's finding count matched. Timings include startup, reads, parsing, visiting and message construction, but exclude build, formatting and fixes. Timed native uses -O2 without sanitizers; sanitized native is used for correctness above. These are corpus-plus-positive-fixture rates, not isolated visitor performance or compiler-only finding rates.

| Rule | Findings | Native /s | Node /s | Go /s |
| --- | ---: | ---: | ---: | ---: |
| @next/next/no-assign-module-variable | 1000 | 926.70 | 1415.44 | 6103.65 |
| @typescript-eslint/default-param-last | 1024 | 919.86 | 1290.15 | 6089.00 |
| structure/tailwind-no-physical-direction | 2000 | 1752.82 | 2573.20 | 10836.62 |

Compiler-only findings were respectively 0, 24 and 0. No performance guarantee is inferred from these measurements.

Evidence: [parity](../rules/structure-tailwind-no-physical-direction/evidence/parity.log), [mutants](../rules/structure-tailwind-no-physical-direction/evidence/mutants.log), [throughput](../rules/structure-tailwind-no-physical-direction/evidence/throughput.log); the same directory includes setup, registry, vet, oracle and default-failure logs.
