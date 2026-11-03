Three independently bound negative programs

Task #redxes2, head 1.

Mutation: Three separate virtual programs: append const _crypto = {}; add an _crypto parameter to createSHA256Hash; append void _crypto;. The whole-file hash gate is bypassed so these exercise the semantic binding guard itself.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/proofs/bindings.cjs /workspace/scratch/wasm-crypto-proof/control /workspace/scratch/wasm-crypto-proof/after /workspace/scratch/adapt-94/stage3/adapt/99-wasm-crypto-import/proofs/bindings.json
```

Exact observed output:

```text
CAUGHT module collision: collision or bare namespace use: _crypto
CAUGHT enclosing parameter shadow: collision or bare namespace use: _crypto
CAUGHT bare namespace value: collision or bare namespace use: _crypto
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
