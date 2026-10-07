Built: Walk, walkNodes and writeValueCss, one helper per .a file; 18 prerequisite removals across six rules.
Commits: implementation 27af0b0; reservation 19fc755 pushed before code; all twenty-three retained slot 04 helpers are tested and pushed.
Checks: touched package PASS 27.626s; uncached filtered input oracle PASS 1.459s; vet/format clean; session setup 165s, nproc 5.
Mutants: dropped root, broken stop propagation, skip descent, unknown-action descent, omitted separators, wrong closing delimiter; all compiled and were caught against Go on source Node, sanitized native and emitted JavaScript. Consumer omission caught.
Limits: no full repository gate or production integration; immutable finite non-nil tree arena and byte-builder view; parser geometry supplied by real Go; zero final blockers removed.

## Exact consumers

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.Walk

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.walkNodes

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.writeValueCss

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The frozen original readiness.json remains unchanged. This local ledger removes only twenty-three uniquely owned slot 04 helpers and assumes no integration of other workers' code. Original 46 helper-ready rules become 47 cumulatively, solely adding @next/next/no-img-element in the first batch. These are prerequisite counts, not completed rule implementations.

## Observations

Pinned Cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Temporary Go overlays invoke the actual Walk, private walkNodes and private writeValueCss. The adapter preserves CSS pointer identity by assigning one stable index per actual pointer, including shared children and repeated roots. Value data remains raw bytes, including malformed UTF-8. Actual Go ParseCSS and ParseValue supply parser geometry for source probes; parse errors yield the parser's returned CSS tree, commonly empty, rather than a claimed independent Adamic parse.

- 537 witness/tree controls produce 7518 matching lines on real Go, source Node, sanitized native and emitted JavaScript. They include empty input, all six CSS kinds, unknown/empty kinds with children, shared nodes, repeated roots, a forty-level tree with shared leaf siblings, nested functions, unknown value kinds with children, Unicode and 500 deterministic arbitrary-byte inputs.
- All 149 captured unique source fixtures cover every consuming rule. Source and actual Go-parser string/template literal values give 4074 matching observation lines. These direct helper probes do not claim full Adamic rule integration, and many TypeScript/class strings are not valid CSS trees.
- Actual upstream instrumentation records 682 helper calls, deduplicated to 21 tagged tree inputs: 3 Walk, 3 walkNodes and 15 writeValueCss. Those geometries are reconstituted with actual Go pointer identity and passed to all three actual helpers; 294 observation lines match. Recursively entered helper lists are captured too.
- Empty input, all 256 singleton bytes and all 65536 byte pairs are rendered through a controlled mixed CSS/value tree with all four container kinds, noncontainer children and an unknown value kind. All 921102 lines match. This exhausts raw writer byte content, not every possible tree topology.
- Each tree runs six visitor policies: continue, skip root, stop nested node, unknown root action, stop root, and skip even identities. Exact traces from Walk and walkNodes are independently compared, along with walkNodes' completion boolean. Traversal is depth-first and preserves aliases; skip suppresses a subtree, stop unwinds every enclosing level, and unknown actions proceed to siblings without descending.
- The writer starts with bytes 137,0,255 and runs twice into the same builder. Exact output proves append behavior, prefix preservation, recursive function parentheses, separator preservation, and omission of unknown kinds and their children. Inputs are not mutated.
- Vet produces no diagnostics, gofmt lists no files, and the filtered input oracle passes with zero cache hits and six probe misses.

## Every mutant

All six credited semantic mutants compile and execute successfully; each differs from real Go in all three Adamic modes.

- Walk drops the first root: first mismatch line 2, missing identities and their DFS descendants.
- walkNodes returns true on stop: line 5, wrong completion boolean plus visits after the nested stop.
- walkNodes descends on skip: line 3, pruned descendants incorrectly visited.
- walkNodes descends on unknown action 99 while preserving skip: line 7, unknown-action descendants incorrectly visited. This refined mutant isolates the unknown-action check rather than being caught first by skip.
- writeValueCss omits separator nodes: line 13, the two spaces and comma disappear.
- writeValueCss closes functions with ]: line 13, byte 93 replaces byte 41.
- Consumer omission removes every canonical-classes fixture row; the readiness-derived coverage assertion catches the missing rule.

The first local adapter used kind/children fields that structurally overlapped the JSON reader's OptionValue shape. The compiler refused that arena under its cycle-capable ownership proof. Renaming the flat projection fields to tag/edges/data separates those ownership roles and passes without weakening the types or modifying shared code. The final report credits only successful compiling semantic mutants, not the initial profile's compiler refusals.

## Commands, ownership and boundaries

README.md records exact reproduction commands. Test stdout/stderr is written directly to logs. evidence/tests.log contains the final full helper gate and all mutant outcomes; capture.log the counts; tailwind-capture-tailwind.log the upstream filtered fixture gate; oracle.log, vet.log and format.log the other checks. No shared registration generator, harness, Cohere worktree source or protected compiler file was edited.

Session setup previously passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5, cgroup cpu.max 400000 100000, memory 17.6GB. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. This continuation sources /workspace/adamic-tools/env.sh and verifies nproc 5; setup was not rerun.

Fetched and checked every claim path on all eighteen visible origin codex/lint-helpers* branches before reserving these highest-count unclaimed six-consumer ties. Higher-count comment helpers were already owned and delivered by the base branch. Post-reservation and final wildcard fetch/claim checks show no competing reservations for these three. No additional helper is reserved.

The full repository gate was not run, and unchanged previous slot packages were not rerun. The arena requires valid stable indices identifying non-nil nodes and finite immutable trees. Nil CSS pointers, invalid identities, mutation during a walk and cyclic value trees are outside the supported adapter boundary; actual captured consumer inputs have none. The byte builder models observable append output, not Go's internal copied-builder panic checks or nil builder pointers. Broader CSS/value parser ports, production AST integration and the ValueToCss entry wrapper remain separate dependencies. Capture selects rule fixture tests and excludes unrelated external live-population tests.
