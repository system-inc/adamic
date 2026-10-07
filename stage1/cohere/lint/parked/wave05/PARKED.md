# Wave 05 parked candidates

These directories are outside rules/ and are not registered. Their implementations and historical evidence are retained here. Reproducer for each: run the indicated test after restoring its directory under rules/.

- boundaries-dependencies: lint-registry rejects missing witness. Full findings need Program/module-resolution/project-root facts absent from RuleContext. `go test ./stage1/cohere/lint -run ^TestOwnedWitnesses$ -count=1` stops before comparisons.
- complexity: the same owned-witness command reports `complexity witness reports no findings`; witness options are not supplied by this landed harness.
- no-cond-assign: `go test ./stage1/cohere/lint -run ^TestMutants$ -count=1` panics in oracleNoCondAssignOptions: cannot unmarshal object into Go string for an all-rule options row.
- tailwind-no-unnecessary-whitespace: `go test ./stage1/cohere/lint -run ^TestRulesAgree$ -count=1` fails upstream tailwind capture with `-run ^(Test|TestNoPhysicalDirection)`. Original regex-shaped upstreamTest was invalid; the valid Test prefix captures every matching public rule case. Tailwind inputs remain necessary.
Function-type and namespace wire labels were corrected to fix. Function-type no-fix rows now retain the finding bounds. Both stay in the landing set. Final TestOwnedWitnesses and TestRulesAgree PASS, covering 2371 captured cases. Full TestMutants completed without failures in the earlier combined run; after the final bounds correction its array-function mutant was rerun and caught on Node, emitted JavaScript and sanitized native. Parser recovery limitations inherited from the area harness are explicitly reported by TestRulesAgree, not changed here.

Initial complete gate log: evidence/initial-gate.log. Follow-up witness logs identify subsequent blockers. Failing setup or oracle execution is not a caught semantic mutant.
