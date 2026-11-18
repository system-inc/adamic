# Last queried position proof (#tdwte64)

The recorded starting tip is `46d8f9a44c7d6c7dcbc9f82cab3cf8742f835bc5`. The candidate is its direct child `4c2e15fbf94edb63565ee5e9ed3a1f8c74c1a1bd` and uses TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Origin/main advanced to `17a200299adcdbafbf0bce731870a28e52f50a21` during measurement. The receipts below use the recorded starting tip throughout.

1. PASS. Stock tsc 6.0.3 reported zero diagnostics on all 79 roots before and after. Both `adamic types` runs exited 0.
2. PASS. The full CommonJSBundle parser census removed the write stop at `src/compiler/parser.ts:10467:39` and recorded zero new findings. Both census children and the comparison exited 0.
3. PASS. The restoring mutant brought the write stop back. Its complete census matched the baseline events and unit statuses. The mutant census and comparison exited 0.
4. PASS. Whole `stage3/apply.sh` and `TestProduct_Stage3AdaptedTree` exited 0. The product's first cold run hit the Go test default timeout and exited 1. The retry used a 45-minute timeout.
5. PASS. The whole lane exited 0 with 106,366 passing, one sanctioned API failure, and zero pending. The upstream oracle exited 1 for the sanctioned failure. The shard plan, manifest, and generated Go declarations matched their input hashes.
6. PASS. The public `typescript.d.ts` files were byte-identical. The comparison exited 0.

## Stops reached in previously blocked units

- `src/compiler/parser.ts:1253:17` in `src/compiler/parser.ts:10432:5`. a member typed any (167(0)) related to QualifiedName at the assertion site

The census wrappers recorded their exact environment and arguments in `.log.command` and their actual child exits in `.log.exit`. Large logs and streams are gzip-compressed.
