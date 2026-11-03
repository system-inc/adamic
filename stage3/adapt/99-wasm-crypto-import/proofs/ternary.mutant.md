Leaving the crypto conditional is a bare namespace value

Task #redxes2, head 1.

Mutation: Replace createHash: createSHA256Hash, with createHash: _crypto ? createSHA256Hash : generateDjb2Hash,. The command plants it on disk and calls the semantic guard directly, bypassing the hash gate.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/proofs/ternary-mutant.cjs /workspace/scratch/wasm-crypto-proof/after
```

Exact observed output:

```text
AssertionError [ERR_ASSERTION]: collision or bare namespace use: _crypto
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
