# Lint wave 1, slot 02

Branch: `codex/lint-wave1-02`.

The helper report links the ordered list in `stage1/cohere/lint/HELPERS.md`.
Positions 4, 5 and 6 in that list are:

1. `@typescript-eslint/no-explicit-any`: skipped, already ported on
   `origin/codex/stage1-lint-batch5` (`dcef3eebfddcb297464bdce398d850775852587d`),
   in `stage1/cohere/lint/no_explicit_any.ts`.
2. `@typescript-eslint/no-inferrable-types`: skipped, already ported on
   `origin/codex/stage1-lint-batch4-typescript`
   (`63782c53678711717f9fd4f12762ce3d103b9a3d`),
   in `stage1/cohere/lint/no_inferrable_types.ts`.
3. `@typescript-eslint/no-restricted-types`: claimed by slot 02; no matching
   stage1 port filename found across fetched origin branches.

This claim is published before any rule implementation.

## Foundation blockers observed before implementation

Base: `origin/main` at `d090af531216ddd3c25a0dede6b82d7c0a6edf76`.
Registration: `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8`, merged successfully.
Helpers: `5d13f5baaecaf11d4ea62de693426f69a1f41bba`, attempted merge conflicts
in README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go
under stage1/cohere/lint. That merge was aborted, preserving both foundations.

The registration generator requires rule.ts and rejects mutant files without
a .ts suffix. This conflicts with the unit's requirement that new Adamic files
are .a and with the rule-directory-only ownership boundary. No shared generator
or dispatch file is changed by this unit. docs/parallel-work.md is absent on
main and on both requested foundation refs. Implementation is blocked pending
compatible foundations; this claim does not assert a completed port.

## Continuation claim, October 7

The user requested the next three unported-on-main and unclaimed rules on the
same branch. Existing work was pushed first, then all origin heads fetched with
--no-recurse-submodules. Main observed: ef3d907e.

The first 45 helper-ready entries are recorded in claims on origin branches.
The final helper entry is available. Continue in inventory order for the next
two syntax-only rules (needs_type_information false, binding_only false):

1. structure/tailwind-no-physical-direction
2. @eslint-community/eslint-comments/require-description
3. @next/next/google-font-display

None appears in claims on any fetched origin branch or in Adamic lint source on
origin/main. Metadata and frequency observations are not implementations.
This update is pushed before new code. New Adamic modules will be .a.
The original no-restricted-types reservation remains in place.

docs/parallel-work.md was found and read on origin/codex/no-shared-lists.
The existing .a registration limitation remains; any compatibility work used
for validation will stay in scratch overlays, not shared production files.

## Continuation result

Tailwind and require-description now have owned `.a` candidates and real-Go
comparison/mutant evidence. They require shared `.a` registration support.
Two upstream findings-plus-fixes cases remain excluded explicitly. Google
font display is blocked on shared JSX parsing and has a three-runtime refusal
probe under gaps/, not a registered placeholder. All three reservations remain
with slot 02; this is not a claim that three complete ports are certified.
See wave1-02-continuation-report.md for commits, commands, outputs and limits.

## Third reservation, October 7

Existing work through 339f88af was pushed before fetching all origin heads.
Main observed: ef3d907ecdc4c771b016f7d9c52372def057a340. The published
46-rule helper-ready list is exhausted by main ports and origin claims.
The first three available syntax-only inventory entries are now reserved:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-non-null-assertion
3. @typescript-eslint/no-this-alias

Selection examined 320 fetched origin refs and 39 claim Markdown files, using
whole public rule names, including names followed by sentence punctuation.
Inventory source: origin/codex/lint-inventory; needs_type_information and
binding_only are both false for each selection. None is ported on main or
named in any fetched origin claim. This update is pushed before new code.
Previous reservations and their explicitly reported integration gaps remain.

## Fourth reservation, October 7

All earlier implementation and gap evidence through 2f48129f was pushed first.
No-restricted-types is now ported and compared; Google font policy is ported
behind the explicitly blocked shared JSX adapter. See the owned rule report.
Fetched all origin heads. Main remains ef3d907ecdc4c771b016f7d9c52372def057a340.
The original 46 helper-ready entries are all on main or named in origin claims.
The first three remaining syntax-only entries in inventory order are reserved:

1. array-callback-return
2. arrow-body-style
3. base/consistency-no-bare-throw

Selection read 335 origin refs and 52 claim Markdown files, matching complete
public names including trailing sentence punctuation. All three have false
needs_type_information and binding_only inventory flags. No name appears in
main port source or any fetched origin claim. This claim is pushed before code.
Shared-harness gaps will be stated, with independent rule logic still ported.

## Fifth reservation, October 7

Earlier rule implementations, comparisons, mutants and explicit shared gaps
through 6de3b402 are pushed. All origin heads were fetched; selection examined
350 origin refs and 53 claim Markdown files. Main remains ef3d907e.
The original 46 helper-ready rules are exhausted. The next three syntax-only
inventory entries, neither ported on main nor named in origin claims, are:

1. dot-notation
2. grouped-accessor-pairs
3. id-length

Both needs_type_information and binding_only are false for each entry.
This reservation is pushed before implementation. Existing shared CFG, JSX,
parser and reporting limitations remain documented in owned rule reports.

## Fifth reservation result

Dot notation, accessor grouping and identifier length now have registered .a
candidates and independent Go comparisons on source Node, emitted JavaScript
and sanitized native, including semantic mutants. Accessor grouping covers all
160 captured cases. Dot and identifier runtime regex options remain explicitly
blocked; all remaining selected upstream cases and the full compiler/stage1
source corpus match. See rules/id-length/REPORT.md for exact coverage, excluded
inputs, throughput and raw logs. No further reservations are taken while those
two option surfaces remain pending. Only owned directories were changed.
