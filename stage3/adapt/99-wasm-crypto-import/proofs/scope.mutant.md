An erased type edit is caught only by the exact source proof

Task #redxes2, head 1.

Mutation: Add | null to clearTimeout's private ambient parameter type. The command first passes both transpile comparisons, then fails exact source undo.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/proofs/scope-mutant.cjs /workspace/scratch/wasm-crypto-proof/control /workspace/scratch/wasm-crypto-proof/after
```

Exact observed output:

```text
PASS TestCryptoTranspileScope: the extra type-only edit is invisible in both JavaScript comparisons
AssertionError [ERR_ASSERTION]: TestCryptoSourceScope: edit outside the crypto import, removed optional load and three hash expressions
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
