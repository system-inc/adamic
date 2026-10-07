# Area landing recheck

Rebased the helper branch onto origin/area/stage1-lint
 d65a8f931c98655936ae04c6899f38f14862b73e, which includes current main
39638d9e278d38bb5aeae887f46d55a70e47aaad and the landed lint harness. Only the
owned helper commits were replayed. The shared helpers/readiness/inventory
foundation has identical blobs on the area and codex/lint-helpers branches.
No shared harness file was edited. No new helper was claimed.

Fresh commands, with source /workspace/adamic-tools/env.sh, all output logged:

- python3 stage1/cohere/lint/helpers/slot14/validate_hex.py: PASS; actual Go
  private helper against source Node, emitted JavaScript and ASan/UBSan native;
  68,844 records and 2,149,768 output bytes per backend.
- python3 stage1/cohere/lint/helpers/slot14/validate_unicode.py: PASS; same
  comparison, 12,935 records and 834,966 output bytes per backend.
- go test ./internal/oracle -run
  'TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read|TestTheOracleCatchesOneByte'
  -count=1 -timeout 30m -v: PASS after recovering disk space. The first two
  attempts failed with no space left on device before any comparison; 628 old
  reproducible Go cache entries (6,446,209,870 bytes) were removed, then the
  unchanged check passed. No fixture or assertion was relaxed.

Each helper's width mutant compiles and exits zero with clean stderr on all
three backends, and only comparison with Go catches it. Both original suites
and targeted actual private calls pass for each of the four consumer rules.
Evidence remains under evidence/hex* and evidence/unicode*, with actual counts.

The two helpers supply four prerequisites apiece for
@next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type,
no-restricted-exports and no-restricted-imports: eight prerequisite entries,
zero complete rule blocker sets removed. Full regexp integration is not claimed.

The separate rule branch is rebased onto the same lint area and applying the
dedup ledger. Its old missing-harness parking blocker is obsolete. A fresh
shared TestOwnedWitnesses instead fails because the unchanged Go oracle panics
on the existing no-lonely-if eleven-pass budget witness. The witness is kept.
The rule branch will report its integrated results and reproducible blockers;
no new work is taken while that branch is not green.

Toolchain setup on the rebased rule worktree: tools ready 0s each, submodules
0s, build cache warm 89s, total 89s; nproc 5, CPU quota 4 cores. No full repository
gate is claimed. Historical withdrawn Tailwind deliveries remain uncredited.
