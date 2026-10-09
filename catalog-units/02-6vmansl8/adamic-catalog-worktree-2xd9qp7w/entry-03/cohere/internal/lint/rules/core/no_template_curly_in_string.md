# `no-template-curly-in-string`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **2** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow template literal placeholder syntax in regular strings

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

2 in the tree. Showing the first few.

**`libraries/structure/source/api/event-sources/shared-worker/generated/EventSourceSharedWorkerCode.ts:4`**

```
'// EventSourceSharedWorker Build Time: 2026-07-10T20:50:37.890Z\n\n(()=>{var k=Object.defineProperty;var y=(o,n,e)=>n in o?k(o,n,{enumerable:!0,configurable:!0,writable:!0,value:e
```

> Unexpected template string expression

**`libraries/structure/source/api/web-sockets/shared-worker/generated/WebSocketSharedWorkerCode.ts:4`**

```
'// WebSocketSharedWorker Build Time: 2026-07-10T20:50:37.018Z\n\n(()=>{var v=Object.defineProperty;var w=(r,e,t)=>e in r?v(r,e,{enumerable:!0,configurable:!0,writable:!0,value:t})
```

> Unexpected template string expression

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

