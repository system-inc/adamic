The method-values topic is held because automatic approval review rejected the
proposed native ABI conflict resolution. It judged the multi-file closure
runtime rewrite unverified and too risky. No part of that rejected edit ran.
The topic cherry-pick was aborted in its isolated review worktree.

The stack has both uncounted and counted closure code pointers, selected by a
flag, plus `adamic_method_entry` tables, canonical closure identity and owning
frame cells. The incoming topic assumes one three-argument code-pointer type
and raw method-pointer tables. Taking either side alone loses behavior.

The proposed resolution is:

* Store extracted class-method wrappers as counted closures, using the existing
  counted dispatcher and `adamic_method_entry`. Leave each original method's
  counted/uncounted selection unchanged.
* Keep the existing packed fixed and rest argument slots. Treat the observable
  argument count separately from packed storage length when prepending a
  receiver. Keep the topic's unsupported-rest and unsupported-slot refusals.
* Keep bound receivers and original callables owned. Preserve the topic's weak
  unbound-wrapper cache and clear it on destruction, alongside the stack's
  canonical frame-cache unlinking.
* Preserve checked-view and field-readiness paths before extracting a callable.
  Keep every current extraction, type, capture and receiver guard unless the
  topic's proof admits that exact source form under the existing ABI.
* In JavaScript, preserve counted direct-call dispatch and static-method tables.
  Extraction and binding must use that dispatch and retain callable identity.

The affected conflict files are `internal/javascript/javascript.go`,
`internal/native/emit_expressions.go`, `internal/native/emit_objects.go`,
`internal/native/runtime/adamic.h`, `internal/native/runtime/closure.c`, and
`internal/native/runtime/heap.c`. A program-level extraction flag would select
the existing counted convention for wrappers, without editing emit.go.

Before delivery, the topic's own positive fixtures must agree with Node in both
backends, its mutants must fail meaningfully, and the compiler package, counts
and catalog checks must pass. A positive fixture that needs weakening a retained
guard makes the whole topic a skip, with its exact location recorded.
