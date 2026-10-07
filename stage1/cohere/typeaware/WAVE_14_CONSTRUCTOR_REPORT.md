# Wave 14 constructor completion

Built native constructor tracking, constant arguments and cooked source mapping.
This resumes pushed ec47031e on codex/typeaware-wave-14; no new claims.
Rule suite PASS 174.233 s; normal and sanitizer corpora agree byte for byte.
Thirteen comparison mutants and the released-handle mutant are caught.
Pattern validation, undefined labels and real JSX remain incomplete.

## Native implementation

no_misleading_character_class.a now checks constructors at file entry, before
its regex-literal pass, using three new owned .a modules:

- regexp_reference_tracker.a ports the RegExp specialization of cohere's global
  ReferenceTracker. It follows direct and aliased globals, global/globalThis/self/
  window members, object destructuring, assignments and defaults, parentheses,
  logical/conditional/comma expressions and type-only wrappers. Writes to a
  global suppress its trace. Local alias traces are flow-insensitive, as Go's
  are, and duplicate routes preserve duplicate findings. Computed keys use the
  scope-less constant reading. Cyclic binding traces terminate.
- regexp_constant.a ports constant strings, templates, primitive literals,
  concatenations and local constant initializers. It distinguishes RegExp objects
  from string values and checks writes to mutable bindings by symbol identity.
  A shorthand assignment uses its value symbol rather than its property symbol.
- regexp_cooked_offsets.a ports literal.CookedToRaw at Go byte granularity, with
  explicit UTF-8/CESU-8 encoding and Go-compatible invalid-rune widths, then
  translates offsets back to native UTF-16 source positions. It preserves Go's
  current escaped-surrogate and raw-tail behavior, including declined mappings.

String/template argument findings point into raw characters. Other constant
arguments report once per message ID at the argument. A literal passed with
flags is checked under the call's flags and excluded from the literal pass;
unknown flags still exclude it. A literal passed alone stays in that pass and
can retain its suggestion. Constructor strings and overridden literal patterns
have no automatic edits or Unicode-flag suggestions, as in Go.

No bridge question was added. Existing raw symbol identity, declaration origin
and resolved-name facts supply the checker information; all trace, constant and
class judgments execute natively. No shared registration generator, test
harness, parser, protected compiler file or cohere production rule changed.
The owned Go oracle gained --class-only to invoke the unchanged production class
rule independently of the separately unfinished pattern validator.

## Byte agreement and sanitizers

| Population | Inputs | Findings | Complete bytes |
| --- | ---: | ---: | ---: |
| Existing three-rule positive/negative controls | 54 | 60 | 19622 |
| Constructor and reference controls | 62 | 58 | 24379 |
| Upstream constructor fixtures | 83 | 89 | 38351 |
| TypeScript src/compiler frozen roots | 77 | 0 | 5318 |
| Repository frozen roots | 287 | 0 | 18485 |

Every row agrees on the entire canonical stream, including every fix and
suggestion field, in normal and ASan/UBSan/LeakSanitizer builds. Native stderr
is empty on supported inputs. agreement.json records final-run hashes and
sizes; complete output, stderr, inputs, source hashes and logs are retained in
validation-wave-14-constructors. Earlier numbered files from failed attempts
are also retained; agreement.json selects the latest run for each population.

The 62 constructor controls include all six class IDs, raw escapes, Unicode and
CRLF, templates/substitutions, flags and pattern bindings, mutable writes and
shadowing, concatenation/deduplication, cyclic constants, RegExp-object inputs,
flag overrides, unknown flags, alias cycles, destructuring/defaults, global
writes, type wrappers and duplicate trace routes.

The upstream extractor reads the pinned Go test's Fires and StaysSilent default
constructor cases without rewriting their source. It finds 84 inputs. The
existing independent Go --valid-sources selector excludes one strict-module
parse refusal: upstream-067.a, a legacy octal string escape. The remaining 83
compare byte for byte. The excluded source is retained, not counted as a pass.
Inputs add export {} to isolate scopes, as the unit's prior controls do.
The extractor and both selected/excluded manifests are retained.

Cohere is pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go to
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the frozen TypeScript v6.0.3
compiler population to 050880ce59e30b356b686bd3144efe24f875ebc8. The compiler and
repository manifests/configs are the original branch populations, not a filtered
replacement chosen to evade findings.

## Mutants

Each of these builds, exits 0 and has empty stderr. The independent full Go byte
comparison is the check that catches it:

