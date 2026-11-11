# tracing array truncations

Task ff5eddz. JavaScript-changing adaptation under docs/rulings.md:306 and :307.
Only pop and popAll are owned here, pinned whole before and after with exact CRLF
bytes and SHA-256 hashes. Unknown owners, partial application and declaration
changes are refused. A second application writes nothing.

pop replaces eventStack.length-- with eventStack.splice(eventStack.length - 1).
popAll replaces eventStack.length = 0 with eventStack.splice(0).
On an empty stack, pop's Debug.assert throws when enabled. With assertions off,
writeStackEvent(-1, ...) destructures eventStack[-1], which is undefined, and
throws TypeError before the decrement. The negative length RangeError cannot be
reached. The stack is a private ordinary array initialized in this module.

Adaptation 32 owns indexed reads, including eventStack[index] in writeStackEvent.
47 validates startTracing's require("fs") producer; 62 and 63 own dumpTypes and
dumpLegend calls, 65 owns builtin declarations, and 99-correctness-tracing-assert
pins stopTracing whole. None pins pop or popAll whole. This ownership audit is
against origin/main 16429a304.

startTracing is deliberately unchanged. Its typeCatalog.length = 0 belongs with
the concurrent amendment to 47 on typescript/meta-spread-checkpoint, which is to
pin startTracing whole for the meta spread. The upstream 71 / adapted 72 length
write is not claimed by this adaptation.

Validation is in progress. No completed lane or census result is claimed yet.
Evidence is kept outside stage3/adapt and will be published separately.
