# Family 5 pragma scout · #pncjpqy

Checkpoint. The scout started from the observed origin/main tip `384c6daa3d230929b7b90a97ae44f9781282153a`. The recursive submodule pins are in [submodules.txt](evidence/submodules.txt). The TypeScript input was v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, as pinned by `stage3/source.json` and fetched by `stage3/meter/twice-daily.sh`.

The prior scout at `464502ff430b1c88b2e79ebf6b21badcb09e02af` identified G43 in `addPragmaForMatch` and G44 in `extractPragmas`. C13 and C14b declared only the first absent optional member, `types`. The trial moved the refusal to `lib` and exposed a string widening into the union of pragma records. The structural-widening rule is at `docs/rulings.md:71`.

## Arguments by pragma

Line references in this table are to upstream v6.0.3. A span is `{ value: string; pos: number; end: number }`.

| Pragma | Constructed own keys and values | Spec |
| --- | --- | --- |
| `reference` | Optional `types`, `lib`, `path` spans. Optional `no-default-lib`, `resolution-mode`, `preserve` strings, with explicit `undefined` possible for empty single quotes. | types.ts:10241–10250 |
| `amd-dependency` | Required own `path`, optional own `name`. Each present value is a string or `undefined`. | types.ts:10252–10255 |
| `amd-module` | Required own `name`, a string or `undefined`. | types.ts:10256–10259 |
| `ts-check` | No keys, even with trailing arguments. | types.ts:10260–10262 |
| `ts-nocheck` | No keys, even with trailing arguments. | types.ts:10263–10265 |
| `jsx` | `factory` string when written. It can be absent. | types.ts:10266–10269 |
| `jsxfrag` | `factory` string when written. It can be absent. | types.ts:10270–10273 |
| `jsximportsource` | `factory` string when written. It can be absent. | types.ts:10274–10277 |
| `jsxruntime` | `factory` string when written. It can be absent. | types.ts:10278–10281 |

For the current table, the name lookup and flag check select exactly the applicable spec. In the positional branch, only that spec's `argument.name` is written. The early return on absent payload occurs before required-argument validation. In the XML branch, `extractPragmas` performs its own flag check and obtains each key and capture mode from the selected spec. A foreign attribute cannot add a key. The XML expression `matchResult[2] || matchResult[3]` yields `undefined` for empty single quotes. For spans, `value.length` throws before the write. For string entries, the parser performs the write with the undefined value. These guarantees concern the current runtime table. `as const` does not freeze it. See `parser.ts:10714–10764,10767–10797`.

## Readers

`processCommentPragmas` populates a map whose repeated names become arrays (`parser.ts:10591–10612`). `processPragmasIntoFields` reads the six reference fields, AMD dependencies and AMD module names. The check pragmas use only the range. The JSX entries are read directly (`parser.ts:10619–10699`). The source-file parser invokes both passes at `parser.ts:1822–1823`, and the services preprocessor invokes them at `services/preProcess.ts:421–422`. The JSX checker reads factory values at `checker.ts:2444–2447,2490–2493,51352–51355` and checks presence only at `checker.ts:33779–33780`. JSX runtime and import-source selection read the factory strings at `utilities.ts:9413–9425`. The classifier uses only table membership (`services/classifier.ts:941–945`).

## Trial

The scratch generator adds internal type-only `PragmaNamesWithKind` and `ParsedPragmaArguments` in `types.ts`, and changes the arguments member of the internal type alias `PragmaPseudoMap`. It preserves the existing entry and map declarations, which pick up the new payload type through that alias. The parser adds type-only imports and local aliases `XmlPragmaName`, `PositionalPragmaName`, `PragmaBuilder`, `XmlPragmaBuilder`, and `PositionalPragmaBuilder`. It changes erased annotations at the name boundaries, construction locals, indexed writes, the two arguments expressions feeding the entry casts, and the return of `getNamedPragmaArguments`. The finite builders cover the eligible keys while mutating. The asserted payloads express the eligible kind records. The existing discriminated entry assertion still bridges the correlation between the runtime spec, the name, and the payload. The trial edits no public declaration directly.

The producer contract must include missing JSX factories and present undefined AMD strings. Those facts expose existing consumer assumptions. `parseIsolatedEntityName` takes a string, and `AmdDependency.path` is a public string property. A type assertion that simply promises a string at either boundary would contradict the witnesses. See [trial generator](evidence/trial.py) and [runtime witnesses](evidence/witness.json).

Measurements are pending at this checkpoint.
