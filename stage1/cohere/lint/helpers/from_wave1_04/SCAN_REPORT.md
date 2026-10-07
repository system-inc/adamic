Built: scanNumber in one .a file, removing four Tailwind dependency edges; cumulative delivery is four helpers, twenty edges and six distinct consumers, with zero final blockers removed alone.
Commits: claim 5ff59dd69 precedes code; rules are parked and green on main c01907a7 as 20e03b57, helpers were rebased and green as c023b39c; implementation follows this evidence.
Commands and outputs: scanner gate PASS 15.868s, complete four-helper gate PASS 35.578s, vet PASS; 101,547 cases and 802,813 Go output bytes match on Node, emitted JavaScript and ASan/UBSan native; setup PASS 40s, nproc 5.
Mutants: accept a dot without fractional digits and consume an incomplete exponent compile and finish cleanly; only actual Go comparison catches them on every target, alongside all seven prior helper mutants.
Not covered: native whole-rule CSS engine/findings, arbitrary malformed Go byte strings, or full repository gate; the separately owned ASCII digit predicate is an explicit dependency, with actual Go membership supplied in these comparisons.

# Exact Go scanner behavior

The scanner is anchored. It accepts an optional sign, ASCII digits, an optional
dot requiring at least one following digit, and an optional exponent only when
all its required digits exist. Preserve Go's exact behavior: 5. consumes zero;
1e and 1e+ consume just the 1. This is not replaced with a JavaScript regexp or
Number conversion. Every consumed character is ASCII, so its returned UTF-16
index equals Go's consumed byte count even when the remaining text is Unicode.
The digit predicate's successful set comes from actual Go isDigit over all bytes.

Consumers in the frozen inventory:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Selection inspected every one of 573 origin refs, twenty distinct helper claim
blobs and the older shared comments claim. Every larger named helper was reserved;
this was the highest remaining unclaimed count. Claim 5ff59dd69 was pushed before
code. Both owned branches are rebased, re-greened and published on c01907a7. The
rule parking blocker remains shared registration unification #zmh9v36, not an
unreported production integration claim. No shared harness file was changed.

# Corpus and mutant evidence

The original 38 Go tests across all four consumer files passed and produced 103
captured sources. The unchanged Go parser extracts 346 literal/field values,
148 / 69 / 64 / 65 by consumer. See trim_fixture_cases.json and trim_capture.json.
This is the consumer source literal domain, not a runtime CSS engine call trace.

Add all 63,488 BMP scalar characters following an ASCII prefix, all 37,449 words
through length five over +-01.eEX, the eleven Unicode/empty boundaries, and 253
number/suffix combinations: 101,547 cases. The exact output is 802,813 bytes.
The first semantic mutant returns the consumed dot when its required fractional
digits are absent; Go returns zero. The second commits an incomplete exponent;
Go leaves it unconsumed. Both compile, return success with empty stderr, and differ
from real Go on source Node, emitted JavaScript and ASan/UBSan native. No compile,
runtime, sanitizer or leak failure is credited as a mutant catch.

With source /workspace/adamic-tools/env.sh:

- go test ./stage1/cohere/lint/helpers/from_wave1_04 -run '^TestScanNumber'
  -count=1 -v -timeout=10m: PASS 15.868s.
- go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m:
  PASS 35.578s, all four real Go comparisons and all nine semantic mutants.
- go vet ./stage1/cohere/lint/helpers/from_wave1_04: PASS, empty log.

Logs are retained in evidence/scan-tests.log, four-helpers-tests.log and scan-vet.log.
The owned oracle fixture adds a scanner mode and real private exports; trim's old
corpus and output remain unchanged, and its comparisons and mutants pass again.
Production Go and shared Adamic harness files remain untouched. Readiness remains
conservative and lists the unimplemented dependencies for all six consumers.
