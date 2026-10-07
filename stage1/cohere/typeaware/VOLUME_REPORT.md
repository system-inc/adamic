# Ten more type-aware rules

Native Adamic now runs sixteen production cohere rules with one checker program
and one shared Adamic parser walk per file. The ten additions produce 12,469
compiler findings and 44 repository findings. Including the previous six, Go and
native agree on 14,232 compiler findings and 46 repository findings, byte for byte
including every fix and suggestion. ASan, UBSan and LeakSanitizer runs agree too.

## Selection and scope

Selection started by counting all 62 checker-dependent `@typescript-eslint`
rules with Go cohere, before implementing the ports. The two populations were
TypeScript's 77 `src/compiler` roots and a frozen pre-port repository manifest of
287 files (212 `.a`, 75 `.ts`), excluding the intentional-invalid directories from
CohereSettings. Counts use every rule's defaults, independently of configured
severity. Declaration roots from each config are preserved; `.a` files are read
as TypeScript. The root prelude matters: an initial counter without it was
incorrect and its counts were discarded.

These are the ten highest combined counts among the remaining TypeScript rule
family. They are not the ten highest among every cohere rule family:

| New rule (`@typescript-eslint/`) | Compiler | Repository | Total |
| --- | ---: | ---: | ---: |
| no-unsafe-type-assertion | 3867 | 3 | 3870 |
| no-unsafe-member-access | 2777 | 0 | 2777 |
| prefer-nullish-coalescing | 1422 | 0 | 1422 |
| no-shadow | 1180 | 18 | 1198 |
| no-unsafe-enum-comparison | 1000 | 0 | 1000 |
| no-unsafe-assignment | 623 | 0 | 623 |
| no-confusing-void-expression | 483 | 1 | 484 |
| consistent-return | 425 | 9 | 434 |
| switch-exhaustiveness-check | 411 | 13 | 424 |
| unbound-method | 281 | 0 | 281 |

The next TypeScript candidates were no-unused-vars (266), no-deprecated (238),
no-unsafe-return (126), no-unsafe-call (95), consistent-type-imports (85) and
no-unnecessary-template-expression (84). A final broader counter covers all
197 checker-dependent registry rules; the original 62 counts did not change.
Higher-volume unported rules in other families include no-use-before-define
(7,843), adamic/no-unchecked-cast (3,586), no-implicit-coercion (924),
nexus/correctness-no-caller-data-mutation (820), no-param-reassign (678),
adamic/invariant-mutable (575), adamic/no-optional-widening (287), and
nexus/consistency-no-property-alias (282). Both full tables are preserved in
[validation-volume](validation-volume). Selection is intentionally within the
TypeScript rule family carried by this bridge work, not a claim about a global
cohere top ten.

## Implementation and byte oracle

`volume_suite.ts` owns the rule decisions. It loads once, parses each root using
Adamic's TypeScript parser, indexes parents, dispatches the sixteen rules and
sorts complete canonical diagnostic lines, retaining duplicates. The Go oracle
has its own loader and AST walk and calls the pinned production `Run` methods
unchanged, using their program views and shared file cache. It imports no bridge
implementation. Sorting changes diagnostic order only; fix order is preserved.
Each fix records byte start/end and replacement text. Every suggestion records
message ID, message text and every fix. UTF-16 frame lengths and UTF-8 ABI lengths
remain distinct and checked.

The ABI stays version 1: new questions use the existing `tsgo_inspect` interface.
The bridge supplies compiler facts, never lint verdicts or edits. New facts are
assignability between live type identities, widened graphs, enum parent types,
type symbol names, binder scope locals, call return types, property types,
contextual types, symbol declaration origins, type declaration origins, property
value-declaration metadata, call signature counts, callback parameter types,
apparent types, base types and the noImplicitThis option. Shapes omit display
names until a message needs them. See [facts.md](../../../bridge/tsgo/facts.md)
and [tsgo.h](../../../bridge/tsgo/tsgo.h) for framing and ownership.

Adamic performs generic assignment recursion, destructuring, scope comparisons,
completion analysis, nullish tests, enum overlap, unsafe assertions, exhaustive
case matching, unbound method exemptions, report spans and edit construction.
Completion sets distinguish normal, thrown, return and targeted break/continue
paths, including labels, loops, switch fallthrough and reachable catches.
A unique readonly suggestion tag avoids structural overlap with Diagnostic,
which otherwise made the compiler's ownership proof recursively expensive.
No protected emitter, lowering, native driver or oracle files were edited.

The compiler output includes 578 fixes and 1,422 suggestions (each with one
fix); the repository output includes three fixes and no suggestions. Complete
fields, not just these counts, match. The final sanitizer comparison preserved:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| Compiler, all 77 files | 14232 | 6717107 |
| Repository, frozen 287-file manifest | 46 | 31862 |
| Generated controls, 49 files | 73 | 24667 |

