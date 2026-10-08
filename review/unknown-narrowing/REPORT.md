Built: runtime unknown/object narrowing, presence checks and tagged property reads, including host fs code/message.
Commits: implementation 3edfb8fb99aac52e31f893bd2c65f53e08f8ff54; host-fs merge 0359837 (d46ab6a); base origin/main f8013f0.
Validation: touched packages, filtered Node oracle on both backends, all recorded allocation counts, go vet and formatting passed.
Mutants: unconditional in and skipped inner typeof changed Node stdout; null dispatch, missing retain and merged scalar layouts failed runtime/oracle checks.
Not covered: pending require branch, full repository gate, and the explicit representation refusals below.

The prior rejection was the unconditional InKeyword refusal plus missing representations for unknown/nonprimitive and object & Record intersections. Lowering now keeps these values in the tagged union representation. Presence has its own IR operation; reading a property produces unknown until the checker proves its tag. Native fixed-shape fields carry scalar type metadata; host error reference fields retain their existing tagged values. Null has a distinct immortal tagged sentinel, including when passed to an unknown parameter. Both emitters implement the new operations.

Presence follows own fields, native class method/accessor tables and the represented Object/Array prototypes. Private storage slots are excluded. Arrays support length, canonical existing indices and prototype presence. Dynamic prototype function values and dynamic array indexed reads remain unsupported. Class scalar fields and host-built code/message strings are supported.

Fixtures cover EEXIST, absent code, number code, string/number/null/undefined object guards, classes and inherited methods, arrays and indices, private fields, caught Error instanceof Error/message, and String on a property that remains unknown. Host import fixture covers ENOENT, primitive/null none, the actual host message, and an existing-directory EEXIST catch. Host library dependency d46ab6a was merged, including its additive lowerer and counts conflicts. Stage 3 fixture 13 and the host adaptation generator at cf976e8 were inspected. The require dependency was absent on origin; review/unknown-narrowing/require_pending.a preserves that requested form without registering it as passing.

Validation commands and observations:

- `bash cloud/setup.sh > /tmp/unknown-setup-retry.log 2>&1`: passed. Timing lines: go 0s; clang 0s; node 0s; submodules 0s; build cache warm 89s; done 89s. `nproc`: 5; cpu.max 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. Initial setup warm overlapped my checkout merge and failed on conflict/undefined lowerer types; rerunning after resolving the merge succeeded.
- All subsequent Go commands sourced `/workspace/adamic-tools/env.sh`.
- `go test ./internal/lower ./internal/native ./internal/fresh ./internal/ir ./internal/javascript -count=1`: final combined run passed native (88.597s), fresh (47.856s), IR (14.238s); JavaScript has no package tests. Lowering had one outdated expected refusal diagnostic; correcting that test and rerunning `go test ./internal/lower -count=1` passed (15.992s). Logs: /tmp/unknown-packages-final.log and /tmp/unknown-lower-final.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestUnknownNarrowingMutants|TestNativeAgreesWithNode/internal/oracle/testdata/(unknown_narrowing|unions|json_stringify_values)|TestCountsAreRecorded' -count=1 -v -timeout 30m`: PASS, 30.433s, 366 native cache misses and 12 Node cache misses. Log: /tmp/unknown-final-oracle.log. Both new fixtures match Node through native and JavaScript backends. Counts checks also compile all registered fixtures; this is not the complete Node oracle gate.
- General fixture: 31 allocations, 31 frees, 42 retains, 63 releases, peak 3; host fixture: 8 allocations, 8 frees, 11 retains, 16 releases, peak 4. Existing counts unchanged; one existing row reordered to harness order.
- `go vet ./internal/lower ./internal/native ./internal/fresh ./internal/ir ./internal/javascript ./internal/oracle`, gofmt and `git diff --check`: passed. /tmp/unknown-vet-final.log.
- Node declarations missing from setup were installed locally using `npm install --prefix stage3/api --no-save --package-lock=false @types/node@25.3.3`; these dependencies are not committed.

Mutants actually run:

1. HasProperty always true: permanent IR mutant test; absent-key and primitive outputs disagree with Node in both JavaScript and native backends.
2. Skip the property string typeof: permanent IR mutant test; numeric 42 prints instead of undefined and Node stdout catches it in JavaScript. Native is not run for this mutant because the removed proof exposes an unchecked string representation cast. Regular unmutated fixtures are tested in both backends.
3. Remove null sentinel check in native typeof dispatch: uncached oracle failed with ASan global-buffer-overflow. /tmp/unknown-mutant-null_typeof_dispatch.log.
4. Remove the dynamic field retain: uncached host oracle failed with ASan heap-use-after-free. /tmp/unknown-mutant-drop_dynamic_field_retain.log.
5. Collapse number/boolean shape metadata to reference-bit layouts: uncached oracle failed with stdout disagreement. /tmp/unknown-mutant-merge_number_and_boolean_shapes.log.

The three temporary source mutants were restored before final verification and committing. Script and exit results: /tmp/unknown-runtime-mutants.py and /tmp/unknown-runtime-mutants.log.

Explicit refusals and practical limits:

- Only literal string keys are lowered for in. Dynamic coercion, symbol keys and NUL-containing keys lack native name representation.
- Tuple views, opaque collections/host types (Map, Set, RegExp, Date, Stats, Hash, Buffer), functions, intrinsic identity objects and class constructors lack the required dynamic descriptors. Ordinary arrays, objects, instances and caught errors are covered.
- Dynamic getter/method/prototype value reads need callable dispatch or ToPrimitive support. Presence of represented inherited methods is supported. Methods omitted from the native dispatch table are refused for presence.
- Nullable field slots cannot distinguish null and undefined dynamically, so such unknown views/reads are refused. Possibly absent spreads and ambient class fields lack JavaScript presence descriptors and are refused. These checks are conservative across the program and may reject unrelated shapes with the same key.
- Error stack/cause and host errno/syscall/path presence lack descriptor metadata and are refused; own code/message are supported.
- String conversion of arbitrary unknown remains refused. The supported unguarded property-read conversion requires a closed-program proof that the named property values are scalar; a const alias is supported.
- Computed class field declarations and general unknown instanceof beyond the existing caught Error path retain their earlier limitations. No broad index-signature record representation was added.

No edits were made to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or internal/oracle/oracle_test.go. No code was copied from cohere. No full gate or pending require execution is claimed.
