Changed: three owned JSX handlers declare ast.Kind names and node: true; no new rules claimed.
Commits: named declarations first pushed as 2b780538b, rebased onto b8fb957a; source SHA is pinned in evidence metadata.
Checks: all 16 landing steps pass; 31 Node fixtures, 131 JSX controls, both corpora, sanitizers and released handles pass.
Mutants: three JSX message IDs, three named listeners, one library data mutation and all prior rule/raw-fact/ownership guards are caught.
Not covered: positive JSX on the shared main parser, shared type-aware context integration, emitted-JavaScript rule comparison and the full repository gate.

## Exhausted selection

The final all-origin audit inspects 579 refs and 33 distinct claim blobs. It
excludes the original ports, native ports and descriptors on main/base, and all
origin claims. No unported and unclaimed ranked rule remains. While the named
listener correction was being validated, wave 18 claimed the last two available
names. No new claim or implementation was made by wave 07. The three original
analysis-heavy React hook claims remain parked for native IR/SSA/capture work.
The pinned audit is in evidence/final-trio-selection.json.gz.

## Kind correction and shared boundary

Rule-owned manifests and exported declarations now use Kind-prefix-free names,
matching the registry at ab70f38d4: fragments use JsxElement, JsxSelfClosingElement,
JsxFragment; undefined names use JsxSelfClosingElement, JsxOpeningElement; adjacent
inline elements use CallExpression, JsxElement. Each manifest sets node: true.
The driver's internal legacy-node normalization resolves each declared name once;
its internal numeric representation is not a public registry API. Handlers retain
the handed node; the indexed driver alone decides relevance. Compiled declarations,
manifests and actual production Go Run keys match on 184 bytes. Each named-listener
mutant substitutes Unknown, compiles, exits 0 with no stderr and is caught only
by full production finding bytes. Three message-ID mutants behave the same way.

The shared harness context at ab70f38d4 contains no checker-program handle. These
standalone type-aware handlers are not claimed as registered with that context.
No shared registry, harness or emitter was edited. Main's JSX parser remains
unchanged by b8fb957a; positive controls still require the isolated published
parser a8a62d62ca49db7415e14c3887dd305022b17309. The Go-byte controls and both corpora
were rerun with the freshly built b8 native compiler and ordinary/sanitized archives.
The three JSX rules contain no Go regex matcher, so no regex translation or
hand-written replacement matcher was introduced. The earlier regex-literal rule
uses the generic regex syntax walk, not a Go regexp matching expression.

## Observed green checks

All 16 owned landing steps have latest exit 0: original trio; timer, process,
streams, rest, promise and regex; continuation and regex question guards; bridge;
fact decoder/refusals/flags; filtered Node oracle; vet; formatting; main JSX blocker
and published-parser dependency probes. The filtered Node oracle includes the new
inherited_static_field_read fixture and passes 31 fixtures, the one-byte mutant
and iterator refusals. Upstream protected files were taken by rebase and untouched.

The 131 JSX controls match on 28,512 bytes / 65 findings. Element mode matches
24,534 / 48; globals 27,829 / 63; both 23,851 / 46. All four also pass sanitizers.
Repository 287 files: 18,485 bytes / zero findings. Compiler 77 files: 5,318 /
zero findings. All messages, ranges, fixes and suggestions are compared; these
production rules offer no fixes or suggestions. Existing symbol metadata succeeds
live and panics 70 after release. Library data matches all 114 external sources
on 3,793,522 bytes under sanitizers; its valid data mutant exits 0 and fails bytes.
Source lint passes 276 rules over 15 owned .a files.

## Failures and recovery

Disk pressure blocked the first rebase. A gate began on the previous base;
it was stopped and not counted. The successful rebase was explicitly verified
before the replacement gate. Named obsolete owned executables and C archives
were removed, preserving sources/logs and leaving shared caches unchanged.
Two later question attempts found dependencies deleted by that cleanup: first
the timer compiler, then the streams normal archive. Both failure logs are
preserved; dependencies were rebuilt and question checks resumed successfully.
The first filtered Node run failed writing its gate cache with no space left;
the inherited static fixture itself passed. More completed owned outputs were
freed and the entire selected Node gate passed on retry. Recovery inventories,
failed logs and latest successes are archived. These failures are not mutant kills.

## Cost observations

Three quiet final process pairs per rule/corpus after all heavy checks. Each
output also matches Go. Times include process startup, checker creation, parsing
and handling; they are not handler-only measurements.

| Corpus | Rule alias | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| repository | fragments | 0.322117 | 0.128375 |
| repository | undef | 0.319322 | 0.137483 |
| repository | adjacent | 0.345811 | 0.134777 |
| compiler | fragments | 2.004307 | 0.294590 |
| compiler | undef | 2.071835 | 0.310129 |
| compiler | adjacent | 1.947400 | 0.296995 |

Setup timing lines: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready
1s, cache warm 44s, done 44s. nproc is 5 with a four-core quota. Environment:
/workspace/adamic-tools/env.sh. Native remains slower; no speedup is claimed.

## Reproduction

From the repository with that environment, save every command's output to a log:

- ADAMIC_WAVE07_LANDING=/workspace/wave-07-latest-landing python3 stage1/cohere/typeaware/wave07_react/landing_gate.py
- On a recoverable failure, use ADAMIC_WAVE07_RESUME=<failed step>; the owned runner checks every skipped step has a recorded passing result.
- python3 stage1/cohere/typeaware/wave07_jsx/validate.py /workspace/wave-07-jsx --compiler /workspace/wave-07-typescript
- python3 stage1/cohere/typeaware/wave07_jsx/validate_libraries.py
- python3 stage1/cohere/typeaware/wave07_jsx/benchmark.py

Fresh evidence records exact command arguments, exits, source hashes and main pin.
Only codex/typeaware-wave-07 is pushed. No main or area branch is pushed.
