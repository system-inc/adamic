These diffs apply to origin/main d29d80ce. Apply one at a time in a clean scratch checkout, source /workspace/adamic-tools/env.sh, run go vet ./internal/lower/, then run:

ADAMIC_BUILD_CACHE_DIR=/tmp/defend-predicates/cache/<ID> timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > <ID>.log 2>&1

Restore the mutated production file before applying the next diff. D2 intentionally panics in one row. Its full-rest replay command is recorded in D2-rest-run.json. Run that command and each target separately to complete the observations. Source comparisons are to origin/main, not the evidence merge commit. The rows passed under each mutation are enumerated in matrix.json. All logs are gzip compressed JSON lines. Profiles and exclusive block files are plain text.

Top-level rows TestOriginalCycleLedger and TestOptionalWideningCensus skip under the recorded environment. They require an external pristine generated TypeScript tree or a chosen inventory project, respectively. Their mutant results are unknown.
