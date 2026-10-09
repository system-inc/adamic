# Step 09 predicates contract

This contract serves roadmap step 09, task #tzd3gjg under #kbcj5nr. It implements the ruling conveyed by @system_adamic: TypeScript predicates are checked in both narrowing directions; Adamic predicates require a body proof. Base is train 4, cloud/land-train-4-views-v3-787cea7a-6c28ada8 at 1e81b051c52eee40d47c44f4a5f68b821e606708. There are several live train-4 variants; this uses the explicitly named tip. No worker branch is merged.

## Measured body obligations

The source is codex/step09-double-casts at 8a7ab17e, with its pinned adapted-tree API inventory and exact main proof results. The inventory has 651 annotations, 580 bodies and 71 bodyless declarations. Main proves three bodies: core.ts:1769:42, core.ts:1773:39 and watchPublic.ts:739:77. The table partitions the remaining **577 bodies** once each. Bodyless declarations are excluded. All original bodies, annotation coordinates and exact diagnostics are retained in stage3/scouts/step09/predicates/evidence/groups.json.gz.

| First missing proof obligation | Bodies | Three annotation coordinates in the pinned adapted tree |
| --- | ---: | --- |
| direct kind comparison | 283 | `checker.ts:7873:81`; `checker.ts:44756:49`; `checker.ts:47505:62` |
| delegation or composition of predicate calls | 115 | `builder.ts:1192:54`; `checker.ts:13778:57`; `checker.ts:13783:67` |
| kind partitions with control flow or extra conditions | 105 | `checker.ts:9408:75`; `checker.ts:32636:54`; `checker.ts:35656:61` |
| property presence or structural reads | 40 | `builder.ts:268:87`; `builder.ts:1176:79`; `builder.ts:1181:58` |
| flags and bit masks | 24 | `checker.ts:4290:132`; `checker.ts:13387:43`; `checker.ts:14997:47` |
| other value, generic, assertion or erased claims | 10 | `checker.ts:26489:43`; `commandLineParser.ts:3038:37`; `debug.ts:213:142` |
| Total | 577 | Every unproven body appears once |

These are measured AST forms and proof obligations, not a claim that 577 predicates lie. Nor does a syntax group prove every member sound. The partition first selects direct parameter-kind comparisons and stable kind aliases. It then selects flags/masks, remaining kind conditions, calls to inventoried predicates, structural property reads, and the remaining value/generic cases. Overlapping bodies keep the first applicable group. The ten final cases retain subcategories: seven semantic/generic conditions, two constant/assertion/erased claims and one primitive test. Three examples of the combined group are shown rather than inventing examples for a one-entry subgroup. The classifier parses the original serialized bodies with TypeScript 6.0.3; it does not search comments or guess from function names alone.

The largest candidate group is **283 direct kind comparisons**. A SyntaxKind comparison proves a tag partition; it does not establish unrelated fields of an open Node interface. Lowering must retain the original complete target and any checked view for those fields. The existing c191f97b proof work is read as a dependency already represented on this train, not copied or merged from its branch. Train status and semantic results are measured separately from this main census. The original corpus had checker diagnostics; this table is not a successful full-corpus build.

## Checked predicates in .ts

At a resolved call of P with saved argument x, source type S and target T:

1. Evaluate the callee and arguments once, in JavaScript order. Execute the predicate once. Save its result and use the post-call value through the saved argument identity.
2. For a true result, check the narrowed type T once. Use its proven runtime tag, or the shared checked view for an untagged target. Additional target fields keep their original contracts. Returning true is not itself evidence.
3. For a false result, check x is not T where the false branch actually narrows the caller's value. Determine that from the checker's source domain and excluded part; an annotation is a claim to validate, never proof. Where the false branch leaves the type unchanged, insert no false-direction check. An empty branch still needs its check if its type narrows; the absence of a later field read is not body proof.
4. For asserts x is T, check the asserted type after every normal return. A throw has no normal return to check. The check belongs to the resolved call so it names that call site.
5. A failed direction exits 70. Its diagnostic names P, the call site's file:line:column, the true/false/assertion branch, S and T. Both backends carry the same check and diagnostic.
6. Every inserted predicate check is recorded in the checked-sites report. Record each call site and direction as checked, proven, or unobservable, with the source and target. An unobservable false direction is recorded separately from a body proof. A proof can retire a check; missing reification cannot.

The default applies to all six groups while their complete runtime contracts are representable. Erased any, unresolved generic T, recursive/unbuilt contracts, unsupported receiver predicates and opaque callable cases stay a named refusal or NotYet; they are pending, never pass. A tag test must not be substituted for an untagged shape check. V1/V2/V3 facilities count as available only if they are on the selected train and pass the fixture there.

