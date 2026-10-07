Built two owned .a rule candidates; Google Font Display remains blocked by shared JSX parsing.
Claim commit: 4612124a; candidate implementation commits follow this report's claim reservation.
Four-way parity passed 377 supported captured cases, 398 corpus rows and 8 option variants.
Both compiling semantic mutants were caught by Node, emitted JavaScript and sanitized native byte comparisons.
Excluded two JSX cases, the third port, default registry integration, invalid-config diagnostics and the full gate.

# Scope and observations

Branch codex/lint-wave1-07 retains its existing foundation; the newer fetched main was used for selection,
not merged into this checkout. The three names were reserved and pushed before implementation.
All changed implementation files are in the two owned rule directories; no shared compiler, parser,
registry, oracle or corpus-list files were edited. The original reservations remain in place.

The default registry fails because it opens rule.ts. The proposed compatibility.patch is kept unapplied
inside the Tailwind directory. validate.py copies shared Go harness files to scratch and uses Go overlays;
this makes the candidates reviewable without representing them as integrated ports.

The parser treats JSX as a type assertion and refuses a valid Google Fonts link at GreaterThanToken.
An independent Go oracle reports googleFontDisplayMissing on that exact fixture, while source Node,
emitted JavaScript and sanitized native each exit 70 with empty stdout and the parser diagnostic.
All 16 upstream Google Font Display tests pass in Go alone. There is no third Adamic implementation,
semantic mutant or throughput measurement. A parser owner must supply JSX support first.

# Reproduction and evidence

Toolchain setup: bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh.
Setup timing: Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 18s, total 18s; nproc 5.
Go 1.27.1, clang 20.1.8, Node 24.19.0. TypeScript v6.0.3 commit
050880ce59e30b356b686bd3144efe24f875ebc8 was cloned to /tmp/wave07-typescript.

Runner: stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py.
Each invocation redirects stdout/stderr to a log; the runner writes test output directly to its scratch validation.log.

```
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py --scratch /tmp/wave07-final2 --run '^(TestWave07Supported|TestWave07Corpus|TestWave07Options|TestWave07Throughput)$'
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py --scratch /tmp/wave07-mutants-final --run '^TestMutants$/(direction-aware_exemption_removed|empty_reason_accepted)$'
go test -overlay=/tmp/wave07-final/overlay.json ./stage1/cohere/lint/registry -count=1 -v
go vet -overlay=/tmp/wave07-final2/overlay.json ./...
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=5m
```

Final parity suite PASS 94.667s. Captured 379 cases; explicitly excluded exactly two JSX cases.
Supported new-rule cases: Tailwind 52, descriptions 108; 217 existing-rule cases also compared.
All four paths matched 116410 bytes. Compiler plus checkout stage1 corpus: 398 file/rule rows,
24481871 identical bytes. Eight valid option variants matched 8602 bytes.
Findings, ranges, messages, diagnostics and unchanged fixed source are compared; these rules have no fixes.
Witness and JSX-gap run PASS 116.772s before the final empty-additional-directive option correction;
final parity, options and mutants were rerun after that correction. Witness bytes: 15760.
Registry tests PASS 0.027s including metadata rejection controls; vet exit 0;
filtered external one-byte oracle PASS 3.180s. No full repository gate was run.

# Mutants

`direction-aware exemption removed` changes the rtl/ltr exemption to false. It compiles and runs;
extra RTL findings differ from Go on Node, emitted JavaScript and sanitized native.
`empty reason accepted` accepts a whitespace-only description. It compiles and runs;
the missing finding differs from Go on all three paths. Both controls PASS, total 28.024s.
Compilation errors, parser refusals and sanitizer errors are not credited as mutant detection.

# Throughput

Synthetic 1000 declarations, 2000 findings per rule; best of five complete subprocess executions,
including process startup. Native throughput uses release mode; parity uses sanitized native.
These are synthetic measurements, not compiler-only throughput.

| Rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| Tailwind physical direction | 207760.46 | 18529.65 | 181861.51 |
| Directive descriptions | 177452.20 | 18798.44 | 179187.53 |

Raw logs: ../rules/structure-tailwind-no-physical-direction/evidence/*.txt.
The default-registry-blocker log records the unresolved .a discovery failure.
The report distinguishes successful bounded comparisons from missing full JSX and integration coverage.
