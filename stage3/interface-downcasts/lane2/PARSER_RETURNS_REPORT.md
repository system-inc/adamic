Built: nullable generic array returns and readonly never-array conditional consumption, with lazy scalar/literal element checks.
Commits: this checkpoint extends 4ed5e301 and the already merged Lane 1 b4cfd1aa (merge d060943c); the push reports its final SHA.
Commands: the focused lower/native/JavaScript/oracle gate passes; native debug, native release and JavaScript match Node on both unchanged parser probes.
Mutants: nullable dispatch, readonly consumer, mutable admission, string kind/literal checks and default string sort ordering caught; undefined-reference comparer stop proved by a mutant; prior array/callable and view/phantom dispatch mutants caught again.
Not covered: full parser slice rerun, 2,936 complete stock-source lowerings, complete native producer/mutation coverage, and full callable certification; Lane 1 owns the pending lazy-admission cast merge.

The probes are byte-identical to codex/stage3-parser-proof 05d1fe88. The nullable
sameMap return prints `2` then `undefined`; toSorted prints `0` then `1`, held to
Node in sanitized native, release native and JavaScript emission. The sameMap
fixture now retains both original assertion chains and the undefined input, with
no added early undefined guard. Its bound index assertions and omitted overloads
remain the adaptations already documented in PARSER_REPORT.md.

A syntactically immediate assertion chain creates no escaping intermediate
writable alias. `readonlyArrayConsumer` resolves its final contextual/return
slot. A nullable source can use that route only when the readonly consumer still
admits undefined; the runtime null pointer then passes through the assertions.
The outer view checks selected elements. Returning the same assertion to a
nonnullable result remains refused with the original castRepair diagnostic.

The toSorted target is SortedReadonlyArray<T>, whose phantom-array base is
ReadonlyArray<T>. Its conditional branches are judged against that readonly
consumer rather than the inferred mutable T[]. The never element has no value;
readonly covariance is therefore safe even for an unconstrained type parameter.
The writable element guard still runs before the bottom-type relation, so a
shared never[] reaching T[] remains refused. Both a direct mutable return and an
assertion chain ending in a mutable result pin the old diagnostic. For a mutable
result TypeScript should construct a fresh `[] as T[]`, rather than alias the
shared emptyArray through a wider writable type. An empty never-array allocation
uses numeric slots; no logical element certification is inferred from a cast.

The unchanged probes additionally require array template conversion and an
optional comparer. Array templates consume and check each selected element,
including closed scalar literals; an undefined array spells `undefined` without
dereferencing it. Sparse slots are selected through the existing holes accessor.
An explicitly undefined comparer uses UTF-16 string ordering, with numeric,
boolean and string controls including [10,9,1] and a supplementary character.
A supplied comparer uses existing closure dispatch. Sort with no comparer
argument remains refused. Recursive/object default conversion and undefined
reference elements with an optional comparer stop with a named NotYet diagnostic.

Small named integration hooks and IR additions:

- `nullableReadonlyArrayCast` joins the existing unified deferred-cast registry;
  `readonlyArrayViewBridge` and `viewArrayCast` use `readonlyArrayConsumer`.
- `impliedTarget` invokes `readonlyArrayConsumer` for conditional branch targets;
  `widened` treats a never element as bottom after writable slot checks.
- `elementType` invokes `viewNeverArrayElement`; `template` invokes
  `viewArrayString`; `arraySort` invokes `viewOptionalArrayComparator`.
- `ArrayJoin.Stringify` and `ArrayJoin.ViewRead` carry template conversion and
  element checks; readiness resolves this read metadata.
- `ArraySort.OptionalComparator` selects `emitViewOptionalArraySort` in the
  lane-owned native/JavaScript wrappers. `emitViewArrayString` emits conversion;
  native `viewArrayStringLiterals` preserves closed scalar refinements.
- Lane 1's `internArrayViewContract` is now the only registered array adapter;
  the older local init override was removed. Full physical metadata migration to
  `adamic_array.element_kind` is still the later native-array item.

Validation: `go test ./internal/lower ./internal/native ./internal/javascript
./internal/ir ./internal/oracle -run 'Test.*Cast|Test.*View|TestPhantom|Test.*Require|
Test.*Sort|Test.*ArrayJoin|Test.*Template' -count=1`, output in
parser-return-logs/final-focus.log. Lower 16.659s, native 11.430s, JavaScript
1.683s, oracle 56.315s; IR matched no tests. This is a focused gate, not a full
gate claim. The prior broad package gate's unrelated failures remain documented
in PARSER_REPORT.md; no new full gate was run.

Run `run-return-mutants.py`, `run-preflight-mutants.py` and
`run-integrated-mutants.py`; each mutation is restored in finally. Seven return-family mutants caught, two prior preflight mutants caught, and
three prior runtime/signature mutants caught. Individual
failure logs and aggregate outputs are in parser-return-logs. The environment
restart interrupted the first string-element mutant attempt; its source was
restored and the complete run repeated before the final gate.

Setup already completed in this session with GOPROXY=https://proxy.golang.org|direct:
Go 0.051s, Node 0.054s, submodules 0.135s, markdown 0.139s, clang 0.301s,
build 46.027s, warm 46.216s, done 46.269s; nproc 5, CPU quota 4. Lane 1's
ab4d6f90 merge attempt was aborted at the user's instruction, with no conflict
resolved locally. Merge its lazy-admission push only after Lane 1 resolves the
cast integration with this checkpoint included.
