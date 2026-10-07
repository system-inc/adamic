Built: native leaked-render judgments and context visitor, completing the code for the three continuation rules.
Commits: previous pushed implementation 0dd49842; this completion and evidence accompany this report.
Commands and outputs: both owned type-aware suites PASS 100.296s; sanitizers, vet and formatting pass.
Mutants: leaked-render byte 51, listener byte 107, namespace byte 4699; released registry caught by required panic.
Not covered: real JSX parsing, emitted-JavaScript comparison, every upstream fixture, full repository gate; no further claim made.

# Leaked-render completion within the parser boundary

The remaining claimed rule now has its own .a implementation. It follows
rendered expressions through ternary branches, logical conjunction/disjunction
and nullish coalescing. It separately follows falsy operands, classifies unions,
intersections, constrained type parameters, nullable/void/never/boolean parts,
number and bigint literals, and numeric enum members. The depth guard matches
Go's production rule. It preserves source spans and complete messages, and
produces no fixes or suggestions, as Go does. The owned three-rule driver now
calls it. No shared parser, registration generator, bridge file or shared
harness was edited. Existing raw checker questions suffice.

Twenty-two isolated .a controls supply only the enclosing JSX context, without
claiming to parse JSX source. Native adds numeric-index ParseNode wrappers;
the independent Go oracle adds factory-created JSX wrappers and invokes the
unchanged production Go listener. Both check the original expression AST and
checker types at their original source positions. The adapters supply no type
verdict or finding. This tests native rule judgments and the context visitor;
it does not prove either parser can consume actual JSX through this integration.

| Comparison | Findings | Identical bytes | Sanitizers |
| --- | ---: | ---: | --- |
| Isolated JSX child contexts | 18 | 7704 | PASS |
| Attribute contexts | 0 | 1133 | Normal comparison |
| Ordinary source without JSX contexts | 0 | 1133 | Normal comparison |
| Compiler, 77 roots | 0 | 5318 | PASS |
| Repository, 287 roots | 0 | 18485 | PASS |
| All three rules' original controls | 15 | 9262 | PASS |

Controls cover Unicode/CRLF, number/bigint, nullable and boolean unions,
zero and nonzero literal unions, a named zero-bigint alias, zero/nonzero numeric
enums, branded numbers, any/unknown, generics, nested conjunctions,
disjunctions, ternaries and nullish coalescing. Every byte comparison includes
fix and suggestion fields even though these rules leave them empty.
The frozen input populations and their hashes remain the ones recorded in
validation-wave-14-next; no new port file was added to those manifests.

## Mutants and ownership

The leaked-render mutant changes the reporting verdict from leaks to silent.
It compiles, exits 0 and has empty stderr. The independent complete-byte oracle
alone catches byte 51. The revalidated listener and namespace mutants likewise
exit 0 with empty stderr and differ at bytes 107 and 4699. These offsets differ
from earlier reports because the artifact paths have different lengths.
The released-handle probe still exits 70 with `invalid or released checker
handle`; retaining the registry entry makes the mutant exit 0. Native control
and corpus sanitizer runs have empty stderr, including LeakSanitizer.

## Remaining blocker and claim gate

The real .a JSX witness still produces the production Go finding at bytes
49:54 when the external Go harness presents its unchanged contents as virtual
TSX. Native exits 70 before rule traversal:

```
adamic: panic: parser slice expected CloseBraceToken, got AmpersandAmpersandToken at 54 in /workspace/wave-14-render-suite-final/jsx-witness.a
```

The native parser cannot produce JSX child nodes. This is a parser gap, not
.a module loading or suggestion serialization in the shared harness. An origin
fetch inspected 325 refs; the named codex/lint-harness-dot-a branch was not in
that fetched snapshot. No shared infrastructure was changed to bypass the gap.
The rule code and isolated judgment tests are complete, but end-to-end JSX
agreement remains unverified. Under the existing instruction to stop on other
blockers and the gate requiring the claimed rules to be complete before more
claims, this turn makes no additional claim. It does not assert that the
remaining ranking is exhausted.

## Timing and commands

Three alternating count-only rounds after validation, for the integrated
three-rule suite on the two non-JSX populations:

| Population | Native process seconds | Go process seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.315045 | 0.293499 | 4.48 |
| Repository | 0.201515 | 0.122835 | 1.64 |

These measurements include loading and are not positive-JSX throughput.
Raw rounds are preserved in validation-wave-14-render/timings.json. The same
existing Go 1.27.1, clang 20.1.8, Node 24.19.0 toolchain was used, sourcing
/workspace/adamic-tools/env.sh. nproc remains 5, with a four-core CPU quota.
The preceding successful setup was 25 seconds, with tools/submodules ready at
0 seconds and cache warm at 25 seconds.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE14_RENDER_ARTIFACTS=/workspace/wave-14-render-final ADAMIC_WAVE14_NEXT_ARTIFACTS=/workspace/wave-14-render-suite-final ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave14(RenderJudgmentsAndMutant|NextAgreementAndMutants)$' -count=1 -v -timeout=30m > /tmp/wave-14-render-final.log 2>&1
go vet ./... > /tmp/wave-14-render-vet.log 2>&1
gofmt -l stage1/cohere/typeaware > /tmp/wave-14-render-gofmt.log
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-14-render-suite-final/wave-14-next /workspace/wave-14-render-suite-final/wave-14-next-oracle /workspace/wave-14-render-bench --corpus compiler /workspace/wave-14-typescript/src/compiler/tsconfig.json /workspace/wave-14-artifacts/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-14-artifacts/repository.manifest > /tmp/wave-14-render-bench.log 2>&1
```

All test output went directly to files. Complete compressed outputs, matching
hashes, sanitizer stderr, mutant output and the real JSX boundary witness are
in [validation-wave-14-render](validation-wave-14-render/). The shared emitted
JavaScript harness and full repository gate were not run. The earlier bridge
and Node regression evidence remains in WAVE_14_NEXT_REPORT.md; this turn
changes no bridge or compiler code.
