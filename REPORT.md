# Paired census for #s8gy89r, family 1, cause U1

Both latent pairs are complete. The ordinary 79-root table is pending.

Base is `6cb495cf2adcd359e59b9d8ead8e148a1593231b`, the sole parent of head `f5008927b8e29375220f16096c204d27c8250bcc`. Fetched `main` was `da83d772d470cc7b0660db3627d709c62febe086` and contains the base. Each executable and overlay was built from its corresponding compiler checkout.

Both stage3 applies used TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. All 304 adaptation checkpoint hashes match, and both final snapshots hash to `fb481d75255cb26e120190f74804e8e51cf06d31`. All measurements use the base-applied tree. Latent flags are `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=0`, `LATENT_TAIL=1` and `LATENT_ASSERT_NO_OUTPUT=1`.

The analysis pairs normalized reads and exact reasons with `stage3/census/pair` and grades findings by kind, phase, reason and message with positions stripped from the message.

The base apply wrote the generated sources and final patch table, then reported a local cache-archive error when tar saw a changing Git pack. A recursive file comparison excluding Git metadata found identical outputs. The census used that directly verified base tree.

## 1. Family 1 site counts

Exact reason is “an object refinement using an open numeric enum as a literal tag”.

| Pass | Base distinct | Base raw | Head distinct | Head raw |
| --- | ---: | ---: | ---: | ---: |
| Compiler directory | 557 | 730 | 503 | 671 |
| `tsc.ts` entry | 557 | 730 | 503 | 671 |

- Compiler directory loses 54 distinct and 59 raw family 1 findings net. 50 distinct base reads lose all findings at that read. 4 distinct reason transitions retain another finding at the same read.
- `tsc.ts` entry loses 54 distinct and 59 raw family 1 findings net. 50 distinct base reads lose all findings at that read. 4 distinct reason transitions retain another finding at the same read.

Workers’ historical count is 110 at `dc9d089c`. GitHub could not resolve that abbreviated commit and the per-site inventory was unavailable. The number of sites in that specific cohort that clear or reach a later stop remains unverified. Section 4 lists the measured changes on the requested snapshots.

## 2. Units whose first stop moved

### Both latent passes

Raw finding grades across all reasons are 0 new, 55 gone, 4 changed, 0 reached.

#### Now lower whole in the scoped measurement

None.

#### Reached a later stop

None.

#### Moved to an earlier or different stop

None.

#### Previously clean units that now stop

None.

### Including each unit’s first recorded refusal-scan stop

The earlier family 1 refusal disappears in these units. The head stops below were already visible in the base stream, so these are effective first-stop moves and add zero findings to the reached grade.

#### Now lower whole in the scoped measurement

None.

#### Reached a later stop

- `createMemberAccessForPropertyName` at `src/compiler/factory/utilities.ts:189:1`. **Base** `src/compiler/factory/utilities.ts:196:66` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:1106:17` (NotYet, lowering) stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet.
- `IdentifierNameMap` at `src/compiler/transformers/utilities.ts:391:1`. **Base** `src/compiler/transformers/utilities.ts:428:56` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/transformers/utilities.ts:429:119` (Refused, refusal_scan) Adamic 0.1 refuses a method read as a value (toKey would lose its object, and this with it); call it in an arrow that keeps the object: (name) => IdentifierNameMap.toKey(name) (unbound-method).
- `getTextOfIdentifierOrLiteral` at `src/compiler/utilities.ts:5438:1`. **Base** `src/compiler/utilities.ts:5439:40` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:1106:17` (NotYet, lowering) stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet.
- `getEscapedTextOfIdentifierOrLiteral` at `src/compiler/utilities.ts:5443:1`. **Base** `src/compiler/utilities.ts:5444:33` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:1106:17` (NotYet, lowering) stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet.

#### Moved to an earlier or different stop

None.

#### Previously clean units that now stop

None.

## 3. Regressions

### Both latent passes