Controls cover Unicode and CRLF spans, chained member access, generic and fresh
assertions, spaced/commented empty Map assertions, nested destructuring, scope
merging and shadowing, nullable repairs, enums, union/unique-symbol switches,
Promise-like callbacks, method binding and default-library origins, return
completion, hexadecimal bigint truthiness and non-ASCII function names.
The earlier 47-file control set had 72 findings and 24,232 bytes and was used for
the fifteen judgment-question mutants; the final controls add two edge cases.
Compressed canonical finding outputs and matching Go/native SHA-256 hashes are
committed beside the manifest and final source hashes. Headers contain the
original absolute paths; hashes therefore describe this run, not relocated bytes.

## Measurements

Three alternating rounds, isolated from builds/tests/counters, load one program
and run all sixteen rules over the entire manifest. Medians below are seconds;
process includes startup and teardown, run excludes checker program loading.
Medians of separate phases need not add exactly to the process median.

| Corpus / implementation | Load | Run | Whole process | Findings/second |
| --- | ---: | ---: | ---: | ---: |
| Compiler native | 0.286 | 25.463 | 25.782 | 552.01 |
| Compiler Go | 0.281 | 3.901 | 4.212 | 3378.54 |
| Repository native | 0.084 | 1.141 | 1.231 | 37.36 |
| Repository Go | 0.079 | 0.223 | 0.313 | 147.12 |

Native compiler: 854,525 queries, 9.434 seconds aggregate query time,
**11.040 microseconds/query**. Repository: 65,196 queries, 0.519 seconds
aggregate, **7.965 microseconds/query**. These are medians of each round's
aggregate divided by its query count, not percentiles of individual queries.
The runtime timer covers input copies, compiler facts, C buffer conversion and
freeing. Adamic frame decoding, rule traversal, judgments and returned-string
release are outside that timer but included in whole-run time. Go's production
rules do not make the native bridge queries; a Go per-query comparison therefore
uses the controlled probes below, rather than dividing its whole run by the
native query count. Raw alternating rounds and stderr phases are preserved.

```sh
python3 bridge/tsgo/profile/volume_bench.py /tmp/tsgo-volume/guard-final/volume /tmp/tsgo-volume/guard-final/volume-oracle /tmp/tsgo-volume/bench-isolated --corpus compiler /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/tsgo-volume/repository.manifest > /tmp/tsgo-volume/bench-isolated.log 2>&1
```

The repeated-query probes each make 10,000 queries, alternate Go/native order
for three rounds and check identical accumulated response lengths. Warm medians
exclude the first query. Widened graph: native 3.057 microseconds, direct Go
2.608 microseconds. Property metadata: native 1.847 microseconds, direct Go
1.553 microseconds. Native includes input copying, C crossing, checker lease,
serialization, UTF conversion, output copying and C-buffer freeing.
It does not decode those frames beyond reading string length. The direct-Go
microbenchmark shares the raw fact implementation; it is a cost baseline, not
the independent production-rule findings oracle. Native property timing excludes
its initial type-identity bootstrap query. The differences are about 0.449 and
0.294 microseconds; they are observed adapter differences, not isolated C call
latencies. The compiler whole-run query average includes larger scope graphs,
cold facts and varied questions, so it is not interchangeable with these probes.
The expanded suite is slower than the previous six-rule suite; no claim of
Go-speed parity or an additional performance optimization is made here.

## Mutants and refusals

Every judgment-question mutant compiled and finished normally with exit 0 and
empty stderr. Only complete finding bytes from independent Go cohere caught it:

| Question | Mutation | First differing byte |
| --- | --- | ---: |
| assignable-types | Reverse source and target | 54 |
| widened-shape | Keep fresh type instead of widening | 481 |
| enum-types | Member instead of parent enum type | 5661 |
| type-symbol | Wrong symbol name | 19638 |
| scope-locals | Wrong binder symbol name | 8112 |
| call-returns | Number instead of signature return type | 12536 |
| property-shape | Number instead of property type | 7098 |
| contextual-shape | Own type instead of contextual type | 7605 |
| symbol-origin | Current file instead of declaration file | 24220 |
| type-origin | Wrong type symbol name | 23635 |
| property-info | PropertySignature instead of value-declaration kind | 19886 |
| call-count | Zero call signatures | 15638 |
| call-parameters | Number instead of apparent callback parameter | 15638 |
| apparent-shape | Number instead of apparent type | 15638 |
| base-shapes | Omit base types | 23635 |

Initial widening and enum mutants survived weaker fixtures. Those fixtures were
strengthened with a fresh excess-property assertion and same-enum member
comparison; the mutants above then failed through bytes alone. Weak controls
were not treated as successful evidence.

The sixteenth question, strict-this, refuses `noImplicitThis:false` before any
finding output: native exits 70 with `volume suite requires noImplicitThis;
implicit-this messages are not yet ported`. Substituting the strict-null option
for noImplicitThis incorrectly permits the input: the refusal expectation catches
exit 0 instead of 70, and the independent Go findings also differ at byte 101.
The suite does not silently omit these unported special message variants.

Querying a released program using call-returns exits 70 with `invalid or released
checker handle`. Keeping the handle in the registry makes it succeed; the
required refusal catches that mutant. Two loader mutants separately prove `.a`
root loading and retention of configured declaration roots: removing the extension
option loses the root; dropping declarations changes its type to `any`.

