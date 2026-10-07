# Runtime mutable storage audit

The original snapshot table is historical. The integrated-runtime section below supersedes its unsafe classifications after concurrency-area integration. Unrelated compiler and external-checker behavior is outside this audit.

Sources: `origin/area/runtime` at `d8759cd66d04a940cba179c1990997c81fa5bb95` and `origin/codex/concurrency-stack` at `5a1963eca6e2b688b7bdaf4668a5c19989defc27`. Lines below refer to those immutable snapshots. A dash means absent. Header extern declarations are included alongside definitions. Keys use file/name/occurrence so moving a declaration does not require a line-only documentation edit.

The Go guard tokenizes every runtime `.c` and `.h`, including inactive preprocessor branches and function-local statics. It excludes genuinely const storage, functions, automatic locals and type declarations. A pointer to const data remains mutable unless the pointer itself is const. Repeated local names have source-order occurrence numbers.

| Audit key (file:name:occurrence) | Runtime line | Stack line | Holds | Writers and timing | Runtime classification | Stack classification |
|---|---:|---:|---|---|---|---|
| `adamic.c:broken:1` | 61 | 60 | Per-stream sticky write-failure flags | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | unsafe today: unguarded output state | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:output:1` | 49 | 48 | 64 KiB console byte buffer | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | unsafe today: unguarded output state | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:output_lock:1` | - | 61 | Mutex serializing ordinary console state access | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | Absent | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:output_mode:1` | 58 | 57 | Lazy stdout buffering mode (uninitialized, buffered, terminal) | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | unsafe today: unguarded output state | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:output_used:1` | 50 | 49 | Bytes currently used in the console buffer | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | unsafe today: unguarded output state | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:output_whole:1` | 55 | 54 | End of the last complete console line, read by the termination handler | write_line/output_flush/panic; finish at exit; stopped signal handler reads output/broken/whole | unsafe today: unguarded output state | Guarded by output_lock for ordinary access; unsafe today: finish and signal handler bypass the lock (requires review) |
| `adamic.c:panicking:1` | - | 62 | Atomic flag selecting the one fatal-message writer | adamic_panic test-and-set before writing fatal message | Absent | Atomic: atomic_flag |
| `adamic.h:adamic_box_false:1` | 660 | 705 | Immortal boolean box or built-in identity | Static initializer only; zero reference count prevents retain/release writes | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `adamic.h:adamic_box_true:1` | 659 | 704 | Immortal boolean box or built-in identity | Static initializer only; zero reference count prevents retain/release writes | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `adamic.h:adamic_literal_mark:1` | 130 | 157 | Address-only sentinel, never read as a character | Zero initialization only; ADAMIC_LITERAL_INDEX takes its address | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `adamic.h:adamic_stack_limit:1` | 835 | 880 | Stack overflow boundary | find_stack_limit constructor; stack_thread_start sets each worker boundary | Written only before pool starts: constructor; base has no worker initialization | Thread-local |
| `adamic.h:adamic_string_empty:1` | 514 | 527 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_string_false:1` | 516 | 529 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_string_true:1` | 515 | 528 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_string_undefined:1` | 521 | 534 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_thrown:1` | 700 | 745 | Pending exception object | Generated throw/catch cleanup and parallel execute_range set/clear it | unsafe today: process-global exception | Thread-local |
| `adamic.h:adamic_typeof_boolean:1` | 676 | 721 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_typeof_function:1` | 679 | 724 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_typeof_number:1` | 674 | 719 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_typeof_object:1` | 678 | 723 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_typeof_string:1` | 675 | 720 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `adamic.h:adamic_typeof_undefined:1` | 677 | 722 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `count.c:adamic_counted:1` | 8 | 8 | Allocation/free/retain/release/live/peak/region counters | count.c allocation hook and count.h macros on every counted heap operation | unsafe today: plain counters | Atomic: every member is _Atomic in count.h; peak uses CAS |
| `count.h:adamic_counted:1` | 28 | 29 | Allocation/free/retain/release/live/peak/region counters | count.c allocation hook and count.h macros on every counted heap operation | unsafe today: plain counters | Atomic: every member is _Atomic in count.h; peak uses CAS |
| `directory.c:denied:1` | 42 | 42 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:directory_type:1` | 31 | 31 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:error_kind:1` | 29 | 29 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:failed:1` | 43 | 43 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:file_type:1` | 30 | 30 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:missing_directory:1` | 40 | 40 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:missing_file:1` | 39 | 39 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:not_directory:1` | 41 | 41 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:ok_kind:1` | 28 | 28 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `directory.c:other_type:1` | 32 | 32 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `exceptions.c:adamic_thrown:1` | 10 | 10 | Pending exception object | Generated throw/catch cleanup and parallel execute_range set/clear it | unsafe today: process-global exception | Thread-local |
| `exceptions.c:error_name:1` | 15 | 15 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `exceptions.c:message_cache:1` | 27 | 27 | Cached shape and message-field slot for uncaught Error reporting | adamic_object_field/find fills on shape miss; uncaught or map iterator entry extraction calls it | unsafe today: process-global shape cache | Atomic: packed cache word uses __atomic_load_n/store_n in adamic.h and object.c |
| `exceptions.c:name_cache:1` | 27 | 27 | Cached shape and name-field slot for uncaught Error reporting | adamic_object_field/find fills on shape miss; uncaught or map iterator entry extraction calls it | unsafe today: process-global shape cache | Atomic: packed cache word uses __atomic_load_n/store_n in adamic.h and object.c |
| `heap.c:chunk_capacity:1` | 82 | - | Capacity of the growable allocator chunk registry | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Absent |
| `heap.c:chunk_count:1` | 81 | - | Number of registered allocator chunks | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Absent |
| `heap.c:chunks:1` | 80 | - | Growable registry of all allocated chunk pointers | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Absent |
| `heap.c:draining:1` | 233 | 275 | Whether this worker is already draining destruction | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:freeing:1` | 230 | 272 | Iterative destruction queue backing pointer | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:freeing_capacity:1` | 232 | 274 | Capacity of the iterative destruction queue | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:freeing_count:1` | 231 | 273 | Pending entries in the iterative destruction queue | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:giving:1` | 78 | 85 | Per-size-class heads of chunks with available allocation slots | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:owned_chunks:1` | - | 87 | Current worker allocator chunk list | new_chunk/take/give/list/release_last mutate during allocation and release | Absent | Thread-local; chunk registry moved to heap_parallel.h |
| `heap.c:spares:1` | 79 | 86 | Head of empty chunks available for another size class | new_chunk/take/give/list/release_last mutate during allocation and release | unsafe today: process-global allocator state | Thread-local; chunk registry moved to heap_parallel.h |
| `heap_parallel.h:chunk_count:1` | - | 12 | Number of registered allocator chunks | register_chunk/thread_number publish pages and increment IDs during allocation | Absent | Atomic: atomic page pointers, slots and fetch-add counters |
| `heap_parallel.h:chunk_pages:1` | - | 11 | Atomic page pointers for the stable allocator chunk registry | register_chunk/thread_number publish pages and increment IDs during allocation | Absent | Atomic: atomic page pointers, slots and fetch-add counters |
| `heap_parallel.h:owner_number:1` | - | 14 | Current worker allocator owner ID | thread_number initializes on first allocation | Absent | Thread-local |
| `heap_parallel.h:thread_count:1` | - | 13 | Atomic allocator owner-ID counter | register_chunk/thread_number publish pages and increment IDs during allocation | Absent | Atomic: atomic page pointers, slots and fetch-add counters |
| `ieee754.c:huge:1` | 1484 | 1484 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:huge:2` | 2253 | 2253 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:huge:3` | 2581 | 2581 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:pi_lo:1` | 1255 | 1255 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:tiny:1` | 1249 | 1249 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:two1023:1` | 1486 | 1486 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `ieee754.c:twom1000:1` | 1485 | 1485 | Volatile fdlibm floating-point constant (forces evaluation) | Static initializer only; math functions only read it | Written only before pool starts: C static initialization, no runtime writes | Written only before pool starts: C static initialization, no runtime writes |
| `input.c:argument_count:1` | 140 | 137 | Saved process argument count | adamic_arguments_save in adamic_start before generated module execution | Written only before pool starts: generated main calls adamic_start before any user parallelMap | Written only before pool starts: same generated-main ordering; C callers must not reinitialize during work |
| `input.c:argument_values:1` | 141 | 138 | Borrowed process argv pointer | adamic_arguments_save in adamic_start before generated module execution | Written only before pool starts: generated main calls adamic_start before any user parallelMap | Written only before pool starts: same generated-main ordering; C callers must not reinitialize during work |
| `input.c:denied:1` | 176 | 173 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:directory:1` | 177 | 174 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:error_kind:1` | 165 | 162 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:failed:1` | 178 | 175 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:missing_directory:1` | 175 | 172 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:missing_file:1` | 174 | 171 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:ok_kind:1` | 164 | 161 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:read_prefix:1` | 172 | 169 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `input.c:write_prefix:1` | 173 | 170 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `json_stringify.c:blanks:1` | 231 | 231 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `json_stringify.c:empty:1` | 230 | 230 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `library_language.c:identities:1` | 76 | 76 | Immortal boolean box or built-in identity | Static initializer only; zero reference count prevents retain/release writes | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `map.c:key_cache:1` | 345 | 351 | Cached shape and key-field slot for Map pair extraction | adamic_object_field/find fills on shape miss; uncaught or map iterator entry extraction calls it | unsafe today: process-global shape cache | Atomic: packed cache word uses __atomic_load_n/store_n in adamic.h and object.c |
| `map.c:value_cache:1` | 345 | 351 | Cached shape and value-field slot for Map pair extraction | adamic_object_field/find fills on shape miss; uncaught or map iterator entry extraction calls it | unsafe today: process-global shape cache | Atomic: packed cache word uses __atomic_load_n/store_n in adamic.h and object.c |
| `maybe.c:adamic_string_undefined:1` | 9 | 9 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `normalize.c:cached_classes:1` | 59 | - | Unicode combining-class lookup results | combining_class/mapping_of/composite replace entries during normalization | Thread-local | Absent |
| `normalize.c:cached_mappings:1` | 84 | - | Pointers to immutable Unicode decomposition mappings | combining_class/mapping_of/composite replace entries during normalization | Thread-local | Absent |
| `normalize.c:cached_pairs:1` | 112 | - | Unicode composition input/result cache entries | combining_class/mapping_of/composite replace entries during normalization | Thread-local | Absent |
| `normalize.c:cached_points:1` | 58 | - | Unicode point keys in a direct-mapped lookup cache | combining_class/mapping_of/composite replace entries during normalization | Thread-local | Absent |
| `normalize.c:cached_points:2` | 83 | - | Unicode point keys in a direct-mapped lookup cache | combining_class/mapping_of/composite replace entries during normalization | Thread-local | Absent |
| `parallel.c:changed:1` | - | 55 | Pool work/completion condition variable | claim/parallel_map/execute_range and shutdown under scheduler | Absent | Guarded by scheduler; pthread primitives initialized statically |
| `parallel.c:created:1` | - | 58 | Number of successfully created pool worker threads | start writes before pthread_once returns; shutdown frees/reset after all joins | Absent | Written only before pool work is published: start under pthread_once; workers deques guarded by scheduler; shutdown only at quiescent exit |
| `parallel.c:enabled:1` | - | 66 | This worker sanitizer perturbation enabled flag | worker_main sets index; tsan_pause initializes and increments budgets | Absent | Thread-local |
| `parallel.c:initialized:1` | - | 66 | Whether this worker read the sanitizer perturbation setting | worker_main sets index; tsan_pause initializes and increments budgets | Absent | Thread-local |
| `parallel.c:scheduler:1` | - | 54 | Pool scheduler mutex | claim/parallel_map/execute_range and shutdown under scheduler | Absent | Guarded by scheduler; pthread primitives initialized statically |
| `parallel.c:started:1` | - | 53 | pthread_once pool initialization token | pthread_once in parallel_map/threads/workers | Absent | Guarded by pthread_once(started, start) |
| `parallel.c:stopping:1` | - | 59 | Scheduler shutdown flag | claim/parallel_map/execute_range and shutdown under scheduler | Absent | Guarded by scheduler; pthread primitives initialized statically |
| `parallel.c:thread_count:1` | - | 57 | Configured pool thread count | start writes before pthread_once returns; shutdown frees/reset after all joins | Absent | Written only before pool work is published: start under pthread_once; workers deques guarded by scheduler; shutdown only at quiescent exit |
| `parallel.c:visits:1` | - | 67 | Per-worker sanitizer perturbation budgets | worker_main sets index; tsan_pause initializes and increments budgets | Absent | Thread-local |
| `parallel.c:worker_index:1` | - | 60 | Current pool worker index | worker_main sets index; tsan_pause initializes and increments budgets | Absent | Thread-local |
| `parallel.c:workers:1` | - | 56 | Pool worker array pointer; each worker also owns a scheduler-guarded deque | start writes before pthread_once returns; shutdown frees/reset after all joins | Absent | Written only before pool work is published: start under pthread_once; workers deques guarded by scheduler; shutdown only at quiescent exit |
| `regexp.c:regex_step_limit:1` | 8 | 8 | Regex execution step budget (zero means unlimited) | adamic_regex_set_step_limit writes; regex_run/test fast paths read | unsafe today: setter can race matching | unsafe today: same shared budget; fix with atomic load/store |
| `share.c:sharing:1` | - | 8 | Mutex serializing reachable-graph publication | adamic_share locks while walking and marking reachable graphs | Absent | Guarded by sharing; mutex initialized statically |
| `stack.c:__stack_low:1` | 22 | - | WASI linker stack boundary symbol | Linker defines address; runtime only takes its address | Written only before pool starts: linker symbol, no C writes | Absent |
| `stack.c:adamic_stack_limit:1` | 18 | 16 | Stack overflow boundary | find_stack_limit constructor; stack_thread_start sets each worker boundary | Written only before pool starts: constructor; base has no worker initialization | Thread-local |
| `string_build_impl.h:adamic_string_empty:1` | 130 | 115 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `string_build_impl.h:adamic_string_false:1` | 132 | 117 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `string_build_impl.h:adamic_string_true:1` | 131 | 116 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `string_index.c:adamic_literal_mark:1` | 108 | 136 | Address-only sentinel, never read as a character | Zero initialization only; ADAMIC_LITERAL_INDEX takes its address | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `tsgo.c:call_ns:1` | 14 | 14 | Accumulated external checker call time | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:facts_bytes:1` | 14 | 14 | Accumulated checker fact-buffer bytes | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:first_query_ns:1` | 12 | 12 | Duration recorded for the first observed checker query | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:input_ns:1` | 14 | 14 | Accumulated input preparation time | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:loaded_ns:1` | 12 | 12 | Accumulated checker creation time | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:output_ns:1` | 14 | 14 | Accumulated output conversion time | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:queried_ns:1` | 12 | 12 | Accumulated checker query time | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:query_count:1` | 13 | 13 | Number of measured checker queries | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `tsgo.c:run_started_ns:1` | 12 | 12 | Timestamp recorded at checker creation | tsgo_program/query_in/type_parts/inspect accumulate; tsgo_release reports and resets | unsafe today: process-global profiling state | unsafe today: same profiling state; fix with _Thread_local (per-thread reports) |
| `union.c:adamic_box_false:1` | 8 | 8 | Immortal boolean box or built-in identity | Static initializer only; zero reference count prevents retain/release writes | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `union.c:adamic_box_true:1` | 7 | 7 | Immortal boolean box or built-in identity | Static initializer only; zero reference count prevents retain/release writes | Written only before pool starts: C static initialization | Written only before pool starts: C static initialization |
| `union.c:adamic_typeof_boolean:1` | 12 | 12 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `union.c:adamic_typeof_function:1` | 15 | 15 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `union.c:adamic_typeof_number:1` | 10 | 10 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `union.c:adamic_typeof_object:1` | 14 | 14 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `union.c:adamic_typeof_string:1` | 11 | 11 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `union.c:adamic_typeof_undefined:1` | 13 | 13 | Immortal string and its units/index caches | Static initializer; string_index.c fills units/index lazily on the base | unsafe today: lazy string caches | Written only before pool starts: static initializer; stack string_index.c never mutates immortal literals |
| `weak.c:table:1` | 37 | 38 | Weak target-to-hidden-handle table backing pointer | weak_of inserts; weak_forget/weak_dropped remove; weak_target/weak_held read | unsafe today: unguarded shared table | Guarded by table_lock; mutex is initialized statically |
| `weak.c:table_capacity:1` | 38 | 39 | Weak table slot capacity | weak_of inserts; weak_forget/weak_dropped remove; weak_target/weak_held read | unsafe today: unguarded shared table | Guarded by table_lock; mutex is initialized statically |
| `weak.c:table_count:1` | 39 | 40 | Live weak table entry count | weak_of inserts; weak_forget/weak_dropped remove; weak_target/weak_held read | unsafe today: unguarded shared table | Atomic: table_count; table mutations also hold table_lock |
| `weak.c:table_lock:1` | - | 41 | Mutex serializing weak table and target access | weak_of inserts; weak_forget/weak_dropped remove; weak_target/weak_held read | Absent | Guarded by table_lock; mutex is initialized statically |
| `weak.c:table_used:1` | 40 | 42 | Weak table occupied/tombstone slot count | weak_of inserts; weak_forget/weak_dropped remove; weak_target/weak_held read | unsafe today: unguarded shared table | Guarded by table_lock; mutex is initialized statically |

Const tables checked by source search include `case_tables.h`, `normalize_tables.h`, `regexp_fold.h`, dtoa cached powers, regex shapes and collection iterator shapes. They have static initialization and no writers. Regex bytecode is generated at compile time; there is no runtime regex compilation cache in these snapshots. Number formatting uses stack-local buffers and bignums. Map hashing computes a hash per call; it has no global hash cache. JSON schemas are generated metadata; the runtime JSON encoder has no shape cache in either audited snapshot. Its two gap strings are listed above. Any later shape cache must acquire a row and protection.

State inside shared heap values also matters although it is not a C global: string units/index/view/cursor, Map iterator tallies, and reference counts. The stack branch publishes shared indexes with CAS, suppresses shared cursor writes, keeps immortal literal fields immutable, and uses atomic shared reference counts and iterator tallies. These need dedicated parallelMap fixtures after integration, alongside the global rows.

Smallest remaining work: integrate the packed atomic field caches; the regex budget and checker profiling atomic fixes are recorded below; integrate allocator/exception/count/shared-string protection from concurrency-area; reconcile normalization caches and target branches; review output exit/signal ordering. No unsafe classification may be treated as a concurrency landing pass.

Guard proof: `go test ./internal/native -run 'TestRuntimeStorageScanner|TestRuntimeStaticsAreListed' -count=1` passed in 0.205s (exit 0, `/tmp/runtime-statics-guard.log`). Adding a new file `runtime/statics_mutant.c` with `static int unlisted_runtime_state;` made `TestRuntimeStaticsAreListed` exit 1, naming that file at line 1 and the variable (`/tmp/runtime-statics-guard-mutant.log`). The file was removed and no mutant remains. This mutant never compiles C, so compiler warnings cannot kill it.

Setup: Go 1.27.1, clang 20.1.8 and Node 24.19.0 ready at 0s; submodules ready at 0s; `nproc` is 5. `bash cloud/setup.sh > /tmp/runtime-statics-setup.log 2>&1` exited 1 at cache warming because `cmd/adamic/tsgo.go` references `Options.Target`, `Options.Request`, `ValidateOptions`, `compileWASI` and `native.WASI` missing from the named base. Subsequent test commands source `/workspace/adamic-tools/env.sh`. Tests write complete output to logs. The full gate is currently blocked by missing native target APIs in `cmd/adamic/tsgo.go` on the named base; no compiler files outside this unit were changed to repair it.

Follow-up fixes on codex/runtime-statics: regex_step_limit and all nine checker profiling accumulators now use _Atomic storage. Implicit C11 loads, stores and increments are atomic, including report resets. This preserves process-wide accumulation rather than dropping worker measurements into separate thread-local reports. Their snapshot columns above describe the original branches, not these follow-up edits. Protection-removal TSan mutants still need to pass before these fixes are certified.

Correction from reading the packed-cache access paths: the stack snapshot's exception and Map field caches are already protected atomically, even though their declarations lack an _Atomic qualifier. All accesses go through packed-word __atomic operations in adamic.h/object.c. The base's two-field caches remain unsafe pending integration.

## Preliminary race proof, before integration

Observed: eleven fixtures (`regex`, `string_bmp`, `string_supplementary`, `normalization`, `numbers`, `map`, `heap_weak`, `field_cache`, `json`, `profiling`, `output`) each ran three times on four pool threads under TSan, without race reports. Every runtime operation is inside an `adamic_parallel_map` callback, after a four-worker rendezvous. The checker fixture uses the actual C wrapper against stateless external-checker ABI stubs; it proves the C profiling accumulators, not the external Go checker.

The tested source was an isolated copy of concurrency-stack, overlaid with this branch's `regexp.c`, `tsgo.c`, and area/runtime's `normalize.c`. It is not a concurrency-area merge. The default tests explicitly skip when `parallel.c` is absent, so a pre-merge green result is not evidence that these race fixtures ran.

Reproduce the exact overlay:

```sh
mkdir -p /tmp/runtime-statics-stack
git archive 5a1963eca6e2b688b7bdaf4668a5c19989defc27 internal/native/runtime | tar -x -C /tmp/runtime-statics-stack
cp internal/native/runtime/regexp.c internal/native/runtime/tsgo.c internal/native/runtime/normalize.c /tmp/runtime-statics-stack/internal/native/runtime/
source /workspace/adamic-tools/env.sh
go test ./internal/native -run '^TestRuntimeStaticsParallel$|^TestRuntimeStaticsProtectionMutants$' -runtime-statics-directory /tmp/runtime-statics-stack/internal/native/runtime -count=1 -v -timeout 30m > /tmp/runtime-statics-parallel-area-overlay.log 2>&1
```

Observed output: both tests passed; package exit 0 in 73.276s. Each protection mutant below compiled with `-Wall -Wextra -Werror -pedantic` and was caught by an explicit `WARNING: ThreadSanitizer: data race` in all three executions. A compiler error, timeout or crash without that report fails the mutant test.

| Mutant | Fixture | TSan finding |
|---|---|---|
| Remove regex budget `_Atomic` | regex | regexp.c, adamic_regex_set_step_limit |
| Remove checker accumulator `_Atomic` qualifiers | profiling | tsgo.c, adamic_tsgo_program |
| Replace packed cache atomic store with plain assignment | field_cache | adamic.h/object.c, adamic_object_field/adamic_slot_cache_store |
| Remove weak table lock and unlock | heap_weak | weak.c, insert |
| Remove allocator giving-list `_Thread_local` | heap_weak | heap.c, list_chunk |
| Increment allocation counter through a plain size_t pointer | heap_weak | count.c, adamic_count_allocation |
| Remove ordinary output lock and unlock | output | adamic.c, buffer/adamic_write_line |
| Remove combining-class cache `_Thread_local` | normalization | normalize.c, combining_class |
| Remove mapping cache `_Thread_local` | normalization | normalize.c, mapping_of |
| Remove composition cache `_Thread_local` | normalization | normalize.c, composite |
| Replace shared index CAS with plain pointer publication | string_bmp | string_index.c, shared_index |

The weak-table mutant initially failed compilation because its unused mutex triggered a warning. That attempt is **not** race evidence. Its corrected version preserves an address-only use of the mutex and removes both lock/unlock, and then produced the three race reports above. No runtime mutant edits the working tree: the runner copies a complete snapshot into a test temporary directory.

Other observed validation:

```sh
go test ./internal/native -count=1 -timeout 30m > /tmp/runtime-statics-native.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(regexp|json_stringify|normalize_coverage_cache|integer_format)' -count=1 -timeout 30m > /tmp/runtime-statics-oracle.log 2>&1
go vet ./internal/native > /tmp/runtime-statics-vet.log 2>&1
```

The native package passed in 166.841s; the selected uncached Node oracle passed in 24.631s; touched-package vet exited 0 without diagnostics. Final focused checks and the hardened guard mutant are in `/tmp/runtime-statics-base-final.log` and `/tmp/runtime-statics-guard-mutant-final.log`. The runtime statics guard's new-file mutant fails with file, line and variable; scanner tests also cover local statics, mutable pointers to const, function pointers, multi-declarators, aggregate variables, header prototypes and statics inside struct-returning functions.

Not covered: the actual concurrency-area integration (the branch is not yet published), complete repository gate, optional WASI race proofs, external Go checker concurrency, and output during a termination signal or exit. The output mutex proof covers ordinary concurrent writes only. Signal/exit safety is a review blocker, not established by that passing fixture. All unsafe base rows still require reconciliation with the integrated runtime before concurrency lands on main.

## Audited files

A newly added runtime C source or header requires review even when it introduces no mutable storage. The guard checks this file manifest as well as the variable table. Both original snapshots are covered.

- `runtime-file:adamic.c`
- `runtime-file:adamic.h`
- `runtime-file:array.c`
- `runtime-file:array_from.c`
- `runtime-file:bitwise.c`
- `runtime-file:case.c`
- `runtime-file:case_tables.h`
- `runtime-file:class_features.c`
- `runtime-file:class_inheritance.c`
- `runtime-file:class_static.c`
- `runtime-file:closure.c`
- `runtime-file:count.c`
- `runtime-file:count.h`
- `runtime-file:directory.c`
- `runtime-file:dtoa.c`
- `runtime-file:exceptions.c`
- `runtime-file:from_codes.c`
- `runtime-file:heap.c`
- `runtime-file:heap_parallel.h`
- `runtime-file:hypot.c`
- `runtime-file:ieee754.c`
- `runtime-file:input.c`
- `runtime-file:json_stringify.c`
- `runtime-file:json_stringify.h`
- `runtime-file:library_array.c`
- `runtime-file:library_language.c`
- `runtime-file:library_math_number.c`
- `runtime-file:library_object.c`
- `runtime-file:map.c`
- `runtime-file:map_set.c`
- `runtime-file:math.c`
- `runtime-file:maybe.c`
- `runtime-file:normalize.c`
- `runtime-file:normalize_tables.h`
- `runtime-file:number.c`
- `runtime-file:object.c`
- `runtime-file:parallel.c`
- `runtime-file:parallel.h`
- `runtime-file:parse.c`
- `runtime-file:radix.c`
- `runtime-file:regexp.c`
- `runtime-file:regexp.h`
- `runtime-file:regexp_fold.h`
- `runtime-file:region.c`
- `runtime-file:set.c`
- `runtime-file:share.c`
- `runtime-file:slab_quarantine.c`
- `runtime-file:slab_quarantine.h`
- `runtime-file:sort.c`
- `runtime-file:sort_undefined.c`
- `runtime-file:spread.c`
- `runtime-file:stack.c`
- `runtime-file:string.c`
- `runtime-file:string_append.c`
- `runtime-file:string_build_impl.h`
- `runtime-file:string_builder_impl.h`
- `runtime-file:string_decode_impl.h`
- `runtime-file:string_from.c`
- `runtime-file:string_index.c`
- `runtime-file:string_repeat_impl.h`
- `runtime-file:string_replace_impl.h`
- `runtime-file:string_search_impl.h`
- `runtime-file:string_share.c`
- `runtime-file:string_slice_impl.h`
- `runtime-file:string_split_impl.h`
- `runtime-file:string_trim_impl.h`
- `runtime-file:string_walk_impl.h`
- `runtime-file:tsan_test.h`
- `runtime-file:tsgo.c`
- `runtime-file:tsgo_runtime.h`
- `runtime-file:typed_array.c`
- `runtime-file:union.c`
- `runtime-file:utf8.c`
- `runtime-file:weak.c`

Final guard validation after adding the audited-file manifest: exit 0 in 0.215s. Two isolated mutants were run and restored: a new runtime file failed both the file-review and variable checks at line 1; adding the same unlisted static to the already audited count.c failed only the variable check at line 36. Logs are `/tmp/runtime-statics-guard-mutant-final.log` and `/tmp/runtime-statics-existing-file-mutant.log`. Final base checks passed in 7.180s, including the existing catastrophic-regex step-limit check; the two pool proof tests explicitly skipped because the base still has no parallel.c.

## Integrated runtime, signals and exit

Merged concurrency-area `b9479aa4beb65409312eaae5f406e94d3a6d7dfb` in `5e19080`. The historical stack classifications now apply, except normalization caches remain C TLS, regex/checker counters are atomic, and output is protected on every path as recorded below. Historical unsafe classifications describe the original snapshots, not the integrated runtime. `output_whole` was removed: volatile and signal fences cannot protect bytes that other threads write.

The installed handlers forward SIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGUSR2, SIGALRM, SIGXCPU, SIGVTALRM, SIGPROF, optional SIGIO/SIGPWR, and public realtime signals through stopped. SIGPIPE, SIGXFSZ and SIGUSR1 are ignored. Fault and abort handlers, including sanitizer handlers, are preserved. stopped may execute on any unblocked thread. It only reads an immutable nonblocking pipe descriptor, saves/restores thread-local errno, and performs async-signal-safe write of a signal byte. It never touches output, locks, counters or the heap. The normal signal thread inherits stop signals blocked, takes output_lock to flush complete lines, installs the default action, unblocks that signal on itself and raises it. Inherited ignored SIGINT/SIGHUP/SIGTERM are reset to stop handlers, matching Node; other inherited ignored stop signals remain ignored. A full pipe already contains a pending termination event.

stop_start initializes descriptors before installing handlers or publishing work. Descriptors remain open and unchanged until process death; OS cleanup deliberately avoids close/reuse races with handlers. stop_end sends zero and joins the signal thread at ordinary exit. Panic uses _exit, so it runs no exit hooks. SIGKILL/default fatal actions run no teardown.

| File, current line, storage | Handler or exit access / writers | Current safety |
|---|---|---|
| `adamic.c:stop_pipe:1`, 130 | Startup initializes; stopped reads write descriptor; stop_loop reads read descriptor; stop_end sends shutdown | Written only before pool starts and handler installation; never closed/changed. Handler uses async-signal-safe nonblocking write |
| `adamic.c:stop_thread:1`, 131 | Startup pthread_create writes identity; stop_end joins | Written only before pool starts; no handler access |
| `adamic.c:output:1`, 66 | buffer/write_line writes; signal-thread, finish and panic flush reads | Guarded by output_lock on every path; no handler access |
| `adamic.c:output_used:1`, 67 | buffer increments; all flush paths read/reset | Guarded by output_lock; no handler access |
| `adamic.c:output_mode:1`, 70 | First write_line sets mode and registers finish | Guarded by output_lock; exit/handler do not read |
| `adamic.c:broken:1`, 73 | Failed writes set; flush and finish read | Guarded by output_lock; finish snapshots failure before unlocking |
| `adamic.c:output_lock:1`, 74 | Ordinary writes, signal thread, finish, panic lock/unlock | Statically initialized mutex; never used by a handler |
| `adamic.c:panicking:1`, 75 | Panic selects the one fatal-message writer | Atomic flag; losing panic threads pause; winner locks output, reports then _exit |
| adamic.c const prefix/message | Panic/unreachable read constant message bytes | Static initialization, no writers, excluded from mutable inventory |
| parallel.c started 55, scheduler 56, changed 57, workers 58, thread_count 59, created 60, stopping 61, worker_index 62 | shutdown sets stopping under scheduler, broadcasts, joins workers, frees workers/resets created; workers read deques/stopping and TLS worker_index | pthread_once publishes startup, scheduler guards live deque/state; shutdown is an exit/quiescent operation. finish explicitly joins before flushing even if first worker output registered finish after pool hook. Repeated shutdown sees created zero. No handler access |
| heap.c giving 94, spares 95, owned_chunks 96, freeing 281, freeing_count 282, freeing_capacity 283, draining 284 | Worker heap_thread_end drains owned remote frees, changes its free lists and frees freeing buffer; heap_end cleans caller | C TLS; each worker cleans its own state. Signal thread allocates no Adamic values |
| heap_parallel.h chunk_pages 11, chunk_count 12, thread_count 13, owner_number 14 | heap_end reads registry and frees chunks/pages; thread cleanup reads owner identity | Atomic registry/counters; owner_number TLS. Constructor registers heap cleanup before main, so later pool join and output/signal hooks complete first. Remote queues are atomic chunk fields |
| count.c adamic_counted 8; count.h 29 | report_at_exit destructor and panic read allocations/frees/retains/releases/live/peak/regions; heap operations write | Every member atomic. Normal destructor follows atexit cleanup on supported clang targets. Panic snapshots may overlap tasks; snprintf is normal code only, never a handler |
| weak.c table 38, table_capacity 39, table_count 40, table_lock 41, table_used 42; exceptions.c name_cache/message_cache 27 | Generated final releases and heap release paths can remove weak entries/access error fields | table_lock guards weak state, count atomic; packed error caches atomic. Generated final releases precede main return; workers join before heap-wide disposal |
| tsgo.c loaded_ns/queried_ns/first_query_ns/run_started_ns 13, query_count 14, input_ns/call_ns/output_ns/facts_bytes 15 | Handle release can report/reset profiling during cleanup | All nine atomic; measurements are snapshots, not transactions; external checker concurrency outside C proof |
| async.c subscriptions/jobs/jobs_last; async_host_impl.h host_* below | async_run explicitly calls teardown before return; host_shutdown retires requests and closes pipe under host_mutex, then releases values | Host publication/shutdown under host_mutex; promises/reactions confined to loop. No signal-handler/stop_end access. Source async work in parallelMap is refused by internal/lower/parallel.go |
| ADAMIC_TARGET_WASI | WASI parallelMap, signals and teardown | n/a: codex/wasm-threads makes parallelMap sequential; available_threads returns 1 before overrides, start creates no workers, build has no threads; native signal loop excluded |

New async storage introduced by integration follows. Loop confined means the documented C loop-thread API contract, rather than C TLS. host_loop_thread checks affinity on host loop entry; source async work in parallelMap is refused. Foreign host workers may only publish copied buffers via mutex-protected resolve/reject/abandon, never manipulate Adamic values. Calling a loop-only C API concurrently is outside that contract.

| Audit key | Line | Holds / writers and timing | Classification |
|---|---:|---|---|
| `async.c:subscriptions:1` | 14 | Subscription registry; await/unregister/teardown on loop | Thread-local by loop confinement; async parallelMap refused |
| `async.c:jobs:1` | 15 | Reaction head; queue/drain_jobs on loop | Thread-local by loop confinement |
| `async.c:jobs_last:1` | 15 | Reaction tail; queue/drain_jobs on loop | Thread-local by loop confinement |
| `async_host_impl.h:host_mutex:1` | 23 | Queue/identity/pipe lock | Statically initialized named host_mutex |
| `async_host_impl.h:host_requests:1` | 24 | Live registry; new/process/shutdown mutate | Guarded by host_mutex |
| `async_host_impl.h:host_completions:1` | 24 | Completion head; publish/process/shutdown mutate | Guarded by host_mutex |
| `async_host_impl.h:host_completions_last:1` | 24 | Completion tail; publish/process/shutdown mutate | Guarded by host_mutex |
| `async_host_impl.h:host_next_identity:1` | 25 | Identity sequence; new increments | Guarded by host_mutex |
| `async_host_impl.h:host_started:1` | 26 | First request flag; loop new writes under mutex, loop process/shutdown read | Thread-local by loop confinement; writes also hold host_mutex |
| `async_host_impl.h:host_closed:1` | 26 | Shutdown flag; shutdown writes, new/hooks read | Guarded by host_mutex for registry; loop-only configuration reads |
| `async_host_impl.h:host_hooks_set:1` | 26 | One-time hook flag; startup loop sets | Thread-local by loop confinement |
| `async_host_impl.h:host_owner_set:1` | 26 | Affinity-established flag; first loop entry writes | Thread-local by documented loop affinity |
| `async_host_impl.h:host_owner:1` | 27 | Loop identity; first loop entry writes, later entries compare | Thread-local by documented loop affinity |
| `async_host_impl.h:host_pipe:1` | 29 | Host wake pipe; new creates, shutdown closes/resets | Guarded by host_mutex for create/wake/close; only loop drains/waits, so shutdown cannot overlap its own read |
| `async_host_impl.h:host_wake_hook:1` | 33 | Wake callback; startup loop configures; foreign publish calls | Written only before host work publication by startup-only API; mutex publication establishes visibility |
| `async_host_impl.h:host_wait_hook:1` | 34 | Wait callback; startup loop configures, async_run calls | Written only before host work publication; otherwise loop confined |


- `runtime-file:async.c`

- `runtime-file:async.h`

- `runtime-file:async_host_impl.h`

## Integrated observations

The previous “not covered” paragraph records d6d6ca4's pre-integration limits. Concurrency-area, signal output and exit output are now exercised; external Go checker concurrency and the complete repository gate remain outside the passing proof.

From internal/native, with `/workspace/adamic-tools/env.sh` sourced:

```sh
go test runtime_statics_test.go runtime_statics_parallel_test.go -run '^TestRuntimeStaticsParallel$|^TestRuntimeStaticsProtectionMutants$' -count=1 -v -timeout 30m > /tmp/runtime-statics-integrated-proof.log 2>&1
go test runtime_statics_test.go runtime_statics_parallel_test.go -run 'TestRuntimeStaticsSignalAndExit|TestRuntimeStaticsAreListed|TestRuntimeStorageScanner' -count=1 -v -timeout 10m > /tmp/runtime-statics-followup-final.log 2>&1
```

Observed: the 11 integrated fixtures and all 11 protection mutants passed in 186.136s, each fixture/mutant executed three times on four pool threads. The seven signal/exit scenarios cover SIGTERM/SIGINT/SIGHUP raised on the caller or a deliberately unblocked pool worker, and normal exit on the caller while pool tasks print. Signals retain their exact terminating signal; exit succeeds; no positive TSan race is accepted. Both new mutants compile and must produce explicit TSan data races in all three runs: an unlocked flush restored inside stopped races in flush/buffer, and restoring the unlocked exit flush without its preceding join races in flush. The initial weaker handler mutant merely read the buffer and survived one run; that attempt is not proof. The final mutant restores an actual unprotected flush and was caught.

The integrated guard passed; adding an unlisted static in a new isolated runtime file failed with file, line 1 and variable name in `/tmp/runtime-statics-followup-guard-mutant.log`. No working-tree mutant remains. Focused Go vet passed, and `git diff --check` passed.

Uncached Node output oracle from repository root:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestASignalLeavesWhatWasPrinted|TestClosedStdoutEndsAsOnNode|TestOneFileHoldsNodesOrder|TestAPromptComesBeforeTheRead' -count=1 -timeout 10m > /tmp/runtime-statics-signal-oracle.log 2>&1
```

