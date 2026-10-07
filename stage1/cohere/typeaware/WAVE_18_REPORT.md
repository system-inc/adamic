# Type-aware wave 18

Built two native default-option ports: `await-thenable` and `class-literal-property-style`.
Claim commit: `27d09b55`; implementation commit: `adea6fbe`.
Commands and outputs: setup, byte agreement, bridge packages, filtered Node oracle and vet logs are in `validation-wave-18`.
Mutants: two rule predicates/spans, two checker facts, released-handle registry, and the existing Node one-byte mutant.
Not covered: the Next JSX rule, nondefault class-literal options, and the unfiltered repository-wide Go test gate.

## Selection and ownership

Base: `0d540f413625f016f20fea39761c7b184f335de6` on `origin/codex/tsgo-c-library`.
Branch: `codex/typeaware-wave-18`. The claim was committed and pushed before implementation.
Positions 52, 53 and 54 in the by-volume population, after excluding the 26 existing ports,
are `@next/next/no-title-in-document-head`, `@typescript-eslint/await-thenable`, and
`@typescript-eslint/class-literal-property-style`. All have zero findings in the frozen corpora.
The ranking combines the two validation-volume count files and breaks ties by rule name.
The separate coverage inventory contains a different population. Every fetched origin branch's
claims and named native ports were checked; no selected rule was skipped for ownership conflict.

## Native implementation

`await_thenable.a` implements ordinary await, Promise aggregation, for-await and await-using
listeners. Raw compiler properties, signature callback parameters, rest element types,
array/tuple elements and known-symbol existence cross the bridge; Adamic decides whether
values are thenable and decides every finding and repair. `class_literal_property_style.a`
implements the production default `fields` mode, including setter pairing, inherited concrete
accessor exemptions, decorators, literal kinds and the production raw-source replacement.
The driver serializes the complete diagnostic protocol including every fix and suggestion.

The new questions are implemented separately in Go and Adamic:

- `iteration-type-facts\n<live type identity>\n<operation>` accepts `iterator`, `asyncIterator`,
  `asyncDispose`, `values` or `callbacks`. Symbol operations return raw property presence.
  Values return raw array-like status and element roots; callbacks return raw first-signature
  parameter types, using the numeric index type for a rest parameter. Graphs reuse the existing
  versioned type frame. Noncanonical, missing or unknown identities and operations are refused.
- `base-member-facts` accepts an exact class getter/property node. It returns the member modifier
  flags, inherited property symbol/check/modifier flags and whether each declaration belongs to
  an interface. The inherited conversion exemption is decided in Adamic.

The sole shared source edit is two one-line switch registrations in `bridge/tsgo/checker/facts.go`.
They remain single lines to honor this wave's shared-file constraint. Protected compiler files,
existing rule suites, shared fact schemas and submodule pins were untouched. Every new Adamic
source file is `.a`.

## Measured JSX gap

The Next rule is **not ported**. Its production visitor requires JSX elements and their nesting;
the current native parser has no JSX element grammar. The independent Go oracle reports the
rule on this positive control:

```tsx
import {Head} from 'next/document';export const page=<Head><title>x</title></Head>;
```

Native parsing exits 70 with `parser slice expected GreaterThanToken, got Identifier` at that
control. The test asserts the Go positive and the specific native refusal, rather than treating
zero findings on JSX-free corpora as a successful port. A faithful third port needs a JSX parser
slice before its visitor can run. Its claim remains explicit and this report records the blockage.
The generated `.tsx` is a TypeScript oracle input, not an Adamic implementation file.

## Validation

Both engines load the same config and roots. The independent Go program uses the pinned cohere
production registry and visitors unchanged; it imports no bridge implementation. Defaults match
the preceding 26 ports. Findings, messages, UTF-8 ranges, fix texts/ranges and suggestion texts/ranges
are sorted and compared as complete bytes, not merely counts. Twenty generated sources provide
38 positive findings, including Promise aliases, computed keys, generic constraints, rest callbacks,
sync/async iteration/disposal, inheritance, private/decorated getters, Unicode and CRLF.

The compiler corpus uses TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, and the existing 77-file compiler manifest.
The repository corpus is the existing frozen 287-file manifest, not a newly expanded manifest.
Both corpora have zero selected-rule findings; their silence is complemented by the positive controls.
Normal and ASan/UBSan builds run all three manifests, with LeakSanitizer enabled by the harness.
No sanitizer stderr is accepted for successful native runs.

