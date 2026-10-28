# TestLatentStatecopySharesSourceIdentities: the generator mutant

The latent run shares one `sourceIdentities` map across every lowering of the program (lower.go.txt's latentRun,
#pv935yw). The state copier passes that field through instead of deep-copying it, and the guard skips it as it skips
`impliedTargetQuery`.

Mutant, in statecopy/main.go: the generator's `sourceIdentities` case made `else if false`, so the copier deep-copies the
map again.

`go test ./stage3/census/latent/statecopy -run TestLatentStatecopySharesSourceIdentities -count=1 -v`, on macOS:

```
    guard_test.go:126: the lowering copier deep-copies sourceIdentities, which every lowering of a latent run shares
--- FAIL: TestLatentStatecopySharesSourceIdentities (0.04s)
```

Restored, `go test ./stage3/census/latent/... -count=1` is green: importclean, refusalrewrite, statecopy, statementrewrite.
