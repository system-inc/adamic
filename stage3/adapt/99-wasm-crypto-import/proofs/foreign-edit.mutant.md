A foreign 99j edit is accepted without changing the owned result

The exact sys.ts site was read from `typescript/axolotl-99j-2`, commit `e0063e177deb458b92993ac872fdcd3cf1fa7fa3`. foreign-site.json records its source and statement. In invokeCallbacks, before our adapter:

```ts
(existing as string[]).push(...fileNames);
```

becomes:

```ts
if (Array.isArray(existing)) existing.push(...fileNames);
```

Command on the isolated before-tree copy:

```sh
node stage3/adapt/99-wasm-crypto-import/adapt.cjs /workspace/scratch/crypto-scoped-proof/probe
```

Exit 0, files=1. The exact changed callback statement survives, and the complete output equals main's crypto-adapted sys.ts plus only that foreign edit. The owned after hashes are unchanged:

- ownerAfterSha256: `9d0f66d64f8d6c9a713b73b514429f485254c918e26791dae9540331a4aa357d`
- importsAfterSha256: `c24afbb7654f6cd3057985c10558464846d134085b4867626146a199a505c62f`

scoped.py restores the before bytes in finally. This is a positive composability probe, not a surviving unsafe mutation. logs/scoped-foreign-edit.log.gz retains the full adapter output.
