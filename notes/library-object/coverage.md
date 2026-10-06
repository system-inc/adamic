# Object branch coverage

Base: origin/codex/library-object at cdf632b47fefa315aefdeaeda468c13abed4d551.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the three branch commit messages, the branch diff, and the changed implementation and evidence files. branch-files.json inventories all 50 changed files, including parsed classification and adapted-source artifacts.

Result: 13 new registered oracle programs, zero output disagreements. Every final program finished with exit 0 and empty stderr on source Node and release native; the uncached oracle also checked JavaScript, ASan/UBSan, release, and leaks. release-results.json contains each exact build command and both complete observations. Only 13 new rows were added to counts.md.

## Cases and existing programs

Names in the existing column refer to internal/oracle/testdata. New names abbreviate coverage_object_<name>.a in that directory. An existing program's presence is distinguished from registration: object_prototype.a exists but is not registered on this branch. The new prototype_keys and nested programs exercise its relevant plain-object methods in the oracle.

| Code condition or value/call case | Existing program before this work | New program or limitation |
|---|---|---|
| keys/names on a complete plain object; canonical integer indexes first, ordinary strings in insertion order | library_object_keys, library_object_order | frozen_copy, nested, integrity_order |
| Integer boundaries 0, 2^32-2 versus 2^32-1; leading zero, signed, fractional, exponent and empty keys | library_object_keys, library_object_order | Already used |
| keys/names number primitives have no keys | library_object_keys (names on 1) | primitive_names uses both methods on zero, negative zero, NaN, infinities and a finite number |
| keys/names boolean primitives have no keys | library_object_keys (keys on true) | primitive_names uses both methods on both booleans |
| strings use UTF-16 indexes; names also adds non-enumerable length | library_object_keys (A plus supplementary character) | primitive_names: empty keys, BMP, supplementary, lone surrogate, combining sequence |
| Returned names arrays are independent and mutable | library_object_keys | Already used |
| Empty object names/keys | library_object_keys (keys only) | empty: both |
| Complete nested object keys/names/hasOwn | None with these methods on the same mixed-key nested shape | nested |
| Homogeneous number/string/boolean values and entries | library_object_keys, library_object_order | frozen_copy on frozen string fields; stable on let/alias fields |
| values/entries empty, heterogeneous, optional or nested fields | No executable case admitted for nested/heterogeneous fields | values_nested, entries_nested probes; scalar homogeneous representation required |
| Plain literal, string/identifier field names, shorthand fields | Literal fields in library_object_keys | shapes: shorthand, parenthesized alias, second alias, direct literal integrity mutation |
| const binding and aliases establish exact shape | library_object_keys, library_object_assign | shapes: alias chain |
| Unannotated unreassigned let establishes exact shape | library_object_keys (keys/names) | stable: values, entries, hasOwn, assign, freeze and integrity |
| Reassigned binding invalidates shape proof | library_object_replaced (unregistered refusal probe) | Existing lower tests cover direct, array/object destructuring and for-of replacements; cannot execute reflection through these unproven shapes |
| Binding annotation, function call, spread, computed/numeric field syntax, NUL, __proto__, private-style key, missing initializer, non-variable declaration, depth >16 do not establish exact shape | Not admitted as complete reflected shapes; lower tests exercise widened/optional/spread views | No successful oracle can reach these validation exits; fixed-shape proof deliberately rejects them |
| hasOwn with declared public literal field; class own fields versus prototype methods | library_object_has_own, library_object_own | nested, stable, integrity_order with integer and ordinary keys |
| hasOwn absent/dynamic/private key | Refusal tests, no successful oracle | Must name a declared public field; prototype methods allow absent/dynamic strings instead |
| assign one, two and three typed sources; last source wins; untouched fields and order stay | library_object_assign, library_object_assign_fields | stable and integrity_order: target is a later source, after an earlier write |
| assign self retains string before releasing old slot | library_object_assign | stable: self is also later source; frozen_copy: frozen source copied to mutable target |
| assign without sources, >3 sources, conflicting/widened types, nested references or new fields | Lower refusal tests | assign_no_sources, assign_nested: any result or reference-cycle/type proof rejection |
| freeze returns same object; alias reads, shallow freeze, primitive identity | library_object_freeze, library_object_freeze_alias | frozen_copy, nested, stable |
| Frozen field write or assign raises TypeError | library_object_freeze_write, library_object_freeze_alias, library_object_freeze_assign | Already used; no differing fixture added |
| Frozen spread copies rather than reuses; copy is extensible/unsealed/unfrozen | library_object_freeze | frozen_copy reads original and copy; nested preserves shallow child mutability |
| seal/prevent primitive fast paths: number, boolean, string, undefined, null | library_object_freeze covers most; no undefined or false mutation there | primitive_integrity: undefined, false, string prevention, negative zero and NaN |
| Integrity mutation on nonempty exact plain literal/binding returns same object | library_object_freeze | shapes: direct literals; stable: let; integrity_return: return owned identity through functions |
| Queries on primitive null/undefined/number/boolean/string | library_object_freeze | primitive_integrity fills isFrozen null/false/empty/NaN |
| Query on non-object heap kind: array/map returns extensible, unsealed, unfrozen | library_object_freeze | primitive_integrity adds closure and Set |
| Query through proven-safe structural object view | None with all three states | class_queries: initial, sealed, frozen view; integrity_return: returned view |
| Query with scalar/object/undefined tagged union | None | integrity_union: primitive fast path and all object states |
| Query on class with private internal slot | None | class_queries: initial flags are clear; private-slot skip in a nonextensible class cannot be reached because class integrity mutation is refused |
| Extensible empty/nonempty object is neither sealed nor frozen | library_object_freeze | empty, integrity_order |
| Nonextensible empty object is sealed/frozen without field loop | library_object_freeze (preventExtensions) | empty also covers seal and freeze |
| Nonextensible nonempty object with configurable fields is not sealed/frozen | library_object_freeze | integrity_order: prevent then seal |
| Sealed writable nonempty object is sealed but not frozen | library_object_freeze | integrity_order: self assign; nested and stable |
| Frozen object is sealed and nonextensible | library_object_freeze | nested and integrity_order: seal/prevent after freeze, repeated seal |
| Nonextensible/sealed spread must allocate fresh extensible shape | library_object_freeze | Already used |
| Heap and region object construction initializes new metadata | library_object_freeze; existing regions program constructs region objects | New ordinary and nested objects query initial states; direct proof of nondefault region flags is limited by allowed mutation/escape rules |
| SameValue numbers: NaN equal; signed zeros distinct; equal/different scalar kinds and object identities | library_object_is, library_object_same | stable/nested identity cases |
| Exact null in either/both positions gives statically proven result | library_object_same (literal null, number-side effects) | null_order: both null operands and undefined/null/object operands with observable evaluation, once left-to-right |
| Nullable union operand must be refused before indistinguishable tags compare | Lower integrity tests | is_nullable: parameter representation rejects it before Object.is dispatch |
| toString/toLocaleString/valueOf on plain objects | object_prototype (unregistered) | nested and prototype_keys: mixed keys, nested/frozen receiver, identity result |
| hasOwnProperty/propertyIsEnumerable with present/absent/inherited dynamic string keys | object_prototype (unregistered), has_own | prototype_keys: mixed integer/ordinary keys, absent canonical spelling and inherited member |
| Own callbacks shadow inherited prototype methods | object_prototype (unregistered) | Existing program uses them; not changed by this branch |
| Call inherited method on receiver or wrap in arrow | object_prototype receiver calls | prototype_keys: dot and arrow; bracket calls rejected |
| inherited isPrototypeOf, detached/call/apply forms | No admitted inherited prototype-chain observation | prototype_chain and prototype_bracket: observable prototype chain and inherited member reads are refused |
| Invalid arity/spread and unsupported host shapes, tuple descriptors, mutation on arrays/maps/classes | Existing lower/prototype refusal tests | Compile-time exits, not execution paths; impossible to compare a native executable for them |

