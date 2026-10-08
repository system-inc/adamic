Superseded by the user correction: NotYet workers stay on area/compiler and
must not merge views integration. No ABI resolution or approval is now needed
for this unit. The attempted merge was aborted before any merge commit existed.
Checked non-null c41c0e06 remains; ! in .a stays refused.

The isolated views certification checkout was not pushed, and its tests and
native emitter change are not carried onto codex/notyet-element-access. Its
observations are outside this compiler-area unit and do not increase coverage
on this branch. The original merge-attempt record follows for audit only.

Views prerequisite merge is pending automatic approval review.

Requested views tip: ffe428ab26eefb73154adbf570ad99e9c1b4f872.
Requested checked non-null c41c0e06 is already an ancestor of this branch, merged
at c7095bc7. Attempting the views merge produced conflicts in 20 files. The
merge was aborted after review rejected the unverified ABI resolution; no
partial compiler changes remain on codex/notyet-element-access.

The exact representation collision is in internal/ir/ir.go. This branch gives
Uint8Array, Int32Array and Float64Array the values 12, 13 and 14. Views integration
reserves slot tags 12 and 13 for null and undefined and gives Record value 14.
Keeping both unchanged cannot distinguish the representations. The proposed
resolution is to keep the integration tags and assign typed arrays 15, 16, 17.
This is a proposal, not applied code or a certified ABI change.

The typed-array native constructor mapping uses symbolic ir.Type constants in
internal/native/typed_arrays.go and distinct C adamic_typed_array_kind values.
But object storage and checked view code carry numeric ir.Type slot tags:
internal/native/runtime/object.c has explicit 12/13/14 handling. Therefore the
combined merger needs typed-array field, union, array, checked-view, ownership
and sanitizer tests, not just successful compilation. The additional merge
conflicts include checked non-null .ts-only policy versus earlier integration
.a acceptance, predicate/non-null reporting, optional array accesses versus
checked array views, fixed tuple indexing and finite own-field dispatch.

Automatic approval review rejected a batch conflict resolver as broad and ad
hoc. A subsequent narrow IR resolver was also rejected: reassigning core tags
could silently miscompile values if runtime/backend paths rely on numeric tags;
the merge request did not authorize that unverified ABI change. Approval to
resolve this explicit ABI collision is required before applying the proposal.
No rejected mutation was executed. Five independent merge conflicts were
resolved temporarily before the second rejection, then restored by merge abort.

Unaffected work continues in /workspace/adamic-element-views on the owned branch
codex/notyet-element-access-views-check, starting from ffe428ab with its ABI
unchanged. It certifies previously blocked element-access shapes in isolation.
Its certification is not a claim that the prerequisite has landed on the
original element-access branch. Only owned branch names are written.
