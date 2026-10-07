Safe imported const-enum read in the original barrel evaluation graph.

Run through scanner/node.mjs with stock TypeScript 6.0.3: stdout `1\n`.
The real Error/non-null scratch compiler refuses reader.a:2:23 with
`nexus/correctness-no-import-cycle-load-time-read`. This four-file probe
reproduces the unstubbed scanner gate without Error or scanner dependencies.
No native binary is produced. See BLOCKERS.md and evidence/error-real-*.
