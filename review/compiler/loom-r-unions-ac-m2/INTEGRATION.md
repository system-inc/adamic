The first full counts regeneration stopped at main's callback_widening/nullable-object-map.a:

```
<source>.a:1:107: stage 0 can't lower sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet
```

Main's slice 2 added adaptArrayCallback, which boxes each supplied argument into the callback's actual representation, including mapping a legacy NULL-for-null pointer to the boxed null token. The carried union guard ran before this adapter and treated the callback as an unadapted shared callable view. No table was written by that failed regeneration.

Keep main's adapter unchanged. The new exception applies only at a direct standard array map callback. Its return type must be identical to the contextual return type; rest callbacks stay outside the exception; every nonprimitive parameter member must be identical on both sides. Primitive alternatives can be adapted by main's existing checked value adapter. Arrays, objects and nested callables cannot silently acquire a different stored protocol. This is a conservative integration choice: other intrinsic callback views retain the carried stop rather than broadening their ABI in this re-cut.

The existing main fixture passes its release, sanitized-native and JavaScript Node comparisons. New scout_nullable_storage_refused/array_callback.a prints null on Node, and the new top-level TestNullableObjectMapKeepsArrayStorageViewNotYet requires the storage-sharing NotYet. Its third parameter would otherwise view a legacy-null array as a tagged null/undefined array, which the adapter does not copy. The negative program lives below *_refused, so flow's top-level fixture census does not try to lower it. It is NotYet, so it correctly has no a-check header. No position-pinned reader was added or changed.

The complete regeneration after the integration adds only the source's sixteen scout rows and changes only taste_stage3_representations.a, to the source's exact values. Main's three self-compare rows at the conflict insertion point remain unchanged. The initial failed output is retained in counts-before-integration.log. The fresh integration mutants and their catchers are in mutant-results.json; historical source mutants are not claimed rerun.

The GCC header configuration proven in the completed chain-followers re-cut is reused only through the command-local scratch PATH. No shared toolchain wrapper or runtime source is changed. Automatic approval review rejected a persistent shared wrapper and explicitly allowed the scratch-local alternative. The earlier recovery details are at review/compiler/loom-r-chain-followers-ac-m2/ENVIRONMENT.md in commit 7a72cfba37691f562b4d8403b0a40e53bedd6677.
