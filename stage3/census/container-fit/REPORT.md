# Container fit census (zrnkpdz)

Preparation in progress. No measurement result is available yet.

Pins: base d4c6b07aca1a7c55785226563c2ad79d552c1adf; head 1197ee885cb653795cda90480705e192c14144d7; logging head f913c480175d88b31adbefa5825d1283ac2da2f2.

The adapted tree is generated once by base stage3/apply.sh from TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 and its lockfile-pinned API packages. Build and adaptation logs are retained under evidence.

The reused first-error file and unit selector and validation tooling came from typescript/any-members-census at f45740c18bccdbe5b87b7d3d5d576fb79002dce9. The selector helper adds `unit <location>` markers to the logging file to attribute subsequent unchanged walk lines. This does not change lowering or findings. Base was built before this annotation and has logging disabled.

Each census child will have a 20 minute timeout. Each piece must be committed and pushed before another piece starts. Publication is checked before spending time on measurements.

First-stop changes, regressions, cap rows, pair maxima and timing comparisons remain unknown. An empty result is not inferred from preparation.
