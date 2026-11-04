Owner-local moved-target evidence

Mutation: in getNodeSystem, change `const isMacOs = process.platform === "darwin";` to use `"darw1n"`. The old newline-outside-owner probe is intentionally no longer a target movement.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/adapt.cjs /workspace/scratch/crypto-scoped-proof/scoped
```

Exit 1, exact red:

```text
AssertionError [ERR_ASSERTION]: uninventoried owner: getNodeSystem
```

The rejected call leaves mutant bytes and nanosecond mtime unchanged. scoped.py restores the exact saved bytes in finally; the full output is logs/scoped-moved-owner.log.gz. prepare.py now produces this same owner-local mutant.
