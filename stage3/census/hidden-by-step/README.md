# Hidden bytes attributed to roadmap steps

This unit attributes the supplied ranking, without rerunning or changing its compiler
measurement. [REPORT.md](REPORT.md) ranks the totals, gives each group's top three
exact reasons, and lists every unplaced reason, including zero-byte reasons.
[RESULT.json](RESULT.json) is the machine-readable result;
[mapping.json](mapping.json) assigns each exact `(kind, reason)` once to a step or
null, with the applied rule. [input.json](input.json) retains every credited reason
row, examples and provenance as a projection of the original JSON.

Input: `ec0b16c04f3bbdfbaf3932ab01307f90b08156fd`,
`stage3/census/hidden-ranking/RESULT.json`, measured compiler
`69501280a81259fb512edbb8dd0e52c6eb0d88c8`. The projection records the SHA-256
of the entire original JSON. The tool parses that entire input, including its
boundaries and attributed segments, but sums only already credited reason rows.
It never adds raw boundary lengths, which would count nested bytes again.

There are 1,712 Refused/NotYet reasons and 15 other-cause rows. Every one of
3,654,880 hidden bytes is conserved: 1,155,139 bytes in 425 placed compiler
reasons, 794,395 bytes in 1,287 unplaced compiler reasons, and 1,705,346 bytes
in 15 unplaced measurement causes. The compiler-reason subtotal remains
1,949,534 bytes. Unplaced is an accounting group, not roadmap step 30.

## Placement rule

Placement is an inference from the diagnostic's operation and type spelling,
not an observation that implementing that step would fix the program. The
ordered rules below are implemented literally by the regular expressions in
[attribute.py](attribute.py). First matching rule wins; the row records its
rule name. This gives each reason at most one step even when several features
appear in its text. There is no family normalization of the input identities.

| Order | Rule | Step | Evidence used |
|---:|---|---:|---|
| 1 | object-relation | 22 | A value seen as another type, or a relation explicitly failing optional field presence/type |
| 2 | generic-mapper | 16 | Generic function/mapper failure, or an overload with additional implementation type parameters |
| 3 | overload | 30 | Other overload compatibility failures and overloaded function values |
| 4 | object-operation | 22 | Structural method dispatch with statics, member assignment, Object methods, computed fields, destructured methods, method replacement, static initialization escapes, property presence, delete, object spread and prototype/receiver views |
| 5 | iteration | 20 | The failing operation is for-of, for-in or iteration, including object/union iteration |
| 6 | optional-call | 18 | The failing operation explicitly says optional call; optional property/index chains do not satisfy this rule |
| 7 | collection-storage | 19 | A Map or Set storage/key failure, or construction of Map from entries; collection storage wins over brands, type parameters and unions appearing inside it |
| 8 | regex-operation | 23 | The failing operation is RegExp construction |
| 9 | exception-operation | 21 | The failing operation begins throw, try, catch or finally |
| 10 | namespace-operation | 15 | The failing operation begins namespace |
| 11 | generic-type | 16 | A value, function/call result or array element diagnostic names a standalone T/U/V/K/R/X/Y or a T followed by an uppercase letter; this is a spelling-based inference, not checker resolution |
| 12 | known-brand | 11 | Direct value/result/array element types __String, Path, ResolvedConfigFileName, ResolvedConfigFilePath or PathPathComponents, optionally unioned with undefined |
| 13 | union-storage | 17 | Explicit union in value/result/element/checked-view-field type, boxed/narrowed/captured union failure, differently held union binary operands or null/scalar comparison |
| 14 | object-view | 22 | Erased function/object descriptors or JSON object reflection metadata |
| 15 | other-explicit-operation | 30 | Concrete non-wall failures: unchecked casts, var, destructuring, rest/spread, templates, Date/JSON parsing, class expressions, this, nested functions, yield, arithmetic/operators, assignment expression returns, array method limitations or typed arrays |
| Otherwise | insufficient-diagnostic | null | The message does not identify one of these operations/types sufficiently |

