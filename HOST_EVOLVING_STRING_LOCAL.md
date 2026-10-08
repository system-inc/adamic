# Evolving string storage, 2026-10-08 UTC

Fixture 25 at 430:9 has an unannotated uninitialized `normalized`. Its writes are proven strings and subsequent checker flow types are strings or undefined. Extend the host branch's evolving object storage proof to string storage, supporting plain string writes, string += and ??=. Every write must agree on its representation; explicit any and any-valued reads remain unsupported. Initialize the local with undefined in its proven storage representation. The contextual types are unchanged.

The evolving_string_local.a probe agrees with Node in native under sanitizers and in the JavaScript backend: owned:true:tail, fallback:false:tail, seen:true, absent. Uncached oracle passed (0.806s). Focused lower controls passed (0.216s). Counted allocation/free 5/5, retain/release 9/18, peak 2, regions 0.

Executed mutant removes the check that every write has the same representation. TestEvolvingStringMixedWritesStayNotYet fails with got <nil>; exit 1. Log /tmp/evolving-string-mixed-mutant.log. The mutant was restored.

Unchanged pinned fixture 25 now stops at 470:73, lastIndexOf with a position argument (/tmp/evolving-string-host25.log). No complete fixture 25 pass, no new nullable reference kind reached. Full scratch gates have previously observed unrelated predicate-marker and virtual-call target failures; this is a focused gate. This fix belongs to the scratch branch because the base branch does not carry the host's evolving local analysis.
