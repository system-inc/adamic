Defaulted object binding parameters lower in source order with one undefined-only fallback evaluation.
Commits: follows freshness fix 52882bdf on codex/generic-function-value; host next stop was 724:5.
Checks: lower package 15.181s; six uncached Node/native/JavaScript regression oracles 1.105s; vet clean; measured counts 12 allocations and frees, 22 retains, 27 releases, peak 3.
Mutants: extra fallback evaluation changes owned:1/owned:2 and calls 2 into owned:2/owned:5 and calls 5 in both backends; treating null as missing changes null into wrong in JavaScript alone. Oracle stdout comparisons catch both, with exit 0 and empty stderr.
Uncovered: defaulted closure binding patterns, defaulted array binding patterns and mixed array/scalar defaults retain their existing NotYet; full counts updater has the recorded regexp_tree.ts project-root blocker; no full repository gate.

Date: 2026-10-08 UTC. The object parameter receives the actual argument, selects
its default only for undefined, and destructures one temporary holder. Parameters
are processed in source order, so an initializer can read earlier bindings.
Ordinary scalar defaults share the same undefined-only flag. JavaScript emission
honors that flag with a single-evaluation expression; native uses its existing
presence path and nullable string sentinel helper. Null is retained by a nullable
string default. No new reference kind was migrated at this stop.

The oracle covers missing, explicit undefined and present arguments, a fallback
that allocates an object with an owned string and increments a counter, and
initializers before and after the binding pattern. Existing defaults, generic
function values and fresh/readonly array witnesses were rerun in both backends.

Evidence logs in /tmp: object-parameter-default-{first,backends3,lower,counts,
final-oracles,vet,mutant-double-fallback,mutant-null-default}.log. The first oracle
run intentionally exposed the JavaScript null-default bug; the final oracle is
green. Previous setup: total 30.366s, clang .186s, build 30.106s, cache 30.340s;
nproc 5.
