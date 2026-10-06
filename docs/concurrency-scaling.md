# Concurrency scaling

Stacked on 137afa5. Claims: runtime/string_index.c and share.c for lazy shared string indexes; runtime/parallel.c for adaptive range sizes; runtime/adamic.h and object.c plus runtime slot-cache declarations and the one declaration hook in emit_objects.go for packed atomic caches. Parallel C harnesses and Go tests, measurement scripts and this report will hold the proofs and measurements. No lower or JavaScript changes.
