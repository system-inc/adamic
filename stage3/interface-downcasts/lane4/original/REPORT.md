Certified original Identifier.escapedText and Symbol.escapedName, with full inherited fields and full __String declaration.
Commits: first certified group on codex/views-mixed-unions-2, based directly on integration 4e67894a.
Checks: uncached TestCheckedViewBrandsOriginalPairs passed, 12 cases, 48.827s; Node, release C, sanitized C with leak checks, JavaScript.
Mutants: one helper member-check bypass per pair, caught by both backend refusal pins; all four mutant runs exit 0 with uncheckedunchecked instead of exit 70.
Uncovered: 28 of the original 30 candidate pairs / 95 of 543 reads; __String | undefined pairs tracked separately; whole-program reachability unmeasured.

The first group certifies two candidate pairs / 448 reads against all 78 emitted
declaration files from pristine official tsc 050880ce. Receiver fields are compared
with the stock checker manifest, so inherited fields cannot be reduced away.
The full brand includes string & phantom, void & phantom and the complete
InternalSymbolName enum. Built string, internal name and explicit undefined agree
with Node; wrong number and null pin exit 70 naming __String and the field.
Missing required properties pin initialization refusal. Checks run inside a helper
that accepts the original interface. No compiler hook was needed for this group.

Reproduce declarations with lane4b/original/prepare.cjs, then this directory's
prepare.cjs, using pristine pinned upstream. Set ADAMIC_BRAND_ORIGINAL_DECLS to
the emitted directory when running TestCheckedViewBrandsOriginalPairs.

Dates: __String working delivery target remains October 9, 2026, 17:00 MDT;
mixed primitives October 13, 2026, 17:00 MDT. These are estimates for certified
per-pair contracts, not whole-program tsc. Remaining groups are in progress.
Setup: Node .061s, Go .069s, submodules .143s, Markdown .170s, clang .443s,
build 95.977s, cache 96.519s, total 96.656s; nproc 5, quota 4 cores.
