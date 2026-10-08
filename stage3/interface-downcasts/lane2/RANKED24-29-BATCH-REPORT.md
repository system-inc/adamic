Certified eighteen more original array-field pairs covering mapper, flow, signature, modifier, JSX, configuration and private directory arrays.
Commits: 8303bd50, 5dcbad43, f947b7b1, 1b416b41, d75544c8; the final directory group and this report share the next commit, pushed separately.
Commands: each group's Node, sanitized native, release native, JavaScript and finishing leak checks pass; scoped counts and oracle vet pass.
Mutants: ten mapper/flow, five signature, four modifier/JSX, three function/children, two configuration and one directory mutant are caught by pinned stopping oracles in all three modes.
Not covered: the lane 7 intersection handoff, remaining 132 pairs / 232 reads, branded sorted arrays and production consumer/intrinsic totals.

| Group | Pairs / reads | Fixtures | Finishing controls | Mutants | Focused verification | Count refresh |
| --- | --- | --- | --- | --- | --- | --- |
| 24 | 3 / 9 | 28 | 16 | 10 | 65.553s, including count verification | 22.584s |
| 25 | 5 / 15 | 50 | 25 | 5 | 59.961s | 37.905s |
| 26 | 4 / 12 | 56 | 36 | 4 | 67.088s | 44.961s |
| 27 | 3 / 9 | 53 | 38 | 3 | 60.556s; added full JSX arm checks 11.416s | 35.744s |
| 28 | 2 / 6 | 16 | 12 | 2 | 21.561s | 6.935s |
| 29 | 1 / 3 | 7 | 5 | 1 | 4.214s | 1.289s |

The batch adds eighteen pairs / fifty-four original static candidate reads. Lane 2 now holds 202 pairs / 2957 reads, remaining 132 / 232 of 334 / 3189. All 210 fixture rows are measured; removing each group's rows reproduces its previous table byte for byte. All 132 finishing source controls and 25 finishing mutants have native leak checks. Each final mutant remains valid C, completes with Node output and no stderr, and changes a reached guard. Failed exploratory inputs and clang-only mutants were corrected and are not counted.

Every group preserves complete original declarations at pin 050880ce59e30b356b686bd3144efe24f875ebc8 and validates its original read spans. Detailed commands, diagnostics, declaration fields and logs are in RANKED24-ARRAYS-REPORT.md through RANKED29-ARRAYS-REPORT.md and each originalNN/evidence directory. Existing production adapters suffice; no production or cross-lane implementation files changed. The private directory binder is in the final group's oracle helper commit.

SourceFile.amdDependencies has three uncredited candidate reads. Its reached element path refuses an unsupported intersection contract and is listed for lane 7, codex/views-intersections, in INTERSECTION-HANDOFFS.md. Lane 2 proceeded with independent pairs. No message was sent to another worker.

The required global count refresh was run once per group and failed in existing fixtures; each failure and scoped workaround is recorded with its log in that group's report. No whole package test or full gate ran. The branch is codex/views-arrays-callables-parser, with integration 432d4913 already merged. Own fields remain 3 / 26 of 30 / 72, and tuple frontier credit remains separate. The remaining queue still includes three-read JSDoc comment unions, diagnostic arrays, a private type-parameter array, own-array fields and then two-read pairs.
