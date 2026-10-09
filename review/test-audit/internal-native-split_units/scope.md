Code under test and oracle were stated before selecting M1-M4 in the session.

Production mutants change the C string runtime. S1-S4 change the suite construction, under the brief's setup-check exception. The oracles are executed Node string operations plus self-written membership, ownership, and allocation invariants.

Clean Go coverage reached these package functions (build preparation is not mutated):

```text

github.com/system-inc/adamic/internal/native/library.go:32:				RuntimeLibraryForSource		100.0%

github.com/system-inc/adamic/internal/native/library.go:46:				featureFlags			80.0%

github.com/system-inc/adamic/internal/native/library.go:56:				runtimeLibrary			80.6%

github.com/system-inc/adamic/internal/native/library.go:102:				readRuntime			75.0%

github.com/system-inc/adamic/internal/native/library.go:121:				runtimeKey			100.0%

github.com/system-inc/adamic/internal/native/library.go:143:				cachedRuntime			18.6%

github.com/system-inc/adamic/internal/native/library.go:212:				RuntimeLinkFlags		66.7%

github.com/system-inc/adamic/internal/native/library.go:219:				slicesContain			50.0%

github.com/system-inc/adamic/internal/native/native.go:26:				cString				81.8%

github.com/system-inc/adamic/internal/native/native.go:78:				Flags				86.4%

github.com/system-inc/adamic/internal/native/native.go:110:				Build				71.0%

github.com/system-inc/adamic/internal/native/target.go:10:				ValidateOptions			18.8%

github.com/system-inc/adamic/internal/native/target.go:37:				compilerName			40.0%

github.com/system-inc/adamic/internal/native/view_unions_read.go:8:			init				100.0%

```

The C entry points called by the harnesses are listed in probes.json. adamic_release, adamic_write_line, and adamic_output_flush are lifetime/output support, not empty-answer entries for this string-operation audit.

Supporting C functions were inspected in string.c, string_build_impl.h, string_decode_impl.h, string_trim_impl.h, string_walk_impl.h, string_builder_impl.h, string_slice_impl.h, string_repeat_impl.h, string_search_impl.h, string_share.c, string_append.c, string_index.c. Their static support includes allocate, sequence, decode, unit_at, is_space, units_next, units_start, builder_add, builder_unit, builder_finish_units, builder_finish, clamp_index, ascii_character, repeat_unchecked, halves_pairs, to_units, unit_index_of, affix, grown, width, build, usable.

This is a conservative source inventory, not measured C function coverage. The explicit complete function inventory was assembled after the mutation menu, a procedural limitation against the requested ordering. The menu was committed to evidence before any mutant execution.

Matrix scope is the ten assigned rows. Direct C API occurrences identify their callers; coverage confirms their Go Build path. Other package rows can also call the same runtime through generated C, so their kills are unknown. No package-wide or repo-wide uniqueness is claimed.

The ten top-level rows have distinct assertions. TestStringIndexMatchesNode groups its four checker inputs internally, so it remains one row. No cross-top-level family, witness, or subprocess helper was identified.

Production menu (fixed before failures): change SHARE_FRACTION 8 to 7; append cached units +1 to +2; indexed low-half constant 0xdc00 to 0xdc01; slice NaN default 0 to 1.

The selector calls getenv in runtime hot paths. Its runtime cost is not a clean test cost. Each test was timed three times before instrumentation. C runtime sources are part of the runtime cache key; all switched bytes remain identical across selections and ADAMIC_MUTANT is evaluated in the native process. ADAMIC_BUILD_CACHE_DIR was separate per selection.
