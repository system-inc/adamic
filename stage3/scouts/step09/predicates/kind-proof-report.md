# Prove kind comparisons

Item 3 builds the largest measured candidate group: 283 direct kind comparisons. This is a body-proof implementation and focused certification, not a remeasurement claiming all 283 original compiler declarations now build. The pinned group table remains the observation of 577 unproven bodies on the measured main. Complete targets and unsupported members stay intact.

## Proven facts

Qualified enum constants enter the existing independent flow proof as known scalar literals. Enum identity comes from the pinned checker; the train's enum lowering already refuses non-finite members. For a closed discriminated union, the checker derives the selected complete source member and the independent verifier checks both return directions. An annotation never supplies an initial fact.

For an open source, the finite kind verifier starts with no helper summaries, partitions declared target and comparison tags plus the other-kind cell, and verifies every normal return. A const scalar alias of this parameter's kind, including a const alias of that alias, carries the same partition through already verified effect-free paths. Rebinding, assignment, unknown calls, mutable kind aliases and opaque expressions supply no proof. Object aliases are not scalar tag facts.

A readonly comparison property is not automatically pure. Enum members and readonly literal data initializers qualify; getters and externally declared properties do not. The getter negative mutates an aliased source object while evaluating the comparison constant and must not establish a kind claim.

Successful tag summaries retire their predicate call wrappers only after the complete target is admitted through the train's shared view builder. Remaining field guards stay active. Optional target fields remain in that view; their declarations are not removed. A required missing field still exits 70 at its read. The former optional-shape refusal test now checks that the extra field is retained, rather than calling admission a whole-data shape proof.

## Observations and mutants

Four own witnesses run unchanged on source Node, sanitized native, release native and JavaScript, in .a and explicit temporary .ts mode. enum selects both closed union members. alias checks the immutable scalar chain on an open source. optional retains its complete optional declaration and checks required reads for both present and absent optional data. fields preserves the remaining required-field guard: source Node prints undefined, while both compiled backends exit 70 with the pinned field-read diagnostic.

Each checked-source positive reports two proven predicate directions, zero checked directions and no predicate wrapper. Open targets retain the escapedText field view. Finished programs pass LeakSanitizer and counted allocation balance.

Six independent source-body mutants reverse the comparison or add an always-false conjunct for enum, alias and optional. Each .a body is refused with a predicate diagnostic. Three executable IR mutants return default false instead of the proven comparison; native and JavaScript disagree with source Node for enum, alias and optional. The enum witness initially failed as a native catcher because its stdout did not distinguish the selected branch. It now prefixes both branches, and the same mutant is caught in both backends. This is recorded as a witness correction, not a passed initial mutant.

A separate compiler mutant treats every readonly property as a pure constant. TestPredicateBodyProof/readonly_getter then fails: the body acquires TaggedView:true with no error instead of the required opaque-test refusal. The guard is restored before final checks.

## Pending

optional_read retains the original optional-read witness separately. Source Node prints absent:missing, present:value and other. Both .a and .ts stop at 'a checked field alias requiring an optional, accessor, or representation conversion'. Its oracle is explicitly skipped as pending, has no allocation row and is never claimed as a backend pass.

Full untagged negative membership, overlapping tag contracts where false narrows, unproved indirect or receiver predicates, and proof of masks or property-presence groups remain pending. Recursive or generic target members rejected by the complete view builder stay refused or NotYet. There is no whole adapted-tree acceptance claim.

Design questions still concern negative view membership without consuming or panicking inside the probe; receiver and indirect-call argument identity; full generic and recursive descriptors; optional field alias conversion; and complete producer domains for masks and effectful accessors. No decision or unlanded dependency is silently substituted for these gaps.

## Counts and checks

The global counts command adds four rows and changes no previous row. enum balances 5 allocations and frees, 4 retains and 6 releases, peak 3. alias balances 3 allocations and frees, 4 retains and 6 releases, peak 3. optional balances 4 allocations and frees, 7 retains and 9 releases, peak 4. fields stops holding 2 allocations, with 1 retain and release, peak 2; panic state is counted where it stops. No optional-read counts are invented.

Final focused lower command: go test ./internal/lower -run 'Predicate|TestConditionAssertionAdmission|TestEveryNeedsCallbackEffects' -count=1 -v; PASS 15.913s, /tmp/predicates-kind-lower-item3.log.

Final oracle command: go test ./internal/oracle -run '^TestCheckedPredicate(Oracle|CountsAreRecorded)$|^TestPredicateKind(ProofOracle|OptionalReadPending|CountsAreRecorded)$' -count=1 -v; PASS 11.266s with optional_read explicitly pending, /tmp/predicates-item3-oracle-final.log.

Counts command: go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts; PASS 48.827s, /tmp/predicates-kind-counts.log.

CLI command: go test ./cmd/adamic -run '^TestExplainChecksOutput$' -count=1; PASS 1.490s, /tmp/predicates-item3-cli.log.

No whole-package tests or full gate run. No code is copied from cohere. The implementation uses existing pinned checker shims and the train's checked views. The feature branch carries its own three items above the named train tip; no other worker branch is merged.
