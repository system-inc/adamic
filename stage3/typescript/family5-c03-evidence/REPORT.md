# Family 5 C03 stop evidence (#pncjpqy)

C03 is rejected on main `150b7b5e8ec45cd88e294dbb728322ec41247737`.
Adding `/** @internal */ constraint?: Type;` clears the 14 specific G03
and G12 refusals. The next field check at every one of those sites names
`default`. Outside the target sites, the continuing refusal scan shows
26 changed findings and 36 additional findings. 52 of the 62 replacement
or additional events belong to units with no G03 or G12 stop.
The proposed adapter is preserved here as evidence. No landing commit or
ledger append was made.

At `checker.ts:17226:44`, in `getNoInferType`, the new mutable variance
refusal reads “a value of type SubstitutionType seen as Type, which can
write Type | undefined where Type is read.” Its other conditional arm,
at `17226:16`, reports “a value of type Type seen as Type, which can write
Type | undefined where Type is read.” Neither site had a refusal on main.
The new optional writable slot exposes a wider write path to
`SubstitutionType.constraint`, which is required. These examples alone
fail the gate. At `checker.ts:5576:16` in `createTypeParameter`, an
existing cast refusal changes the unproven optional field from
`constraint` to `default`. That unit also has no G03 or G12 site.

## Pins and inputs

- Main `150b7b5e8ec45cd88e294dbb728322ec41247737`
- Scout `464502ff430b1c88b2e79ebf6b21badcb09e02af`
- TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`
- Cohere `5360753e19e4417675a90e92a1823bea66b13918`
- Checker submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`

The main pin search covered every adapter, owners file and targets file.
Adaptation 75 addresses three named members on `Type` and validates their
compatibility with `InstantiableType`. There is no whole-owner or whole-file
pin covering `Type.constraint`. The proposal pins the composed interface
before and after. Main and candidate `checker.ts` are byte-identical.

## Proofs

```text
PASS main Stage 3 apply (exit 0)
PASS adapter-apply (exit 0)
PASS census-scan-before (exit 0)
PASS census-scan-after (exit 0)
FAIL census-scan-grade (exit 1)
```

The serial preflight loaded the 79 compiler roots and selected
`checker.ts` with `LATENT_ONLY_FILE`. Both headers report zero loader
diagnostics. A scratch build of the ordinary full-mode overlay emitted
the complete continuing refusal scan before unit registration and lowering.
The only preflight instrumentation is in [scan-only.patch.gz](scan-only.patch.gz).
Both children emitted the selected file and the completion trailer.
The preflight graded every recorded finding by kind, phase, location,
reason, text and unit, with paths normalized and multiplicity retained.
Total findings went from 1,502 to 1,538. Full unit lowering, stock checks,
mutant census, product, whole lane and public declaration comparison
were skipped after the gate failed.

The TypeScript input was bootstrapped as in the meter by a depth-one
clone of the upstream repository and tag in main's `stage3/source.json`.
The main apply used `ADAMIC_TYPESCRIPT_SOURCE` set to that checkout.
The scratch candidate was a copy of the resulting `src` and
`package.json` with only the proposed adapter applied. The scanner was
built with the checked-in `make_overlay.py` and the included preflight
patch. Both invocations used `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`
and `CENSUS_MODULE_FORMAT=CommonJSBundle`. The raw streams, all deltas,
and exact proof log are retained alongside this report.

## Every reached stop

- `checker.ts:6968:54: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:7131:128: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:7142:90: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:9276:95: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:13244:71: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:13276:58: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:13612:16: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:16607:33: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:19948:30: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:23690:67: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:23696:71: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:25216:31: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source InterfaceType, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:31813:31: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source InterfaceType, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`
- `checker.ts:35806:48: Adamic 0.1 refuses optional property default in TypeParameter absent from structural source Type, which can hide fields; declare default on the source type, or build a fresh object with known fields (adamic/no-optional-widening)`

## Findings outside the target sites

