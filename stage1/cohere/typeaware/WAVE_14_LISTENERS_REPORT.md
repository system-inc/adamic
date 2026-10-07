Added numeric SyntaxKind listener declarations to all nine claimed rule classes.
Continues pushed 9c59f97f on main e8ba3d5d; no rebase or new claims needed.
All four native oracle suites and declaration checks PASS 408.352 s; Go vet passes.
Nineteen semantic byte mutants, nine declaration mutants and released-handle mutants caught.
Numeric dispatch, real JSX, undefined labels and native pattern validation remain blocked or incomplete.

Each rule now exposes readonly syntaxKinds: readonly number[]. Values come from
the pinned production Go parser's ast.Kind constants, with declarations checked
against the actual production Go rule.Listeners composite literals:

| Owned class | SyntaxKind values |
| --- | --- |
| NoArrayDelete | DeleteExpression 221 |
| NoBaseToString | BinaryExpression 227, CallExpression 214, TemplateExpression 229 |
| NoExtraneousClass | ClassDeclaration 264, ClassExpression 232 |
| NoGlobalListenerTargetAssertion | CallExpression 214 |
| NoMockOnModuleNamespace | CallExpression 214 |
| NoLeakedNumberRender | JsxExpression 295 |
| NoLabelVar | LabeledStatement 257 |
| NoInvalidRegexp | CallExpression 214, NewExpression 215 |
| NoMisleadingCharacterClass | SourceFile 307, RegularExpressionLiteral 13 |

The class rule's file listener is needed for reference tracking and checked
literal exclusions before the literal listener. Declarations contain no kind
strings or bridge calls. The owned verifier reads production Go listener keys
and resolves their numeric identities through the pinned external ast shim.
Replacing each declaration with [999999] is caught independently. These are
metadata assertion mutants, not claims of executed shared-driver dispatch.

The speed contract cannot be fully applied with the shared API currently on
main: stage1/typescript/parser/nodes.ts exposes only readonly kind: string.
Parser.node(index) returns that ParseNode; there is no numeric kind field and
the existing drivers invoke run() instead of handing a numeric-kind node to a
listener. Existing rule bodies still use that legacy API. The declarations are
ready for the incoming shared kind-indexed driver; this unit does not claim
numeric dispatch, elimination of string kind comparisons, elimination of
per-rule node fetches, or a speedup. Updating the shared parser/harness is outside
the authorized territory, so neither was edited. No new bridge question was added.

Only the nine owned rule declaration lines and an owned verifier changed code.
The sole branch pushed by this worker remains based on origin/main e8ba3d5d.
Remote checks confirmed main and the previously pushed branch SHA were unchanged
during testing. No new rules were selected or claimed.

Setup succeeded in 25 s: Go, clang, Node and submodules ready at 0 s, cache warm
at 25 s. nproc 5; cpu.max 400000 100000; memory 17.6 GB. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Commands source /workspace/adamic-tools/env.sh. Every test invocation
writes directly to a log file, retained in validation-wave-14-listeners.

```sh
go test ./stage1/cohere/typeaware \
  -run '^TestWave14NumericListenerDeclarations$' -count=1 -v
go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
go vet ./...
gofmt -l stage1/cohere/typeaware/wave_14_listeners_test.go
git diff --check
```

The full owned invocation uses the original frozen compiler and repository
populations through ADAMIC_TYPESCRIPT_SOURCE, ADAMIC_WAVE14_COMPILER_MANIFEST,
ADAMIC_WAVE14_REPOSITORY_MANIFEST and the corresponding NEXT manifest variables.
All four artifact directory variables point to /workspace/wave-14-listeners-*
so complete streams are retained. The compiler manifest has 77 roots and the
repository manifest 287 roots. No corpus was filtered to hide a finding.

| Suite | Result |
| --- | --- |
| Numeric declarations | PASS 0.00 s |
| Listener assertion and module namespace | PASS 75.16 s |
| Leaked-render judgments | PASS 46.32 s |
| Original three rules | PASS 67.29 s |
| Third batch supported paths | PASS 219.57 s |
| Complete owned invocation | PASS 408.352 s |

All supported findings, fixes and suggestions agree byte for byte with the
independent unchanged production Go rules, in normal and ASan/UBSan/LeakSanitizer
builds. Original controls have 33 findings; compiler roots have three original
rule findings. Continuation controls have 15 findings and four module settings
also agree. Render controls have 18 findings in synthetic JSX contexts. Third
controls have 60 findings, constructor controls 58. Both frozen corpora agree
in every suite. Native stderr is empty for supported cases. Scratch path lengths
change stream sizes relative to previous runs, with both implementations agreeing.

Nineteen compiled semantic mutants exit 0 with empty stderr and fail only the
complete diagnostic-byte comparison. First differing bytes from this run:

| Mutant | Byte |
| --- | ---: |
| Array delete | 57 |
| Stringification | 1546 |
| Extraneous class | 4376 |
| Listener assertion | 99 |
| Module namespace | 4659 |
| Leaked render | 55 |
| Cooked mapping | 59 |
| Constant writes | 8710 |
| Constructor flags | 463 |
| Overridden literal checked twice | 10280 |
| Constant deduplication | 6690 |
| Alias tracking | 12805 |
| Global writes | 21991 |
| Label membership | 54 |
| Flag precedence | 5164 |
| Unicode quoting | 5767 |
| Surrogate decoding | 18113 |
| Combining class | 6923 |
| Scope meaning Value to Variable | 1362 |

Original, continuation and third suites separately prove that released handles
panic 70; retaining the released registry entry exits 0 and is caught by that
required-panic assertion. The nine declaration mutants are separate from these
runtime mutants. Go vet, formatting and whitespace checks produce no diagnostics.

Quiet alternating three-round whole-process count medians after testing:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.704653 | 0.340303 | 5.01x |
| Repository | 0.231234 | 0.123125 | 1.88x |
| Controls | 0.021347 | 0.028423 | 0.75x |
| Constructors | 0.023782 | 0.028351 | 0.84x |
| Upstream | 0.026773 | 0.031790 | 0.84x |

Compilation and sanitizers are excluded. Constructors and upstream use
--class-only in both implementations. Upstream count is 89; its full byte
comparison and sanitizer evidence from the preceding landing unit remain in
WAVE_14_LANDING_REPORT.md. This resume does not repeat that standalone comparison.
Benchmark rounds, streams and the script are retained. The declarations are
ignored by current dispatch, so these timings do not establish a dispatch gain.

Other existing boundaries remain demonstrated: real JSX fails the shared
parser; undefined labels fail its Identifier-only label gate. Native pattern
validation and regexp2-compatible errors are still missing substantive owned
implementation: production Go reports invalidRegexp for new RegExp('['), while
native explicitly panics NotYet with exit 70. No further rules are claimed while
these paths remain incomplete. Nondefault options, new-rule emitted JavaScript,
the complete repository gate and the shared numeric-driver integration are not
claimed covered. Bridge packages and the filtered Node oracle were green in the
preceding landing unit; no bridge or compiler implementation changed here.