## Unsupported requested forms

These are compilation failures, not disagreements between executing outputs. Each .a probe in this folder has its Node behavior and exact build diagnostic in unsupported-results.json. No probe is registered as a passing oracle fixture.

- from_entries_repeated.a: Object.fromEntries explicitly rejects its unproven index-signature result. Node prints `2|plain|01` then `9|1|3`; there is no native executable.
- values_nested.a and entries_nested.a: nested values have no admitted scalar homogeneous element representation.
- assign_nested.a: reference fields and possible cycles are explicitly refused.
- assign_no_sources.a: TypeScript's result is any, which Adamic refuses.
- freeze_null.a: null is missing from the primitive freeze fast path and cannot establish a plain shape.
- is_nullable.a: number/null/undefined union parameters lack a representation.
- keys_array.a: array holes/descriptors are not represented by own-name reflection.
- nested_bracket.a: integer-named plain-object element access is not lowered.
- null_return.a: functions returning null (and exact undefined) are not lowered. The passing null_order fixture uses evaluated conditional expressions instead.
- prototype_bracket.a: inherited bracket method calls are treated as forbidden member reads. The passing counterpart uses dot calls.
- prototype_chain.a: inherited isPrototypeOf is explicitly refused.

The branch's guard-only validation paths cannot be made into successful programs, and the existing Go refusal tests cover those paths. Arbitrary descriptors, prototype mutation, host constructors, symbol keys and deep recursive shape proofs remain outside the branch's represented subset.

