# V4 direct-call oracle divergences

Argument/result contract violations intentionally stop with exit 70 at invocation. Node running the original misfit source evaluates the argument, then executes the producer and prints its result. The checked native and JavaScript backends preserve argument output, then stop before an invalid producer argument or after an invalid returned result. Diagnostics name view.run, the call site, and producer/view types. Scalar, object, and array argument probes and the result probe pin this difference.

A noncallable member evaluates its argument before Node throws TypeError. Both checked backends preserve that output, then report exit 70 with the callable and call site. TestV4DirectCalleeOrder pins this difference.

Compatible calls, methods, boxed representations, optional receivers, null, void, unknown producer parameters, class results, and unused misfit callables match Node. Recursive-domain refusal probes execute successfully in Node; Adamic refuses the unproven relation with path and fix.

Escaping .ts callable adapters defer signature checks to invocation. Source Node runs the misfit function unchecked; both generated backends instead exit 70 after argument effects for an argument misfit, or after producer effects/return for a result misfit. TestV4EscapeAdapterArgument and TestV4EscapeAdapterResult pin stdout and identical generated-backend diagnostics, and omission mutants must finish leak-clean with source Node's output. Unused signature misfits do not fail at read. Current adapter blame names the read site and the escaping-call boundary; the actual later source call site awaits precision 5.
