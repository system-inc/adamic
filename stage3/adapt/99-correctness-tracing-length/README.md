# tracing array truncations

Task ff5eddz. JavaScript-changing adaptation under docs/rulings.md:306 and :307.
Only pop and popAll are owned here, pinned whole before and after with exact CRLF
bytes and SHA-256 hashes. Unknown owners, partial application and declaration
changes are refused. A second application writes nothing.

pop replaces eventStack.length-- with an explicit empty-stack RangeError guard
followed by eventStack.splice(eventStack.length - 1). popAll replaces
eventStack.length = 0 with eventStack.splice(0).

An empty pop at entry throws from Debug.assert, or from writeStackEvent(-1, ...)
when assertions are explicitly stubbed off. However, while writing an event,
JSON.stringify(args) can invoke a getter that reenters popAll and empties the
stack. The original decrement then executes at length 0 and throws RangeError.
The explicit guard preserves that reachable class and "Invalid array length"
message. The reentrant Node witness demonstrates the original unguarded splice
would return normally, and requires the guarded adaptation to throw identically.

Adaptation 20 owns the optional args annotation in the eventStack element type.
32 owns indexed reads, including eventStack[index] in writeStackEvent, and the
private TraceRecord.typesPath annotation. 40 changes erased any annotations.
The typeCatalog declaration has no whole-owner pin in these adaptations.
47 validates startTracing's require("fs") producer; 62 and 63 own dumpTypes and
dumpLegend calls, 65 owns builtin declarations, and 99-correctness-tracing-assert
pins stopTracing whole. None pins pop or popAll whole. This ownership audit is
against origin/main 16429a304.

startTracing is deliberately unchanged. Its typeCatalog.length = 0 belongs with
the concurrent amendment to 47 on typescript/meta-spread-checkpoint, which is to
pin startTracing whole for the meta spread. The upstream 71 / adapted 72 length
write is not claimed by this adaptation.

Validation and replay instructions, including whole-lane counts, before/after
census quotations, Node outcomes, restoring mutants and all emitted hunks, are
in the [proof](https://github.com/system-inc/adamic/blob/typescript/tracing-length-evidence/stage3/tracing-length-evidence/PROOF.md) on
[typescript/tracing-length-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-length-evidence/stage3/tracing-length-evidence).
