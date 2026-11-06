# 99r measurement amendments

## 2026-11-06: toSorted shared empty arm, task b5w3ycg

Measured before editing any existing 99r owner, adapter, contract or README.
Branch base: origin/main ce4480011426adbd0213af313cc87ddf0fbf55fd.
Requested comparison base: 37f8314b. Their complete stage3/adapt inputs are
byte-identical. The full 37f8314b apply with --build-product exited 0.
Upstream is 050880ce59e30b356b686bd3144efe24f875ebc8; the checker is stock
TypeScript 6.0.3. Setup, both API installs and both unmodified compiler builds
also exited 0. No compiler source or adaptation owner was changed for these
measurements. Candidate source exists only in isolated scratch copies or in
an in-memory CompilerHost.

### Stock checker at core.ts:1049:34

| Expression | Type | Contextual type |
| --- | --- | --- |

| conditional | `T[]` | `unknown` |

| emptyArray arm | `never[]` | `unknown` |

| slice().sort(comparer) arm | `T[]` | `unknown` |


The emptyArray identifier itself is never[], contextual unknown. Stock TypeScript
reduces the conditional to T[]. The mutable T[] is an implied target in Adamic,
not the stock checker's contextual type at the read. These are distinct facts.

### Refusal decision

The reason is constructed by wideningRefusal in internal/lower/invariance.go:737.
Its caller refuseWidening (449-587) asks impliedTarget (834-850) for the reduced
conditional type when there is no relevant slot, and freshOrWidened/widened walks
the writable array element; wideningRefusal emits the type-parameter reason.
It would accept that shared arm as readonly T[], whose element cannot be written,
rather than exposing never[] as mutable T[].

### Candidate spellings and emitted bytes

Each candidate was substituted into the complete adapted core.ts in memory.
All reported equality checks compare the complete emitted JavaScript, not a
function fragment. The pre-existing unknown/SortedReadonlyArray mint is retained.

| Candidate | Conditional type | Arm type | Pure upcast | JS comments kept | JS comments removed | Qualification |
| --- | --- | --- | --- | --- | --- | --- |

| `emptyArray as readonly never[]` | `readonly never[] | T[]` | `readonly never[]` | yes | identical | identical | Byte-identical readonly upcast |

| `emptyArray as readonly T[]` | `readonly T[]` | `readonly T[]` | yes | identical | identical | Byte-identical readonly upcast |

| `emptyArray as ReadonlyArray<T>` | `readonly T[]` | `readonly T[]` | yes | identical | identical | Byte-identical readonly upcast |

| `<readonly T[]>emptyArray` | `readonly T[]` | `readonly T[]` | yes | identical | identical | Byte-identical readonly upcast |

| `emptyArray as readonly []` | `readonly [] | T[]` | `readonly []` | no | identical | identical | Excluded: not an upcast; stock checker also reports TS2352 |

| `emptyArray satisfies readonly T[]` | `T[]` | `never[]` | no new cast | identical | identical | satisfies retains mutable never[]; does not produce a readonly expression view |

| `(emptyArray as readonly never[]) satisfies readonly T[]` | `readonly never[] | T[]` | `readonly never[]` | yes | identical | identical | Byte-identical readonly upcast |

| `conditional satisfies readonly T[]` | `T[]` | `never[]` | no new cast | identical | identical | Readonly context only; ordinary compiler still stops at 1049:34 |


`conditional satisfies readonly T[]` means inserting `satisfies readonly T[]`
between the conditional's existing closing parenthesis and `as unknown`.
Stock TypeScript supplies readonly T[] as the empty read's contextual type,
but the arm remains never[] and the conditional remains T[]. The unmodified
37f8314b compiler still refuses the same read. A new annotated local would change
JavaScript and does not qualify. The readonly empty tuple is not a pure upcast:
a never[] does not prove length zero. No new non-upcast assertion qualifies.

