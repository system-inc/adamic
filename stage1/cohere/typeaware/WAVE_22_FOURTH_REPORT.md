Built: completed the nine previous claims before claiming and porting prefer-numeric-literals, prefer-object-has-own and prefer-object-spread; twelve rules on this branch.
Commits: nine completed through a7e33d0c; fourth claim 712cddd7 was pushed before code; implementation is the commit containing this report.
Commands and outputs: agreement PASS 126.715s; bridge PASS 142.675s; vet/format PASS; native/Go medians repository 0.440992/0.258058s, compiler 3.316988/0.653543s.
Mutants: numeric-prefix, has-own-fix-span and spread-parens caught only by exact diagnostic byte comparison; retained registry and all seven bridge mutants caught.
Not covered: full repository gate, other compiler/configuration profiles, shared CLI registration and JavaScript execution of these FFI rules; no requested rule is blocked.

## Selection and scope

The previous nine claims were complete, tested and pushed through a7e33d0c before this selection. After pushing everything and fetching every origin head, 348 fetched refs, 33 distinct claim documents, 129 mentioned rule names and the 26 base/main ports were checked against the descending combined-volume ranking. Claimed names, including released reservations, were excluded conservatively. The first three available names were prefer-numeric-literals, prefer-object-has-own and prefer-object-spread, each with zero recorded findings on both frozen corpora. Claim 712cddd7 was pushed before implementation. validation-wave-22-fourth/selection.json records the refs, inventories and available ordering. No additional rules were claimed and no PR was opened.

Each rule is in its own .a file. PreferenceSupport and ObjectAssignReferences are new .a helpers in this unit's directory. All judgments and fixes run on the native Adamic tree. The existing global-binding, global-source, resolved-name and symbol-identities questions already supply every compiler fact needed; no new bridge question or shared registration edit is necessary. The shared generator, existing harness files, compiler emitter and lowerer are untouched. A dedicated wave_22_fourth_suite.a, wave_22_fourth_test.go and independent production Go oracle provide this unit's comparison.

Numeric literals use cooked string and canonical numeric tokens. The repair retains Go's explicit uint64 limit, converting two exact 32-bit limbs to a double once and comparing that with the production parseInt's sequential double accumulation. Invalid digits, separators, signs, empty strings, overflow and value changes decline fixes. Callee text, comments and Go's byte-based identifier adjacency determine the final repair.

Has-own matches all three production receivers, including the empty object whose repair introduces Object. That case resolves Object in the checker scope at the literal, so a local shadow declines. Parenthesized callees retain the wrappers outside the replaced member. Comments inside the callee decline repairs.

Object spread follows the production reference tracker through global objects, aliases, object patterns, defaults, assignments, logical expressions, conditional branches and type-only wrappers. A written global declines its own trace; aliases remain flow-insensitive. Its recursion stack is path-local, preserving duplicate discoveries. Every repair is emitted in production order, including argument parentheses, empty/trailing-comma literals, line-comment whitespace, assignment/arrow/conditional spreads and leading semicolons. Only linked syntax-tree nodes participate: the parser's speculative orphan nodes are not Go AST children.

The shared native parser accepts type arguments without consulting JavaScript script kind. Go's tryParseTypeArgumentsInExpression explicitly declines them in JavaScript. The native rules apply that same local call-shape guard using existing global-source metadata. The extracted JavaScript generic-call cases distinguish the paths and compare exactly; the shared parser is unchanged.

## Exact agreement

Cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and TypeScript source at 050880ce59e30b356b686bd3144efe24f875ebc8. The independent oracle invokes unchanged production registry rules and imports no bridge code. Both programs keep configured declaration roots and use the same options and manifest.

| Population | Roots | Findings | Identical bytes, normal and sanitizers |
| --- | ---: | ---: | ---: |
| Extracted and targeted paired controls | 641 | 425 | 169,502 |
| Frozen repository | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,857 |

The controls contain 322 unique production/targeted source texts, each written as JavaScript and Adamic input. A module marker isolates each file's shadows from other fixtures. The independent Go parser accepts 641 of 644 roots; three syntax-invalid roots are recorded with their full source in controls.json and excluded from both executions. Additional controls cover uint64 overflow and rounding, cooked escapes, Unicode/CRLF, optional and computed members, comments, shadowed Object introduced by fixes, alias cycles/destructuring/defaults, computed concatenated/template keys, global writes, accessor/spread declines, generic calls, parentheses and automatic semicolon insertion.

