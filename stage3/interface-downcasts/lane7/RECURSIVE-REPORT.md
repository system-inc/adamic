Built runtime checking for recursive plain object/scalar intersection payloads.
Commits: selected arms 9c8c720e; central integration merge 9a044725; recursive group is this tip.
Validation: Node, sanitized C, release C, JavaScript, successful leak controls, touched packages and vet.
Mutants: skip, shape, nested, canonical, optional dispatch and three literal constraints each break three refusal pins.
Pending: 17 candidate pairs / 66 reads; full upstream certifications remain zero.

Working date: October 11, 2026, 23:00 UTC. The latest central integration merged
is 1d165a7c95a5c1f5f0c05de199fba04166c2b838. Its plan append conflict retained
both sides; production shared hooks merged cleanly. No peer branch was merged.

The canonical optional descriptor link repairs the reservation-copy gap: a Link
optional copy built before Link.Fields completes must not certify an empty shape.
Runtime recursion checks complete object obligations using (object, contract)
active pairs rather than dropping compiler recursion backedges. Native walks
immutable descriptor tables with explicitly freed path storage; JavaScript uses
an active Map/Set. Tables contain only object descriptors reachable from the read.

Fourteen recursive source cases cover finite chains, optional absence, a bad
number three links deep, missing count, wrong nested object, embedded-NUL string
literals, optional number and boolean literals, optional root payload/absence,
and null root payload/absence. An unread cast with a bad payload remains admitted.
Node controls print true (or ok for unread); wrong checked reads exit 70 with
exact field, expected type and found-value messages in every backend mode.

The merged nullish emitter exposed a second gap: its physical admission returned
a present optional intersection without invoking its contract. An optional-root
fixture caught it, and the two minimal named nullish hooks now retain that check.
These shared hooks and the optional canonical link are listed in the plan for
the integrator. The optional-dispatch and canonical-link mutants independently
prove both fixes observable at a demanded root read.

Each mutant produces valid execution with stdout true and exit zero, causing
three independent expected-exit-70 assertion failures (sanitized, release, JS):
skip deletes aggregate fields; shape accepts boolean Link.count; nested deletes
canonical Link fields; canonical deletes ObjectPresent links; optional clears
intersection dispatch; literal-mode, literal-code and literal-enabled drop the
string, numeric and boolean membership obligations respectively.

Limitations: compound recursive unions, arrays, callbacks, nominal classes and
nullable descendants do not gain runtime support here. Already unsupported
children retain lazy refusals at their own reads. Allocated cyclic source
programs and complete upstream tsc declarations are not certified. Original
witness metadata verifies spans, not whole shapes. Static candidates remain
17 object pairs / 66 reads plus 28 delegated pairs / 703 reads; zero unclassified
overlaps. Full tsc and the full repository gate were not run.

Validation commands and final outputs:
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView' -count=1: PASS 92.230s.
- Final SelectedArms|RecursiveDemand filter after nullable controls: PASS 8.634s.
- RecursiveDemand verbose source controls: PASS 6.463s.
- Touched package TestView|TestLazyView|TestChecked|TestOptional filter: lower 2.160s, native 4.618s, JavaScript 0.820s; IR has no selected tests.
- go vet on lower/native/JavaScript/IR/oracle and git diff --check: PASS.
- run-recursive-mutants.py: eight mutants each caught in all three execution modes.
- Unknown discriminator mutant (adding 999 to the selected tag arm): three valid-execution refusal failures. This supplements the selected-arm group's skip/shape/nested mutants.

Logs are retained in recursive-evidence. The prior guard-only recursive evidence
is historical; runtime checking now replaces that plain-object refusal. The
untagged compound refusal stays active. The two nullable root tests were added
after the broad gate and passed both separately and in the final focused suite.
