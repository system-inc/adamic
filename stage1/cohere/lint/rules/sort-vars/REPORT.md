Built `sort-vars` in Adamic, preserving Go string ordering, ignoreCase defaults, literal-only fixes and original separator text.
Validation: all 71 upstream cases passed; final original corpus covered 224 files / 672 cases / 37,782,851 identical bytes on Go, Node, emitted JS and sanitized native.
Mutant: sort_vars_comparison_disabled compiled and ran, then failed only output comparison on all three runtimes.
Findings/s, native / Node / Go: 876.18 / 1257.95 / 5021.62, 77 compiler files plus 1,000 positive sources, 1,018 total findings, best of five.
Limits: shared profile compilation and .a registration still need codex/lint-harness-dot-a; full uncached repository gate was not run.

Claim 7915fd3 preceded source. The Go adapter retains upstream defaults and accepts captured decoded settings, including the exported IgnoreCase spelling. An initial decoder incorrectly treated captured settings as raw user configuration and was corrected. Rule-local initialization now tests the separator before the final child, so equals signs inside a type do not classify that type as an initializer.

`validation_test.go.txt` is an owned virtual Go test file. The already-owned Tailwind validate.py maps it through a scratch Go overlay. Shared dispatch, corpus and oracle sources remain untouched. Commands (all outputs sent directly to log files):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py > /tmp/wave10-original-prepare.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestOriginalWitnesses$' -count=1 -v > /tmp/wave10-original-witness.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestOriginalUpstream$' -count=1 -v > /tmp/wave10-original-upstream.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestMutants/sort_vars_comparison_disabled$' -count=1 -v > /tmp/wave10-original-sort-mutant-final.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-10-next/typescript go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestOriginalCorpus$' -count=1 -v -timeout 20m > /tmp/wave10-original-corpus-final.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-10-next/typescript go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestOriginalThroughput$' -count=1 -v > /tmp/wave10-original-throughput.log 2>&1
```

TypeScript src/compiler is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Native correctness uses ASan/UBSan, empty stderr and exit zero. Native timing is optimized without sanitizers; measurements include whole-process startup and input handling, excluding compilation, on a worker concurrently running other validation. The 18 extra sort-vars findings come from actual compiler sources. Mutants must compile and finish successfully before wrong output is compared. No shared file was changed by these tests.

Final TestOriginalCorpus passed in 72.342s after the final initializer fix and snapshot addition. The retained raw evidence includes significant trailing whitespace in diagnostics and fixture text; whitespace checks are for source, not byte-preserved evidence.
