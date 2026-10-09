Defense evidence at origin/main cf735d9fba9e38de6368575e5630e44375a86eaf.

[
  {
    "test": "TestDefaultTaggedInterfaceNeedsNoFlag",
    "defense": "cannot-judge",
    "unique_mutant": null,
    "reason": "No exclusive coverage. Its source duplicates an admitted inline-cast use within the subsumer, except for an unused factory declaration. The seven-run budget was spent on distinct safeguards; no aimed mutant for this row.",
    "name_assertion_finding": "The test executes Node agreement now, but does not unset or vary an admission flag. The name promises flag independence beyond an assertion under the inherited environment."
  },
  {
    "test": "TestIteratorGapsAreExplicit",
    "defense": "cannot-judge",
    "unique_mutant": null,
    "reason": "One aimed attempt survived: no row failed. Three honest aimed attempts were not completed within the seven-run budget.",
    "name_assertion_finding": "Checks only NotYet category; does not establish which iterator gap caused rejection. The name promises explicit gaps more broadly than this category assertion."
  },
  {
    "test": "TestIteratorViewsCannotHideReturn",
    "defense": "defended",
    "unique_mutant": "D01 internal/lower/iteration.go:290",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. This defense protects diagnostic wording, not refusal category or runtime semantics.",
    "name_assertion_finding": "Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established."
  },
  {
    "test": "TestIteratorViewsCannotEraseReceivers",
    "defense": "defended",
    "unique_mutant": "D02 internal/lower/iteration.go:265",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. This defense protects diagnostic wording, not refusal category or runtime semantics.",
    "name_assertion_finding": "Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established."
  },
  {
    "test": "TestLiteralMethodCapturesCannotMakeCycles",
    "defense": "defended",
    "unique_mutant": "D05 internal/lower/iteration.go:57",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. The mutant changes the guarded production behavior.",
    "name_assertion_finding": "Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established."
  },
  {
    "test": "TestLiteralMethodViewsDoNotLoseThis",
    "defense": "cannot-judge",
    "unique_mutant": null,
    "reason": "The row now checks Node agreement. Its positive path is distinct from the refusal subsumer, but is also exercised by the new TestLiteralMethodSignatureViewsDoNotLoseThis. No aimed receiver-behavior mutant fit the seven-run budget.",
    "name_assertion_finding": "No observed mismatch between the name and current JavaScript receiver assertion: Node agreement observes this.value. Native execution is not checked, but the name does not explicitly promise native execution."
  },
  {
    "test": "TestRepresentedMethodReplacementIsNotYet",
    "defense": "defended",
    "unique_mutant": "D03 internal/lower/class.go:418",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. This defense protects diagnostic wording, not refusal category or runtime semantics.",
    "name_assertion_finding": "Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established."
  },
  {
    "test": "TestDestructuredMethodsCannotLoadOwnSlots",
    "defense": "defended",
    "unique_mutant": "D06 internal/lower/iteration_origin.go:80",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. The mutant changes the guarded production behavior.",
    "name_assertion_finding": "Accepts either Refused or NotYet without a reason. An unrelated refusal can satisfy it, so the asserted cause is weaker than the named own-slot safeguard."
  },
  {
    "test": "TestGenericIteratorViewsPreserveNativeArguments",
    "defense": "cannot-judge",
    "unique_mutant": null,
    "reason": "The row reaches an earlier nominal refusal, with no lowered product. Coverage and source were inspected, but no aimed nominal-argument mutant fit the seven-run budget.",
    "name_assertion_finding": "Yes. The name promises native argument preservation; assertions check only Refused plus nominal ancestry in What, before native arguments exist."
  },
  {
    "test": "TestIteratorSymbolKeysAreNotStringKeys",
    "defense": "defended",
    "unique_mutant": "D04 internal/lower/class_features.go:78",
    "reason": "Only this row failed in the executable default package matrix; two inventory rows skipped. This defense protects diagnostic wording, not refusal category or runtime semantics.",
    "name_assertion_finding": "Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established."
  }
]

