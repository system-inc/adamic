Built: native validator cores for all three claimed React rules, with native post-dominance; source-to-HIR integration remains missing.
Commits: existing landing-ready branch 01f9d6f6; native cores 76e453e7, based on main e8ba3d5d.
Checks: TestWave18ReactCores PASS 455.423s; 91 valid controls, 48 findings; both corpora match normally and under sanitizers; vet clean.
Mutants: render guard byte 456, effect setter byte 1556, static creator byte 57, adjacency-row alias byte 3123; all caught only by Go bytes.
Not covered: native source lowering/SSA, source gates/memo shadowing, end-to-end timing, full gate, or deterministic parity for ambiguous two-creator Go phis.

The landing cap was satisfied before continuation: 01f9d6f6 already contained
current main e8ba3d5d, its oracle had been rerun and its own origin branch was
up to date. A new fetch inspected 466 origin refs. Wave 21 now publishes native
validator cores, but not a native source adapter. The cores and private test-only
fixture producer were adapted from its e9ec024c tip into owned wave-18 files.
The final main fetch still reports e8ba3d5d as an ancestor. No new rules were claimed.
Only codex/typeaware-wave-18 is pushed; main and area/ branches are not targets.

## Native code and validation boundary

All six new Adamic modules are .a under cores/. They define typed HIR places,
identities, instruction operands, phis, captures, nested functions and blocks,
then execute the three validators and post-dominance analysis in native Adamic.
The render validator translates captures, tracks unconditional setter functions,
handles memo callbacks and excludes optional calls. The effect validator carries
setter provenance, follows synchronous callbacks and effect events, propagates
ref-derived values/patterns and applies control-frontier exemptions. Static-
components carries creator provenance through loads/stores/phis and preserves
Go's single forward walk and tag ranges.

The private Go fixture producer invokes actual production HIR Lower/Construct,
standalone compilation-unit selection and memo erasure/inlining, then exports
raw prepared graphs, ranges and type names as generated .a inputs. Native code
computes the findings. A separate execution of the unchanged production Go
registry rules supplies expected findings, fixes and suggestions. The expected
path imports no bridge code and does not execute native validators.

These are prepared-HIR checks, not native source frontend coverage. Unit selection,
SSA construction, memo annotations and shadow scope handling are dependencies of
the fixture provider. Native source analysis remains explicitly refused by the
existing reporter entry points. No Go HIR lowering was installed in the production
checker bridge. No bridge query or handle path is added by these pure-input cores;
new-question released-handle tests therefore do not apply to this increment.
Existing bridge ownership/mutant evidence remains in WAVE_18_LANDING_REPORT.md.

## Findings and mutants

The extractor retains all 100 candidate fixture strings. Go's parse filter admits
91; nine invalid source strings are excluded explicitly, not because of findings.
Twelve batches retain all 91 valid inputs and yield 48 findings and
37138 canonical bytes. Every batch matches normal and
ASan/UBSan/LeakSanitizer execution, with empty sanitizer stderr. Full diagnostic
records retain zero-fix and zero-suggestion counts for these three rules.

The compiler corpus (77 sources) matches 5318 bytes and the frozen repository
corpus (287 sources) matches 18485 bytes, with zero findings, normally and under
sanitizers. These use prepared Go HIR too and do not prove native parsing/lowering.

Each mutant compiles, exits 0 and has empty stderr. Only complete comparison with
independent Go bytes rejects it:

| Mutant | First differing byte |
| --- | ---: |
| Reverse render unconditional membership guard | 456 |
| Reverse effect typed-setter guard | 1556 |
| Delete call/new creator propagation | 57 |
| Share one mutable predecessor row across keys | 3123 |

The last mutant reproduces a genuine post-dominance adjacency aliasing failure.
Fresh rows keep separate predecessors; sharing the empty row loses the loop-after-
throw finding. No mutant is killed by clang or an unrelated panic.

## Independently observed Go ambiguity

An additional source assigns createA() and createB() on opposite branches and
renders the resulting local component. Thirty-two executions of unchanged Go,
with identical source/config, produced two different complete finding streams,
frequencies 27 and 5. Both have one static-components finding, but name different
creation sites. Phi.Operands is a Go map and the validator selects its first
dynamic operand. The native core preserves supplied operand order; deterministic
byte equality against a separate Go process cannot be promised for this input.
No output was normalized, and this probe is not counted among passing controls.
Its exact source, manifest, all stdout/stderr and summary are retained under
validation-cores/phi-order. Reproduce by running the prepared test's independent
react-core-oracle repeatedly with tsconfig.json and that manifest.

## Commands, timings and limits

    source /workspace/adamic-tools/env.sh
    export ADAMIC_WAVE18_REACT_CORE_ARTIFACTS=/workspace/wave18-react-cores-final
    export ADAMIC_WAVE18_COMPILER_CONFIG=/workspace/wave-18-typescript/src/compiler/tsconfig.json
    export ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest
    export ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest
    go test ./stage1/cohere/typeaware -run '^TestWave18ReactCores$' -count=1 -timeout=30m -v > /tmp/wave18-react-cores-final.log 2>&1
    go vet ./stage1/cohere/typeaware > /tmp/wave18-react-cores-vet.log 2>&1

Test PASS 455.423s; vet exit 0. Go formatting and git diff --check are clean.
The first preliminary run matched its first batch, then was deliberately stopped
before source formatting. The six modules were formatted with CoHere's native
printer through the existing private adapter. Mutation anchors were adjusted and
the complete final run used those formatted sources. The preliminary log is
preserved and is not presented as a completed pass.

Setup: Go ready 0s, clang ready 1s, Node ready 1s, submodules 1s, cache warm 91s,
done 91s; nproc 5, CPU quota 4. The printed step timings overlap/round and are
reported exactly as printed. Toolchain is Go 1.27.1, clang 20.1.8, Node 24.19.0.

Across 91 controls, observed prepared-input native process time is 0.018864s,
full Go source-analysis time is 0.256744s and Go graph preparation is 0.288546s.
These pipelines do different work. Native time excludes source loading/lowering;
there is no end-to-end native/Go speed claim. No comparable source lint ratio is
available until the native adapter exists.

The prior reporter/refusal and JSX parser evidence remains in REPORT.md and
WAVE_18_REACT_BLOCKER_REPORT.md. The source adapter, gates and memo-shadow policy
are still missing; the base parser still lacks the required JSX paths. No shared
harness, registration generator, native parser or protected compiler file was
edited. No full root gate was run. Claims remain unfinished and reserved; no
additional batch is taken. Logs are gzip-compressed without truncation, and
inputs, source hashes and summaries are committed with this report.
