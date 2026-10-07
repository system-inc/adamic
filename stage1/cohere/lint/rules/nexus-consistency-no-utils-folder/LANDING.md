# No utils folder filename fix validation

Rebased onto the shared filename-transport fix
59451e23eefc9d83102afd6d1a26050d6c447227. The existing witness is now
testdata/src/utils/format.ts.txt. Rule code, messages, options adapter and
mutant are unchanged. No shared files were edited.

Commands:

    go run ./cmd/lint-registry
    gofmt -l stage1/cohere/lint/rules/nexus-consistency-no-utils-folder/oracle.go
    go vet ./...
    go test ./stage1/cohere/lint -run '^Test(OwnedWitnesses|Mutants|RulesAgree)$' -count=1 -v -timeout 30m

All commands passed. Gofmt and vet produced empty logs. The full three-test
run is green. TestRulesAgree compares 13,072,570 identical bytes across Go,
source Node, emitted JavaScript and sanitized native. Its captured corpus now
preserves all eight paths from both upstream tests selected by
TestConsistencyNoUtilsFolder. TestOwnedWitnesses reports 113,755 identical
bytes and requires a nonzero upstream finding for the relocated witness.

The complete discovered mutant suite passes. The owned
no-utils-folder-utils-segment-omitted mutant builds and runs successfully, and
ordinary Go byte comparisons catch it on source Node, emitted JavaScript and
sanitized native at case 145, line 2738, for src/utils/format.ts. The same
mutation previously survived when the harness removed directory segments.

Exact test timings and the full intact log are in evidence/filename-fix.
BLOCKER_HISTORY.md and the earlier report/logs preserve the previous failure
and independent real-path proof. They are historical, not current blockers.
The full repository gate remains unrun. No main or area ref was pushed.
