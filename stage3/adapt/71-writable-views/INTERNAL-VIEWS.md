Adapted eight shared-sentinel views and fourteen selected array-clear views, with four collateral clear removals.
Prior accepted source: f8dea48f; main source: ef3141e9; table input: d35a81d3; all public declarations unchanged.
Complete census 1047 -> 1021, cumulative 1174 -> 1021; exact table writable observations 637 -> 615, cumulative 755 -> 615.
Sentinel undo returns one row; clear undo returns eighteen; writer, runtime-copy, any and unknown mutants reject before writes.
Remaining requested scope: 82 shared views, 370 other views and 163 diagnostic views; full feasibility coverage remains incomplete.

Internal adaptations

sentinels.json selects six reader expressions: a TypeParameter spread in getOuterTypeParameters; a Declaration iteration in getResolvedMembersOrExportsOfSymbol; a Declaration search in markEntityNameOrEntityExpressionAsReference; a string spread in forEachFileNameOfModule; string iteration in getAllModulePathsWorker; and string mapping in updateSharedExtendedConfigFileWatcher. Each emptyArray fallback becomes a readonly element-array view. The existing nonempty array keeps its element domain; the shared never[] sentinel gains no element writer. No sentinel is widened to a mutable array, and no allocation or identity changes.

The internal getUnionSignatures and getUnionIndexInfos return types become readonly. Their local result builders remain mutable. All consumers receive readonly lists: structured-type construction, union-property/index resolution, and JSX signature selection. The public TypeChecker already returns readonly signature/index-info lists. find and arrayToMap only read their input container, never pass it to callbacks, and are included in the pinned runtime bodies. These eight annotation changes remove eight exact selected never[] sites: three string, two Declaration, one TypeParameter, one Signature and one IndexInfo. The trailing mutable-result || emptyArray expression in getUnionSignatures remains a census finding; this unit does not count a family as fully removed when only one return was adapted.

clear.json changes the internal clear parameter from unknown[] to { length: number }. Its entire body is array.length = 0. This is a genuine length writer, but it cannot insert a value of a wider element type. The receiver exposes exactly its written field, so callers no longer give it writable unknown elements. Fourteen selected views disappear: twelve Node[] calls and two BindingElement[] calls. Four Type[]/Symbol[] calls in symbolWalker disappear as collateral of the same receiver change; their reasons were absent from the exact writable table selection and receive no selected-site credit. No any or unknown is added; the old unknown annotation is removed.

Guards and measurements

Every owner body is pinned before edits. The existing whole-file erased-AST guard rejects runtime syntax changes, and the any/unknown guard rejects added broad types. These are type changes in internal helpers and expressions; the source-side emptyArray declaration, public contracts, and API sanction remain unchanged. Reapplying the adapter changes zero files. All 746 source files of the independently applied oracle tree match the census tree. All ten built JavaScript artifacts and typescript.d.ts match main byte for byte.

The same complete eight-body census probe measures all 79 compiler files with LATENT_ASSERT_NO_OUTPUT=1. Checker diagnostics remain exactly equal at 323 and eligibility remains 4489, with no added finding or skipped body. Before/after, main cumulative, source identity and emitted identity evidence is under evidence/internal-views/. Normalized identical streams use checked symlinks with a digest ledger.

family-mutant.cjs sentinels restores getOuterTypeParameters's mutable fallback: 1021 -> 1022, exactly checker.ts:13079:64 returns. family-mutant.cjs clear restores unknown[]: 1021 -> 1039, eighteen returning rows and no other change, including the four collateral symbolWalker rows. Both comparisons preserve diagnostic and eligible-body sets. Their output logs and comparisons are preserved.

family-guards.cjs plants emptyArray.push(undefined!) in getUnionSignatures and an indexed assignment in clear. Both exit 1 before source edits. The sentinel writer is caught by the already pinned enclosing createTypeChecker body. The three previous family writer mutants also pass their rejection assertions. plan-mutants.cjs plants a changed fresh-array holder, a slice allocation, any and unknown: all four exit 1 under their intended guard and preserve all source hashes. The first guard run expected the inner body name, but the enclosing body rejected first; that assertion was corrected and the full guard run repeated successfully.

Commands run with output redirected to logs:

- bash stage3/apply.sh /tmp/unit71-next-apply --write-table: exit 0; unit 71 is 23 files, 97 lines added and removed; aggregate 79 files, 5224 added, 5198 removed.
- UNIT71_TABLE_PROBE=1 UNIT71_TABLE_TREE=<tree> LATENT_ASSERT_NO_OUTPUT=1 /tmp/unit71-table-census2 <tree>/src/compiler <output.jsonl>: completed final and both undo measurements.
- measure.py and reconcile.py: complete-file checks, fresh counts and unchanged diagnostic/eligibility sets; the refreshed table has 615 retained writable observations.
- node family-guards.cjs, node plan-mutants.cjs, node adapt.cjs: rejection mutants pass; final application is idempotent.
- NODE_OPTIONS=--max-old-space-size=1536 taskset -c 0-3 bash stage3/lane/run.sh /tmp/unit71-next-lane2: runs the baseline stage3/oracle with all suites and four workers. Verdict PASS; apply exit 0, install/build exit 0, full oracle 106366 passing, one failing and zero pending, identical to main. The only baseline difference is the existing api/typescript.d.ts sanction (222 declarations), unchanged by this unit. Tests took 434.799 seconds; oracle 469.747 seconds; complete lane 604.228 seconds. The raw oracle exits 1 as on main. Execution metadata records the precommit HEAD f8dea48f; source-identity.json pins the exact uncommitted source tested, all 746 files, matching this continuation.

The first new lane and one undo census were interrupted by disk exhaustion. The build reported ENOSPC, not a type error, and the oracle tests did not start. Obsolete unit-owned scratch data was removed, the failed lane was discarded, and both measurements were repeated in fresh output paths. The failed build/report are preserved as disk-full-* evidence. The incomplete census is excluded.

Retained families

BLOCKED-DIAGNOSTICS.md lists all 158 location and five detached-location observations with file:line and required fields. file, start and length need protection; array views also need element-slot protection. The ledger describes the capability needed and distinguishes it from evidence of an actual undefined write in tsc. The public declaration and sanction stay fixed.

FAMILY-INVENTORY.md lists the original 90 shared and 384 other sites by source family, largest first, with adapted/retained status and original source. Confirmed public boundaries include mutable TypeChecker.getBaseTypes(), TypeChecker.getPropertiesOfType(), and InterfaceTypeWithDeclaredMembers's declared signature/index-info arrays. A blanket readonly propagation across those contracts would change the public API; a fresh copy would change runtime identity and allocation. The corresponding shared storage paths remain.

The 370 other retained views include FlowNode's mutable storage, ResolvedType's optional-member cache state, NodeBuilderContext's tracker aliases, generated-node metadata and generic node mutation. Truthiness expressions also produce apparent unrelated-type views that need attribution/narrowing review. This continuation has not completed those alias proofs and does not classify all retained observations as unavoidable API/runtime edits or genuine language questions. No supplied adaptation row is newly reclassified. The existing .a position and flag writer witnesses remain untouched. No new internal/oracle fixtures were added, so counts.md needs no refresh.

Toolchain setup was already completed for this unit with GOPROXY=https://proxy.golang.org|direct: Node .066s, Go .090s, clang .599s, Markdown 1.613s (install 1.425s), submodules 21.354s, warm cache 288.734s, total 288.872s. nproc remains 5; the cgroup quota is four CPUs, so the full oracle uses four workers. No whole-package Go test or full gate was run. No partial push occurred during this continuation.
