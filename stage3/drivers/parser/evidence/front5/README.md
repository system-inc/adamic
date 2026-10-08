Fresh, detached scratch integration starts at compiler/stage3-front-2 860a0d5
and explicitly verifies import-cycles 779ff9d is an ancestor. Merge area/stage3
4ad53a4, library/qualified-as-const 8280fd0, records-lowering 70fb62b and
runtime-records 754e666. Runtime records is already in the records-lowering
ancestry. Scratch head 4831de9; never pushed. Exact SDK pins remain cohere
7945d102 / TypeScript d92d9bfee. Scratch module-path fixes are evidence only.

Node-loader conflicts and resolution are the same as front3: load.go and
source_fs.go preserve typed paths and Node globals; cast.go retains the front's
checked class/tag cast proof, with the qualified-const identifier guard in
cast_proof.go. Counts keep front rows. Record conflicts:

- expression.go: retain enumNeverIdentity/enumNeverValue; inside enumNeverValue
  check enumExpression then recordExpression, each with its own guarded return.
- refusals.go: preserve nodeLibraryRefusal, enumRefusal, predicates and checked
  downcasts; also retain recordLiteralRead, index-signature storage restrictions,
  delete target checks, and recordStorageView. No check is removed to get green.
- oracle/counts.md: keep front counts; no full count gate is claimed.

Two record filename predicates need string conversion for the new typed-path
SDK, in expression.go and records.go. This is a compile adjustment, not a
textual merge conflict. Exact source integration is in the compressed patch.
Compiler builds, and focused load/lower Node, Cast, ImportCycle and Record tests
pass (command and output in report.json).

Parser build is still red at core.ts:11:52: indirect call/class construction
before enum initialization, at new Map<never, never>(). The previously committed
native-enum-map.a reproduces the same NotYet on this compiler. No workaround is
introduced. No native parser binary, actual parser diff or parser-output mutant.
All three Node dumps match the required 35,456,964 bytes / SHA256 2014ef06...;
best user time 5.531759 seconds. perf is absent. Entry is still 1,990 code
declarations; the separate diagnostic-driver root gives 1,991.

The requested type-only MapLike<T> question has a measured positive answer on
this merge: native-type-only-map-like.a builds and native/Node both print
type-only, empty stderr, exit 0; byte comparison exits 0. Flipping exactly one
byte of this real native output makes comparison exit 1. This is the type-only
probe, not parser output; it does not prove all runtime generic record forms.
