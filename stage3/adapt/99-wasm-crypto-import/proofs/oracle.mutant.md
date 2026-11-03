An extra baseline line is caught by the real lane byte checker

Task #redxes2, head 1.

Mutation: Append an extra baseline mutant line to the completed oracle-after/baseline.diff, leaving both API snapshots unchanged. Both api_check and the real lane CLI must reject it.

Command:

```sh
CRYPTO_PROOF_SCRATCH=/workspace/scratch/wasm-crypto-proof python3 stage3/adapt/99-wasm-crypto-import/proofs/oracles.py
```

Exact observed output:

```text
baseline.diff does not exactly describe the API snapshots (extra, missing, or changed diff lines)
FAIL stage3 landing lane: baseline.diff does not exactly describe the API snapshots (extra, missing, or changed diff lines) [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
