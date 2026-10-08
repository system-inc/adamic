Built: declaration-backed witnesses for SourceFile.externalModuleIndicator and Diagnostic.messageText, plus correct optional absence handling.
Commits: builds on f4e5993e and integration ba59427c; implementation tip is the commit containing this report.
Commands/results: selected production gate passed, oracle 49.176s; all nine original cases passed; vet and diff check passed.
Mutants: outer removal, wrong-member/shape acceptance, transitive removal, readiness loss and rejecting allowed absence caught in both release backends.
Uncovered: 40 candidate pairs / 146 candidate reads remain; exact whole-tsc execution reachability remains unmeasured.

Two original field-contract pairs are now certified by standalone source witnesses:
SourceFile.externalModuleIndicator (20 candidate reads) and Diagnostic.messageText
(15 candidate reads). The original target and object-member declarations are
imported in full, including inheritance, qualifiers, recursive references and
unsupported unread descendants. No reduced Node or Chain declaration is used.
This certification covers the original field contracts with representative source
reads and propagation controls. It does not mean the 35 original enclosing tsc
functions were compiled or executed. The 35 sites were audited byte for byte at
their UTF-16 positions against the pinned upstream tree. Exact production
reachability still waits for a checker-clean compiler.

The TypeScript tree is fetched from microsoft/TypeScript at the repository's
stage3/source.json pin 050880ce59e30b356b686bd3144efe24f875ebc8. No code is copied
from cohere. prepare.cjs rejects tracked source modifications, verifies all
recorded sites, and emits 78 declaration files outside Adamic. Its manifest hashes
all emitted files and records the exact complete field sets for SourceFile, Node,
Diagnostic and DiagnosticMessageChain. The oracle checks those hashes and rejects
missing/reduced field sets in the lowered contracts. Declaration emission is not
whole-program source compilation; each bound witness and all its declarations
pass ordinary Adamic load diagnostics. No diagnostics are disabled in that path.

Nine source cases cover both valid union alternatives, false outside the true
literal contract, invalid Diagnostic boolean, wrong nested SyntaxKind/code,
explicit undefined, absent optional property, uninitialized present property,
and SourceFile generic helpers/callbacks/stored aliases. Both backends pin field,
expected, found and exit 70. Native positives include ASan/UBSan and leak checks.

Original SourceFile declares externalModuleIndicator optional. The prior reduced
control declared it present with an undefined alternative, so it did not expose
missing optional storage. The owned C/JS emitters now pass field absence and
receiver optionality separately. The native helper allows a missing optional slot
before inspecting payload; a present uninitialized slot still enters the common
readiness refusal. No shared file changes were necessary in this group.

Independent release mutant observations (native / JavaScript):

- SourceFile wrong-member acceptance: false / false, exit 0.
- SourceFile outer-check removal: absent / false, exit 0; helper/callback variant
  has the same independent result.
- SourceFile wrong nested kind acceptance and transitive removal: 0 / false,
  exit 0, independently failing the SyntaxKind refusal pin.
- Diagnostic wrong-member acceptance and outer removal: wrong / wrong, exit 0.
- Diagnostic wrong nested code acceptance and transitive removal: chain:0 /
  chain:false, exit 0, independently failing the number refusal pin.
- Lose readiness on the present optional slot: absent / absent, exit 0, failing
  the named uninitialized refusal pin.
- Reject allowed absence: both stop at exit 70 after true/80/absent; the positive
  four-value Node pin catches this erroneous refusal.

Validation (output retained in logs):

```
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/object-primitive-upstream
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
node stage3/adapt/00-setup/adapt.cjs /tmp/object-primitive-upstream
node stage3/interface-downcasts/lane4b/original/prepare.cjs /tmp/object-primitive-upstream /tmp/object-primitive-original-declarations
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/object-primitive-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitive|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestCheckedViewArrays' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
git diff --check
```

The original suite is opt-in to its generated external declarations; it ran with
that environment set and no original cases skipped. Native and JS packages had
no matching package-local tests; their emitters ran through the oracle. Opt-in
whole-program lazy census tests skipped. This is a selected gate, not the full
repository gate. Existing reduced and array controls passed alongside originals.

