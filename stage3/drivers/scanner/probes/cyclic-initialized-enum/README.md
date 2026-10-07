Safe imported const-enum read in the original barrel evaluation graph.

Run through scanner/node.mjs with stock TypeScript 6.0.3: stdout `1\n`.
The real Error/non-null scratch compiler refuses reader.a:2:23 with
`nexus/correctness-no-import-cycle-load-time-read`. This four-file probe
reproduces the unstubbed scanner gate without Error or scanner dependencies.
No native binary is produced. See BLOCKERS.md and evidence/error-real-*.

October 7 cycles fix: native and Node both print `1\n` with exit 0 when
codex/import-cycles is combined with codex/flag-enums ec67b02. Generated C
contains no ReferenceError/TDZ guard. Current main without the existing enum
feature rejects this syntax with TS1294; the cycles change does not add enum
support. The enum combination was validated in an isolated worktree and was
not pushed into the cycles branch.
