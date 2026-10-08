Final scratch integration also merges area/developer-tools
2adf65c2514eefbbb073bfdab922d00a4e80a3d3. It merges without source conflicts;
the only later conflict is restoring the scratch module-path stash in go.mod.
Keep developer dependency versions, restore exact SDK replacements. Scratch
merge 92bb9616 is never pushed. Previous record and loader conflict resolutions
are described in front5, with selected final source patches preserved here.
Compiler and explicitly gated build-metrics helper both build successfully.

Read internal/native/CLANG_UNITS.md whole before measurement. Split is enabled
with ADAMIC_NATIVE_SPLIT=1 / Options{Split:true,Jobs:5}; groups are 16 functions.
The measurement tool uses a fresh XDG_CACHE_HOME for its first split attempt,
then the same directory for the repeated split. ADAMIC_GATE_UNCACHED is unset.
The helper times native.Build separately from checker/lowering/C generation.
Its build-time scope includes runtime preparation, splitting, preprocessing,
cache access, object compilation and linking; it is not a sum of parallel clang
process times. Runtime preparation would need warming for a comparable green
clang-only benchmark. No native.Build measurement is reached in this red run.

| Slice measurement | Unsplit | Split | Repeated split |
| --- | --- | --- | --- |
| Generated C bytes / lines | Unreached | Unreached | Unreached |
| Clang wall time | Unreached | Unreached | Unreached |
| Parser binary bytes | Unreached | Unreached | Unreached |
| CLI attempt wall seconds | 0.767176 | 0.770804 | 0.846200 |
| Exit | 1 | 1 | 1 |
| Cache files after attempt | 0 | 0 | 0 |

Those CLI times stop before C emission and are not clang timings. The repeated
split is not a warm-cache sample: no objects exist. nproc=5, jobs=5. First
error in C emission and all three actual build commands is core.ts:11:52,
new Map<never, never>(), indirect call/class construction before enum
initialization. No hypothetical green size or timing is substituted.

Final native-proof.py run also stops at that error. All three Node dumps
compare equal, 35,456,964 bytes with the required SHA256 2014ef06..., best user
time 5.465565 seconds. Native parser diff, binary size/runtime, instruction
count and parser-specific output mutant remain unrun. perf is not installed.
The separate type-only MapLike probe's actual native/Node comparison and
one-byte native-output mutant remain recorded in front5; that is not parser
output. Minimal native-enum-map.a records the full parser's failure class.

Focused native split tests pass: TestSplitTokensDoNotRewriteLiterals,
TestUnitsPreserveSharedState and TestUnitSystemHeaderProvenance. Loader/lower
Node/Cast/ImportCycle/Record checks passed before the clean developer merge.
No full repository gate, full native parser proof or new upstream suite is
claimed for this scratch integration.

Reproduce the instruments in the scratch compiler's module:

```sh
# Helper is gated because the parser branch's base native.Options lacks Split.
GOWORK=off go build -tags stage3_split_metrics -mod=mod -o METRICS ./stage3/drivers/parser/build-metrics
python3 stage3/drivers/parser/measure-builds.py COMPILER METRICS ENTRY SCRATCH NEW_OUTPUT --jobs 5
# Existing full oracle, timing and output-mutant protocol:
ADAMIC_NATIVE_SPLIT=0 PARSER_TYPESCRIPT=STOCK_TS_6_0_3 python3 stage3/drivers/parser/native-proof.py COMPILER SCRATCH SLICE CORPUS NODE_OUT NEW_PROOF_OUT
```

No compiler edits land on the parser branch. Helper and scripts are in the
owned driver directory; scratch source resolutions are evidence only.