Observed exit 0, package 44.498s. This includes signal behavior under native sanitizers compared to Node, output ordering, broken stdout and prompt flushing.

Setup rerun: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node 24.19.0 ready 0s, submodules ready 0s; nproc 5. It failed during test-cache warming because integrated `internal/native/map_hash_test.go:90` uses `Options{slabs: ...}` while the field is now `Slabs`. The old missing-target-API failure is resolved by integration. Rather than editing another unit's test, the commands above run the runtime proof/guard Go files directly; they compile and exercise the actual integrated C runtime. The complete native package and repository gate remain blocked by that unrelated merged test build error.

Final signal/exit, two new mutants, scanner and coverage checks passed together in 59.332s (`/tmp/runtime-statics-followup-final.log`). The two new mutants each produced three explicit TSan race summaries. The integrated 11-fixture proof was run after the signal-loop/exit-lock implementation; the final additional code change retries an interrupted handler pipe write and was covered by this final signal run.

## String view character pool

| Audit key | Holds / writers and timing | Classification |
|---|---|---|
| `string_slice_impl.h:bytes:1` | 128 ASCII bytes, C aggregate initializer only | Const storage; initialized before any pool starts, no runtime writers |
| `string_slice_impl.h:characters:1` | 128 one-unit immortal string headers, C aggregate initializer only | Written only before pool starts by static initialization. Counts are zero, units are known, and indexes carry the literal marker; retain/release and string caches never mutate them |

