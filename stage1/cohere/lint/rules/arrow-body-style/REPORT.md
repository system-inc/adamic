Built `arrow-body-style` in .a: ported rule and exact repair builder; shared multi-edit application is blocked.
Commits: claim f9232c1b was pushed before implementation; the implementation commit is the commit containing this report.
Validation: upstream 199/202 parser-supported cases and corpus 230 files / 690 cases pass on Go, Node, emitted JavaScript and sanitized native; 37,967,495 corpus bytes agree.
Mutant: arrow_findings_suppressed compiles and exits successfully, then only output comparison catches it on Node, emitted JavaScript and ASan/UBSan native.
Not covered: shared engine integration of multiple edits, corpus-wide repair application, and the three shared-parser arrow refusals.

Implements AsNeeded, Always and Never; the returned-object exemption; all five messages; ASI repair refusal; comment-preserving removal; comma and for-initializer parentheses; and forced object parentheses on member/call left spines. Exact ordered disjoint edits are available in Rule.repairs, paired with the corresponding context finding. Each Repair.body is an AST index; edit positions are UTF-16 units. The rule-local repairs.a driver parses every source independently with Adamic, calls the actual port and compares finding spans, IDs, complete messages and each edit span/text against unmodified Go rule answers from repairs.go.txt. It does not import a Go AST or precomputed repair into the port. The 184-case repair corpus comprises 108 arrow cases and all 76 bind cases.

Findings per second, best of five wall-clock runs including startup, over 77 TypeScript compiler sources plus one 1,000-line positive file: native **406.40**, Node **705.87**, Go **5060.82** (1046 findings, 78 files). Native timing is unsanitized; correctness uses ASan/UBSan. These are finding-count rates, not fix-application rates. An earlier run without ADAMIC_TYPESCRIPT_SOURCE measured only one file and is explicitly not used here.

The current shared Finding can represent one edit span/replacement. The shared Go oracle additionally panics on len(d.Fixes) != 1. Both rules genuinely propose multiple fixes, including separated spans needed to preserve intervening parentheses/comments. The integration adapter must consume Rule.repairs and serialize/apply those edits; this unit does not collapse them into a different Go edit shape. Until that happens, ordinary registered findings intentionally have no repair. The rule-local diagnostic overlay removes fixes only from output serialization for these two rules and leaves the source unchanged; the actual Go judgments are unmodified. Consequently the large corpus and standard upstream run assert full diagnostic parity, not fixed-source parity. The separate 184-case repair run compares every actual edit and its text byte for byte, but does not assert Go engine convergence or corpus-wide fixed output. This is a declared shared-harness blocker, not a completed end-to-end fixer.

Commands, from the repository root, with every test's output redirected to a log file:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/max-lines/validate.py > /tmp/wave10-next-overlay.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-10-next/typescript go test -overlay=/tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestNext(Witnesses|Upstream|Corpus|Mutants|Repairs|Throughput)$' -count=1 -v -timeout 15m > /tmp/wave10-next-all.log 2>&1
```

Observed targeted results: Witnesses PASS 21.67s / 5,380 bytes; Upstream PASS 26.63s / 568,886 bytes (108 arrow, 15 max-lines, 76 bind); Corpus PASS 83.97s / 37,967,495 bytes; Mutants PASS 104.75s, all nine runtime mutant comparisons failed as required; Repairs PASS 11.71s / 49,634 bytes; Throughput PASS 59.65s. The exact commands actually executed split these tests across the evidence logs; the combined command above is a reproduction command, not a claim that it was run.

Preliminary combined commands exited 1: one corpus invocation lacked ADAMIC_TYPESCRIPT_SOURCE, and the owned repair driver first hit nominal nested-map conversion, then an untyped never-array fallback. The corrected corpus and final repair reruns passed. The initial compile also rejected multi-argument array push; owned code now pushes one value at a time. No compiler or shared harness source was edited to resolve any of these.

The exact three parser refusals are recorded in ../max-lines/testdata/parser-exclusions.json. One is a Go recovery-tree fixture that Go's strict adapter also refuses; the two for-initializer return expressions are accepted by Go but refused by the current Adamic parser. They are excluded explicitly, never treated as clean. Actual Go rule tests were run and their asserted cases captured before replay.

Setup was run earlier in this unit: ordinary setup failed at shared profile_test.go:32 because portFiles had become a function. With the existing isolated compatibility overlay, setup passed in 17s: Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 17s; nproc 5. Full uncached repository gate was not run. The shared harness branch published at 2650ad595b82220c368631ea13139fad4b306ed6 is an integration dependency, not an excuse to edit shared files here. Only owned rule directories changed after the pushed claim.

Public synthetic fixture output and compiler/stage1 evidence are retained under ../max-lines/evidence. Auxiliary Go sources use .go.txt; authored Adamic sources use .a. No .ts source was authored.
