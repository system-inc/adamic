Built six candidate pairs with 40 source controls for finite caches, optional dictionary fields and numeric enum maps.
Commits: follows 45b334d0; minimal finiteDictionaryKeys, optionalDictionaryField and dictionaryReadContract hooks documented in the lane plan.
Checks: targeted lower/native/oracle gate passed (4.049s / 43.767s / 75.223s); vet passed; Node, sanitized C, release C and JavaScript controls passed.
Mutants: skip check, accept array as object and drop nested scalar check compile and lose the named refusal in C/JS; forced optional-receiver read fails the absence control in both.
Not covered: five candidate pairs / 48 reads remain; rich union selection, erased any and numeric string indexing remain uncertified; enumeration remains a separate obligation.

Working date remains October 11, 2026, 23:00 UTC. Static candidate inventory is
29 pairs / 228 reads. This batch adds six pairs / ten reads, bringing certification
to 24 pairs / 180 reads. Exact runtime reachability remains unmeasured.

The finite-key shapes are IndexedAccessType (three reads), IterationTypes (two),
IterableOrIteratorType (one), and optional Signature caches (one). Each has good,
wrong scalar, absent, array masquerading as object, and nested wrong-field cases.
The required IterationTypes selection refuses absence; optional caches allow it.
Original representative interfaces are used, rather than copied compiler sources.

CompilerOptions | undefined.paths adds one candidate read. Its eight controls
cover receiver absence, field absence, container kind, entry kind, nested array
elements and receiver evaluation exactly once. Four additional required-field
controls prove that an absent receiver is allowed while a present receiver with
an absent or explicitly undefined required field refuses. The declaration's type,
not the optional expression's widened result type, selects the field contract.

MapLike<WatchDirectoryFlags> adds two candidate reads. Eight controls exercise
fixed objects and genuine shared record producers, including absent keys, wrong
strings and numeric values outside the named enum members. Existing numeric enum
selection admits these numbers; no new union selection or record storage is added.

All wrong-value cases pin full exit-70 messages naming the read, expected type
and actual kind. Valid and absent cases agree with Node, including leak controls.
The three finite-key semantic mutants compile in both backends: skip check and
accept wrong shape print object instead of refusing; dropped transitive checking
prints the wrong leaf instead of refusing. The optional short-circuit mutant
executes the forbidden absent-receiver field read and exits 70 instead of Node's
successful missing result. Raw gate and mutant output are in logs/group8-*.log.

An initial misplaced finite-key hook failed Go compilation; it was moved into
elementAccess before validation. This diagnostic is not claimed as mutant evidence.
The targeted gate included dictionary, lazy-view, optional-read, readiness,
cast, record, object-refusal and enum-slot tests across ir/flow/fresh/lower/native/
javascript/oracle. ir/flow/fresh/javascript built with no matching tests. Full
repository gate was not run; previously documented unrelated failures remain.

Remaining candidate rows: any dynamic access (20), string numeric indexing (14),
Path numeric indexing (3), CompilerOptions rich dynamic selection (10), and
BuildOptions rich dynamic selection (1). The last two use the accepted lane 4/4b
selection handoffs. No pending row is removed or mislabeled as exact reachability.
Lookup, absence and enumeration remain this lane's responsibility.