Whole slices retain their input without writing its units field, including immortal character headers. Shared input cache construction uses string_index.c's existing release/acquire CAS publication; new view metadata is filled before the view escapes. All owner count queries use adamic_reference_count and retention uses the shared atomic path. Region storage with no retainable count copies rather than becoming an immortal byte owner.

## Record runtime storage

- `runtime-file:record.c`: all static storage is const, initialized at compile time.
  The record/iteration field-name arrays have const pointers, the reference masks
  and shapes are const, and the missing-member and assignment diagnostics are
  const byte arrays. There are no lazy writes or runtime static caches. Mutable
  records and iterators are counted heap objects using the existing Map/Object
  ownership and concurrency rules.

## signals-area reconciliation

Base: codex/string-views-concurrency at 4b9f0c6; merge: area/runtime at 14504c7. Keep the threaded stop handler and output_lock, rather than restoring the area's signal-handler buffer access. Every handler still touches only the immutable write descriptor and thread-local errno. stop_start masks the full external stop list, including realtime signals, when creating the signal thread. The caller's mask is restored. Whole-line flushing needs no output_whole/fence: write_line holds output_lock through the terminating newline, and the signal thread waits for that lock before flushing. The signal thread holds output_lock through default-action re-raising, so another writer cannot start an incomplete line between the stop flush and termination. Signal handling never acquires the lock on the interrupted thread.