The next original pair in the inventory is the overlapping object/intersection
family (11 reads), followed by LiteralType.value (10). Neither is certified here.
The working whole-family target remains October 13, 2026, 23:00 UTC. Candidate
coverage is tracked in certification.json and the ranked lazy-pair-progress.json;
completed_reads denotes candidate reads attached to witnessed pairs, never
measured whole-program runtime reads.

October 8, next original-pair checkpoint

Built: full-declaration LiteralType.value witnesses and an explicit bindable intersection frontier.
Commits: builds on d70bc4dd and integration ba59427c; tip is the commit containing this checkpoint.
Commands/results: selected original, reduced, array, lazy and intersection gate plus go vet and diff check; timings recorded in logs/next-production-gate.log.
Mutants: wrong scalar acceptance, skipped outer read check, wrong nested boolean acceptance and dropped transitive check caught in both release backends.
Uncovered: 39 candidate pairs / 136 candidate reads; bindable expression's 11 reads remain included; exact whole-tsc reachability unmeasured.

LiteralType.value (10 attached candidate reads) now joins the two certified
standalone original field contracts. The complete upstream LiteralType and
PseudoBigInt declarations are imported, inherited fields and mutable qualifiers
included. Positive output is plain/42/true:123. Boolean at value pins the named
string | number | PseudoBigInt refusal; nested negative and base10Value failures
pin expected boolean/string and found number, exit 70 in native and JavaScript.
The type does not allow absence, so there is no allowed-absence fixture.
All ten original site texts are audited against the pinned upstream source.
These are representative original field-contract witnesses, not executions of
the original enclosing compiler functions.

Release mutant outputs (native / JS), each exit 0 and caught by the refusal pin:
accept boolean at value: wrong / wrong; skip outer check: wrong / wrong;
accept wrong nested negative: true:123 / 42:123; drop its transitive check:
false:123 / 42:123. Existing SourceFile and Diagnostic mutants remain enabled.

The higher-ranked BindableStaticAccessExpression.expression (11 candidate reads)
is audited and explicitly remains uncertified. A cast-only original parent
witness prints admitted in Node and both backends. Its demanded expression read
refuses with the pinned unsupported representation conversion contract message.
A minimal named hook in lower/library_object.go calls the owned
objectPrimitiveIntersectionStorage helper to recognize a physical object slot
for plain structural union/intersection members with aggregate fields. This
does not certify the intersection. Existing lazy contract refusal is unchanged,
and existing intersection checks run in the selected gate. No emitter or
view_lazy.go hook is added. The integrator should reconcile this storage hook.

Provenance preparation now checks the four highest-ranked candidate pairs
(including the uncertified frontier), hashes all emitted external declarations,
and records complete LiteralType, PseudoBigInt and bindable field sets. The
original suite is opt-in and ran with declarations enabled. Whole-program lazy
census remains opt-in and unmeasured. Selected tests only; not the full gate.
The whole-family working target remains October 13, 2026, 23:00 UTC.

Final selected command (all output in logs/next-production-gate.log):

```
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/object-primitive-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitive|TestViewIntersection|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestCheckedViewArrays' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
git diff --check
```

Passed: lower 1.411s, oracle 53.104s; native and JavaScript had no matching
package-local tests and ran via the oracle. All 15 original cases passed,
including both frontier cases; the frontier is deliberately not certified.
Vet and diff checks passed. Fetch confirmed integration remains ba59427c.
One initial original run exposed an unread control containing an unused reader;
the cast-only control now contains no read. A subsequent rebuild failure during
removal of an unnecessary draft helper was corrected before this final gate;
neither failure was counted as a mutant kill.

October 8, five-pair original witness batch

Built: original DiagnosticRelatedInformation, AutoGenerateInfo, ConditionalType and both location diagnostic variants, plus five higher-ranked frontier witnesses.
Commits: follows 166cf410; implementation tip is the commit containing this batch.
Checks: uncached selected regression gate passed, oracle 73.194s; final complete original suite passed 53.297s; vet and diff checks passed.
Mutants: outer omission, wrong member, nested type acceptance and transitive omission caught in both release backends; optional string uses its outer wrong-member mutant plus nested omission.
Uncovered: 34 candidate pairs / 122 candidate reads remain; no whole-tsc execution or exact reachability claim.

