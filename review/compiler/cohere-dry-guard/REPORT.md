Built the test-only DRY guard toward #xjdce2d, unit #w3d0ww1, with 25 decided entries.
Implementation commit d5e097ff; main integration commit 2e2a788e (origin/main af8b9db2).
Focused guard passes (48.259s); TestCallTargetReaders and integration lane checks pass; exact logs accompany this report.
All eleven isolated input mutants are caught; three real repository mutants make TestCohereMirrors exit 1.
Not covered: semantic clones without attribution or identifiable IDs, computed IDs across files, approved stage1 ports, or the full gate.

The old 9045cdb4 attempt was read whole and rebuilt, never merged. No cohere code was copied. The guard code derives from the historical Adamic test. The delivery branch began at origin/main 2b1be383 and merged updated main af8b9db2 before re-greening.

The inventory lives at internal/dry/testdata/cohere-mirrors.json. It is runtime-free test input and belongs in the test-only lane. Seven flow lifts name #2kb2kje. Relation and nominal judgments name #brf7xmj; other local soundness mirrors name the parent #xjdce2d. One keep decision explicitly records the existing rule_runner consumer, which already references cohere. The id-free suppression refusal is pinned by its diagnostic wording. The old no-definite-assignment entry is absent because current main implements readiness checks rather than that old refusal.

Conservative assumption: a local diagnostic using a registered cohere ID counts as a mirror until its owner migrates it. Test assertions are references, but test Rule registrations and Refused diagnostics remain guarded. The cohere soundness set validates parsing, and detection uses all parsed registrations, not only that set. Leading source headers are scanned in Go, Adamic, TypeScript, C/headers, JavaScript, Python and shell; parsed Go comments are checked beyond the leading header too. Stage1 has its own approved port/registration guards; rule scanning there is excluded while provenance headers remain checked.

Setup: first timeout 360 bash cloud/setup.sh reached submodule readiness but was terminated by the hard limit during cache warming. Retrying with export GOPROXY='https://proxy.golang.org|direct' and timeout 120 bash cloud/setup.sh succeeded. Both logs are preserved. Sourced /workspace/adamic-tools/env.sh for every test. Retry timing: Go 0.038s, Node 0.056s, submodules 0.116s, markdown dependencies 0.132s, clang 0.403s, go build 70.951s, test binaries deferred 71.216s, cache warm 71.221s, done 71.300s. nproc 5, cgroup cpu.max 400000 100000 (four CPUs).

Commands run, with output redirected to the named logs:

```sh
timeout 90 go test ./internal/dry -run 'Test(CohereMirrors|Unlisted.*|AssembledMirrorRule|Stale.*|MissingMirrorFile)$' -v -count=1 -timeout 90s
timeout 120 python3 review/compiler/cohere-dry-guard/run-mutants.py
timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
git diff --check
```

The clone only mapped main to a remote-tracking ref. The first lane attempt could not resolve origin/cloud/merge-tree after the requested fetch; explicit refs/heads to refs/remotes/origin fetches fixed it. Lane output: lane checks 3.4 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages

Final test leaf durations and mutant evidence:

| Leaf | Seconds | Catcher |
|---|---:|---|
| TestAssembledMirrorRule | 26.64 | constant concatenation of ID |
| TestStaleMirrorHeader | 26.68 | existing file lost provenance |
| TestStaleMirrorRule | 26.90 | existing file lost rule diagnostic |
| TestMissingMirrorFile | 27.08 | listed file deleted |
| TestCohereMirrors | 28.37 | unmutated inventory |
| TestUnlistedTestMirrorRule | 18.62 | unlisted test Rule registration |
| TestUnlistedCoreMirrorRule | 18.60 | unlisted ordinary cohere ID |
| TestUnlistedAdamicMirrorHeader | 18.47 | unlisted .a header |
| TestStaleEvidenceMirrorRule | 19.32 | existing file lost pinned id-free diagnostic |
| TestUnlistedTestMirrorHeader | 17.68 | unlisted test header |
| TestUnlistedMirrorHeader | 2.74 | unlisted Go header |
| TestUnlistedMirrorRule | 2.94 | unlisted soundness ID |

The external runner applied unlisted-header, unlisted-rule and stale-entry independently to real repository input, ran only TestCohereMirrors, and restored each source. Each returned exit 1 with the intended guard diagnostic, not a build failure. Their source evidence is .go.txt, never compilable Go. Missing-file, stale-rule and the other mutants run in separate t.TempDir trees, each checked green before mutation.

No new oracle fixture or compiler behavior changed, so counts.md needs no refresh. No whole package suite or full gate was requested or run. Setup warmed builds as required. This guard recognizes static provenance and rule IDs; it cannot prove arbitrary program equivalence or discover a deliberately unlabelled semantic clone. Metadata/schema rejection paths were not exhaustively mutated.
