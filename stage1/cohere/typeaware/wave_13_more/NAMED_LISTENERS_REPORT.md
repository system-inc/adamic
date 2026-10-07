Built: all nine wave 13 rule.json files and native listener declarations now use ast.Kind names; no new rules claimed.
Commits: based on pushed ee6c4d9e and current main c01907a7; named-listener code and evidence are committed together.
Commands and outputs: original agreement gate PASS 226.502s; next 129/129, more 430/433 controls equal in normal and ASan runs; frozen corpora equal.
Mutants: nine rule, two export-question and one lifetime mutant caught again; nine source, nine named-JSON and nine numeric-JSON metadata mutants rejected.
Not covered: three shared parser recovery inputs, supplied-node driver integration, full repository gate and emitted-JavaScript rule comparison.

This report supersedes the numeric-listener contract in SPEED_LISTENERS_REPORT.md,
JSON_LISTENERS_REPORT.md and earlier landing reports. Ahra's corrected contract
requires typescript-go ast.Kind names, with no numeric listener kinds.
All nine owned JSON descriptors now contain string kinds. The native .a class
fields likewise declare string arrays, without a runtime name-to-number mapper.

The independent tests parse production Go rule.Listeners declarations, resolve
keys to actual pinned ast.Kind constants and derive names with
strings.TrimPrefix(kind.String(), "Kind"), matching the shared registry's name
validation. Wrong Identifier listeners are rejected independently for each source
and JSON descriptor. Numeric JSON kinds are rejected for each descriptor as well.
The first metadata test run exposed a Go declaration error; it was fixed and the
unfiltered package gate then passed. No failing result was treated as green.

## Verification

The original trio's full agreement gate rebuilds native and sanitized suites,
compares complete production Go output, repeats mutants, checks released handles
and repeats both frozen corpora. Controls agree on 19,008 bytes and 32 findings;
the extra 51 bytes relative to the previous report come from one additional
character in each scratch-path header. Compiler output is 7,087 bytes with four
findings; repository output is 19,144 bytes with one finding.

The original unassigned, caught and exports mutants compile and exit normally
with empty stderr and are caught by finding-byte comparison. Export-chain-flags
and export-module-lookup likewise require byte comparison. Released handles refuse
with exit 70, and the registry mutant exits 0 and is caught by that refusal check.

Both continuation suites were freshly built from the named-listener sources in
normal and ASan modes. Next has 129 equal controls. More has 430 equal controls
and exactly these three parser refusals in both modes: 53d1c0a9ffc2fefa,
aff6ced2ca61fe89 and defcb4c4ce6a2921. No sanitizer diagnostics appear in the retained
control logs. The first two require parser recovery for JSX-like syntax in .ts;
the third requires accepting yield labels and break yield. The parser refuses
before any rule runs. The React analysis parking exception does not apply.

Continuation mutants process-state, race-handle, blocking-order, namespace-alias,
literal-parentheses and executor-span were rebuilt with the new declarations;
all compile and exit 0 with empty stderr, caught only by byte comparison.
Both continuation trios agree over all 77 frozen compiler roots (4933 bytes) and
287 repository roots (18485 bytes), with zero findings. Positive controls and
mutants provide nonempty evidence. Scoped go vet and git diff --check pass.

The prior foundation bridge mutants and the continuation raw-fact mutants were
not repeated because neither their code nor the compiler changed. Their evidence
remains in LANDING_F801_REPORT.md. No full repository Go gate was run.

## Native time against Go

After builds and checks finished, three alternating full-output rounds were run
for each population. Each round compares bytes. These measurements include
checker loading; native remains slower than Go and no driver speed improvement
is claimed.

| Trio | Population | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 6.425 | 2.078 | 3.09 |
| original | repository | 0.401 | 0.124 | 3.23 |
| next | compiler | 2.757 | 0.447 | 6.17 |
| next | repository | 0.427 | 0.142 | 2.99 |
| more | compiler | 4.890 | 0.660 | 7.41 |
| more | repository | 0.470 | 0.133 | 3.53 |

## Integration limits

The shared harness SHA ab70f38d4 remains outside current main. Existing own suites
call per-file run() methods over typed bridge contexts; they are not wired to the
shared supplied-node listener callbacks. Legacy relevance checks therefore remain
inside those run() methods. Named metadata alone does not prove supplied-node or
kind-indexed execution, and no such compliance or speed gain is claimed. The
remaining integration dependency is the shared typed driver, not numeric parser
kinds. These files are owned metadata descriptors rather than a claim that the
syntax-only registry's full factory/class contract is already implemented.

No production Go regex occurs in these nine rules, and the shared regex table
has no rows for them. No regex translation or hand-rolled matcher was added.
The test metadata parser's Go regexp is test code, not a ported rule regex.
The unchanged toolchain was reused: its prior setup total is 217s, nproc 5.
No shared files, allocator leak-check changes, main or area branches were edited.
Three parser controls remain blocked; no additional claim was taken.
