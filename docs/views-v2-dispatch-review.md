# V2 dispatch scope dependency

V2 completion is being verified locally; no new candidate has been pushed. The published V1 dispatch hook remains 8e09bce347a054271d56f14cd124eac2eb0e78a7. The local production union adapters, producer certificates, representation migration and restored admission assertions are uncommitted.

The compiler ruled that array membership belongs to V3. V2 array descriptors carry the named unsupported arm, demanded reads return NotYet naming views-v3: array element kind, and both matchers fail closed for injected array descriptors. No array element_kind or hole accessor prerequisite is imported. Array admission tests remain pending in the acceptance-dependency form.

The element representation prerequisite is af0a0ae70f58c5e63ef7cab7d6eb7ff89bbb4518: adamic.h adds element_kind to adamic_array; array.c initializes it; native allocation emission records physical element representations. The hole accessor originates in 41603d77d89c99b9634766d8ce01b0c13d727843: adamic.h adds sparse storage and the accessor API, array.c initializes sparse storage, array_holes.c implements the accessor, and heap.c manages its ownership. These dependencies require an owner-approved prerequisite or an explicit ruling before importing them. Inferring physical storage from the asserted target would be unsound; bypassing array checks would admit unchecked unions.

Shared runtime hunks currently applied locally, all derived from V2’s own commits:

- internal/native/runtime/object.c: include the union probe header and add the readiness-aware snapshot (0af43ea7213a80b80193d885e25eb60c2ebb0cf3); allow initialized undefined reference slots to receive reference writes (f426929a0cf100ab74d6dde9f78b62a852008055); initialize tuple identity to false (f753567dd8837e90914e7c229df88d8536a3bf8b).
- internal/native/runtime/adamic.h: add tuple identity to adamic_object (f753567dd8837e90914e7c229df88d8536a3bf8b).
- internal/native/runtime/region.c: initialize tuple identity to false (f753567dd8837e90914e7c229df88d8536a3bf8b).

No additional array runtime hunk, graph admission or ownership emission change has been applied. The normalized snapshot preserves static-owner readiness and interprets representation 9 as the stack’s packed maybe-boolean convention.

Verification evidence:

- /tmp/views-v2-active-contract-tests.log: the formerly pending lower contract assertions passed.
- /tmp/views-v2-native-runtime-build.log: TestViewMixedUnionUnknownAndUnavailable fails in both release and sanitized variants while compiling view_unions_untagged.c: undeclared adamic_array_holes_at and missing element_kind.
- This compilation failure is not a mutant kill. Node differential source admissions, representation mutants, wrong-adapter mutant and the full slice sweep remain unverified. No landing claim is made for these additions.

Current completion evidence supersedes the extraction status above. Non-array mixed, object-plus-primitive, untagged object and certified callable reads now use the shared checked dispatch. Fixed object-backed tuples are admitted; array matchers remain unavailable. Nullable array, Map and object slots retain their previous reference ABI. Phantom primitive normalization applies only to checked-view contracts; ordinary branded generic returns remain NotYet.

Runtime review commits: 845a6ba6c reads readiness and type bytes by the returned slot index and includes both slot tails in allocation sizing; 88202e158 guards initialized-undefined reference writes; 572f51a34 stores tuple identity beside frozen and initializes it along the existing allocation header path. Runtime 05ab568e is absent, so the tuple initializer must move with that header path when its allocator change lands. Copies already copy both per-slot tails. No ownership emission or array runtime change is included.

The cache-index mutant cannot be distinguished by a valid fixture on this base: cache hits return slots[cache.index], and misses update cache.index before returning that same slot. The packed atomic cache described by runtime is absent. This is an untested concurrency distinction, not a claimed mutant kill.

The configured counts update passed and adds only 49 V2 rows. The package sweep passed IR, flow and JavaScript, with one lower failure: TestNodeFSFileScratchOptionsBorrow refuses an evaluated fs option expression. Its base comparison remains to be run. Focused V2 source differentials include sanitized native, release native and JavaScript; representation mutants cover null, undefined and all three emitted typed-array tags. Membership, tuple identity, producer certificate and wrong-family mutants are pinned. Full oracle, stage3 record moves, platform-guard sweep, stage1 comparison and the outside-package .a audit remain outstanding.

Pending admission skips use acceptance dependency: compiler/views-v3 787cea7a, with comments naming missing array element-kind and hole support. This V3 tip does not yet provide that support. The pure phantom-brand contract graph assertions are active; this does not admit ordinary branded generic returns.

Runtime stack compatibility correction: the fallback slot-index expression now lives as a guarded macro in view_representations.h, after adamic.h, so it never defines a duplicate static inline. view_write indexes both reference flags and type bytes by the returned slot. adamic_object_size reserves sizeof(size_t) per slot for c1 insertion order, and both current allocators use that function. The base has no order accessor or dynamic adamic_object_copy_reserving_checked allocator. That allocator must initialize tuple = false beside frozen when c1 is merged; the scratch proof applies that integration hunk explicitly. This absent allocator is not imported into V2.
