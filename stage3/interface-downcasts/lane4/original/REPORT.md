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

Second group: four further original pairs / 48 candidate reads, bringing totals
to six / 496 certified and twenty-four / 47 remaining in the 30 / 543 table.
Uncached focused oracle passed in 63.094s, 36 cases. PrivateIdentifier and
TransientSymbol preserve their full upstream field sets. MemberName and the
explicit Identifier | PrivateIdentifier type receive checked member downcasts
via normal upcasts; both alternatives are exercised. The initial fixture attempted
a broad Base-to-union cast, which correctly refused at compile time. It was
corrected to the approved cast path; no admission rule was weakened.
Six helper-check removal mutants (one for each union alternative and one each
for the two single receivers) were caught in both C and JavaScript by exit-70
pins, rather than compiler warnings or sanitizer failures. Raw group logs retained.

Third group: all five original __String | undefined field pairs / nine reads
certified separately from the thirty-pair table: UnionType.keyPropertyName, both
SourceFile local JSX names, SymbolLinks.typeOnlyExportStarName, and
WideningContext.propertyName. Uncached oracle passed in 62.461s, thirty cases.
Original receiver objects are reached through a viewed carrier then passed into
an interface helper. Missing optional properties and explicit undefined agree
with Node; number and null pin named exit-70 refusals. Five helper member-check
bypass mutants run exit 0 instead of the refusal, caught in both backends.
Full original fields are checked against the expanded upstream manifest.
Combined delivered coverage is eleven pairs / 505 candidate reads. The original
__String table remains six / 496 certified and twenty-four / 47 pending.

Fourth group: eleven original namespace field pairs / twenty-three reads certified,
bringing the thirty-pair table to seventeen / 519 certified, thirteen / 24 pending.
Including the optional-branded group, twenty-two pairs / 528 reads are certified.
Uncached oracle passed in 66.045s, sixty-six cases. Eleven bypass mutants are
caught in both C and JavaScript. No namespace or static receiver is trusted in
the fixtures: a viewed carrier holds the namespace-shaped value and passes it
into a helper whose static type is the complete original typeof namespace.
Private JsxNames and ReactNames are omitted by ordinary exported declarations.
The adapter derives declaration-only namespaces from pristine checker.ts AST,
preserving every original export const and its exact __String cast annotation.
It rejects any changed member form, retains all ten JsxNames fields and React's
Fragment, imports the full upstream __String declaration, and hashes both source
and generated declarations. It does not compile tsc namespace implementations
or weaken Adamic's namespace syntax refusal. Runtime values come from fixtures.

Fifth group: original __String[] dynamic element pair / six reads certified;
plain table now eighteen / 525 certified and twelve / 18 pending. Optional fields
remain five / nine certified; combined twenty-three / 534 delivered.
Uncached oracle passed in 5.769s. Built strings, internal names, undefined and
bounds agree with Node in both backends and sanitized/leak-checked C. Number
elements pin exit 70: element read failed: values[index] expected __String, found
number. The element-check bypass runs valid C/JS release code and is caught by
that pin. The complete original element descriptor is a ViewUnion represented
as String, with the void phantom admitted as undefined; its primitive members
are verified. An initial scalar-only oracle assertion was corrected after
inspecting that actual descriptor. No compiler admission was changed.
Null-only arrays remain a named compile refusal, not runtime coverage; that
supplementary producer gap is retained. No broader primitive array adapter or
null-array producer support is claimed.
