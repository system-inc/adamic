# Carry facts mutant

Guard: `TestLatentCarryFactsRequiresEveryLoweringFieldPolicy`, beside `statecopy/guard_test.go`'s existing guards. It calls `t.Parallel()`, runs no Go or compiler subprocess, and its green run took 0.099 s.

Mutant planted: remove the `"entriesRecords": programFact` entry from `loweringFieldPolicies` in `statecopy/main.go`, leaving the production lowering field in place. Command:

```sh
go test ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsRequiresEveryLoweringFieldPolicy$' -count=1 > /workspace/scratch/carry-final-guard-mutant.log 2>&1
```

The test went red with:

```text
guard_test.go:207: lowering.entriesRecords has no census memo/copy policy
```

The original table was restored in `finally`, and the whole `go test ./stage3/census/...` run subsequently passed. The guard also checks capture/install assignments and that the carried map keeps an owned typed snapshot copy. See [guard-mutant.log.txt](guard-mutant.log.txt) and [guard-green.log.txt](guard-green.log.txt).

An earlier Boolean-only version was separately proven red by deleting `entriesProgramRecords`; the final map-deletion mutant above is the committed version's proof.