| Audit key | Current line | Holds / writers and timing | Classification |
|---|---:|---|---|
| `adamic.c:default_action:1` | 132 | Default signal disposition; adamic_start fills handler and mask, stop_loop reads after pthread_create publication | Written only before pool starts, before signal thread creation and handler installation; immutable afterward, no handler access |
| adamic.c stop_signals | 135 | Full ordinary external stop list, static const initializer; startup installs and masks these signals | Const, no runtime writers; public realtime limits queried at startup |

Closed descriptors 0/1/2 are reopened on /dev/null in descriptor order before arguments or signal pipes are initialized. Native startup ignores SIGXFSZ so file-size-limit failures follow the write-error path, and SIGUSR1 so the program continues as Node's inspector startup does. Regular files and terminals flush each line; pipes retain buffered output. write_all waits through nonblocking EAGAIN, detects zero writes, and retries EINTR. Panic replaces lone surrogates through write_text. All new POSIX startup, signals, poll and destination-stat behavior remains excluded under ADAMIC_TARGET_WASI; parallelMap there stays sequential.

The new TestRuntimeStopWholeLines observes Node first, requires two exact newline-terminated lines and termination by SIGUSR2, then checks native three times. Its mutant removes the signal thread's whole-line flush; it must still compile and terminate by the same signal, but must lose the two lines. No compiler failure, sanitizer report or timeout counts as catching this behavior mutant. The pool signal/exit TSan proof now raises every ordinary Linux stop signal plus realtime endpoints (34 and 64), on both the caller and a deliberately unblocked pool worker. Existing unsafe handler/exit flush mutants remain required.

