Added numeric syntaxKinds exports for all twelve owned rules, with one listeners.a per directory.
Base: origin/main e8ba3d5d; prior pushed landing commit 14e11171; this commit's SHA accompanies the push report.
Command: python3 stage1/cohere/lint/rules/id-length/verify-numeric-listeners.py; PASS, 136 identical bytes against Go on source Node, emitted JavaScript and sanitized native.
Mutants: each rule's first numeric listener increased by one; all twelve compiled and ran cleanly and were caught by output comparison on all three backends.
Not covered: whole-rule parity after rebase, findings throughput, numeric dispatch or handed-in-node execution; landing remains blocked and no helpers are claimed.

Numeric values come from pinned Go cohere's AST.Kind via its public shim, not from counting enum lines or assuming stock TypeScript values. Cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db; its TypeScript parser is at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. The current registry's string kind names resolve to these same Go kinds. Numeric identifiers must be regenerated when the parser pin changes. Each entry module re-exports syntaxKinds for the forthcoming shared driver. Existing string descriptors remain valid for today's registry.

The reproducible verifier invokes the actual Go enum, regenerates the declarations, builds the compiler, and imports only listener declarations in an isolated program. It checks original source through oracle/node.mjs, independently emitted JavaScript with that runtime loader, and a --sanitize native build. No whole-rule integration is inferred from this isolated check. Native, emitted and source runs return successfully for every mutant; only comparing their bytes to Go detects the changed listener.

The first attempt's runner printed an array index without proving it present and without converting it to a string; the checker refused it. The second attempt omitted the runtime loader for emitted JavaScript and Node rejected the adamic package import. Both failures are preserved in evidence/numeric-listeners.log and numeric-listeners-retry.log. The corrected full run is evidence/numeric-listeners-final.log. Neither refused attempt counts as a mutant.

All new changes are inside the twelve owned directories. There are zero additional shared-file hunks. The four shared-file reconciliation hunks remain exactly those listed in LANDING_REPORT.md. No main rule was removed and no shared dispatch or parser was edited.

The existing parser exposes ParseNode.kind as a string and no numeric kind field. Current rule contexts refetch nodes by index. Therefore the numeric declarations are ready, but the requested numeric-only, handed-in-node execution cannot yet be wired to the shared API. Main's static Go oracle and runner also do not select these directory rules; the last owned witness gate fails with missing findings. Previous main-only oracle success does not establish parity for these ports. No fresh rule findings-per-second measurement is asserted.

Automatic approval review previously rejected the proposed broader registration bridge and option-text plumbing because it exceeded rebase-only shared-file authorization. That command did not execute. This change stays in owned directories and leaves that blocker intact. The work-in-progress cap prevents any helper claim until the landing gate is green.
