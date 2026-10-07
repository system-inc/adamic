# Type-aware wave 25 continuation

Built native .a ports of collection misuse, discarded Nexus outcomes and discarded pure results.
Claim b9a19759 was pushed first; implementation and evidence are in the accompanying commit on codex/typeaware-wave-25.
Rule gate PASS 74.103s: 60 controls, 57 findings, 364 corpus files; ASan/UBSan/LSan and released-handle checks passed; native timings below.
Three rule mutants are killed only by Go bytes; ancestry-count mutant is killed by the raw-fact assertion; registry and ABI mutants are detailed below.
Not covered: full repository gate, emitted-JavaScript comparison for these ports, CLI suppression/edit application, exhaustive TypeScript programs or successful cohere CLI lint of .a inputs.

## Selection and ownership

Fetch covered all 320 origin refs. Selection used the full VOLUME_REPORT.md ranking:
descending combined compiler and repository volume, lexical rule-name ties.
Ports on origin/codex/tsgo-c-library at 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6
and origin/main at ef3d907ecdc4c771b016f7d9c52372def057a340 were excluded.
Every rule mentioned in the 30 distinct claim blobs on origin branches was also
excluded. The first three remaining rules were:

- nexus/correctness-no-collection-misuse
- nexus/correctness-no-discarded-outcome
- nexus/correctness-no-discarded-pure-result

Each has recorded combined volume zero. The exact ref snapshot, ranking and
exclusion evidence are in validation-wave-25-next/selection.json.
Claim commit b9a19759 was pushed successfully before implementation. No more
rules were claimed following Ahra's correction. The earlier work and its logs
were already pushed in 51eb2a06 and 585a2a19.

Each rule owns a .a file. Nexus symbol decoding, policy messages, the native
runner and tests are new files. No shared registration generator, shared test
harness, protected compiler files or submodule pins changed. The raw checker
question has its own Go and .a files; the pre-correction registration changes
one existing fallback dispatch line in bridge/tsgo/checker/facts.go. Its import
is in the new outcome rule. No subsequent edits to shared files were needed.
The existing harness loads explicit .a roots successfully, so the .ts fallback
was unnecessary.

## Behavior and independent comparison

Collection misuse preserves all four cases: in on array/tuple operands with
non-index missing keys; constant comparisons of built-in length/size; bracket
keys missing every collection member; and default-library Object listing
methods on default-library Map/Set types. Constraints, every-member union
requirements, literal property lookup, canonical array-index boundaries,
comparison mirroring and lookalike exemptions are decided in Adamic.

Discarded outcomes match literal-arm declarations and their top-level union
aliases by filename and alias name. Distinct declaration identities must cover
every original union arm. Optional calls, generic instantiations, extended
unions and awaited producers retain this provenance; narrowed success arms,
floating promises, void and local aliases stay quiet. Findings span the whole
expression statement.

Discarded pure results require every resolved method declaration to be an
allowlisted MethodSignature on a default-library String, Array or ReadonlyArray
interface. Any callback argument, including a callable union or any/unknown,
exempts the call. The call itself is the reported span, omitting enclosing
parentheses and semicolon. All three rules produce no fixes or suggestions.

The new type-declaration-ancestry question returns raw AST ancestry, direct type
links, local identities and union arm counts for a raw or awaited compiler type.
It contains no rule name, Nexus path, allowlist or lint verdict. Exact selector,
versioned UTF-16 field framing and released program validation remain enforced.
Protocol documentation is bridge/tsgo/checker/type_declaration_ancestry.md.

The independent Go oracle loads the unchanged production cohere registry rules
at submodule pin 715ba94f3608a6500086b1076ce5cb7e51b836db and independently walks
Go ASTs. It does not read Adamic decisions or expected findings.

| Population | Files | Findings | Identical bytes | Sanitized equality |
| --- | ---: | ---: | ---: | --- |
| Extracted production controls plus Unicode/CRLF and index/numeric edges | 60 | 57 | 29080 | Passed |
| Frozen repository corpus | 287 | 0 | 18485 | Passed |
| TypeScript v6.0.3 src/compiler | 77 | 0 | 5010 | Passed |

Controls produce 25 collection, 19 outcome and 13 pure-result findings.
Source strings are extracted from pinned production test tables. Four external
Nexus declaration fixtures are written as .a files, with scratch-only .ts
symlinks preserving the production rule's required TypeScript declaration path.
No .ts source file was written or added. The TypeScript corpus is pinned at
050880ce59e30b356b686bd3144efe24f875ebc8. Both frozen manifests are the populations
already used by the branch; new ports are not silently added to that population.
Positive controls are necessary because both corpus runs are finding-free.
Output bytes include spans, rule IDs, messages, fixes and suggestions. All
normal and sanitized native runs have empty stderr.

