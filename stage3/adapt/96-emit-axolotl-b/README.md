# 96b: transformer.ts's sparse feature flags

Axolotl (#axolotl), task #vy7ncrg. **This is a JavaScript-changing adaptation** under #wks9251, as 91 is. Its value is the same on Node, and it changes no public API. It edits two statements in src/compiler/transformer.ts.

## The cause

`transformNodes` keeps one feature word per SyntaxKind in `enabledSyntaxKindFeatures = new Array<SyntaxKindFeatureFlags>(SyntaxKind.Count)`. That array is sparse: a slot is a hole until a feature is first enabled for its kind.

`enableSubstitution` and `enableEmitNotification` set a bit with `enabledSyntaxKindFeatures[kind] |= Flag`. That reads the slot, so the checker reports TS2532 at 367 and 395.

- Adaptation 33 declined these two as a deferred assignment shape.
- 33 gave the same slots' reads in `isSubstitutionEnabled` and `isEmitNotificationEnabled` its reviewed U-zero default, `?? 0`.

## The repair

`enabledSyntaxKindFeatures[kind] = (enabledSyntaxKindFeatures[kind] ?? 0) | Flag`, in both functions.

**Why it isn't byte-identical:** the sparse `new Array(Count)`. A hole reads `undefined`, and nothing short of a default makes that read a number.

- **On Node:** `undefined | F` is already `F` (ToInt32 of undefined is 0), so `(undefined ?? 0) | F` gives the same value, and so does a present slot.
- **In Adamic:** the default is what lets the read type-check without claiming the slot is there.

Correctness agreed on #axochk (Nov 1): no honest byte-identical form exists.

## Node witnesses

`witness.cjs <before> <after>` takes the real `enableSubstitution`, `isSubstitutionEnabled`, `enableEmitNotification` and `isEmitNotificationEnabled` from each tree by name. It transpiles them with stock TypeScript 6.0.3 and runs them on Node over a fresh sparse array for each scenario:

- **unset bit:** a kind with no feature yet gets each flag.
- **same bit already set:** each flag is set twice.
- **other bit already set:** one kind gets EmitNotifications, then Substitution, and another the reverse.

For each scenario it compares the array's own keys, its values, and both readers over every kind. Before and after agree in all three. proofs/witness.json is that run (Node v24.14.1).

It also refuses trees that emit the same JavaScript, so it can't pass by comparing a tree with itself. A mutant default, `?? 2` on Substitution, turns it red: "unset bit: after differs from before".

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3 and finds each function by name. It accepts only a body of the `Debug.assert` state guard and then the one write, in its before or after text. It refuses a half-adapted file and plans both edits before its single write.
- **Diagnostics:** `adamic types` on the adapted tree at main a23ba096, alone:
  - whole program, 144 before and 142 after
  - transformer.ts own-file, 6 before and 4 after
  - The two removed are exactly 367 and 395, with no new diagnostic anywhere. The four left are 96a's.
- **Idempotence:** a second run edits nothing.

**Lane:** run as part of Correctness's wave, which carries the manifest regeneration. The gap: no lane run before the wave's own, since no box could run one on Nov 1.
