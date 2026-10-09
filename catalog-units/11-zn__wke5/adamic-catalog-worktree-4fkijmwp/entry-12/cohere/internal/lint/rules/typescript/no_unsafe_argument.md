# `@typescript-eslint/no-unsafe-argument`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **90** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow calling a function with a value with type `any`

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

90 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:425`**

```
const result = dictionaryFrom(input);
```

> Unsafe argument of type `any` assigned to a parameter of type `object`

**`libraries/structure/libraries/nexus/source/encoding/StreamTransforms.ts:13`**

```
chunks.push(value);
```

> Unsafe argument of type `any` assigned to a parameter of type `Uint8Array<ArrayBufferLike>`

**`libraries/structure/libraries/nexus/source/encoding/StreamTransforms.ts:98`**

```
controller.enqueue(JSON.parse(textDecoder.decode(data)));
```

> Unsafe argument of type `any` assigned to a parameter of type `JsonType | undefined`

**`libraries/structure/libraries/nexus/source/events/EventEmitter.test.ts:548`**

```
emitter.on('simple', testObject.handler.bind(testObject));
```

> Unsafe argument of type `any` assigned to a parameter of type `EventHandlerType<TestEventsInterface, "simple">`

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.test.ts:198`**

```
const imported = await importHmacKey(keyBase64);
```

> Unsafe argument of type `any` assigned to a parameter of type `string | BufferSource`

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

