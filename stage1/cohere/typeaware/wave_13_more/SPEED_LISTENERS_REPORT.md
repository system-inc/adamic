Built: added numeric listenerKinds declarations to all nine owned .a rules; full numeric dispatch and supplied-node handling remain blocked by the shared parser/driver interface. No new claims.
Commits: implementation f147551e on main e8ba3d5d, following previously pushed 4f64511a; changes are confined to owned rule files and this directory's check/evidence.
Commands and outputs: production-listener check PASS 0.022s; original trio oracle gate PASS 242.664s; normal and sanitized Nexus 129/129 and core 430/433 controls agree; both frozen corpora agree; touched-package vet and source whitespace checks pass.
Mutants: nine compiling semantic rule mutants and two export-fact mutants are caught only by byte comparison; nine in-memory numeric-declaration mutants fail the production-listener check; the released-registry mutant fails the lifetime assertion.
Not covered: removing legacy string-kind reads and per-rule node refetches awaits the shared numeric-node/driver handoff; three no-object-constructor parser refusals remain. No whole-branch green result, new claim, full Go gate or emitted-JavaScript parity is claimed.

## Listener contract

Each public rule class declares `readonly listenerKinds: number[]`. Today's
suite drivers still call run() and ignore this metadata, as anticipated by the
request. The values are the pinned parser's numeric SyntaxKind constants, checked
against the independent production Go listener table rather than a copied list
of expected native numbers:

| Rule | Numeric listenerKinds | Parser kinds |
| --- | --- | --- |
| no-unassigned-vars | 261 | VariableDeclaration |
| preserve-caught-error | 258 | ThrowStatement |
| consistent-type-exports | 307 | SourceFile |
| correctness-no-process-exit-after-output | 307 | SourceFile |
| correctness-no-uncleared-race-timeout | 214 | CallExpression |
| correctness-require-blocking-standard-streams | 307 | SourceFile |
| no-obj-calls | 307 | SourceFile |
| no-object-constructor | 214, 215 | CallExpression, NewExpression |
| no-promise-executor-return | 220, 254 | ArrowFunction, ReturnStatement |

SourceFile subscriptions match Go cohere's actual registrations for those four
whole-file analyses. No invented all-kinds subscription is used. The listener
checker parses each production Go rule's rule.Listeners composite literal and
compares its keys using actual ast.Kind constants from the pinned parser shim.
It rejects absent, string-valued, duplicate or wrong numeric declarations.
Nine mutations change the first literal to the next number in the real native
source text; each is rejected by that comparison. These are in-memory metadata
mutants, not compiled native semantic variants. The separately repeated nine
semantic rule mutants below satisfy the native byte-comparison requirement.

## Exact shared-interface blocker

stage1/typescript/parser/nodes.ts defines ParseNode.kind as string, and its
constructor accepts a string kind. It exposes no numeric kind field. Existing
OutputContext.node(index) returns that string; Bindings/shape and raw declaration
facts also use strings. Existing suite drivers hand a file/context to run(),
with no shared per-node dispatch or supplied-node callback contract.

Consequently these declarations alone do not satisfy the complete speed rule:
legacy string reads, scans and per-rule refetches remain in the old execution
path. Converting them requires a shared numeric ParseNode field and the incoming
driver's node handoff. Adding a private string-to-number lookup or checker query
per node would preserve the overhead the request is removing, so no such
workaround was introduced. Ahra's instruction forbids changing shared parser,
harness or registration generator files and says to stop on other blockers.
Those files were untouched. This unit completed the expressly requested numeric
metadata now; full callback migration is still blocked and is not claimed done.
No performance gain is inferred from metadata that the current driver ignores.

The separate parser refusals remain exactly the same: 53d1c0a9ffc2fefa has JSX
`<foo />` in .ts and refuses GreaterThanToken/SlashToken at 5; aff6ced2ca61fe89
has `<foo></foo>` in .ts and refuses GreaterThanToken/Identifier at 7;
defcb4c4ce6a2921 has a yield label/break yield and refuses a semicolon at 37.
Go exits 0 with 699, 699 and 611 bytes. Both native modes exit 70 before rules
run, with zero stdout. These three failures are retained, not made green.

## Fresh oracle and sanitizer evidence