## Dependency proof

One line in internal/native/runtime/object_names.c was changed from `if (has_length) {` to `if (has_length && length > 0) {`. coverage_object_primitive_names.a compiled and finished in both modes, but Node's second line was `length` and native's second line was empty. The oracle reported `stdout differs`; this was not a compiler or sanitizer failure. mutant.txt contains the full receipt. The exact original file was restored and the same program passed uncached. No compiler change is committed.

## Commands

All final toolchain commands sourced /workspace/adamic-tools/env.sh. Test output went to files and those files were read.

```sh
git fetch origin main codex/library-object
git fetch origin refs/heads/codex/library-object:refs/remotes/origin/codex/library-object
git switch -c coverage/library-object origin/codex/library-object
git log --format='%h %s%n%b' origin/main..origin/codex/library-object
git diff origin/main...origin/codex/library-object
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_object_' -count=1 -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
# For each of the 13 final .a files, with its basename as NAME:
go run ./cmd/adamic build internal/oracle/testdata/coverage_object_NAME.a -o /tmp/adamic-gate/object-release/coverage_object_NAME
/tmp/adamic-gate/object-release/coverage_object_NAME
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/coverage_object_NAME.a
# For each unsupported .a probe, the same build and source-Node commands were run.
# With the one-line mutation, then restore:
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_object_primitive_names.a' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_object_primitive_names.a' -count=1 -timeout 30m
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
git diff --check
```

The focused oracle was run while drafting, producing refusal-only failures for the unsupported syntax retained as probes. It was run again after each adaptation, finally with all 13 passing. Direct builds were also repeated for the final sources. Setup was run twice: the first warm-up was invalidated by my checkout change; the second completed successfully on the stable branch.

Final setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (60s)
setup: done in 60s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

## Base-branch flow check

The full gate encounters the pre-existing, intentionally refused internal/oracle/testdata/library_object_replaced.a in four internal/flow tests. The failure was independently reproduced on the unmodified base branch in a detached worktree, with no new programs present. base-flow.txt contains the receipt. The probe and flow implementation were not changed.

```sh
git worktree add --detach /tmp/adamic-gate/object-base origin/codex/library-object
rmdir /tmp/adamic-gate/object-base/cohere
ln -s /workspace/adamic/cohere /tmp/adamic-gate/object-base/cohere
cd /tmp/adamic-gate/object-base
go test ./internal/flow -run TestEveryFunctionIsInSingleAssignment -count=1 -timeout 30m
```

This returned exit 1 on library_object_replaced.a: Object.values cannot use its replaced binding as a proven complete shape.

## Final gate result

Formatting (`gofmt -l cmd internal`), vet (`go vet ./...`), and `git diff --check` produced no findings. The full uncached gate exited 1 solely because four internal/flow tests discover the base branch's intentionally refused library_object_replaced.a. Every other package passed, including the entire oracle, native, Unicode and stage packages. full-gate.txt records the complete output. The base reproduction above confirms this is pre-existing.

Commit command: `git commit -m "Add oracle programs for Object branch coverage."`
Push command: `git push -u origin coverage/library-object`
