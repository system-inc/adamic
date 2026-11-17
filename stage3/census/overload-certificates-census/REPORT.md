# Paired overload certificate census — #5fws6q8

In progress. Base: `910f21afffd892c747bd82f72229743a899586ea`.
Head: `812c342689ded7e1cc8a31c35a86f3546f7b0772`, exactly one commit on base.
No adaptation, ledger, or internal change is made by this report branch.

Both measured worktrees initialized recursive submodules before any build:
cohere `5360753e19e4417675a90e92a1823bea66b13918`, TypeScript-go
`d92d9bfee114c80be2c375d72edae966176e3a4f`.

`git diff BASE HEAD -- stage3/apply.sh stage3/apply.py stage3/adapt stage3/api
stage3/source.json` is empty. One unchanged base-adapted tsc tree is being
prepared for both. Each sha independently supplies its latent generator,
meter entry overlay, scan-all generator, compiler sources and binaries.

The first-error entry pass will match the meter: tsc.ts root,
CommonJSBundle, LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1.
Each full run selects exactly one absolute file with LATENT_ONLY_FILE and uses
LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1. The caller list is resolved
with getResolvedSignature from the entry closure and complete src tree. Files
inside the entry closure retain that same tsc.ts root; extra src callers use
their own entry closure, so they are measured rather than silently omitted.

No declaration-pass count, certified-call count, or other-overload conclusion
is established yet. The final report will retain child exits, complete streams,
position-stripped finding tuples, reached-stop distinctions, source hashes,
and probe build provenance. A blocked/incomplete run will be labeled as such.
