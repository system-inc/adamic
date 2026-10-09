# Any-call coverage is incomplete

The additional table records one `a call returning any` site at sys.ts:378:63.
Current main's latent census observes zero rows with that exact reason, before
and after 41. The historical call is timer registration in
scheduleNextPoll, storing host.setTimeout's handle on its matching polling queue.
The host handle must be correlated with clearTimeout; an arbitrary concrete
host return type would lose supported host implementations.

Zero call-return owners are adapted. Zero observations are not proof that main
supports every any call: this census stops at each eligible unit's first error
and skips checker-diagnosed bodies. The small ambient-call probe still fails,
reported as `a value of type any` at the receiving declaration. It is retained
as a cross-kind control, not claimed to be a call-return-specific mutant.
No successful call-kind adaptation or dedicated call-return mutant is claimed.

Minimal contract program for @system_adamic_typescript:

```a
function keepHandle<H>(start: () => H, stop: (handle: H) => void): void {
    const handle = start();
    stop(handle);
}
```

This illustrates the intended owner correlation without any. Applying it to
System and WatchHost requires carrying H through all timer storage and public
host contracts. That work is unfinished; a truthfully correlated generic is
possible and has not received its owner/caller proof here.
