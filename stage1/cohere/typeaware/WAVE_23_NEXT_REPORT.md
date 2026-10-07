Built: nexus/correctness-no-collection-misuse, nexus/correctness-no-discarded-outcome and nexus/correctness-no-discarded-pure-result, each in its own .a file.
Commits: claim 4107fa50 pushed before implementation; implementation f06724a1; same codex/typeaware-wave-23 branch.
Commands and outputs: new wave PASS 109.926s; original wave regression PASS 95.031s; bridge PASS 76.728s; final checker PASS 0.090s; filtered Node oracle PASS 24.801s; vet and gofmt logs empty.
Mutants: three rule mutants and two new raw-fact mutants caught only by Go finding bytes; released-registry mutant caught by the required refusal; foundation ownership, length, position, region and link mutants caught.
Not covered: full repository test gate, every upstream fixture and compiler-option combination, suppression/edit application, JSX populations, emitted-JavaScript rule comparison, and pinned cohere CLI linting of .a files.

The original three ports and their evidence were already pushed at e67b14f4
before the continuation request. `git push` reported everything up to date.
The fetch explicitly requested every origin head because the configured default
refspec fetches main only. Selection checked 320 origin refs, the 26 production
ports on main and codex/tsgo-c-library, and 90 claimed rules across 30 distinct
claim blobs. It identified these as the first three of 82 remaining entries in
the combined VOLUME_REPORT count tables, descending by count with lexical ties.
All three original counts are zero. Ports were identified from the native suites
and their Go production-rule oracles, including abbreviated native rule names.
The exact fetched snapshot, exclusions and ranking are preserved in
[selection.json](validation-wave23-next/selection.json). No conflicting claim was
found, and no additional rules were claimed after Ahra's correction.

At selection, main was ef3d907ecdc4c771b016f7d9c52372def057a340 and the bridge
branch was 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Implementation continued on
the existing branch, originally based on 0d540f4, as requested. Cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db and typescript-go remains
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No submodule pin changed.

The three native files hold the decisions. The collection rule checks exact
library symbols, constrained union constituents, canonical array indices,
property presence, nonnegative size comparisons in either operand order, and
Object receiver/method origins. It declines subclasses and intersections as Go
does. The pure-result rule checks every declaration's interface and library
origin, the precise method lists, and callable/any/unknown arguments. It reports
the call span rather than the enclosing statement. The outcome rule groups
instantiated literal declarations by their written union, requires all arms,
checks top-level alias identity and the five Nexus file/name pairs, and chooses
the first complete alias in production order. Its awaited path asks for the raw
awaited type. It reports the statement span, including parentheses and semicolon.

The native runner is [wave_23_next_suite.a](wave_23_next_suite.a). It loads one
checker program, parses each root with Adamic, indexes parents, runs these three
rules and sorts full canonical diagnostic lines, retaining duplicates. The Go
oracle independently loads the program and invokes the three unchanged pinned
registry rules with default options. It imports no bridge implementation. Every
finding field, ordered fix and complete suggestion is compared. These rules
produce no repairs or suggestions in production, and all 120 controls have
zero of each, matching Go.

Two new raw questions have separate Go and Adamic files:
[declaration-ancestry and awaited-shape](wave_23_next_facts.md). They supply raw
syntax/locations and checker type graphs. No lint verdict, known Nexus name,
filename filter or message crosses the bridge. Registration changes one line
in this worker's existing type_alias_info.go provider. The shared facts.go
registration, shared registration generator and shared test harness are untouched
by this continuation. New dedicated tests reuse the existing harness API without
editing it. No protected compiler file was changed. The native runner already
loads .a modules, so no temporary .ts implementation or codemod was needed.

| Population | Roots | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| Targeted ESNext controls | 42 | 120 | 56,117 |
| TypeScript compiler | 77 | 0 | 5,318 |
| Frozen repository | 287 | 0 | 18,485 |

Controls comprise 38 executable .a roots and four external TypeScript declaration
fixtures. The latter retain the production rule's required `/nexus/.../Json.ts`
path identity; they are checker inputs modeled on Go cohere's existing Nexus
fixtures, not Adamic implementation files. Their full source is preserved in
[controls-sources.json](validation-wave23-next/controls-sources.json).
They cover all five Nexus aliases, generic instantiation, optional calls, awaited
and floating promises, wrappers, complete/partial/multiple unions, extra arms,
parenthesized types, lookalikes outside Nexus, and aliases inside a namespace.

The 49 collection findings include all four messages: 9 inOnArray, 22 impossible
size comparisons, 13 bracket accesses and 5 Object calls. Controls cover canonical
indices at 2^32 - 2 and 2^32 - 1, literal unions, sparse indexes, real members,
wide keys, any/unknown/symbol keys, constraints, exact library origins, subclasses,
intersections, optional receivers, both operand orders, all comparison operators,
hex/binary/octal/exponent/underscore numbers and unary signs. The 55 pure-result
findings cover every method in both lists with ESNext declarations, while the
negative cases cover function/any/unknown/union arguments, spreads, kept results,
void, mutating methods and project lookalikes. Outcome controls produce 16
findings. Unicode and CRLF exercise byte spans. Empty corpus findings are supported
by these positive controls and mutants, rather than used alone as proof.

