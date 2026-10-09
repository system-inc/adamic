# Integration and hooks

Claim 2f8c5f3 was pushed first with only claim.md. It starts at origin/wasm/integrate
6f7dce3. On the user's base update that history was retained: local merge fba6fd2
integrates origin/codex/concurrency ced5530; local merge b9052ec integrates
origin/area/runtime bb48dc9. No concurrency source branch was edited.

Fetched tips: concurrency ced5530, scaling 0f67284, async 55d2c89, moves 8aa7b62,
compiler 2e284e1. Logs against origin/main and newest reports were inspected.
Compiler is an ancestor of all four runtime branches. Concurrency is an ancestor
of async and moves; scaling forks earlier at 137afa5. No branch contains all others.
The integrated pool/compiler baseline is concurrency, whose integration report
records the accepted fixture/native gate and distinguishes async/moves as later
work. This selection excludes sibling scaling, async and move prototypes.

Hooks relative to b9052ec:
- parallel.c:9, guard the signal include; available_threads:54 returns one on
  WASI regardless of ADAMIC_THREADS; start:222 initializes one executor and
  returns before CPU discovery, signals, attributes or pthread_create.
  The existing sequential map body and its exception/ownership behavior stay intact.
- stack.c:70, thread startup returns on WASI, retaining the constructor's
  linear-memory stack limit instead of the native 8 MiB worker calculation.
- adamic.c:199, skip the competing-thread panic wait/pause on WASI.
- No changes to adamic.h, count.c, count.h, exceptions.c, heap.c, object.c,
  string_index.c or weak.c: clang's -mno-atomics lowers atomics and TLS plainly.

Two additional out-of-territory blockers were found and scope was requested:
- native.go:88 unconditionally adds -pthread, rejected with -mno-atomics.
  Needed: add -pthread only outside the wasm32-wasi target.
- share.c:11 hashes uintptr_t with shifts by 33. Needed: WASI-only uint64_t
  widening at the top of seen_slot, as weak.c already does.
