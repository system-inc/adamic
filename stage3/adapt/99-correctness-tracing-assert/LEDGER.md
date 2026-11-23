- baobab, `99-correctness-tracing-assert`: stopTracing passes Debug.assert a
  boolean (`!!tracing`) where upstream passed the namespace object or undefined.
  The asserts-expression narrowing of tracing is lost; nothing after the call
  reads tracing, and its next occurrence assigns undefined. The assertion message
  and all other stopTracing bytes are preserved. Both 79-root type checks and
  the complete sharded lane pass; the lane records 0 assertion calls and 0 outcome
  differences, with positive Node controls. See the adaptation PROOF.md.
