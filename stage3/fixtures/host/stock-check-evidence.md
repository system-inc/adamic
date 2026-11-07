# h2c34ad evidence

Base: af42d3dc1261a3db536064c3dfbb5eaa7e8b6405, fetched origin/main.
Stock TypeScript 6.0.3; bundled @types/node 25.3.3. The prelude has no process
declaration. Console is rewritten by parsed identifier and statement span;
process is rejected if reintroduced, rather than silently retaining its type.

Rewritten prelude SHA256, before:
`f69ba5c76f59175f2b028660ae0d30cbca3934133e7993840485793626f505ae`

Rewritten prelude SHA256, after (also the four-space mutant):
`f69ba5c76f59175f2b028660ae0d30cbca3934133e7993840485793626f505ae`

| Input | Old code | New code |
| --- | --- | --- |
| Today's prelude | exit 0, all 25 rows agree | exit 0, all 25 rows agree |
| Console tabs changed to four spaces | not needed | identical rewritten hash; exit 0, all 25 rows agree |
| Console declaration removed | exit 0, all 25 rows agree silently | exit 1: expected exactly one console declaration, found 0 |
| Console declaration duplicated | not needed | exit 1: expected exactly one console declaration, found 2 |
| Process declaration added | not needed | exit 1: stock check would otherwise type-check against the prelude's process instead of Node's and the rewrite must come back |

The documented direct invocation passed, as did its existing --mutant variant
which caught TS2322 for _os.EOL assigned to number:

```
CENSUS_TYPESCRIPT=/tmp/h2c34ad/api/node_modules/typescript/lib/typescript.js node stage3/fixtures/host/stock-check.cjs "$PWD" /tmp/h2c34ad/final-direct
```

Small sibling mutant files transform the real prelude, preserving all unrelated
bytes. stock-prelude.mutant.cjs executes the real script with only its prelude
read intercepted. Mode hash executes the script through its rewrite and hashes
the resulting text; mode full executes its entire fixture loop. Reproduce with
CENSUS_TYPESCRIPT pointing at a 6.0.3 TypeScript API and STOCK_CHECK_BEFORE at
stock-check.cjs extracted from the base commit:

```
node stage3/fixtures/host/stock-prelude.mutant.cjs old real hash /tmp/stock-old-hash
node stage3/fixtures/host/stock-prelude.mutant.cjs new real hash /tmp/stock-new-hash
node stage3/fixtures/host/stock-prelude.mutant.cjs new spaces hash /tmp/stock-spaces-hash
node stage3/fixtures/host/stock-prelude.mutant.cjs old removed full /tmp/stock-old-removed
node stage3/fixtures/host/stock-prelude.mutant.cjs new spaces full /tmp/stock-spaces
node stage3/fixtures/host/stock-prelude.mutant.cjs new removed full /tmp/stock-removed
node stage3/fixtures/host/stock-prelude.mutant.cjs new duplicate full /tmp/stock-duplicate
node stage3/fixtures/host/stock-prelude.mutant.cjs new process-added full /tmp/stock-process
```

Runs used file logs under /tmp/h2c34ad and wrappers reporting child exit codes.
No Go build ran. Read stock-check.cjs whole and searched stage3 and internal.
No stage3/fixtures test runs this script: fixtures_test.go's TestFixturesHost
runs the fixture directory oracle, and the host configuration/status tests run
that Go fixture runner. Therefore no Go test was added. The actual executable
caller internal/hostconfiguration/readers_test.cjs runs stock-check in a VM;
configuration_test.go supplies CENSUS_TYPESCRIPT. Host ADAPTED.md documents the
direct invocations used above.

Existing CJS caller harness: exit 0, all nine reader cases passed.
