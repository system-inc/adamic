# Wave 23 shared fixes evidence

Read ../WAVE_23_SHARED_FIXES_REPORT.md for counts and categorized blockers. lint-receipt.json contains exact final command, environment, times, load and counts. fixes-lint-complete.jsonl.gz is the complete final gate; the two earlier compressed logs are explicitly interrupted attempts, not final results.

rule-matrix-summary.json reports every owned rule. The isolated-cases archive contains all 978 inputs, options, strict project configs, manifests, checker transcripts and four raw runtime outputs. The problem, smallest and await archives preserve original and reduced failures. No compiled binaries or private checker answers are shipped. Native transcripts came from the area's RuleContext.checker and are replayed for Node/emitted JavaScript. final-build-hashes.json identifies the exact gate-built artifacts; native-sanitizer-symbols.log proves ASAN linkage.

The Python runners regenerate comparisons from captured.json; scratch paths are recorded as used on this box and can be changed for a fresh workspace. runtime-observations.json records passing-case Go and native times including startup/recording under concurrent load, not a controlled benchmark. Upstream-prefix inventory, merge logs, local first-failure logs and registry/vet/format checks are retained separately.
