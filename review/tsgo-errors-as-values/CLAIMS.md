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

This error shape is explicit in the task brief. The successful alternatives
and any design-specific naming are pending access to checker-handle-design.md
section 5. That attachment is absent from this checkout and available task
connectors; the owner has been asked to provide it. This file claims the error
contract early and does not claim that implementation or validation is done.

## Observed starting state and remaining work

The C ABI already returns status/error buffers for all five operations. Go
checker panics are recovered into these buffers. Native `checked()` discards
that recoverability by panicking; argument/path validation also panics.
JavaScript bridge calls are currently refused at compile time. A supported,
honest JavaScript execution path is needed for the requested two-backend
failure fixture; this claim does not invent a JavaScript checker.

WASI cannot link the Go checker. Preserve a named target refusal and keep the
ordinary runtime archive buildable without the bridge.

No implementation, mutants, counts, or gate results are claimed yet.
