# Scratch compiler integration conflicts

No compiler edits or scratch merge commits were pushed. Final scratch SHA:
6ade8c38626a2a7d95cd0a21e62c9682361c39e0. All requested commits are ancestors:
area/stage3 4ad53a4, compiler/stage3-front-2 80fb9b75, import-cycles 779ff9d,
library/qualified-as-const 8280fd08. The prior scanner feature scratch e034d3d
was retained, including Error 6469f37 and non-null a02613e.

Lowering conflicts while retaining the prior scanner feature integration:

- expression.go: combine import-cycle checkedModuleRead with Readiness.
- invariance.go: keep wideningRefusal factoring and the enum-specific refusal body.
- predicates.go: retain the newer predicateRefusal admission seam and body prover.
- refusals.go: retain predicate/enum/namespace admission, non-null support, and Error checks.
- statements.go: retain named/star exports and nested declarations; remove duplicate export dispatch.

Lowering conflicts when merging front-2:

- cast.go: retain front-2 castProof, including its runtime-check refusal proofs.
- object.go: retain the enum checked-default body while the front-2 fallthrough changes merge.
- refusals.go: retain castProof preflight alongside predicate/enum/namespace/Error admission.

Other conflicts: load.go and source_fs.go combined the rooted-path checker API
with the Node declaration loader; load_test.go takes front-2 house-style options;
escape-hatches docs, counts, fixture NOTICE, oracle fixtures and panic formatting
were resolved in scratch. Further rooted-filename conversions in node_library.go
and library_node.go were needed to build the merged compiler.

The qualified-as-const merge conflicts only in oracle counts; its lowering changes
(invariance.go and regexp.go) merged cleanly. Refreshing area/stage3 and front-2
reported already up to date; explicitly merging 779ff9d did too.

Compiler build succeeds. The safe-cycle probe prints 1 natively. The real scanner
then refuses the MapLike type's index signature before lowering its bodies. This
result does not certify every scratch conflict resolution or the full compiler gate.
