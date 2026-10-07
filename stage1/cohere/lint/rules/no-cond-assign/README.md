# no-cond-assign

Implements upstream except-parens/always behavior, compound assignment tokens and conditional ancestry. The owned assignment-token mutant is caught by all three comparison runtimes.

Candidate implementation in `.a`; default directory discovery still requires `rule.ts`. Shared infrastructure was not edited. Four-way comparison uses the reproducible scratch overlay in [wave1-05 evidence](../../claims/wave1-05-evidence/REPORT.md). These candidates are not certified for the full requested corpus until the integration and parser boundaries are resolved.
