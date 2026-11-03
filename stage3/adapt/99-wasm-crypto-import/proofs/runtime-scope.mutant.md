An unrelated terminal-width runtime edit is outside the adaptation

Task #redxes2, head 1.

Mutation: Replace return process.stdout.columns ?? 0; with return process.stdout.columns ?? 1; before the check, as prepare.py does.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/proofs/check.cjs /workspace/scratch/wasm-crypto-proof/control /workspace/scratch/wasm-crypto-proof/after
```

Exact observed output:

```text
AssertionError [ERR_ASSERTION]: TestCryptoSourceScope: edit outside the crypto import, removed optional load and three hash expressions
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
