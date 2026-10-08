Built the highest-count required dictionary container fixture: ParsedCommandLine.options, with unread rich union children retained lazily.
Commits: this group follows 26e4694d on codex/views-dictionaries; no shared implementation hook changed.
Checks: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionary' -count=1 -v passed in 18.127s; candidate reproduction passed.
Mutants: skip-check, accept-wrong-shape and drop-transitive-check reran successfully in C and JavaScript for fixed and record producers; producer-certificate mutant also passed.
Not covered: rich dynamic element unions, enumeration and uncertified writes; 27 candidate pairs / 111 candidate reads remain, exact reachability unmeasured.

October 18 was overestimated: I treated rich child contracts as a prerequisite
for reading their dictionary container. Lazy admission removes that dependency.
The newly certified container pair alone accounts for 111 candidate reads.
The fixtures are original representative read-contract shapes, not a copy of
CompilerOptions or a claim that its full dynamic element union is supported.
A valid container containing unused null and object/array alternatives agrees
with Node in sanitized C, release C and the JavaScript backend. A number stops
with exit 70 and exactly:

    adamic: panic: field read failed: view.options is not a CompilerOptions; expected CompilerOptions, found number

An absent required container stops with exit 70 and exactly:

    adamic: panic: field read failed: view.options is not initialized; expected CompilerOptions, found missing

The initial missing-message expectation was corrected to the observed named
initialization refusal before the final gate passed. There is no allowed absent
case for this required property; optional-container cases remain covered by the
existing paths tests. This batch changes fixtures and inventory only.

Revised whole-family working estimate: October 11, 2026, 23:00 UTC. Confidence
is limited until the remaining dynamic union and optional receiver fixtures are
measured. This is an estimate for per-pair contracts, not whole-program tsc.
No other-lane handoff is required to continue the independent container pairs.

Proposed overlap split, for the user to relay to lanes 4 and 4b:

- Lane 4: scalar and nullish selection at dictionary element extraction.
- Lane 4b: object/array selection for CompilerOptionsValue and the additional
  TsConfigSourceFile alternative, preserving descendant read contracts.
- Dictionaries: shared record storage, lookup and absence, optional receivers,
  enumeration adapters and transitive propagation. Enumeration has no separate
  candidate row here and does not explain the former eleven-day estimate.

The relevant rich union candidate rows are CompilerOptions dynamic reads
(10 candidate reads) and BuildOptions dynamic reads (1). Container certification
must not silently credit either. Candidate any/string/Path rows also need
classification; they are not automatically dictionary-object reads.

Raw final gate: logs/group4-oracle.log. Previous broad baseline failures remain
as documented in GROUP3.md; no broader gate was claimed green in this batch.