- Previously clean measured units now stopped: 0.
- New independent findings: 0.
- Different findings at existing sites: 4. The head diagnoses below are newly recorded under their reasons and have the grade changed.
  - **Base** `src/compiler/binder.ts:916:51` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:916:51` (NotYet, refusal_scan) stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet.
  - **Base** `src/compiler/binder.ts:920:45` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:920:45` (NotYet, refusal_scan) stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet.
  - **Base** `src/compiler/binder.ts:921:52` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:921:52` (NotYet, refusal_scan) stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet.
  - **Base** `src/compiler/binder.ts:921:79` (Refused, refusal_scan) Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag). **Head** `src/compiler/binder.ts:921:79` (NotYet, refusal_scan) stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet.
- Newly reached findings: 0.
- New panics: 0.

## 4. Changes by predicate

The changed reads below come from the measured pair. Calls were identified from the stock TypeScript control-flow graph and checked against the adapted source. At the two checker reads of `declaration` at lines 11961 and 11968, the checker narrows the variable and leaves the accessed fields at their declared types. Two other reads clear under `supportsUnlabeledBreak`, which also compares a finite set of kinds. The historical probe’s per-site list is needed to match the original 110-site population one by one.

Both latent passes have the same changed reads and predicate attributions.

| Predicate | Cleared at read | Another finding at read | Raw family 1 sites lost | Unit first stop moved later |
| --- | ---: | ---: | ---: | ---: |
| isMemberName | 20 | 0 | 20 | 4 |
| isPropertyNameLiteral | 0 | 0 | 0 | 0 |
| isBindingPattern | 8 | 0 | 12 | 0 |
| hasOnlyExpressionInitializer | 18 | 0 | 18 | 0 |
| isClassLike | 1 | 0 | 1 | 0 |
| canHaveLocals | 0 | 4 | 4 | 0 |
| isPropertyName | 0 | 0 | 0 | 0 |
| isAssignmentPattern | 1 | 0 | 2 | 0 |
| supportsUnlabeledBreak | 2 | 0 | 2 | 0 |
| Unresolved | 0 | 0 | 0 | 0 |

#### `isMemberName`

- `src/compiler/emitter.ts:5253:58`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5254:31`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5257:38`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5258:58`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5259:51`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5263:30`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/emitter.ts:5265:24`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/emitter.ts:5252:13`.
- `src/compiler/factory/utilities.ts:196:66`. Cleared at this read. The unit’s next recorded stop is `src/compiler/binder.ts:1106:17` (NotYet, lowering) src/compiler/binder.ts:1106:17: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet. Candidate call locations are `isMemberName` at `src/compiler/factory/utilities.ts:195:13`.
- `src/compiler/factory/utilities.ts:197:65`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/factory/utilities.ts:195:13`.
- `src/compiler/transformers/utilities.ts:428:56`. Cleared at this read. The unit’s next recorded stop is `src/compiler/transformers/utilities.ts:429:119` (Refused, refusal_scan) src/compiler/transformers/utilities.ts:429:119: Adamic 0.1 refuses a method read as a value (toKey would lose its object, and this with it); call it in an arrow that keeps the object: (name) => IdentifierNameMap.toKey(name) (unbound-method). Candidate call locations are `isMemberName` at `src/compiler/transformers/utilities.ts:428:34`.
- `src/compiler/transformers/utilities.ts:428:96`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/transformers/utilities.ts:428:34`.
- `src/compiler/transformers/utilities.ts:428:128`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/transformers/utilities.ts:428:34`.
- `src/compiler/utilities.ts:5439:40`. Cleared at this read. The unit’s next recorded stop is `src/compiler/binder.ts:1106:17` (NotYet, lowering) src/compiler/binder.ts:1106:17: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5439:12`.
- `src/compiler/utilities.ts:5439:68`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5439:12`.
- `src/compiler/utilities.ts:5439:103`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5439:12`.
- `src/compiler/utilities.ts:5439:111`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5439:12`.
- `src/compiler/utilities.ts:5444:33`. Cleared at this read. The unit’s next recorded stop is `src/compiler/binder.ts:1106:17` (NotYet, lowering) src/compiler/binder.ts:1106:17: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5444:12`.
- `src/compiler/utilities.ts:5444:72`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5444:12`.
- `src/compiler/utilities.ts:5444:114`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5444:12`.
- `src/compiler/utilities.ts:5444:147`. Cleared at this read. Candidate call locations are `isMemberName` at `src/compiler/utilities.ts:5444:12`.

#### `isBindingPattern`

