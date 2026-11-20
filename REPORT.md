# #837zjkt paired compiler census

## 1. Getter casts in the full utilities census

Every row has the same `Refused` reason, `a cast the runtime can't check`, and the same full diagnostic at base and head.

> Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)

| Adapted `utilities.ts` site | Getter | Cast operand | Base `f48a0df1` | Head `181888a5` |
| --- | --- | --- | --- | --- |
| 8555:82 | `getNodeConstructor` | `Node` | Refused in `refusal_scan` and `lowering` | Refused in `refusal_scan` and `lowering` |
| 8556:32 | `getTokenConstructor` | `Token` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8557:37 | `getIdentifierConstructor` | `Identifier` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8558:44 | `getPrivateIdentifierConstructor` | `Node` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8559:37 | `getSourceFileConstructor` | `Node` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8560:33 | `getSymbolConstructor` | `Symbol` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8561:31 | `getTypeConstructor` | `Type` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8562:36 | `getSignatureConstructor` | `Signature` | Refused in `refusal_scan` | Refused in `refusal_scan` |
| 8563:42 | `getSourceMapSourceConstructor` | `SourceMapSource` | Refused in `refusal_scan` | Refused in `refusal_scan` |

Both streams report 933 units and 1,162 finding records. Pairing found zero new, gone, changed or reached findings in the full file. [Base utilities](evidence/base-utilities.jsonl.gz) · [Head utilities](evidence/head-utilities.jsonl.gz)

## 2. Symbol

Head leaves Symbol's road refused at the getter cast on `8560:33`. The adaptation 41 expression is `Symbol as typeof Symbol & ReturnType<ObjectAllocator["getSymbolConstructor"]>`. Its own stop is the full unchecked-cast diagnostic in section 1. Within the `objectAllocator` initializer unit at `8554:1`, the first lowering stop on both commits is the earlier `getNodeConstructor` cast at `8555:82`, with those same words. The full refusal scan reaches Symbol's cast independently.

The separate `Symbol` function unit begins at `8473:1` with `this: Partial<Symbol> & { links?: TransientSymbol["links"] | undefined; }`. Both compilers attempt this unit and record `NotYet` in `lowering` at `8473:17`. Its full first-stop diagnostic is

> stage 0 can't lower a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented) yet

Both streams also record a `Boundary` for the same stop at the unit location with reason `function signature or prologue failed; continued at next declaration`. [Base utilities](evidence/base-utilities.jsonl.gz) · [Head utilities](evidence/head-utilities.jsonl.gz)

## 3. First-error entry census

The checker-clean `tsc.ts` entry closure reached 81 source files on each compiler. Each stream has 4,568 unit records, including 4,064 attempted units and 504 units deferred to generic instantiation. Each has 3,098 finding records.

| Grade or regression | Count |
| --- | ---: |
| New findings | 0 |
| Gone findings | 0 |
| Changed findings | 0 |
| Reached findings | 0 |
| Units that lowered at base and stop at head | 0 |
| New panics | 0 |

The entry regression count is **0**. The two scoped tail findings are unchanged, and neither entry stream contains a panic. [Base entry](evidence/base-entry.jsonl.gz) · [Head entry](evidence/head-entry.jsonl.gz)

## Pins and reproduction

| Role | Fetched branch | Measured commit |
| --- | --- | --- |
| Base | `main` | `f48a0df1a5a26171252cfa1b77b30334bee0fdae` |
| Slice 1 | `compiler/function-constructor-allocator-7` | `1da3e5922b2195b7cbd998e308e154b2c14e1ddb` |
| Head | `compiler/function-constructor-intersection-8` | `181888a59a0c08ebbff251236a3640734adbc251` |

Fetched main pointed to `8f91fd0b630a746e9ba6b33b6bd2d095609d0cfa`. The measured base is slice 1's parent and an ancestor of that main tip. Slice 1 is head's parent.

Each detached worktree initialized the recursive submodules at cohere `5360753e19e4417675a90e92a1823bea66b13918` and checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. The checker download was retried with the meter's shallow submodule update using a local clone of the exact gitlink. After the base submodule update and `npm ci --ignore-scripts --no-audit --no-fund` in `stage3/api`, the base `stage3/apply.sh` produced the shared adapted tree. The TypeScript input came from the meter's shallow clone of the `v6.0.3` tag in `stage3/source.json` and resolved to `050880ce59e30b356b686bd3144efe24f875ebc8`.

Go 1.27.1 built each pin's own latent census and meter entry overlay and its own instrumented scan-all overlay. The streams below are from the two latent passes.

```sh
LATENT_TIMES="$out/entry-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 \
  CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_TAIL=1 \
  "$bin/entry-census" "$adapted/src/tsc/tsc.ts" "$out/entry.jsonl"

LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
  LATENT_ONLY_FILE="$adapted/src/compiler/utilities.ts" \
  "$bin/latent-census" "$adapted/src/compiler" "$out/utilities.jsonl"
```

The utilities driver loaded the compiler program and attempted the selected file in full mode. The entry driver used the `tsc.ts` import closure and first-error mode with its scoped tail enabled. Each measurement process used `GOMAXPROCS=2`.

The adapted `utilities.ts` SHA-256 is `20fea160988c2a5552e835e8e8c567ed7b6594da5d51303010affea982c5141a`. All three compilers read that same file in the same adapted tree.

I paired the entry streams and the utilities streams with the base's `stage3/census/pair/pair.py`, using the stock TypeScript 6.0.3 parser installed from the base lockfile. I also compared the finding multiset at each source site and owner unit by kind, phase, reason and full message with positions stripped. Both comparisons show zero deltas in each pass. The entry regression check includes the scoped tail.

Slice 1 was measured on the shared tree as a control. Its full utilities and entry finding signatures match base and head. [Slice 1 utilities](evidence/slice-utilities.jsonl.gz) · [Slice 1 entry](evidence/slice-entry.jsonl.gz)
