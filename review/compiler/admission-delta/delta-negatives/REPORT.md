Built SHA-256 ruled-negative witness support with exact stop validation and pinned type-correct repairs.
Delivery: compiler/admission-delta, based on 841e335c with current main 98008bbb merged first; final SHA is reported by the worker.
Commands: targeted tool leaves pass; TestCallTargetReaders passes in 19.96s; tool package vet passes; final committed lane check follows.
Mutants: type-correct input with a retained stopping backend is rejected; wrong exit/message, source edit, invalid declaration/repair and forbidden valid hash are rejected.
Not covered: general type-lie inference or imported witnesses, arbitrary runtime contracts, the full gate, or other platforms.

This implements #936mrbr's admission-delta ruling toward Outcome 24. The ledger uses a valid JSON `_comment` header rather than invalid JSON comments. New entries require @system_adamic's ruling. Its schema is documented in that header and cmd/adamic-admission-delta/README.md. The map uses SHA-256 of exact source bytes, not the filename or the corpus Git blob. Names are documentary. A changed source has no exemption. Valid repair hashes cannot occur as witness keys; the tool independently validates the ruled literal contradiction at its declaration and the repair hash, rather than asking TypeScript to prove its own assertions. The supported forms are deliberately bounded; other lies need a validator plus a ruling. Imported witnesses are rejected, keeping content identity self-contained. The eight delivered sources contain no imports or exports.

Ordinary agreement remains exactly:

```go
return a.Error == "" && b.Error == "" && c.Error == "" && a.Exit >= 0 && a.Exit == b.Exit && a.Exit == c.Exit && a.Stdout == b.Stdout && a.Stdout == c.Stdout
```

It compares stdout and exit code, not stderr. Negative witnesses additionally require source Node exit 0, exact ruled exit 70 and whole stderr in each backend, with matching backend stdout. Compilation failures and timeouts never qualify. JSON keeps `agree=false` separate from `negative_witness=true`; it records agreeing/witness counts, and stderr prints list size including zero. The invoking repository supplies the ledger, enabling slice 5 to depend on this executable without merging the unlanded tool branch.

The changed legacy unsupported-view probe is now marked `a-check: refused unsupported dictionary contract`, matching current main's sound refusal. The first sourced lane check exposed this stale inherited expectation; no compiler behavior was changed.

Validation commands (all outputs retained here):

- `go test ./cmd/adamic-admission-delta -run '^(TestNegative.*|TestNewAdmissionMismatch|TestHeadInputIsolation|TestManifestMismatch|TestNewlyRefusedIsInformational|TestEmptyCorpus|TestOutputsAgree|TestBuildRevisions|TestDiffCoverage|TestShardCommandScope)$' -count=1 -json -timeout 90s`
- `go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s`
- `go vet ./cmd/adamic-admission-delta`
- Required integration lane command after commit, with the toolchain environment sourced.

The ten new top-level leaves and touched existing command leaves are recorded in leaf-seconds.json; every leaf is below 60s. TestNegativeTypeCorrectMutant replaces `undefined!` with `"ready"` in a committed fake-compiler input while both backend artifacts still stop: Node exits 0, both backends exit 70, and the tool returns fail with zero accepted witnesses. The independent real slice-5 source repair and retained-artifact mutant is recorded on compiler/chain-slice-5. No oracle fixture was added, so counts.md requires no new row.

Setup: initial shared-cache run was stopped; rerun with ADAMIC_GOCACHE_OFF=1 passed. GOPROXY=https://proxy.golang.org|direct; environment /workspace/adamic-tools/env.sh. Timing lines: go/node 0.021s; submodules 0.060s; markdown 0.066s; clang 0.146s; shared cache off 0.148s; go build 166.134s; test binaries deferred 166.263s; cache warm 166.264s; done 166.293s. nproc=5, cgroup CPU quota=4. Full setup.log retained.

The real slice-5 census with this tool passes on main 98008bbb: 30 newly admitted, 22 ordinary agreements, eight listed witnesses, zero omissions and zero compiler failures. Source Node and both backends for factory-use.a print `true\n` and exit 70; it is not listed. The type-correct speculative repair prints `1\n` and exits 0 on all three clean implementations. Supplying the original stopping artifacts for that repaired source makes the tool fail: its hash is unlisted, `agree=false`, `negative_witness=false`, Node stdout `1\n`/exit 0 versus empty stdout/exit 70 in both backends. Complete observations belong to the separately delivered slice branch. Runtime layout and compiler production behavior are unchanged by this tool unit. No cohere source was copied.

Mandatory repair rechecks: admission-only sampling could skip a repaired placeholder if main already admitted it. The final tool schedules listed content, pinned repair hashes and admitted edits at declared witness paths even when classed accepted-by-both, under any budget. Paths grant coverage only; exceptions remain content-hash-only. TestNegativeAlreadyAcceptedRepairMutant proves that source Node exit 0 versus retained backend exit 70 fails even with zero new admissions; TestNegativeEditedWitnessMandatory proves an admitted edit cannot be sampled away. Both are separate parallel leaves and are recorded in leaf-seconds.json.
