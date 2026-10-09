# `no-underscore-dangle`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **100** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow dangling underscores in identifiers

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

100 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:405`**

```
return this._value;
```

> Unexpected dangling '_' in '_value'

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:408`**

```
this._value = value;
```

> Unexpected dangling '_' in '_value'

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:415`**

```
expect(result._value).toBe('internal');
```

> Unexpected dangling '_' in '_value'

**`libraries/structure/libraries/nexus/source/protocols/rpc/RpcResponse.ts:85`**

```
response.__version === '1.0' &&
```

> Unexpected dangling '_' in '__version'

**`libraries/structure/libraries/nexus/source/protocols/rpc/RpcResponse.ts:87`**

```
response.__protocol === 'BaseRPC'
```

> Unexpected dangling '_' in '__protocol'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