## Predicates in .a

The body must establish its claim on every normal return path. Both truth directions must hold: true implies T and false excludes T wherever a false narrowing is claimed. An assertion's normal return establishes T. Unproved declarations, overload claims and callbacks stay refused. Imported .ts predicates retain their checked contract; the declaration's extension determines its proof mode. A .a caller does not erase those checks.

Rebinding the parameter, mutation through aliases, getters, unknown calls or delegated predicates cannot introduce trusted facts. A helper annotation supplies no proof. Proof summaries begin empty and are added only after independent body verification; recursive helper cycles without an independent seed stay unproved.

## Proof signatures in lowering

Let S be the original source domain, T the complete declared target, K a discriminant partition and E the body's established facts. These are logical signatures, not casts:

- kindProof(S, T, kind, K, returns): prove all true returns select exactly T's tag set and all false returns exclude it. For a closed discriminated union, selected source members must satisfy T in full. For an open base, prove only the tag partition and retain the shared checked view for remaining fields. A tag table comes from complete original declarations and verified source domains, never from believing P's annotation. Qualified enum constants and stable local aliases resolve through checker identities.
- delegatedProof(S, T, verifiedHelpers, returns): substitute only independently proved helper summaries, preserving the tested argument identity and both truth directions. Compose conjunction, disjunction and negation; prove every return path and invalidate facts across effects. This covers the 115 delegation candidates and helper portions of the 105 mixed-kind bodies only when their extra conditions also prove the full claim.
- flagsProof(S, T, mask, valueDomain, returns): prove the JavaScript ToInt32 bit test over a known finite tag/flag domain selects exactly T, including the complement. Overlapping masks, combined flags and broad number inputs require independent domain evidence. This is the obligation for the 24 mask candidates; no arbitrary nonzero mask implies an interface.
- presenceProof(S, T, property, readiness, valueType, returns): prove actual presence/readiness and the member type required by T. Distinguish missing, present undefined and truthy values. A property alone does not prove unrelated fields, nominal ancestry or producer invariants. This addresses the 40 structural candidates where their full target can be established.
- valueProof(S, T, primitiveOrNominalPartition, returns): use independently checkable typeof, literal, array or nominal facts; generic targets need concrete instantiation and complete descriptor evidence. Constant true, empty assertions and semantic subset filters require additional producer/domain facts and cannot be accepted on their annotations.

Item 3 builds the largest group's rule and its conservative boundaries. Proven predicate-direction checks retire; the checked view's independent field-read checks remain where the source domain does not statically establish every target field. There is no claim that merely proving a tag removes all structural checks.

## Design decisions and questions

The ruling settles true checks, false checks where narrowing occurs, assertions on normal return, exit 70, named diagnostics and counting. Remaining implementation questions are tracked conservatively:

- What stable reference and checker flow establish that a false call actually narrows, including negation, boolean aliases, continuation after return, dotted arguments and indexed arguments? Do not infer this solely from lexical then/else placement or a later read.
- Which full untagged contracts can the selected train test negatively without consuming a view or panicking inside a membership probe? Missing/unbuilt negative reification is pending.
- How do indirect calls, methods, this predicates, generic overloads and callback escapes retain a reifiable contract and a saved argument identity? Unsupported forms remain named pending cases.
- What effects invalidate an established tag and which getters can run during membership? Preserve the original evaluation count and order; no speculative extra getter call or copied mutable object.
- Which tag/flag domains have complete verified producers? A matching numeric tag on an arbitrary Base object does not prove its extra fields. Keep checked views or refuse that proof.
- Can optional targets and overlapping discriminants be proved without weakening original declarations? Keep all original fields and exact optional semantics; do not reduce interfaces to obtain a pass.

No answer is requested between items. Each implementation reports its concrete supported and pending forms and its evidence on the train.

## Evidence and checks

Reproduction: source /workspace/adamic-tools/env.sh, then run node stage3/scouts/step09/predicates/group.cjs with evidence/inventory.json.gz, evidence/main-results.json.gz and an output JSON path. The classifier asserts 651 distinct observations, 580 original bodies, 577 unproven bodies and the pinned parser version. Three independent measurement mutants fail: dropping an observation, dropping a body and inventing a proven result. Their logs and intended catchers are retained under evidence. This documentation item adds no oracle fixture and changes no counts row.

