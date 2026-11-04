A changed node: import block is rejected before writing

Append an unused module-scope `import * as cryptoScopedProbe from "node:util";` to the adapted file, leaving getNodeSystem unchanged.

Exact red:

```text
AssertionError [ERR_ASSERTION]: uninventoried node: import block
```

The call leaves mutant bytes and nanosecond mtime unchanged. scoped.py restores the original file in finally; full output is logs/scoped-moved-imports.log.gz.
