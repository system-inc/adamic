# Parser scout corpus

Branch parser-scout/census, base ad7bd06632f1. Shared checkout and parser sources are unchanged.

Initial checkpoint: every .ts file in microsoft/TypeScript src/compiler at v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8 (77 files), the parser gate's pinned corpus.

The full quiet hundred lives on Kirk's Mac and is unavailable in this environment. GitHub is reachable. This scout is shallow-fetching the 23 provided public repository snapshots, listed with explicit repository mappings and exact SHAs in stage1/typescript/parser/census/public-pins.tsv. Those snapshots are a public subset, not the quiet hundred. Every physical .ts/.tsx file in each successfully checked-out snapshot is selected, including declarations and malformed fixtures. No installs or repository scripts run. See census/fetch-results.json for exact per-repository file inventory and any failed fetches.

The census driver, shortest corpus witnesses, complete trees, and counts are under stage1/typescript/parser/census; the report is stage1/typescript/parser/CENSUS.md. Adapter/runner failures are separated from differences. Broader census pending at this initial checkpoint.