CODE UNDER TEST and ORACLE: code-and-oracles.json and report.json. Production only was mutated. No test, harness, Node oracle, or preparation code was changed.
Coverage: individual row and named subsumer profiles instrument internal/lower. coverage-differences.json lists covered blocks and source lines exclusive to each row. profile paths refer to the starting commit. NeedsNoFlag has no exclusive lines; shared-line semantic input differences were also inspected.
Matrix: all 275 discovered top-level rows have observed pass/fail/skip results under every mutant. matrix.json contains exact passing lists for each defended row, failing lists and raw failure output. Native product caches are separate per mutant.
Standalone diffs: D01.diff through D07.diff; all passed go vet ./internal/lower/ and apply checks against the restored starting tree. Selector instrumentation was removed; production sources and tests are byte-identical to the starting commit.
Four mutants change production diagnostic constants. They defend only wording contracts, not rejection semantics. D05 drops literal-method closure registration; D06 allows the callee exception for a destructured method view; D07 disables the collected-element representation guard.

Brief feedback, costs and limits:
* The audit report is named REPORT.md, rather than report.md. Initial lookup failed, then the branch tree revealed its name. I initially read the test files during discovery before correcting the report lookup; the report was read before mutation planning or runs.
* Main changed from the audit's 8171b317 to the starting commit above. Package discovery grew from 239 to 275 tests. Positive admission/receiver tests were strengthened to Node agreement, and a signature-view receiver row was added. Old vacuity findings do not describe those current tests.
* Exclusive coverage is not a verdict. Generic refusal and positive receiver rows have exclusive lines, but a distinct mutation can still be shared by other rows. NeedsNoFlag has none and duplicates the subsumer's inline-cast behavior with one unused factory absent.
* Ten rows with up to three aimed attempts can require 30 full matrices, while a near-minute package run permits seven. This run obeyed the seven cap. Unattempted or singly attempted nonunique rows are cannot-judge, not not defended. Three failed attempts were not claimed.
* The definition of defended says no other package row catches the mutant, but two project-specific inventory rows skip by default: TestOriginalCycleLedger and TestOptionalWideningCensus. Their external/project inputs were not prepared. Defenses concern the default executable package gate; behavior under those opt-in inventories is unknown.
* Diagnostic-constant mutants are on the allowed menu and change real production output. Their unique catches prove wording sensitivity only. They do not prove that no other row guards the underlying unsafe admission.
* Whole-file combined reads were sometimes truncated by tool output; focused reads of the relevant test bodies and production safeguards resolved the mutation sites.
* Per-row coverage requires an instrumentation rebuild. The first two coverage commands took about 26s each including compilation, while later commands took about 3s. Commands were run with two workers; their wall durations cannot be summed as elapsed wall time.
* D07 survived the complete executable package matrix. Its before/after Lower witness is saved in D07-witness-clean.log and D07-witness-D07.log. Both runs produced the same earlier spread-representation NotYet. A direct Array.from probe also produced the same earlier optional-widening Refused (D07-witness-from-*.log). No output change was demonstrated, so D07 is an equivalent candidate, not proven unguarded behavior. The witness source and timings are preserved.
* No test was deleted, rewritten or weakened. No repository-wide uniqueness is claimed. No packages outside internal/lower were tested.
* Nondefended name/assertion findings appear per row in report.json. GenericIteratorViewsPreserveNativeArguments only asserts an earlier nominal refusal. NeedsNoFlag does not vary a flag. LiteralMethodViewsDoNotLoseThis now observes receiver behavior through Node, so no name/assertion mismatch was established for its current JavaScript contract.

Timing: warm env.sh worked, setup skipped; nproc=5; npm ci stage3/api 1.775s. Clean package binary 70.193s, command wall 72.718s. Seven matrix command wall times total 498.85s. Individual coverage and vet/build timings are in coverage-times.json and runs.json. No production run exceeded the binary's 90s budget.
