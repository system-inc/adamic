Added lowering guards for inherited field order, override diagnostic classification and nonempty inheritance answers.
Base: origin/main 291ee604ad0dd90894ad4503054eea60bf4f5ee6; delivery: compiler/inherit-guards.
Focused lowering and Node oracle checks passed before mutations and after restoration.
M02, M07 and P_LOWER each failed their intended checks.
Production code unchanged; no whole-package gate or repo-wide mutant sweep.

Task #a898y40 adds test guards at the layer computing class metadata. The shared fixture is the audit's W1: A declares x, B extends A and declares y, and a B instance prints its inherited x. Source Node prints 1 with exit 0 and empty stderr. Generated JavaScript, sanitized native, release native and leak checks agree.

Commands use the toolchain environment from /workspace/adamic-tools/env.sh, GOPROXY=https://proxy.golang.org|direct and outer timeout 120. Go test commands use -count=1 -timeout 90s. Logs retain complete output.

- Lower: go test ./internal/lower -run 'TestInheritance(DerivedFieldOrder|VoidBooleanOverrideReason|AllowsSoundOverrides|KeepsNominalTupleDestructuring|NativeSignatureNeighbors)$' -v.
- Oracle: go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(inherited_fields_guard|class_inheritance)\.a$' -v.
- Counts: go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts.
- Vet: go vet ./internal/lower ./internal/oracle.

Restored test seconds: DerivedFieldOrder 0.12; VoidBooleanOverrideReason 0.09; sound overrides and nominal tuple acceptance below 0.2 each; native signature neighbor leaves 0.06–0.13. New oracle witness 0.37; existing class_inheritance.a 0.53. First baseline oracle has zero Node cache hits (four misses). Every new leaf remains below 60 seconds.

Independent mutants, copied as .diff from test-audit/internal-lower-class_inheritance:
- M02 removes inherited metadata.Fields append in internal/lower/class.go. TestInheritanceDerivedFieldOrder fails with B.Fields=[y], expecting [x,y], 0.04 seconds. New inherited_fields_guard.a oracle fails: Node prints 1, native stops exit 70 on a missing checker-proven field. Existing class_inheritance.a also catches M02 downstream. Oracle leaves 0.54 and 0.70 seconds.
- M07 changes overrideResultRepresentation's void/never representation from 0 to ir.Boolean in internal/lower/class_inheritance.go. TestInheritanceVoidBooleanOverrideReason fails in 0.03 seconds: Refused widened return type instead of NotYet native result representation with its repair.
- P_LOWER returns nil,nil from Lower. All three acceptance-family members fail the nonempty-answer floor, including all four native signature neighbor leaves; no nil dereference. Package test elapsed 0.119 seconds.

Each exact audit diff was applied independently and reversed before restored checks. No mutant selection or production change is committed.

Setup succeeded: node 0.022s; Go 0.024s; markdown dependencies 0.091s; submodules 0.093s; clang 0.193s; Go build 36.889s; cache warm 37.122s; total 37.151s. nproc=5; cpu.max=400000 100000 (4 CPUs). Setup log includes tool versions and individual timing lines.

Counts regeneration passed in 50.712 seconds. Only inherited_fields_guard.a was added: this records the new inherited-instance witness; existing fixture rows did not change. Exact allocation, free, retain, release, peak and region values are in counts.diff. Vet succeeded with empty output.

Integration lane checks passed: "lane checks 1.1 s: gofmt and tools on 3 Go files, t.Parallel on 2 test packages; vet 2 packages". Test commit: 1bf711d5a. Counts row: allocations 2, frees 2, retains 1, releases 4, peak live 2, in regions 0.
