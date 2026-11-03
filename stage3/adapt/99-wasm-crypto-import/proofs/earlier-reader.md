98-e is an earlier reader, not a later pass

The official apply runs 98-e before 99-wasm-crypto-import. Its store guard at `stage3/adapt/98-axolotl-e/adapt.cjs:141` reads the conditional property that this adaptation removes. Running it alone on the final output is intentionally rejected. This is its input moving, not a defect in its guard.

```sh
node stage3/adapt/98-axolotl-e/adapt.cjs /workspace/scratch/wasm-crypto-proof/after
```

Exit 1. Exact red:

```text
the nodeSystem store: want one, got 0
```

`apply.py` sorts at line 272, then executes its sole non-setup adapter loop at lines 315-321. There is no post-adaptation pass. The separate 00-setup generation occurs before the pristine snapshot. `order.json` records the complete order. No adapter sorts after this one, and no shared runner was changed. The failed standalone rerun writes no source bytes; the restored whole-source check passes immediately afterwards.
