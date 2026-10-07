# Type-aware lint wave 08

Branch: codex/typeaware-wave-08.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Positions 22, 23 and 24 in the combined by-volume ranking referenced by
VOLUME_REPORT.md, excluding the 26 rules already ported on the base:

22. no-loop-func (compiler 9, repository 2, combined 11)
23. nexus/correctness-no-import-cycle-load-time-read (compiler 10, repository 0, combined 10)
24. @typescript-eslint/no-require-imports (compiler 9, repository 0, combined 9)

All fetched origin stage1 trees were checked for corresponding implementation
filenames and claim Markdown containing these names. No match was found.
No rule is skipped. This claim is committed and pushed before implementation.
New Adamic files use .a.

## Continuation claim

The original three rules are complete and pushed in d0d6310f. After fetching all
origin heads, checked all 33 claim documents across 325 refs and native rule
registrations on origin/main and origin/codex/tsgo-c-library. The combined
compiler/repository ranking has 197 checker-dependent entries; 25 of the base
ports appear in that ranking and 96 entries are named in origin claim files.

The first three neither ported on those two branches nor claimed on any origin
branch are, with lexical ordering for equal combined volume:

- nexus/correctness-no-process-exit-after-output (0)
- nexus/correctness-no-uncleared-race-timeout (0)
- nexus/correctness-require-blocking-standard-streams (0)

Claimed for wave 08 continuation. This update is committed and pushed before
implementation. New Adamic files remain .a; existing shared harness and generator
files will not be edited.

Continuation status: all three ports are complete in the owned native profile.
Normal and ASan/UBSan output matches unmodified Go cohere on 100 upstream process
programs, compiler 77, repository 287 and race controls 22. Each rule has a
compiling byte-only mutant; raw-fact, state, released-handle and full bridge
ownership/sanitizer gates pass. Shared registration/profile integration is left
to the assigned harness worker; an owned scratch overlay links the new raw
questions without editing shared files. See ../wave08-next/REPORT.md.

## Second continuation claim

All six earlier claims are complete and pushed through cfa7115d and a116e599.
Fetched all origin heads again and checked all claim Markdown blobs across 356
origin refs, plus full and shorthand native rule registrations on origin/main
and origin/codex/tsgo-c-library. Of 197 ranked checker-dependent rules, 25 are
already ported and 136 are named in claims; 36 remain eligible.

The first three eligible rules in the combined by-volume ranking, with lexical
ordering for tied zero totals, are:

- react-hooks/globals (0)
- react-hooks/immutability (0)
- react-hooks/no-deriving-state-in-effects (0)

Reserved for this branch. This claim update is committed and pushed before
implementation. Shared harness, generator and dispatcher files stay untouched;
new native implementation files will be .a.

Second continuation status: PARKED per Ahra. Globals
has native AST/scope decisions and exact normal/sanitized parity on 35 non-JSX
controls and both corpora, but the shared parser rejects JSX. Immutability and
no-deriving-state-in-effects have tested native kernels; their production entry
points refuse because native React HIR lowering/SSA and the associated passes
are unavailable. Kernel mutants are not complete rule mutants. Shared files
were not edited. These parked reservations count as finished for the landing cap.
Native React analysis is being ported on #dnv6f2c; JSX is landing on
area/stage1-lint. No push targets those integration branches.
See ../wave08-react/REPORT.md.

## Third continuation claim

All earlier work is pushed and landing-ready on main f8013f0ba, with the six
complete ports and the supported React partials re-green in 443a69d33. React
reservations are parked as directed above. Fetched all 529 origin refs, inspected
33 distinct claim Markdown blobs, and checked main/bridge registrations against
the 197-entry combined volume ranking. React-associated rules, including
structure/react-hook-no-any-type, are excluded under the React parking directive.

The first three eligible non-React rules are:

- require-atomic-updates (0)
- require-await (0)
- symbol-description (0)

These are reserved for wave 08. Core atomic updates uses the existing ECMAScript
CFG family rather than React HIR/SSA. This claim update is pushed before code.
New native sources use .a, declare numeric listeners and rule.json kinds, and
will receive nodes through an owned indexed profile while the shared driver is
pending. Shared source remains untouched. Frozen selection is in
../wave08-core-next/selection.json.gz.

Third continuation status: COMPLETE in owned native profiles. All three rules
have full finding/fix/suggestion byte parity on captured upstream programs and
both frozen corpora, normal/sanitized, with one compiling byte-only mutant per
rule. Numeric listeners, manifests, released handles and missing-metadata refusal
are checked. The branch is rebased onto main c01907a70, and earlier six ports are
re-green. Shared production factory/dispatcher integration is still owned by the
harness worker; no shared source was edited. See ../wave08-core-next/REPORT.md.

## Harness landing

Rebased onto origin/area/stage1-lint at 7481e0324, including the merged harness
and JSX parser. Globals is now COMPLETE in the owned native profile: all 111
typed upstream programs (58 findings), controls and both frozen corpora match
Go normally and under sanitizers. Its compiling byte-only mutant is caught in
seven upstream programs. The checker-free upstream witness retains its original
Go assertion. The old JSX refusal expectation is replaced by exact parity.

Immutability and no-deriving-state-in-effects remain PARKED for native React
HIR/SSA/capture support on #dnv6f2c. Shared production factory/dispatcher and
checker-program integration remain pending; owned profiles do not claim shared
registration certification. No additional rules are reserved: a fresh audit of
592 origin refs finds every ranked rule ported or named in an existing claim.

Runtime landing: rebased onto origin/area/stage1-lint at d65a8f931, including
the native release and string runtime changes. All ten completed owned profiles
and both parked kernels are re-green, with byte-only mutants and sanitizer/
released-handle checks. A fresh 606-ref audit still finds no unclaimed rule.
See ../wave08-core-next/LANDING_D65_REPORT.md for the current evidence and times.

Current-main landing: rebased onto area/stage1-lint b84a9d931, containing main
c7991b900 and its lowering/record runtime changes. All ten completed profiles
and the two parked kernels pass again, with all rule mutants and released-handle
checks. A 631-ref audit has no unclaimed entry. No selected check skipped.
See ../wave08-core-next/LANDING_B84_REPORT.md.

Registry migration landing: rebased onto area/stage1-lint b46914832, still
containing main c7991b900. All ten completed profiles, parked kernels, byte-only
mutants and sanitizer/released-handle checks pass again. Shared registry and
complete suggestion serialization tests pass. A 639-ref audit finds no unclaimed
rule. No selected Go check skipped; the full gate was not run.
See ../wave08-core-next/LANDING_B469_REPORT.md.
