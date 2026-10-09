# Config conversion carries caller types

TypeScript 6.0.3, source pin `050880ce59e30b356b686bd3144efe24f875ebc8`.
Main baseline: `45487a809f89885a3fc651cd590e7dabf31362dc`.
Ruling: October 8, 12:00, internal caller-specific generics; the existing
`isCompilerOptionsValue` predicate stays exactly as written.
Contracts: `origin/codex/step09-public-any`, scout commit `4e6121d3`.
Patterns read: main's adaptation 41 and adaptation 43 on
`origin/codex/stage3-real-any`. No public API sanction is taken from either.

## Reviewed edits

`rules.json` contains twelve edits, all in `src/compiler/commandLineParser.ts`.
The adapter reads the current tree after lower-numbered adapters, rather than
reconstructing pristine source. It accepts exactly one reviewed original or
adapted span, supports CRLF and LF, parses before writing, checks exact erased
JavaScript, and is idempotent. `predicate.json` protects the entire existing
predicate, including its shallow array acceptance.

| Site | Contract |
|---|---|
| S04 parseJsonConfigFileContentWorker | Carry the raw root `TRaw` to parseConfig. |
| parseConfig | Carry `TRaw` to the raw-config owner; do not claim validation. |
| S07 convertCompileOnSaveOptionFromJson | Carry the raw property's `T` through `{ compileOnSave?: T }`; retain diagnostic conversion and boolean result. |
| S10 compiler-options worker | Accept caller `TRaw`; produce existing CompilerOptions defaults and normalized fields. |
| S11 type-acquisition worker | Accept caller `TRaw`; produce existing TypeAcquisition defaults and fields. |
| S12 watch-options worker | Accept caller `TRaw`; retain optional WatchOptions result. |
| S13 converter overload | Raw `TRaw`, undefined defaults, optional WatchOptions result. |
| S14 converter overload | Raw `TRaw`, normalized default family union, including actual falsy-input undefined return. |
| S15 converter implementation | Enumerate raw `TRaw` and pass each indexed value into generic value conversion. |
| convertJsonOption | Internal export accepts `TRaw`; existing predicate, validation, normalization and output contract remain. |
| convertJsonOptionOfListType | Carry the actual `TValues extends readonly unknown[]`, including unions of array types. |
| list-element caller | Explicit `map<TValues[number], CompilerOptionsValue>` passes the exact element union to recursive conversion. |

No cast, runtime statement, public declaration or public alias is added.
The stock checker found that an element generic alone fails on the actual
union-of-arrays caller. The collection generic and indexed element type solve
that mismatch without a cast. Raw values may be invalid, primitive, functions,
accessors or cyclic objects; the generics do not falsely call them JSON or
normalized options. Public any inputs still infer any at their boundary.

The mutable default object is deliberately **not** returned as arbitrary
`TOptions`. For example, `{ strict: true }` becomes `{ strict: false }` when
normalizing raw `{ strict: false }`. Carrying its literal subtype would be a
false promise. The return remains the normalized family union plus undefined.
The checker witness and actual function execution cover this distinction.

## Remaining sites and limits

- S06 `parseOwnConfigOfJson` still accepts any. It reads arbitrary properties,
  writes a boolean into `compileOnSave`, and forwards unchecked `extends` to an
  existing CompilerOptionsValue parameter. An unconstrained root generic yields
  nine actual checker diagnostics, recorded in evidence/types.json. A constraint
  claiming validated fields would hide invalid public inputs; preserving an
  arbitrary original property subtype across mutation would lie. Further work
  needs a truthful raw field/mutation owner, not a cast. This is unfinished, not
  a claim that no internal representation is possible.
- `ParsedTsconfig.raw` and the public raw input/output declarations stay exact.
  S01, S03, S08 and S09 retain their public any boundaries. The forwarding root
  generics do not eliminate the boundary at S06 or the public raw result field.
- S02 stays exact by ruling. Its shallow acceptance does not prove every array
  element is a CompilerOptionsValue. No native predicate proof is claimed.
- S05 `canJsonReportNoInputFiles`, JsonConversionNotifier/onPropertySet's raw
  value contract, `normalizeNonListOptionValue`, and the list converter's any[]
  output are unchanged. The callback mixes raw and normalized values; path
  normalization lacks a declared relation between option kind and string value;
  list output needs an option/element/result relationship. None is hidden by a
  new cast or silently treated as complete.
- Adaptation 43 is not on this main. Its unchanged referenced adapter rejects
  current main's convertConfigFileToObject anchor (47 now uses a push-only
  diagnostic sink), and its timer anchor precedes 41's Node-specific type.
  The full measurement here is main plus 44. Integration of a rebased 43 is
  not claimed; 44's edits do not overlap its JSON return edits.
- No Adamic/native compilation, latent census delta, whole Go gate or new .a
  fixture is claimed. These are real Node compiler and stock-checker tests.
  No fixture was added, so no counts.md refresh is needed.

## Reproduce

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adapt44-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Run the main and adapted lanes serially in new result directories.
NODE_OPTIONS=--max-old-space-size=1400 bash /tmp/adapt44-main/stage3/lane/run.sh /tmp/adapt44-before > /tmp/adapt44-before.log 2>&1
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/lane/run.sh /tmp/adapt44-final > /tmp/adapt44-final.log 2>&1
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
node stage3/adapt/44-config-generics/check-types.cjs /tmp/adapt44-final/adapted-tree /tmp/adapt44-types.json > /tmp/adapt44-types.log 2>&1
node stage3/adapt/44-config-generics/proof.cjs /tmp/adapt44-before /tmp/adapt44-final /tmp/adapt44-proof > /tmp/adapt44-proof.log 2>&1
```

The lane itself runs stage3/apply.sh and the complete stage3/oracle, with eight
workers and no test filter. Evidence/report.md records observed results.
The checked-in patch table changes only 44's incremental row. The generated
full table in evidence carries the measured total for this tree.
