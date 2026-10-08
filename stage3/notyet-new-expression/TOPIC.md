Built topic-only class constructor support and dense Array constructors on current main; sparse construction stays NotYet.
Base origin/main 6998ebc24ae353193cb1495d3d51308131a4b5c7; final SHA accompanies the single push to codex/notyet-new-expression-topic.
Own fixtures passed 1.514s; counts passed 58.431s; final touched-package output and mutant summaries are attached in evidence.
Four Array mutants plus the ten retained class/cache mutants were rerun on this base; exact catchers are listed below.
Morning census: Identifier 22 roots, Parenthesized 8; 0 certified lowered roots, four explicit hole stops advanced, no views dependency used.

## Topic provenance

Fresh branch from fetched origin/main 6998ebc2. No merge commits above this base and no unlanded worker branch. Cherry-picked only our surviving non-merge commits: 39cfcae6 -> 2dc62fce, 974c2bc2 -> 33f19823, 6ea88f0d -> 6f1cc007, aae36609 -> 7917b8c4, 968172dc -> 55b4be12, a625251c -> 0d7f1ba4. Historical reports/evidence carried with them describe their original base, not this branch. Counts conflicts were resolved by retaining main's rows, then regenerating counts. The withdrawn b50442f2 Uint16 implementation was omitted with its reversal; only our final refusal fixtures and named NotYet were ported. No Int32-backed Uint16 helper exists here. No c41 or views branch was imported; any non-null behavior present comes solely from main. This is the conservative final-state interpretation of carrying our own non-merge work.

## Remaining morning kinds, largest first

Morning input is codex/stage3-notyet-table e8c283b5ed32477805357b652a170b85a04b2469, stage3/notyet-table/rerun-0730/after/roots.csv. It has 30 roots for the assigned newExpression kinds; PropertyAccessExpression is absent. The separate new Map from pairs stop originates in collections.go, outside this function's assigned kinds.

| Kind | Roots | Result |
| --- | ---: | --- |
| new an Identifier | 22 | Dense Array forms built and proven; 9 original sparse-Array sites remain unsupported, 6 weak collections need runtime lifetime support, 3 ordinary function allocators need receiver semantics/ruling, 1 class-expression binding needs registration, 3 Date sites need native Date representation/runtime |
| new a ParenthesizedExpression | 8 | Existing checked class caches held again to Node; original caches are ordinary functions, still blocked. NodeLinks uses any and needs a ruling. |
| new a PropertyAccessExpression | 0 | No remaining morning root; no cancellation/clearance inferred from its absence. |

Four numeric Array sites now reach the named NotYet an Array constructor creating holes (checker:14408,24218,33116 and transformer:249). Four untyped constructions reach an array of any (core:1325,2199,2200 and watch:135); this is another named stop, not acceptance. checker:27143 is masked by a number-as-condition refusal; tsbuildPublic:206 Date is masked by a value-as-condition refusal. The other twenty exact signatures reproduce. Full guarded replay results and exact commands for all thirty sites are in evidence/morning-topic-roots.json. Every census measurement is on a checker-rejected program. No original root is certified lowered, and no masked root is called an echo.

Checked views were not needed by the implemented forms. Sparse arrays require a hole/presence-aware array implementation, including callback skipping and key enumeration; checked views alone cannot supply that. WeakMap/WeakSet need runtime-owned lifetime/ephemeron semantics, Date needs native time storage and its host methods, SymbolLinks needs class-expression registration, and ordinary allocator functions need a sound this/new/return-replacement model. Those dependencies were named and skipped, never merged or substituted with incorrect behavior.

## Implementation and checks

newDenseArray uses existing ArrayLiteral IR. Empty constructors produce empty growable arrays; multiple arguments are elements; a sole nonnumeric argument is one element. Expression order and exactly-once evaluation come from the existing literal emitters, proven with side effects. Both Number and MaybeNumber single arguments stay NotYet: nullable numeric storage can still contain a number, and that number creates holes. Spread constructors and unsupported element representations stay NotYet. There are no IR, backend, runtime or other-worker lowering-function edits.

The own .a fixtures cover empty, numeric/multiple, string/single, string/multiple and object/multiple forms. The sparse fixture is held to Node output 3,0,",," proving callbacks skip holes. A nullable-length fixture is held to Node output 2,"," and pins the same refusal. The previous Uint16 70000 conversion fixture still pins NotYet with Node output 4464. All emitted positive fixtures run native, release, JavaScript and sanitizers through the existing oracle.

Commands, all redirected to log files:

- go test ./internal/oracle -run 'TestNewExpression|TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_' -count=1 -timeout 10m: passed 1.514s.
- python3 stage3/notyet-new-expression/run-dense-array-mutants.py: reversed elements and dropped singleton fail Node stdout; accepted numeric/nullable numeric lengths each fail their exact pre-emission NotYet assertion with nil. No build-error kill counted.
- python3 stage3/notyet-new-expression/run-class-value-mutants.py; run-cache-or-mutant.py; run-cache-local-mutants.py: retained checks rerun with scratch overlays.
- go test ./internal/lower ./internal/oracle -count=1 -timeout 30m: passed lower 66.209s and oracle 251.808s; output in evidence/topic-packages.log.txt.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts: passed 58.431s; only four own rows added, no main rows altered. Dense Array row allocations/frees 21/21, retains/releases 12/33, peak 8, regions 0.
- Scratch-only 9a1f14c5 replay tooling generated the guarded overlay over this current production tree. Production Lower/Load and all output paths remain disabled in the measurement binary. No replay-tooling branch or production overlay was committed.
- git diff --check and no-merge ancestry check pass. No whole-repository gate run. Setup/environment reused from the preceding unit: 231.369s, nproc 5.

All retained mutant catchers: eager ??, forced static allocator, constructor reread after arguments, repeated constructor and repeated argument fail Node stdout; no cache store fails allocator trap/Node exit comparison. Eager || fails stdout. Lexical eager initialization fails stdout, missing lexical store fails trap/exit comparison, and removed closed-initializer proof fails the exact additional-capture NotYet test before emission. Historical Uint16 implementation mutants were superseded by withdrawal and make no claim about supported Uint16.