- `src/compiler/checker.ts:32006:79`. Cleared at this read. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:32006:28`.
- `src/compiler/checker.ts:32012:57`. Cleared at this read. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:32006:28`.
- `src/compiler/checker.ts:32040:110`. Cleared at this read. Additional raw positions are `src/compiler/checker.ts:32040:122`. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:32040:72`.
- `src/compiler/checker.ts:32041:50`. Cleared at this read. Additional raw positions are `src/compiler/checker.ts:32041:62`. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:32040:72`.
- `src/compiler/checker.ts:42262:48`. Cleared at this read. Additional raw positions are `src/compiler/checker.ts:42262:54`. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:42261:75`.
- `src/compiler/checker.ts:42262:60`. Cleared at this read. Additional raw positions are `src/compiler/checker.ts:42262:66`. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:42261:75`.
- `src/compiler/checker.ts:45125:65`. Cleared at this read. Candidate call locations are `isBindingPattern` at `src/compiler/checker.ts:45124:32`.
- `src/compiler/emitter.ts:5450:31`. Cleared at this read. Candidate call locations are `isBindingPattern` at `src/compiler/emitter.ts:5449:22`.

#### `hasOnlyExpressionInitializer`

- `src/compiler/checker.ts:11946:60`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11947:28`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11947:57`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11948:70`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11948:106`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11948:150`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11953:59`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11953:100`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11957:80`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11960:36`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11961:64`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11962:69`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11963:47`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11963:115`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11968:45`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11969:78`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11970:47`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.
- `src/compiler/checker.ts:11970:115`. Cleared at this read. Candidate call locations are `hasOnlyExpressionInitializer` at `src/compiler/checker.ts:11946:13`.

#### `isClassLike`

- `src/compiler/checker.ts:20379:47`. Cleared at this read. Candidate call locations are `isClassLike` at `src/compiler/checker.ts:20379:24`.

#### `canHaveLocals`

- `src/compiler/binder.ts:916:51`. Remaining stops at this read include `src/compiler/binder.ts:916:51` (NotYet, refusal_scan) src/compiler/binder.ts:916:51: stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet. Candidate call locations are `canHaveLocals` at `src/compiler/binder.ts:916:22`.
- `src/compiler/binder.ts:920:45`. Remaining stops at this read include `src/compiler/binder.ts:920:45` (NotYet, refusal_scan) src/compiler/binder.ts:920:45: stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet. Candidate call locations are `canHaveLocals` at `src/compiler/binder.ts:916:22`.
- `src/compiler/binder.ts:921:52`. Remaining stops at this read include `src/compiler/binder.ts:921:52` (NotYet, refusal_scan) src/compiler/binder.ts:921:52: stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet. Candidate call locations are `canHaveLocals` at `src/compiler/binder.ts:916:22`.
- `src/compiler/binder.ts:921:79`. Remaining stops at this read include `src/compiler/binder.ts:921:79` (NotYet, refusal_scan) src/compiler/binder.ts:921:79: stage 0 can't lower a read of container narrowed to ArrowFunction | CallSignatureDeclaration | ClassStaticBlockDeclaration | ConstructSignatureDeclaration | ... 15 more ... | TypeAliasDeclaration where a call may have rewritten it (container is assigned inside bindSourceFile) yet. Candidate call locations are `canHaveLocals` at `src/compiler/binder.ts:916:22`.

#### `isAssignmentPattern`

- `src/compiler/transformers/es2018.ts:732:81`. Cleared at this read. Additional raw positions are `src/compiler/transformers/es2018.ts:732:86`. Candidate call locations are `isAssignmentPattern` at `src/compiler/transformers/es2018.ts:732:13`.

#### `supportsUnlabeledBreak`

- `src/compiler/transformers/generators.ts:2472:32`. Cleared at this read. Candidate call locations are `supportsUnlabeledBreak` at `src/compiler/transformers/generators.ts:2471:30`.
- `src/compiler/transformers/generators.ts:2480:32`. Cleared at this read. Candidate call locations are `supportsUnlabeledBreak` at `src/compiler/transformers/generators.ts:2479:25`.

## 5. First stops across the 79 roots

The base ordinary per-root run is in progress.

Raw streams, timing streams and paired results are under [evidence](evidence/).
