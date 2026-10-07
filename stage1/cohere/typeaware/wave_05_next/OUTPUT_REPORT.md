Built: all three continuation claims complete; the two output rules use a wave-owned native CFG and raw bridge facts.
Commits: continuation claim d8b312d9; timer d4320843; output implementations 948dd987, pushed.
Commands and outputs: output verifier PASS; checker tests PASS 0.096s; filtered Node oracle PASS 13.525s; vet and format checks clean.
Mutants: process and blocking verdicts plus CFG edges caught only by Go bytes; three bridge guards and unsafe-name regression caught directly.
Not covered: full repository gate, complete upstream fixture matrix, emitted-JavaScript harness comparison, and nondefault options.

# Completed output ports

The remaining reservations are `nexus/correctness-no-process-exit-after-output` and `nexus/correctness-require-blocking-standard-streams`. Each owns its native judgments in a separate `.a` file. `output_identity.a` shares symbol identity and one-level helper reads; `output_cfg.a` is a wave-owned syntax CFG adapted from the previously validated dead-store walk. The old constructor refusal was bypassed by accepting already constructed bindings. Inheritance and shared bindings/compiler edits were unnecessary. The working small probe is [output_cfg_main.a](gaps/output_cfg_main.a); the earlier refused probe remains historical evidence.

Raw questions `output-callee`, `output-symbol` and `runtime-modules` each have their own Go exporter and `.a` decoder, with one physical registration line apiece. Their protocols are documented beside the native files. Go supplies resolved declarations, ancestry/name kinds and module-resolution edges; native code chooses identities, entrypoints, blocking order, paths and diagnostics. No lint verdict moved to Go. The shared registration generator, shared harness, protected compiler files and original three-rule driver are untouched.

An independent overlay-built Go executable runs the unmodified production rules, with its own compiler host and AST walk and no bridge imports. Comparisons hold canonical ranges, IDs, complete messages, fixes and suggestions byte for byte. Both output rules specify no fixes or suggestions; every compared diagnostic retains those zero fields. The wave-owned verifier is [validate_output.py](validate_output.py).

| Population | Process-exit findings | Blocking-stream findings | Checks |
| --- | ---: | ---: | --- |
| 30 direct controls | 15 | 13 | Normal and ASan/UBSan/LSan |
| 30 separate module programs | 28 | 10 | Normal and ASan/UBSan/LSan |
| Existing 22 DOM timer controls | 0 | 0 | Normal and ASan/UBSan/LSan; 12 timer findings unchanged |
| Existing 22 Node timer controls | 0 | 0 | Normal and ASan/UBSan/LSan; 12 timer findings unchanged |
| Frozen 287-file repository | 0 | 0 | 18,485 identical bytes, normal and ASan/UBSan/LSan |
| TypeScript compiler, 77 files | 0 | 0 | 5,318 identical bytes, normal and ASan/UBSan/LSan |

Controls exercise writes and exits by global and imported symbols, local look-alikes, retained exit codes, loop backedges, catch/finally paths, exits in write arguments and catch bindings, Unicode/CRLF, one-level local/module callees, async/await, overload/generator/never declines, imported and shebang entries, type-only imports, load-time blocking, first/late blocks, callbacks, local const and mutable functions, argument blocking, recursion, computed loads, ambient calls and class initializers. Generated `.ts` module fixtures are TypeScript oracle inputs; every new Adamic implementation is `.a`. The exact StandardStreams suffix is part of the production rule's identity check.

Two real failures were corrected before the final pass. The initial native code used the ordinary TypeScript `Never` enum value; the pinned Go checker uses 262144, and independent bytes caught the extra diagnostic. The compiler corpus then exposed a `Node.Text` panic on a binding-pattern ancestor. The exporter now preserves nontextual name kinds without asking them for text, with a direct regression assertion and an unsafe-name mutant.

## Mutants and lifetime

- Reversing the process rule's write-state condition builds and exits 0 with empty stderr; only independent Go bytes reject its findings.
- Reversing the blocking rule's write-state condition does the same.
- Bypassing a loop's CFG body edge builds and exits normally; only the Go comparison catches the missing loop findings.
- Disabling each new bridge suffix guard fails `TestWave05OutputFacts` with `accepted suffix` for that question.
- Removing the raw-name kind guard fails the destructuring regression with `Unhandled case in Node.Text: *ast.BindingPattern`.
- Each of the three new questions, invoked after program release, exits 70 with exactly `adamic: panic: invalid or released checker handle`. All new native controls and both full corpora pass ASan, UBSan and LSan with empty stderr. The original wave's lifecycle and sanitizer mutants remain recorded in its evidence; those were not rerun wholesale.

## Reproduction

```bash
source /workspace/adamic-tools/env.sh
WAVE05_OUTPUT_SKIP_BENCH=1 python3 stage1/cohere/typeaware/wave_05_next/validate_output.py > /workspace/wave-05-output-verified.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-05-output-checker-tests.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /workspace/wave-05-output-node-oracle.log 2>&1
go vet ./bridge/tsgo/... > /workspace/wave-05-output-vet.log 2>&1
```

The absolute frozen manifests and TypeScript checkout are the original wave's inputs. The independent Go oracle and C archive are rebuilt by the verifier. No test output is piped. Format-only verification uses the already documented scratch `.a` formatting overlay because pinned cohere does not expose `.a` linting. Native compilation checks the `.a` types. The full gate and emitted-JavaScript shared harness remain outside this verification.

## Native versus Go time

Three alternating count-only rounds were measured after every other build/test job finished. Medians include program loading, parsing, all three continuation rules and teardown, on the same zero-finding corpora:

| Population | Go | Native | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.124803s | 0.276087s | 2.21x |
| Compiler | 0.347531s | 2.379856s | 6.85x |

Native is slower. These are aggregate process measurements, not active-finding benchmarks or individual-rule timings. The compiler native run made 32 fact queries; query time was about 13ms, while native run time was about 2.12s. The existing parser/driver dominates this measurement. Raw rounds, source hashes, oracle streams, mutants, release and sanitizer logs are in [output_evidence](output_evidence).

Cloud setup was already successful for this branch: Go 1.27.1 0s, clang 20.1.8 0s, Node 24.19.0 0s, submodules 0s, warm build 82s, total 82s. `nproc` remains 5; cgroup CPU capacity is four cores. No toolchain or pinned submodule changed.