| Mutant | Observation and catcher |
| --- | --- |
| Collection: lower canonical array-index upper bound by one | Exit 0, empty stderr; Go bytes differ at 1,851 |
| Outcome: count each distinct arm twice | Exit 0, empty stderr; Go bytes differ at 28,828 |
| Pure result: report the statement instead of the call | Exit 0, empty stderr; Go bytes differ at 17,081 |
| Raw ancestry: add one to direct child counts | Exit 0, empty stderr; Go bytes differ at 28,828 |
| Awaited shape: return the unawaited raw type | Exit 0, empty stderr; Go bytes differ at 31,881 |
| Registry: keep the released program live | Mutant exits 0 with empty stderr; normal exits 70 with exactly `invalid or released checker handle` |

The raw-fact test compares all eight ancestor fields against the actual checker
AST, including alias written-type spans, parentheses and instantiated arm origins.
It compares the awaited opaque identity against direct checker operations and
rejects malformed questions. The final original-wave regression also repeats its
three rule mutants, alias-argument/rest-callback mutants, both corpora normal and
sanitized, and released-handle mutant. All pass through the final provider.

The bridge foundation checks 100 C ABI queries, output survival after release,
zero/stale handles and nonreused handles. Its external corpus compares 1,600
positions across four compiler files and 54,982 bytes under ASan/UBSan/LSan.
Input/output length mutants trigger ASan, omitted C frees and heap-instead-of-region
allocation trigger LeakSanitizer, wrong position differs at byte 6, and stale
registry/link opt-in mutants fail their refusal expectations. The foundation run
preceded the final move of registration into this worker's own provider; final
wave and direct checker tests exercise that final path. ASan instruments native/C
memory, not the external Go heap.

Three quiet alternating rounds, with no builds or tests running, compare complete
finding streams and include load, parse, decisions, rendering and teardown:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.824 s | 0.836 s | 3.38 |
| Repository | 0.322 s | 0.146 s | 2.21 |

Native query counts are 73,411 and 8,560 respectively. The implementation is
slower than Go on this measured machine. Timing excludes builds and sanitizers.
Raw rounds and phase logs are in
[measurements.json](validation-wave23-next/measurements.json). Test-run timing
lines were collected with other validation work active and are not the quiet
comparison above. Compressed full outputs, matching hashes, source hashes,
portable manifests, mutant outputs, logs and measurement script are preserved
in [validation-wave23-next](validation-wave23-next).

All commands sourced `/workspace/adamic-tools/env.sh`; test output went to files,
never pipes. The toolchain reuses the successful original setup: Go 1.27.1 ready
0s; clang 20.1.8 ready 0s; Node 24.19.0 ready 0s; submodules 0s; cache warm 83s;
done 83s. That original log is preserved. The continuation rechecked `nproc`: 5,
and cpu.max: 400000 100000. TypeScript v6.0.3 remains pinned to
050880ce59e30b356b686bd3144efe24f875ebc8. The 77 and 287 roots are the same frozen
populations as the first wave, not expanded to include the new implementation.

```sh
ADAMIC_WAVE23_NEXT_ARTIFACTS=/workspace/wave-23/next-final \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave23NextAgreementAndMutants$' -count=1 -timeout=30m -v > /workspace/wave-23/next-final-test.log 2>&1

# Same corpus variables; original three-rule regression.
ADAMIC_WAVE23_ARTIFACTS=/workspace/wave-23/previous-regression \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave23AgreementAndMutants$' -count=1 -timeout=30m -v > /workspace/wave-23/next-previous-regression.log 2>&1

ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/next-bridge-test.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-23/next-checker-final.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/next-node-oracle.log 2>&1
go vet ./... > /workspace/wave-23/next-vet-final.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/next-gofmt-final.log 2>&1

python3 stage1/cohere/typeaware/validation-wave23-next/measure.py /workspace/wave-23/next-final/native /workspace/wave-23/next-final/wave23-next-oracle /workspace/wave-23/typescript /workspace/wave-23/next-bench /workspace/adamic > /workspace/wave-23/next-bench.log 2>&1
```

The filtered Node oracle selected six fixtures and the one-byte oracle mutant,
comparing native/Node/emitted-JavaScript behavior with sanitizers. It is compiler
regression evidence, not emitted-JavaScript execution of these checker-linked
rules. That shared lint-harness work belongs to codex/lint-harness-dot-a.

Initial builds refused an unsupported optional call and two inferred never-array
expressions. Those were replaced with explicit guards/invariants and a slice.
The first successful native control run exposed a namespace mismatch because the
shared Diagnostic constructor defaults to @typescript-eslint/. Each rule now
uses the established explicit-namespace constructor pattern. Those initial
failures are preserved and are not counted as mutant kills. Subsequent ES2022
and expanded ESNext gates pass, followed by the final owned-provider gate.

The pinned cohere CLI still refuses these .a paths as unsupported and reports
none in its program. Its dry-run refusal is preserved and is not counted as a
passing lint gate. No shared extension-support code was changed. The type-aware
loader and native build support these exact .a sources and were checked under
sanitizers. The full repository `go test ./...` gate was not run; the explicit
package/filtered gates above are the evidence. No PR was opened. Work stops after
these already claimed rules; no further claims are made.
