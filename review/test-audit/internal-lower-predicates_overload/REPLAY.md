Apply one diffs/MNN.diff to origin/main at 7709c91213f476eba7e3dbfbf4ee65e6988038cb. Each diff is standalone, without a selector. Source /workspace/adamic-tools/env.sh; run go vet ./internal/lower/; then ADAMIC_BUILD_CACHE_DIR=/tmp/u042/replay/MNN timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > replay.log 2>&1. Restore before the next diff.

M01 aborted; the bounded reruns are separate per-row logs. Other columns completed 239 tests. completion-checks.json and bounded-mutants.json distinguish them. Raw logs are tracked explicitly despite the repository log ignore rule.

Empty-answer probes are separate: probes/PXX.switched.diff plus audit_selector.go from switched-source, selecting ADAMIC_MUTANT=PXX. probes.json names their entries and rows. They do not support mutation verdicts. Production fixtures and oracle implementations were not mutated.
