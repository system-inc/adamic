# Speculative private IR state copy

Base: 89c1c7958 (origin/main at task start).

The speculative struct copy starts at zero, consults the generated private-field policy, and shallow-copies only exported fields before cloning mutable fields. Both copy modes share pointers selected by latentSharedIRPointer. Unknown private IR fields are still refused by loadGenerator before the overlay is published.

The whole statecopy package passes (26.078 seconds). TestLatentSpeculativePrivateIRFields compiles the actual copier and generated policies against real IR. It populates Program.methodTables, checks the copy starts empty while the original stays populated, and checks GeneratedType pointer identity. Its build-refusal subtest requires Program.addedLater and its declaration location in the refusal.

The saved old-branch-mutant.go.txt is main's full.go.txt. Selecting it with LATENT_STATECOPY_MUTANT makes the new case fail with unexported IR field methodTables (exit 1). Unit logs are packed in unit-results.tar.gz; unit-results.sha256 lists the unpacked files.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
go test -count=1 ./stage3/census/latent/statecopy
LATENT_STATECOPY_MUTANT="$PWD/stage3/census/latent/evidence/speculative-copy/old-branch-mutant.go.txt" GOWORK=off go test -count=1 -run '^TestLatentSpeculativePrivateIRFields$' ./stage3/census/latent/statecopy
```

Overlay build, first-error identity and project audit evidence will be recorded separately.
