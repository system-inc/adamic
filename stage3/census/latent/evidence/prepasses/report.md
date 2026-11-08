# Census prepasses, task sygfnmc

Tested code checkpoint. Base: origin/main `0816e1851`. Both guarded entry overlays
compile. Required Go packages, predicates family audit and all 15 pairing tests
pass. No census counts are claimed yet.

The prior absent-optional report and both literal/decodeJson prepass proofs at
`2c62ea4e` were read before implementation. They show an unprimed head misses
both writes, while its corrected driver reports them.

The guard follows Lower's local call graph because current main delegates through
LowerWithOptions, lowerChecked and lowerWithElementKey. It extracts calls between
moduleOrder and refuse without a fixed prepass list. Both isolated in-process
mutants reject the modified text, naming the method and template.

The four templates call collectViewedElements, noteInheritance, noteAccessorNames,
and noteOmittedOptionals in production order. The exemption table beside the guard
explains enumInitialization and namespaceInitialization. Independent census units
do not execute module initialization schedules. Census registration prepares enum
objects and actual uses still perform production enum validation and runtime
readiness checks. This does not claim a census of whole-program readiness errors.
collectViewedElements is called with the census's existing nil element key, which
retains conservative checking of every element read. It is an O(1) call in this
mode, not a measurement of production's nonnil-key discovery cost.

For entry censuses, original loader roots are saved before expanding the observation
ledger. Prepasses use ESMModuleOrder, the exact ordering implementation called by
moduleOrder, from those original roots. Directory drivers use the stable deduplicated
union of each original root's order. Type-only files and unrelated observation files
can therefore be excluded even when they are present in program.Files(). This matters
for all three walking fact collectors: inheritance facts, accessor names and omitted
optional fields can otherwise be contributed by modules production does not execute.
The order is irrelevant to their set unions, but membership matters. The readiness
passes are order-sensitive, which is another reason not to substitute Files() for
a production module schedule. The helper computes the list only; it does not move
moduleOrder's dynamic-import refusals or cyclic read-proof effects earlier in units.
Those remain at their actual production query sites in the measurement overlay.

Omission discovery reads contextual checker types with no explicit unit slots and
writes only omittedOptionals. The registered snapshot captures that program-only
map. latentUseProgramFacts installs an owned clone per unit, and statecopy's table,
capture/install generation and owned-map guard include its exact symbol-keyed type.
The historical predicates driver has a minimal latentUseProgramFacts helper for
the same omission map, with an owned clone in every fresh unit; it has no statement
snapshot or entriesRecords layer.

The cost-only overlay times eight fresh preparations per entry, including the two
exempt readiness passes on disposable lowerings. It never emits usable IR. These
are samples, not a claim that every possible independent unit was benchmarked.

Validation logs retain the cold statecopy publication timeout as well as the later
passing whole-package run. Both overlay builds exited 0. The full-history source
clone was stopped with its own exit 143 during indexing and replaced with a
successful exact pinned shallow fetch. No measurement used the partial clone.