Reconciliation setup: Go 1.27.1, clang 20.1.8 and Node 24.19.0 ready at 0s; WASI SDK 27 and submodules ready at 8s; build cache warm and setup done at 250s. nproc is 5, CPU quota 4. This integration resolves the historical map_hash_test field-build failure; the whole native package now compiles.

Counts were regenerated with `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts` (exit 0, 298.672s, `/tmp/signals-area-counts.log`). All 453 existing rows are unchanged. Five rows imported from stdout edges are added, yielding 458 rows:

| Fixture | Allocations / frees / retains / releases / peak / regions | Explanation |
|---|---|---|
| output_edges/fsize.a | 8 / 8 / 3 / 12 / 5 / 0 | Finite file-size-limit write/result fixture |
| output_edges/fsize_out.a | 60000 / 60000 / 0 / 60000 / 3 / 0 | Formats 60000 lines for stdout size-limit behavior |
| output_edges/closed.a | 0 / 0 / 0 / 0 / 0 / 0 | Writes immortal strings to standard descriptors |
| output_edges/panic_surrogate.a | 3 / 1 / 0 / 2 / 2 / 0 | Builds surrogate-bearing strings, then panic leaves live values through _exit |
| output_edges/usr1.a | 2 / 2 / 0 / 2 / 2 / 0 | Finite work/formatting fixture for continued execution after SIGUSR1 |

