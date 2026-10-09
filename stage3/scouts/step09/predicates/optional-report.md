# Predicate optional alias conversion

Built ahead on the explicitly named optional-presence dependency compiler/optional-presence a774d31656ebbb183d259e445b56135a71d5a1fb, merged only into compiler/step09-predicates-ahead. That dependency is not an ancestor of the selected train-4 tip 1e81b051. Semantic results here are build-ahead evidence, pending on that dependency for train acceptance.

The conversion uses actual insertion-rank presence before readiness and physical payload types. Missing optional strings return undefined without manufacturing an own field. A present undefined value remains present, and is accepted only when the non-missing field type explicitly admits undefined. The pinned checker's GetNonMissingTypeOfSymbol supplies that distinction; no checker option changed. Required fields remain required. Scalar required fields through optional receivers short-circuit once when the receiver is absent. Getter and accessor aliases, optional numeric/boolean slots, array/object optional aliases, and combined optional receivers and optional slots retain their named stops.

Node observations for optional_alias.a are `missing:false:missing`, `undefined:true:missing`, and `value:true:payload`, each on its own line. Both source modes, native with ASan/UBSan and leak checks, release native, JavaScript, and balanced allocation counting agree. optional_read.a's former pending stop becomes an executable oracle. optional_receiver.a holds the corpus's numeric receiver conversion to Node. optional_wrong.a holds present undefined against a string-only optional slot: source Node prints `undefined:missing`; both compiled backends stop at exit 70, naming node.text and the string/nullish mismatch. This checked failure is deliberately distinguished from Node's unvalidated output.

The independent presence mutant replaces the presence query with a payload undefined test. Both backends print `undefined:false:missing` instead of Node's `undefined:true:missing`, with exit 0 and empty stderr. Existing default-false kind-body mutants also remain caught in both backends. Optional number slots and combined optional string slots/receivers keep pinned NotYet negatives. Two independent measurement mutants, dropping a body and inventing an admission despite its diagnostic, fail the exact-body audit.

## Exact baseline subset

All 79 source hashes from 8a7ab17e remain unchanged; all 580 bodies and all 320 checker diagnostics are retained. The body-only probe does not measure callers or claim executable corpus output. It compares the exact 289 bodies whose old first stop was debug.ts:301:48, node?.kind.

| Group | Selected bodies | Complete view admissions | Pending |
| --- | ---: | ---: | ---: |
| direct kind comparison | 277 | 0 | 277 |
| delegation or composition | 12 | 0 | 12 |
| Total | 289 | 0 | 289 |

Every selected body passes the old first stop. The new first stops are debug.ts:383:139, symbol.declarations, for 151 bodies, and utilitiesPublic.ts:768:16, node.original, for 138. These need the array/object optional view adapters. Logical proofs remain 309; complete predicate admissions remain 18. This is zero newly admitted bodies, not 289 passes. Per-body old/new diagnostics are retained in evidence/optional-admissions.json.gz.

## Verification

Commands write directly to logs, retained compressed under evidence/optional-*:

- `go test ./internal/lower -run 'Predicate|TestConditionAssertionAdmission|TestEveryNeedsCallbackEffects' -count=1 -v`: PASS, 8.405s.
- `go test ./internal/oracle -run '^TestCheckedPredicateOracle$|^TestPredicateKindProofOracle$|^TestPredicateOptionalAliasPresenceMutant$|^TestOptionalFieldPresence' -count=1 -v`: PASS, 14.371s.
- After the static owner resolver was retained: `go test ./internal/oracle -run '^TestPredicateKindProofOracle$|^TestPredicateOptionalAliasPresenceMutant$' -count=1 -v`: PASS, 8.120s.
- `go test ./cmd/adamic -run '^TestExplainChecksOutput$' -count=1`: PASS, 1.432s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts`: PASS, 64.718s.
- Body measurement: generate measure-overlay.py, build measure-probe with its overlay and hatch_predicate_measurement tag, then run `HATCH_BODY_ONLY=1` against /workspace/scratch/predicates-corpus. audit-optional-admissions.py compares the retained baseline, raw body output and that tree. PASS.

Counts add four own rows: optional_read (now admitted), optional_alias (three presence states), optional_wrong (checked stop), and optional_receiver (undefined and numeric receiver). No existing row changes. The dependency carries its separate seven rows. The initial old pending test failed because its stop had been removed; it was replaced by the semantic oracle, not relabelled skipped. The first presence mutant attempt changed no instruction, and the second had an invalid receiver ABI; neither counted as a catcher. The corrected mutant compiles and produces the observed semantic disagreement above.

Setup used GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh: Node 0.027s, Go 0.029s, markdown ready 0.072s, submodules 0.072s, clang 0.174s, build 42.412s, tests deferred 42.514s, warm 42.515s, done 42.544s; nproc=5. No full gate or whole-package test ran.

Q5 is answered with independent presence and non-missing types for the implemented string conversion. Q1's condition-only ruling is now received and will be a separate push. Q2–Q4 and Q6 remain recorded in corpus-report.md, pending the owning views slices. General accessors and optional array/object/boolean/numeric slots are not claimed as implemented. This work serves roadmap step 09.
