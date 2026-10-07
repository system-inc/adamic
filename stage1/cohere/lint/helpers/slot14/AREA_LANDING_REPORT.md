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

## Compiler landing refresh at c7991b900

Rebased onto area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da,
containing current main c7991b900362796aefd111474e65eb5398e91953. The dedup
ledger is unchanged. Main's compiler changes make input-identity reuse invalid,
so both helpers and both mutants were rebuilt and executed again in full.

Both validate_hex.py and validate_unicode.py PASS. Fixed hex: 68,844 actual
Go/control records, 2,149,768 identical bytes on each backend. Unicode: 12,935
records, 834,966 identical bytes on each backend. Each width mutant compiles,
exits zero with clean stderr, and only Go byte comparison catches it on source
Node, emitted JavaScript and ASan/UBSan native. All four original consumer suites
and every targeted real-rule private call PASS again. Fresh raw evidence is in
hex* and unicode* logs, not reused from the preceding compiler.

The rebased rule branch's full 368-file corpus PASSes across all four sides:
20,750,769 identical bytes. All ten rule mutants are freshly caught on all
three Adamic backends. Nine complete original-rule groups (579 cases) PASS;
the constructor parser and the shared Go oracle's budget panic remain failures.
No new claim is made while that branch is not green.

The inherited-static-field compiler fixture and one-byte oracle control PASS
again. go vet ./... PASSes. Setup: tools ready 0s each, submodules 0s, build
cache warm 149s, total 149s; nproc 5, quota 4 cores. Two confirmed inactive Go
scratch-build directories from earlier failed runs were removed to make space;
no source, corpus input, fixture or assertion was removed or weakened.

The four consumers and eight prerequisite entries above are unchanged; zero
complete rule blocker sets are removed. No full repository gate, full native
regexp integration or new throughput measurement is claimed. No shared harness
file was edited. The remote main/area tips were verified before pushing.

## Registry migration refresh at b46914832

Rebased onto area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53,
containing unchanged main c7991b900362796aefd111474e65eb5398e91953. The ledger
is unchanged. Both validate_hex.py and validate_unicode.py were executed again,
with all output saved under evidence/hex* and evidence/unicode*. Both PASS:
68,844 fixed-hex records / 2,149,768 identical bytes; 12,935 Unicode records /
834,966 identical bytes. Actual Go private helpers, source Node, emitted
JavaScript and ASan/UBSan native agree. Both compiling, exit-zero, clean-stderr
width mutants are caught only by Go byte comparison on all three backends.
All four original consumer suites and targeted actual leaf calls PASS again.

The filtered inherited-static-field compiler fixture and one-byte oracle control
PASS (0.200s, gate-cache hits); go vet ./... PASS. The rule branch's fresh
415-file corpus PASSes with 20,751,296 identical bytes on all four sides. Its
modified-destructured constructor parser failure remains, so no new helper is
claimed. No shared harness file was edited. Eight prerequisites for the same
four consumers still remove zero complete blocker sets. No full repository gate,
new throughput measurement or full regexp integration is claimed. Setup evidence
remains the preceding unchanged-toolchain run: 149s total, nproc 5, quota 4.
Remote bases were checked before pushing.
