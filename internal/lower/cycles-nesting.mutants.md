# Cycle finder nesting (#ykd89v9)

One edit, run on a house box or a sealed one, then put back.

## The enter forced ok

In `use` (cycles.go), keep the call and its deferred leave, and force the answer:

| Old text | New text |
|---|---|
| `	if !ok {` (the first one in `use`) | `	if ok = true; !ok {` |

This mutant survives through lowering. Since #ykd89v9 the expansive-generic pre-pass stops every
program with a generic that instantiates itself deeper before the cycle finder runs, so no program
reaches the bound. It stays as defense, for a growth the pre-pass doesn't see (polymorphic recursion
through an inferred call, an `import()` type).
