Restored lane 5 aggregate metadata for destructured checked fields.
Commit: see git log for Restore callable view metadata on destructured fields.
Tests: oracle TestV4Destructured 0.973s; lower ViewCallable|ViewArray 1.011s; native 5.360s; JavaScript 0.984s; TestCallTargetReaders 13.603s.
Mutants: bypassing the destructured adapter is caught in native and JavaScript; the native mutant matches Node and is ASan/UBSan/LSan clean.
Remaining: mixed aggregate literal boxing, higher-order producer checking, Set receiver integration and rank-165 readonly control.

The four lines restored from 3e1a7f6f attach ViewTypeID, ViewReceiverTypeID, ViewWhere and escaping callable metadata to destructured properties, preserving current initializeLocal handling. Reading a mismatched .ts callable remains successful; its first incompatible call checks producer arguments. The .a refusal pins its path and fix. New leaves: argument 0.96s, read-only 0.67s, refusal 0.27s. Inline temporary fixtures add no registered oracle rows; counts.md unchanged.

Already present: CallableMasks, DiscardContract, DiscardView and payload maps from 670a2939; boxed parameter/result adaptation from cb6adb25 with V4 call-time replacements and lane-5 misfit tests in 5acb88b0b; aggregate schemas and returned-allocation propagation from 3e1a7f6f; array representation and element payload hooks from 164dac2c. These are coverage observations, not a claim that every historical lane-5 fixture passes. The mixed aggregate object-literal hook from ebb1f8a6 is missing, as is that commit's doubled expected-type diagnostic capacity in runtime/object.c.

Fetched codex/views-callables-code to obtain f339ca8d and 2efc8949. The higher-order patch is absent; its old logical producer rejection occurs at read time and cannot be copied over the ruled adapter behavior. Both runtime callable-domain emitters currently reject ViewCallable children, and lower/view_callable_calls.go returns no domain for callable-valued arguments. Rank-165 readonly remains pending on that port.

The Set patch depends on 360084ada and 6f7902e05, also absent: there is no runtime/view_set_intrinsics.c, constructor-owned intrinsic Set identity, or checked Set method dispatcher on V4. Applying only 2efc8949 cannot supply its receiver certificate or demonstrate its wrong-domain mutant. The dependency changes must be ported into call-time dispatch rather than retaining their signature rejection at read time.

Runtime-callback blame follow-up: carry source call sites on callback-bearing IR (including array visits/sort and library callbacks), emit invocation-site context in both backends, and restore context across nested calls and exceptions. The existing pending test remains pending.

Setup: go build 44.416s; total 45.268s; nproc 5, cgroup quota 4 CPUs. Every test command used -count=1 -timeout 90s and an outer timeout 90.
