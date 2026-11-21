# Checked write trace for #rgrtt32

The head rejects the source assignment `symbol.valueDeclaration = node` at `src/compiler/utilities.ts:4315:9` while registering a predicate view. It visits the assignment as an AST node in other census units. Lowering the assignment's own function, `setValueDeclaration`, produces the same unchecked store on both revisions.

| Census scope | Main `49ecfe75` | Head `fcc0f8da` |
| --- | --- | --- |
| Affected units traced | The registration walk judges the field unchecked. `CheckedField(9351, "valueDeclaration")` is false. The units emit zero stores at this site, so the store receiver has no reach result in these units. | The registration walk finds `CheckedField(9351, "valueDeclaration")` true and returns the optional-object-field `NotYet`. The units emit zero stores at this site. The rejecting view has yet to append its origin to `ViewOrigins`. There is no reach result for the write. |
| `setValueDeclaration` as its own unit | One `ir.SetProperty` with `ViewReceiverTypeID:9351`, `Site:1`, and `DeclaredType:"Declaration \| undefined"`. `CheckedField` is false. Receiver reach is **Unknown**, with zero allocation IDs and zero viewed allocation IDs. | Same store, check and reach result. |

The static type ID is present on the emitted store even with `CheckedField=false`. That ID keys the check table. The independent setter unit has no tracked producer for the `symbol` parameter. Its Unknown result cannot establish caller-specific allocation reach.

