# Slot 04 loader and CSS registration helpers

These isolated helpers port Go orchestration and mutation. External engine,
filesystem and parsing helpers remain explicit dependencies; this does not
integrate a complete Tailwind engine or rule. Each production helper owns one
.a file. The runner uses the existing options reader only for test data.

`loadDesignSystemThrough` takes callbacks for the program's project root,
entry point lookup, platform directory extraction, package lookup, engine load,
recording stat and descriptor-table construction. Callbacks must use the same
recording filesystem. Numeric handles identify externally owned system/table
objects; zero is nil. Loaded success must hold a nonnil system. Failed load is
an explicit boolean independent of error text, including an empty error.
The adapter preserves error identity/message separately: this helper exposes
presence/message, while Go error wrapping identities remain adapter-owned.
Returned results have no reads snapshot: the separately owned program wrapper
sets that after the load, as Go does. Stat outcomes are ignored as Go ignores
them, but every stylesheet is stated in order, duplicates included. Exceptions
or panics inside dependencies are outside this successful-callback API.

`ingestUtilityBlock` takes a collector, raw parameters, a readonly arena body,
source path and exact Go TrimSpace callback. `kinds` uses membership bits 1 for
static and 2 for functional; adapters supply only 0..3. Every accepted operation
preserves both memberships. Functional definitions append, including duplicate
roots; statics replace their own body only. Body arrays are retained directly
and must remain shared and immutable. Empty-string return is nil Go error;
rejection messages are exact. Caller-created invalid kind bits, raw invalid
UTF-8 strings and distinctions between nil/empty slices are outside this view.
Go recognizes Unicode whitespace; do not wire JavaScript trim as a substitute.

`normalizeValueFunctionNodes` takes a finite parsed value arena, root indices
and printer, top-level comma segmenter, argument normalizer and parser callbacks.
Function targets are exactly --value and --modifier. Other functions recurse;
other kinds are left alone. Each target prints its whole child slice, segments
that text, normalizes each argument, joins with commas, then replaces its child
slice from the parser. Reparsed children are not traversed again. The parser
must allocate fresh edges/nodes in the same arena. Valid indices and tree input
are required; cycles, synthetic pointer alias graphs and nil dependency objects
are outside the parser-derived boundary. The runner's dependency tables are
exported from actual Go segment/normalizer/parser results; they are test data,
not a production replacement for those helpers.

The six consuming rule names and residual blockers are in readiness.json.
That file subtracts these three helpers only from the original frozen ledger.
It removes 18 dependency entries and completes zero additional rules.

The Go overlay exports real private utility ingestion and normalization. The
loader oracle copies the exact upstream Go body, changing only its parameter
and dependency names for controlled external effects. A drift check binds that
copy to the pinned upstream source. Neither Go cohere nor shared Adamic compiler,
registration or harness files are edited. Temporary capture overlays exercise
real rule suites and are deleted after capture.
