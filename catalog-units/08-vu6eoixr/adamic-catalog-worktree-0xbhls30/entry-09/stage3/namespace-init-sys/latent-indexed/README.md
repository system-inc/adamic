This unit measures the 99 `noUncheckedIndexedAccess` ledger rows on compiler baseline `5ad36d2cf4fce475ec072f3cc68e21c393dffc45`. It changes no production compiler source or refusal. The existing latent census tool is extended in a scratch Go overlay, with its production `Lower` and ordinary loader output paths disabled.

The source input is the exact 79-file adapted TypeScript tree used in `../production-final`, produced from adaptation commit `3b25512566206bc603b93264e8d072c55e075d64` and upstream TypeScript commit `050880ce59e30b356b686bd3144efe24f875ebc8`. `classify.py` verifies every file against `stage3/stricter-indexed-all/evidence/ledger-source-hashes.json`. The compiler project's own options stay in effect: this does not force the stricter option onto the production checker. All 99 indexed sites must have scheduled dispositions. Remaining option errors are retained, without the previous production probe's exclusion overlay.

Run from the repository root with the prepared adapted source tree at `/tmp/namespace-init-sys-typescript`:

```bash
source /workspace/adamic-tools/env.sh
python3 stage3/namespace-init-sys/latent-indexed/build.py --scratch /tmp/namespace-init-sys-latent-v2 > /tmp/namespace-init-sys-latent-v2-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/namespace-init-sys-latent-v2/latent-census /tmp/namespace-init-sys-typescript/src/compiler /tmp/namespace-init-sys-latent-v2/raw.jsonl > /tmp/namespace-init-sys-latent-v2-run.log 2>&1
python3 stage3/namespace-init-sys/latent-indexed/classify.py --raw /tmp/namespace-init-sys-latent-v2/raw.jsonl --source /tmp/namespace-init-sys-typescript --output stage3/namespace-init-sys/latent-indexed/evidence > /tmp/namespace-init-sys-latent-v2-classify.log 2>&1
python3 stage3/namespace-init-sys/latent-indexed/audit_rows.py stage3/namespace-init-sys/latent-indexed/evidence /tmp/namespace-init-sys-typescript > /tmp/namespace-init-sys-latent-v2-row-audit.log 2>&1
python3 stage3/census/latent/audit.py /tmp/namespace-init-sys-latent-v2/latent-census > /tmp/namespace-init-sys-latent-v2-audit.log 2>&1
python3 stage3/namespace-init-sys/latent-indexed/audit.py /tmp/namespace-init-sys-latent-v2/latent-census > /tmp/namespace-init-sys-latent-v2-indexed-audit.log 2>&1
python3 stage3/namespace-init-sys/latent-indexed/audit_guard_mutant.py /tmp/namespace-init-sys-latent-v2 /tmp/namespace-init-sys-latent-v2-ir-mutant > /tmp/namespace-init-sys-latent-v2-ir-mutant.log 2>&1
python3 stage3/namespace-init-sys/latent-indexed/audit_eligibility_mutant.py /tmp/namespace-init-sys-latent-v2 /tmp/namespace-init-sys-latent-v2-eligibility-mutant > /tmp/namespace-init-sys-latent-v2-eligibility-mutant.log 2>&1
python3 stage3/census/latent/audit_output_guards.py /workspace/adamic /tmp/namespace-init-sys-latent-v2 /tmp/namespace-init-sys-latent-v2-guard-mutants > /tmp/namespace-init-sys-latent-v2-guard-audit.log 2>&1
```

The builder adapts stale tool assumptions about the current loader return, compiler entry collection, and record/discriminant refusal returns. It preserves the original production `refuse` function. The adapted refusal collector's own tests run in `/tmp/namespace-init-sys-latent-v2/refusalrewrite` with `GOWORK=off go test -count=1 ./...`.

Scheduled JSON diagnostics are excluded from checker-body skip eligibility, matching production. Unresolved ordinary diagnostics and the five remaining option errors keep the existing body skip policy. The `.a` audit fixture is materialized as a scratch `.ts` input solely to exercise the owning project profile.

Constructing a guard is insufficient evidence of emission. The tool records indexed-read attribution but counts only panic-bearing guards surviving in `ir.InsertedChecks` after an isolated attempt. Function wrappers record the original error and rethrow the original panic. They do not retry a nested function, fabricate its lexical environment, or relax a refusal. The smallest enclosing function owns each ledger row. An actual function failure takes priority; otherwise a direct syntax refusal establishes a local blocker. Without either, a surviving guard qualifies as emitted. Other rows retain the enclosing stop and its exact location/message as their not-reached reason.

The counted table assigns one primary blocker to each blocked/unreached row. It does not enumerate every alternative refusal in that function; the compressed raw census retains all observed findings. Generic declarations are measured as written. Module ordering, final ownership checks, and both native backends are outside the latent measurement.