At the head, **`(*lowering).recordSourceSlotWrite` returns the `NotYet` at `internal/lower/source_slot_certificate.go:114`**. The stack goes through `predicateRefusal`, `view`, `viewWalkIndexed`, then `judgeViewAccess`, which calls `recordSourceSlotWrite`. `viewSchema` has already registered the check. `view` returns the error before appending the new origin. This path uses the checker type and the declared field type, before allocation-flow analysis. A whole-program reach answer for this parameter would need its lowered callers and arguments in the same IR graph. [Return site](https://github.com/system-inc/adamic/blob/fcc0f8dab52790fe7a36a203cb100f96affab4c8/internal/lower/source_slot_certificate.go#L96-L114) · [View registration and origin append](https://github.com/system-inc/adamic/blob/fcc0f8dab52790fe7a36a203cb100f96affab4c8/internal/lower/interface_cast.go#L90-L118)

## Reproduction

I used detached worktrees at `49ecfe7541f1317e8bde90b469bd09a9b4b28591` and `fcc0f8dab52790fe7a36a203cb100f96affab4c8`, with recursive submodules at their recorded pins. The requested head SHA is reachable through `compiler/view-field-unsupported-4`. Both compiler pins have identical Stage 3 sources. I cloned TypeScript v6.0.3 as in `stage3/meter/twice-daily.sh`, verified the upstream pin `050880ce59e30b356b686bd3144efe24f875ebc8`, installed the locked `stage3/api` dependencies and ran main's `stage3/apply.sh` once.

Each census binary used its pin's `make_overlay.py` and `entry_overlay.py`. The [diagnostic patch](trace-overlay.py) applies to the generated entry overlay and keeps the original compiler files untouched. The selector limits independent first-error unit attempts to the named functions while keeping the full checker program and the census's normal dependency work. Both runs reported 81 entry-reachable sources and zero checker diagnostics. Allocation numbering and reach run on a disposable snapshot after the unit produces a store.

```bash
# From each detached compiler worktree. overlay_root is a scratch directory for that revision.
python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay_root/latent"
gofmt -w "$overlay_root"/latent/*.go
python3 stage3/meter/entry_overlay.py "$overlay_root/latent" "$overlay_root/entry"
python3 "$report_checkout/trace-overlay.py" "$PWD" "$overlay_root/entry"
gofmt -w "$overlay_root"/entry/*.go
go build -buildvcs=false -overlay="$overlay_root/entry/overlay.json" \
  -o "$overlay_root/entry-census" ./stage3/census/latent/tool

ADAMIC_RGRRT32_TRACE=1 \
ADAMIC_RGRRT32_UNITS=getOutputPathsFor,isOuterExpression,skipOuterExpressions,isCommaExpression,setValueDeclaration \
LATENT_ASSERT_NO_OUTPUT=1 LATENT_FULL=0 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
  "$overlay_root/entry-census" "$adapted/src/tsc/tsc.ts" "$overlay_root/latent.jsonl" \
  2>"$overlay_root/trace.stderr"
```

## Raw diagnostics

The stack excerpts start at the return site and retain the overlay's reported source lines.

### `setValueDeclaration` (main)

```text
RGRRT32 store unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 ir=ir.SetProperty{Nullish:0, DeclaredType:"Declaration | undefined", Record:false, Uninitialized:false, Unset:false, Object:ir.Read{Local:546, Of:4, Checked:false, Readiness:"symbol", Unset:false}, Name:"valueDeclaration", Value:ir.Read{Local:547, Of:4, Checked:false, Readiness:"node", Unset:false}, Class:0, Site:1, Define:false, ViewReceiverTypeID:9351} checked=false
RGRRT32 completed unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 name=setValueDeclaration site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 stores=1 receiverTypeID=9351 checked=false viewOrigins=0
RGRRT32 reach unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 writeSite=1 receiver=ir.Read{Local:546, Of:4, Checked:false, Readiness:"symbol", Unset:false} receiverSites=[] receiverUnknown=true reasons=["local or result has no tracked producer" "no reaching allocation established"] viewedReceiverSites=[] viewOrigins=0 viewOriginsUnknown=false checked=false
```

### `setValueDeclaration` (head)

```text
RGRRT32 store unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 ir=ir.SetProperty{Nullish:0, DeclaredType:"Declaration | undefined", Record:false, Uninitialized:false, Unset:false, Object:ir.Read{Local:546, Of:4, Checked:false, Readiness:"symbol", Unset:false}, Name:"valueDeclaration", Value:ir.Read{Local:547, Of:4, Checked:false, Readiness:"node", Unset:false}, Class:0, Site:1, Define:false, ViewReceiverTypeID:9351} checked=false
RGRRT32 completed unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 name=setValueDeclaration site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 stores=1 receiverTypeID=9351 checked=false viewOrigins=0
RGRRT32 reach unit=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4306:1 writeSite=1 receiver=ir.Read{Local:546, Of:4, Checked:false, Readiness:"symbol", Unset:false} receiverSites=[] receiverUnknown=true reasons=["local or result has no tracked producer" "no reaching allocation established"] viewedReceiverSites=[] viewOrigins=0 viewOriginsUnknown=false checked=false
```

### `isOuterExpression` (main)

```text
RGRRT32 judgment unit=/workspace/scratch/rgrtt32/adapted/src/compiler/factory/utilities.ts:632:1 site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 receiverTypeID=9351 known=false judgedChecked=false checked=false viewOrigins=0 error=<nil>
RGRRT32 completed unit=/workspace/scratch/rgrtt32/adapted/src/compiler/factory/utilities.ts:632:1 name=isOuterExpression site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 stores=0 receiverTypeID=9351 checked=false viewOrigins=1
```

### `isOuterExpression` (head)

```text
RGRRT32 refusal unit=/workspace/scratch/rgrtt32/adapted/src/compiler/factory/utilities.ts:632:1 site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 receiverTypeID=9351 checked=true ir=unavailable reach=unavailable viewOrigins=0 error=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet
github.com/system-inc/adamic/internal/lower.(*lowering).recordSourceSlotWrite(0x346cfd4ba008, 0x346cddbed4a0, 0x346ce09c1268, {0x1e74610, 0x346d1011c4e0})
	/workspace/scratch/rgrtt32/head/internal/lower/source_slot_certificate.go:114 +0x17e
github.com/system-inc/adamic/internal/lower.(*lowering).judgeViewAccess(0x346cfd4ba008, 0x346cddbed4a0, 0x0)
	/workspace/scratch/rgrtt32/head/internal/lower/interface_cast.go:199 +0x392
github.com/system-inc/adamic/internal/lower.(*lowering).viewWalkIndexed(0x346cfd4ba008, {0x346d09623088?, 0x346cfd4a5088?, 0x346cfd4ba008?}, 0x346d0fed8120)
	/workspace/scratch/rgrtt32/head/internal/lower/view_sites.go:165 +0x3a5
github.com/system-inc/adamic/internal/lower.(*lowering).viewWalk(0x346cfd4ba008?, 0x346d0fed8120?, {0x346d09623088?, 0x346ce14b0300?, 0x1e49578?}, 0x346ce14b0300?)
	/workspace/scratch/rgrtt32/head/internal/lower/view_sites.go:113 +0x45
github.com/system-inc/adamic/internal/lower.(*lowering).view(0x346cfd4ba008, 0x346cdd32d798, {0x0, 0x0}, 0x346ce455e780, 0x346ce14b0300)
	/workspace/scratch/rgrtt32/head/internal/lower/interface_cast.go:110 +0x2ac
github.com/system-inc/adamic/internal/lower.(*lowering).predicateMembership(0x346cfd4ba008, 0x346cdd32d798, {0x1e74530, 0x346d056dc4b0}, 0x346d061fdc60?, 0x0?, 0x346d061fe5b0, 0x0)
	/workspace/scratch/rgrtt32/head/internal/lower/predicates_proof.go:953 +0x1129
github.com/system-inc/adamic/internal/lower.(*lowering).checkedPredicateCheck(0x346cfd4ba008, 0x346cdd379808, 0x346cdd32d798, {0x1e74530, 0x346d056dc4b0}, 0x346ce455e780, 0x346ce14b0300, {0x1e74530, 0x346d056dc4e0}, 0x0, ...)
	/workspace/scratch/rgrtt32/head/internal/lower/predicates_checked.go:787 +0x388
RGRRT32 judgment unit=/workspace/scratch/rgrtt32/adapted/src/compiler/factory/utilities.ts:632:1 site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 receiverTypeID=9351 known=false judgedChecked=true checked=true viewOrigins=0 error=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet
RGRRT32 completed unit=/workspace/scratch/rgrtt32/adapted/src/compiler/factory/utilities.ts:632:1 name=isOuterExpression site=/workspace/scratch/rgrtt32/adapted/src/compiler/utilities.ts:4315:9 stores=0 receiverTypeID=9351 checked=true viewOrigins=0
```


The emitter trace is still running at this checkpoint.
