- **JavaScript-changing, 99-correctness-polling-no-enum-write (Adamic, task baobab, Nov 13, 2026):** With TSC_WATCHFILE=DynamicPriorityPolling and
  TSC_WATCH_POLLINGINTERVAL_LOW/MEDIUM/HIGH=17/41/89, TypeScript 6.0.3's shipped
  _tsc.js schedules 250/500/2000 and throws on a configured caller (17/41/89).
  Its source, transpiled without bundle enum inlining, schedules 17/41/89 and
  throws on a default caller (250/500/2000). The bundle inlines enum reads while
  setCustomPollingValues mutates the runtime enum object. Adapted tsc instead
  keeps the enum immutable: defaults are scheduled and default callers throw
  nothing. Chunk-size and unchanged-threshold customization retain their existing
  computation and assignments. Adamic ruled road (1), task baobab, November 13;
  docs/rulings.md:249 refuses writes into enum objects. Whole owner:
  99-correctness-polling-no-enum-write. Evidence and unfiled upstream issue draft:
  typescript/polling-no-enum-write-evidence, stage3/census/polling-no-enum-write.