| Mutant | First differing byte |
| --- | ---: |
| Cooked mapping shifted one byte | 61 |
| Mutable-binding writes ignored | 8740 |
| Constructor Unicode flags ignored | 467 |
| Overridden regex literal checked a second time | 10326 |
| Constant-argument message deduplication removed | 6710 |
| Alias target followed as global root instead of constructor | 12871 |
| Writes to traced globals ignored | 22097 |
| Label membership inverted | 56 |
| Duplicate flag precedence lost | 5208 |
| Unicode printable quoting inverted | 5817 |
| Lone surrogate decoded as one replacement instead of three | 18217 |
| Combining-mark detector inverted | 6979 |
| Scope meaning Value changed to Variable | 1374 |

The additional ownership mutant retains a released program in the Go registry:
it exits 0, while the proper program panics 70 with invalid or released checker
handle. That assertion catches it separately from diagnostic comparison.

## Commands and observed corrections

Toolchain setup succeeded: Go, clang, Node and submodules ready at 0 s, build
cache warm at 19 s, setup done in 19 s. nproc is 5, with four cgroup cores.
Go 1.27.1, clang 20.1.8, Node v24.19.0. Commands source
/workspace/adamic-tools/env.sh; all output went to log files.

```sh
bash cloud/setup.sh > /tmp/wave-14-constructor-setup.log 2>&1

ADAMIC_WAVE14_THIRD_ARTIFACTS=/workspace/wave-14-constructor-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript \
ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest \
ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest \
go test ./stage1/cohere/typeaware \
  -run '^TestWave14ThirdAgreementAndMutants$' -count=1 -v -timeout=30m \
  > /tmp/wave-14-constructor-final.log 2>&1

go vet ./... > /tmp/wave-14-constructor-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware \
  > /tmp/wave-14-constructor-gofmt.log 2>&1
git diff --check > /tmp/wave-14-constructor-diffcheck.log 2>&1
```

Final suite PASS 174.233 s. Static Go and whitespace checks pass. Pinned native
Go formatting was applied only to the owned .a files through a scratch overlay
and virtual .ts filenames; no Adamic .ts file was authored.

The first full-byte constructor comparison caught the shorthand symbol mismatch
and lone-surrogate mapping discrepancy. Both were fixed, with those inputs kept
in the final controls. An expanded compiler run then exposed speculative native
parser nodes with no AST parent; indexing now ignores nodes not linked into the
file, matching Go's actual AST walk. orphan.log preserves that failure and the
final corpus comparisons prove the correction. Stage 0 also explicitly refused
an inferred never[] and multi-value push; typed local arrays and separate pushes
fit existing support without changing the compiler. The direct-only milestone
passed in 164.298 s before reference tracking was added; it is historical evidence,
not the final result. Saved failed and milestone logs keep these observations
separate from the final passing checks.

The upstream checks use third-oracle/third/third-asan with the independent
valid.manifest and --class-only. The extractor, manifests and complete streams
are saved for reproduction. The full repository test gate and new-rule emitted
JavaScript comparison were not run. Earlier full bridge and filtered Node gates
remain documented in WAVE_14_THIRD_REPORT.md; this change adds no Go bridge code.

## Native time against Go

Quiet, alternating three-round whole-process --count medians follow verified
byte agreement. Compilation and sanitizers are outside these timings. Constructors
and upstream use --class-only in both implementations. Raw timings and the
reproducible measurement script are retained.

| Population | Native | Go | Native / Go | Findings |
| --- | ---: | ---: | ---: | ---: |
| compiler | 1.638156 s | 0.346199 s | 4.73x | 0 |
| repository | 0.196181 s | 0.117518 s | 1.67x | 0 |
| controls | 0.021180 s | 0.028211 s | 0.75x | 60 |
| constructors | 0.023150 s | 0.028502 s | 0.81x | 58 |
| upstream | 0.023329 s | 0.033957 s | 0.69x | 89 |

## Remaining work

The owned default class rule no longer refuses direct constructor strings or
constructor aliases. Its production registration and shared JavaScript harness
integration are left to the shared worker. allowEscape configuration beyond the
branch's default-option protocol and the excluded legacy-octal strict-module
case are not claimed covered.

no-invalid-regexp still explicitly refuses literal-pattern paths because native
ECMAScript rewrite and regexp2-compatible validation/error messages are missing.
That is substantive owned implementation, not a shared harness gap. The final
Go-positive new RegExp('[') witness demonstrates the boundary.
no-label-var still hits the shared parser's Identifier-only label gate on
undefined: for(;;) {break undefined;}; its Go-positive witness and native
exit-70 refusal remain. Real JSX still blocks the earlier leaked-number-render
end-to-end path. Neither shared parser was edited or merged into this branch.

All available work in this resume is committed and pushed on the existing branch.
No additional rules are selected or claimed while these claimed paths remain
incomplete. Earlier reports remain historical records; this report supersedes
their constructor-tracking and cooked-mapping gap descriptions.
