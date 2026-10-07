# Lint wave 1 slot 10

Branch: codex/lint-wave1-10. Owner: wave1-10.

The helper report delegates its ordered rule list to stage1/cohere/lint/HELPERS.md.
Positions 28, 29 and 30 in that list are claimed here:

28. react/no-unsafe
29. react/self-closing-comp
30. sort-vars

No implementation of these three names was found in main or the fetched origin
branch source trees on October 6, 2026. Helper descriptors and inventory rows
are readiness metadata, not ports. No rule is skipped as already ported.

This claim is committed and pushed before any rule implementation.
New Adamic source files for this unit must use .a, never .ts.

Foundation integration: the registration merge was clean. The helpers merge
conflicted in six shared lint driver files. The registration versions were
retained, with the standalone helper package and inventory added by the merge.

## Continuation claim, October 7

Fetched all origin heads without recursive submodule fetching. origin/main is
 ef3d907ecdc4c771b016f7d9c52372def057a340. Previous branch work is fully pushed.

The helper report points to HELPERS.md. Its measured comment handoff now expands
the original 46 to 62 helper-ready rules; option-gap bullets are not ready rules.
After excluding ports on main and claims on every fetched origin branch, the
first three candidates in measured handoff order are reserved by this unit:

- structure/tailwind-no-physical-direction (original handoff position 46)
- @typescript-eslint/ban-tslint-comment (comment handoff position 1)
- @typescript-eslint/no-invalid-this (comment handoff position 2)

The latest selection criterion excludes ports on main, rather than all branch
ports. ban-tslint-comment has older origin-branch implementations outside main;
these are reference material, not a main port or a claim in claims/.
No new implementation is written before this claim is committed and pushed.
The original three reservations remain as recorded above.

## Third claim, October 7

All previously claimed rule implementations and their bounded validation are pushed through 68cf11aa. Self-closing-comp remains blocked on independent JSX parsing, with all 79 Go-AST rule-engine cases and its mutant held across Node, emitted JS and sanitized native.

Fetched all origin heads again and checked main ef3d907ecdc4c771b016f7d9c52372def057a340 and every origin branch claims directory. The first three unclaimed, unported helper-ready rules in the measured handoff are:

- arrow-body-style
- max-lines
- no-extra-bind

This update is committed and pushed before writing implementation code. No inventory fallback was needed. The shared harness dependency is origin/codex/lint-harness-dot-a at 2650ad595b82220c368631ea13139fad4b306ed6; no shared files are edited here.

## Fourth claim, October 7

Previous implementations, tests, reports and explicit shared-harness gaps are fully
pushed through 39d19398. Fetched all 348 origin refs and checked actual claim
documents under every claims directory against main
 ef3d907ecdc4c771b016f7d9c52372def057a340.

The first three available rules in the helper report's measured handoff are:

- default-case
- no-extra-label
- no-fallthrough

Selection JSON inventories under another unit's evidence directory mention these
names while explicitly recording claims: []. Such inventory rows are not claims.
They caused the previous scan to skip default-case incorrectly; it is the first
unreserved helper-ready name and is reserved now. No syntax-only fallback was needed.

This claim is committed and pushed before writing any implementation. Shared
multi-edit fixes and JSX parsing remain blocked as previously reported; the latest
harness f4d98cab still rejects multiple fixes. Only owned rule directories will be
changed after this claim.
