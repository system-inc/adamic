# Type-aware lint wave 14

Branch: `codex/typeaware-wave-14`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 40, 41 and 42 after excluding the 26 existing ports:

1. `@typescript-eslint/no-array-delete`
2. `@typescript-eslint/no-base-to-string`
3. `@typescript-eslint/no-extraneous-class`

Selection uses the combined compiler and repository checker-dependent counts
linked by VOLUME_REPORT.md in validation-volume, descending total with lexical
rule-name ties. Each selected rule has compiler 1, repository 0, total 1.
All fetched origin branch trees were checked for matching claims and rule-named
port files before this claim. No selected rule was found claimed or ported.
No rules were skipped.

## Continuation on October 7, 2026

After pushing c043d14a, fetched all 320 origin refs without recursing submodules.
The first three remaining checker-dependent rules in VOLUME_REPORT.md's combined
compiler and repository counts, descending volume with lexical ties, are:

1. `nexus/correctness-no-global-listener-target-assertion`
2. `nexus/correctness-no-leaked-number-render`
3. `nexus/correctness-no-mock-on-module-namespace`

Each has zero compiler and repository findings. Excluded the original 26 ports
on origin/main and origin/codex/tsgo-c-library and every rule named in Markdown
under stage1/cohere/typeaware/claims on every fetched origin branch, including
incomplete reservations. The earlier collection and discarded-result candidates
are already claimed by other continuations. These three have no matching native
port on either base branch and no matching claim. This update must be pushed
before implementation. Positive controls and comparison-only mutants are required.

Continuation status after Ahra's correction: the listener assertion and module
namespace rules are implemented and validated. The leaked-number-render rule
remains unimplemented because the native parser cannot parse JSX child nodes.
A Go-positive .a witness, parsed as virtual TSX by the independent Go harness,
reports bytes 49:54; the native parser refuses with panic 70 at byte 54.
This is a partial continuation, not three completed ports. Implementation stops
at this blocker, with no further rules claimed and no shared harness or generator
edits. The six existing checker-dispatch lines were added before the correction.
See ../WAVE_14_NEXT_REPORT.md for commands, artifacts and limits.

## Leaked-render judgments completed

On the next resume, the leaked-render rule's native judgments and JSX-context
visitor were implemented in no_leaked_number_render.a and integrated into the
owned wave_14_next.a driver. Twenty-two isolated controls supply only synthetic
JSX parent contexts to both native and the unchanged production Go listener.
They match on 18 complete findings and 7704 bytes, including empty fixes and
suggestions; attributes and ordinary source stay silent. Sanitizers pass and
an exit-0, empty-stderr judgment mutant differs at byte 51.
Real JSX source remains blocked by the shared native parser, independently of
.a module support in the shared lint harness. No shared parser, generator or
harness file was changed. This is not a claim of end-to-end JSX parity.
No additional rules are claimed while that original completion gate is blocked.
See ../WAVE_14_RENDER_REPORT.md for the evidence and exact scope.

## Third batch, October 7, 2026

The preceding six rule implementations and their available comparisons are pushed
at 5bd31ee5. Real JSX remains a shared-parser gap, documented in
WAVE_14_RENDER_REPORT.md. The latest instruction permits moving on after pushing
the available implementation and reporting that gap.

Fetched all 335 origin refs. After excluding the 26 ports on origin/main and
origin/codex/tsgo-c-library and every reservation under typeaware/claims on every
origin branch, the first three remaining rules in the combined by-volume ranking
are claimed here:

1. `no-invalid-regexp`
2. `no-label-var`
3. `no-misleading-character-class`

All three have zero compiler and repository findings. Neither base branch has
rule-named native files for them. There are 58 remaining unclaimed ranking rows
before this reservation. This claim is pushed before implementation. New native
files use .a; shared parser, registration generator and harness remain untouched.

Third-batch status: native scoped-label judgments, constructor selection and
flag validation, and regex-literal class judgments are implemented. Fifty-one
supported controls produce 49 complete byte-identical findings under all three
sanitizers, with a mutant per rule plus scope-meaning and Unicode-quote mutants.
Both frozen corpora also agree. This remains partial: undefined labels are a
shared parser gap; string-pattern validation, constructor reference/source
mapping and lone-surrogate flags are explicit refused dependency paths. These
are demonstrated by independent Go-positive witnesses. No further batch is
claimed. See ../WAVE_14_THIRD_REPORT.md.

## Surrogate flags completed on resume

Fetched origin and removed the lone-surrogate flags refusal in the owned native
rule. Three additional .a controls bring the supported comparison to 54 inputs,
60 findings and 19,352 identical bytes in normal and sanitizer builds. Six
compiled comparison-only mutants and the released-handle mutant are caught.
Both frozen corpora still agree. The rule suite passes in 99.844 s.
Undefined labels, native pattern validation and constructor tracking/source
mapping remain explicit Go-positive boundaries; real JSX remains blocked in the
preceding batch. No further rules are claimed while these incomplete paths
remain. See ../WAVE_14_SURROGATE_REPORT.md for the current commands, evidence and
native-versus-Go timing.
