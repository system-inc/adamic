The exact proposed fix is proposed-opaque-diagnostic.patch. It is evidence only, not applied.

The user authorizes runtime changes subject to reporting: "Runtime C files (internal/native/runtime): if any runtime line changes versus ... own runtime lines, list every changed line; I take it to runtime for clearance." The overall instruction also says "make it right before it goes."

Static scope proof: view_unions_mixed.h describes logical kinds and requires complete owning adapters for reference-member certification. The existing unchanged native test deliberately passes a four-byte int address as an opaque object placeholder with no adapter or no contract. The source commit now reads adamic_object tuple/class fields before rejecting; ASan proves an out-of-bounds read.

The proposal removes that unsafe inspection from the generic selector, restoring main's rejection behavior. It moves the existing diagnostic-only classification to the end of adamic_view_untagged_plain_select, after all matching has failed. That function already normalizes and dereferences the heap object, establishing its layout. It changes only a local failed-selection snapshot immediately before the nonreturning diagnostic call. Accepted matches return earlier; matching, contracts, ownership, allocations and the diagnostic formatter are unchanged. No test is weakened or edited; no new API or runtime feature is added.

Required validation if authorized: rerun existing missing-adapter ASan control, all D tests, diagnostic kind mutants at their relocated lines, and reverse-fix mutant; revalidate final counts, reader guard and vet as appropriate. Every additional runtime line will be listed for the user's runtime review.
