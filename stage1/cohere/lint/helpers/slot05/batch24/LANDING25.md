Landing only: took current lint area onto all 68 completed helpers; no new claim.
Previous publication dedecf24; tested own-branch merge a9e1389b; publication SHA follows in final response.
Fresh batch24/batch15, shared link guard, seven Node probes, vet and format PASS; setup 26.303s, nproc 5.
All 34 rerun compiling semantic mutants caught by Go result or dependency-trace comparisons; exact witnesses in evidence/landing25-proof.json.
Not covered: full repository gate and seventeen required stage 1 external-input checks; no skip credited.

Main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. Area advanced to b28757f339f3253ed3796a4dd6138d9ee736d887. Its six changed files add link-only node-table validation, fixed-source path reparsing and bounded fix-pass behavior. The merge was clean and shared source was taken intact. This worker publishes only codex/lint-helpers-05.

All 2334 tracked input entries covering compiler, dependencies, configuration and helpers match dedecf24 exactly. The previous complete 24-package, 290-mutant run therefore remains applicable to the helper source. This landing freshly reran batch24 and batch15, including 34 of those same mutants, rather than counting them as new variants. Every native variant compiles and exits successfully under ASan/UBSan before a semantic difference earns credit. Helper sources, tests and corpora were unchanged.

Commands (each build shell sources /workspace/adamic-tools/env.sh; output is redirected directly to the named evidence log):

```
bash cloud/setup.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch24 ./stage1/cohere/lint/helpers/slot05/batch15 -count=1 -v -timeout=15m
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestNodeTableIsLinkOnly$' -count=1 -v -timeout=20m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v
go vet ./...
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05
```

Observed: batch24 PASS 284.706s, batch15 PASS 177.887s; shared guard PASS 117.816s with identical 13069337 bytes across 2154 rows; Node oracle PASS 1.425s with seven uncached probe misses; vet/format empty logs. No failed or skipped selected test. All exact mutant names, substitutions and first differing Go witnesses are retained in the JSON and lossless gzip log. These remain bounded helper checks, not whole-rule diagnostic coverage. The new shared guard was run as upstream supplied; no new mutant proof is claimed for it.

Setup timing lines are preserved in evidence/landing25-setup.log: go 0.022s, Node 0.023s, markdown 0.070s (validated installed bytes), submodules 0.070s, clang 0.190s, Go build 25.987s, deferred test-binary warming 26.272s, cache 26.273s, done 26.303s. Five processors, cgroup quota four cores. Full repository and seventeen external-input correctness checks were not run, relaxed or credited. All existing helper claims are finished; no unfinished reservation remains.
