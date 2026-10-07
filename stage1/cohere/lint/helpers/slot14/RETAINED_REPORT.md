# Slot 14 retained helpers and landing

Two executable .a helpers are delivered, each in its own file:
regexp.decodeFixedHex and regexp.decodeUnicodeEscape. Claims were pushed before
code; the latter reuses the former. The earlier Tailwind races remain withdrawn
and archived, with zero delivery/readiness credit.

Both owned branches are rebased onto origin/main
39638d9e278d38bb5aeae887f46d55a70e47aaad and re-green. Rule branch 48b945b37 is
parked with its shared context/registration/Diagnostic blocker named in the
owned module-alias PARKED.md. No live shared harness file is edited.

Commands, with source /workspace/adamic-tools/env.sh:
- python3 stage1/cohere/lint/helpers/slot14/validate_hex.py
- python3 stage1/cohere/lint/helpers/slot14/validate_unicode.py
- rule branch: validate_parking.py, then latest fixture/mutant selection and
  validate_rebase.py after proving compiled-input/corpus identity
- go vet ./... on the rule branch
- correctly filtered inherited_static_field_read.a and one-byte compiler oracle

All test output is retained in evidence/log files. Fixed hex: 68,844 records /
2,149,768 bytes. Unicode: 12,935 records / 834,966 bytes. Actual private Go
matches source Node, emitted JavaScript and sanitized native for each. Both
compiling clean-exit width mutants are caught only by comparison on all three.
All four original consumer suites pass, with actual targeted leaf invocations
in each consumer. Details, call counts, API domains and limitations are in
HEX_REPORT.md and UNICODE_REPORT.md.

Thirteen owned rule ports pass their supported 331-file corpus, fixtures,
semantic mutants and extra refusal/UTF-8/numeric/convergence controls. The enum
comparison includes 1,413,265,526 bytes, 680 findings. Throughput for every rule
(native release, source Node, Go; best of five count-only runs) is in PARKED.md.
Full corpus was rebuilt after the compiler change at b8fb957aa; it was reused
only under exact compiled-input/corpus identity after the stage3 landing.
Latest fixtures and mutants were executed again, with the rebase reports
distinguishing fresh execution from cached complete corpus evidence.

Each helper supplies four prerequisite entries for:
@next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type,
no-restricted-exports and no-restricted-imports. Total eight entries, zero
complete rule blocker sets removed. This is helper readiness, not four finished
rule ports or complete native regexp integration.

Limits: shared default rule-driver integration; explicit JSX/destructuring
parser exclusions and split-UTF-8 suggestion refusals; bounded helper domains;
no full repository gate. Archived Tailwind original consumer coverage remains
blocked by missing Kirk-local fixtures and is not counted as passing. Setup
timing lines: tools ready 0s each, submodules 16s, build/cache warm 318s, reported
total 318s; nproc 5, CPU quota 4. The temporary mistyped remote branch created
during a helper push retry was deleted with an exact-tip lease; both canonical
owned tips were verified. No main or area branch was pushed.
