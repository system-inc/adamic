Built: numberWithSuffix in one .a file, removing four Tailwind dependency edges; cumulative delivery is five helpers, twenty-four edges, six distinct consumers and zero final blockers removed alone.
Commits: claim 44a0ba625 precedes code; rules remain parked and green as 20e03b57 on main c01907a7, previous helper delivery 05ec8b0b; this implementation follows its evidence.
Commands and outputs: suffix gate PASS 59.386s, complete five-helper gate PASS 91.597s, vet PASS; 1,218,564 cases and 27,938,475 exact Go bytes match on Node, emitted JavaScript and ASan/UBSan native; setup PASS 40s, nproc 5.
Mutants: accept zero numeric consumption, accept a suffix prefix and scan twice all compile and complete cleanly; only actual Go result/call-trace comparison catches them on every target, alongside nine previous helper mutants.
Not covered: whole native Tailwind findings/CSS engine, malformed Go byte strings or the full repository gate; scanner and digit dependencies remain explicit and their held implementations execute in these comparisons.

# Contract and consumers

numberWithSuffix(value, suffixes, scanNumber) calls the scanner once with the
unchanged value. Zero consumption returns false. Otherwise the entire remaining
string must exactly equal one suffix. An empty suffix may match an empty tail;
nil and empty suffix lists never match. Source and tail contents are unchanged.
The scanner callback is implemented by the independently held scan_number.a in
these comparisons, using actual Go digit membership.

Consumers from the frozen readiness ledger:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Selection refreshed all 582 origin refs and inspected every one of twenty distinct
helper claim blobs, plus the older shared comment claim. This was the highest
remaining unclaimed named count. Both owned branches were already rebased, green
and pushed on c01907a7. Claim 44a0ba625 was pushed before writing code. No shared
finding, context, main, registry, oracle or lint_test comparison file changed.

# Actual Go comparison

The 101,547 values include all four consumer fixture literal domains, every BMP
scalar after an ASCII prefix, every word through length five over +-01.eEX and
number/suffix boundaries. Cross twelve suffix sets: nil, empty, empty-string,
percentage, length, angle, mixed, case-sensitive, Unicode, NUL and duplicate
sets. Total 1,218,564 calls. Original consumer capture covers all 38 test functions,
which passed in Go and yielded 103 distinct sources. Fixture literal extraction
is an input-domain replay, not a claim to record runtime CSS engine calls.

An observer-only Go overlay records actual scanNumber calls while the unchanged
real numberWithSuffix executes. It adds no decision logic and never substitutes
a result. Each observation contains the boolean result, scanner call count and
unchanged scanner argument. The native/source/emitted drivers record the same
callback observations. There are 27,938,475 identical bytes on all three targets.
The direct scanner and trim comparisons remain unchanged and pass again with the
observer disabled. The owned copy helper resolves the scanner dependency only for
the suffix driver; scanner mutants still load their actual mutated scratch file.

The three mutants change zero-consumption rejection, whole-tail membership and
callback count. All compile, return success with empty stderr and no sanitizer or
leak findings. The third preserves results but calls twice, proving the comparison
observes the dependency contract rather than only finding-equivalent booleans.
Every mutant differs from actual Go on source Node, emitted JavaScript and native.

With source /workspace/adamic-tools/env.sh:

- go test ./stage1/cohere/lint/helpers/from_wave1_04 -run '^TestNumberWithSuffix'
  -count=1 -v -timeout=10m: PASS 59.386s.
- go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m:
  PASS 91.597s, all five real Go comparisons and all twelve semantic mutants.
- go vet ./stage1/cohere/lint/helpers/from_wave1_04: PASS, empty log.

Evidence is retained in suffix-tests.log, five-helpers-tests.log and suffix-vet.log
under evidence/. Readiness removes only delivered symbols and conservatively
retains all other prerequisites. The shared registration blocker #zmh9v36 remains
named on the separate parked rule branch; no shared harness handoff SHA was named.
