Claim: immortal retain and field-store release elision, October 7.

Base: merge graph-regions fallback 410cb1c; fetched area/runtime 94a9c83 does not contain it.
Changes: new internal/lower/field_writes.go with reusable collectFieldWrites and exported entry; immortal helpers in internal/native/branch_shape.go; retained's small named static test in emit_ownership.go; SetProperty release proof in emit_statements.go; immortal check before existing dispatch in graph_regions.go. New pinned tests cover rejected heap writes and sanitized ownership mutants.

All field names share writes conservatively across structural views and class inheritance; unknown writers disable completeness. No runtime code, field-store layout, stack-check placement, or symbol naming edits. Requested permission separately for one per-emitter proof cache field in emit.go; absent permission, that file remains untouched.
