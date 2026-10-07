Built: Native .a ports of no-throw-literal, no-useless-backreference and prefer-arrow-callback, including its exact fixes and native RegExp tracking.
Commits: earlier six ports 24c600f0 and 4f3a9411; third claim 612007a4 was pushed before implementation; implementation is the commit containing this report.
Commands/results: third gate PASS 144.054s; bridge PASS 121.232s; checker PASS 0.148s; filtered Node PASS 34.274s; vet, formatting and whitespace pass.
Mutants: three rule mutants exit 0 with empty stderr and fail only the Go byte comparison; released-registry mutant, seven bridge mutants and Node's byte mutant are caught.
Not covered: full repository gate, nondefault callback options, complete production project matrix, shared registration/profile integration, emitted-JavaScript rule comparison and speed parity.

The earlier six claimed rules were all tested and pushed before this batch. The all-heads fetch found 348 origin refs, with 33 distinct Markdown claim blobs naming 126 ranked rules. Excluding claims and the implementations on current origin/main and origin/codex/tsgo-c-library left 46 unclaimed ranked rules. These were the first three by combined volume, all zero in both frozen corpora. No selected rule was skipped. A source-literal scan of the 34 distinct origin typeaware trees found no matching implementation. The claim was pushed at 612007a4f0104aa028c1c280325608ccb4707963 before code was written. The snapshot is preserved in validation/selection.json.

Each rule is in its own .a file. The directory also owns a regex structure scanner, reference tracker, constant evaluator, raw reference decoder, suite and test harness. It reuses the prior batch's native AST helper and raw symbol/file-mode questions. The only existing tracked shared-file change is one dispatch line in bridge/tsgo/checker/facts.go. The new reference-symbol question has its own Go and Adamic files and returns raw value-reference identity and declaration locations, including shorthand and local export targets. It returns no lint verdict or repair. Every throw judgment, regex classification, reference traversal, constant evaluation, callback ownership decision and fix is computed in native Adamic. Shared registration generators, shared test harnesses, parser and protected compiler files were not edited.

The independent Go runner calls the pinned production rules through the registry and has its own loader and AST walk; it imports no bridge implementation. Native and Go sort and serialize the complete finding span, rule name, message ID, message, ordered fixes and suggestions. Only defaults are compared: allowNamedFunctions false and allowUnboundThis true. The native callback class implements both flags, but this run does not claim nondefault option coverage. The three rules produce no suggestions. The two other rules produce no fixes; callback fixes are fully compared.

| Population | Roots | Findings | Identical complete bytes | Native sanitizer stderr bytes |
| --- | ---: | ---: | ---: | ---: |
| Controls | 322 | 181 | 70,138 | 0 |
| Frozen repository | 287 | 0 | 18,485 | 0 |
| TypeScript src/compiler | 77 | 0 | 5,010 | 0 |

The 322 controls are deduplicated actual literal sources extracted from the pinned production tests plus explicit controls. Diagnostic snapshot strings are excluded from source extraction. The actual source inputs all remain in the comparison: 321 have no Go parser diagnostics, and the pinned legacy octal-string source `RegExp('\1(a)')` has one. It is retained as a script; the independent oracle explicitly logs its recovered parser diagnostic and invokes the unmodified production rules on the recovered AST. This is a recovered-input comparison, not a claim that the source is valid TypeScript. The control config uses strict true with alwaysStrict false; every other control is an explicit module. Neither frozen corpus has parser diagnostic recovery. Both runners agree on the legacy control's empty findings.

Controls produce 18 throw findings, 106 backreference findings and 57 callback findings. Backreference controls cover all five IDs: forward 41, nested 24, backward 16, disjunctive 14 and intoNegativeLookaround 11. They include numbered/named/duplicate groups, lookaround direction, v-flag nested classes, unknown/constant flags, invalid Unicode patterns, constant bindings, global aliases, shadowed constructors and object destructuring. Callback controls cover resolved self-name versus shadows, arguments ownership, nested arrows/functions, this/super/new.target, generator declines, binds, comments, duplicate and this parameters, async line breaks and parenthesization. Of the 57 reported callbacks, 49 carry 119 individual edits and eight decline repair. Normal and ASan/UBSan/LSan outputs match the production Go bytes for every population.

