Built `max-lines` in .a: ported and validated; no repair serialization dependency.
Commits: claim f9232c1b was pushed before implementation; the implementation commit is the commit containing this report.
Validation: upstream 199/202 parser-supported cases and corpus 230 files / 690 cases pass on Go, Node, emitted JavaScript and sanitized native; 37,967,495 corpus bytes agree.
Mutant: max_lines_findings_suppressed compiles and exits successfully, then only output comparison catches it on Node, emitted JavaScript and ASan/UBSan native.
Not covered: the unmodified shared registration/profile gate; the three arrow parser exclusions do not exclude any max-lines case.

Counts CRLF, CR, LF, U+2028 and U+2029, omits the empty line after a final break, and applies the actual Go whitespace set plus BOM removal. Comment-only skipping consumes parser-anchored shared comments and converts their byte spans to UTF-16 units. The finding begins on the first counted line after the maximum and ends at EOF. This rule has no fixes. Captured Maximum/SkipBlankLines/SkipComments options and raw integer/max aliases are supported.

Findings per second, best of five wall-clock runs including startup, over 77 TypeScript compiler sources plus one 1,000-line positive file: native **29.58**, Node **59.11**, Go **284.45** (57 findings, 78 files). Native timing is unsanitized; correctness uses ASan/UBSan. These are finding-count rates, not fix-application rates. An earlier run without ADAMIC_TYPESCRIPT_SOURCE measured only one file and is explicitly not used here.

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