These additions increase table totals solely through new fixture work; there is no rise in any existing fixture. POSIX descriptors and the signal thread allocate no counted Adamic values.

`TestWASI` explicitly enabled with SDK clang passed in 105.015s (`/tmp/signals-area-wasi.log`): 52 strict C11 runtime units, 35/35 Node equivalences, and 100000 request rounds with zero live values and constant memory. This is a passing target test, not an opt-in skip.

Final focused mutant run passed in 38.102s (`/tmp/signals-area-final-mutants.log`). The handler mutant repeats the same unlocked flush before forwarding, to expose competing writes without adding a happens-before edge; a single flush in an early run occasionally terminated before a race observation and is not certified. Each of the three final executions emitted an explicit TSan data race in flush. Whole-line stop positive and drop-flush mutant each ran three times; the latter compiled and died by SIGUSR2 with empty stdout while Node printed both complete lines. The guard's new-file/unlisted-static mutant failed at line 1 (`/tmp/signals-area-guard-mutant.log`), and touched-package vet passed (`/tmp/signals-area-vet.log`). No working-tree mutant remains.

Final full native package: `go test ./internal/native -count=1 -timeout 30m` passed in 1407.283s (`/tmp/signals-area-native-final.log`). It includes all eleven cache/table fixtures and their mutants, every expanded signal/exit scenario, whole-line stop proof, and the static inventory guard. Full uncached native oracle: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m` passed in 944.127s (`/tmp/signals-area-oracle-native-final.log`), including the output-edge tests and regenerated-count verification. The dedicated output-edge run also passed in 34.025s (`/tmp/signals-area-output-edges.log`).

An initial combined native/WebAssembly oracle hit its 30-minute package timeout under simultaneous compilation and exposed a timing assumption in the Node signal probe: a 750ms delay could signal Node before its first line. The probe now waits for Node's actual newline and preserves consumed bytes before signaling; native buffered output retains the existing delay. Exact stdout and signal-status assertions are unchanged. Final native and WebAssembly oracle commands run separately. The initial weak handler mutant was replaced by the repeated unsafe flush described above; only final passing runs are evidence.

Final uncached WebAssembly oracle: `ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASI' -count=1 -v -timeout 30m` passed in 1237.955s (`/tmp/signals-area-oracle-wasi-final.log`). All Node comparisons (973.13s), byte/exit oracle and runner mutants (0.51s/0.48s), and every emission check (263.78s) passed. The gate reported zero cache hits and 463 Node misses. Final touched-package vet and staged whitespace checks passed. No full-repository `go test ./...` was run; this unit's complete native and oracle packages, opt-in target test, and complete WebAssembly oracle were run explicitly.

