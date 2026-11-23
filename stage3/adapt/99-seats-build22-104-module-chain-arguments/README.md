Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, family 104); Compiler accepts the head after its review and census.

The Nov 22 meter reported a saved-tuple spread in
`createModuleNotFoundChain` at `utilities.ts:848:13`. The v6.0.3
baseline at ac44317c builds the Node10 diagnostic arguments from
`[alternateResult] as const` and the other resolution kind's arguments
from a two-element literal marked `as const`. Neither tuple escapes or is
written. The saved pair is tested for truthiness and read only to pass
its diagnostic and spread its strings into the rest parameter of
`chainDiagnosticMessages`.

The direct calls pass one and two strings respectively, as Adamic ruled
under [rulings.md row 220](../../../docs/rulings.md#L220). The
`alternateResult` truthiness test still sends an empty string to the
existing fallback. The same comparison selects Node10, and the
alternate-result path computes the mangled package name only when the
other resolution kind supplies a path under `node_modules/@types/`.

Run the test with the pinned API packages and a tree that
`stage3/apply.sh` wrote. It reads this adapter's before and after
snapshots from its `refs/stage3/` refs and runs the adapted function
beside the original.

```sh
adapter=stage3/adapt/99-seats-build22-104-module-chain-arguments
NODE_PATH="$PWD/stage3/api/node_modules" \
    node "$adapter/test.cjs" "<applied tree>"
```
