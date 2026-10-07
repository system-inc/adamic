# Cohere lint rule inventory summary

The complete machine-readable inventory is [inventory.json](inventory.json).
It records all 491 rules, source files and line counts, test locations and counts,
type-information needs, fixes, options, shared dependencies, firing frequencies,
and observed stage 1 branch status. Tools should read that JSON file.

This compact summary replaces the 52,487-line expanded Markdown inventory,
which caused the whole-document oracle preflight to time out. Keep expanded
per-rule dependency reports outside tracked Markdown documents.

| Family | Rules |
|---|---:|
| adamic | 7 |
| base | 16 |
| boundaries | 1 |
| core | 172 |
| next | 22 |
| nexus | 51 |
| react | 81 |
| structure | 32 |
| tailwind | 13 |
| typescript | 96 |

## Evidence and handoff

- Pinned cohere revision: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- Inventory observations and limitations remain in `inventory.json`.
- Shared helper claims and measured readiness: [HELPERS.md](../HELPERS.md).
- First helper handoff: [helpers/REPORT.md](../helpers/REPORT.md).
- Comment helper handoff: [helpers/comments/REPORT.md](../helpers/comments/REPORT.md).

Helper-ready counts assume the common AST/reporting adapter; they do not claim
completed rule implementations or full fixture parity. Use the current helper
handoff ledgers alongside the frozen inventory when assigning workers.
