Built: directional scalar/boxed-union callable adapters, independently recorded union-member masks, collection callback hooks and exceptional cleanup.
Commits: follows 15b30747; this group is identified by git log on codex/views-callables.
Commands: focused IR/lower/native/JavaScript packages PASS (26.517/.634/5.621/.832s), vet PASS, counts PASS (9.188s), restored uncached callable oracle PASS (40.277s).
Mutants: scalar result/parameter value corruption, parameter cleanup omission, native/JS member omissions and native/JS reversed variance all failed executable tests.
Uncovered: aggregate transitivity, optional/default and overload variance, nominal parameter contracts, arbitrary intrinsic producers, and full compiler execution.

Native selects the immutable actual code slot, adapts borrowed arguments, calls
through the existing code(self, arguments, count) convention, releases newly made
parameter boxes even when adamic_thrown is set, and adapts owned results only on
normal return. Numbers acquire a box; booleans use immortal boxes; owned references
retain their ownership when becoming a union. Arguments and callable identities
are not reevaluated or replaced.

Result members must be a subset of the requested result; requested parameter
members must be a subset of the producer's parameter members. Two unions with the
same physical representation no longer count as compatible merely because both
are Union. Wrong shapes stop at the callable field read, before storing/calling.
Optional absent callbacks read undefined. Unknown producers at union conversion
boundaries panic in both backends, rather than interpreting the wrong word.

The original mixed-result probe now prints 5 on Node, native release, sanitized
native and JavaScript, with no leak. Additional fixtures cover booleans, named
forwarders, stored callbacks, filters including a real predicate parameter,
ordinary/reused map paths, forEach/reduce/sort, Map/Set visits, an undefined argument,
dynamic strings and throwing callbacks. Counts were measured for all boxing fixtures.
All seven mutants are restored and their raw logs are in logs/boxing.

Candidate counts are unchanged by these shape-level scaffolds: conservative
6/2818 pairs and 115/11063 reads certified; 2812/10948 remain. Static 4/308 pairs
and 34/1503 reads certified; 304/1469 remain. No new original tsc pair is claimed
until the corresponding original member fixture is certified. Exact reachability
remains unmeasured. Aggregate groups follow this adapter checkpoint.

A broader initial package selection also ran TestGraphClosureEnvironment. Its
untouched harness has code(self, arguments), while the integrated runtime already
requires code(self, arguments, count), causing clang's incompatible function pointer
diagnostic. git show 15b30747:internal/native/graph_regions_test.go confirms the
same two-argument harness in the parent. The final package selection includes all
callable and call-target guards and excludes this unrelated stale harness. Full
repository gate is not claimed. Debug.assert's guard remains pending the integrator's
predicate tip; it was not removed to obtain these passes.
