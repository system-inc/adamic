# Stage 3 fixtures

Each bucket owns its .a programs and status.json. The schema is the shared Stage
3 fixture contract: file, tsc source spans, census reason, recorded Node stdout,
stderr and exit, and stage0 outcome plus full diagnostic text. Diagnostics use
repository-relative paths, omit the CLI's adamic prefix and final newline, and
preserve diagnostic content. TypeScript attribution is in NOTICE.

This unit seeds runner/ with NotYet, Refused, and matching Compiles observations.
The fixture runner implementation is currently blocked by automatic approval
review: the unit-specific request assigns fixtures_test.go to this worker, but
the later shared instructions say a separate worker must build it. Review rejected
two attempts to write that file. No runner or oracle hook has been written.
Consequently the runner's Node/status/native mutants and -update behavior have
not been tested. Clarification of this conflicting ownership is pending.

No other worker's bucket has been edited.