The ordinary binder probe of `emptyArray as readonly T[]` advances to
core.ts:1076:100, Refused: a cast the runtime can't check, in both ordinary rows.
The whole-conditional satisfies probe retains core.ts:1049:34 and the exact
original reason in both rows. These are isolated candidate measurements, not an
amendment, a full census, or a native compilation claim.

The measured useful spelling is:

```ts
return (array.length === 0 ? emptyArray as readonly T[] : array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>;
```

This adds one pure upcast. No measured cast-free spelling both produces the
readonly arm view and clears the refusal. Step 2's explicit instruction to stop
if no spelling exists without a cast therefore prevents an amendment unless the
user clarifies that the earlier pure-upcast exception also satisfies that condition.
At this measurement checkpoint all existing 99r files, including owners.json,
remain unchanged. No Step 2 or checks 1-5 after an amendment are claimed.

### Rulings

No ruling names emptyArray or specifically a shared empty-array view. The general
rule at docs/rulings.md:91 is:


> - **A mutable location (array element, mutable property, map value, class type argument) is invariant: `Dog[]` as `Animal[]`, `{ pet: Dog }` as `{ pet: Animal }` and `Map<string, Dog>` as `Map<string, Animal>` are refused, through a type parameter's constraint or an intersection too, and so is a `readonly` field seen again as writable; a `readonly` location is covariant.** Refuse. Oct 5 2026, Kirk and @system_adamic; source: #8vpywxp, #qffy2j8. Spec: The soundness rules (row `adamic/invariant-mutable`, and "Taken together"); Five programs 0.1 refuses, 5. Test: internal/lower/lower_test.go:170 TestAMutableLocationSeenWiderIsRefused; cohere/internal/lint/rules/adamic/invariant_mutable_test.go:36.


### Other emptyArray sites in the latent census on main

The tracked main latent census is historical main
b8fb957aa839a9e8cb0b54279dd9864fa317bd30, not 37f8314b. Its raw source is
stage3/census/latent/data/unmerged4/main-unmerged.jsonl.gz. The same bytes are
present on 37f8314b. It is explicitly measured on a checker-rejected program,
with refusal-scan and independent-unit findings, and is not an ordinary 79-root
first-stop census. No fresh latent run on 37f8314b is claimed.

Deduplicating by (where, reason), the exact `never seen as T ... constraint any`
reason occurs once, at historical core.ts:1039:34, toSorted, pinned by 99r.
Thus there are **zero other sites with the identical reason**. The analogous
parameter wording occurs once more at core.ts:414:22, flatMap, using U and
constraint {}; flatMap is pinned by 99m. Its readonly return could use the same
readonly-upcast shape if needed; this historical finding does not prove a
current 37f8314b stop.

For completeness, the broader invariant-mutable family has **25 other shared
emptyArray reads** below, excluding fresh [] literals, whole-container findings
that do not point at an emptyArray read, and toSorted itself. Coordinates are
mapped to the completed 37f8314b adapted source; the archived raw findings retain
the historical coordinates. These are historical findings mapped to current
source, not a claim that all 25 still refuse on 37f8314b. A readonly view is only
sufficient where the receiving contract is readonly; mutable result/storage
contracts need their own investigation. None was edited.

| Adapted compiler file:line:column | Function or initializer | Existing whole-owner pin |
| --- | --- | --- |

| `builderState.ts:502:63` | `getAllFileNames` | none |

| `builderState.ts:562:63` | `getAllFilesExcludingDefaultLibraryFile` | none |

| `core.ts:414:22` | `flatMap` | 99-correctness-axolotl-m |

| `emitter.ts:606:27` | `createAddOutput.getOutputs` | none |

| `moduleSpecifiers.ts:401:143` | `getModuleSpecifiersWithCacheInfo` | none |

| `moduleSpecifiers.ts:555:69` | `computeModuleSpecifiers` | none |

| `parser.ts:10594:67` | `processCommentPragmas` | none |

| `program.ts:649:57` | `getPreEmitDiagnostics` | none |

