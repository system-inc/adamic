# Checked predicate calls

Item 2 builds checked admission on train 4, 1e81b051c52eee40d47c44f4a5f68b821e606708. The declaration's source extension governs admission. Stored .a witnesses are copied to temporary .ts files explicitly for checked mode. Their original sources also run on Node. Unproved .a bodies and overload claims remain refused. This does not add a checker option or erase a claimed type.

Named calls preserve argument evaluation order and execute the body once. True returns check the target once, even if the result is unused. False returns check the complement when the pinned checker actually narrows the caller; empty branches, negation, boolean aliases and continuation flows are included. A false branch that leaves number unchanged for a literal predicate, or rejoins the source domain through an unrelated conjunction, has no check. Assertions check normal return. Each failure exits 70 with the predicate, original call position, direction and both types. The existing checked-sites report includes plain calls and overload calls; unobservable directions no longer count as body proofs.

The call wrapper reuses the train's primitive, literal, nominal, array and checked-view reification. Positive tag results retain complete shared field views. Negative structural membership is admitted only over a complete tag partition; overlapping tags with different full contracts remain pending. Full untagged targets and unproved indirect calls remain named pending cases. No result here depends on another worker's unlanded branch.

## Oracles and mutants

Nine own source witnesses cover lying true, unused true, lying false with empty branches, lying false through a negated boolean alias, a compound false branch without narrowing, lying assertion, valid call order, valid assertion and unchanged number on a literal predicate's false branch. Node's original programs exit 0 with pinned stdout and empty stderr. Checked lying programs intentionally differ: both backends exit 70 with pinned diagnostics. Sanitized native, release native and JavaScript execute every witness. Finished valid programs pass LeakSanitizer and counted allocation balance; panic programs are counted at their stopping point and are not claimed to release their live panic state.

Five independent removed-guard IR mutants remove exactly the true, unused true, false, false-alias or assertion guard. Each mutant builds and executes in sanitized native and JavaScript, matches Node's unchecked source result, and disagrees with its exit-70 contract. The three branch rules therefore have executable catchers. The same lying .a true, false and assertion bodies stay refused.

Existing overload runtime witnesses now opt into temporary .ts mode. Three formerly unchecked lies now stop: the empty-array false branch, a true result passed only to console.log, and a nominal lie whose boolean is printed. The CLI explanation witnesses also opt into .ts, preserving their stored .a Node inputs.

## Counts

The nine new allocation rows are labelled with their stored witness path and .ts checked mode. Literal true and unused true, the assertion lie and unchanged number allocate nothing. False and false-alias rows hold one allocation at exit 70. The compound unchanged branch balances one allocation. The valid call-order row balances one allocation, two retains and three releases. The valid assertion balances three allocations and releases. No previous allocation row changes.

All changed direction rows follow the ruling. array_alias drops its unobservable direction from Proven. checked and false each change from one checked plus one unobservable/proven to two checked. every changes from one checked and three unobservable/proven to two checked plus two unobservable. nominal and once each replace two unobservable/proven directions with one checked and one unobservable. some replaces four unobservable/proven directions with two checked plus two unobservable. some_empty replaces two unobservable/proven with one checked and one unobservable. some_false_read, some_false_valid, some_objects and some_true_only each become two checked rather than one checked plus one unobservable/proven. Assertion, erased and callback rows retain their counts. The report labels an assertion direction asserts rather than true.

A counts attempt during development conservatively rejected the positive array_alias witness. The final rule uses the already landed shared checked view for positive membership and requires full negative membership only where false narrows. This is a resolved implementation failure, not a pending dependency and not a passed run.

## Commands

All output is retained in log files. Focused lowering uses -run 'Predicate|TestConditionAssertionAdmission|TestEveryNeedsCallbackEffects'. The focused oracle uses -run '^TestCheckedPredicate(Oracle|CountsAreRecorded)$'. CLI reporting uses -run '^TestExplainChecksOutput$'. The required global counts command is go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts. No whole package or full gate is run.

The independent overlap-admission mutant removes the negative tag-domain gate. TestCheckedPredicatePendingContracts/open_tag then fails with got <nil>, demonstrating that the pending boundary prevents admission. Final oracle and own counts verification pass in /tmp/predicates-oracle-item2.log (5.092s); global counts pass in /tmp/predicates-counts-item2.log (43.954s); CLI explanations pass in /tmp/predicates-cli-final.log (4.943s). Setup timing correction: both Go and Node ready are 0.022s.

Final focused lowering passes in /tmp/predicates-lower-item2-final.log (8.096s).