Checker rejection, skipped dependency, panic and competing-cause conflict rows
are always null (`measurement-cause`). A conflict cannot be assigned the bytes
of both competing steps. A checker-rejected body can contain many features;
its whole span cannot be credited to one without additional evidence.

The five known brand aliases were checked against upstream TypeScript 6.0.3
commit `050880ce59e30b356b686bd3144efe24f875ebc8`: compiler/types.ts:23
(Path), :4557 (ResolvedConfigFileName), :6196 (__String),
compiler/tsbuildPublic.ts:186 (ResolvedConfigFilePath), and
compiler/path.ts:457 (PathPathComponents). An interface that happens to inherit
a brand is not automatically a branded-type lowering failure. Likewise
`reading excludeRegex`, `reading currentNamespace` and `reading exception`
identify properties in tsc's program, not regex, namespace or exception failures.
They stay unplaced. Opaque aliases such as CapturedThis, InitializedVariableDeclaration
and ImmediatelyInvokedArrowFunction also stay unplaced; resolving their layouts
and pinpointing the failed lowering operation is outside this ranking-only unit.
Zero bytes for steps 15 and 21 mean no matching reason, not that those steps work.

Step 30 contains only explicit operations outside the named wall steps, not a
fallback for ambiguous messages. Refused entries still describe current refusal;
putting them in a roadmap bucket does not promise their future admission.
All credit inherits the ranking's outermost-cause estimate: these are potentially
exposed source bytes, not proven bytes that would successfully compile if fixed.

## Fixture and validation

[fixture.json](fixture.json) is a three-reason synthetic ranking with a
hand-written answer: Object.entries 60 plus Object.defineProperty 40 makes
step 22 total 100, above namespace's step 15 total 90. This proves aggregation
before ranking, exact top-reason ordering, and placement. It is JSON input to a
reporting tool, not an Adamic program; there are no new .a files or compiler
fixture registrations. [counts.md](counts.md) records its local checks.

The mutant appends an additional assignment of Object.entries to step 30,
retaining its step 22 assignment. The validator rejects it with
`reason mapped more than once`, before aggregating or crediting any bytes.
The same known answer catches a +1 credited-byte mutant. Seven further mutants
exercise duplicate input, missing/extra coverage, negative/boolean/string byte
counts and an unknown step. All nine are caught; the proof tests exit zero
only after the expected failures occur. Six focused tests pass, including the
full pinned-ledger recomputation and direct per-group sums/top-three checks.

Run from the repository root (outputs must go to logs):

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/hidden-by-step/attribute.py > /tmp/step05-attribute.log 2>&1
python3 stage3/census/hidden-by-step/test_attribute.py > /tmp/step05-tests.log 2>&1
git diff --check > /tmp/step05-diff.log 2>&1
```

Regeneration needs the pinned ranking commit fetched. Tests need only the local
projection and mapping. Reporting regenerates deterministically, sorting groups
by descending byte count then step number (null last on ties), and reasons by
descending bytes then kind and exact text. Focused logs and setup output are
preserved under evidence/ as text. No whole-package tests or full gate were run.

Setup used `GOPROXY='https://proxy.golang.org|direct'`. Timing lines: Node
0.093s, Go 0.143s, submodules 0.328s, markdown 0.704s (validation 0.042s),
clang 1.200s, Go build 184.390s, test binaries deferred 184.660s, cache warm
184.663s, done 184.722s. `nproc` is 5; cgroup CPU quota is 4.
Base main is `031a1259bc7973934792dc6cb1bd4074fc2204b9`.

## Limits

I did not infer a step for the 1,287 ambiguous compiler reasons or the 15
measurement causes, rerun the source census, or test fix-alone compiler
counterfactuals. The ranking does not provide enough evidence for those
attributions. Exact unplaced identities and credits remain visible in the
report and JSON so a later source/checker audit can refine them without losing
or double-counting bytes. This unit changes only hidden-by-step reporting files.
