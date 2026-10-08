# Step 08 React landing

Construction, clone and replay use the shared concrete index arena; CFG indices
are merged from lint-helpers/ecmascript-control-flow-graph-rules 49a62b32, including
stage1-arena/index ed236663. All classes remain in ../arena/arena_index.a with
private minting, checked reads and undefined absent edges. SSA is imported.
No pass implementations, planning files or evidence archives are included.
The generic symbol-query bridge required by resident bindings is included.
For-header roles are derived inside HIR; the parser is unchanged.

Required certificate: native = Node on all 1,442 non-Flow originals plus 72
probes; fresh checkpoint replay covers all 1,465 originals including 23 Flow
graphs. Static-components requires its 22 upstream cases and three owned
witnesses, with creation mutant caught. Full package certificates are run with
-count=1 -timeout=3h, every compiler/stage1/JSX/profile input set enabled and
all HIR semantic mutants. New top-level Go tests call t.Parallel.

Recorded compiler workarounds: gap 1 stores optional boolean presence/value in
a required-field record. Gap 2 copies clone records explicitly instead of
spreading in a program with private-index static constructors. Sites name the
gap; testdata/optional-boolean-gap.a and static-constructor-spread-gap.a retain
the shortest proving programs. Selector-style tests require Node output and
exact lower.NotYet.What (respectively `a field of type boolean | undefined` and
`spreading in a program with static constructor objects`) and fail when native
lowering succeeds, triggering removal of the workaround.

Validation: full HIR and arena packages pass; 79/79 lowering mutants, clone
alias, cache, corrupt-index and brand checks pass. Both compiler-gap tests pass.
Lint ran with every input set and -timeout=3h. Its 97/97 registered mutants,
static-components 22+3 certificate, 366-file JSX, 77-file throughput, shards
and profiles pass. The initial full run found a dirty-tree precondition and
consistent-return scanning detached rows; after committing and restricting that
rule to attached scopes, all three affected checks pass: 7,223 upstream
combinations, 950 compiler/stage1 files and detached-row isolation. The CFG
rule's mutant also passes on Node and sanitized native. Go cohere is unchanged.

Allocations: the byte-identical native/Node 12-function probe allocates and frees
4,087 objects: +875 from the original 3,212, -77 from the previous 4,164.
Retains/releases 16,325/14,909; peak live 231; regions 0. Earlier growth includes
boxed concrete indices, handle arrays, SSA callbacks and pattern arrays; this
whole-port delta does not isolate any one cause. The brands remain enforced.