| Rule | Findings | Ordered repairs | Suggestions |
| --- | ---: | ---: | ---: |
| prefer-numeric-literals | 166 | 100 | 0 |
| prefer-object-has-own | 80 | 74 | 0 |
| prefer-object-spread | 179 | 1,241 | 0 |

Full byte streams include file headers, rule/message IDs, exact messages and byte spans, fix/suggestion counts and every ordered repair. The production rules offer no suggestions, and their zero suggestion counts are compared. Normal and ASan/UBSan/LSan executions have empty native stderr. The zero corpus findings are backed by required positive controls for every rule.

## Mutants and ownership

All three rule mutants build and exit 0 with empty stderr. They preserve findings and change only repair metadata, so findings-only comparison would miss them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| numeric-prefix | Binary repair starts with 0o instead of 0b | 13,494 |
| has-own-fix-span | Callee replacement ends one byte later | 17,047 |
| spread-parens | Wrapped replacement loses its closing parenthesis | 292 |

Dedicated global-binding, resolved-name and symbol-identities probes after release exit 70 with exactly `adamic: panic: invalid or released checker handle`. A scratch registry mutant retains the released handle and exits 0, proving the required refusal catches it. Existing global-source release coverage was run in the preceding completed batch; the registry guard is common to every inspect question.

The full bridge package tree passed again: 100 C ABI queries, retained output after release, zero/stale handles, and 1,600 compiler positions across four files with 54,982 identical bytes under sanitizers. All seven existing mutants were caught: input/output lengths, registry retention, wrong source position, omitted link guard, omitted output free and incorrect region allocation. The filtered Node gate passed earlier in this same unit in 63.365s, including the one-byte oracle mutant; no compiler code has changed since. That validates the compiler separately and is not a claim that these FFI rules execute through JavaScript.

## Timing and reproduction

Three alternating native and Go runs on each frozen corpus compare complete output before recording wall time. Program load, native parsing, rule evaluation and serialization are included.

| Corpus | Native median seconds | Go median seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.440992 | 0.258058 | 1.709 |
| Compiler | 3.316988 | 0.653543 | 5.075 |

Native is slower on both workloads. validation-wave-22-fourth/timings.json contains every observation, bridge timing counters and output hash; benchmark.py reproduces the measurement. These are observed times, not speed guarantees.

The original setup for this continuing unit completed successfully: Go 1.27.1, clang 20.1.8 and Node 24.19.0 ready in 0s each; submodules 0s, warm cache 118s and total 118s. nproc is 5, with a four-core quota and 17.6GB memory. The same toolchain remains active through /workspace/adamic-tools/env.sh; setup was not repeated.

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE22_FOURTH_ARTIFACTS=/workspace/wave-22-fourth-complete ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22FourthAgreement$' -count=1 -timeout=30m -v > /workspace/wave-22-fourth-complete.log 2>&1
TMPDIR=/workspace/wave-22-scratch ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-22-fourth-bridge.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /workspace/wave-22-fourth-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_22_fourth_test.go stage1/cohere/typeaware/testdata/oracle_wave_22_fourth.go > /workspace/wave-22-fourth-gofmt.log 2>&1
python3 stage1/cohere/typeaware/validation-wave-22-fourth/benchmark.py /workspace/wave-22-fourth-complete /workspace/adamic /workspace/wave-22-typescript-pinned /workspace/wave-22-fourth-benchmark > /workspace/wave-22-fourth-benchmark.log 2>&1
```

The final agreement run passes in 126.715s. The bridge package passes in 142.675s and checker package in 1.599s. Vet and formatting return 0 with empty logs. Evidence includes all full compressed streams and hashes, actual accepted/rejected control sources, portable frozen manifests, configuration, timing counters and gate logs. The dedicated harness reproduces each rule and released-registry mutant.

The full repository-wide Go gate was not rerun. Other compiler/configuration profiles, shared CLI registration and JavaScript execution of these FFI rules are not covered here. The requested native builds, both corpus comparisons, complete repairs, per-rule mutants and sanitizer/released-handle checks are complete. No shared-harness gap leaves a requested rule blocked.
