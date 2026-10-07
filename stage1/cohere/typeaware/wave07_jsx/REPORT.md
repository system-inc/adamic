Built: three owned numeric JSX rule handlers in .a; analysis-dependent React claims remain parked.
Commits: claim fe438566c was pushed before implementation; the final implementation SHA is recorded in evidence metadata.
Checks: 131 controls, four option combinations and both frozen corpora match Go; sanitizers and released-handle checks pass.
Mutants: three message IDs, three numeric dispatch listeners and one library-data mutation compile, exit 0 and fail only byte comparison.
Not covered: positive JSX on the shared main parser, shared registration/profile integration, emitted-JavaScript comparison and the full repository gate.

## Selection and scope

The original nine ports were landing-ready before this claim. The three earlier
React hook claims are parked under Ahra's instruction: native high-level IR,
single-assignment/value flow and closure/capture analysis are pending on #dnv6f2c.
The all-origin audit inspected 529 refs and 33 distinct claim blobs, excluding
ports on main/base and every origin claim. `react/jsx-no-constructed-context-values`
was skipped because its stability/escape walk needs that analysis. The next
eligible three were `react/jsx-fragments`, `react/jsx-no-undef` and
`react/no-adjacent-inline-elements`. Their measured ranking volumes are zero.
The claim push preceded all implementation.

Each rule lives in its own directory. Its `rule.json` and exported declarations
use production Go's numeric SyntaxKind keys: fragments [285,286,289], undefined
names [286,287], adjacent inline elements [214,285]. Each handler accepts the
handed Node; it neither refetches that node nor reads its kind as a string.
The owned driver normalizes legacy parser strings once, indexes listeners by
numeric kind and invokes only matching handlers. Declaration metadata is decoded
at the bridge boundary. No shared parser, generator, harness or dispatcher was edited.
The existing Diagnostic model serializes findings, fixes and suggestions.

## Oracle and observed results

The independent Go loader invokes unchanged production Rule.Run implementations.
It does not import the bridge or copy its predicates. The comparison includes
full messages, IDs, UTF-8 ranges, rule identity, fixes and suggestions. These
three production rules intentionally emit no repairs or suggestions, including
fragment type arguments. The port preserves that behavior.

Default controls: 28,512 identical bytes, 65 findings: fragments 21, undefined
names 16, adjacent inline 28. Element mode: 24,534 bytes / 48 findings; global
scope mode: 27,829 / 63; both options: 23,851 / 46. All four runs also match under
ASan/UBSan/LSan with empty native stderr. The 131 source files cover alias imports,
requires, destructuring, shadowing and merged declarations, cross-file fragments,
CJS globals, intrinsic/member/namespace/Unicode tag names, fragments and props,
inline JSX adjacency, createElement arrays, literals, holes and Go's Unicode
whitespace boundaries. The first adjacent pair alone is reported, as in Go.

The frozen repository corpus is 287 files and 18,485 identical bytes; compiler
corpus is 77 files and 5,318 bytes. All three rules report zero findings there.
Both corpora match in ordinary and sanitized isolated builds. They also match
using the unchanged main parser; positive controls establish non-vacuous behavior.

All six rule mutants are valid compiled programs, exit 0 with empty stderr and
are caught only by production Go bytes. Message mutations change one rule's ID;
listener mutations replace its first real numeric kind with Unknown (0), causing
missed real dispatch. Compiled listener declarations and rule.json match actual
production Go Run keys. The existing node-symbol-details question succeeds live;
an uncached request after release panics 70 with the required invalid-handle error.

The owned pinned-library probe matches 3,793,522 external bytes from all 114
checker library sources under sanitizers. Its valid copyright-text data mutant
also compiles and exits 0; only the external-byte comparison catches it. Source
lint reports zero findings, 276 rules, 15 owned .a files. The parent table uses
an owned map: an attempted sparse array write was caught as a native panic and
corrected before the final gate. The initial failure is not counted as a mutant.

## Bridge boundary and source assets

No new checker question is required. Existing node-symbol-details supplies symbol
identity and declaration spans. The native handlers inspect declaration syntax.
Go's bundled:///libs paths are virtual, so ordinary native file I/O cannot load
them. Automatic approval review rejected the proposed shared bridge source-text
file because Ahra restricted changes to owned rule directories. The safe owned
alternative is libraries.a: an Apache-2.0 snapshot of all 114 libraries pinned
to checker 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, preserving license notices
and exact bytes. library-pins.json records every SHA256; generate_libraries.py
checks the pin before regenerating. The snapshot adds roughly 4 MB of source and
must be regenerated deliberately if the checker pin changes.

## Parser integration blocker

On current main, control-000.tsx exits 70:
`parser slice expected GreaterThanToken, got SlashToken at 97`.
The branch's existing probes establish the same shared JSX gap. Positive JSX
validation uses an isolated copy of the published parser revision
`a8a62d62ca49db7415e14c3887dd305022b17309`; every copied file is pinned by hash.
Owned rule bodies are identical. Shared files remain untouched. Integration of
that parser from area/stage1-lint is pending. Consequently these are implemented
and proven isolated handlers, not a claim of working positive JSX in today's
production native driver. Shared module/profile/suggestion harness work and the
batch-8 Diagnostic migration remain pending; no batch-8 SHA was supplied.

## Time observations

Median seconds over three quiet separate-process pairs. Every timed output is
also compared in full with Go. These times include startup, checker creation,
parsing and handling; they are not measurements of handler-only cost.

| Corpus | Rule alias | Native seconds | Go seconds | Native/Go |
| --- | --- | ---: | ---: | ---: |
| repository | fragments | 0.357243 | 0.154166 | 2.32 |
| repository | undef | 0.319685 | 0.137628 | 2.32 |
| repository | adjacent | 0.310928 | 0.125244 | 2.48 |
| compiler | fragments | 2.135686 | 0.311901 | 6.85 |
| compiler | undef | 2.067379 | 0.302973 | 6.82 |
| compiler | adjacent | 2.258282 | 0.318658 | 7.09 |

Setup passed in 28 seconds: Go 0, clang 1, Node 1, submodules 1, cache 28,
done 28; nproc is 5 (4-core quota). Toolchain environment:
/workspace/adamic-tools/env.sh. No speedup is claimed.

## Commands

With that environment sourced, all output was saved to files:

- python3 wave07_jsx/validate.py /workspace/wave-07-jsx --compiler /workspace/wave-07-typescript
- python3 wave07_jsx/validate_libraries.py
- python3 wave07_jsx/benchmark.py
- cohere-adamic --no-cache --no-fix <the 15 owned .a files>
- ADAMIC_WAVE07_LANDING=/workspace/wave-07-latest-landing python3 wave07_react/landing_gate.py

The latest fetch advanced main from f8013f0b to c01907a7. The owned branch was
rebased onto c01907a7 and the full owned landing gate was rerun. The relevant
upstream diff changes only internal/oracle/stage3_hook_test.go; parser/compiler,
bridge and cohere rule sources are unchanged. No main or area branch is pushed.
Owned obsolete executables and mutant C archives were deleted to recover disk
space. Sources, logs and prior committed proof were preserved; shared caches
were not changed. Recovery inventories are archived. See evidence metadata
and compressed command logs for actual exits and pins.
The full repository gate and emitted-JavaScript differential were not run.
