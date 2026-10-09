# `require-atomic-updates`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **31** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow assignments that can lead to race conditions due to usage of `await` or `yield`

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

31 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/CountingSemaphore.test.ts:295`**

```
criticalSectionActive = false;
```

> Possible race condition: `criticalSectionActive` might be reassigned based on an outdated value of `criticalSectionActive`

**`libraries/structure/source/components/maps/Map.tsx:85`**

```
countryFeatures = decodeFeatures(MapCountriesBase64);
```

> Possible race condition: `countryFeatures` might be reassigned based on an outdated value of `countryFeatures`

**`libraries/structure/source/components/maps/Map.tsx:93`**

```
stateFeatures = decodeFeatures(MapStatesBase64);
```

> Possible race condition: `stateFeatures` might be reassigned based on an outdated value of `stateFeatures`

**`libraries/structure/source/modules/account/shared-state/AccountState.ts:443`**

```
accountRequestInProgress = false;
```

> Possible race condition: `accountRequestInProgress` might be reassigned based on an outdated value of `accountRequestInProgress`

**`modules/apple/apns/ApnsApi.ts:42`**

```
cachedJsonWebToken = await signJsonWebToken(
```

> Possible race condition: `cachedJsonWebToken` might be reassigned based on an outdated value of `cachedJsonWebToken`

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

