# No document import in page candidate

Public name: @next/next/no-document-import-in-page. Independent oracle:
cohere/internal/lint/rules/next.NoDocumentImportInPage at cohere commit
715ba94f3608a6500086b1076ce5cb7e51b836db.

The listener reports whole static import declarations with the exact decoded
module specifier next/document, including type-only, named, namespace and
side-effect imports. Re-exports, dynamic imports, require and near misses are
silent. The document exemption preserves the Go helper's final split on the
bare text pages and prefix /_document or backslash-_document. This intentionally
includes _documentation and directory forms, and excludes nested user/_document.
No fix or suggestion is offered. New Adamic modules use .a.

This is a candidate awaiting shared integration. The actual registry rejects .a;
the compatibility proposal in the boolean-outcome directory remains unapplied.
The existing profile API mismatch also blocks the default package gate. The
scratch overlay enables comparison of this candidate on Go, source Node,
emitted JavaScript and sanitized native without shared source edits.

Witnesses, ten path cases and 215 compiler/stage1 files agree byte for byte.
The mutant inverts the document exemption, compiles, exits normally with no
stderr, and is caught only by comparison on all three Adamic execution paths.
Count-only best-of-five throughput, 77 compiler files plus 1,000 reporting
imports in one synthetic file: native 682.24, Node 951.59, Go 3883.93 findings/s.
The compiler files alone produce zero findings. Native throughput uses the
release build; correctness and mutants use ASan/UBSan. Startup and parsing are
included, compilation and formatted finding/fix output are excluded. These runs
overlapped other bounded checks; individual rounds are retained in the logs.

Full captured cohere fixture parity does not pass. The capture loses filenames
and the shared Go oracle hardcodes ScriptKindTS. JSX in the original document
rule fixtures is rejected as invalid corpus; the stage1 parser independently
refuses JSX. No original JSX fixture is simplified or dropped to claim parity.
The Go family's original tests pass independently, which proves its reference
behavior rather than this candidate's full parity. No full repository gate or
unconditional merge readiness is claimed.

validate.py reconstructs the scratch overlay from the earlier reviewable
compatibility.patch and this directory's validation_test.go.txt. Source
/workspace/adamic-tools/env.sh, then use --scratch and --typescript (the latter
points at the pinned TypeScript v6.0.3 checkout). Its bounded parity filter
excludes the known-failing TestRulesAgree and runs witnesses, compiler/stage1,
paths and rates. Full-fixture failure is separately recorded in the
[batch report](../../claims/wave1-12-batch3-evidence/REPORT.md).