Setup first failed because the generated Go build cache occupied 25 GB and the disk was full. The failed train checkout was restored from its known Git blobs and retried after go clean -cache. A setup retry overlapped that branch transition and saw inconsistent package inputs; it is not a train result. Setup on the stable train succeeds: Node 0.022s, Go 0.022s, markdown 0.064s, submodules 0.073s, clang 0.189s, go build 41.530s, test binaries deferred 41.614s, cache warm 41.615s, done 41.642s. nproc is 5, cpu.max is 400000 100000; Go 1.27.1, clang 20.1.8, Node 24.19.0. Raw setup logs are retained compressed. No full gate or whole-package test is run by this item.

## Item 2 implementation

Named checked .ts calls and their direction report are built on the selected train. See stage3/scouts/step09/predicates/checked-report.md for source Node pins, independent removed-guard mutants, allocation rows and supported forms. Complete untagged membership, overlapping tag contracts on a narrowing false branch, and unproved indirect calls remain pending. Positive tagged results use the complete shared checked views already present on the train. .a bodies and overload claims still require independent body proof.

## Item 3 implementation

Independent qualified-enum and immutable kind-alias proofs are built for the largest candidate group. Checked predicate directions retire when the body proves the tag partition and the complete target's remaining field view is admitted. Optional declarations remain intact; an optional field alias read still stops on this train and is recorded as pending. See stage3/scouts/step09/predicates/kind-proof-report.md for body-liar refusals, executable default-return mutants, the readonly-getter purity mutant, exact commands and all four new counts rows. The original 577-body measurement is not relabelled as an adapted-tree pass.

## Corpus measurement after the build

The original 580 bodies are remeasured at 90a78f49 with all source hashes preserved: 309 logical proofs, 18 admitted proofs, 291 .a NotYet view stops and 271 .a Refused bodies. In .ts, 18 predicates construct 54 call checks; 495 predicates are pending on views. The complete group table, every checked call and the remaining unobserved/other cases are in [the corpus report](../stage3/scouts/step09/predicates/corpus-report.md). It preserves all 320 baseline checker diagnostics and distinguishes a proved condition-only Debug.assert body from its 445 calls still blocked by target reification. Current origin/main 031a1259 was merged into the feature branch; Git reports it is already included. Focused checks and counts pass with no new count row.

## Optional alias conversion dependency

Q5 uses compiler/optional-presence a774d316's independent per-slot presence ranks. Optional string aliases preserve missing, present undefined and present value; the pinned checker's non-missing field type governs present payloads. Scalar optional receivers retain one evaluation. Build-ahead fixtures and mutants pass on that dependency, which remains outside the selected train. Remeasurement of the exact 289 earlier alias stops admits zero complete views: all pass node?.kind, then 151 stop on optional symbol.declarations arrays and 138 on optional node.original objects. See [the optional conversion report](../stage3/scouts/step09/predicates/optional-report.md). Q2–Q4 and Q6 remain pending the views slices. Q1's condition-only assertion ruling is received and is built as a separate delivery.

## Condition-only assertions ruling and implementation

Q1 is settled: a condition assertion is proved only when every normal return establishes the argument's truthiness independently of runtime assertion-level switches. The checker retains its own condition narrowing. Otherwise .ts inserts a truthiness check immediately after the call using its already-evaluated argument, never a second evaluation; failure exits 70 naming the assertion and call site. .a still requires a body proof. Sites count as proven or checked in the shared report. All 445 pinned Debug.assert calls are now proven, zero checked and zero pending. Empty and switched-off assertions retain their checks, with independently caught proof, removed-check and repeated-evaluation mutants. See [the condition assertion report](../stage3/scouts/step09/predicates/condition-report.md). Q2–Q4 and Q6 remain recorded.

## Delegated proof implementation

The delegation unit builds on 223f233a. Independently proved direct helpers compose through conjunction, disjunction, negation, boolean aliases, conditionals and fixed kind switches. Literal operator chains and multi-literal kind targets keep a cell for values outside every tested literal. Helpers observing node.kind preserve the original object argument identity; unknown calls and writes refuse surviving paths. Recursive annotations never seed summaries. Both directions and every normal return are checked independently before a predicate check retires; complete target views remain.

The exact 115 delegation bodies move from 13 to 24 logical proofs and from one to eight complete admissions. Three predicates still construct five .ts call checks; 91 bodies remain refused in .a and 16 independently proved bodies remain pending on complete views. The 105 mixed-kind bodies move from three to 18 logical proofs, with only three complete admissions. All 138 optional node.original stops remain; landed V2 still deliberately refuses optional alias reads. See [the delegation report](../review/compiler/step09-predicates-ahead/delegation/delegation-report.md) and its per-body evidence for before/after counts, dependencies, Node witnesses, independent mutants and test-leaf seconds. Q2–Q4 and Q6 stay recorded.