## Slab quarantine

slab_quarantine.c compiles its storage only in a sanitized build with ADAMIC_SLABS (address or thread sanitizer); a release build compiles the file to an empty object. slab_quarantine.h declares no storage. Each thread that frees a slab slot holds it in its own ring of 4096 (at most 1 MB) before heap.c gives it back, so no lock orders one thread's frees against another's. heap.c's adamic_heap_thread_end drains the calling thread's ring through give before draining its remote frees, and adamic_heap_end does the same for the caller after workers join. TestQuarantineIsPerThreadUnderTSan runs memory.c and the pool's map.c under TSan with ADAMIC_SLABS, and requires a race report from a mutant that makes the ring one shared static.

| Audit key | Current line | Holds / writers and timing | Classification |
|---|---:|---|---|
| `slab_quarantine.c:held:1` | 49 | The thread's ring of held slots, calloc'd on its first slab free; adamic_slab_quarantine writes entries, adamic_slab_quarantine_drain gives them back and frees the ring at thread end | Thread-local; only the owning thread reads or writes it, and the slots it names are on no free list or remote list until given back |
| `slab_quarantine.c:next:1` | 50 | Index of the oldest held slot and of the next entry; written by adamic_slab_quarantine, reset by adamic_slab_quarantine_drain | Thread-local |

## Typed arrays, October 7, 2026

typed_array.c adds no mutable static runtime storage. Uint16Array uses the existing
counted typed-array header and owner pattern, with a two-byte `uint16_t` buffer
allocated per owning array. A subarray retains its ultimate owner and points
into that buffer. The generic iterator retains the array. Existing heap freeing
releases owners, buffers and iterators; no new global cache or static counter is
introduced. This entry records this extension, not an inventory of earlier units.

