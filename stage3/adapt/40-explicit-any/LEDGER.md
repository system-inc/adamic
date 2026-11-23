- z5d0jye, Nov 12 2026, type-only: amend 40-explicit-any rule 103 to emit
  `interface PathPathComponents<T extends Path = Path> extends Array<T>` with
  the existing undefined phantom marker. Its type argument is the base's own;
  all 13 uses remain bare through the default. The type is internal;
  no API sanction row is added. Refresh sanctioned-api.json’s class-rules proof
  hash and shard-plan.json’s runtime hash; the sanction count remains 267.
  Ten JavaScript outputs and both public
  declaration files are byte-identical; Adamic types and stock tsc have zero
  diagnostics on all 79 roots. Full apply passes, 40's post-stage replay makes
  zero edits, and restoring the non-generic interface fails 40's check.
  The path.ts full
  latent site totals are unchanged: 20 NotYet, 6 Refused, 32 Boundary. Eleven
  identities change only the displayed type to PathPathComponents<Path>.
  The existing getPathComponents overload-result refusal remains for
  Optimizer's follow-up. Evidence: typescript/path-components-evidence.
