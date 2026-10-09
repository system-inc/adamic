# Runtime clearance scope

No runtime file changed. In particular, internal/native/runtime/object.c is byte-for-byte identical to the requested base 3e2166d5.

All six reported programs already agree with Node on that base. Receiver-keyed checked reads leave unrelated object reads on the plain path. The checked-field summary no longer puts their writes through the program-wide name guard. Their ordinary reference slots therefore accept the existing null and undefined values and can subsequently hold an array, Map or string.

The string-or-undefined viewed read in p04 already uses the union-contract selector. No modification to adamic_object_view or adamic_object_view_write is submitted, and no broader writable asserted-view contract is claimed.
