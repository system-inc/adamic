Temporary: comes out when function widening lands

Plan published before implementation. Slice only: treat Debug.fail's optional
stackCrawlMark as opaque metadata (`{}` rather than AnyFunction), including the
same marker parameter on reached assert and assertEqual. Nothing calls this
marker; V8 receives exactly the same function object as before. This answers
the method-signature-style refusal of stackCrawlMark || fail, where AnyFunction
has a never[] parameter list. Keep captureStackTrace and all runtime statements.

Validate all 509,014 Node tokens, failure behavior, and the complete upstream
baseline. Record the next actual compiler diagnostic before claiming it closed.