Foundation regression tests again catch input and output lengths off by one with
ASan heap-buffer-overflow, missing output frees with LeakSanitizer, a type from
the source-file position with the byte oracle, removed link opt-in with refusal,
and a region allocation on the heap with LeakSanitizer. The C ABI checks 100
queries, nonreused handles, rejection of zero/stale handles and output strings
that survive program release. Its independent foundation sample checks 162
positions and 3,261 output bytes under sanitizers.

## Commands, environment and observed outputs

Base branch: codex/tsgo-c-library at 3c6744dca3ffdb898c271b85a4c4c99b82433b4e.
Cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. typescript-go:
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. TypeScript v6.0.3:
050880ce59e30b356b686bd3144efe24f875ebc8. No submodule pins changed.

`bash cloud/setup.sh` succeeded. Timing lines: go ready (0s), clang ready (0s),
node ready (0s), submodules (0s), build cache warm (49s), done (49s).
`nproc`: 5; cpu.max: 400000 100000; memory: 17.6GB. Go 1.27.1,
clang 20.1.8, Node 24.19.0 on Linux amd64. Each command sourced
`/workspace/adamic-tools/env.sh`. c-archive works on this toolchain/platform.

```sh
# Build counters/oracle through an overlay in cohere, replacing a virtual main
# with testdata/count_volume.go or testdata/oracle_volume.go, respectively.
(cd cohere && go build -overlay /tmp/tsgo-volume/overlay.json -o /tmp/tsgo-volume/count-all /workspace/adamic/cohere/adamic_volume_count.go)
/tmp/tsgo-volume/count-all /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest > /tmp/tsgo-volume/compiler-all.counts 2> /tmp/tsgo-volume/compiler-all.log
/tmp/tsgo-volume/count-all /workspace/adamic/tsconfig.json /tmp/tsgo-volume/repository.manifest > /tmp/tsgo-volume/repository-all.counts 2> /tmp/tsgo-volume/repository-all.log

ADAMIC_VOLUME_ARTIFACTS=/tmp/tsgo-volume/validation-final ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript ADAMIC_VOLUME_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest go test ./stage1/cohere/typeaware -run '^TestVolumeAgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/tsgo-volume/volume-final.log 2>&1
# PASS 343.157s; all 15 question mutants, stale handle, controls, both corpora normal/ASan.
ADAMIC_VOLUME_GUARD_ARTIFACTS=/tmp/tsgo-volume/guard-final ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript ADAMIC_VOLUME_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest go test ./stage1/cohere/typeaware -run '^TestVolumeConfigGuardAndMutant$' -v -count=1 -timeout 30m > /tmp/tsgo-volume/guard-final.log 2>&1
# PASS 120.591s; final edge controls, strict-this mutant, both final corpora ASan.
go test -v -count=1 -timeout 30m ./bridge/tsgo > /tmp/tsgo-volume/bridge-final.log 2>&1
# PASS 41.604s; ABI, link guard, length/ownership/position/region mutants.
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^(TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' > /tmp/tsgo-volume/fact-regression.log 2>&1
# PASS 39.446s; frame guards and request refusals.
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' > /tmp/tsgo-volume/node-filtered.log 2>&1
# PASS 12.968s; five native/Node/JS sanitizer cases and one-byte mutant.
```

The initial combined regression command `go test -count=1 -timeout 30m
./bridge/tsgo ./bridge/tsgo/checker ./internal/load ./internal/lower
./internal/native` passed checker (0.142s), load (1.687s), lower (5.767s), native
(81.569s), but bridge archive linking failed with `ar: No space left on device`.
The scratch filesystem is 8.8GB; named obsolete local mutant binaries/archives
were removed and the bridge rerun passed in full as recorded above. Tests now
remove reproducible per-mutant binaries/archives and preserve their logs.

Final `go vet ./...` and `gofmt -l cmd internal bridge/tsgo
stage1/cohere/typeaware` produced empty logs. `go test -v -count=1 -timeout
30m ./bridge/tsgo/checker ./bridge/tsgo/cost` passed the checker in 0.063s
(the cost package has no tests). The final `/tmp/tsgo-lint-cohere --no-cache
--no-fix` run over the sixteen changed/new TypeScript source paths reported
`276 rules, 17 checked, 100% Adamic-ready`, with zero findings or formatting
changes. Its log and the earlier regression logs are preserved in
validation-volume. All test output went to files, without pipes.

## Limits

The full `go test ./...` gate was not run; touched packages and the filtered Node
oracle above were used, with actual disk exhaustion documented. The rule suite
has default options only, no suppression engine or edit application, and is not
a replacement for the full cohere CLI. JSX/JavaScript populations and every
upstream per-rule fixture/options matrix were not covered. The fixed suite
explicitly refuses noImplicitThis:false. Supported default strict configurations,
77 compiler files, the frozen repository population, generated controls and
per-question mutants are the measured evidence. Source hashes describe the final
validation inputs; adding the port files to the population would change the
population used for selection. Other rule families' higher counts are recorded
but their rules were not ported in this unit.
