# eslint-comments-require-description

Implements upstream directive recognition, descriptions, ignore/additional-directive options and comment spans outside literals. One upstream JSX case and one unterminated-comment case remain outside the shared parser/driver contract.

Candidate implementation in `.a`; default directory discovery still requires `rule.ts`. Shared infrastructure was not edited. Four-way comparison uses the reproducible scratch overlay in [wave1-05 evidence](../../claims/wave1-05-evidence/REPORT.md). These candidates are not certified for the full requested corpus until the integration and parser boundaries are resolved.
