A newline moves the whole-file target before any write

Task #redxes2, head 1.

Mutation: Append one newline byte to sys.ts before invoking the adapter. prepare.py checks that both mutant bytes and nanosecond mtime stay unchanged until restoration.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/proofs/../adapt.cjs /workspace/scratch/wasm-crypto-proof/after
```

Exact observed output:

```text
AssertionError [ERR_ASSERTION]: target source moved: sys.ts
```

Every on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.
