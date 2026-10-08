No-inferrable-types now passes the unified harness with automatic multi-edit fixes.
Rebased onto requested 59451e23eefc9d83102afd6d1a26050d6c447227, which contains fca1616e6; tested source d31ef0e459301a488bf15e43305396c2014e31d6 plus the new owned witnesses.
TestRulesAgree PASS 56.52s, TestMutants PASS 766.79s, TestOwnedWitnesses PASS 22.19s; total 845.513s.
All 41 registered mutants are caught, including number-inference on source Node, emitted JavaScript and sanitized native.
The former shared multifix blocker is resolved; full repository and performance gates were not run.

The fetched area b28757f33 did not yet contain fca1616e6 or 59451e23e, so the user-specified
commit is the rebase target. Only the owned port branch is pushed. No shared code changed.
The existing rule already passes the complete edits array to context.reportNode, preserving
upstream edit order and independently proposed removals. No implementation change was needed.

New owned witnesses:

- multifix.ts.txt: optional parameter `const fn = (a?: number = 5) => {};` and definite-assignment
  property `class C { value!: number = 5; }`, each carrying two automatic edits.
- ignore-parameters, ignore-properties and ignore-both, each with its own options.json sidecar,
  retain a variable violation so every configured witness still fires. Their source also includes
  parameter/property controls and a readonly property exemption.

Commands after sourcing /workspace/adamic-tools/env.sh:

- go run ./cmd/lint-registry: exit 0.
- go vet ./...: exit 0; gofmt -l on owned oracle.go: empty output.
- go test ./stage1/cohere/lint -run '^TestRulesAgree$|^TestOwnedWitnesses$|^TestMutants$' -count=1 -v -timeout=30m:
  exit 0, PASS 845.513s, no test skips.

TestRulesAgree captures 2128 unique source/rule/options combinations and compares 13,101,430 bytes
identically between independent Go cohere, source Node, emitted JavaScript and sanitized native.
The TestNoInferrableTypes prefix retains every one of the five upstream test functions. Shared
method-signature malformed cases remain explicit recovery refusal checks, not silent skips or
newly certified parser recovery. They are not failures of this owned rule.
TestOwnedWitnesses compares 124,131 identical bytes, including selected and all-rule runs, options,
findings, all fix-edit rows and converged fixed output. The original optional-parameter reproducer
and definite-assignment property now compare successfully instead of panicking.

Full TestMutants runs every registered descriptor, rather than only the owned subtest. The owned
number-inference mutant compiles and executes; changing number to empty removes real findings
and fixes and only byte comparisons catch it on all three execution paths. Exact catches and
all 41 passing subtests are retained in evidence/multifix-refresh/tests.log. ASan/UBSan and the
shared native ownership checks are part of these comparisons. Historical reports/logs remain
unchanged apart from the superseding link in REPORT.md.

Not covered by this unit: the full repository gate, a new throughput benchmark, every project
outside captured upstream/witness/generated inputs, or unrelated shared parser recovery support.
No guards, upstream rules, dispatch lists or oracle serializers were modified by this branch.