## Mutants and lifecycle checks

| Mutant | Observation | Catcher |
| --- | --- | --- |
| collection-zero: size comparison <= 0 becomes < 0 | Native build succeeds, exits 0, empty stderr; first difference byte 1462 | Independent Go finding bytes |
| outcome-incomplete: one arm qualifies as complete | Native build succeeds, exits 0, empty stderr; first difference byte 17275 | Independent Go finding bytes |
| pure-callback: callable signature count must exceed 999 | Native build succeeds, exits 0, empty stderr; first difference byte 26239 | Independent Go finding bytes |
| ancestry-count: raw union arm count plus one | Go protocol test fails with lost union arm count | New independent raw metadata assertion |
| released registry retains live entry | Native exits 0 instead of required panic 70 | Released-handle exit/message assertion |
| C input length plus one | ASan heap-buffer-overflow | Bridge sanitizer guard |
| C output string length plus one | ASan heap-buffer-overflow | Bridge sanitizer guard |
| Source-file type substituted for selected node | Oracle mismatch at byte 6 | Bridge's outside Go oracle |
| Link opt-in guard removed | Unlinked checker calls no longer refused | Bridge link refusal assertion |
| Output C free removed | Leaked buffers | LeakSanitizer |
| Region result allocated on heap | Unowned result leaks | LeakSanitizer |
| Native result byte changed | Native disagrees with Node | TestTheOracleCatchesOneByte |

The new ancestry request after tsgoRelease exits 70 with exactly
`adamic: panic: invalid or released checker handle`. The registry mutant
proves that assertion can fail. The foundational bridge gate additionally
holds 100 C ABI checks, outputs surviving release, zero/stale handles, distinct
handles, 162 selected positions and 3261 exact bytes under sanitizers. Its own
released-entry mutant is also caught by the stale-handle assertion.

The first new protocol test selected sync() inside async(), so its exact-range
selector failed. Restricting the selector to the complete line fixes the test;
checker.log records the passing package rerun. bridge-initial.log retains the
initial failure, rather than hiding it. bridge.log is the final full bridge gate.

## Commands, toolchain and timing

The session setup was `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Its recorded timing lines are:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (89s)
setup: done in 89s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc is 5; the quota is 4 cores. Go 1.27.1, clang 20.1.8 and Node v24.19.0.
Tests wrote logs directly, without piping their output.

```
ADAMIC_WAVE25_NEXT_ARTIFACTS=/workspace/wave-25-next-validation \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus \
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware \
  -run '^TestWave25NextAgreementAndMutants$'
go test -v -count=1 -timeout 15m ./bridge/tsgo/...
go test -v -count=1 -timeout 5m ./bridge/tsgo/checker
go vet ./...
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
go test -v -count=1 -timeout 10m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions)\.a$'
```

The rule gate passes in 74.103s, final bridge gate in 64.233s (checker package 0.164s), checker rerun in
0.098s, and the filtered Node
oracle with closures, method closures, generic functions and the byte mutant
passes in 5.654s. Vet and gofmt logs are empty. The ancillary ancestry mutant
uses a Go build overlay; the test fails as expected in 0.015s.

Three interleaved Go/native runs compare complete output, without timing stderr.
Medians include process launch, checker program load, parsing, rule decisions,
output and release:

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.281303055 | 0.799212455 | 2.8544 |
| Repository | 0.306369213 | 0.140447521 | 2.1814 |

Observation: native is slower. Instrumented single runs issue 17539 compiler
queries and 5981 repository queries; query times are 597ms and 71ms respectively.
No inference that optimizing the bridge alone will close the whole-process gap
is claimed. The full repeat samples are in bench.json.

## Evidence and limits

validation-wave-25-next holds test and mutant logs, nine compressed compared
outputs, their uncompressed hashes, manifests, input hashes, setup timings,
selection snapshot and repeated timings. The reproducible test regenerates all
controls and mutations. Scratch binaries and subprocess logs remain at
/workspace/wave-25-next-validation.

The full repository Go test gate and all pre-existing type-aware suites were not
rerun. These three use production defaults and have no options. The filtered
Node oracle holds the compiler/runtime features used here; it does not compare
emitted JavaScript for the three lint modules. That shared support belongs to
codex/lint-harness-dot-a, per Ahra. No shared harness or generator was edited.
The pinned cohere CLI previously rejected .a roots as non-TypeScript, so no
passing cohere CLI lint result for these files is claimed. Generic program
shapes outside these controls and corpora, suppression and edit application
remain untested.