Commands (all output redirected to logs):

```sh
bash cloud/setup.sh > /tmp/wave-18-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE18_ARTIFACTS=/workspace/wave-18-validation-final \
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest \
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave18AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave-18-test-final.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /tmp/wave-18-bridge-gate.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  -count=1 -timeout=15m -v > /tmp/wave-18-node-oracle.log 2>&1
go vet ./... > /tmp/wave-18-vet.log 2>&1
gofmt -l cmd internal > /tmp/wave-18-gofmt.log
```

The added Go files were also checked with gofmt; `git diff --check` passed. Setup reported Go,
clang, Node and submodules ready in 0s each, cache warming 76s, total 76s. `nproc` was 5;
the cgroup quota was four CPUs. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

Pinned cohere's CLI declines `.a`, so its direct CLI check is not counted as a gate pass.
Its actual native formatter was invoked through an isolated Go test overlay, using the TypeScript
printer on the five new `.a` buffers. All five were formatted and the second pass was byte-identical.
No new `.ts` implementation file was created. An attempted isolated newer cohere tool could not
build against the pinned older TypeScript API; neither submodule pin was changed. Native compilation
checks the new Adamic sources' types. A full `.a`-aware cohere CLI lint gate remains unrun.

Two initial validation corrections were made before the final run: the stock config loader needed
a declaration-file anchor to avoid a no-inputs error for generated `.a` controls, and the direct
base-facts test needed the compiler's raw transient symbol flags rather than the bare accessor enum.
Neither initial failure is counted as an agreement pass.

## Mutants and results

The final wave test passed in 92.300s. Controls matched 10,881 bytes (38 findings), the repository
matched 18,485 bytes (zero findings), and the compiler matched 5,318 bytes (zero findings), in both
normal and sanitized builds. All bridge packages passed; `TestBridge` took 89.71s and the checker
package took 0.159s. The filtered Node oracle passed in 18.216s, including all five selected fixtures.
Vet, formatting and whitespace checks had no findings.

| Mutant | What caught it |
| --- | --- |
| Await flips the never-thenable decision | Go byte comparison at byte 55 |
| Literal getter diagnostic end advances one byte | Go byte comparison at byte 4448 |
| Iteration known-symbol presence always returns false | Go byte comparison at byte 1710 |
| Inherited base symbol flags become Property | Go byte comparison at byte 7843 |
| Registry retains a released program | Exact stale-handle refusal assertion; mutant exits 0 |
| Node oracle's existing one-byte result mutant | Independent Node comparison |

The first four mutants compile and finish with exit 0 and empty stderr. No compiler or sanitizer
failure is credited as catching those mutants. The healthy stale-handle control exits 70 with
`adamic: panic: invalid or released checker handle`. The retained-registry mutant incorrectly
succeeds with empty stderr.
The controls mint a live type identity before releasing the program so that the mutation exercises
the handle gate rather than failing later on an unknown type.

## Observed time against Go

Three alternating whole-process `--count` rounds, seconds; medians are observations on this machine.
Both include program loading and their own AST traversal. The Go oracle registers all three selected
production visitors; the native driver registers the two completed ports. The two corpora are JSX-free.

| Input | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| 20 positive-control sources, 38 findings | 0.020976 | 0.024691 | 0.85 |
| Frozen repository, 287 files | 0.192059 | 0.111676 | 1.72 |
| TypeScript compiler, 77 files | 1.281899 | 0.298943 | 4.29 |

The corpus timing rounds are preserved in `wave-18.log`; the additional quiet positive-control
rounds are in `controls-timing.json`. Native made 165 queries on the controls, zero on the repository,
and one on the compiler. Native is slower on the large silent corpora. These measurements do not
establish per-rule speedups or performance on untested syntax. Existing bridge-gate mutants also
passed: input/output length errors caught by ASan, wrong source position by byte mismatch,
unlinked checker access by refusal, buffer-free removal and heap region entry by LeakSanitizer.

## Limits

This unit completes two of its three rules. It does not claim arbitrary TypeScript parity beyond
the frozen manifests and targeted controls. Class-literal's nondefault `getters` option is absent,
matching the default-only suite pattern. The native parser's existing supported syntax is inherited;
JSX is an explicit failing measurement. No PR was opened. Only the touched package tests and
filtered Node oracle were run, rather than `go test ./...`; the exact filter is above.
