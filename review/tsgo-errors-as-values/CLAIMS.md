# typescript-go errors as values

Owner branch: `codex/tsgo-errors-as-values`.
Base: `48c05d091f0a43c31cbe051b1d6578d99eeedf19`.
Task: RuleContext #45rq89s, design review #k4fm1vf, section 5.

## Error contract claimed for dependent work

The shared error type is named `TSGoError`, exported from `adamic`:

```typescript
export type TSGoError = {
    readonly kind: 'Error';
    readonly message: string;
};
```

Every public bridge operation (`tsgoProgram`, `tsgoQuery`, `tsgoInspect`,
`tsgoTypeParts`, `tsgoRelease`) will return a union containing `TSGoError`.
Callers narrow on `kind`. A failure is never an `Ok` with error text, nor a
checker error routed through `panic`. C error-buffer text is preserved using
its explicit byte length, decoded as UTF-8; a buffer without text requires an
Adamic-owned, platform-independent message. Error messages are data, not thrown
Node errors or platform-dependent libc messages.

Section 5 was supplied verbatim by the task owner and confirms this contract.
`TSGoResult<T> = { readonly kind: 'Ok'; readonly value: T } | TSGoError` is
exported beside `TSGoError`. Program creation uses `T = number`; query uses the
existing `{ nodeKind, symbolName, type }` payload; inspection and type-parts use
`T = string`. Release has no payload: `{ readonly kind: 'Ok' } | TSGoError`.
The harness owns `Answer`, refusal handling and the refusal wire format.

Assumption: the brief leaves the success payload's field name open. This unit
uses the standard `value` field consistently, and omits it for release.

## Observed starting state and remaining work

The C ABI already returns status/error buffers for all five operations. Go
checker panics are recovered into these buffers. The starting native `checked()` discarded
that recoverability by panicking; argument/path validation also panicked.
The starting JavaScript backend refused checker calls at compile time. This
unit adds an explicit Node-API adapter to the real archive for the two-backend
proof; it does not invent a JavaScript checker.

WASI cannot link the Go checker. Preserve a named target refusal and keep the
ordinary runtime archive buildable without the bridge.

Validation evidence is recorded separately; this file defines the public
contract claimed for dependent work.
