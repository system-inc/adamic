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

Sixth group: five original generated/unique-symbol/nullable-receiver pairs / nine
reads certified. Plain table twenty-three / 534 certified, seven / nine pending.
With nullable fields, twenty-eight pairs / 543 reads certified. Uncached oracle
passed in 57.134s, thirty cases. Both generated identifier declarations keep
their unread EmitNode intersection obligations. Nullable Identifier and Symbol
receivers are read with original optional-chain expressions in interface helpers.
Five read bypass mutants are caught by named exit-70 pins in both backends.
The adapter also verifies three repeated receiver ids at their exact upstream
AST read positions: the full field sets match the original named interfaces and
the declared field type is __String. They are not credited until their own
source fixtures and mutants run in the next group.

Seventh group: four original repeated/mapped receiver pairs / six reads certified.
Plain __String table: 27 pairs / 540 reads certified, 3 pairs / 3 reads remaining.
Optional __String | undefined table: 5 pairs / 9 reads certified, none remaining.
Combined: 32 / 549 certified of 35 / 552 candidates. Focused green oracle
passed in 53.955s; four bypass mutants were caught in both backends. Each repeated
receiver id has its own source fixture group, checked against the exact original
AST receiver's full field set and declared __String type. Mutable<Identifier>
imports the complete original utilities.d.ts mapped type, with all fields intact.

Observed blockers, retained as eighteen Node controls with named compile refusals
(TestCheckedViewBrandsOriginalBlockedPairs, PASS 10.564s):

| Original pair | Reads | Refusal |
| --- | ---: | --- |
| ActiveLabel.name, id 103520 | 1 | field name with unsupported never contract |
| { name: __String; oldSymbol: Symbol; }.name, id 70885 | 1 | field name with unsupported never contract |
| (LeftHandSideExpression & Identifier).escapedText, id 9477 | 1 | field child with unsupported recursive intersection payload contract |

The private ActiveLabel declaration preserves every original binder.ts member;
the anonymous type is emitted from the stock checker's exact original read
receiver. Node prints word-built for the valid fixtures, but lowering refuses
before either backend executes. None of these three is credited complete.
Code inspection indicates name's supported descriptor is contaminated by the
field-name-only fallback over unrelated reachable descriptors. This is an
inference about the cause; the observed evidence is the named refusal. A fix
must scope obligations using the existing shared allocation flow and retain
refusals through broader interfaces, helpers and Unknown flow. Disabling the
fallback would discard its soundness guard and is not a valid fix.
The recursive intersection adapter is a separate blocker. Newest integration
322bd65d was inspected and retains both behaviors; no individual lane was merged.
The old reduced intersection fixture completion is explicitly superseded in the
lazy ledger: that table now has one brand pair / one read pending.

Dates: 27 original brand pairs and all five optional fields are delivered now.
October 9, 2026, 17:00 MDT remains the full __String working target, conditional
on resolving these three shared admission blockers. October 13, 2026, 17:00 MDT
remains the mixed primitive working target. In the original mixed column, the
32 certified brand pairs cover 549 of 690 candidate reads; 31 pairs / 141 reads
remain, including the three blocked brands. These are candidate counts, not
checker-clean whole-program reachability or whole-tsc compilation.
