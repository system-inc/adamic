u009 checker audit, cmd/adamic-meter
Starting origin/main: b955d3030e79e9f01d12a96ab1fb4b167fc91c0c

CODE UNDER TEST: Adamic meter Go report construction, in-memory import and optional-property adaptation, and annotated-return adaptation. No compiler, Node, fixture or existing test was mutated.
ORACLES: self-written counts/labels/strings, unmodified linked TypeScript checker acceptance and diagnostic results, and Node runtime output in three rows. external-run for linked TypeScript means in-process execution, not an external command. No external-authority classification is claimed.

functions.txt lists all reached named package functions and baseline statement coverage. Closures in these functions are included in their mutation scope. All functions except main were reached; returnAdaptations only reached its guard, with 3.8% coverage. Total coverage 66.8%.

The fixed 20 mutants were recorded before any matrix failure was examined. M14 changes the return at origin/main line 413. M19 changes the returned rewrites value to nil. No supplemental inserted-statement mutant was used.

M1.diff through M20.diff are independent changes against the base, with no switch. Each was applied separately and passed go vet ./cmd/adamic-meter/. E_*.diff are separate empty-answer probes; they do not support kills or unique_kills. All five probes were separately vetted too. switch.diff is only instrumentation for selecting all changes in one compiled test binary; the production files were restored before the final gate.

Matrix replay command, with ID replaced:
ADAMIC_MUTANT=ID timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-meter/ -run . > ID-matrix.log 2>&1
To replay a standalone change, apply its diff and omit ADAMIC_MUTANT.

M16 whole-package run timed out at 90 seconds. Every original row was rerun alone. Four positive optional-adaptation rows timed out again; ten rows passed. These timeouts are excluded from assertion kills and are listed separately. Rows with unfinished M16 behavior have bounded=true and list all rows rerun. Sacred uniqueness is established by complete, different mutant runs. Subsumption only concerns observed assertion kills, not the cooked M16 result.

E_MEASURE panicked the combined binary, so every original row was rerun alone. Panics observed on the row's own isolated run are probe failures. Missing rows from the aborted package run are never inferred.

probe_kills lists every observed probe failure, including propagation through downstream calls. own_entry_probes identifies the entries used to judge vacuity. measure is the entry for corpus/import reports, run for CLI JSON, normalize for normalization, adaptations for optional rows, and returnAdaptations for the return-contract rows. The implicit-return row calls both adaptations and returnAdaptations. Preparation Load/LoadOverlay calls were not probed. Negative-only rows pass their empty-answer probe and are vacuous under the brief's definition, even when they catch unsafe adaptations.

Survivors are M10, M18 and M20. observe_source.txt is an independent temporary Go observation test, not a mutation of audited tests and not used for verdicts. It was compiled with the switch, run as:
ADAMIC_MUTANT=ID /tmp/u009/observation.test -test.v -test.run '^TestAuditObservation$' > survivor-ID.log 2>&1
An empty selector is the control. M10 changes column-20's UTF-16 position from byte 21 to byte 22. M18 changes one return rewrite to zero using an explicitly synthetic TS7030 diagnostic input; today's Load emits no TS7030 for the current return fixtures. M20 removes actual resolver diagnostic TS2322 on a required string initialized by an unchecked array read. No survivor is merely an equivalent candidate.

Oracle limitations: several rows assert diagnostic counts or codes rather than complete diagnostic identity. The implicit-return Node run checks original source, not a rewritten product. Existing return tests exercise the adaptation guard, not the TS7030 rewrite path. The live-method row passed every production mutant in this fixed set; this does not establish that no conceivable mutant could fail it.

Brief interpretations and costs: Go-only unit required npm ci anyway (395ms); linked compiler execution classified external-run with that detail named. Empty-answer vacuity is expected for tests whose entire intended answer is no adaptation, but is still reported as required. M16's package timeout necessitated all-row isolation, costing four further 90-second positive runs. An initial switch build had two incorrect selector return types, corrected before matrices; all standalone mutants had already vetted successfully. /usr/bin/time was absent, so observation compilation was timed with Python. All evidence tests go to logs.

No other packages were tested, no repo-wide uniqueness established, no native products built, no PR opened, and main was not pushed. Production changes are fully reverted. The restored full package and vet pass. Subsumption is a small-set hint, not a deletion recommendation. All raw commands, timings, logs and independent diffs accompany report.json.
