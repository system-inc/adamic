The cherry-pick applied only 05eb9043^..05eb9043 to dbdf977f5b. These are every marked conflict hunk. Exact three-way texts and resolutions are in conflict-hunks.json.

- `cloud/admission-corpus/negative-witnesses.json`, hunk 1: Retain main schema/comment and install the source eight exact-hash ruled witnesses.
- `internal/javascript/javascript.go`, hunk 1: Preserve placeholder global declaration check and main async readiness exclusion.
- `internal/native/emit_objects.go`, hunk 1: Keep main slot-derived metadata index and source distinction between unset payload and uninitialized storage.
- `internal/native/reuse.go`, hunk 1: Keep main slot-derived metadata index and source distinction between unset payload and uninitialized storage.
- `internal/native/runtime/adamic.h`, hunk 1: Use one main packed atomic cache snapshot, adding source per-object presence check on its local slot.
- `internal/native/runtime/class_static.c`, hunk 1: Publish the field while retaining main slot-derived readiness bookkeeping.
- `internal/native/runtime/object.c`, hunk 1: Skip absent construction fields and retain main zero initializer for packed cache and owner-aware copy below.
- `internal/native/runtime/object.c`, hunk 2: Keep atomic packed lookup and add per-object presence without rereading shared cache.
- `internal/native/runtime/object.c`, hunk 3: Initialize and publish the own slot, preserving main rule not to initialize a static parent.
- `internal/native/runtime/union.c`, hunk 1: Retain source construction scalar tagging and main acquire publication of immutable shape metadata.
- `internal/oracle/counts.md`, hunk entire table: Restored the pinned main file verbatim, then regenerated once on Linux with clang 20.1.8 after compilation; no hand merge. 41 added and 21 changed rows, all source-attributed, no removals.

Additional compatible-merge adaptations: placeholder.go now asks the runtime for the actual static owner and derives metadata indices from the returned slot. Main's owner-aware field helper is exposed in adamic.h/object.c. library_object.c publication uses a returned-slot index. object.c reserved-field writes search immutable shape names instead of re-reading a shared packed cache. Every runtime edit-line difference is listed in runtime-line-differences.md.

Main changes retained: f93a65fb async/parallel lowering, fcf5c2ad concurrent runtime infrastructure, and e3282288 traced admission compiler products and ledger schema. Packed cache snapshots, slot-derived indices, static owner tracking, acquire/release shape registration, and async readiness exclusion remain.