| Site | Unit start | Previous reason | Candidate reason |
| --- | --- | --- | --- |
| checker.ts:1885:121 | 1490:1 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:2127:133 | 1490:1 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:5576:16 | 5575:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:6950:69 | 6777:9 (contains a target) | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:6956:87 | 6777:9 (contains a target) | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:6962:97 | 6777:9 (contains a target) | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:7043:56 | 6777:9 (contains a target) | No finding | a cast the runtime can't check |
| checker.ts:7173:93 | 7119:13 (contains a target) | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:14969:132 | 14955:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15159:84 | 15158:5 | an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15176:59 | 15174:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15180:73 | 15174:5 | No finding | a cast the runtime can't check |
| checker.ts:15359:68 | 15339:9 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15381:67 | 15379:9 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15382:25 | 15379:9 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:15443:70 | 15379:9 | No finding | a cast the runtime can't check |
| checker.ts:17226:16 | 17225:5 | No finding | a value of type Type seen as Type, which can write Type \| undefined where Type is read |
| checker.ts:17226:44 | 17225:5 | No finding | a value of type SubstitutionType seen as Type, which can write Type \| undefined where Type is read |
| checker.ts:17234:97 | 17229:5 | No finding | a cast the runtime can't check |
| checker.ts:17241:59 | 17239:5 | No finding | a cast the runtime can't check |
| checker.ts:17245:16 | 17244:5 | No finding | a value of type Type seen as Type, which can write Type \| undefined where Type is read |
| checker.ts:17247:13 | 17244:5 | No finding | a value of type SubstitutionType seen as Type, which can write Type \| undefined where Type is read |
| checker.ts:17256:24 | 17250:5 | No finding | a cast the runtime can't check |
| checker.ts:17264:30 | 17263:5 | No finding | a value of type SubstitutionType seen as Type, which can write Type \| undefined where Type is read |
| checker.ts:18996:67 | 18994:5 | No finding | a cast the runtime can't check |
| checker.ts:19497:20 | 19488:5 | No finding | a cast the runtime can't check |
| checker.ts:19498:18 | 19488:5 | No finding | a cast the runtime can't check |
| checker.ts:19499:44 | 19488:5 | No finding | a cast the runtime can't check |
| checker.ts:19499:105 | 19488:5 | No finding | a cast the runtime can't check |
| checker.ts:19501:21 | 19488:5 | No finding | a cast the runtime can't check |
| checker.ts:19760:43 | 19758:5 | No finding | a cast the runtime can't check |
| checker.ts:19957:47 | 19955:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:19958:47 | 19955:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:20883:24 | 20878:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:21129:50 | 21072:5 | No finding | a cast the runtime can't check |
| checker.ts:21133:52 | 21072:5 | No finding | a cast the runtime can't check |
| checker.ts:22340:66 | 22334:5 | No finding | a cast the runtime can't check |
| checker.ts:22340:131 | 22334:5 | No finding | a cast the runtime can't check |
| checker.ts:22947:116 | 22890:9 | an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeVariable: optional field default has no proven compatible presence/type |
| checker.ts:22948:59 | 22890:9 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:23599:47 | 23560:9 (contains a target) | No finding | a cast the runtime can't check |
| checker.ts:23599:86 | 23560:9 (contains a target) | No finding | a cast the runtime can't check |
| checker.ts:23600:52 | 23560:9 (contains a target) | No finding | a cast the runtime can't check |
| checker.ts:23600:93 | 23560:9 (contains a target) | No finding | a cast the runtime can't check |
| checker.ts:23914:60 | 23560:9 (contains a target) | an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeVariable: optional field default has no proven compatible presence/type |
| checker.ts:25005:32 | 24998:5 | No finding | a cast the runtime can't check |
| checker.ts:25129:86 | 25128:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:26531:31 | 26530:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:26931:182 | 26800:9 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:26982:33 | 26800:9 | No finding | a cast the runtime can't check |
| checker.ts:26983:63 | 26800:9 | No finding | a cast the runtime can't check |
| checker.ts:30499:21 | 30497:5 | No finding | a cast the runtime can't check |
| checker.ts:34420:37 | 34404:5 | No finding | a cast the runtime can't check |
| checker.ts:34435:81 | 34432:5 | No finding | a cast the runtime can't check |
| checker.ts:34611:31 | 34507:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:34611:106 | 34507:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:34611:166 | 34507:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:34630:57 | 34622:5 | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | an unproven relation from Type to TypeParameter: optional field default has no proven compatible presence/type |
| checker.ts:54062:35 | 53883:5 | No finding | a cast the runtime can't check |
| checker.ts:54062:70 | 53883:5 | No finding | a cast the runtime can't check |
| checker.ts:54066:31 | 53883:5 | No finding | a cast the runtime can't check |
| checker.ts:54066:68 | 53883:5 | No finding | a cast the runtime can't check |