Five new standalone original field-contract pairs carry 14 candidate reads:
DiagnosticRelatedInformation.messageText (6), AutoGenerateInfo.prefix (4),
ConditionalType.resolvedConstraintOfDistributive (2),
DiagnosticWithLocation.messageText (1), and
DiagnosticWithDetachedLocation.messageText (1). Together with prior witnesses:
8 certified pairs / 59 attached candidate reads. The original targets and
member types are imported in full. prepare.cjs audits the 34 highest-ranked
rows against pristine pinned upstream site texts and hashes all declarations;
this broader audit does not certify the unsupported rows. The oracle asserts
complete inherited field sets for every new target and its object member.

The diagnostic variants hold both valid alternatives and boolean/nested code
refusals. AutoGenerateInfo holds plain string, nested prefix, missing outer
property, missing nested optional prefix and explicit undefined. GeneratedNamePart
is complete; its unread node union stays deferred. ConditionalType holds false,
Type flags, missing property, and wrong true/flags; the true poison is boxed
through an actual runtime union producer so the omitted-check mutant can run.
Every wrong read pins exit 70, expression, expected declared type and found type.

Independent release mutant outputs, native / JavaScript, each exit 0:
- Diagnostic variants: accept wrong member or omit outer check prints wrong /
  wrong; accept wrong nested code or omit its check prints chain:0 / chain:false.
- AutoGenerateInfo: wrong-member acceptance and outer omission print wrong /
  wrong; nested check omission prints undefined / false. A draft nested type
  mutation broke the static optional-string conversion, so it was discarded and
  is not a kill. The required wrong-shape acceptance is independently held at
  the union member.
- ConditionalType: false-literal relaxation and outer omission print true /
  true; wrong flags acceptance or nested omission prints 0 / false. The first
  unboxed true omission mutant crashed; that was not counted. The final boxed
  input executes normally and independently fails only the refusal pin.

Uncertified higher-ranked frontiers are checked against source Node before
lowering: JSDoc.parent (10) and both bindable left fields (9 each) retain named
read-stage representation-conversion refusals. Original JSDoc.comment (7) is
NotYet at the union field. Original CompilerOptions dynamic-key (10) reaches an
up-front unchecked-cast refusal in this direct probe; that is a remaining lazy
admission gap, not compliant runtime checking and not certification. These
frontiers remain in the remaining counts. Previously recorded 11-read bindable
expression is also uncertified. All 35 original cases, including 7 frontiers,
passed their positive/refusal pins in the final complete original suite.

Dictionary member handoff is accepted from ce4eeaa4 GROUP4.md. Object/array and
TsConfigSourceFile member selection belong here; lookup, absence and enumeration
belong to lane 6 and scalar/nullish selection to lane 4. CompilerOptions dynamic
reads (10, rank 6) and BuildOptions dynamic reads (1, rank 36) already exist in
our inventory and are not added twice. resume/dictionary-member-handoff.json
retains the original rows, scope and counts. Container-only certification by
lane 6 does not credit either element pair. The newer integration tip discovered
during validation is 267fd75b; it will be merged after this batch is pushed.

Commands (full output retained in batch-*.log):

```
node stage3/interface-downcasts/lane4b/original/prepare.cjs /tmp/object-primitive-upstream /tmp/object-primitive-original-declarations
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/object-primitive-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitive|TestViewIntersection|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestCheckedViewArrays' -count=1 -v -timeout 10m
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/object-primitive-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOriginalPairs$' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
git diff --check
```

Native/JS package-local tests did not match; both emitters ran in the oracle.
Original suite was explicitly enabled, no original cases skipped; opt-in whole
compiler census was not run. No production hook is added in this batch.
Working whole-family target remains October 13, 2026, 23:00 UTC, now including
the explicitly accepted dictionary-member scope within the existing candidates.
