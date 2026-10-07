# watchUtilities.ts at zero

Four findings become zero; the fair whole-tree census is 374 to 370, with no
new diagnostics elsewhere. Forty of the same 78 input files now have no checker
diagnostics (36 before these closure waves).

The cached directory host deliberately writes four optional wrappers as undefined
when their host method is absent. Its owning @internal DirectoryStructureHost
now admits explicit undefined. Indexed extraction of each unchanged method
signature preserves its original parameter variance; the writeFile host may
require a boolean even though this existing internal method makes it optional.
A plain function-property conversion was rejected after the full census caught
a new program.ts error. The final method extraction introduces no new error.

The other three findings lose a key-dependent contract inside two private
wrappers. All four constructors use a literal key matching their returned
WatchFactory field; that immutable key selects the matching file/directory
callback and flags. No local factory escapes for external mutation. The selected
host method invokes its logging callback with the same key's argument tuple;
that tuple is forwarded unchanged to cb.call, including optional modifiedTime.
The two selected call receivers and the callback receiver receive erased local
function-type bridges. The logging rest parameter becomes the truthful union
of the existing FileWatcherCallback and DirectoryWatcherCallback argument tuples.
No value is read twice, moved, skipped, defaulted or initialized differently.

watch-protocol.json guards the reviewed entire getWatchFactory AST body by
stock 6.0.3 emitted JavaScript hash. A changed key is rejected before any source
write. Removing each of the eight declaration/receiver/rest annotations fails
its shape guard; reverting them all restores all four census diagnostics.
A required moduleSpecifiers read changed from ! to ?? 0 fails the partition's
independent emitted-byte check and occurrence contract. This file needs no new
indexed assertions: the keyed fields and first callback tuple member are present,
and the optional second callback argument is already handled nearby.

All 26 partition files preserve CRLF and stock-emitted JavaScript bytes. The
second adapter run makes zero edits. The default oracle passes 106,367 tests
with no baseline differences, including the mechanically checked exact API
exceptions: adaptation 20's 189 and 40's 28 lines. Adaptation 70 is not present
in this integration input; no other public API lines are accepted.
