# Current-pin landing evidence

This directory retains fresh checks of slot 05's 95 delivered helpers on cohere 7945d102a6c18dd36adf9114a758ce646e8b2359, main 48c05d091f0a43c31cbe051b1d6578d99eeedf19 and lint area 65017b318da1995237ff3ea2c80f59b055b39ac3. The directory is new; earlier reports, corpora and verdicts remain untouched. Final status and observed package results are recorded in REPORT.md once verification completes.

The exact Go declaration audit is reproducible from the cohere directory:

```
go run ../stage1/cohere/lint/helpers/slot05/landing_7945d102/testdata/audit.go ../stage1/cohere/lint/helpers/slot05/landing_7945d102/symbols.json > /tmp/lint05-body-audit.json 2> /tmp/lint05-body-audit.stderr
```

The tool parses production Go declarations from both Git pins, resolves the named receiver and function, formats the full signature and body without comments, and records hashes. It refuses missing or ambiguous definitions. The initial duplicate-init diagnostic is retained; init is not a named delivered helper and was excluded from the lookup. The final audit covers exactly 95 helpers: 92 declarations are identical, NormalizedFileName adds FileName().AsString(), and the two ClassLiteralReader methods switch to the current compiled selector predicates. This audits the named declarations, not every transitive dependency or TypeScript parser implementation.

Owned test adapters retain an exact Go pin assertion. Parser construction uses the new typed rooted file and PathKey. The filename test's oracle-only constructor intentionally converts raw strings to the typed filename to preserve its original unnormalized boundary probes; it does not normalize the string or alter Go's target helper.

Callee and variable helpers now accept readsCallee(index) and matches(name) callbacks respectively. These callbacks stand for separately owned current reader selectors. The isolated oracle supplies actual Go predicate answers for the same nodes/names, and separately checks the actual private Go helper output. This proves the helpers' orchestration and payload handling, not a selector engine implementation. All current selector configuration changes, member-call controls and the public callback API are described in ../batch2/README.md. Source, emitted JavaScript and sanitized native are checked for this changed package. Older packages retain their existing backend coverage; the final report does not imply every original package added emitted-JavaScript comparison.

run-owned.sh records the initial broad test invocation. Batch2 failed before its old oracle field was migrated, so that aggregate invocation exits 1. The completed current batch2 is separately held by batch2-complete.log. All other successful package observations from the aggregate run remain valid because their source and adapters were unchanged during that run. Batch33 was independently checked in first-batch33.log before the broad run. Root slot05 checks select the three owned TestSlot05 tests, not other workers' root tests. observations.json retains each individual package result and every semantic mutant witness; no initial failure is credited as a caught mutant.

The full repository gate, the 17 required external stage 1 correctness checks and whole-rule findings/fix integration are outside this bounded worker gate. No skip, check or guard is relaxed or removed. The current dynamically constructed RegExp blocker is checked by batch34 and remains an explicit NotYet; it contributes no delivered helper or semantic mutant.
