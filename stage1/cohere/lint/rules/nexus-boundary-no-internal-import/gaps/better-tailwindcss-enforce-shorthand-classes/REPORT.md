Built and pushed the prior claimed work, then three unregistered Tailwind decision kernels in .a; shared adapters still block complete ports.
Commits: prior 97b9922b; new claim 5675637e; shorthand 41843c9a; interpolation 1c87d678; conflicts 78b81442; certification is the subsequent report commit.
Checks: owned package PASS 18.606s, 562 cases and 121,355 identical Go/Node/emitted-JavaScript/sanitized-native bytes; vet and whitespace checks exit 0.
Mutants: invert already-present suppression, reverse seam fragment selection, invert composing-class exclusion; each compiles, exits 0 without stderr and is caught only by comparison on all three backends.
Not covered: registered rule integration, JSX/class-value extraction, design-system resolution, complete fixture/corpus findings and source ranges, whole-rule throughput, or the full repository gate.

## Selection and ownership

Previous claimed work was pushed through 97b9922b before taking this batch. That includes optional-chain assertions (1fe40525), unnecessary constraints (a93ba8ef), and the five earlier JSX decision candidates (6d7aded4). The earlier this-alias implementation remains e41b30da. Their exact certification and remaining adapter limits are in [the prior report](../typescript-no-non-null-asserted-optional-chain/REPORT.md).

After fetching origin again, selection examined 339 origin refs and 52 unique Markdown claim blobs, plus main's lint .a/.ts sources and descriptors. Main is ef3d907ecdc4c771b016f7d9c52372def057a340. All 46 helper-ready rules were unavailable. The first three remaining syntax-only inventory entries, in array order on origin/codex/lint-inventory, were enforce-shorthand-classes, no-concatenated-classes and no-conflicting-classes in better-tailwindcss. Claim 5675637e was pushed before code. Evidence/selection.json records the result. No further rule is claimed.

All new production language files are .a. Changes stay in these three owned directories and the previously authorized claim file. No shared generator, test harness, parser, compiler or Go cohere source is edited. The Go overlays are test-only; the working submodule remains pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db. No PR is opened.

## Observations and certification boundary

The shorthand kernel mechanically retains all 48 patterns in 18 families. It tries stable longest-first patterns and first-appearance variant groups, preserves first capture without backtracking, agrees on signs and important markers, selects trailing important spelling when present, and suppresses shorthand classes already written. The message is the pinned Go message. A real project must additionally ask its design system whether every generated shorthand exists.

The interpolation kernel distinguishes ASCII seam whitespace from Go strings.Fields Unicode whitespace, rejects empty seams, and names the last fragment before a hole or first fragment after it. It does not pretend that plus concatenation is template interpolation. The full class-template reader is external to this kernel.

The conflict kernel accepts already-resolved facts, deduplicates class names, compares variants and selector shapes, requires the complete property set, excludes composing utilities, reports symmetrically, and sorts properties in Go UTF-8 scalar order. It emits the pinned Go message. Despite commentary mentioning a suggestion, the actual pinned Go Run only reports a range; this candidate invents no edit.

The final cases comprise 394 shorthand lists, 34 interpolation segments and 134 resolved-fact lists. They include every table entry, signs/important placement, variants with nested colons, mismatched values, existing shorthands, duplicates, composed utilities, property-set overlap, selector/state mismatch, ASCII/Unicode whitespace, and supplementary Unicode property ordering. These are extracted decision inputs, not original whole-rule fixtures.

The oracle calls actual private Go shorthandCollapses, boundaryBeforeHole/boundaryAfterHole and message functions via an additional virtual package file. For conflicts, it injects resolved facts into the existing conflictFindingsIn resolution boundary. A byte-level guard proves that the committed Go overlay is the pinned original function with only that resolution block replaced; pairing and reporting code remain intact. The test also rejects Go pin drift. Source Node runs through oracle/node.mjs; emitted JavaScript uses the real lowerer/emitter; native uses ASan/UBSan. Compiler or sanitizer failures do not count as mutant kills.

The initial probe failed checked array accesses, then Adamic refused sort without a comparator. Both implementation issues were fixed. The original unrolled 561-case suite then passed in 255.903s. The final compact probe adds the Unicode ordering case and passes in 18.606s. Raw failure and passing logs are retained separately; failed compilation is never credited as parity or mutant proof.

## Exact shared blockers

The owned adapter-gap-probe.a parses `const view = <div className="w-4 h-4" />;` using the shared stage1 Parser. Source Node exits 70 with `adamic: panic: parser slice expected GreaterThanToken, got Identifier at 18 in tailwind-gap.tsx`. The shared parser has no JSX node construction. See evidence/jsx-gap.log.txt.

The helper branches' readClassValues/calleeValues/variableValues retain explicit external collection, traversal, payload and regex dependencies. They do not supply a production ClassLiteralReader/ClassTemplatesIn adapter over the stage1 AST with decoded text, inherited edges and exact source ranges. This also prevents full configured-callee and variable coverage, beyond the JSX failure.

RuleContext on this branch and the fetched codex/lint-harness-dot-a branch has no Program or design-system resolver. The latter branch fixes .a/suggestion/profile concerns, but not Tailwind entry points, stylesheet loading, CSS class facts, class membership or project skip/error behavior. No-conflicting-classes is silent in Go for nil Program; comparing empty raw findings would certify nothing. Its facts are explicitly supplied by the decision test, not generated by an Adamic engine.

Accordingly there are no rule.json registrations here. These are partial ports with independently tested decision logic. Complete TypeScript src/compiler, stage1 and original-rule-fixture findings/fixes cannot be certified until those adapters exist. Whole-rule findings/second for these three are unavailable for the same reason; decision cases are not represented as end-to-end performance measurements.

Previously completed suggestion rules did have measured whole-rule rates on the compiler plus 1,000 positive declarations, best of five, with process startup included:

| Completed rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| no-non-null-asserted-optional-chain | 547.46 | 880.69 | 4197.86 |
| no-unnecessary-type-constraint | 536.99 | 888.15 | 4067.75 |

Those rates and their corpus/fixture proofs belong to the prior linked report and are not performance claims for the new Tailwind kernels.

## Commands and evidence

Toolchain setup earlier in this continuation completed in the isolated published-harness worktree: go ready 0s, clang ready 3s, node ready 3s, submodules ready 22s, cache warm 427s, done in 427s on 5 processors (cgroup cpu.max 400000 100000), 17.6 GB. This batch reuses it; nproc again prints 5. Environment: /workspace/adamic-tools/env.sh. Earlier setup failures/workarounds are documented in the prior report, not hidden.

All test output goes directly to files. From the repository root after sourcing the environment:

    go test ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes -count=1 -v -timeout=15m > /tmp/wave12-tailwind-final.log 2>&1
    go vet ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes > /tmp/wave12-tailwind-vet-final.log 2>&1
    git diff --check > /tmp/wave12-tailwind-diff-final.log 2>&1
    node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes/adapter-gap-probe.a > /tmp/wave12-tailwind-jsx-gap.log 2>&1

The first three exit 0; the JSX refusal exits 70, as recorded. The full repository gate is not run. evidence/ contains unchanged copies of the logs. The owned test can be rerun without installing or modifying the shared harness.
