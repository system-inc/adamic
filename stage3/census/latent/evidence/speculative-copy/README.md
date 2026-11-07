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

Overlay generation and both binary builds pass. On the archived small.a input, the entire first-error JSONL is byte-identical to main: 1,919 bytes, SHA256 `a1711c95746b1ef78cc9403fb8021e295fc97e16887b77e9902a9f46f3f966f2`. RESULT.json records the measured byte size authoritatively. The main-equivalent overlay changes only latent_full.go back to the saved main source; its generated policies and other hooks are identical.

Acceptance is **not met**. The pinned, unadapted tsc entry, passed as one census root, completes in 36.177 seconds with **3 state-copy panic findings**, all `unsupported function field Boundary`. There are **6 units: 0 function units, 6 attempted statement units, 0 unit-status panics**. Recovered panic findings are distinct from the unit-status count. This is one entry-file record, not a completed 79-file closure audit. No result is claimed for the incomplete adapted tree.

The new private-field fix exposes the independent function-slice schema guard: ir.Function.Boundary contains mutable parameter/schema data, and latentSpecFunctionSchema does not list it among its explicit copies. The speculative synthetic fixture exits 2 at that guard before any file record or unit completes (1 unhandled state-copy panic). The original audit_deferred.py stops earlier in legacy mode on both main and the fix: its called generic lacks an instantiation witness because lowering reports no import-declaration route. Both failed audits and the separate speculative failure are retained; no assertion was suppressed.

Full setup succeeds. The complete adaptation was stopped at 70-readonly-views once the independent guard was confirmed. No adapted-corpus result or zero-panic claim is made. All task-started processes have exited or were stopped. The implementation remains the one code commit; this follow-up records evidence only.

Validation raw inputs, generated overlays, observations, metrics and logs are packed in validation-results.tar.gz. validation-results.sha256 lists unpacked members; binaries.sha256 identifies the two executables. RESULT.json contains exact scope and counts. The tsc run used:

```sh
LATENT_ASSERT_NO_OUTPUT=1 LATENT_SPECULATIVE=1 GOMEMLIMIT=6GiB python3 stage3/census/speculative/timed_run.py 1200 8192 /workspace/scratch/9k-tsc.metrics.json /workspace/scratch/9k-tsc.log /workspace/scratch/9k-census /workspace/typescript-source/src/tsc/tsc.ts /workspace/scratch/9k-tsc.jsonl
```

Every command's stdout/stderr went to a named log. The timed run exited 0 because its three panic findings were recovered; that exit does not satisfy the zero-panic condition.