| Population | Findings | Equal bytes, normal / sanitized |
| --- | ---: | ---: |
| Original trio controls, 51 roots | 32 | 19008 / 19008 |
| Nexus controls, 129 programs / 254 roots | 166 | 120897 / 120897 |
| Supported core controls, 430 programs | 289, 171 suggestions | 162621 / 162621 |
| Original compiler, 77 roots | 4 | 7087 / 7087 |
| Original repository, 287 roots | 1 | 19144 / 19144 |
| Each later trio compiler, 77 roots | 0 | 4933 / 4933 |
| Each later trio repository, 287 roots | 0 | 18485 / 18485 |

Full finding/fix/suggestion bytes, ordering, spans, identifiers, explanations
and replacement text are compared against unchanged production Go. The original
control byte count changes only with the scratch path prefix on both sides.
Final normal and sanitized per-control result JSONs are identical. There are no
new failures or sanitizer findings; even the three parser refusals have no
sanitizer report. Both corpus manifests remain the original frozen populations.

Every semantic rule mutant compiles, exits 0 with empty stderr and is caught only
by complete byte comparison:

| Mutant | First differing byte |
| --- | ---: |
| unassigned | 1348 |
| caught | 4146 |
| exports | 14997 |
| process-state | 134 |
| race-handle | 118 |
| blocking-order | 1690 |
| namespace-alias | 72 |
| literal-parentheses | 587 |
| executor-span | 101 |

The export-chain-flags and export-module-lookup mutants also compile and exit 0,
with empty stderr; comparisons catch them at bytes 14715 and 17159. The native
released-handle check exits 70 with invalid or released checker handle; retaining
the released registry handle changes exit to 0 and is caught by its lifetime
assertion. The other unchanged bridge facts, seven foundation mutants and Node
compiler fixtures retain their prior current-main evidence in LANDING_REPORT.md;
they were not redundantly repeated for this metadata-only rule change.

## Commands and native time against Go

```
go test ./stage1/cohere/typeaware/wave_13_more -run '^TestWave13ListenerKind' -count=1 -v
ADAMIC_WAVE13_STAGE0=/workspace/wave13-landing-final-adamic ADAMIC_WAVE13_ARCHIVE=/workspace/wave13-landing-checker.a ADAMIC_WAVE13_ARTIFACTS=/workspace/wave13-speed-original ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus ADAMIC_WAVE13_COMPILER_MANIFEST=/workspace/wave-13-compiler.manifest ADAMIC_WAVE13_REPOSITORY_MANIFEST=/workspace/wave-13-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave13AgreementAndMutants$' -count=1 -v -timeout=30m
go vet ./stage1/cohere/typeaware/...
```

Rebuilt both continuation suite.a files with the current-main compiler and normal
or sanitized checker archives. Ran each directory's validate.py and corpora.py
on both modes, plus mutants.py using the unchanged production Go control bytes.
Existing reports document these positional arguments. All output went directly
to files. Reused the unchanged toolchain and current-main compiler; prior setup
reported Go/clang/Node/submodule 0s, cache warm 200s, total 200s; nproc is 5.
No new setup timing is claimed.

Three alternating full-output runs per corpus were performed after all builds
and tests finished. These medians include checker loading:

| Suite / population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original / compiler | 6.186187 | 1.915524 | 3.23x |
| original / repository | 0.427004 | 0.124587 | 3.43x |
| next / compiler | 2.969367 | 0.441660 | 6.72x |
| next / repository | 0.418892 | 0.128149 | 3.27x |
| more / compiler | 3.382823 | 0.462135 | 7.32x |
| more / repository | 0.480723 | 0.129480 | 3.71x |

Native remains slower on every population. These measurements establish the
current result, not an improvement caused by listener declarations. No timing
claim uses concurrent gate runs.

validation/speed-listeners retains normal/sanitized per-control results, corpus
hashes, semantic mutant logs, all nine metadata catches, isolated measurements
and a compressed archive of complete stdout/stderr and mutant evidence.
sha256.json checks every retained top-level artifact. Formatting and whitespace
checks pass. No protected compiler, shared parser/driver or registration file
changed; no new bridge question was needed. No new Adamic .ts file was added.
No new rules were claimed, no PR opened, and publication targets only the unit's
own codex/typeaware-wave-13 branch, never main or area/.