| `programDiagnostics.ts:168:52` | `createProgramDiagnostics.getCombinedDiagnostics.<anonymous>` | none |

| `programDiagnostics.ts:240:93` | `createProgramDiagnostics.createDiagnosticExplainingFile` | none |

| `programDiagnostics.ts:269:28` | `createProgramDiagnostics.createDiagnosticExplainingFile` | none |

| `semver.ts:211:99` | `VersionRange.constructor` | none |

| `sys.ts:846:22` | `createDirectoryWatcherSupportingRecursive.updateChildWatches` | none |

| `sys.ts:864:57` | `createDirectoryWatcherSupportingRecursive.updateChildWatches` | none |

| `transformer.ts:128:26` | `getScriptTransformers` | none |

| `tsbuildPublic.ts:619:37` | `createBuildOrder` | none |

| `tsbuildPublic.ts:620:23` | `createBuildOrder` | none |

| `utilities.ts:640:42` | `getDeclarationsOfKind` | none |

| `utilities.ts:1714:21` | `getScriptTargetFeatures.memoized callback` | none |

| `utilities.ts:1913:21` | `getScriptTargetFeatures.memoized callback` | none |

| `utilities.ts:1927:21` | `getScriptTargetFeatures.memoized callback` | none |

| `utilities.ts:4678:22` | `getJSDocCommentsAndTags` | none |

| `utilities.ts:10097:13` | `getModuleSpecifierEndingPreference.inferPreference` | none |

| `utilities.ts:10141:24` | `getRequiresAtTopOfFile` | none |

| `watchUtilities.ts:414:84` | `updateSharedExtendedConfigFileWatcher` | none |


### Evidence and reproduction

measure-tosorted.cjs records checker types, contexts, diagnostics, assignability,
and complete core JavaScript hashes in both comment modes. toSorted is never
rewritten on the baseline tree. Commands were logged to files, without piping
runs into head or tail. The archive retains the exact scripts, logs, raw stock
checker results, candidate ordinary rows, historical census, source inventory and
SHA256SUMS. The archive itself has an adjacent SHA-256 list.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# Run from the detached 37f8314b checkout, with its pinned API installed:
stage3/apply.sh --build-product /workspace/scratch/tosorted/before
# Run from this checkout:
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/measure-tosorted.cjs /workspace/scratch/tosorted/before
```

## 2026-11-06 20:05Z: authorized upcast amendment

The user relayed Adamic's ruling on b5w3ycg: pure upcasts qualify; fold
`emptyArray as readonly T[]` into 99r. This resolves the stop in the earlier
measurement section. The earlier closeout JSON remains historical Step 1 evidence.

The applicable rule is docs/rulings.md:14: an as upcast is accepted only under
Adamic's own sound relation, immutable covariance included, and is then erased.
The revised contract runs the actual extracted cast and a relation-only satisfies
witness through the built compiler, not TypeScript assignability as an oracle.
Both probe functions preserve the shared never[] declaration and unconstrained T,
and are instantiated for number and string. The readonly-tuple and branded
downcast mutants turn this relation contract red without owner hash checks.
The mutable T[] control passes the compiler's checker and fails Lower, proving
that the check reaches Adamic's writable-slot rule beyond assignability.

Main changed under stage3/adapt/95-checker-axolotl-h after the original base.
The branch was rebased onto 73cdea837469a7a9e27ff692e747421646a068dd before the
amendment. The previously pushed d7f59a541 commit is also retained as a merge
parent, allowing every subsequent push to remain a plain fast-forward push.
The fresh before apply uses that exact main revision; the after apply uses its
adaptations plus this amendment. The independent census compares 37f8314b's
ordinary compiler/tree with the newly based ordinary compiler and amended tree.

Only toSorted's after text and its after hash change in owners.json. Its before
pin and every other owner remain unchanged. The full source/output comparison
and census results are recorded by the continuation's proof and census files.
