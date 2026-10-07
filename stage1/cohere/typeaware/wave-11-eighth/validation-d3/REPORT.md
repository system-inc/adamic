Rebased wave 11 onto lint integration d3a37422, including current main b6b1538b and inherited typeof-null fixes.
The evidence commit is pushed only to codex/typeaware-wave-11.
All eight owned suites, full bridge tests, filtered compiler oracle, extra typeof oracle and package vet PASS.
All 46 worker mutant observations revalidated through independent byte comparison and ownership checks.
No unclaimed rules: 657 origin refs, 641 distinct trees, 34 reservation documents; no new claims or implementation.

Complete commands, exits and logs are adjacent. Seventh process 94.248s,
eighth 115.682s, bridge 93.385s, compiler 8.272s. Additional command:
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_(dispatch|null|null_compare|null_roll|null_slots|null_switch|string_literal)[.]a$' -count=1 -timeout 10m -v,
with output directly to typeof.log, PASS 12.028s. It compares Node, native
sanitizers and emitted JavaScript. Applicable rule suites compare controls
and both frozen corpora including findings/fixes/suggestions, sanitizers
and released handles. Evidence writing initially hit disk exhaustion; only
named obsolete worker archives/binaries were removed, retaining source/logs.
No protected compiler or shared harness/generator source changed. Tested main
and area remote heads were verified unchanged after checks.

Full gate and its 17 required external-input checks were not run. Older
standalone numeric metadata is not claimed compatible with the shared
name-based registry. Shared emitted-JavaScript lint and nondefault options
remain uncovered; four analysis claims remain parked. Original quiet
native/Go times remain repository 0.9227/0.1380s and compiler 7.0781/0.3421s
on c01907a7; current concurrent timing observations are in logs.
