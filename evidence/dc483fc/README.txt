#dc483fc evidence, run as root on Linux, TMPDIR=/tmp/adamic-gate.
Head A b899c2c1 (compiler/audit-names-the-difference), head B 90094d52 (compiler/settle-compares-racing-copies), base origin/main 876fdf02.
main-race.log, a-race.log, b-race.log: go test -race -count=1 -timeout 30m ./internal/buildcache/ on each (A and B with their files staged, the content later committed). *-failing.txt: each run's failing tests; all three equal (the five known root failures). No DATA RACE in any.
a-vet.log, b-vet.log: go vet ./internal/buildcache/ (empty, exit 0).
a-focus.log, a-focus2.log: head A's tests verbose; a-focus2.log shows the audit-difference line (witness.o, ELF section .rodata.str1.1).
a-mutant.log: head A mutant, auditHit rebuilding into the stored product's directory: the new test fails (red).
b-focus.log: head B's test verbose. b-mutant.log: head B mutant, the comparison skipped: the different case fails (red).
