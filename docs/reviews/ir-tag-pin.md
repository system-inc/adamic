# IR representation tag pin

Rider for area/runtime; no representation is renumbered. No AGENTS.md existed in
the checkout. The Oct 8 07:44 system_adamic ruling belongs to the later meeting
with area-views: null 12, undefined 13, Record 14, Uint8Array 15, Int32Array 16,
Float64Array 17, Uint16Array 18, Promise 19. This rider preserves today's 1–16.

## Numeric consumers outside ir.go

- runtime/map_set.c: collection_reference and collection_next mirrored Number 1,
  Boolean 2 and MaybeNumber 7. They now use adamic_ir_* definitions in adamic.h.
  Collection iterator state stores key/value representation IDs in number slots.
- native/library_map_set.go: emits key/value ir.Type numbers to that C ABI.
- native/class_accessors.go: emits accessor descriptor types, setter argument
  types and conversion comparisons from Go constants (Number, Boolean,
  MaybeNumber, Union). adamic.h stores descriptor type as int; object.c forwards
  it to the generated setter. These descriptors live in emitted C.
- native/library.go: disk runtime archive and process build keys now contain the
  complete Go numbering fingerprint, alongside all runtime C/header bytes.
- native/units.go: object disk/process keys use runtimeKey, inheriting that
  fingerprint; emitted and preprocessed C/header bytes are also hashed.
- oracle/cache_test.go: native result/count/leak gate observations include emitted
  C and the runtime archive identity; nativeResultKey and nodeResultKey now also
  include the numbering fingerprint directly. Count contexts flow through those
  result keys. Cache envelopes retain their existing content integrity hash.
- ir/tags.go: the named numbering registry is the new fingerprint input.

The adamic_kind heap enum is a separate allocation/cleanup discriminator:
union.c's boxing, equality and typeof switch use those symbols, not ir.Type
numbers. adamic_typed_array_kind is a separate 1–4 element-width enum;
native/typed_arrays.go maps Go types to C symbols explicitly. Neither should be
compared numerically to ir.Type. count.h counts operations, not representation IDs.
record.c's numeric switch is string length; input/regexp/Unicode numeric switches
are unrelated protocols. JavaScript uses representation switches to choose code
and constructor names; no numeric IR tag table is emitted. No serialized IR or
other persisted tag table was found.

## Tests

TestIRTypeTags reads the embedded C header and compares every named definition
against Go constants; unknown, duplicate and missing definitions fail.
TestTypeTagsComplete reads the Go Type constant block and checks registry
coverage. TestTypeTagFingerprintChanges changes each registered value in turn
and requires a different digest.

Validation passed:
- cloud/setup.sh with the requested GOPROXY, followed by sourcing the printed env.
- go test ./internal/ir -count=1
- go test ./internal/ir ./internal/native -run 'Test(TypeTag|IRTypeTags|RuntimeKey|UnitKey)' -count=1
- go test ./internal/native -run 'Test(IRTypeTags|Runtime|Unit|Units)' -count=1
  (includes the existing runtime protection mutants).
- go test ./internal/oracle -run '^TestGateCache' -count=1
- Restored pin and TestLibraryMapSetIteratorResources passed.
- git diff --check

Mutant: changed only adamic_ir_Promise in adamic.h from 16 to 99, ran
go test ./internal/native -run '^TestIRTypeTags$' -count=1, and required
nonzero exit plus the diagnostic:
    Promise: C=99 Go=16
The test failed as required. Restored 16 and reran successfully. No mutated
number or change to ir.go is retained.
