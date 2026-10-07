Built: rebased wave 13 onto origin/area/stage1-lint at 7481e032, including requested 50a5f105, harness 41eb6eab2 and current main 39638d9e; no new claims.
Commits: rebased code 862dae3e replaces pushed 0a5a2641; this report records the renewed area landing evidence.
Commands and outputs: original agreement gate PASS 243.270s; next 129/129, more 430/433 controls equal in normal and ASan runs; frozen corpora, metadata, Node and scoped vet checks pass.
Mutants: nine rule, two export-question, one lifetime and 27 metadata mutants caught again.
Not covered: three shared parser recovery inputs, checker-context integration into the supplied-node registry, full repository gate and emitted-JavaScript rule comparison.

The explicitly requested area rebase completed without conflicts. All incoming
shared harness, finding-model, registry and JSX parser changes were accepted.
No shared files were edited or reverted by this unit. The old report's statement
that the shared harness is outside this branch is superseded: it is now an ancestor.

The compiler and bridge source trees, go.mod and go.sum are byte-identical to the
previous own push 0a5a2641. Stage 0 and the C archives were therefore reused. All
rule native programs were freshly rebuilt against the new area parser, in normal
and ASan modes. The existing toolchain was reused after its prior setup total of
217s; nproc remains 5. No main or area branch is pushed by this unit.

## Renewed checks

TestWave13AgreementAndMutants passes in 243.270s: 18,957 control bytes and 32 findings,
compiler 7,087 bytes and four findings, repository 19,144 bytes and one finding.
Normal and sanitized findings, fixes and suggestions match production Go exactly.
Released handles refuse with exit 70. The registry mutant exits 0 and is caught by
the required refusal check.

Original rule mutants unassigned, caught and exports compile and exit normally
with empty stderr, caught only by byte comparison at offsets 1345, 4136 and 14959.
Export-chain-flags and export-module-lookup likewise require byte comparison, at
14677 and 17114. All were rebuilt against the area parser.

Both continuation suites reran every upstream control under normal and ASan builds.
Next has 129 equal controls. More has 430 equal controls and exactly three parser
refusals among 433 controls in both modes: 53d1c0a9ffc2fefa, aff6ced2ca61fe89 and
defcb4c4ce6a2921. The area parser still enables JSX only for JSX-enabled extensions;
the first two controls are JSX-like recovery cases in .ts. The third needs a yield
label and break yield. They exit 70 before rules execute. No unsupported control
was omitted, rewritten or weakened, and no sanitizer diagnostics occur in the
retained control stderr streams. The React analysis parking exception does not
apply to this core rule.

Continuation mutants process-state, race-handle, blocking-order, namespace-alias,
literal-parentheses and executor-span compile and exit 0 with empty stderr, caught
only by byte comparison. Both continuation trios agree over all 77 frozen compiler
roots (4933 bytes) and 287 repository roots (18485 bytes), with zero findings.
Their positive controls and mutants provide nonempty evidence.

The owned metadata package passes all nine source-name, nine JSON-name and nine
numeric-JSON mutant rejection checks. All listener metadata uses ast.Kind names.
The filtered Node gate passes inherited_static_field_read.a, class_layouts.a and
TestTheOracleCatchesOneByte; worker oracle caches were reused where applicable.
Scoped go vet passes. The complete repository gate was not run. Other raw-fact
and foundation bridge mutants were not repeated because their code is unchanged;
prior evidence remains in LANDING_F801_REPORT.md. This run repeats the original
trio's lifetime checks; continuation released-handle evidence remains unchanged.

## Native time against Go

Three alternating full-output rounds ran after all builds and gates completed.
Every round compares canonical bytes. Measurements include checker loading;
native remains slower than Go. No shared-dispatch speed improvement is claimed.

| Trio | Population | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 8.377 | 3.257 | 2.57 |
| original | repository | 0.474 | 0.138 | 3.45 |
| next | compiler | 3.168 | 0.479 | 6.61 |
| next | repository | 0.457 | 0.140 | 3.25 |
| more | compiler | 3.516 | 0.524 | 6.72 |
| more | repository | 0.560 | 0.151 | 3.71 |

## Checker-context integration blocker

The shared supplied-node registry and finding model are now present. Its
stage1/cohere/lint/context.ts RuleContext contains source, parser, scanner,
settings and parent links, with report, reportNode and reportRange. It has no
checker handle, program/file identity or bridge question context. These rules use
Bindings, Rules and OutputContext over the checker session, including cross-file
module and symbol queries. The existing shared context cannot construct those
inputs for their native factories without shared harness/context integration.
Ahra prohibited shared registration-generator and harness edits; none was added.

Own suites therefore still invoke their per-file run() methods over typed bridge
contexts. Legacy relevance checks remain inside those methods. Named metadata
alone does not prove supplied-node or kind-indexed execution. This limitation is
explicit and no full shared-registry factory/class migration is claimed. The
finding wire protocol does not need changing for these implementations.

No production Go regex occurs in these nine rules, and no hand-rolled regex matcher
was added. The existing regex audit remains applicable. Three parser controls
remain partially blocked; no additional claim was taken under the landing cap.