The native build initially refused an inferred never[]; explicit typed empty arrays fixed that inside the owned helper. The private harness initially misread expected-diagnostic strings as source and forced a legacy script into a strict module. Source extraction and explicit recovery logging corrected those harness errors. No shared parser/compiler change or source omission was used to make the comparison pass. Extracted snippets do not reproduce every production test's helper projects or option configuration, so the complete production project matrix is not claimed.

| Mutant actually run | Catch |
| --- | --- |
| Throw conditional requires both branches to possibly be an Error, instead of either | Exit 0, empty stderr; Go comparison differs at byte 5,578 |
| Backreference message counts the first named group again among the other groups | Exit 0, empty stderr; Go comparison differs at byte 969 |
| Callback repair inserts an extra space after the arrow token | Exit 0, empty stderr; fix-only byte difference at byte 3,802 |
| Keep released reference handle in registry | Mutant exits 0; required panic 70 is absent |
| C ABI input or output length +1 | ASan heap-buffer-overflow |
| Bridge released handle retained | Stale-handle assertion |
| Bridge type read from SourceFile | Independent Go byte mismatch at byte 6 |
| Remove bridge link opt-in guard | Refusal assertion |
| Remove C output free | LeakSanitizer |
| Allocate region entry result on heap | LeakSanitizer |
| Change one Node oracle output byte | Differential oracle mismatch |

The original and sanitized reference-symbol released probes both exit 70 with exactly `adamic: panic: invalid or released checker handle`. The new checker test verifies that an ordinary reference, shorthand read and local export read resolve to one binding and its VariableDeclaration location, and that the question refuses a SourceFile. Existing stream/checker tests also pass. The bridge gate validates 100 C ABI queries, independent output lifetime, stale/zero handles and 162 Go/native positions producing 3,261 identical bytes under sanitizers. The filtered Node oracle runs eight fixtures through Node, the JavaScript backend and sanitized native, plus its byte mutant. The earlier six ports and their mutants are documented in ../WAVE_28_REPORT.md and ../wave28_next/REPORT.md.

Three isolated fresh-process timing rounds ran after the parallel gates finished, alternating subject order. Every timed stdout was byte-compared again. These include loading and serialization and use warm filesystem caches; they are observations on this machine.

| Population | Median native seconds | Median Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.434379 | 0.287574 | 1.51 |
| Compiler | 2.903774 | 0.589507 | 4.93 |

Native is slower. Individual isolated rounds are in validation/timing.json and isolated-timing.log; the gate log retains its earlier rounds, some concurrent with other gate work. Complete canonical outputs and hashes are in validation/comparisons.json. The arrow-fix mutant stream is gzip-compressed to preserve its intentional trailing fix spaces without treating those evidence bytes as source whitespace. Absolute file headers affect output hashes on another machine. Portable frozen manifests and root-content hashes are copied from the original wave. TypeScript is 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3), cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, and typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

The original setup remains in use: Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node v24.19.0 ready 1s, submodules ready 1s, build cache warm 114s, done 114s. nproc is 5, CPU quota 4, memory 17.6 GB. Setup succeeded; see the original wave report and its committed setup log.

Commands from the repository root, writing each test's output to a log:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_THIRD_ARTIFACTS=/workspace/wave-28-third-validation \
  ADAMIC_WAVE28_THIRD_FULL=1 \
  go test ./stage1/cohere/typeaware/wave28_third -count=1 -timeout=25m -v \
  > /tmp/wave-28-third-full.log 2>&1
go test ./bridge/tsgo -count=1 -timeout=15m -v > /tmp/wave-28-third-bridge.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave-28-third-checker3.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /tmp/wave-28-third-node.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_third ./bridge/tsgo/... > /tmp/wave-28-third-vet.log 2>&1
```

The full private gate uses /workspace/wave-28-artifacts/{repository,compiler}.manifest and /workspace/wave-28-corpus/src/compiler/tsconfig.json. Materialize validation's portable manifests there to reproduce the complete run. With no ADAMIC_WAVE28_THIRD_FULL=1, the test runs controls only; absent external corpora are explicitly logged in the full gate. Full go test ./... and full vet were not run. Shared generator/profile/emitted-JavaScript rule integration remains for codex/lint-harness-dot-a. The pinned ordinary cohere CLI rejects .a input, demonstrated in the original wave's lint refusal; this is not a clean production lint claim. These integration gaps did not block native .a compilation, the full diagnostic/fix comparisons or sanitizer checks here.
