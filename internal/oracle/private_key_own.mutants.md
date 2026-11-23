# Mutant for #a0xw8mg's private-key stop

**P1, the rule dropped.** `privateKeyHazard` (internal/lower/private_key.go) answers "" before reading the key,
so no own-key or presence test stops for a key that may spell a private slot's native name.

Must turn red:
- `TestPrivateKeyOwnStops/literal.a`, `computed.a`, `detached.a` and `in_literal.a`: each lowers where it must stop.
- `TestReviewProgramsRefuse/fxspptb_903f25b_private_mangled.a`: the agreement daily's program is accepted again.

Must stay green: `TestPrivateKeyOwnStops/hasown.a` (Object.hasOwn's own # rule) and `TestPrivateKeyOwnNeighborsAgree`.
