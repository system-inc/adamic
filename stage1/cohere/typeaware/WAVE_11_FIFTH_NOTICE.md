# Fifth batch provenance

no_throw_literal.a, no_useless_backreference.a and prefer_arrow_callback.a port
production Go rule decisions from cohere/internal/lint/rules/core at
715ba94f3608a6500086b1076ce5cb7e51b836db. Regex scanning, reference tracking and
constant folding follow the corresponding cohere ecmascript helpers. Messages
preserve production text byte for byte. Cohere is distributed under its existing
MIT license; see the pinned submodule license. Production Go oracle sources are
unchanged. The new bridge query returns raw declaration-file metadata only.
