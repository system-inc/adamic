# Streaming stop report, task 20pb346

Stopped before implementation: immediate final file records have a census-wide
future dependency. This is a source-inspection result, not a performance result.
The requested change and performance proofs have not been completed.

Base inspected: origin/main, `31d7a244808fa8092643926d21ab9d310a54cce3`.
Read the requested stall report first, from evidence branch commit
`d84dac36b19a7199705eba1e7d885091d2a5b12b`, path
`stage3/meter/evidence/stall-20261105/REPORT.md`.

## Why immediate emission changes bytes

`stage3/census/latent/lower.go.txt:219` initializes the census-owned
`latentInstantiated` map, accumulates records through `latentRun`, then walks the
records after the run. For each `deferred_generic_instantiation` unit it sets
`Instantiated = latentInstantiated[unit.Where]` before emitting that record.
There is no end-of-run file sort here. Records preserve the callback order from
`latentRun`, including full mode's `program.Files()` traversal in `units.go.txt`.
Keeping that order alone does not preserve the record contents.

`stage3/census/latent/make_overlay.py:117` installs the body-entry witness in
`lowerFunction`: when `latentConcreteCallDepth > 0`, it records the declaration's
location in that map. The body can be reached by a concrete caller in a later
file. At completion of the declaration's own file that witness need not exist.
An immediate callback with the current ledger would encode `"instantiated":false`
where today's final callback encodes `"instantiated":true`. The two JSON tokens
also have different byte lengths.

This behavior has an existing cross-file fixture:
`stage3/census/latent/audit_deferred.py:17` writes `a.a` containing exported generic
functions `called` and `outer`; line 25 writes `z.a` importing and calling them.
Lines 58 and 59 require their final `instantiated` values to be true, in legacy,
full and speculative modes. The unused generic must remain false. This is an
existing executable audit, not a new test or an audit run claimed by this report.

The contract is explicit in `stage3/census/latent/README.md:277`: witnesses are
computed after all files, survive rollback, and include callers in later files.
That README also documents the prior observed single-file versus multi-file
record mismatch at line 310. The evidence is described in
`stage3/census/latent/evidence/carry-facts/report.md`.

## Reader implications and possible scope changes

`stage3/census/latent/unit_status.py:33` counts instantiated generics and builds
the never-reached coverage-gap list from these booleans. The full report and
historical compare reports consume those counts. The census pair tool explicitly
excludes `instantiated` from its stable-unit comparison at
`stage3/census/pair/pair.py:87`, so a passing pair comparison alone would not prove
record identity. This is a content issue, not merely a streamed-order issue.
An exhaustive reader/order equivalence audit was not completed after finding it.

An append-only stream of final records cannot supply the future witness at file
completion in the current one-pass traversal. Possible redesigns include a
witness-discovery prepass or provisional records followed by file rewriting.
Neither has been implemented or measured here. A rewrite needs an explicit
contract for live readers, interruption, and changing record lengths; a prepass
needs proof that witnesses match actual attempts and that the meter becomes
readable sooner. Reordering files alone does not generally solve mutual
cross-file generic calls. This report does not claim all such redesigns are
impossible; it identifies why removing the current buffer is not byte-preserving.

Speculative mode has another future dependency: `speculative.go.txt` buffers
records to resolve failed boundaries and finding depth after all files. The
meter's full mode already encounters the generic-witness dependency above.

## Work and validation status

Only this stop report is added. No compiler or lowering package is changed.
Entry timing and private-field caching were not partially implemented after the
streaming blocker. `cloud/fast-gate/compiler-dependencies.json` was inspected;
no package's compiler use changes and no manifest update is needed.

No new Go tests were added. No overlay build, census pass, profile, growth
sampling, or mutant run is claimed. There are no before/after
wall times or profile shares to report. No census processes were started.
Validation was attempted after the final main fetch/rebase, which reported the
branch up to date. The initial Go invocation found an unrelated system `go`
command (`Go: Unknown option: test`). After sourcing
`/workspace/adamic-tools/env.sh`, the requested `go test
./stage3/census/latent/... ./stage3/meter` stopped before testing because the
uninitialized `cohere/TypeScript/tsc/go.mod` required by `go.work` is absent.
Meter Python discovery ran 65 tests, with 8 failures and 5 skips. The pair tool
suite failed setup because `stage3/api/node_modules/typescript` is absent, so
it ran zero tests. These are unsuccessful validation attempts, not passing
checks. Complete outputs are in the adjacent `*-tests.log` files. No real
census passes were launched by this work.

The branch contains this report and validation logs in one commit, to be pushed
plainly to `typescript/meter-stream-records`, with a final main fetch/rebase and
scope check before publication.

Task: #20pb346
