Built: rebased wave-09 onto lint area d65a8f931 with its allocator/string-search improvements preserved and re-greened all checks.
Commits: old remote 63b1fd68c; rebased implementation 0228face0; area d65a8f931 includes current main 39638d9e2; evidence commit follows.
Commands/output: eighteen owned checks, original rules, registry, bridge/checker, filtered Node oracle and targeted vet PASS.
Mutants: all existing semantic, handle and sanitizer mutants reran; handed-node refetch and dynamic/static RegExp checks pass too.
Not covered: complete regex rule parity, shared installation of owned checker rules or the full repository gate; no new claims.

Origin/area/stage1-lint advanced from 7481e0324 to
d65a8f931c98655936ae04c6899f38f14862b73e, integrating lint runtime
profiling, allocator and lastIndexOf improvements. The explicit area rebase
completed without conflicts and the compiler was rebuilt. Shared source was
not edited or reverted. Final main fetch confirms current main
39638d9e278d38bb5aeae887f46d55a70e47aaad remains an ancestor.
Publication is solely to codex/typeaware-wave-09, with an exact lease on old
remote 63b1fd68ca1805678c0cac3cc009b6db27c49ad1. No main or area push.

Commands/manifests are those in LANDING_AREA_REPORT.md, replacing the
log/artifact prefix with /workspace/wave-09-runtime. All eighteen owned
Python verifiers were rerun on the rebuilt compiler, including the corrected
named descriptors and handed-node rule. Package commands are repeated there;
the independent Node selection additionally includes runtime_last_index_of.a
and tests matching LastIndex. Test output goes to files, never pipes.
Exact fresh logs are retained in validation-landing-runtime.

Original suite PASS 108.437 seconds, checker PASS 0.176 seconds,
bridge PASS 70.483 seconds, filtered Node PASS 5.001 seconds. Registry and
targeted vet pass. Original rules match Go's 44 control, 14 compiler and
4 repository findings, including complete fixes/suggestions and sanitizers.
Label controls retain 18 findings / 7700 bytes; 77 compiler sources zero /
7859 bytes; 287 repository sources zero / 18485 bytes. The one parser-refused
control remains explicitly recorded. Full literal-slice and pure component
comparisons retain native, sanitizer, source Node and emitted JS agreement.

All mutants enumerated in LANDING_AREA_REPORT.md reran: rule eligibility,
radix and assertion judgments, raw flags/scope, regex sequence/parser/mapping
and formatter decisions, suggestions, constructor spans, constants and alias
eligibility, named listener declarations, Unicode folding, equivalence groups
and class escapes. They compile/run and fail independent Go comparisons.
The handed-node target-refetch mutant exits zero and differs at byte 43;
named-kind mutations differ at byte 276. Handle-retention mutations remove
required panic 70 and are caught. Bridge memory/length/ownership mutations
are caught by its sanitizer/refusal tests. The dynamic constructor still
refuses nonconstant patterns, while its static-input mutant builds/runs with
sanitizers and is caught by the required-refusal assertion. Fresh logs retain
every catcher; no new mutant was needed for a unit with no implementation edit.

Original native/Go whole-process times: compiler 2.445370/0.594211 seconds,
repository 0.344062/0.175057 seconds. Label medians:
compiler 1.515877/0.351535 seconds, repository 0.242100/0.157275 seconds.
Concurrent runs are not isolated throughput evidence. Successful setup is
reused: 88 seconds, nproc 5. No full repository gate was run.

The shared harness remains present on the area base. Named kinds and the
handed-node API are available; historical absent/numeric API blockers are
resolved. Shared checker registry installation and remaining legacy visitor
migration are unfinished integration work. Dynamic native RegExp remains an
independently reproduced lowering blocker. No hand-rolled matcher, regex
parser or rewrite was added. Both regex claims remain incomplete and the
React parking exception does not apply. No further batch is claimed while
these remain unfinished. The required area/runtime rebase was this unit.
