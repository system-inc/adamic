# Typeaware deletion-set replay

Candidate: `TestVolumeProfileCorpora family`. No test was deleted or modified.

Starting main: `b8bcadb2c493173855f19d7e5c508b34f5eeb5b6`, detached with pinned submodules. Warm tools used; setup skipped. CPU count: 5.

The mutant list was saved before the baseline. Audit M1/M4 and defender D1/D3 qualify; audit M2/M3 and defender D2/D4 do not record a candidate failure. Every selected standalone diff applies to current main without adjustment. `TestVolumeAgreementRepository` is present on current main. Its historical catches are not assumed to persist; the enabled replay establishes its current catches.

Baseline and each replay use `ADAMIC_GATE_UNCACHED=1`, a fresh per-run `ADAMIC_BUILD_CACHE_DIR`, and:

```
go test -json -count=1 -timeout 30m ./stage1/cohere/typeaware/ -skip '^(TestVolumeProfileCorpora)($|_|/)'
```

The requested expression excludes 218 corpus leaves and their setup. It retains `TestVolumeProfileCorporaUnion`, which only checks enumeration and does not execute the port. Other default skips remain recorded in the logs. The first run left repository agreement opt-ins unset. The recovered 287-file defender repository manifest is enabled for a second clean baseline and the D1/D3 replays. That baseline passed in 493.329 command wall seconds (491.317 test-binary seconds), with 264 top-level passes and 95 reported skips; the interrupted unconfigured D1 run supports no conclusion. Baseline passed: 231 top-level passes, 128 reported skips, 606.639 test-binary seconds, 608.619 command wall seconds.

Mixed test names are classified by the actual failing case. `TestVolumeAgreementAndMutants_030` hashes to `controls`, and `_014` to `controls-asan`; both compare the unmutated native volume port with Go truth. Their controls mismatch at volume_agreement_shards_test.go:365 is an ordinary catch, not a planted-mutant witness. Other built-in mutant cases cannot be the sole catcher.

Slow replays may be stopped only after a clean ordinary failure. Such runs do not claim that later rows passed. Passing runs run to completion. Go test panic aborts require additional replays excluding the panicking top-level test. Individual native product panic assertions are distinguished from Go test-binary panics.

Initial disk checks: /tmp 8.6 GB free of 8.8 GB total, /workspace 12 GB free. Only earlier-unit `/tmp/defend-class-inheritance` scratch was removed. The requested 15 GB floor is impossible on the /tmp filesystem. No disk-full failure occurred. Completed per-run caches are removed after their logs and results are captured.

`mutant-list.json` preserves original recorded candidate members and other failures, branch provenance, diff paths, current-main source lines, and apply status. `matrix.json` records commands, cache identities, wall times, observed passes/failures/skips, failure output, and panic reruns. `result.json` contains the final deletion-set decision. Production sources are restored after each replay. Only this evidence directory is published.

## Result

All four gathered mutants remain caught outside the candidate family. No stale diff, Go test-binary panic, witness-only catch, or disk-full failure was observed. M1 failed ordinary controls and sanitized controls; M4 failed the ordinary cross-file agreement case; D1 and D3 failed repository sanitized group 1 (`TestVolumeAgreementRepository_017`). All mutant runs stopped after an ordinary top-level failure. Unobserved later rows are not inferred to pass.

Replay wall seconds: M1 345.733, M4 244.453, D1 122.676, D3 375.624. Baseline wall seconds: default 608.619, corpus-enabled 493.329. The interrupted unconfigured D1 run is retained but excluded from the decision.

The candidate is deletable only relative to this demonstrated mutant set. This report does not claim the corpus family guards no other behavior. Native and Go production sources were restored; `git diff --exit-code` passed before publication.
