# watch.ts at zero under the public-owner sanction

The input is this branch's pushed handoff batch on integration 86cc9b7, with
all adaptations present, including 47 and 70. The unchanged latent run 0
feature compiler checks all diagnostic codes in 79 roots (including hostErrors).
Watch is **2 -> 0**, whole-tree findings **321 -> 314**, and whole original files
**53 -> 54 of 78**. No finding is added elsewhere. The three preceding handoff
closures remain zero: parser.ts, factory/emitNode.ts and transformers/utilities.ts.

The existing returned host objects explicitly store undefined via maybeBind,
whose body is fn?.bind(obj). The owning host declarations therefore need truthful
undefined unions. public-host-sites.json lists exactly 15 declarations: 11
public and four internal. Method type extraction retains the original signature
and its bivariant method parameters while allowing an explicitly absent value.
BuilderProgramHost's function property keeps its original parameter variance.
The receiving cache and directory host owners are widened too, preventing
errors from shifting into previously completed files.

Widening ModuleResolutionHost exposes 15 calls to directoryProbablyExists.
Its parameter's optional property excludes explicit undefined, although its
body already handles absence with !host.directoryExists ||
host.directoryExists(directoryName). Utilities belongs to 30 and is unchanged.
Each existing call receives an erased local function view describing that same
implemented contract. Callee and argument reads stay at the original evaluation
points. Parsed stock-runtime hashes guard both returned-host producers, maybeBind,
the utility helper and the Buffer reader. Exactly six owned files change;
watch.ts itself and every other partition's source remain byte-identical.

The adapter recognizes only adaptation 47's seven already-reviewed Buffer
assertions on full-tree reruns, inserting none. This compatibility landed with
the handoff batch and remains fully ledgered. All plans precede any source write. The handoff owner guards also explicitly
preserve the existing readonly modifiers; removing AutoGenerateInfo.prefix
readonly is rejected by a separate guard mutant.

public-api.cjs reconstructs the complete snapshot from pristine parsed owners:
20's 189, 40's exactly 28, 70's proven public readonly view, the two already
sanctioned handoff unions, and exactly these 11 public host owners. Stock 6.0.3
independently emits the replacement signatures; upstream's pinned dprint
configuration formats the result. Host method representation adds exactly 40
physical API lines. All 60,930 other reference baselines remain pristine.

Default oracle: **106,367 passing, zero failing, empty baseline diff** after the
exact API proof. Stock JavaScript bytes match untouched input for all 26 files;
CRLF and a zero-edit second run pass. Mutants remove the required owner (one
watch TS2375 returns), replace a required ! by ?? 0 (emission and ledger fail),
change the returned host producer (runtime hash fails before any write), remove
one helper view (occurrence/type guard fails), and add an unrelated API declaration
(exact reconstruction fails). All mutants are run and archived.

Tracing's catch already uses adaptation 47 errorMessage(e), retaining its
original || e fallback. Its four remaining TS2591 findings are fs, require and
two process.pid Node host bindings at lines 41, 64, 83 and 84. The current compiler
lacks those bindings; no source shim or invented linkage is introduced. This
file is **not zero**. PerformanceCore's two require bindings stay skipped as
directed. Sourcemap retains its one owner dependency on partition 33; nodeFactory
retains its two unproved generic writes in partition 30. Exact remainders are
in after.json. No compiler, native, or outside-owner file is edited.
