Built: @typescript-eslint/prefer-enum-initializers candidate port in .a, complete repair records and owned oracle adapter.
Commits: claim 058844c1; implementation commit recorded in git history.
Commands and outputs: see the final parity log and replay command in ../typescript-prefer-as-const/REPORT.md.
Mutant: mutant.json changes behavior while compiling and exiting cleanly; comparison catches it on all three backends.
Not covered: standard shared frontend integration is blocked; this candidate is not merge-ready.

The rule exposes complete findings and edits through `analyze`, with directory registration through rule.json. The registered legacy `visit` refuses repair shapes its renderer cannot retain. No shared registration, harness, compiler, or parser file was edited. The three candidate directories must be integrated together: complete records and the verification driver are owned by typescript-prefer-as-const.

See [the unit evidence](../typescript-prefer-as-const/REPORT.md) for commands, corpus coverage, findings per second, failures, and reproduction.
