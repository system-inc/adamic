Built: collapse.isVector in is_vector.a, preserving exactly three numeric components, ASCII-space separators, full anchoring and scanner observations.
Commits: claim 8ed34d030885015701526e542f9e50a550b910e9; base main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; delivery commit follows this report.
Commands and outputs: isolated vector Go comparison and mutants PASS 39.515s, owned-package vet PASS; 170,369 cases and 4,378,863 identical bytes on Node, emitted JavaScript and ASan/UBSan native.
Mutants: two components, tab separator, ignored remainder and twice scanning all compile, finish cleanly and are caught only by actual Go comparison on each target.
Not covered: full CSS engine runtime calls and whole-rule findings integration; literal-domain replay is bounded, and the scanner is an explicitly supplied dependency.

The helper removes one listed prerequisite from each of these four rules:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Together six owned helpers remove 28 prerequisite edges for six distinct rules.
None of those rules becomes completely helper-ready: readiness.json preserves all
other listed prerequisites. The production helper is one .a file; vector_main.a
is only its comparison driver. No shared harness, compiler or rule files change.

The already captured original Go rule test corpus covers all 38 test functions,
103 distinct sources and 346 literal/field helper-domain values. The scanner's
101,547-case population includes those four consumer domains, every BMP scalar
after an ASCII prefix, all words through length five over +-01.eEX and numeric
boundaries. Vector coverage adds 63,488 BMP scalar separator points, 5,324 signed,
fractional, exponent and invalid numeric triples with single/multiple spaces,
tab and newline separators, plus ten whole-string anchoring controls. Total
170,369. This uses original rule fixture literals as inputs; it does not claim
these values were actual runtime CSS engine helper calls.

A temporary observer-only overlay adds a recording hook to the real pinned Go
scanNumber. The unchanged isVector body executes in Go, recording its actual
lazy scanner count and unchanged arguments. The held scanner port supplies the
same dependency to the Adamic implementation. Each row compares result, callback
count and all three scanner arguments. The twice-scan mutant preserves boolean
answers but fails the recorded dependency contract. All four mutants return zero
and produce no stderr, sanitizer or leak failure; comparison is their only catch.

With source /workspace/adamic-tools/env.sh:

go test ./stage1/cohere/lint/helpers/from_wave1_04 -run '^TestVector' -count=1
-v -timeout=10m: PASS 39.515s. go vet on the owned package: PASS, empty log.
The full six-helper rerun is recorded below when it finishes. Evidence files
vector-tests.log, vector-vet.log and six-helpers-tests.log retain actual output.

Landing-first check before claim: rules 77e7562fc and helpers a9fe6fd71 are both
rebased, green and pushed on main b8fb957a. Every one of 594 origin refs and every
one of twenty distinct helper claim blobs, plus HELPERS.md, was inspected.
isVector ties for the highest unclaimed reach at four. Claim pushed before code.
The rule branch remains parked on #zmh9v36 as its owned PARKING.md explains.
No production numeric-dispatch performance gain or new findings/second
measurement is claimed by this helper delivery. Setup PASS 87s; nproc 5.

Complete owned six-helper gate: go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m PASS 125.376s. All six actual Go comparisons and sixteen compiling semantic mutants pass on all three targets.
